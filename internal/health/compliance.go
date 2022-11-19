package health

import (
	"encoding/json"
	"log"
	"math"
	"sort"
	"time"

	"github.com/eternal-flame-AD/yoake/internal/util"
	"github.com/google/uuid"
)

type ComplianceLog struct {
	UUID       string `json:"uuid,omitempty"`
	MedKeyname string `json:"med_keyname"`

	Expected ComplianceDoseInfo `json:"expected"`
	Actual   ComplianceDoseInfo `json:"actual"`

	// 0 = closest to expected time +1 = closest to next expected dose
	// get a cumsum of this to get a compliance stat
	DoseOffset f64OrNan `json:"dose_offset"`

	EffectiveLastDose *ComplianceLog `json:"effective_last_dose,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ComplianceDoseInfo struct {
	Time time.Time `json:"time"`
	Dose int       `json:"dose"`
}

type f64OrNan float64

func (n *f64OrNan) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*n = f64OrNan(math.NaN())
		return nil
	}
	var f float64
	if err := json.Unmarshal(b, &f); err != nil {
		return err
	}
	*n = f64OrNan(f)
	return nil
}

func (n f64OrNan) MarshalJSON() ([]byte, error) {
	if math.IsNaN(float64(n)) {
		return []byte("null"), nil
	}
	return json.Marshal(float64(n))
}

func doseOffset(dir Direction, this ComplianceLog, last ComplianceLog) float64 {
	if last.UUID == "" {
		return math.NaN()
	}

	offset := float64(this.Actual.Time.Sub(last.Actual.Time))/
		float64(time.Duration(dir.PeriodHours)*time.Hour) - 1

	// for prn ignore positive offsets
	if util.Contain(dir.Flags, DirectionFlagPRN) {
		if offset > 0 {
			return 0
		}
	}

	// ad lib ignore negative offsets
	if util.Contain(dir.Flags, DirectionFlagAdLib) {
		if offset < 0 {
			return 0
		}
	}

	return offset
}

type ComplianceLogList []ComplianceLog

func (c ComplianceLogList) findEffectiveLastDose(dir Direction, this ComplianceLog) ComplianceLog {
	// for ad lib directions, this finds the last dose
	// for default scheduling, this find the earliest dose that does not cumulatively exceed a whole dose

	var lastDose ComplianceLog
	var cumDosage int
	for ptr := 0; ptr < len(c); ptr++ {
		if c[ptr].MedKeyname == dir.KeyName() && c[ptr].Actual.Time.Before(this.Actual.Time) {
			if dir.OptSchedule == OptScheduleWholeDose {
				return c[ptr]
			}

			cumDosage += c[ptr].Actual.Dose
			if cumDosage > dir.Dosage {
				return lastDose
			} else if cumDosage == dir.Dosage {
				return c[ptr]
			}
			lastDose = c[ptr]

		}
	}
	return lastDose
}

func (c ComplianceLogList) ProjectNextDose(dir Direction) (nextDose ComplianceLog) {
	tmpUUID := uuid.New().String()

	nextDose = ComplianceLog{
		UUID:       tmpUUID,
		MedKeyname: dir.KeyName(),
		Expected: ComplianceDoseInfo{
			Time: time.Now(),
			Dose: dir.Dosage,
		},
		Actual: ComplianceDoseInfo{
			Time: time.Now(),
			Dose: dir.Dosage,
		},
		DoseOffset: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	lastDose := c.findEffectiveLastDose(dir, nextDose)
	if lastDose.UUID == "" /* not found */ {
		return
	}

	nextDose.EffectiveLastDose = &lastDose
	nextDose.Expected.Time = lastDose.Actual.Time.Add(time.Duration(dir.PeriodHours) * time.Hour)
	nextDose.DoseOffset = f64OrNan(doseOffset(dir, nextDose, lastDose))
	return
}

func (c ComplianceLogList) UpdateDoseOffset(dir Direction) {
	sort.Sort(c)

	for i := range c {
		if c[i].MedKeyname == dir.KeyName() {
			lastDose, thisDose := c.findEffectiveLastDose(dir, c[i]), c[i]
			if lastDose.UUID == "" /* not found */ {
				return
			}

			c[i].DoseOffset = f64OrNan(doseOffset(dir, thisDose, lastDose))
			log.Printf("thisDose: %+v, \nlastDose: %+v\n-->offset: %f\n", thisDose, lastDose, c[i].DoseOffset)
		}
	}
}

func (c ComplianceLogList) Len() int {
	return len(c)
}

func (c ComplianceLogList) Less(i, j int) bool {
	return c[i].Actual.Time.After(c[j].Actual.Time)
}

func (c ComplianceLogList) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
