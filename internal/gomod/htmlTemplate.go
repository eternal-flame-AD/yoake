package gomod

import "text/template"

type goGetHtmlTemplateCtx struct {
	ModulePath string
	ProxyBase  string
}

var goGetHtmlTemplate = template.Must(template.New("tpl").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="go-import" content="{{.ModulePath}} mod {{.ProxyBase}}">
</head>
<body>
go get {{.ModulePath}}
</body>
</html>`))
