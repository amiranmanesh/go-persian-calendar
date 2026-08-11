package ptime

// An AmPm specifies the 12-hour clock marker.
type AmPm int

// The 12-hour clock markers.
const (
	Am AmPm = iota
	Pm
)

// A DayTime represents a part of the day, derived from the hour.
type DayTime int

// Parts of the day, each covering a three hour window.
const (
	Midnight DayTime = iota
	Dawn
	Morning
	BeforeNoon
	Noon
	AfterNoon
	Evening
	Night
)

var amPmNames = [2]string{
	"قبل از ظهر",
	"بعد از ظهر",
}

var shortAmPmNames = [2]string{
	"ق.ظ",
	"ب.ظ",
}

var dayTimeNames = [8]string{
	"نیمه\u200cشب",
	"سحر",
	"صبح",
	"قبل از ظهر",
	"ظهر",
	"بعد از ظهر",
	"عصر",
	"شب",
}

// String returns the Persian name of the 12-hour marker.
func (a AmPm) String() string {
	return amPmNames[a.index()]
}

// Short returns the abbreviated Persian name of the 12-hour marker.
func (a AmPm) Short() string {
	return shortAmPmNames[a.index()]
}

// index returns the index of a, clamped to [0, 1].
func (a AmPm) index() int {
	if a <= Am {
		return 0
	}

	return 1
}

// String returns the Persian name of the part of the day.
// Out of range values are clamped to Midnight and Night.
func (d DayTime) String() string {
	switch {
	case d < Midnight:
		return dayTimeNames[0]
	case d > Night:
		return dayTimeNames[len(dayTimeNames)-1]
	default:
		return dayTimeNames[d]
	}
}
