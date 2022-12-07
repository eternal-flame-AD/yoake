package page

import "github.com/maxence-charriere/go-app/v9/pkg/app"

func BasePage(elems ...app.UI) app.UI {
	return app.Main().Class("col-md-9", "ms-sm-auto", "col-lg-10", "px-md-4").Body(elems...)
}

func BasicHeading(name string) app.UI {
	return app.Div().Body(
		app.H1().Class("page-header").Text(name),
		app.Hr(),
	)
}

type Card struct {
	app.Compo

	HeaderClass []string
	Header      app.UI

	BodyClass []string
	Body      []app.UI
}

func (c *Card) Render() app.UI {
	return app.Div().Class("card", "border").Body(
		app.Div().Class(append(c.HeaderClass, "card-header")...).Body(c.Header),
		app.Div().Class(append(c.BodyClass, "card-body")...).Body(c.Body...),
	)
}

func Abbr(abbr string, title string, class string, href string) app.UI {
	abbrEle := app.Abbr()
	if title != "" {
		abbrEle = abbrEle.Title(title)
	}
	if class != "" {
		abbrEle = abbrEle.Class(class)
	}
	if href != "" {
		return app.A().Target("_blank").Rel("noopener noreferrer").Href(href).Body(abbrEle.Text(abbr))
	}
	return abbrEle.Text(abbr)
}
