package gomod

import (
	"archive/zip"
	"fmt"
	"math/rand"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/eternal-flame-AD/yoake/internal/echoerror"
	"github.com/labstack/echo/v4"
	"golang.org/x/mod/module"
)

type Info struct {
	Version string    // version string
	Time    time.Time // commit time
}

const backendProxy = "https://proxy.golang.org/"

func random143String() string {
	randItems := []string{
		"143",
		"143.",
		"4.3",
		"143",
		".143",
		"1.43",
		"1.4.3",
		"14.3",
		"omo",
		"om.o",
		"o.m.o",
		"o.mo",
	}

	var ret strings.Builder
	ret.WriteString("v1.4.3-")
	for i := 0; i < 10; i++ {
		rand.Intn(len(randItems))
		ret.WriteString(randItems[rand.Intn(len(randItems))])
	}
	ret.WriteString("+incompatible")
	return ret.String()
}

func resolveModule(c echo.Context, modPathUnesc string, pRequest string) error {
	modPath, err := module.UnescapePath(modPathUnesc)
	if err != nil {
		return echoerror.NewHttp(400, fmt.Errorf("invalid module path: %w", err))
	}
	if !strings.HasPrefix(modPath[strings.IndexRune(modPath, '/')+1:], "test-") {
		return echoerror.NewHttp(400, fmt.Errorf("please prefix your module path with 'test-'"))
	}
	if err != nil {
		return echoerror.NewHttp(400, fmt.Errorf("invalid module path: %v", err))
	}
	if pRequest == "list" {
		return c.String(143, random143String()+"\n")
	}
	ext := path.Ext(pRequest)
	version := strings.TrimSuffix(pRequest, ext)
	if err := module.Check(modPath, version); err != nil {
		return echoerror.NewHttp(400, fmt.Errorf("invalid version: %v", err))
	}
	switch ext {
	case ".info":
		return c.JSON(143, Info{
			Version: version,
			Time:    time.Now().Add(-time.Hour),
		})
	case ".mod":
		return c.String(143, strings.Repeat("\n", 143-1)+"Welcome.to.white.space.\n")
	case ".zip":
		zipFile := zip.NewWriter(c.Response().Writer)
		mainGo, err := zipFile.Create(fmt.Sprintf("%s@%s/main.go", modPath, version))
		if err != nil {
			zipFile.Close()
		}
		_, err = mainGo.Write([]byte(("package main\n\nimport \"fmt\"\n\nfunc main() {\n\n}\n")))
		if err != nil {
			zipFile.Close()
			return echoerror.NewHttp(500, fmt.Errorf("failed to write main.go: %v", err))
		}
		c.Response().Status = 143
		return zipFile.Close()
	}
	return c.Redirect(302, backendProxy+modPathUnesc+"/@v/"+pRequest)
}

func Register(baseURI string, baseG *echo.Group) (gogetMiddleware echo.MiddlewareFunc) {
	baseG.GET("*", func(c echo.Context) error {
		fullURI := c.Request().RequestURI
		if !strings.HasPrefix(fullURI, baseURI) {
			return echo.ErrNotFound
		}
		fullURI = strings.TrimPrefix(fullURI, baseURI)
		fullURIUnEscaped, err := url.PathUnescape(fullURI)
		if err != nil {
			return echoerror.NewHttp(400, fmt.Errorf("invalid URI: %v", err))
		}
		if strings.Contains(fullURIUnEscaped, "/@v/") {
			split := strings.SplitN(fullURIUnEscaped, "/@v/", 2)
			return resolveModule(c, strings.TrimPrefix(split[0], "/"), split[1])
		}
		return echo.ErrNotFound
	})

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method == "GET" && c.QueryParam("go-get") == "1" {
				unescapedPath, err := url.PathUnescape(c.Request().URL.Path)
				if err != nil {
					return echoerror.NewHttp(400, fmt.Errorf("invalid module path: %v", err))
				}
				ctx := goGetHtmlTemplateCtx{
					ModulePath: c.Request().Host + unescapedPath,
					ProxyBase:  c.Scheme() + "://" + c.Request().Host + baseURI,
				}

				return goGetHtmlTemplate.Execute(c.Response().Writer, ctx)
			}
			return next(c)
		}
	}
}
