package ui

import (
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/eternal-flame-AD/yoake/internal/uinext/compo"
	"github.com/eternal-flame-AD/yoake/internal/uinext/ui/page"
	"github.com/eternal-flame-AD/yoake/internal/uinext/webapp"
	"github.com/eternal-flame-AD/yoake/internal/util"
	"github.com/eternal-flame-AD/yoake/internal/version"
	"github.com/labstack/echo/v4"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type App struct {
	app.Compo

	navbar  *compo.Navbar
	sidebar *compo.Sidebar
}

func (a *App) Render() app.UI {
	if a.navbar == nil {
		a.navbar = &compo.Navbar{}
	}
	if a.sidebar == nil {
		a.sidebar = &compo.Sidebar{}
	}
	return app.Div().ID("app").Body(
		a.navbar,
		app.Div().Class("row").Body(
			a.sidebar,
			&page.Router{
				Rules: []page.RouterRule{
					{
						CheckNav: page.RouterMatchPathRegexp(regexp.MustCompile(`^/$`)),
						Page:     &page.Dashboard{},
					},
				},
			},
		),
	)
}

const BasePath = "/uinext"

func Register(g *echo.Group) {
	webapp.Singleton.BasePath = BasePath
	webapp.Singleton.TrimaImgBase = "https://yumechi.jp/img/trima/"

	handler := &app.Handler{
		Name:               "Yoake PMS",
		ShortName:          "夜明け",
		Description:        "Yoake PMS - 夜明け",
		RawHeaders:         util.Join(headTagBootstrap, headTagDayjs, headTagCustom),
		Lang:               "en",
		AutoUpdateInterval: 10 * time.Second,
		BackgroundColor:    "#FEDFE1",
		ThemeColor:         "#FEDFE1",
		LoadingLabel:       "Loading...",
		Version:            version.Version + "-" + version.Date,
		Icon: app.Icon{
			Default: webapp.Singleton.TrimaImgBase + "icon_squeeze.gif",
		},
	}

	app.RouteWithRegexp("^"+BasePath+"/.*", new(App))

	g.Group("/web").GET("*", func(c echo.Context) error {
		handler.ServeHTTP(c.Response(), c.Request())
		return nil
	})
	g.GET("/web/app.wasm", func(c echo.Context) error {
		tryPaths := []string{
			"web/app.wasm",
			"dist/web/app.wasm",
			"app.wasm",
		}
		for _, path := range tryPaths {
			if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
				return c.File(path)
			}
		}
		return c.NoContent(404)
	})
	g.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if util.Contain(staticFiles, strings.ToLower(c.Request().URL.Path)) {
				handler.ServeHTTP(c.Response(), c.Request())
				return nil
			}
			return next(c)
		}
	})
	g.Group(BasePath).Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			next(c)
			if !c.Response().Committed {
				handler.ServeHTTP(c.Response(), c.Request())
				return nil
			}
			return nil
		}
	})
}
