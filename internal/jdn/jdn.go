// Package jdn converts calendar dates to and from the Julian Day Number, a
// continuous count of days that is independent of any calendar.
//
// Every conversion in this module passes through a JDN. Going
// Persian -> JDN -> Gregorian (and back) costs one extra step and keeps the
// two calendars exact across the 1582 Gregorian reform, because the reform is
// handled in exactly one place: dates before it are read in the Julian
// calendar, dates on or after it in the Gregorian calendar.
//
// The Gregorian formulas follow https://aa.usno.navy.mil/faq/JD_formula.
package jdn

// gregorianReform is the JDN of October 15, 1582, the first day of the
// Gregorian calendar.
const gregorianReform = 2299160

// FromGregorian returns the Julian Day Number of a Gregorian date, reading
// dates before the 1582 reform in the Julian calendar.
func FromGregorian(year, month, day int) int {
	if isAfterGregorianReform(year, month, day) {
		return fromGregorianPostReform(year, month, day)
	}

	return fromGregorianPreReform(year, month, day)
}

// ToGregorian returns the Gregorian date of a Julian Day Number, returning
// dates before the 1582 reform in the Julian calendar.
func ToGregorian(jdn int) (year, month, day int) {
	if jdn > gregorianReform {
		return toGregorianPostReform(jdn)
	}

	return toGregorianPreReform(jdn)
}

// isAfterGregorianReform reports whether the given Gregorian date falls on or
// after the reform of October 15, 1582.
func isAfterGregorianReform(year, month, day int) bool {
	return year > 1582 ||
		(year == 1582 && month > 10) ||
		(year == 1582 && month == 10 && day > 14)
}

// fromGregorianPostReform converts a Gregorian date on or after the reform
// into its Julian Day Number.
//
// The formula splits the work into four parts: the year is shifted by a large
// offset so that all intermediate values stay positive, the leap year factor
// counts whole four year cycles, the month factor spreads the 12 months over
// the year, and the century factor applies the Gregorian rule that century
// years are leap only when divisible by 400.
func fromGregorianPostReform(year, month, day int) int {
	const (
		// 1461 is the number of days in a four year Julian cycle (365.25 * 4).
		daysInFourYearCycle     = 1461
		yearOffset              = 4800
		centuryAdjustmentOffset = 4900
		monthCycleFactor        = 367
		baseDayAdjustment       = 32075
	)

	adjustedYear := year + yearOffset + ((month - 14) / 12)
	leapYearFactor := (daysInFourYearCycle * adjustedYear) / 4

	adjustedMonth := month - 2 - 12*((month-14)/12)
	monthFactor := (monthCycleFactor * adjustedMonth) / 12

	centuryFactor := (3 * ((year + centuryAdjustmentOffset + ((month - 14) / 12)) / 100)) / 4

	return leapYearFactor + monthFactor - centuryFactor + day - baseDayAdjustment
}

// fromGregorianPreReform converts a Julian calendar date, that is a date
// before the Gregorian reform of 1582, into its Julian Day Number.
//
// The Julian calendar has no century correction, so the leap year factor is a
// plain "one day every four years" adjustment.
func fromGregorianPreReform(year, month, day int) int {
	const (
		yearOffset        = 5001
		monthCycleFactor  = 275
		yearCycleFactor   = 367
		baseDayAdjustment = 1729777
	)

	adjustedYear := year + yearOffset + (month-9)/7
	leapYearFactor := (7 * adjustedYear) / 4
	monthFactor := (monthCycleFactor * month) / 9

	return yearCycleFactor*year - leapYearFactor + monthFactor + day + baseDayAdjustment
}

// toGregorianPostReform converts a Julian Day Number on or after the
// Gregorian reform into a Gregorian date.
func toGregorianPostReform(jdn int) (year, month, day int) {
	const (
		daysInFourYearCycle = 1461
		// 2447 drives the month/day split: 80/2447 approximates the average
		// month length once the year starts in March.
		daysInMonthMultiplier = 2447
		julianDayOffset       = 68569
		// 1461001 is the number of days in a 4000 year cycle (365.25 * 4000).
		daysIn4000YearCycle = 1461001
		// 146097 is the number of days in a 400 year Gregorian cycle.
		daysIn400YearCycle = 146097
	)

	offsetJDN := jdn + julianDayOffset

	century := 4 * offsetJDN / daysIn400YearCycle
	offsetJDN -= (daysIn400YearCycle*century + 3) / 4

	yearBase := 4000 * (offsetJDN + 1) / daysIn4000YearCycle
	offsetJDN = offsetJDN - daysInFourYearCycle*yearBase/4 + 31

	monthFactor := 80 * offsetJDN / daysInMonthMultiplier
	day = offsetJDN - daysInMonthMultiplier*monthFactor/80
	offsetJDN = monthFactor / 11
	month = monthFactor + 2 - 12*offsetJDN
	year = 100*(century-49) + yearBase + offsetJDN

	return year, month, day
}

