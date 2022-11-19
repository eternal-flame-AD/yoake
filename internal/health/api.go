package health

import (
	"fmt"
	"html"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eternal-flame-AD/yoake/internal/auth"
	"github.com/eternal-flame-AD/yoake/internal/comm"
	"github.com/eternal-flame-AD/yoake/internal/comm/model"
	"github.com/eternal-flame-AD/yoake/internal/comm/telegram"
	"github.com/eternal-flame-AD/yoake/internal/db"
	"github.com/eternal-flame-AD/yoake/internal/util"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/labstack/echo/v4"
)

func telegramHandler(database db.DB) telegram.CommandHandler {
	return func(bot *telegram.Bot, role telegram.Role, msg *tgbotapi.Message) error {
		if role != telegram.RoleOwner {
			bot.SendHTML(msg.Chat.ID, "EPERM: You are not authorized to use this command.")
			return nil
		}
		switch msg.Command() {
		case "medtake":
			argStr := strings.TrimSpace(msg.CommandArguments())
			if len(argStr) == 0 || argStr == "help" {
				bot.SendHTML(msg.Chat.ID, "Usage: /medtake <keyname> [dose] [YYYY-MM-DDZHH:mm:ss]")
				return nil
			}
			args := strings.Split(argStr, " ")
			meds, err := DBMedListGet(database)
			if err != nil {
				return err
			}
			for _, med := range meds {
				if med.KeyName() == strings.ToLower(args[0]) {
					dose := med.Dosage
					if len(args) > 1 {
						dose, err = strconv.Atoi(args[1])
						if err != nil {
							return err
						}
					}
					timestamp := time.Now()
					if len(args) > 2 {
						timestamp, err = time.Parse("2006-01-02Z15:04:05", args[2])
						if err != nil {
							return err
						}
					}
					var log ComplianceLog
					log.MedKeyname = med.KeyName()
					log.Actual.Time = timestamp
					log.Actual.Dose = dose
					if err := DBMedComplianceLogSetOne(database, med, &log); err != nil {
						return err
					}
					return bot.SendHTML(msg.Chat.ID, "Success.\n\nDose: %d/%d\nTime Actual: %s\nTime Expected: %s\n\nOffset: %.2f", log.Actual.Dose, log.Expected.Dose,
						log.Actual.Time.Format("2006-01-02 15:04:05"), log.Expected.Time.Format("2006-01-02 15:04:05"), log.DoseOffset)
				}
			}
			return bot.SendHTML(msg.Chat.ID, "keyname %s not found", args[0])
		case "medinfo":
			meds, err := DBMedListGet(database)
			if err != nil {
				return err
			}
			logs, err := DBMedComplianceLogGet(database, util.DateRangeAround(time.Now(), 1))
			if err != nil {
				return err
			}
			argStr := strings.TrimSpace(msg.CommandArguments())
			if argStr == "help" {
				bot.SendHTML(msg.Chat.ID, "Usage: /medinfo [[keyname...]|all]")
				return nil
			}
			all := true
			keynames := strings.Split(msg.CommandArguments(), " ")
			var replies []strings.Builder
			if argStr == "" || argStr == "all" {
				all = argStr == "all"
				replies = make([]strings.Builder, len(meds))
			} else {
				replies = make([]strings.Builder, len(keynames))
			}

			for i, med := range meds {
				index := -1
				if argStr != "" && argStr != "all" {
					for j, keyname := range keynames {
						if med.KeyName() == keyname {
							index = j
							break
						}
					}
					if index == -1 {
						continue
					}
				} else {
					index = i
				}
				name, dir := med.ShortHand()
				fmt.Fprintf(&replies[index], "<b>%s</b> <i>%s</i>\n", html.EscapeString(name), html.EscapeString(dir))
				fmt.Fprintf(&replies[index], "Dosage: %d\n", med.Dosage)
				nextDose := logs.ProjectNextDose(med)
				fmt.Fprintf(&replies[index], "Next Dose: %d%s at %s\n", nextDose.Expected.Dose, html.EscapeString(med.DosageUnit), nextDose.Expected.Time.Format("2006-01-02 15:04:05"))
				stateStr := "unknown"
				if util.Contain(med.Flags, DirectionFlagPRN) && nextDose.DoseOffset >= 0 {
					stateStr = "available"
				} else if nextDose.DoseOffset > 0 {
					stateStr = "DUE"
				} else if util.Contain(med.Flags, DirectionFlagAdLib) {
					stateStr = "available"
				} else if nextDose.DoseOffset < 0 {
					stateStr = "scheduled"
				}
				fmt.Fprintf(&replies[index], "Offset: %.2f (%s)\n", nextDose.DoseOffset, stateStr)
				if !all && nextDose.DoseOffset < 0 {
					replies[index].Reset()
				}
			}
			var out strings.Builder
			for _, reply := range replies {
				if reply.Len() > 0 {
					out.WriteString(reply.String())
					out.WriteString("\n")
				}
			}
			return bot.SendHTML(msg.Chat.ID, out.String())
		}
		return nil
	}

}

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

	if tgBot := comm.GetMethod("telegram"); tgBot != nil {
		bot := tgBot.(*telegram.Bot)
		handler := telegramHandler(database)
		if err := bot.RegisterCommand("medtake", "take 1 med", handler); err != nil {
			log.Printf("failed to register telegram command: %v", err)
		}
		if err := bot.RegisterCommand("medinfo", "current med info", handler); err != nil {
			log.Printf("failed to register telegram command: %v", err)
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
				if err := comm.SendGenericMessage("gotify", &model.GenericMessage{
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
