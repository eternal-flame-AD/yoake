package session

import (
	"log"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

const SessionStoreKeyPrefix = "_session_store_"

const maxAge = 86400 * 90

type Provider func(driver string, name string, path string) map[any]any

func ManagedSession(c echo.Context) (p Provider, close func()) {
	checkedOutSessions := make(map[string]*sessions.Session)
	return func(driver string, name string, path string) map[any]any {
			store := c.Get(SessionStoreKeyPrefix + driver).(sessions.Store)

			if s, ok := checkedOutSessions[name]; ok {
				return s.Values
			}
			s, _ := store.Get(c.Request(), name)

			s.Options = &sessions.Options{
				Path:   path,
				MaxAge: maxAge,
			}
			checkedOutSessions[name] = s
			return s.Values
		}, func() {
			for name, s := range checkedOutSessions {
				if err := s.Save(c.Request(), c.Response()); err != nil {
					log.Printf("error saving session %s: %v", name, err)
				}
			}
		}
}
