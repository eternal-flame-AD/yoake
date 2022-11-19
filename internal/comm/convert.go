package comm

import (
	"bytes"
	"fmt"
	"html/template"
	"path"
	"strings"
	textTemplate "text/template"

	"github.com/PuerkitoBio/goquery"
	"github.com/eternal-flame-AD/yoake/internal/comm/model"
	"github.com/eternal-flame-AD/yoake/internal/servetpl/funcmap"
	"github.com/eternal-flame-AD/yoake/internal/util"
	"github.com/gomarkdown/markdown"
)

func unique[T comparable](s []T) []T {
	var result []T
	for _, e := range s {
		if !util.Contain(result, e) {
			result = append(result, e)
		}
	}
	return result
}

type ErrorMIMENoOverlap struct {
	MessageMIME   string
	supportedMIME []string
}

func (e *ErrorMIMENoOverlap) Error() string {
	return fmt.Sprintf("message MIME type %s is not supported by this communicator. Supported MIME types are: %s", e.MessageMIME, e.supportedMIME)
}

func ConvertOutMIMEToSupportedInMIME(outMIMEs []string) (inMIMEs []string) {
	inMIMEs = outMIMEs
	for _, out := range outMIMEs {
		if out == "text/plain" {
			inMIMEs = append(inMIMEs, "text/html", "text/markdown")
		}
		if out == "text/html" {
			inMIMEs = append(inMIMEs, "text/markdown")
		}
	}
	for _, in := range inMIMEs {
		inMIMEs = append(inMIMEs, in+"+html/template", in+"+text/template")
	}
	inMIMEs = unique(inMIMEs)
	return
}

func ConvertGenericMessage(msgOrig *model.GenericMessage, supportedMIMES []string) (*model.GenericMessage, error) {
	if util.Contain(supportedMIMES, msgOrig.MIME) {
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
			if err := tpl.ExecuteTemplate(&output, tplName, msgOrig.Context); err != nil {
				return nil, err
			}
			msg.MIME = strings.TrimSuffix(msgOrig.MIME, "+text/template")
		}
		msg.Body = output.String()
	}
	if util.Contain(supportedMIMES, msg.MIME) {
		return &msg, nil
	}

	// convert markdown to html
	if msg.MIME == "text/markdown" && !util.Contain(supportedMIMES, "text/markdown") {
		msg.Body = string(markdown.ToHTML([]byte(msg.Body), nil, nil))
		msg.MIME = "text/html"
	}
	// convert html to text
	if msg.MIME == "text/html" && !util.Contain(supportedMIMES, "text/html") && util.Contain(supportedMIMES, "text/plain") {
		docBuf := strings.NewReader(msg.Body)
		doc, err := goquery.NewDocumentFromReader(docBuf)
		if err != nil {
			return nil, err
		}
		msg.Body = doc.Text()
		msg.MIME = "text/plain"
	}

	if !util.Contain(supportedMIMES, msg.MIME) {
		return nil, &ErrorMIMENoOverlap{
			MessageMIME:   msg.MIME,
			supportedMIME: supportedMIMES,
		}
	}
	return &msg, nil
}
