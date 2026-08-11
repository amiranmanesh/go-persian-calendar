package ptime

import "time"

// A Weekday specifies a day of the week, starting from Shanbeh = 0.
type Weekday int

// Days of the week in the Persian calendar.
const (
	Shanbeh Weekday = iota
	Yekshanbeh
	Doshanbeh
	Seshanbeh
	Charshanbeh
	Panjshanbeh
	Jomeh
)

var weekdays = [7]string{
	"شنبه",
	"یک\u200cشنبه",
	"دوشنبه",
	"سه\u200cشنبه",
	"چهارشنبه",
	"پنج\u200cشنبه",
	"جمعه",
}

var shortWeekdays = [7]string{
	"ش",
	"ی",
	"د",
	"س",
	"چ",
	"پ",
	"ج",
}

// String returns the Persian name of the day of the week.
// Out of range values are clamped to Shanbeh and Jomeh.
func (d Weekday) String() string {
	return weekdays[d.index()]
}

// Short returns the single letter Persian abbreviation of the day of the week.
// Out of range values are clamped to Shanbeh and Jomeh.
func (d Weekday) Short() string {
	return shortWeekdays[d.index()]
}

// IsValid reports whether d is in the range [Shanbeh, Jomeh].
func (d Weekday) IsValid() bool {
	return d >= Shanbeh && d <= Jomeh
}

// index returns the index of d, clamped to [0, 6].
func (d Weekday) index() int {
	switch {
	case d < Shanbeh:
		return 0
	case d > Jomeh:
		return 6
	default:
		return int(d)
	}
}

// weekdayOf maps a Gregorian weekday onto the Persian week, which starts on
// Saturday. time.Saturday is 6, so shifting by one and wrapping lands on 0.
func weekdayOf(wd time.Weekday) Weekday {
	return Weekday((int(wd) + 1) % 7)
}
