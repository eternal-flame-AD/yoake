package compo

import (
	"log"

	"github.com/eternal-flame-AD/yoake/internal/uinext/apicall"
	"github.com/eternal-flame-AD/yoake/internal/uinext/webapp"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type Sidebar struct {
	app.Compo

	routePath string

	Auth apicall.RequestAuth
}

func (c *Sidebar) OnMount(ctx app.Context) {
	ctx.Async(func() {
		apicall.GetJSON(ctx, "/api/auth/auth.json", &c.Auth, func(ctx app.Context, err error) {
			if err != nil {
				log.Printf("Failed to get auth info: %v", err)
				return
			}
			log.Printf("Got auth info: %+v", c.Auth)
		})
	})
}

func (c *Sidebar) OnNav(ctx app.Context) {
	c.routePath = ctx.Page().URL().Path
	c.Update()
}

func (c *Sidebar) Render() app.UI {
	return app.Nav().ID("sidebar").
		Class("col-md-3", "col-lg-2", "d-md-block", "bg-light", "sidebar", "collapse").
		Body(
			app.Div().Class("position-sticky", "pt-3", "sidebar-sticky").Body(
				&SidebarItem{
					Name:    "Dashboard",
					Link:    webapp.Singleton.BasePath + "/",
					CurPath: c.routePath,
					Auth:    c.Auth,
				},
				SidebarHeading("Entertainment"),
				&SidebarItem{
					Name:    "YouTube Playlist",
					Link:    webapp.Singleton.BasePath + "/entertainment/youtube",
					CurPath: c.routePath,
					Auth:    c.Auth,
				},
			),
		)
}

type SidebarItem struct {
	app.Compo

	Name    string
	Link    string
	CurPath string

	state       SidebarItemState
	checkAccess func(auth apicall.RequestAuth) bool

	Auth      apicall.RequestAuth
	HasAccess bool
}

func (c *SidebarItem) OnMount(ctx app.Context) {
	c.Update()
}

func (c *SidebarItem) OnUpdate(ctx app.Context) {
	if c.checkAccess != nil {
		c.HasAccess = c.checkAccess(c.Auth)
	} else {
		c.HasAccess = true
	}

	if c.Link == c.CurPath {
		if c.HasAccess {
			c.state = SidebarItemStateActive
		} else {
			c.state = SidebarItemStateAccessDenied
		}
	} else {
		if c.HasAccess {
			c.state = SidebarItemStateAvailable
		} else {
			c.state = SidebaritemStateAuthRequired
		}
	}
}

type SidebarItemState int

const (
	SidebarItemStateUnknown      SidebarItemState = 0
	SidebarItemStateActive       SidebarItemState = 1
	SidebarItemStateAvailable    SidebarItemState = 2
	SidebaritemStateAuthRequired SidebarItemState = 3
	SidebarItemStateAccessDenied SidebarItemState = 4
)

func (c *SidebarItem) Render() app.UI {
	// TODO: allow theming
	var img app.HTMLImg
	switch c.state {
	case SidebaritemStateAuthRequired:
		img = app.Img().Src(webapp.Singleton.TrimaImgBase + "icon_t_vista_procedure_ineligible.gif")
	case SidebarItemStateAccessDenied:
		img = app.Img().Src(webapp.Singleton.TrimaImgBase + "icon_t_vista_procedure_invalid.gif")
	case SidebarItemStateActive:
		img = app.Img().Src(webapp.Singleton.TrimaImgBase + "icon_t_vista_procedure_optimal.gif")
	case SidebarItemStateAvailable:
		img = app.Img().Src(webapp.Singleton.TrimaImgBase + "icon_t_vista_procedure_valid.gif")
	default:
		img = app.Img().Src(webapp.Singleton.TrimaImgBase + "icon_t_vista_procedure_questionable.gif")
	}
	return app.Ul().Class("nav", "flex-column").Body(
		app.Li().Class("nav-item").Body(
			app.A().Class("nav-link").Href(c.Link).Body(
				app.Span().Class("px-1").Body(img.Style("height", "2.5rem").Style("width", "2.5rem")),
				app.Text(c.Name),
			),
		),
	)
}

func SidebarHeading(title string) app.UI {
	return app.H6().Class("sidebar-heading", "d-flex", "justify-content-between", "align-items-center", "px-3", "mt-4", "mb-1", "text-muted").Body(
		app.Span().Text(title),
	)
}
