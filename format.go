package ptime

import (
	"strconv"
	"strings"
	"time"
)

// Predefined layouts for use with [Time.Format] and [Parse].
const (
	// RFC3339 is the RFC 3339 profile of ISO 8601, e.g. "1394-07-02T12:59:59+03:30".
	RFC3339 = "yyyy-MM-ddTHH:mm:ssZ"
	// RFC3339Nano is RFC3339 with a fractional second, trailing zeros removed.
	RFC3339Nano = "yyyy-MM-ddTHH:mm:ss.999999999Z"
	// DateTime is a date and a clock reading, e.g. "1394-07-02 12:59:59".
	DateTime = "yyyy-MM-dd HH:mm:ss"
	// DateOnly is a date without a clock reading, e.g. "1394-07-02".
	DateOnly = "yyyy-MM-dd"
	// TimeOnly is a clock reading without a date, e.g. "12:59:59".
	TimeOnly = "HH:mm:ss"
	// Kitchen is a short 12-hour clock reading, e.g. "12:59 ب.ظ".
	Kitchen = "h:mm a"
	// LongDate spells out the weekday and month, e.g. "پنج‌شنبه 2 مهر 1394".
	LongDate = "E d MMM yyyy"
)

// String returns t in [RFC3339Nano] format.
//
// The zero Time has no calendar date to render and is reported as
// "0000-00-00T00:00:00Z".
func (t Time) String() string {
	if t.IsZero() {
		return "0000-00-00T00:00:00Z"
	}

	return t.Format(RFC3339Nano)
}

// Format returns the textual representation of t according to layout.
//
// The layout is built from the following pattern letters. Anything else is
// copied to the output verbatim.
//
//	yyyy, yyy, y  year (e.g. 1394)
//	yy            2-digit year (e.g. 94)
//	MMM           Persian month name (e.g. فروردین)
//	MMI           Dari month name (e.g. حمل)
//	MM            2-digit month (e.g. 01)
//	M             month (e.g. 1)
//	rw            remaining weeks of year
//	w             week of year
//	W             week of month
//	RD            remaining days of year
//	D             day of year
//	rd            remaining days of month
//	dd            2-digit day of month (e.g. 01)
//	d             day of month (e.g. 1)
//	E             Persian weekday name (e.g. شنبه)
//	e             short Persian weekday name (e.g. ش)
//	A             Persian 12-hour marker (e.g. قبل از ظهر)
//	a             short Persian 12-hour marker (e.g. ق.ظ)
//	HH            2-digit hour [00-23]
//	H             hour [0-23]
//	kk            2-digit hour [01-24]
//	k             hour [1-24]
//	hh            2-digit hour [01-12]
//	h             hour [1-12]
//	KK            2-digit hour [00-11]
//	K             hour [0-11]
//	mm            2-digit minute [00-59]
//	m             minute [0-59]
//	ss            2-digit second [00-59]
//	s             second [0-59]
//	n             part of the day (e.g. صبح)
//	ns            nanosecond, as a plain decimal number
//	S             3-digit millisecond (e.g. 001)
//	.000          fractional second, 3 digits
//	.000000       fractional second, 6 digits
//	.000000000    fractional second, 9 digits
//	.999          fractional second, 3 digits, trailing zeros removed
//	.999999       fractional second, 6 digits, trailing zeros removed
//	.999999999    fractional second, 9 digits, trailing zeros removed
//	z             time zone name (e.g. Asia/Tehran)
//	Z             time zone offset (e.g. +03:30)
//
// A ".999" style fraction is omitted entirely, dot included, when the
// fractional second is zero.
func (t Time) Format(layout string) string {
	if layout == "" {
		return ""
	}

	// The formatted value is usually longer than the layout itself, mostly
	// because Persian names are multi-byte.
	return string(t.AppendFormat(make([]byte, 0, 2*len(layout)), layout))
}

