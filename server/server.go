package server

import (
	"errors"
	"log"
	"strings"

	"github.com/eternal-flame-AD/go-apparmor/apparmor"
	"github.com/eternal-flame-AD/go-apparmor/apparmor/magic"
	"github.com/eternal-flame-AD/yoake/config"
	"github.com/eternal-flame-AD/yoake/internal/util"
	"github.com/labstack/echo/v4"
)

type (
	Host struct {
		Echo *echo.Echo
	}
)

var hosts = map[string]*Host{}

func New() *echo.Echo {
	var Server = echo.New()
	hatServe := config.Config().Listen.AppArmor.Serve
	if hatServe != "" {
		store, err := magic.NewKeyring(nil)
		if err != nil {
			log.Panicf("failed to initialize magic token store: %v", err)
		}
		if magic, err := magic.Generate(nil); err != nil {
			log.Panicf("failed to generate apparmor magic token: %v", err)
		} else {
			if err := store.Set(magic); err != nil {
				log.Panicf("failed to store apparmor magic token: %v", err)
			}
		}
		hatMagic := func() uint64 {
			magic, err := store.Get()
			if err != nil {
				log.Panicf("failed to get magic token: %v", err)
			}
			return magic
		}
		Server.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				var err error
				if errAppArmor := apparmor.WithHat(hatServe, hatMagic, func() {
					err = next(c)
				}); errAppArmor != nil {
					c.Logger().Errorf("apparmor error: %v", errAppArmor)
					return errors.New("apparmor process transition error")
				}
				return err
			}
		})
		aaEnforcer := util.AAConMiddleware(func(label string, mode string) (exit int, err error) {
			if !strings.HasSuffix(label, "//"+hatServe) {
				return 1, errors.New("apparmor process transition error")
			}
			return 0, nil
		})

		Server.Pre(aaEnforcer)
		Server.Use(aaEnforcer)
		for _, h := range hosts {
			h.Echo.Pre(aaEnforcer)
			h.Echo.Use(aaEnforcer)
		}
	}

	Server.Any("/*", func(c echo.Context) (err error) {
		req := c.Request()
		res := c.Response()
		host := hosts[strings.ToLower(req.Host)]

		if host == nil {
			host = hosts[""]
			if host == nil {
				err = echo.ErrNotFound
				return
			}
		}

		host.Echo.ServeHTTP(res, req)

		return
	})

	return Server
}

func RegisterHostname(hostname string, h *Host) {
	hosts[hostname] = h
}
