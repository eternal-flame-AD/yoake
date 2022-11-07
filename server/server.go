package server

import (
	"strings"

	"github.com/labstack/echo/v4"
)

type (
	Host struct {
		Echo *echo.Echo
	}
)

var Server = echo.New()
var hosts = map[string]*Host{}

func init() {
	Server.Any("/*", func(c echo.Context) (err error) {
		req := c.Request()
		res := c.Response()
		host := hosts[strings.ToLower(req.Host)]

		if host == nil {
			host = hosts[""]
			if host == nil {
				return echo.ErrNotFound
			}
		}

		host.Echo.ServeHTTP(res, req)
		return
	})
}

func RegisterHostname(hostname string, h *Host) {
	hosts[hostname] = h
}
