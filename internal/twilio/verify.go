package twilio

import (
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/eternal-flame-AD/yoake/config"
	"github.com/eternal-flame-AD/yoake/internal/auth"
	"github.com/eternal-flame-AD/yoake/internal/session"
	"github.com/twilio/twilio-go/client"
)

func firstUrlValues(vals ...url.Values) map[string]string {
	res := make(map[string]string)
	for _, val := range vals {
		for k, v := range val {
			res[k] = v[0]
		}
	}
	return res
}

const verifySessionName = "twilio-verify"

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
			store := c.Get(session.SessionStoreKeyPrefix + "cookie").(sessions.Store)
			if reqAuth := auth.GetRequestAuth(c); reqAuth.Valid && reqAuth.HasRole(auth.RoleAdmin) {
				return next(c)
			}

			bypassOk := false
			sess, _ := store.Get(c.Request(), verifySessionName)
			if ts, ok := sess.Values["verified"].(int64); ok && time.Now().Unix() < ts {
				bypassOk = true
			}

			cleanPath := path.Clean(c.Request().URL.Path)
			//log.Printf("cleanPath: %s", cleanPath)
			if cleanPath == prefix || strings.HasPrefix(cleanPath, prefix+"/") {
				fullReq := c.Request().Clone(c.Request().Context())
				log.Printf("original request URL: %v, scheme=%s, host=%s, user=%s", c.Request().URL, c.Request().URL.Scheme, c.Request().URL.Host, c.Request().URL.User)
				fullReq.URL = baseURL.ResolveReference(c.Request().URL)
				fullReq.URL.User = nil
				if err := TwilioValidate(c, fullReq); err != nil {
					log.Printf("twilio verify failed: %v", err)
					if !bypassOk {
						c.String(http.StatusOK, "We are sorry. Request Validation Failed. This is not your fault.")
						return nil
					}
				} else {
					sess.Values["verified"] = time.Now().Add(5 * time.Minute).Unix()
					sess.Save(c.Request(), c.Response())
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
		query := c.QueryParams()
		form, err := c.FormParams()
		if err != nil {
			return err
		}

		if !requestValidator.Validate(req.URL.String(), firstUrlValues(form), signature) {
			req.URL.RawQuery = ""
			if !requestValidator.Validate(req.URL.String(), firstUrlValues(form, query), signature) {
				return fmt.Errorf("twilio signature verification failed")
			}
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
