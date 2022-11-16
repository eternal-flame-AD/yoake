package health

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/eternal-flame-AD/yoake/internal/util"
)

type Direction struct {
	Name string `json:"name"`

	PeriodHours int `json:"period_hours"`

	Dosage      int    `json:"dosage"`
	DosageUnit  string `json:"dosage_unit"`
	DosageRoute string `json:"dosage_route"`

	Flags []DirectionFlag `json:"flags"`

	DirectionShorthand string `json:"shorthand"`

	OptSchedule OptSchedule `json:"schedule"`

	Disclaimer string `json:"__disclaimer"`
}

const DirectionDisclaimer = "For personal use only. No warranty of accuracy."

type DirectionFlag string
type OptSchedule string

const (
	DirectionFlagAM      DirectionFlag = "qam"
	DirectionFlagHS      DirectionFlag = "qhs"
	DirectionFlagPRN     DirectionFlag = "prn"
	DirectionFlagAdLib   DirectionFlag = "ad lib"
	OptScheduleDefault   OptSchedule   = "default"
	OptScheduleWholeDose OptSchedule   = "whole"
)

func ParseShorthand(shorthand string) (*Direction, error) {
	res := new(Direction)
	res.Disclaimer = DirectionDisclaimer
	res.Flags = make([]DirectionFlag, 0)
	words := strings.Split(shorthand, " ")

	optionsRegex := regexp.MustCompile(`^([a-zA-Z]+)\((\w+)\)$`)
	for i := len(words) - 1; i >= 0; i-- {
		if match := optionsRegex.FindStringSubmatch(words[i]); match != nil {
			name, value := match[1], match[2]
			switch strings.ToLower(name) {
			case "sched":
				fallthrough
			case "schedule":
				if res.OptSchedule != "" {
					return nil, fmt.Errorf("duplicate schedule option")
				}
				switch strings.ToLower(value) {
				case "default":
					fallthrough
				case "":
					res.OptSchedule = OptScheduleDefault
				case "whole":
					res.OptSchedule = OptScheduleWholeDose
				default:
					return nil, fmt.Errorf("invalid schedule option: %s", value)
				}
			default:
				return nil, fmt.Errorf("unknown option %s", name)
			}

			words = words[:i]
		}
	}
	if res.OptSchedule == "" {
		res.OptSchedule = OptScheduleDefault
	}

	// combined numbers and units
	for i := range words {
		digits := regexp.MustCompile(`^\d+$`)
		if digits.MatchString(words[i]) {
			words[i] = words[i] + words[i+1]
			words[i+1] = ""
		}
		if strings.ToLower(words[i]) == "ad" && strings.ToLower(words[i+1]) == "lib" {
			res.Flags = append(res.Flags, DirectionFlagAdLib)
			words[i] = ""
			words[i+1] = ""
		} else if strings.ToLower(words[i]) == "adlib" {
			res.Flags = append(res.Flags, DirectionFlagAdLib)
			words[i] = ""
		}
	}
	words = util.AntiJoin(words, []string{""})

	// find prn keyword
	for i := len(words) - 1; i >= 0; i-- {
		if strings.EqualFold(words[i], "prn") {
			if util.Contain(res.Flags, DirectionFlagAdLib) {
				return nil, fmt.Errorf("cannot use 'ad lib' and 'prn' together")
			}
			res.Flags = append(res.Flags, DirectionFlagPRN)
			words = append(words[:i], words[i+1:]...)
			break
		}
	}

	freqIdx := len(words) - 1
	if lastWord := strings.ToLower(words[len(words)-1]); lastWord == "bid" {
		res.PeriodHours = 12
		words = words[:len(words)-1]
	} else if lastWord == "tid" {
		res.PeriodHours = 8
		words = words[:len(words)-1]
	} else if lastWord == "qid" {
		res.PeriodHours = 6
		words = words[:len(words)-1]
	} else {
		for i := len(words) - 1; i >= 0; i-- {
			if strings.HasPrefix(strings.ToLower(words[i]), "q") {
				freqIdx = i
				break
			}
		}

		freqStr := strings.ToLower(strings.Join(words[freqIdx:], ""))[1:]
		if freqStr == "am" {
			res.Flags = append(res.Flags, DirectionFlagAM)
			res.PeriodHours = 24
		} else if freqStr == "hs" {
			res.Flags = append(res.Flags, DirectionFlagHS)
			res.PeriodHours = 24
		} else {
			if !(freqStr[0] >= '0' && freqStr[0] <= '9') {
				freqStr = "1" + freqStr
			}
			freqRegexp := regexp.MustCompile(`^([0-9]+)([a-z]+)$`)
			freqMatch := freqRegexp.FindStringSubmatch(freqStr)
			if freqMatch == nil {
				return nil, fmt.Errorf("invalid frequency: %s", freqStr)
			}
			freq, err := strconv.Atoi(freqMatch[1])
			if err != nil {
				return nil, fmt.Errorf("invalid frequency number : %s", freqMatch[1])
			}
			if freqMatch[2] == "d" {
				res.PeriodHours = freq * 24
			} else if freqMatch[2] == "h" {
				res.PeriodHours = freq
			} else {
				return nil, fmt.Errorf("invalid frequency unit: %s", freqMatch[2])
			}
		}
	}

	words = words[:freqIdx]

	dosageRegexp := regexp.MustCompile(`^([0-9]+)([a-z]*)$`)
	var dosageMatch []string
	if dosageMatch = dosageRegexp.FindStringSubmatch(words[len(words)-1]); dosageMatch == nil {
		if dosageMatch = dosageRegexp.FindStringSubmatch(words[len(words)-2]); dosageMatch == nil {
			return nil, fmt.Errorf("invalid dosage: %s", words[len(words)-2:])
		} else {
			res.DosageRoute = words[len(words)-1]
			words = words[:len(words)-2]
		}
	} else {
		words = words[:len(words)-1]
	}
	dosage, err := strconv.Atoi(dosageMatch[1])
	if err != nil {
		return nil, fmt.Errorf("invalid dosage number: %s", dosageMatch[1])
	}
	res.Dosage = dosage
	res.DosageUnit = dosageMatch[2]

	res.Name = strings.Join(words, " ")

	s1, s2 := res.ShortHand()
	res.DirectionShorthand = s1 + " " + s2
	return res, nil
}

