package ptime

import (
	"math"
	"time"
)

// minGregorianYear is the oldest Gregorian year the conversion algorithms
// support. New returns the zero Time for anything older.
const minGregorianYear = 1097

// A Time represents an instant in the Persian (Solar Hijri) calendar with
// nanosecond precision.
//
// The zero value is reported by [Time.IsZero]. Like [time.Time], a Time is
// small enough to pass by value; every method with a value receiver returns a
// new Time instead of mutating the receiver.
type Time struct {
	year   int
	month  Month
	day    int
	hour   int
	minute int
	sec    int
	nsec   int
	loc    *time.Location
	wday   Weekday
}

// New converts a Gregorian time into the Persian calendar.
//
// It returns the zero Time if the Gregorian year of t is below 1097.
func New(t time.Time) Time {
	if t.Year() < minGregorianYear {
		return Time{}
	}

	var pt Time
	pt.SetTime(t)

	return pt
}

// Date returns the Time corresponding to the given Persian calendar date and
// clock reading.
//
// The arguments may be outside their usual ranges and are normalized during the
// conversion. If loc is nil, the local time zone is used.
func Date(year int, month Month, day, hour, minute, sec, nsec int, loc *time.Location) Time {
	if loc == nil {
		//nolint:gosmopolitan
		loc = time.Local
	}

	var t Time
	t.Set(year, month, day, hour, minute, sec, nsec, loc)

	return t
}

// Unix returns the Time corresponding to the given Unix time: sec seconds and
// nsec nanoseconds since January 1, 1970 UTC.
func Unix(sec, nsec int64) Time {
	return New(time.Unix(sec, nsec))
}

// UnixMilli returns the Time corresponding to the given Unix time: msec
// milliseconds since January 1, 1970 UTC.
func UnixMilli(msec int64) Time {
	return New(time.UnixMilli(msec))
}

// UnixMicro returns the Time corresponding to the given Unix time: usec
// microseconds since January 1, 1970 UTC.
func UnixMicro(usec int64) Time {
	return New(time.UnixMicro(usec))
}

// Now returns the current time in the Persian calendar and the local time zone.
func Now() Time {
	return New(time.Now())
}

// Time converts t back into the Gregorian calendar.
func (t Time) Time() time.Time {
	var year, month, day int

	jdn := persianToJDN(t.year, int(t.month), t.day)

	if jdn > gregorianReformJulianDay {
		year, month, day = jdnToGregorianPostReform(jdn)
	} else {
		year, month, day = jdnToGregorianPreReform(jdn)
	}

	loc := t.loc
	if loc == nil {
		//nolint:gosmopolitan
		loc = time.Local
	}

	return time.Date(year, time.Month(month), day, t.hour, t.minute, t.sec, t.nsec, loc)
}

// SetTime sets t to the Persian calendar equivalent of the Gregorian time ti.
func (t *Time) SetTime(ti time.Time) {
	t.nsec = ti.Nanosecond()
	t.sec = ti.Second()
	t.minute = ti.Minute()
	t.hour = ti.Hour()
	t.loc = ti.Location()
	t.wday = weekdayOf(ti.Weekday())

	gy, gmm, gd := ti.Date()
	gm := int(gmm)

	var jdn int
	if isAfterGregorianReform(gy, gm, gd) {
		jdn = gregorianPostReformToJDN(gy, gm, gd)
	} else {
		jdn = gregorianPreReformToJDN(gy, gm, gd)
	}

	year, month, day := jdnToPersian(jdn)

	t.year = year
	t.month = Month(month)
	t.day = day
}

// SetUnix sets t to the Unix time sec seconds and nsec nanoseconds since
// January 1, 1970 UTC.
func (t *Time) SetUnix(sec, nsec int64) {
	t.SetTime(time.Unix(sec, nsec))
}

// Set sets every field of t at once.
//
// The arguments may be outside their usual ranges and are normalized.
// loc must not be nil.
func (t *Time) Set(year int, month Month, day, hour, minute, sec, nsec int, loc *time.Location) {
	if loc == nil {
		panic("ptime: the Location must not be nil in call to Set")
	}

	// Normalize nsec, sec, minute and hour, overflowing into day.
	sec, nsec = norm(sec, nsec, 1e9)
	minute, sec = norm(minute, sec, 60)
	hour, minute = norm(hour, minute, 60)
	day, hour = norm(day, hour, 24)

	// Normalize month, overflowing into year.
	m := int(month) - 1
	year, m = norm(year, m, 12)

	switch {
	case m < 0:
		m = 0
	case m > 11:
		m = 11
	}

	if isLeap(year) {
		m, day = normDay(m, day, pMonthCount[m][1])
	} else {
		m, day = normDay(m, day, pMonthCount[m][0])
	}

	year, m = norm(year, m, 12)

	t.year = year
	t.month = Month(m) + 1
	t.day = day
	t.hour = hour
	t.minute = minute
	t.sec = sec
	t.nsec = nsec
	t.loc = loc
	t.resetWeekday()

	t.normalize()
}

