//go:build linux

package util

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/eternal-flame-AD/go-apparmor/apparmor"
	"github.com/labstack/echo/v4"
)

type AAConMiddlewareEnforcer func(label string, mode string) (exit int, err error)

func AAConMiddleware(enforce AAConMiddlewareEnforcer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			label, mode, err := apparmor.AAGetCon()
			if err != nil {
				log.Printf("failed to get apparmor label: %v", err)
				label = "[ERROR]"
			}
			var sanitizedLabel string
			if idx := strings.Index(label, "//"); idx == -1 {
				sanitizedLabel = "//"
			} else {
				sanitizedLabel = label[idx:]
			}
			c.Response().Header().Set("X-App-Con", fmt.Sprintf("%s (%s)", sanitizedLabel, mode))
			if enforce != nil {
				if exitCode, err := enforce(label, mode); err != nil {
					if exitCode == 0 {
						c.Response().After(func() {
							os.Exit(exitCode)
						})
					}
					return err
				}
			}
			return next(c)
		}
	}
}
