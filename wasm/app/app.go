//go:build wasm

package main

import (
	"github.com/eternal-flame-AD/yoake/internal/uinext/ui"
	"github.com/labstack/echo/v4"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func main() {
	ui.Register(echo.New().Group(""))
	app.RunWhenOnBrowser()
}