// SetYear sets the year of t.
func (t *Time) SetYear(year int) {
	t.year = year
	t.normDay()
	t.resetWeekday()
}

// SetMonth sets the month of t.
func (t *Time) SetMonth(month Month) {
	t.month = month
	t.normMonth()
	t.normDay()
	t.resetWeekday()
}

// SetDay sets the day of month of t.
func (t *Time) SetDay(day int) {
	t.day = day
	t.normDay()
	t.resetWeekday()
}

// SetHour sets the hour of t.
func (t *Time) SetHour(hour int) {
	t.hour = hour
	t.normHour()
}

// SetMinute sets the minute of t.
func (t *Time) SetMinute(minute int) {
	t.minute = minute
	t.normMinute()
}

// SetSecond sets the second of t.
func (t *Time) SetSecond(sec int) {
	t.sec = sec
	t.normSecond()
}

// SetNanosecond sets the nanosecond of t.
func (t *Time) SetNanosecond(nsec int) {
	t.nsec = nsec
	t.normNanosecond()
}

// At sets the clock reading of t.
func (t *Time) At(hour, minute, sec, nsec int) {
	t.SetHour(hour)
	t.SetMinute(minute)
	t.SetSecond(sec)
	t.SetNanosecond(nsec)
}

// In returns a copy of t associated with loc.
//
// It changes only the location, not the calendar fields. loc must not be nil.
func (t Time) In(loc *time.Location) Time {
	if loc == nil {
		panic("ptime: the Location must not be nil in call to In")
	}

	t.loc = loc
	t.resetWeekday()

	return t
}

// Location returns the time zone of t.
func (t Time) Location() *time.Location {
	return t.loc
}

// Zone returns the time zone abbreviation and its offset in seconds east of UTC.
func (t Time) Zone() (string, int) {
	return t.Time().Zone()
}

// IsZero reports whether t is the zero Time.
func (t Time) IsZero() bool {
	return t == Time{}
}

// Before reports whether t is before u.
func (t Time) Before(u Time) bool {
	return t.Time().Before(u.Time())
}

// After reports whether t is after u.
func (t Time) After(u Time) bool {
	return t.Time().After(u.Time())
}

// Equal reports whether t and u represent the same instant.
//
// Two times can be equal even if they are in different locations: 6:00 +0200
// and 4:00 UTC are equal.
func (t Time) Equal(u Time) bool {
	return t.Time().Equal(u.Time())
}

// Compare compares t with u: it returns -1 if t is before u, +1 if t is after
// u, and 0 if they represent the same instant.
func (t Time) Compare(u Time) int {
	tt, ut := t.Time(), u.Time()

	switch {
	case tt.Before(ut):
		return -1
	case tt.After(ut):
		return 1
	default:
		return 0
	}
}

// Unix returns the number of seconds since January 1, 1970 UTC.
func (t Time) Unix() int64 {
	return t.Time().Unix()
}

// UnixMilli returns the number of milliseconds since January 1, 1970 UTC.
func (t Time) UnixMilli() int64 {
	return t.Time().UnixMilli()
}

// UnixMicro returns the number of microseconds since January 1, 1970 UTC.
func (t Time) UnixMicro() int64 {
	return t.Time().UnixMicro()
}

// UnixNano returns the number of nanoseconds since January 1, 1970 UTC.
func (t Time) UnixNano() int64 {
	return t.Time().UnixNano()
}

// Date returns the year, month and day of month of t.
func (t Time) Date() (int, Month, int) {
	return t.year, t.month, t.day
}

// Clock returns the hour, minute and second of t.
func (t Time) Clock() (int, int, int) {
	return t.hour, t.minute, t.sec
}

// Year returns the year of t.
func (t Time) Year() int {
	return t.year
}

// Month returns the month of t in the range [Farvardin, Esfand].
func (t Time) Month() Month {
	return t.month
}

// Day returns the day of month of t.
func (t Time) Day() int {
	return t.day
}

