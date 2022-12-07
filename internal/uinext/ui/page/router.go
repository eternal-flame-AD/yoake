package page

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/eternal-flame-AD/yoake/internal/uinext/webapp"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type Router struct {
	app.Compo
	matchedPage app.UI

	Rules    []RouterRule
	NotFound app.UI
}

func (r *Router) OnNav(ctx app.Context) {
	defer r.Update()
	for _, rule := range r.Rules {
		urlCopy := *ctx.Page().URL()
		if !rule.NoStripPrefix {
			bp := webapp.Singleton.BasePath
			if bp != "" {
				urlCopy.Path = strings.TrimPrefix(urlCopy.Path, bp)
				urlCopy.RawPath = strings.TrimPrefix(urlCopy.RawPath, bp)
			}
		}
		if rule.CheckNav(&urlCopy) {
			r.matchedPage = rule.Page
			return
		}
	}
	r.matchedPage = nil
}

func (r *Router) Render() app.UI {
	if r.matchedPage != nil {
		return r.matchedPage
	}
	if r.NotFound != nil {
		return r.NotFound
	}
	return BasePage(BasicHeading("404 Not Found"), app.P().Text("The page you are looking for does not exist."))
}

type RouterRule struct {
	NoStripPrefix bool
	CheckNav      func(*url.URL) bool
	Page          app.UI
}

func RouterMatchPathRegexp(regex *regexp.Regexp) func(*url.URL) bool {
	return func(u *url.URL) bool {
		return regex.MatchString(u.Path)
	}
}
