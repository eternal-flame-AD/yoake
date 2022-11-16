package health

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShortHandParser(t *testing.T) {

	cases := [][]interface{}{
		{"Atorvastatin 10mg QD", &Direction{
			Name:        "Atorvastatin",
			PeriodHours: 24,
			Dosage:      10,
			DosageUnit:  "mg",
			Flags:       make([]DirectionFlag, 0),
			OptSchedule: OptScheduleDefault,
			Disclaimer:  DirectionDisclaimer,
		}},
		{"Atorvastatin 10mg TAB 10mg PO bid", &Direction{
			Name:        "Atorvastatin 10mg TAB",
			PeriodHours: 12,
			Dosage:      10,
			DosageUnit:  "mg",
			DosageRoute: "PO",
			Flags:       make([]DirectionFlag, 0),
			OptSchedule: OptScheduleDefault,
			Disclaimer:  DirectionDisclaimer,
		}},
		{"metformin 500mg qHS", &Direction{
			Name:        "metformin",
			PeriodHours: 24,
			Dosage:      500,
			DosageUnit:  "mg",
			Flags:       []DirectionFlag{DirectionFlagHS},
			OptSchedule: OptScheduleDefault,
			Disclaimer:  DirectionDisclaimer,
		}},
		{"Amphetamine 10mg qam", &Direction{
			Name:        "Amphetamine",
			PeriodHours: 24,
			Dosage:      10,
			DosageUnit:  "mg",
			Flags:       []DirectionFlag{DirectionFlagAM},
			OptSchedule: OptScheduleDefault,
			Disclaimer:  DirectionDisclaimer,
		}},
		{"Something 10mg tid ad lib", &Direction{
			Name:        "Something",
			PeriodHours: 8,
			Dosage:      10,
			DosageUnit:  "mg",
			Flags:       []DirectionFlag{DirectionFlagAdLib},
			OptSchedule: OptScheduleDefault,
			Disclaimer:  DirectionDisclaimer,
		}},
		{"Hydroxyzine 50mg qid prn sched(whole)", &Direction{
			Name:        "Hydroxyzine",
			PeriodHours: 6,
			Dosage:      50,
			DosageUnit:  "mg",
			Flags:       []DirectionFlag{DirectionFlagPRN},
			OptSchedule: OptScheduleWholeDose,
			Disclaimer:  DirectionDisclaimer,
		}}}

	for _, c := range cases {
		input, expected := c[0].(string), c[1].(*Direction)
		actual, err := ParseShorthand(input)
		if expected == nil {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			expected.DirectionShorthand = actual.DirectionShorthand
			assert.Equal(t, expected, actual)
			name, encoded := actual.ShortHand()
			assert.Equal(t, expected.Name, name)
			encodedDecoded, err := ParseShorthand(name + " " + encoded)
			assert.NoError(t, err)
			assert.Equal(t, expected, encodedDecoded)
		}
	}
}