// Hour returns the hour of t in the range [0, 23].
func (t Time) Hour() int {
	return t.hour
}

// Hour12 returns the hour of t on a 12-hour clock, in the range [0, 11].
func (t Time) Hour12() int {
	if t.hour >= 12 {
		return t.hour - 12
	}

	return t.hour
}

// Minute returns the minute of t in the range [0, 59].
func (t Time) Minute() int {
	return t.minute
}

// Second returns the second of t in the range [0, 59].
func (t Time) Second() int {
	return t.sec
}

// Nanosecond returns the nanosecond of t in the range [0, 999999999].
func (t Time) Nanosecond() int {
	return t.nsec
}

// Weekday returns the day of the week of t.
func (t Time) Weekday() Weekday {
	return t.wday
}

// AmPm returns the 12-hour clock marker of t.
func (t Time) AmPm() AmPm {
	if t.hour > 12 || (t.hour == 12 && (t.minute > 0 || t.sec > 0)) {
		return Pm
	}

	return Am
}

// DayTime returns the part of the day t falls into:
//
//	[0,3)   midnight
//	[3,6)   dawn
//	[6,9)   morning
//	[9,12)  before noon
//	[12,15) noon
//	[15,18) afternoon
//	[18,21) evening
//	[21,24) night
func (t Time) DayTime() DayTime {
	return DayTime(t.hour / 3)
}

// IsLeap reports whether t falls in a leap year.
func (t Time) IsLeap() bool {
	return isLeap(t.year)
}

// YearDay returns the day of the year of t, in the range [1, 366].
func (t Time) YearDay() int {
	return pMonthCount[t.month.index()][2] + t.day
}

// RYearDay returns the number of days remaining in the year of t.
func (t Time) RYearDay() int {
	days := 365
	if t.IsLeap() {
		days++
	}

	return days - t.YearDay()
}

// RMonthDay returns the number of days remaining in the month of t.
func (t Time) RMonthDay() int {
	return pMonthCount[t.month.index()][t.leapIndex()] - t.day
}

// MonthWeek returns the week of the month of t.
func (t Time) MonthWeek() int {
	return int(math.Ceil(float64(t.day+int(t.FirstMonthDay().Weekday())) / 7.0))
}

// YearWeek returns the week of the year of t.
func (t Time) YearWeek() int {
	return int(math.Ceil(float64(t.YearDay()+int(t.FirstYearDay().Weekday())) / 7.0))
}

// RYearWeek returns the number of weeks remaining in the year of t.
func (t Time) RYearWeek() int {
	return 52 - t.YearWeek()
}

// BeginningOfWeek returns the first day of the week of t at 00:00:00.
func (t Time) BeginningOfWeek() Time {
	nt := t.AddDate(0, 0, int(Shanbeh-t.wday))
	nt.At(0, 0, 0, 0)

	return nt
}

// FirstWeekDay returns the first day of the week of t, keeping the clock of t.
func (t Time) FirstWeekDay() Time {
	if t.wday == Shanbeh {
		return t
	}

	return t.AddDate(0, 0, int(Shanbeh-t.wday))
}

// LastWeekday returns the last day of the week of t, keeping the clock of t.
func (t Time) LastWeekday() Time {
	if t.wday == Jomeh {
		return t
	}

	return t.AddDate(0, 0, int(Jomeh-t.wday))
}

// BeginningOfMonth returns the first day of the month of t at 00:00:00.
func (t Time) BeginningOfMonth() Time {
	return Date(t.year, t.month, 1, 0, 0, 0, 0, t.loc)
}

// FirstMonthDay returns the first day of the month of t, keeping the clock of t.
func (t Time) FirstMonthDay() Time {
	if t.day == 1 {
		return t
	}

	return Date(t.year, t.month, 1, t.hour, t.minute, t.sec, t.nsec, t.loc)
}

// LastMonthDay returns the last day of the month of t, keeping the clock of t.
func (t Time) LastMonthDay() Time {
	ld := pMonthCount[t.month.index()][t.leapIndex()]
	if ld == t.day {
		return t
	}

	return Date(t.year, t.month, ld, t.hour, t.minute, t.sec, t.nsec, t.loc)
}

// BeginningOfYear returns the first day of the year of t at 00:00:00.
func (t Time) BeginningOfYear() Time {
	return Date(t.year, Farvardin, 1, 0, 0, 0, 0, t.loc)
}

