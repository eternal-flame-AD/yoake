package funcmap

import (
	"github.com/eternal-flame-AD/yoake/internal/auth"
	"github.com/labstack/echo/v4"
)

func AuthGet(c echo.Context) auth.RequestAuth {
	a := auth.GetRequestAuth(c)
	if !a.Valid {
		return auth.RequestAuth{}
	} else {
		return a
	}
}