// AppendFormat is like [Time.Format] but appends the result to b and returns
// the extended buffer, avoiding a string allocation.
func (t Time) AppendFormat(b []byte, layout string) []byte {
	for i := 0; i < len(layout); {
		// peek returns the layout byte after the current one, or 0.
		peek := func() byte {
			if i+1 >= len(layout) {
				return 0
			}

			return layout[i+1]
		}

		rest := layout[i:]

		switch layout[i] {
		case 'A':
			b = append(b, t.AmPm().String()...)
			i++
		case 'D':
			b = appendInt(b, t.YearDay())
			i++
		case 'E':
			b = append(b, t.wday.String()...)
			i++
		case 'H':
			if peek() == 'H' {
				b = appendPadded(b, t.hour, 2)
				i += 2
			} else {
				b = appendInt(b, t.hour)
				i++
			}
		case 'K':
			if peek() == 'K' {
				b = appendPadded(b, t.Hour12(), 2)
				i += 2
			} else {
				b = appendInt(b, t.Hour12())
				i++
			}
		case 'M':
			switch {
			case strings.HasPrefix(rest, "MMM"):
				b = append(b, t.month.String()...)
				i += 3
			case strings.HasPrefix(rest, "MMI"):
				b = append(b, t.month.Dari()...)
				i += 3
			case peek() == 'M':
				b = appendPadded(b, int(t.month), 2)
				i += 2
			default:
				b = appendInt(b, int(t.month))
				i++
			}
		case 'R':
			if peek() == 'D' {
				b = appendInt(b, t.RYearDay())
				i += 2
			} else {
				b = append(b, 'R')
				i++
			}
		case 'S':
			b = appendPadded(b, t.nsec/1e6, 3)
			i++
		case 'W':
			b = appendInt(b, t.MonthWeek())
			i++
		case 'Z':
			b = append(b, t.ZoneOffset()...)
			i++
		case 'a':
			b = append(b, t.AmPm().Short()...)
			i++
		case 'd':
			if peek() == 'd' {
				b = appendPadded(b, t.day, 2)
				i += 2
			} else {
				b = appendInt(b, t.day)
				i++
			}
		case 'e':
			b = append(b, t.wday.Short()...)
			i++
		case 'h':
			if peek() == 'h' {
				b = appendPadded(b, orMax(t.Hour12(), 12), 2)
				i += 2
			} else {
				b = appendInt(b, orMax(t.Hour12(), 12))
				i++
			}
		case 'k':
			if peek() == 'k' {
				b = appendPadded(b, orMax(t.hour, 24), 2)
				i += 2
			} else {
				b = appendInt(b, orMax(t.hour, 24))
				i++
			}
		case 'm':
			if peek() == 'm' {
				b = appendPadded(b, t.minute, 2)
				i += 2
			} else {
				b = appendInt(b, t.minute)
				i++
			}
		case 'n':
			if peek() == 's' {
				b = appendInt(b, t.nsec)
				i += 2
			} else {
				b = append(b, t.DayTime().String()...)
				i++
			}
		case 'r':
			switch peek() {
			case 'w':
				b = appendInt(b, t.RYearWeek())
				i += 2
			case 'd':
				b = appendInt(b, t.RMonthDay())
				i += 2
			default:
				b = append(b, 'r')
				i++
			}
		case 's':
			if peek() == 's' {
				b = appendPadded(b, t.sec, 2)
				i += 2
			} else {
				b = appendInt(b, t.sec)
				i++
			}
		case 'w':
			b = appendInt(b, t.YearWeek())
			i++
		case 'y':
			switch {
			case strings.HasPrefix(rest, "yyyy"):
				b = appendPadded(b, t.year, 4)
				i += 4
			case strings.HasPrefix(rest, "yyy"):
				b = appendPadded(b, t.year, 4)
				i += 3
			case peek() == 'y':
				b = appendTwoDigitYear(b, t.year)
				i += 2
			default:
				b = appendPadded(b, t.year, 4)
				i++
			}
		case 'z':
			b = append(b, t.locationName()...)
			i++
		case '.':
			if digits, trim, width := fractionLayout(rest); width > 0 {
				b = appendFraction(b, t.nsec, digits, trim)
				i += width

				continue
			}

			b = append(b, '.')
			i++
		default:
			b = append(b, layout[i])
			i++
		}
	}

	return b
}

