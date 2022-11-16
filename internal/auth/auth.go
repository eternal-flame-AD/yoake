package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/eternal-flame-AD/yoake/config"
	"github.com/eternal-flame-AD/yoake/internal/echoerror"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const AuthSessionName = "auth_session"

var dummyHash string

func init() {
	var dummyPassword [16]byte
	_, err := rand.Read(dummyPassword[:])
	if err != nil {
		panic(err)
	}
	dummyHash, err = argon2id.CreateHash(string(dummyPassword[:]), Argon2IdParams)
	if err != nil {
		panic(err)
	}
}

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

type RoleInsufficientError struct {
	RoleRequired   Role
	RolesAvailable []string
}

func (e RoleInsufficientError) Error() string {
	return fmt.Sprintf("role insufficient: required %v, you have %v", e.RoleRequired, e.RolesAvailable)
}

func (e RoleInsufficientError) Code() int {
	if len(e.RolesAvailable) == 0 {
		return http.StatusUnauthorized
	}
	return http.StatusForbidden
}

func (a RequestAuth) RequireRole(role Role) error {
	if config := config.Config(); config.Auth.DevMode.GrantAll && !config.Listen.Ssl.Use {
		log.Printf("dev mode: role %v granted without checking", role)
		return nil
	}
	if a.HasRole(role) {
		return nil
	}
	return RoleInsufficientError{RoleRequired: role, RolesAvailable: a.Roles}
}

func RequireMiddleware(role Role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := GetRequestAuth(c)
			if err := auth.RequireRole(role); err != nil {
				return err
			}
			return next(c)
		}
	}
}

func Middleware(store sessions.Store) echo.MiddlewareFunc {
	yubiAuthLazyInit()
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			sess, _ := store.Get(c.Request(), AuthSessionName)
			sess.Options = &sessions.Options{
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
				Secure:   config.Config().Listen.Ssl.Use,
				MaxAge:   config.Config().Auth.ValidMinutes * 60 * 5,
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

func issueSession(c echo.Context, period time.Duration, roles []string) error {
	sess, _ := c.Get("auth_store").(sessions.Store).Get(c.Request(), AuthSessionName)
	sess.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   config.Config().Listen.Ssl.Use,
		MaxAge:   config.Config().Auth.ValidMinutes * 60 * 5,
	}
	if period == 0 {
		period = time.Duration(config.Config().Auth.ValidMinutes) * time.Minute
	}
	if period < 0 {
		sess.Values["expire"] = (time.Time{}).Format(time.RFC3339)
		sess.Values["roles"] = ""
	} else {
		sess.Values["expire"] = time.Now().Add(period).Format(time.RFC3339)
		sess.Values["roles"] = roles
		log.Printf("Issued session for %v, roles: %v", period, roles)
	}
	return sess.Save(c.Request(), c.Response())
}

type LoginForm struct {
	Username    string `json:"username" form:"username"`
	Password    string `json:"password" form:"password"`
	OtpResponse string `json:"otp_response" form:"otp_response"`
}

var errInvalidUserPass = echoerror.NewHttp(http.StatusUnauthorized, errors.New("invalid username or password"))

func Register(g *echo.Group) (err error) {
	g.GET("/auth.json", func(c echo.Context) error {
		return c.JSON(http.StatusOK, GetRequestAuth(c))
	})

	loginRateLimiterStore := middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
		ExpiresIn: 300 * time.Second,
		Rate:      2,
		Burst:     4,
	})
	loginRateLimiter := middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: loginRateLimiterStore,
	})
	g.POST("/login", func(c echo.Context) error {
		var form LoginForm
		if err := c.Bind(&form); err != nil {
			return err
		}
		var verifiedOtpPubId string
		if form.OtpResponse != "" {
			form.OtpResponse = strings.TrimSpace(form.OtpResponse)
			if yubiAuth == nil {
				return echo.NewHTTPError(http.StatusNotImplemented, "Yubikey authentication not configured")
			}
			if yr, ok, err := yubiAuth.Verify(form.OtpResponse); err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Yubikey authentication failed: "+err.Error())
			} else if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Yubikey authentication failed")
			} else {
				// sessionUseCounter := yr.GetResultParameter("sessionuse")
				// sessionCounter := yr.GetResultParameter("sessioncounter")
				keyPublicId := yr.GetResultParameter("otp")[:12]
				verifiedOtpPubId = keyPublicId
			}
		}

		if form.Username == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "username required")
		}
		if form.Password == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "password required")
		}
		if user, ok := config.Config().Auth.Users[form.Username]; ok {
			if match, _ := argon2id.ComparePasswordAndHash(form.Password, user.Password); match {
				if len(user.PublicKeyId) > 0 {
					if verifiedOtpPubId == "" {
						return echo.NewHTTPError(http.StatusUnauthorized, "otp required")
					}
					found := 0
					for _, pubId := range user.PublicKeyId {
						found += subtle.ConstantTimeCompare([]byte(pubId[:12]), []byte(verifiedOtpPubId))
					}
					if found == 0 {
						return echo.NewHTTPError(http.StatusUnauthorized, "incorrect key used")
					}
				} else if verifiedOtpPubId != "" {
					return echo.NewHTTPError(http.StatusBadRequest, "otp not required but you provided one, this may be an configuration error")
				}

				issueSession(c, 0, user.Roles)
				c.JSON(http.StatusOK, map[string]interface{}{"message": "ok", "ok": true})
				return nil
			} else {
				return errInvalidUserPass
			}
		}
		argon2id.ComparePasswordAndHash(form.Password, dummyHash)
		return errInvalidUserPass
	}, loginRateLimiter)
	g.DELETE("/login", func(c echo.Context) error {
		return issueSession(c, -1, nil)
	})
	return nil
}

func GetRequestAuth(c echo.Context) RequestAuth {
	if a, ok := c.Get("auth_" + AuthSessionName).(RequestAuth); ok {
		return a
	} else {
		return RequestAuth{Present: false, Valid: false}
	}
}