// toGregorianPreReform converts a Julian Day Number before the Gregorian
// reform into a Julian calendar date.
func toGregorianPreReform(jdn int) (year, month, day int) {
	const (
		daysInFourYearCycle   = 1461
		daysInMonthMultiplier = 2447
		julianDayOffset       = 1402
		epochYearOffset       = 4716
	)

	offsetJDN := jdn + julianDayOffset

	quadrennialCycle := (offsetJDN - 1) / daysInFourYearCycle
	remainingDays := offsetJDN - daysInFourYearCycle*quadrennialCycle
	yearAdjustment := (remainingDays-1)/365 - remainingDays/daysInFourYearCycle
	dayOfYear := remainingDays - 365*yearAdjustment + 30

	monthFactor := 80 * dayOfYear / daysInMonthMultiplier
	day = dayOfYear - daysInMonthMultiplier*monthFactor/80
	yearFraction := monthFactor / 11
	month = monthFactor + 2 - 12*yearFraction
	year = 4*quadrennialCycle + yearAdjustment + yearFraction - epochYearOffset

	return year, month, day
}

// ToPersian converts a Julian Day Number into a Persian (Solar Hijri) date.
//
// The Persian calendar repeats on a 33 year cycle containing 8 leap years, so
// the year is recovered by peeling off whole 33 year cycles, then whole four
// year cycles, then the remainder. The first six months have 31 days and the
// rest have 30, which makes the month/day split a pair of divisions.
func ToPersian(jdn int) (year, month, day int) {
	const (
		julianDayToPersianOffset = 1365393
		daysIn33YearCycle        = 12053 // 33 * 365.24
		daysInFourYearCycle      = 1461
		daysInFirstHalf          = 186 // 6 * 31
		epochYearOffset          = 1595
		longMonthDays            = 31
		shortMonthDays           = 30
	)

	daysSinceEpoch := jdn - julianDayToPersianOffset

	cyclesOf33Years := daysSinceEpoch / daysIn33YearCycle
	year = -epochYearOffset + 33*cyclesOf33Years
	remainingDays := daysSinceEpoch % daysIn33YearCycle

	cyclesOf4Years := remainingDays / daysInFourYearCycle
	year += 4 * cyclesOf4Years
	remainingDays %= daysInFourYearCycle

	if remainingDays > 365 {
		year += (remainingDays - 1) / 365
		remainingDays = (remainingDays - 1) % 365
	}

	if remainingDays < daysInFirstHalf {
		month = 1 + remainingDays/longMonthDays
		day = 1 + remainingDays%longMonthDays
	} else {
		month = 7 + (remainingDays-daysInFirstHalf)/shortMonthDays
		day = 1 + (remainingDays-daysInFirstHalf)%shortMonthDays
	}

	return year, month, day
}

// FromPersian converts a Persian (Solar Hijri) date into its Julian Day Number.
func FromPersian(year, month, day int) int {
	const (
		persianToJulianOffset = 1365392
		leapYearCycle         = 33
		leapYearsPerCycle     = 8
		daysInFirstHalf       = 186 // 6 * 31
		longMonthDays         = 31
		shortMonthDays        = 30
		epochYearOffset       = 1595
	)

	adjustedYear := year + epochYearOffset

	// Leap days accumulated so far: 8 per full 33 year cycle, plus one per
	// completed four year step inside the current cycle.
	leapDays := (adjustedYear/leapYearCycle)*leapYearsPerCycle +
		((adjustedYear%leapYearCycle + 3) / 4)

	var dayOfYear int
	if month < 7 {
		dayOfYear = (month - 1) * longMonthDays
	} else {
		dayOfYear = (month-7)*shortMonthDays + daysInFirstHalf
	}

	return persianToJulianOffset + 365*adjustedYear + leapDays + dayOfYear + day
}
