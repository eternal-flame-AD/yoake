package health

import (
	"fmt"
	"sync"
	"time"

	"github.com/eternal-flame-AD/yoake/internal/db"
	"github.com/eternal-flame-AD/yoake/internal/echoerror"
	"github.com/eternal-flame-AD/yoake/internal/util"
	"github.com/labstack/echo/v4"
)

const yearAbsZero = 2000

func RESTComplianceLogGet(database db.DB) func(c echo.Context) error {
	return func(c echo.Context) error {
		filterKeyname := c.Param("med")
		from := c.QueryParam("from")
		to := c.QueryParam("to")
		if to == "" {
			to = time.Now().Format("2006-01-02")
		}
		if from == "" {
			from = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
		}
		fromTime, err := time.Parse("2006-01-02", from)
		if err != nil {
			return echoerror.NewHttp(400, err)
		}
		toTime, err := time.Parse("2006-01-02", to)
		if err != nil {
			return echoerror.NewHttp(400, err)
		}
		period := util.NewDateRange(fromTime, toTime)
		if days := period.Days(); days > 180 {
			return echoerror.NewHttp(400, fmt.Errorf("invalid date range: %v", period))
		}
		logs, err := DBMedComplianceLogGet(database, period)
		if db.IsNotFound(err) || logs == nil {
			return c.JSON(200, ComplianceLogList{})
		} else if err != nil {
			return echoerror.NewHttp(500, err)
		}
		if filterKeyname != "" {
			filtered := make(ComplianceLogList, 0, len(logs))
			for _, log := range logs {
				if log.MedKeyname == filterKeyname {
					filtered = append(filtered, log)
				}
			}
			logs = filtered
		}

		return c.JSON(200, logs)
	}
}

func RESTComplianceLogPost(db db.DB, writeMutex *sync.Mutex) echo.HandlerFunc {
	return func(c echo.Context) error {
		var input ComplianceLog
		if err := c.Bind(&input); err != nil {
			return echoerror.NewHttp(400, err)
		}
		if input.Actual.Time.IsZero() {
			return echoerror.NewHttp(400, fmt.Errorf("invalid date"))
		}
		writeMutex.Lock()
		defer writeMutex.Unlock()

		meds, err := DBMedListGet(db)
		if err != nil {
			return echoerror.NewHttp(500, err)
		}

		var dir *Direction
		for _, med := range meds {
			d := med
			if med.KeyName() == input.MedKeyname {
				dir = &d
			} else if med.Name == input.MedKeyname {
				input.MedKeyname = med.KeyName()
				dir = &d
			}
		}
		if dir == nil {
			return echoerror.NewHttp(404, fmt.Errorf("med not found"))
		}

		if err := DBMedComplianceLogSetOne(db, *dir, &input); err != nil {
			return err
		}

		if input.Actual.Dose <= 0 {
			return c.NoContent(204)
		}

		return c.JSON(200, input)
	}
}

func RESTComplianceLogProjectMed(db db.DB) func(c echo.Context) error {
	return func(c echo.Context) error {
		keyName := c.Param("med")
		meds, err := DBMedListGet(db)
		if err != nil {
			return echoerror.NewHttp(500, err)
		}

		var dir *Direction
		for _, med := range meds {
			if med.KeyName() == keyName {
				d := med
				dir = &d
			}
		}
		if dir == nil {
			return echoerror.NewHttp(404, fmt.Errorf("med not found"))
		}

		complianceLog, err := DBMedComplianceLogGet(db, util.DateRangeAround(time.Now(), 1))
		if err != nil {
			return echoerror.NewHttp(500, err)
		}

		return c.JSON(200, complianceLog.ProjectNextDose(*dir))
	}
}

func RESTRecalcMedComplianceLog(db db.DB, writeMutex *sync.Mutex) func(c echo.Context) error {
	return func(c echo.Context) error {
		meds, err := DBMedListGet(db)
		if err != nil {
			return echoerror.NewHttp(500, err)
		}

		from := time.Date(yearAbsZero, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Now()
		if fromStr := c.QueryParam("from"); fromStr != "" {
			from, err = time.Parse("2006-01", fromStr)
			if err != nil {
				return echoerror.NewHttp(400, err)
			}
		}
		if toStr := c.QueryParam("to"); toStr != "" {
			to, err = time.Parse("2006-01", toStr)
			if err != nil {
				return echoerror.NewHttp(400, err)
			}
		}

		writeMutex.Lock()
		defer writeMutex.Unlock()
		for year := from.Year(); year <= to.Year(); year++ {
			for month := 1; month <= 12; month++ {
				if year == from.Year() && month < int(from.Month()) {
					continue
				}
				if year == to.Year() && month > int(to.Month()) {
					continue
				}

				log, err := DBMedComplianceLogGet(db, util.DateRangeAround(time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC), 1))
				if err != nil {
					return echoerror.NewHttp(500, err)
				}
				if len(log) == 0 {
					continue
				}

				for _, dir := range meds {
					log.UpdateDoseOffset(dir)
				}
				if err := DBMedComplianceLogAppend(db, log); err != nil {
					return echoerror.NewHttp(500, err)
				}
			}
		}

		return c.NoContent(204)
	}
}
