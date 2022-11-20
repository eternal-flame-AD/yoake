package servetpl

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"strings"
	textTemplate "text/template"

	"github.com/eternal-flame-AD/yoake/config"
	"github.com/eternal-flame-AD/yoake/internal/servetpl/funcmap"
	"github.com/eternal-flame-AD/yoake/internal/session"
	"github.com/labstack/echo/v4"
)

type TemplatePath struct {
	File string
	Name string
}

type Context struct {
	CleanPath    string
	Config       func() config.C
	C            echo.Context
	Request      *http.Request
	Response     *echo.Response
	WriteHeaders func(headers ...string) error
	Session      session.Provider
	Global       map[string]interface{}
}

type bodyBuffer struct {
	resp    *echo.Response
	bodyBuf bytes.Buffer

	committed bool
}

func (b *bodyBuffer) Write(p []byte) (int, error) {
	if !b.committed {
		return b.bodyBuf.Write(p)
	}
	return b.resp.Write(p)
}

func (b *bodyBuffer) WriteHeader(headers ...string) error {
	if b.committed {
		return nil
	}
	for _, header := range headers {
		h := strings.SplitN(header, ":", 2)
		if len(h) != 2 {
			return fmt.Errorf("invalid header %s", header)
		}
		h[0] = strings.TrimSpace(h[0])
		h[1] = strings.TrimSpace(h[1])
		b.resp.Header().Set(h[0], h[1])
	}
	b.committed = true
	if _, err := io.Copy(b.resp, &b.bodyBuf); err != nil {
		return err
	}
	return nil
}

var (
	errUndefinedTemplate = errors.New("undefined template")
	errTplExtNotStripped = errors.New("this is a template file and should be requested without the template extension inflix")
)

func ServeTemplateDir(dir string) echo.MiddlewareFunc {
	dir = filepath.Clean(dir)
	var tplFiles []TemplatePath

	if err := filepath.Walk(dir, func(file string, stat os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if stat.IsDir() {
			return nil
		}
		ext := path.Ext(file)
		secondExt := path.Ext(strings.TrimSuffix(file, ext))
		if secondExt == ".tpl" {
			relPath, err := filepath.Rel(dir, file)
			if err != nil {
				return err
			}
			tplFiles = append(tplFiles, TemplatePath{File: file, Name: "/" + relPath})
		}
		return nil
	}); err != nil {
		log.Panicf("templates failed to parse: %s", err)
	}

	templates := template.New("").Funcs(funcmap.GetFuncMap())
	textTemplates := textTemplate.New("").Funcs(funcmap.GetFuncMap())
	for _, file := range tplFiles {
		log.Printf("parsing template: %s", file.Name)

		if path.Ext(file.File) == ".html" {
			if _, err := ParseTemplateFileAs[template.FuncMap](templates, file.Name, file.File); err != nil {
				log.Panicf("templates failed to parse: %s", err)
			}
		} else {
			if _, err := ParseTemplateFileAs[textTemplate.FuncMap](textTemplates, file.Name, file.File); err != nil {
				log.Panicf("templates failed to parse: %s", err)
			}
		}

	}
	dispatchTemplate := func(file string) func(wr io.Writer, data any) error {
		ext := path.Ext(file)
		tplName := file[:len(file)-len(ext)] + ".tpl" + ext
		if path.Ext(file[:len(file)-len(ext)]) == ".tpl" {
			if _, err := os.Stat(filepath.Join(dir, file)); err == nil {
				return func(wr io.Writer, data any) error {
					return errTplExtNotStripped
				}
			}
		}

		tplPath := filepath.Join(dir, tplName)
		if _, err := os.Stat(tplPath); err == nil {
			if ext == ".html" {
				return func(wr io.Writer, data any) error { return templates.ExecuteTemplate(wr, tplName, data) }
			} else {
				return func(wr io.Writer, data any) error { return textTemplates.ExecuteTemplate(wr, tplName, data) }
			}
		}
		return func(wr io.Writer, data any) error { return errUndefinedTemplate }
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req, resp := c.Request(), c.Response()
			p := path.Clean("/" + req.URL.Path)
			ext := path.Ext(p)
			c.Response().Header().Set(echo.HeaderContentType, mime.TypeByExtension(ext))

			body := &bodyBuffer{resp: resp}
			defer body.WriteHeader()

			sess, sessClose := session.ManagedSession(c)
			defer sessClose()
			if err := dispatchTemplate(p)(body, Context{
				Config:       config.Config,
				C:            c,
				CleanPath:    p,
				Request:      req,
				Response:     resp,
				WriteHeaders: body.WriteHeader,
				Session:      sess,
				Global:       map[string]interface{}{},
			}); err == errUndefinedTemplate {
				return next(c)
			} else if errors.Is(err, funcmap.ErrEarlyTermination) {
				return nil
			} else if errors.Is(err, errTplExtNotStripped) {
				c.String(http.StatusBadRequest, err.Error())
			} else if err != nil {
				c.Response().Write([]byte(fmt.Sprintf("<!-- ERROR: %v -->", err)))
				return err
			}
			return nil
		}
	}
}
