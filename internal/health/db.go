package health

import (
	"errors"
	"sort"
	"time"

	"github.com/eternal-flame-AD/yoake/internal/db"
	"github.com/eternal-flame-AD/yoake/internal/echoerror"
	"github.com/eternal-flame-AD/yoake/internal/util"
	"github.com/google/uuid"
)

func DBMedListGet(database db.DB) ([]Direction, error) {
	txn := database.NewTransaction(false)
	defer txn.Discard()

	var meds []Direction
	if err := db.GetJSON(txn, []byte("health_meds_list"), &meds); db.IsNotFound(err) {
		err = DBMedListSet(database, []Direction{})
		return meds, err
	} else if err != nil {
		return nil, err
	}

	return meds, nil
}

func DBMedListSet(database db.DB, meds []Direction) error {
	txn := database.NewTransaction(true)
	defer txn.Discard()

	for i, med := range meds {
		_, meds[i].DirectionShorthand = med.ShortHand()
	}
	if err := db.SetJSON(txn, []byte("health_meds_list"), meds); err != nil {
		return err
	}

	return txn.Commit()
}

const dbMedComplianceLogPrefix = "health_meds_compliance_log_"

func DBMedComplianceLogGet(database db.DB, dates util.DateRange) (ComplianceLogList, error) {
	txn := database.NewTransaction(false)
	defer txn.Discard()

	endIndex := dates.To.UTC().Format("2006-01")
	indexesToFetch := []string{endIndex}
	for indexNow := dates.From.UTC().AddDate(0, -1, 0); indexNow.Before(dates.To.AddDate(0, 1, 0)); indexNow = indexNow.AddDate(0, 1, 0) {
		indexesToFetch = append(indexesToFetch, indexNow.Format("2006-01"))
	}
	indexesToFetch = util.Unique(indexesToFetch)
	sort.Strings(indexesToFetch)

	var res ComplianceLogList
	for _, index := range indexesToFetch {
		var log []ComplianceLog
		if err := db.GetJSON(txn, []byte(dbMedComplianceLogPrefix+index), &log); db.IsNotFound(err) {
			continue
		} else if err != nil {
			return nil, err
		}
		res = append(res, log...)
	}
	sort.Sort(res)
	return res, nil
}

func DBMedComplianceLogAppend(database db.DB, pending ComplianceLogList) error {
	txn := database.NewTransaction(true)
	defer txn.Discard()

	for len(pending) > 0 {
		index := pending[0].Actual.Time.UTC().Format("2006-01")
		var origLogs ComplianceLogList
		if err := db.GetJSON(txn, []byte(dbMedComplianceLogPrefix+index), &origLogs); db.IsNotFound(err) {
			origLogs = []ComplianceLog{}
		} else if err != nil {
			return err
		}

		for i := len(pending) - 1; i >= 0; i-- {
			if pending[i].Actual.Time.UTC().Format("2006-01") != index {
				continue
			}
			origLogs = append(origLogs, pending[i])
			pending = append(pending[:i], pending[i+1:]...)
			uuidMap := make(map[string]int)
			for j := len(origLogs) - 1; j >= 0; j-- {
				if _, ok := uuidMap[origLogs[j].UUID]; ok {
					origLogs = append(origLogs[:j], origLogs[j+1:]...)
				} else {
					uuidMap[origLogs[j].UUID] = j
				}
			}
		}

		sort.Sort(origLogs)

		if err := db.SetJSON(txn, []byte(dbMedComplianceLogPrefix+index), origLogs); err != nil {
			return err
		}
	}

	return txn.Commit()
}

func DBMedComplianceLogSetOne(database db.DB, dir Direction, log *ComplianceLog) error {

	index := log.Actual.Time.UTC().Format("2006-01")

	existingLogs, err := DBMedComplianceLogGet(database, util.DateRangeAround(log.Actual.Time, 1))
	if err != nil {
		return err
	}

	txn := database.NewTransaction(true)
	defer txn.Discard()

	del := false
	if log.Actual.Dose == 0 {
		return echoerror.NewHttp(400, errors.New("dose cannot be zero"))
	} else if log.Actual.Dose < 0 {
		del = true
	}

	if log.UUID != "" {
		foundIdx := -1
		for i, existingLog := range existingLogs {
			if existingLog.UUID == log.UUID {
				log.UpdatedAt = time.Now()
				foundIdx = i
				break
			}
		}
		if foundIdx < 0 {
			return echoerror.NewHttp(404, errors.New("log with specified UUID not found"))
		}
		origLog := existingLogs[foundIdx]
		log.CreatedAt = origLog.CreatedAt
		origLogIdx := origLog.Actual.Time.UTC().Format("2006-01")

		if origLogIdx == index {
			// update and return
			if del {
				existingLogs = append(existingLogs[:foundIdx], existingLogs[foundIdx+1:]...)
			} else {
				log.UpdatedAt = time.Now()
				existingLogs[foundIdx] = *log
			}
			sort.Sort(existingLogs)
			if err := db.SetJSON(txn, []byte(dbMedComplianceLogPrefix+index), existingLogs); err != nil {
				return err
			}
			return txn.Commit()
		} else {
			// delete from old index
			existingLogs = append(existingLogs[:foundIdx], existingLogs[foundIdx+1:]...)
			if err := db.SetJSON(txn, []byte(dbMedComplianceLogPrefix+origLogIdx), existingLogs); err != nil {
				return err
			}
		}
	}

	if del {
		return txn.Commit()
	}

	if log.UUID == "" {
		log.UUID = uuid.New().String()
		log.CreatedAt = time.Now()
		if log.Expected.Dose == 0 {
			nextDose := existingLogs.ProjectNextDose(dir)
			log.Expected.Time = nextDose.Expected.Time
			log.Expected.Dose = nextDose.Expected.Dose
		}
	}

	log.UpdatedAt = time.Now()
	var logs ComplianceLogList
	if err := db.GetJSON(txn, []byte(dbMedComplianceLogPrefix+index), &logs); db.IsNotFound(err) {
		logs = []ComplianceLog{*log}
	} else if err != nil {
		return err
	} else {
		logs = append(logs, *log)
	}

	logs.UpdateDoseOffset(dir)

	uuid := log.UUID
	sort.Sort(ComplianceLogList(logs))

	if err := db.SetJSON(txn, []byte(dbMedComplianceLogPrefix+index), logs); err != nil {
		return err
	}

	for i, l := range logs {
		if l.UUID == uuid {
			*log = logs[i]
		}
	}

	return txn.Commit()

}
