package health

import (
	"math"
	"sort"
	"time"

	"github.com/eternal-flame-AD/yoake/internal/util"
)

type ComplianceLog struct {
	UUID       string `json:"uuid,omitempty"`
	MedKeyname string `json:"med_keyname"`

	Expected ComplianceDoseInfo `json:"expected"`
	Actual   ComplianceDoseInfo `json:"actual"`

	// 0 = closest to expected time +1 = closest to next expected dose
	// get a cumsum of this to get a compliance stat
	DoseOffset float64 `json:"dose_offset"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ComplianceDoseInfo struct {
	Time time.Time `json:"time"`
	Dose int       `json:"dose"`
}

type ComplianceLogList []ComplianceLog

func (c ComplianceLogList) ProjectNextDose(dir Direction) (nextDose ComplianceLog) {
	sort.Sort(c)

	var lastDose ComplianceLog
	var cumDosage int
	for ptr := 0; ptr < len(c); ptr++ {
		if c[ptr].MedKeyname == dir.KeyName() {
			if dir.OptSchedule == OptScheduleWholeDose {
				lastDose = c[ptr]
				break
			} else {
				cumDosage += c[ptr].Actual.Dose
				if cumDosage < dir.Dosage {
					continue
				} else {
					lastDose = c[ptr]
					break
				}
			}
		}
	}
	if lastDose.UUID == "" /* not found */ {
		nextDose = ComplianceLog{
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
		}
	} else {
		nextDose = ComplianceLog{
			MedKeyname: dir.KeyName(),
			Expected: ComplianceDoseInfo{
				Time: lastDose.Actual.Time.Add(time.Duration(dir.PeriodHours) * time.Hour),
				Dose: dir.Dosage,
			},
			Actual: ComplianceDoseInfo{
				Time: time.Now(),
				Dose: dir.Dosage,
			},
			CreatedAt: time.Now(),
		}
		nextDose.DoseOffset, _, _ = c.ComputeDoseOffset(dir, &nextDose)
	}
	return
}

func (c ComplianceLogList) ComputeDoseOffset(dir Direction, newLog *ComplianceLog) (float64, bool, error) {
	sort.Sort(c)

	var lastTwoDoses []ComplianceLog
	if newLog != nil {
		lastTwoDoses = []ComplianceLog{*newLog}
	}
	for ptr := 0; len(lastTwoDoses) < 2 && ptr < len(c); ptr++ {
		if c[ptr].MedKeyname == dir.KeyName() {
			if len(lastTwoDoses) == 0 || lastTwoDoses[0].Actual.Time.After(c[ptr].Actual.Time) {
				lastTwoDoses = append(lastTwoDoses, c[ptr])
			}
		}
	}
	if newLog != nil {
		if newLog.Expected.Dose == 0 && dir.KeyName() == newLog.MedKeyname {
			newLog.Expected.Dose = dir.Dosage
		}
		if newLog.Expected.Time.IsZero() {
			if len(lastTwoDoses) == 2 {
				newLog.Expected.Time = lastTwoDoses[1].Actual.Time.Add(time.Duration(dir.PeriodHours) * time.Hour)
			} else {
				newLog.Expected.Time = newLog.Actual.Time
			}
		}
		lastTwoDoses[0] = *newLog
	}
	if len(lastTwoDoses) < 2 {
		return 0, false, nil
	}

	// now we have:
	//             *exp  ~actual
	//          * ~        ~ *
	// offset = (new_actual - last_expected) / diff(new_expected, last_actual) - 1

	if lastTwoDoses[0].Actual.Time.IsZero() {
		lastTwoDoses[0].Actual.Time = time.Now()
	}
	offset := float64(lastTwoDoses[0].Actual.Time.Sub(lastTwoDoses[1].Expected.Time))/
		float64(lastTwoDoses[0].Expected.Time.Sub(lastTwoDoses[1].Actual.Time)) - 1

	// for prn ignore positive offsets
	if util.Contain(dir.Flags, DirectionFlagPRN) {
		if offset > 0 {
			offset = 0
		}
	}

	// ad lib ignore negative offsets
	if util.Contain(dir.Flags, DirectionFlagAdLib) {
		if offset < 0 {
			offset = 0
		}
	}

	if math.Abs(offset) > 2 {
		// stop counting if three or more doses are missed
		return 0, false, nil
	}
	return offset, true, nil
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
