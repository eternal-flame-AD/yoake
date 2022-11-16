package health

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/eternal-flame-AD/yoake/internal/db"
	"github.com/eternal-flame-AD/yoake/internal/echoerror"
	"github.com/labstack/echo/v4"
)

func RESTMedGetDirections(db db.DB) func(c echo.Context) error {
	return func(c echo.Context) error {
		defer func() {
			if err := recover(); err != nil {
				c.Error(echoerror.NewHttp(500, fmt.Errorf("internal error: %v", err)))
			}
		}()
		meds, err := DBMedListGet(db)
		if err != nil {
			return echoerror.NewHttp(500, err)
		}
		return c.JSON(200, meds)
	}
}

func RESTMedPostDirections(db db.DB, writeMutex *sync.Mutex) func(c echo.Context) error {
	return func(c echo.Context) error {
		var input Direction
		if err := c.Bind(&input); err != nil {
			return echoerror.NewHttp(400, err)
		}
		if input.Name == "" {
			return echoerror.NewHttp(400, fmt.Errorf("name cannot be empty"))
		}
		if input.Dosage <= 0 {
			return echoerror.NewHttp(400, fmt.Errorf("dosage must be positive"))
		}
		if input.PeriodHours <= 0 {
			return echoerror.NewHttp(400, fmt.Errorf("period must be positive"))
		}
		writeMutex.Lock()
		defer writeMutex.Unlock()
		meds, err := DBMedListGet(db)
		if err != nil {
			return echoerror.NewHttp(500, err)
		}
		found := false
		for i, med := range meds {
			if med.KeyName() == input.KeyName() {
				meds[i] = input
				found = true
			}
		}
		if !found {
			meds = append(meds, input)
		}
		if err := DBMedListSet(db, meds); err != nil {
			return echoerror.NewHttp(500, err)
		}
		return c.JSON(200, meds)
	}
}

func RESTMedDeleteDirections(db db.DB, writeMutex *sync.Mutex) func(c echo.Context) error {
	return func(c echo.Context) error {
		name := c.Param("name")
		writeMutex.Lock()
		defer writeMutex.Unlock()
		meds, err := DBMedListGet(db)
		if err != nil {
			return echoerror.NewHttp(500, err)
		}
		found := false
		for i, med := range meds {
			if strings.EqualFold(med.KeyName(), name) {
				meds = append(meds[:i], meds[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			return echoerror.NewHttp(404, fmt.Errorf("med not found"))
		}
		if err := DBMedListSet(db, meds); err != nil {
			return echoerror.NewHttp(500, err)
		}
		return c.NoContent(http.StatusNoContent)
	}
}