// TimeFormat returns the textual representation of t according to a standard
// library reference-time layout, as used by [time.Time.Format].
//
//	2006        4-digit year (e.g. 1394)
//	06          2-digit year (e.g. 94)
//	01          2-digit month (e.g. 01)
//	1           month (e.g. 1)
//	Jan         month name (e.g. مهر)
//	January     month name (e.g. مهر)
//	02          2-digit day of month (e.g. 07)
//	2           day of month (e.g. 7)
//	_2          day of month, space padded to 2 characters (e.g. " 7")
//	Mon         short weekday name (e.g. ش)
//	Monday      weekday name (e.g. شنبه)
//	Morning     part of the day (e.g. صبح)
//	03          2-digit hour [01-12]
//	3           hour [1-12]
//	15          2-digit hour [00-23]
//	04          2-digit minute
//	4           minute
//	05          2-digit second
//	5           second
//	.000        fractional second, 3 digits
//	.000000     fractional second, 6 digits
//	.000000000  fractional second, 9 digits
//	.999        fractional second, 3 digits, trailing zeros removed
//	.999999     fractional second, 6 digits, trailing zeros removed
//	.999999999  fractional second, 9 digits, trailing zeros removed
//	PM          12-hour marker (e.g. قبل از ظهر)
//	pm          short 12-hour marker (e.g. ق.ظ)
//	MST         time zone name
//	-0700       time zone offset (e.g. +0330)
//	-07         time zone offset (e.g. +03)
//	-07:00      time zone offset (e.g. +03:30)
//	Z0700       time zone offset (e.g. +0330)
//	Z07:00      time zone offset (e.g. +03:30)
//
// Month names are Dari when the location of t is [Afghanistan] and Iranian
// Persian otherwise.
func (t Time) TimeFormat(layout string) string {
	if layout == "" {
		return ""
	}

	return string(t.AppendTimeFormat(make([]byte, 0, 2*len(layout)), layout))
}

// AppendTimeFormat is like [Time.TimeFormat] but appends the result to b and
// returns the extended buffer.
func (t Time) AppendTimeFormat(b []byte, layout string) []byte {
	for i := 0; i < len(layout); {
		rest := layout[i:]

		if digits, trim, width := fractionLayout(rest); width > 0 {
			b = appendFraction(b, t.nsec, digits, trim)
			i += width

			continue
		}

		token, width := nextStdToken(rest)
		if width == 0 {
			b = append(b, layout[i])
			i++

			continue
		}

		b = t.appendStdToken(b, token)
		i += width
	}

	return b
}

// ZoneOffset returns the time zone offset of t as [+|-]HH:mm.
//
// An optional layout selects a different shape; it must be one of "-0700",
// "-07", "-07:00", "Z0700" or "Z07:00". The "Z" forms render UTC as "Z".
func (t Time) ZoneOffset(layout ...string) string {
	format := "-07:00"

	if len(layout) > 0 {
		switch layout[0] {
		case "-0700", "-07", "-07:00", "Z0700", "Z07:00":
			format = layout[0]
		}
	}

	_, offset := t.Zone()

	if offset == 0 {
		switch format {
		case "-0700":
			return "+0000"
		case "-07":
			return "+00"
		case "Z0700", "Z07:00":
			return "Z"
		default:
			return "+00:00"
		}
	}

	sign := byte('+')
	if offset < 0 {
		sign = '-'
		offset = -offset
	}

	hours, minutes := offset/3600, (offset%3600)/60

	b := make([]byte, 0, 6)
	b = append(b, sign)
	b = appendPadded(b, hours, 2)

	switch format {
	case "-07":
		return string(b)
	case "-0700", "Z0700":
		return string(appendPadded(b, minutes, 2))
	default:
		return string(appendPadded(append(b, ':'), minutes, 2))
	}
}

// appendStdToken renders a single reference-time token.
func (t Time) appendStdToken(b []byte, token string) []byte {
	switch token {
	case "2006":
		return appendPadded(b, t.year, 4)
	case "06":
		return appendTwoDigitYear(b, t.year)
	case "January", "Jan":
		return append(b, t.localizedMonthName()...)
	case "01":
		return appendPadded(b, int(t.month), 2)
	case "1":
		return appendInt(b, int(t.month))
	case "02":
		return appendPadded(b, t.day, 2)
	case "_2":
		if t.day < 10 {
			b = append(b, ' ')
		}

		return appendInt(b, t.day)
	case "2":
		return appendInt(b, t.day)
	case "Monday":
		return append(b, t.wday.String()...)
	case "Mon":
		return append(b, t.wday.Short()...)
	case "Morning":
		return append(b, t.DayTime().String()...)
	case "15":
		return appendPadded(b, t.hour, 2)
	case "03":
		return appendPadded(b, t.Hour12(), 2)
	case "3":
		return appendInt(b, t.Hour12())
	case "04":
		return appendPadded(b, t.minute, 2)
	case "4":
		return appendInt(b, t.minute)
	case "05":
		return appendPadded(b, t.sec, 2)
	case "5":
		return appendInt(b, t.sec)
	case "PM":
		return append(b, t.AmPm().String()...)
	case "pm":
		return append(b, t.AmPm().Short()...)
	case "MST":
		return append(b, t.locationName()...)
	case "-0700", "-07:00", "-07", "Z0700", "Z07:00":
		return append(b, t.ZoneOffset(token)...)
	default:
		return append(b, token...)
	}
}

