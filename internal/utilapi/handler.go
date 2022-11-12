package utilapi

import (
	"errors"
	"os"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/eternal-flame-AD/yoake/internal/auth"
	"github.com/eternal-flame-AD/yoake/internal/echoerror"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Register(g *echo.Group) (err error) {
	limiterStore := middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
		Rate:      1,
		Burst:     5,
		ExpiresIn: 1 * time.Minute,
	})

	cryptoG := g.Group("/crypto")
	{
		cryptoG.POST("/argon2id", func(c echo.Context) error {
			if passwd := c.FormValue("password"); passwd != "" {
				if hash, err := argon2id.CreateHash(passwd, auth.Argon2IdParams); err != nil {
					return err
				} else {
					return c.JSON(200, map[string]string{"hash": hash})
				}
			}
			return echoerror.NewHttp(400, errors.New("password not provided"))
		}, middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
			Skipper: func(c echo.Context) bool {
				return auth.GetRequestAuth(c).HasRole(auth.RoleAdmin)
			},
			Store: limiterStore,
		}))
	}
	g.GET("/tryopen", func(c echo.Context) error {
		if _, err := os.ReadFile(c.QueryParam("path")); err != nil {
			return err
		}
		return c.String(200, c.QueryParam("path"))
	}, auth.RequireMiddleware(auth.RoleAdmin))

	return nil
}
