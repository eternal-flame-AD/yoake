package server

import (
	"strings"

	"github.com/eternal-flame-AD/yoake/config"
	"github.com/eternal-flame-AD/yoake/internal/apparmor"
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
	hatChanged := false
	Server.Any("/*", func(c echo.Context) (err error) {
		if !hatChanged {
			appArmor := config.Config().Listen.AppArmor
			if appArmor.Serve != "" {
				if key, err := apparmor.GetMagicToken(); err != nil {
					return err
				} else {
					if err := apparmor.ChangeHat(appArmor.Serve, key); err != nil {
						return err
					}
				}
			}
			hatChanged = true
		}
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
