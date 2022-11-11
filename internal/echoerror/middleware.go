package echoerror

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"
)

var htmlErrorTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Error</title>
</head>
<body>
	<h1>{{.Code}} {{.CodeText}}</h1>
	<hr>
	<p>Refer to this message: {{.Message}}</p>
`))

type htmlErrorTemplateCtx struct {
	Code     int
	CodeText string
	Message  string
}

type ErrorWriter func(c echo.Context, err error) error

var (
	JSONWriter = func(c echo.Context, err error) error {
		if httpError, ok := err.(HTTPError); ok {
			jsonStr, err := json.Marshal(map[string]interface{}{
				"code":    httpError.Code(),
				"ok":      false,
				"message": httpError.Error(),
				"error":   err,
			})
			if err != nil {
				jsonStr, err = json.Marshal(map[string]interface{}{
					"code":    httpError.Code(),
					"ok":      false,
					"message": httpError.Error(),
					"error":   nil,
				})
				if err != nil {
					return err
				}
			}
			return c.JSONBlob(httpError.Code(), jsonStr)
		}
		return c.JSON(500, map[string]interface{}{
			"code":    500,
			"ok":      false,
			"message": err.Error(),
		})
	}

	HTMLWriter = func(c echo.Context, err error) error {
		errContext := htmlErrorTemplateCtx{
			Code:    500,
			Message: "Internal Server Error",
		}
		if httpError, ok := err.(HTTPError); ok {
			errContext.Code = httpError.Code()
			errContext.Message = httpError.Error()
		} else {
			errContext.Message = err.Error()
		}
		errContext.CodeText = http.StatusText(errContext.Code)
		c.Response().Status = errContext.Code
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		return htmlErrorTemplate.Execute(c.Response().Writer, errContext)
	}
)

func Middleware(errorWriter ErrorWriter) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				if errEcho, ok := err.(*echo.HTTPError); ok {
					err = NewHttp(errEcho.Code, fmt.Errorf("%s", errEcho.Message))
				}
				if err := errorWriter(c, err); err != nil {
					return err
				}
			}
			return nil
		}
	}
}