func (d Direction) KeyName() string {
	return strings.ToLower(strings.SplitN(d.Name, " ", 2)[0])
}
func (d *Direction) ShortHand() (name string, direction string) {
	builder := new(strings.Builder)
	builder.WriteString(strconv.Itoa(d.Dosage))
	if d.DosageUnit != "" {
		builder.WriteString(" ")
		builder.WriteString(d.DosageUnit)
	}
	if d.DosageRoute != "" {
		builder.WriteString(" ")
		builder.WriteString(d.DosageRoute)
	}
	if d.PeriodHours%24 == 0 {
		qNd := d.PeriodHours / 24
		qNdS := strconv.Itoa(qNd)
		if qNd == 1 {
			if util.Contain(d.Flags, DirectionFlagAM) {
				qNdS = "AM"
			} else if util.Contain(d.Flags, DirectionFlagHS) {
				qNdS = "HS"
			} else {
				qNdS = "d"
			}
		} else {
			qNdS += "d"
		}
		fmt.Fprintf(builder, " q%s", qNdS)
	} else if d.PeriodHours == 12 {
		builder.WriteString(" bid")
	} else if d.PeriodHours == 8 {
		builder.WriteString(" tid")
	} else if d.PeriodHours == 6 {
		builder.WriteString(" qid")
	} else {
		fmt.Fprintf(builder, " q%sh", strconv.Itoa(d.PeriodHours))
	}
	if util.Contain(d.Flags, DirectionFlagPRN) {
		builder.WriteString(" PRN")
	} else if util.Contain(d.Flags, DirectionFlagAdLib) {
		builder.WriteString(" ad lib")
	}
	if d.OptSchedule != "" && d.OptSchedule != OptScheduleDefault {
		if d.OptSchedule == OptScheduleWholeDose {
			builder.WriteString(" sched(whole)")
		}
	}

	return d.Name, builder.String()
}

type Dose struct {
	Time time.Time
	Dose int
}
