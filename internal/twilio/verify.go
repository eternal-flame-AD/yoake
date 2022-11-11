package twilio

import (
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/eternal-flame-AD/yoake/config"
	"github.com/eternal-flame-AD/yoake/internal/auth"
	"github.com/twilio/twilio-go/client"
)

func firstUrlValues(val url.Values) map[string]string {
	res := make(map[string]string)
	for k, v := range val {
		res[k] = v[0]
	}
	return res
}

func VerifyMiddleware(prefix string, baseurlS string) echo.MiddlewareFunc {
	baseURL, err := url.Parse(baseurlS)
	if err != nil {
		log.Fatalf("invalid twilio baseurl: %v", baseurlS)
	}
	log.Printf("twilio baseurl is %v", baseURL)
	var basicAuth echo.MiddlewareFunc
	if userpass := baseURL.User.String(); userpass != "" {
		basicAuth = middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
			ui := url.UserPassword(username, password)
			return subtle.ConstantTimeCompare([]byte(ui.String()), []byte(userpass)) == 1, nil
		})
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		verifySignature := func(c echo.Context) error {
			if reqAuth := auth.GetRequestAuth(c); reqAuth.Valid && reqAuth.HasRole(auth.RoleAdmin) {
				return next(c)
			}

			cleanPath := path.Clean(c.Request().URL.Path)
			//log.Printf("cleanPath: %s", cleanPath)
			if cleanPath == prefix || strings.HasPrefix(cleanPath, prefix+"/") {
				fullReq := c.Request().Clone(c.Request().Context())
				log.Printf("original request URL: %v, scheme=%s, host=%s, user=%s", c.Request().URL, c.Request().URL.Scheme, c.Request().URL.Host, c.Request().URL.User)
				fullReq.URL = baseURL.ResolveReference(c.Request().URL)
				fullReq.URL.User = nil
				if err := TwilioValidate(c, fullReq); err != nil {
					c.String(http.StatusOK, "We are sorry. Request Validation Failed. This is not your fault.")
					log.Printf("twilio verify failed: %v", err)
					return err
				}
			}
			return next(c)
		}
		if basicAuth != nil {
			return basicAuth(verifySignature)
		}
		return verifySignature
	}
}

func TwilioValidate(c echo.Context, req *http.Request) error {
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
		form, err := c.FormParams()
		if err != nil {
			return err
		}
		if !requestValidator.Validate(req.URL.String(), firstUrlValues(form), signature) {
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
