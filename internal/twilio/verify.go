package twilio

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/eternal-flame-AD/yoake/config"
	"github.com/twilio/twilio-go/client"
)

func firstUrlValues(val url.Values) map[string]string {
	res := make(map[string]string)
	for k, v := range val {
		res[k] = v[0]
	}
	return res
}

func VerifyMiddleware(prefix string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {

		return func(c echo.Context) error {
			cleanPath := path.Clean(c.Request().URL.Path)
			//log.Printf("cleanPath: %s", cleanPath)
			if cleanPath == prefix || strings.HasPrefix(cleanPath, prefix+"/") {
				if err := TwilioValidate(c.Request()); err != nil {
					c.String(http.StatusOK, "We are sorry. Request Validation Failed. This is not your fault.")
					log.Printf("twilio verify failed: %v", err)
					return err
				}
			}
			return next(c)
		}
	}
}

func TwilioValidate(req *http.Request) error {
	conf := config.Config().Twilio
	signature := req.Header.Get("X-Twilio-Signature")
	if signature == "" {
		if conf.SkipVerify {
			return nil
		}
		return fmt.Errorf("no twilio signature present")
	}
	requestValidator := client.NewRequestValidator(conf.AuthToken)
	if req.Method == "POST" {
		if !requestValidator.Validate(req.URL.String(), firstUrlValues(req.PostForm), signature) {
			return fmt.Errorf("twilio signature verification failed")
		}
	} else if req.Method == "GET" {
		if !requestValidator.Validate(req.URL.String(), nil, signature) {
			return fmt.Errorf("twilio signature verification failed")
		}
	} else {
		return fmt.Errorf("twilio signature verification failed: unsupported method %s", req.Method)
	}

	return nil
}
