package apicall

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func GET(ctx app.Context, path string, dispatch func(app.Context, *http.Response, error)) {
	resp, err := http.Get(path)
	ctx.Dispatch(func(ctx app.Context) {
		dispatch(ctx, resp, err)
	})
}

func GetJSON(ctx app.Context, path string, result interface{}, dispatch func(app.Context, error)) {
	resp, err := http.Get(path)
	if err != nil {
		ctx.Dispatch(func(ctx app.Context) {
			dispatch(ctx, err)
		})
		return
	}
	ctx.Defer(func(app.Context) { resp.Body.Close() })

	dec := json.NewDecoder(resp.Body)

	ctx.Dispatch(func(ctx app.Context) {
		err = dec.Decode(result)
		dispatch(ctx, err)
	})
}

type RequestAuth struct {
	Present bool
	Valid   bool
	Roles   []string
	Expire  time.Time
	Ident   UserIdent
}

type UserIdent struct {
	Username    string `json:"username"`
	PhotoURL    string `json:"photo_url"`
	DisplayName string `json:"display_name"`
}