// FirstYearDay returns the first day of the year of t, keeping the clock of t.
func (t Time) FirstYearDay() Time {
	if t.month == Farvardin && t.day == 1 {
		return t
	}

	return Date(t.year, Farvardin, 1, t.hour, t.minute, t.sec, t.nsec, t.loc)
}

// LastYearDay returns the last day of the year of t, keeping the clock of t.
func (t Time) LastYearDay() Time {
	ld := pMonthCount[Esfand.index()][t.leapIndex()]
	if t.month == Esfand && t.day == ld {
		return t
	}

	return Date(t.year, Esfand, ld, t.hour, t.minute, t.sec, t.nsec, t.loc)
}

// Yesterday returns the day before t.
func (t Time) Yesterday() Time {
	return t.AddDate(0, 0, -1)
}

// Tomorrow returns the day after t.
func (t Time) Tomorrow() Time {
	return t.AddDate(0, 0, 1)
}

// Add returns t+d.
func (t Time) Add(d time.Duration) Time {
	return New(t.Time().Add(d))
}

// AddDate returns the time corresponding to adding the given number of years,
// months and days to t. The result is normalized the same way [Time.Set] is.
func (t Time) AddDate(years, months, days int) Time {
	t.Set(t.year+years, Month(int(t.month)+months), t.day+days, t.hour, t.minute, t.sec, t.nsec, t.loc)

	return t
}

// Sub returns the duration t-u.
func (t Time) Sub(u Time) time.Duration {
	return t.Time().Sub(u.Time())
}

// Truncate returns t rounded down to a multiple of d since the zero time.
// If d <= 0, t is returned unchanged.
func (t Time) Truncate(d time.Duration) Time {
	if d <= 0 {
		return t
	}

	return New(t.Time().Truncate(d))
}

// Round returns t rounded to the nearest multiple of d since the zero time,
// rounding half away from zero. If d <= 0, t is returned unchanged.
func (t Time) Round(d time.Duration) Time {
	if d <= 0 {
		return t
	}

	return New(t.Time().Round(d))
}

// Since returns the absolute number of seconds between t and t2.
//
// Deprecated: use [Time.Sub] instead, which returns a signed [time.Duration].
func (t Time) Since(t2 Time) int64 {
	return int64(math.Abs(float64(t2.Unix() - t.Unix())))
}

// leapIndex returns the pMonthCount column to use for the year of t.
func (t Time) leapIndex() int {
	if t.IsLeap() {
		return 1
	}

	return 0
}

func (t *Time) normalize() {
	t.normNanosecond()
	t.normSecond()
	t.normMinute()
	t.normHour()
	t.normMonth()
	t.normDay()
}

func (t *Time) normNanosecond() {
	clamp(&t.nsec, 0, 999999999)
}

func (t *Time) normSecond() {
	clamp(&t.sec, 0, 59)
}

func (t *Time) normMinute() {
	clamp(&t.minute, 0, 59)
}

func (t *Time) normHour() {
	clamp(&t.hour, 0, 23)
}

func (t *Time) normMonth() {
	switch {
	case t.month < Farvardin:
		t.month = Farvardin
	case t.month > Esfand:
		t.month = Esfand
	}
}

func (t *Time) normDay() {
	clamp(&t.day, 1, pMonthCount[t.month.index()][t.leapIndex()])
}

func (t *Time) resetWeekday() {
	t.wday = weekdayOf(t.Time().Weekday())
}

// isLeap reports whether the given Persian year is a leap year.
func isLeap(year int) bool {
	return divider(25*year+11, 33) < 8
}

// norm returns nhi, nlo such that
//
//	hi*base + lo == nhi*base + nlo
//	0 <= nlo < base
func norm(hi, lo, base int) (int, int) {
	if lo < 0 {
		n := (-lo-1)/base + 1
		hi -= n
		lo += n * base
	}

	if lo >= base {
		n := lo / base
		hi += n
		lo -= n * base
	}

	return hi, lo
}

// normDay is norm for one based units such as the day of month, where the
// valid range is [1, base] rather than [0, base).
func normDay(hi, lo, base int) (int, int) {
	if lo < 1 {
		n := (-lo-1)/base + 1
		hi -= n
		lo += n * base
	}

	if lo > base {
		n := lo / base
		hi += n
		lo -= n * base
	}

	return hi, lo
}

func clamp(value *int, minValue, maxValue int) {
	switch {
	case *value < minValue:
		*value = minValue
	case *value > maxValue:
		*value = maxValue
	}
}

func divider(num, den int) int {
	if num > 0 {
		return num % den
	}

	return num - ((((num + 1) / den) - 1) * den)
}
