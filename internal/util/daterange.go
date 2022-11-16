package util

import "time"

type DateRange struct {
	From time.Time
	To   time.Time
}

func NewDateRange(from time.Time, to time.Time) DateRange {
	from, _ = time.Parse("2006-01-02", from.Format("2006-01-02"))
	to, _ = time.Parse("2006-01-02", to.Format("2006-01-02"))
	if from.After(to) {
		from, to = to, from
	}
	return DateRange{from, to}
}

func (d DateRange) Days() int {
	return int(d.To.Sub(d.From).Hours() / 24)
}

func DateRangeAround(date time.Time, months int) DateRange {
	return NewDateRange(date.AddDate(0, -months, 0), date.AddDate(0, months, 0))
}
