package page

import "github.com/maxence-charriere/go-app/v9/pkg/app"

type Dashboard struct {
	app.Compo
}

func (d *Dashboard) Render() app.UI {
	return BasePage(BasicHeading("Dashboard"),
		app.Div().Class("container").Body(
			app.Div().Class("row").Body(
				app.Div().Class("col").Body(
					&Card{
						Header:    app.Text("Welcome"),
						BodyClass: []string{"text-center"},
						Body: []app.UI{
							app.Blockquote().Class("blockquote").Body(
								app.P().Text("夜明け前が一番暗い"),
								app.P().Text("The night is darkest just before the dawn."),
							),
							app.Hr(),
							app.Div().ID("welcome").Body(
								app.P().Body(
									app.Text("Welcome to yoake.yumechi.jp, Yumechi's "),
									Abbr("PIM", "Personal Information Manager", "initialism", "https://en.wikipedia.org/wiki/Personal_information_manager"),
									app.Text("."),
								),
								app.P().Body(
									app.Text("Built with "),
									Abbr("Echo", "Echo HTTP Framework", "", "https://echo.labstack.com/"),
									app.Text(", "),
									Abbr("Bootstrap", "Bootstrap CSS Framework", "", "https://getbootstrap.com/"),
									app.Text(", and "),
									Abbr("go-app", "Go PWA Framework", "", "https://go-app.dev/"),
									app.Text("."),
								),
							),
						},
					},
				),
			),
		),
	)
}
