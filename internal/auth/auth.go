package auth

import (
	"log"
	"net/http"
	"time"

	"github.com/eternal-flame-AD/yoake/config"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

const AuthSessionName = "auth_session"

type RequestAuth struct {
	Present bool
	Valid   bool
	Roles   []string
	Expire  time.Time
}

func (a RequestAuth) HasRole(role Role) bool {
	if !a.Valid {
		return false
	}
	for _, r := range a.Roles {
		if r == string(role) {
			return true
		}
	}
	return false
}

func Middleware(store sessions.Store) echo.MiddlewareFunc {
	yubiAuthLazyInit()
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			sess, _ := store.Get(c.Request(), AuthSessionName)
			sess.Options = &sessions.Options{
				Path:     "/",
				HttpOnly: true,
			}

			var auth RequestAuth
			if expireTs, ok := sess.Values["expire"].(string); ok {
				auth.Present = true
				if expireTime, err := time.Parse(time.RFC3339, expireTs); err != nil {
					log.Printf("invalid expireTime: %v", expireTs)
				} else {
					auth.Expire = expireTime
					if time.Now().Before(expireTime) {
						auth.Valid = true
					}
				}
			}

			if existingRoles, ok := sess.Values["roles"].([]string); !ok {
				sess.Values["roles"] = []string{}
				sess.Save(c.Request(), c.Response())
			} else if auth.Valid {
				auth.Roles = existingRoles
			}

			c.Set("auth_"+AuthSessionName, auth)
			c.Set("auth_store", store)

			return next(c)
		}
	}
}

func issueSession(c echo.Context, period time.Duration, baseRole string) error {
	sess, _ := c.Get("auth_store").(sessions.Store).Get(c.Request(), AuthSessionName)
	sess.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
	}
	if period == 0 {
		period = time.Duration(config.Config().Auth.ValidMinutes) * time.Minute
	}
	if period < 0 {
		sess.Values["expire"] = (time.Time{}).Format(time.RFC3339)
		sess.Values["roles"] = ""
	} else {
		roles := []string{baseRole}
		if baseRole == string(RoleAdmin) {
			roles = append(roles, string(RoleUser))
		}

		sess.Values["expire"] = time.Now().Add(period).Format(time.RFC3339)
		sess.Values["roles"] = roles
		log.Printf("Issued session for %v, roles: %v", period, roles)
	}
	return sess.Save(c.Request(), c.Response())
}

func Login(c echo.Context) (err error) {
	if c.Request().Method == http.MethodDelete {
		return issueSession(c, -1, "")
	}
	switch c.FormValue("type") {
	case "userpass":
		return echo.NewHTTPError(http.StatusNotImplemented, "userpass login not implemented")
		// username, password := c.FormValue("username"), c.FormValue("password")
	case "yubikey":
		if yubiAuth == nil {
			return echo.NewHTTPError(http.StatusNotImplemented, "Yubikey authentication not configured")
		}
		otp := c.FormValue("response")
		if yr, ok, err := yubiAuth.Verify(otp); err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Yubikey authentication failed: "+err.Error())
		} else if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "Yubikey authentication failed")
		} else {
			// sessionUseCounter := yr.GetResultParameter("sessionuse")
			// sessionCounter := yr.GetResultParameter("sessioncounter")
			keyPublicId := yr.GetResultParameter("otp")[:12]
			for _, authorizedKey := range config.Config().Auth.Method.Yubikey.Keys {
				if authorizedKey.PublicId[:12] == keyPublicId {
					issueSession(c, 0, authorizedKey.Role)
					return nil
				}
			}
			return echo.NewHTTPError(http.StatusUnauthorized, "Yubikey authentication failed: key "+keyPublicId+" not authorized")
		}
	default:
		return echo.NewHTTPError(400, "invalid auth type")
	}

}
func GetRequestAuth(c echo.Context) RequestAuth {
	if a, ok := c.Get("auth_" + AuthSessionName).(RequestAuth); ok {
		return a
	} else {
		return RequestAuth{Present: false, Valid: false}
	}
}
