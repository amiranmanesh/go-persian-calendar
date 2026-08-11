package ptime

// A Month specifies a month of the year, starting from Farvardin = 1.
type Month int

// Months of the Persian calendar, as used in Iran.
const (
	Farvardin Month = 1 + iota
	Ordibehesht
	Khordad
	Tir
	Mordad
	Shahrivar
	Mehr
	Aban
	Azar
	Dey
	Bahman
	Esfand
)

// Months of the Persian calendar, as used in Afghanistan (Dari names).
// They are aliases of the Iranian names above and share the same values.
const (
	Hamal Month = 1 + iota
	Sur
	Jauza
	Saratan
	Asad
	Sonboleh
	Mizan
	Aqrab
	Qos
	Jady
	Dolv
	Hut
)

var months = [12]string{
	"فروردین",
	"اردیبهشت",
	"خرداد",
	"تیر",
	"مرداد",
	"شهریور",
	"مهر",
	"آبان",
	"آذر",
	"دی",
	"بهمن",
	"اسفند",
}

var dariMonths = [12]string{
	"حمل",
	"ثور",
	"جوزا",
	"سرطان",
	"اسد",
	"سنبله",
	"میزان",
	"عقرب",
	"قوس",
	"جدی",
	"دلو",
	"حوت",
}

// pMonthCount holds {days, leap_days, days_before_start} for every month.
var pMonthCount = [12][3]int{
	{31, 31, 0},   // Farvardin
	{31, 31, 31},  // Ordibehesht
	{31, 31, 62},  // Khordad
	{31, 31, 93},  // Tir
	{31, 31, 124}, // Mordad
	{31, 31, 155}, // Shahrivar
	{30, 30, 186}, // Mehr
	{30, 30, 216}, // Aban
	{30, 30, 246}, // Azar
	{30, 30, 276}, // Dey
	{30, 30, 306}, // Bahman
	{29, 30, 336}, // Esfand
}

// String returns the Iranian Persian name of the month.
// Out of range values are clamped to Farvardin and Esfand.
func (m Month) String() string {
	return months[m.index()]
}

// Dari returns the Dari (Afghan Persian) name of the month.
// Out of range values are clamped to Hamal and Hut.
func (m Month) Dari() string {
	return dariMonths[m.index()]
}

// IsValid reports whether m is in the range [Farvardin, Esfand].
func (m Month) IsValid() bool {
	return m >= Farvardin && m <= Esfand
}

// index returns the zero based index of m, clamped to [0, 11].
func (m Month) index() int {
	switch {
	case m < Farvardin:
		return 0
	case m > Esfand:
		return 11
	default:
		return int(m) - 1
	}
}
