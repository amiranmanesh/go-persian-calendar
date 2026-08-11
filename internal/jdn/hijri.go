package jdn

// The Hijri (Islamic lunar) conversions below implement the *tabular* civil
// calendar, sometimes called the Kuwaiti algorithm: a 30 year cycle of 354 and
// 355 day years, with the 11 leap years falling at fixed positions in the
// cycle.
//
// The tabular calendar is an arithmetic approximation. Iran determines the
// start of each lunar month by actual moon sighting, which can differ from the
// table by a day in either direction. Callers that need the officially
// announced date must correct the result rather than trust it — see the
// holiday package, which treats an uncorrected lunar date as an estimate.

// hijriEpoch is the JDN of 1 Muharram 1 AH in the tabular calendar,
// corresponding to 16 July 622 CE in the Julian calendar.
const hijriEpoch = 1948440

// FromHijri returns the Julian Day Number of a tabular Hijri date.
func FromHijri(year, month, day int) int {
	// ceil(29.5 * (month-1)) spreads alternating 30 and 29 day months over the
	// year; (3+11*year)/30 counts the leap days elapsed in the 30 year cycle.
	monthDays := (59*(month-1) + 1) / 2
	leapDays := (3 + 11*year) / 30

	return day + monthDays + (year-1)*354 + leapDays + hijriEpoch - 1
}

// ToHijri returns the tabular Hijri date of a Julian Day Number.
func ToHijri(jdn int) (year, month, day int) {
	// Shift into the cycle-relative domain, then peel off whole 30 year
	// cycles (10631 days each) before resolving the year within the cycle.
	n := jdn - hijriEpoch + 10632
	cycles := (n - 1) / 10631
	n = n - 10631*cycles + 354

	inCycle := ((10985-n)/5316)*((50*n)/17719) + (n/5670)*((43*n)/15238)
	n = n - ((30-inCycle)/15)*((17719*inCycle)/50) - (inCycle/16)*((15238*inCycle)/43) + 29

	month = (24 * n) / 709
	day = n - (709*month)/24
	year = 30*cycles + inCycle - 30

	return year, month, day
}

// IsHijriLeap reports whether a tabular Hijri year has 355 days, which makes
// Dhu al-Hijjah 30 days long instead of 29.
func IsHijriLeap(year int) bool {
	return (11*year+14)%30 < 11
}

// HijriMonthDays returns the number of days in a tabular Hijri month.
func HijriMonthDays(year, month int) int {
	switch {
	case month%2 == 1: // odd months have 30 days
		return 30
	case month == 12 && IsHijriLeap(year): // Dhu al-Hijjah gains the leap day
		return 30
	default:
		return 29
	}
}
