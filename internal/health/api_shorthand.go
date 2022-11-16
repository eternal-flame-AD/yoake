package health

import (
	"fmt"
	"strings"

	"github.com/eternal-flame-AD/yoake/internal/echoerror"
	"github.com/labstack/echo/v4"
)

func RESTParseShorthand() func(c echo.Context) error {
	return func(c echo.Context) error {
		defer func() {
			if err := recover(); err != nil {
				c.Error(echoerror.NewHttp(500, fmt.Errorf("internal error: %v", err)))
			}
		}()
		var inputStr string
		if c.Request().Method == "GET" {
			inputStr = c.QueryParam("shorthand")
		} else if c.Request().Method == "POST" {
			inputStr = c.FormValue("shorthand")
		} else {
			return echoerror.NewHttp(405, fmt.Errorf("unsupported method"))
		}
		inputStr = strings.TrimSpace(inputStr)
		parsed, err := ParseShorthand(inputStr)
		if err != nil {
			return echoerror.NewHttp(400, err)
		}
		return c.JSON(200, parsed)
	}
}

func RESTFormatShorthand() func(c echo.Context) error {
	return func(c echo.Context) error {
		defer func() {
			if err := recover(); err != nil {
				c.Error(echoerror.NewHttp(500, fmt.Errorf("internal error: %v", err)))
			}
		}()
		var input Direction
		if err := c.Bind(&input); err != nil {
			return echoerror.NewHttp(400, err)
		}
		name, formatted := input.ShortHand()
		return c.JSON(200, map[string]string{
			"name":         name,
			"direction":    formatted,
			"__disclaimer": DirectionDisclaimer,
		})
	}
}
