package compo

import (
	"fmt"

	"github.com/eternal-flame-AD/yoake/internal/uinext/webapp"
	"github.com/eternal-flame-AD/yoake/internal/version"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// source: includes/navbar.tpl.html

type Navbar struct {
	app.Compo

	LoginUsername string
}

func (n *Navbar) renderBrand() app.UI {
	return app.A().Class("navbar-brand", "col-md-3", "col-lg-2", "me-0", "px-3", "fs-6").
		Href(webapp.Singleton.BasePath+"/").
		Body(
			app.Text("夜明け"),
			app.Small().Class("fw-lighter", "text-muted", "px-2").Text(fmt.Sprintf("%s - %s", version.Version, version.Date)),
		)
}

func (n *Navbar) renderNavBtn() app.UI {
	return app.Button().Class("navbar-toggler", "position-absolute", "d-md-none", "collapsed").
		Type("button").
		Aria("aria-label", "Toggle navigation").
		Attr("data-bs-toggle", "collapse").
		Attr("data-bs-target", "#sidebar").
		Attr("aria-controls", "sidebarMenu").
		Attr("aria-expanded", "false").
		Body(
			app.Span().Class("navbar-toggler-icon"),
		)
}

func (n *Navbar) renderAuthUsername() app.UI {
	return app.Div().Class("navbar-nav").Body(
		app.Div().Class("nav-item", "text-nowrap", "px-3").Body(
			app.Text(n.LoginUsername),
		),
	)
}

func (n *Navbar) Render() app.UI {
	return app.Nav().Class("navbar", "sticky-top", "flex-md-nowrap", "p-0").Body(
		n.renderBrand(),
		n.renderNavBtn(),
		n.renderAuthUsername(),
	)
}