// localizedMonthName returns the Dari month name in Afghanistan and the
// Iranian Persian name everywhere else.
func (t Time) localizedMonthName() string {
	if t.locationName() == Afghanistan().String() {
		return t.month.Dari()
	}

	return t.month.String()
}

func (t Time) locationName() string {
	if t.loc == nil {
		//nolint:gosmopolitan
		return time.Local.String()
	}

	return t.loc.String()
}

// stdTokens lists the reference-time tokens, longest first so that a linear
// scan always finds the longest match at the current position.
var stdTokens = []string{
	"January", "Morning", "Monday",
	"Z07:00", "-07:00",
	"Z0700", "-0700",
	"2006",
	"Jan", "Mon", "MST", "-07",
	"PM", "pm", "15", "06", "01", "02", "03", "04", "05", "_2",
	"1", "2", "3", "4", "5",
}

// nextStdToken returns the reference-time token at the start of layout and its
// length, or a zero length when no token matches.
func nextStdToken(layout string) (string, int) {
	for _, token := range stdTokens {
		if strings.HasPrefix(layout, token) {
			return token, len(token)
		}
	}

	return "", 0
}

// fractionLayout recognizes a ".000"/".999" style fractional second at the
// start of layout. It returns the number of digits, whether trailing zeros are
// removed, and the length of the matched layout, or a zero length when the
// layout does not start with a fraction.
func fractionLayout(layout string) (digits int, trim bool, width int) {
	if len(layout) < 4 || layout[0] != '.' {
		return 0, false, 0
	}

	var pad byte

	switch layout[1] {
	case '0':
		pad = '0'
	case '9':
		pad, trim = '9', true
	default:
		return 0, false, 0
	}

	for digits = 1; digits < len(layout)-1 && layout[digits+1] == pad; digits++ {
		continue
	}

	switch digits {
	case 3, 6, 9:
		return digits, trim, digits + 1
	default:
		return 0, false, 0
	}
}

// appendFraction writes a fractional second of the given number of digits.
// When trim is set, trailing zeros are removed and an all zero fraction is
// omitted entirely, dot included.
func appendFraction(b []byte, nsec, digits int, trim bool) []byte {
	value := nsec
	for i := 9; i > digits; i-- {
		value /= 10
	}

	if trim {
		for digits > 0 && value%10 == 0 {
			value /= 10
			digits--
		}

		if digits == 0 {
			return b
		}
	}

	return appendPadded(append(b, '.'), value, digits)
}

// appendPadded writes v in decimal, left padded with zeros to width digits.
// Values wider than width are written in full.
func appendPadded(b []byte, v, width int) []byte {
	if v < 0 {
		return appendInt(b, v)
	}

	for p := pow10(width - 1); p > 1 && v < p; p /= 10 {
		b = append(b, '0')
	}

	return appendInt(b, v)
}

// appendTwoDigitYear writes the last two digits of year, zero padded.
func appendTwoDigitYear(b []byte, year int) []byte {
	if year < 0 {
		return appendInt(b, year)
	}

	return appendPadded(b, year%100, 2)
}

func appendInt(b []byte, v int) []byte {
	return strconv.AppendInt(b, int64(v), 10)
}

func pow10(n int) int {
	p := 1
	for ; n > 0; n-- {
		p *= 10
	}

	return p
}

// orMax maps hour 0 onto the top of a one based clock, so that 0 reads as 12
// on a 12-hour clock and as 24 on a 24-hour clock.
func orMax(value, maxValue int) int {
	if value == 0 {
		return maxValue
	}

	return value
}
