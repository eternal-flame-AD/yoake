package webroot

import (
	"encoding/json"
	"log"
	"os"

	"github.com/eternal-flame-AD/yoake/config"
	"github.com/eternal-flame-AD/yoake/internal/auth"
	"github.com/eternal-flame-AD/yoake/internal/servetpl"
	"github.com/eternal-flame-AD/yoake/internal/session"
	"github.com/eternal-flame-AD/yoake/internal/twilio"
	"github.com/eternal-flame-AD/yoake/server"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Init(hostname string) {
	e := echo.New()

	webroot := config.Config().WebRoot
	if webroot.Root == "" {
		log.Panicf("webroot not set, use . to override")
	}
	if webroot.SessionKey == "" {
		log.Panicf("webroot session key not set")
	}

	sessionCookie := sessions.NewCookieStore([]byte(webroot.SessionKey))
	fsCookie := sessions.NewFilesystemStore(webroot.SessionDir, []byte(webroot.SessionKey))

	if webroot.Secure != nil {
		e.Use(middleware.SecureWithConfig(*webroot.Secure))
	}
	if webroot.Log != nil {
		filters := logCompileFilters(webroot.Log.Filter)
		logOut := log.New(os.Stdout, "webroot: ", log.Ldate|log.Ltime)
		lc := loggerConfig
		lc.LogValuesFunc = func(c echo.Context, values middleware.RequestLoggerValues) (err error) {
			entry := processLoggerValues(c, values)
			if logFilterCategories(c, filters) {
				var j []byte
				if webroot.Log.Indent {
					j, err = json.MarshalIndent(entry, "", "  ")
				} else {
					j, err = json.Marshal(entry)
				}
				if err != nil {
					return err
				}
				logOut.Println(string(j))
			}
			return nil
		}
		e.Use(middleware.RequestLoggerWithConfig(lc))
	}

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer context.Clear(c.Request())
			c.Set(session.SessionStoreKeyPrefix+"cookie", (sessions.Store)(sessionCookie))
			c.Set(session.SessionStoreKeyPrefix+"fs", (sessions.Store)(fsCookie))
			return next(c)
		}
	},
		middleware.Gzip(),
		logMiddleware("twilio", twilio.VerifyMiddleware("/twilio")),
		auth.Middleware(sessionCookie),
		middleware.Rewrite(map[string]string{"*/": "$1/index.html"}),
		logMiddleware("template", servetpl.ServeTemplateDir(webroot.Root)),
		logMiddleware("static", middleware.Static(webroot.Root)))

	server.RegisterHostname(hostname, &server.Host{Echo: e})
}
