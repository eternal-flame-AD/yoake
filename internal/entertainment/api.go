package entertainment

import (
	"github.com/eternal-flame-AD/yoake/internal/db"
	"github.com/labstack/echo/v4"
)

func Register(g *echo.Group, database db.DB) {
	youtube := g.Group("/youtube")
	registerYoutube(youtube, database)
}
