package comm

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"path"
	"strings"
	textTemplate "text/template"

	"github.com/PuerkitoBio/goquery"
	"github.com/eternal-flame-AD/yoake/internal/comm/model"
	"github.com/eternal-flame-AD/yoake/internal/servetpl/funcmap"
	"github.com/gomarkdown/markdown"
)

func contains[T comparable](s []T, e T) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

type ErrorMIMENoOverlap struct {
	MessageMIME   string
	supportedMIME []string
}

func (e *ErrorMIMENoOverlap) Error() string {
	return fmt.Sprintf("message MIME type %s is not supported by this communicator. Supported MIME types are: %s", e.MessageMIME, e.supportedMIME)
}

func ConvertGenericMessage(msgOrig *model.GenericMessage, supportedMIMES []string) (*model.GenericMessage, error) {
	if contains(supportedMIMES, msgOrig.MIME) {
		return msgOrig, nil
	}
	msg := *msgOrig
	if strings.HasSuffix(msgOrig.MIME, "+html/template") ||
		strings.HasSuffix(msgOrig.MIME, "+text/template") {
		var output bytes.Buffer
		var err error
		var tplName string
		if strings.HasSuffix(msgOrig.MIME, "+html/template") {
			tpl := template.New("").Funcs(funcmap.GetFuncMap())
			if strings.HasPrefix(msgOrig.Body, "@") {
				tplFiles := strings.Split(msgOrig.Body[1:], ",")
				tplName = path.Base(tplFiles[len(tplFiles)-1])
				tpl, err = tpl.ParseFiles(tplFiles...)
				if err != nil {
					return nil, err
				}
			} else {
				tpl, err = tpl.Parse(msgOrig.Body)
				if err != nil {
					return nil, err
				}
			}
			log.Printf("template name is: %s %s", tpl.Name(), tpl.DefinedTemplates())
			if err := tpl.ExecuteTemplate(&output, tplName, msgOrig.Context); err != nil {
				return nil, err
			}
			msg.MIME = strings.TrimSuffix(msgOrig.MIME, "+html/template")
		} else {
			tpl := textTemplate.New("").Funcs(funcmap.GetFuncMap())
			if strings.HasPrefix(msgOrig.Body, "@") {
				tplFiles := strings.Split(msgOrig.Body[1:], ",")
				tplName = path.Base(tplFiles[len(tplFiles)-1])
				tpl, err = tpl.ParseFiles(tplFiles...)
				if err != nil {
					return nil, err
				}
			} else {
				tpl, err = tpl.Parse(msgOrig.Body)
				if err != nil {
					return nil, err
				}
			}
			log.Printf("template name is: %s %s", tpl.Name(), tpl.DefinedTemplates())
			if err := tpl.ExecuteTemplate(&output, tplName, msgOrig.Context); err != nil {
				return nil, err
			}
			msg.MIME = strings.TrimSuffix(msgOrig.MIME, "+text/template")
		}
		msg.Body = output.String()
	}
	if contains(supportedMIMES, msg.MIME) {
		return &msg, nil
	}

	// convert markdown to html
	if msg.MIME == "text/markdown" && !contains(supportedMIMES, "text/markdown") {
		msg.Body = string(markdown.ToHTML([]byte(msg.Body), nil, nil))
		msg.MIME = "text/html"
	}
	// convert html to text
	if msg.MIME == "text/html" && !contains(supportedMIMES, "text/html") && contains(supportedMIMES, "text/plain") {
		docBuf := strings.NewReader(msg.Body)
		doc, err := goquery.NewDocumentFromReader(docBuf)
		if err != nil {
			return nil, err
		}
		msg.Body = doc.Text()
		msg.MIME = "text/plain"
	}

	if !contains(supportedMIMES, msg.MIME) {
		return nil, &ErrorMIMENoOverlap{
			MessageMIME:   msg.MIME,
			supportedMIME: supportedMIMES,
		}
	}
	return &msg, nil
}
