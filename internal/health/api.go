package health

import (
	"sync"

	"github.com/eternal-flame-AD/yoake/internal/auth"
	"github.com/eternal-flame-AD/yoake/internal/comm"
	"github.com/eternal-flame-AD/yoake/internal/db"
	"github.com/labstack/echo/v4"
)

func Register(g *echo.Group, db db.DB, comm *comm.Communicator) {
	megsG := g.Group("/meds")
	{
		shortHands := megsG.Group("/shorthand")
		{
			shortHands.GET("/parse", RESTParseShorthand())
			shortHands.POST("/parse", RESTParseShorthand())

			shortHands.POST("/format", RESTFormatShorthand())
		}

		writeMutex := new(sync.Mutex)
		directions := megsG.Group("/directions", auth.RequireMiddleware(auth.RoleAdmin))
		{
			directions.GET("", RESTMedGetDirections(db))
			directions.POST("", RESTMedPostDirections(db, writeMutex))
			directions.DELETE("/:name", RESTMedDeleteDirections(db, writeMutex))
		}

		compliance := megsG.Group("/compliance", auth.RequireMiddleware(auth.RoleAdmin))
		{
			complianceByMed := compliance.Group("/med/:med")
			{
				complianceByMed.GET("/log", RESTComplianceLogGet(db))
				complianceByMed.GET("/project", RESTComplianceLogProjectMed(db))
			}

			compliance.GET("/log", RESTComplianceLogGet(db))

			compliance.POST("/log", RESTComplianceLogPost(db, writeMutex))
		}
	}
}
