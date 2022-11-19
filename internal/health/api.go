package health

import (
	"log"
	"sync"
	"time"

	"github.com/eternal-flame-AD/yoake/internal/auth"
	"github.com/eternal-flame-AD/yoake/internal/comm"
	"github.com/eternal-flame-AD/yoake/internal/comm/model"
	"github.com/eternal-flame-AD/yoake/internal/db"
	"github.com/eternal-flame-AD/yoake/internal/util"
	"github.com/labstack/echo/v4"
)

func Register(g *echo.Group, database db.DB, comm *comm.Communicator) {
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
			directions.GET("", RESTMedGetDirections(database))
			directions.POST("", RESTMedPostDirections(database, writeMutex))
			directions.DELETE("/:name", RESTMedDeleteDirections(database, writeMutex))
		}

		compliance := megsG.Group("/compliance", auth.RequireMiddleware(auth.RoleAdmin))
		{
			complianceByMed := compliance.Group("/med/:med")
			{
				complianceByMed.GET("/log", RESTComplianceLogGet(database))
				complianceByMed.GET("/project", RESTComplianceLogProjectMed(database))

			}

			compliance.GET("/log", RESTComplianceLogGet(database))

			compliance.POST("/log", RESTComplianceLogPost(database, writeMutex))

			compliance.POST("/recalc", RESTRecalcMedComplianceLog(database, writeMutex))
		}
	}

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		notified := make(map[string]time.Time)
		for {
			func() {
				txn := database.NewTransaction(false)
				defer txn.Discard()

				existingNotified := make(map[string]time.Time)
				err := db.GetJSON(txn, []byte("health_meds_compliance_notified_meds"), &existingNotified)

				if err != nil && !db.IsNotFound(err) {
					log.Println("Error getting notified meds: ", err)
					return
				}
				for k, v := range existingNotified {
					o := notified[k]
					if o.Before(v) {
						notified[k] = v
					}
				}
				txn.Discard()
				txn = database.NewTransaction(true)
				err = db.SetJSON(txn, []byte("health_meds_compliance_notified_meds"), notified)
				if err != nil {
					log.Println("Error setting notified meds: ", err)
					return
				} else if err := txn.Commit(); err != nil {
					log.Println("Error committing notified meds: ", err)
					return
				}
			}()

			meds, err := DBMedListGet(database)
			if err != nil {
				log.Println("Error getting med list:", err)
				continue
			}

			logs, err := DBMedComplianceLogGet(database, util.DateRangeAround(time.Now(), 1))
			if err != nil {
				log.Println("Error getting med compliance log:", err)
				continue
			}

			var notifications []CommCtx

			hasNew := false
			for _, med := range meds {
				nextDose := logs.ProjectNextDose(med)
				if nextDose.Expected.Time.Before(time.Now()) {
					if lastNotified, ok := notified[med.KeyName()]; !ok ||
						lastNotified.Before(nextDose.Expected.Time) ||
						lastNotified.Add(4*time.Hour).Before(time.Now()) {
						{
							if !util.Contain(med.Flags, DirectionFlagPRN) {
								hasNew = true
							}
						}
						notifications = append(notifications, CommCtx{
							Med:  med,
							Dose: nextDose,
						})
					}
				}
			}
			if hasNew {
				if err := comm.SendGenericMessage("gotify", model.GenericMessage{
					Subject: "Medications Due",
					Body:    commTemplate,
					MIME:    "text/markdown+html/template",
					Context: notifications,
				}, true); err != nil {
					log.Println("Error sending med compliance notification:", err)
				}
				for _, n := range notifications {
					notified[n.Med.KeyName()] = time.Now()
				}
			}

			<-ticker.C
		}
	}()
}
