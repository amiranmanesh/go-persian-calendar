package ptime

import (
	"strings"
	"time"
)

// A ParseError describes a problem parsing a time string.
type ParseError struct {
	// Layout is the layout that was given to the parse function.
	Layout string
	// Value is the value that was given to the parse function.
	Value string
	// LayoutElem is the layout element that failed to match.
	LayoutElem string
	// ValueElem is the remainder of the value at the point of failure.
	ValueElem string
	// Message describes the failure when it is not a plain mismatch.
	Message string
}

// Error implements the error interface.
func (e *ParseError) Error() string {
	var sb strings.Builder

	sb.WriteString(`ptime: cannot parse "`)
	sb.WriteString(e.ValueElem)
	sb.WriteString(`" as "`)
	sb.WriteString(e.LayoutElem)
	sb.WriteString(`" in "`)
	sb.WriteString(e.Value)
	sb.WriteString(`"`)

	if e.Message != "" {
		sb.WriteString(": ")
		sb.WriteString(e.Message)
	}

	return sb.String()
}

// Parse parses a formatted Persian date string using the layout language of
// [Time.Format].
//
// Fields absent from the layout default to Farvardin 1 of year 1 at 00:00:00.
// Computed fields such as the day of year or the week of month are matched and
// discarded, since they carry no information the other fields do not.
//
// In the absence of a time zone in the value, Parse returns a time in UTC. Use
// [ParseInLocation] to pick a different default.
func Parse(layout, value string) (Time, error) {
	return ParseInLocation(layout, value, time.UTC)
}

// ParseInLocation is like [Parse] but interprets a value without a time zone as
// being in the given location.
func ParseInLocation(layout, value string, loc *time.Location) (Time, error) {
	p := newParser(layout, value)
	if err := p.parse(); err != nil {
		return Time{}, err
	}

	return p.result(loc)
}

// ParseTimeFormat parses a formatted Persian date string using the standard
// library reference-time layout language of [Time.TimeFormat].
//
// In the absence of a time zone in the value, ParseTimeFormat returns a time in
// UTC. Use [ParseTimeFormatInLocation] to pick a different default.
func ParseTimeFormat(layout, value string) (Time, error) {
	return ParseTimeFormatInLocation(layout, value, time.UTC)
}

// ParseTimeFormatInLocation is like [ParseTimeFormat] but interprets a value
// without a time zone as being in the given location.
func ParseTimeFormatInLocation(layout, value string, loc *time.Location) (Time, error) {
	p := newParser(layout, value)
	if err := p.parseTimeFormat(); err != nil {
		return Time{}, err
	}

	return p.result(loc)
}

// clock12 records which 12-hour convention a layout used, so that an AM/PM
// marker can be applied correctly.
type clock12 int

const (
	clock12None  clock12 = iota
	clock12OneTo         // "hh"/"h" and "03"/"3": 12 means midnight or noon
	clock12Zero          // "KK"/"K": 0 means midnight or noon
)

type parser struct {
	layout string
	value  string

	rest string // unconsumed part of value

	year   int
	month  Month
	day    int
	hour   int
	minute int
	sec    int
	nsec   int

	hour12      int
	clock12Kind clock12
	pm          bool
	hasAmPm     bool

	zoneOffset  int
	hasZone     bool
	zoneName    string
	hasZoneName bool
}

func newParser(layout, value string) *parser {
	return &parser{
		layout: layout,
		value:  value,
		rest:   value,
		year:   1,
		month:  Farvardin,
		day:    1,
	}
}

// parse walks a [Time.Format] style layout.
func (p *parser) parse() error {
	layout := p.layout

	for i := 0; i < len(layout); {
		peek := func() byte {
			if i+1 >= len(layout) {
				return 0
			}

			return layout[i+1]
		}

		rest := layout[i:]

		var err error

		switch layout[i] {
		case 'A':
			err = p.readAmPm(amPmNames[:], "A")
			i++
		case 'a':
			err = p.readAmPm(shortAmPmNames[:], "a")
			i++
		case 'E':
			_, err = p.readName(weekdays[:], "E")
			i++
		case 'e':
			_, err = p.readName(shortWeekdays[:], "e")
			i++
		case 'n':
			if peek() == 's' {
				p.nsec, err = p.readNum("ns", 1, 9)
				i += 2
			} else {
				_, err = p.readName(dayTimeNames[:], "n")
				i++
			}
		case 'D':
			_, err = p.readNum("D", 1, 3)
			i++
		case 'W':
			_, err = p.readNum("W", 1, 2)
			i++
		case 'w':
			_, err = p.readNum("w", 1, 2)
			i++
		case 'R':
			if peek() == 'D' {
				_, err = p.readNum("RD", 1, 3)
				i += 2
			} else {
				err = p.readLiteral("R")
				i++
			}
		case 'r':
			switch peek() {
			case 'w':
				_, err = p.readNum("rw", 1, 2)
				i += 2
			case 'd':
				_, err = p.readNum("rd", 1, 2)
				i += 2
			default:
				err = p.readLiteral("r")
				i++
			}
		case 'H':
			if peek() == 'H' {
				p.hour, err = p.readNum("HH", 2, 2)
				i += 2
			} else {
				p.hour, err = p.readNum("H", 1, 2)
				i++
			}
		case 'k':
			var h int

			if peek() == 'k' {
				h, err = p.readNum("kk", 2, 2)
				i += 2
			} else {
				h, err = p.readNum("k", 1, 2)
				i++
			}

			p.hour = h % 24
		case 'h':
			p.clock12Kind = clock12OneTo

			if peek() == 'h' {
				p.hour12, err = p.readNum("hh", 2, 2)
				i += 2
			} else {
				p.hour12, err = p.readNum("h", 1, 2)
				i++
			}
		case 'K':
			p.clock12Kind = clock12Zero

			if peek() == 'K' {
				p.hour12, err = p.readNum("KK", 2, 2)
				i += 2
			} else {
				p.hour12, err = p.readNum("K", 1, 2)
				i++
			}
		case 'M':
			i, err = p.parseMonth(rest, i, peek())
		case 'd':
			if peek() == 'd' {
				p.day, err = p.readNum("dd", 2, 2)
				i += 2
			} else {
				p.day, err = p.readNum("d", 1, 2)
				i++
			}
		case 'm':
			if peek() == 'm' {
				p.minute, err = p.readNum("mm", 2, 2)
				i += 2
			} else {
				p.minute, err = p.readNum("m", 1, 2)
				i++
			}
		case 's':
			if peek() == 's' {
				p.sec, err = p.readNum("ss", 2, 2)
				i += 2
			} else {
				p.sec, err = p.readNum("s", 1, 2)
				i++
			}
		case 'S':
			var ms int

			ms, err = p.readNum("S", 3, 3)
			p.nsec = ms * 1e6
			i++
		case 'y':
			i, err = p.parseYear(rest, i, peek())
		case 'Z':
			err = p.readZoneOffset("Z")
			i++
		case 'z':
			err = p.readZoneName("z")
			i++
		case '.':
			if digits, _, width := fractionLayout(rest); width > 0 {
				err = p.readFraction(rest[:width], digits)
				i += width
			} else {
				err = p.readLiteral(".")
				i++
			}
		default:
			r := decodeRune(rest)
			err = p.readLiteral(r)
			i += len(r)
		}

		if err != nil {
			return err
		}
	}

	return p.checkFullyConsumed()
}

// parseTimeFormat walks a [Time.TimeFormat] style layout.
func (p *parser) parseTimeFormat() error {
	layout := p.layout

	for i := 0; i < len(layout); {
		rest := layout[i:]

		if digits, _, width := fractionLayout(rest); width > 0 {
			if err := p.readFraction(rest[:width], digits); err != nil {
				return err
			}

			i += width

			continue
		}

		token, width := nextStdToken(rest)
		if width == 0 {
			r := decodeRune(rest)
			if err := p.readLiteral(r); err != nil {
				return err
			}

			i += len(r)

			continue
		}

		if err := p.readStdToken(token); err != nil {
			return err
		}

		i += width
	}

	return p.checkFullyConsumed()
}

func (p *parser) readStdToken(token string) error {
	var err error

	switch token {
	case "2006":
		p.year, err = p.readNum(token, 1, 4)
	case "06":
		var y int

		if y, err = p.readNum(token, 2, 2); err == nil {
			p.year = expandTwoDigitYear(y)
		}
	case "January", "Jan":
		err = p.readMonthName(token)
	case "01":
		var m int

		if m, err = p.readNum(token, 2, 2); err == nil {
			p.month = Month(m)
		}
	case "1":
		var m int

		if m, err = p.readNum(token, 1, 2); err == nil {
			p.month = Month(m)
		}
	case "02":
		p.day, err = p.readNum(token, 2, 2)
	case "2":
		p.day, err = p.readNum(token, 1, 2)
	case "_2":
		p.rest = strings.TrimPrefix(p.rest, " ")
		p.day, err = p.readNum(token, 1, 2)
	case "Monday":
		_, err = p.readName(weekdays[:], token)
	case "Mon":
		_, err = p.readName(shortWeekdays[:], token)
	case "Morning":
		_, err = p.readName(dayTimeNames[:], token)
	case "15":
		p.hour, err = p.readNum(token, 2, 2)
	case "03":
		p.clock12Kind = clock12OneTo
		p.hour12, err = p.readNum(token, 2, 2)
	case "3":
		p.clock12Kind = clock12OneTo
		p.hour12, err = p.readNum(token, 1, 2)
	case "04":
		p.minute, err = p.readNum(token, 2, 2)
	case "4":
		p.minute, err = p.readNum(token, 1, 2)
	case "05":
		p.sec, err = p.readNum(token, 2, 2)
	case "5":
		p.sec, err = p.readNum(token, 1, 2)
	case "PM":
		err = p.readAmPm(amPmNames[:], token)
	case "pm":
		err = p.readAmPm(shortAmPmNames[:], token)
	case "MST":
		err = p.readZoneName(token)
	case "-0700", "-07:00", "-07", "Z0700", "Z07:00":
		err = p.readZoneOffset(token)
	default:
		err = p.readLiteral(token)
	}

	return err
}

func (p *parser) parseMonth(rest string, i int, peek byte) (int, error) {
	switch {
	case strings.HasPrefix(rest, "MMM"):
		return i + 3, p.readMonthNames(months[:], "MMM")
	case strings.HasPrefix(rest, "MMI"):
		return i + 3, p.readMonthNames(dariMonths[:], "MMI")
	case peek == 'M':
		m, err := p.readNum("MM", 2, 2)
		p.month = Month(m)

		return i + 2, err
	default:
		m, err := p.readNum("M", 1, 2)
		p.month = Month(m)

		return i + 1, err
	}
}

func (p *parser) parseYear(rest string, i int, peek byte) (int, error) {
	switch {
	case strings.HasPrefix(rest, "yyyy"):
		y, err := p.readNum("yyyy", 1, 4)
		p.year = y

		return i + 4, err
	case strings.HasPrefix(rest, "yyy"):
		y, err := p.readNum("yyy", 1, 4)
		p.year = y

		return i + 3, err
	case peek == 'y':
		y, err := p.readNum("yy", 2, 2)
		p.year = expandTwoDigitYear(y)

		return i + 2, err
	default:
		y, err := p.readNum("y", 1, 4)
		p.year = y

		return i + 1, err
	}
}

// result validates the accumulated fields and assembles the Time.
func (p *parser) result(defaultLoc *time.Location) (Time, error) {
	if defaultLoc == nil {
		defaultLoc = time.UTC
	}

	loc, err := p.location(defaultLoc)
	if err != nil {
		return Time{}, err
	}

	hour, err := p.resolveHour()
	if err != nil {
		return Time{}, err
	}

	if !p.month.IsValid() {
		return Time{}, p.rangeError("month", p.month.String())
	}

	if p.day < 1 || p.day > daysInMonth(p.year, p.month) {
		return Time{}, p.rangeError("day of month", p.value)
	}

	switch {
	case p.minute > 59:
		return Time{}, p.rangeError("minute", p.value)
	case p.sec > 59:
		return Time{}, p.rangeError("second", p.value)
	case p.nsec > 999999999:
		return Time{}, p.rangeError("nanosecond", p.value)
	}

	return Date(p.year, p.month, p.day, hour, p.minute, p.sec, p.nsec, loc), nil
}

func (p *parser) location(defaultLoc *time.Location) (*time.Location, error) {
	if p.hasZoneName {
		loc, err := time.LoadLocation(p.zoneName)
		if err != nil {
			if !p.hasZone {
				return nil, &ParseError{
					Layout: p.layout, Value: p.value,
					LayoutElem: "z", ValueElem: p.zoneName,
					Message: "unknown time zone",
				}
			}

			return time.FixedZone(p.zoneName, p.zoneOffset), nil
		}

		return loc, nil
	}

	if p.hasZone {
		if p.zoneOffset == 0 {
			return time.UTC, nil
		}

		return time.FixedZone("", p.zoneOffset), nil
	}

	return defaultLoc, nil
}

func (p *parser) resolveHour() (int, error) {
	hour, err := p.hourOnClock()
	if err != nil {
		return 0, err
	}

	if hour < 0 || hour > 23 {
		return 0, p.rangeError("hour", p.value)
	}

	return hour, nil
}

// hourOnClock folds a 12-hour reading and its AM/PM marker into a 24-hour one.
func (p *parser) hourOnClock() (int, error) {
	switch p.clock12Kind {
	case clock12None:
		return p.hour, nil

	case clock12OneTo: // 12 reads as 0
		if p.hour12 < 1 || p.hour12 > 12 {
			return 0, p.rangeError("hour", p.value)
		}

		return p.applyPm(p.hour12 % 12), nil

	case clock12Zero: // 0 reads as 0
		if p.hour12 > 11 {
			return 0, p.rangeError("hour", p.value)
		}

		return p.applyPm(p.hour12), nil

	default:
		return p.hour, nil
	}
}

func (p *parser) applyPm(hour int) int {
	if p.hasAmPm && p.pm {
		return hour + 12
	}

	return hour
}

func (p *parser) readNum(elem string, minDigits, maxDigits int) (int, error) {
	n, i := 0, 0

	for i < len(p.rest) && i < maxDigits && p.rest[i] >= '0' && p.rest[i] <= '9' {
		n = n*10 + int(p.rest[i]-'0')
		i++
	}

	if i < minDigits {
		return 0, p.mismatch(elem)
	}

	p.rest = p.rest[i:]

	return n, nil
}

func (p *parser) readName(names []string, elem string) (int, error) {
	for i, name := range names {
		if strings.HasPrefix(p.rest, name) {
			p.rest = p.rest[len(name):]

			return i, nil
		}
	}

	return 0, p.mismatch(elem)
}

func (p *parser) readMonthNames(names []string, elem string) error {
	i, err := p.readName(names, elem)
	if err != nil {
		return err
	}

	p.month = Month(i + 1)

	return nil
}

// readMonthName accepts either the Iranian or the Dari name, since a
// reference-time layout does not distinguish them.
func (p *parser) readMonthName(elem string) error {
	if err := p.readMonthNames(months[:], elem); err == nil {
		return nil
	}

	return p.readMonthNames(dariMonths[:], elem)
}

func (p *parser) readAmPm(names []string, elem string) error {
	i, err := p.readName(names, elem)
	if err != nil {
		return err
	}

	p.pm = i == int(Pm)
	p.hasAmPm = true

	return nil
}

func (p *parser) readFraction(elem string, digits int) error {
	if !strings.HasPrefix(p.rest, ".") {
		// A trailing-zeros-removed fraction is omitted when it is zero.
		if strings.HasSuffix(elem, "9") {
			return nil
		}

		return p.mismatch(elem)
	}

	rest := p.rest[1:]

	value, read := 0, 0
	for read < len(rest) && read < digits && isDigit(rest[read]) {
		value = value*10 + int(rest[read]-'0')
		read++
	}

	if read == 0 {
		return p.mismatch(elem)
	}

	p.rest = rest[read:]

	// Scale the digits that were actually present up to nanoseconds.
	p.nsec = value * pow10(9-read)

	return nil
}

func (p *parser) readZoneOffset(elem string) error {
	if strings.HasPrefix(p.rest, "Z") {
		p.rest = p.rest[1:]
		p.zoneOffset = 0
		p.hasZone = true

		return nil
	}

	if p.rest == "" || (p.rest[0] != '+' && p.rest[0] != '-') {
		return p.mismatch(elem)
	}

	sign := 1
	if p.rest[0] == '-' {
		sign = -1
	}

	p.rest = p.rest[1:]

	hours, err := p.readNum(elem, 2, 2)
	if err != nil {
		return err
	}

	p.rest = strings.TrimPrefix(p.rest, ":")

	minutes := 0
	if len(p.rest) >= 2 && isDigit(p.rest[0]) && isDigit(p.rest[1]) {
		if minutes, err = p.readNum(elem, 2, 2); err != nil {
			return err
		}
	}

	if hours > 23 || minutes > 59 {
		return p.rangeError("time zone offset", elem)
	}

	p.zoneOffset = sign * (hours*3600 + minutes*60)
	p.hasZone = true

	return nil
}

func (p *parser) readZoneName(elem string) error {
	i := 0
	for i < len(p.rest) && isZoneNameByte(p.rest[i]) {
		i++
	}

	if i == 0 {
		return p.mismatch(elem)
	}

	p.zoneName = p.rest[:i]
	p.rest = p.rest[i:]
	p.hasZoneName = true

	return nil
}

func (p *parser) readLiteral(lit string) error {
	if !strings.HasPrefix(p.rest, lit) {
		return p.mismatch(lit)
	}

	p.rest = p.rest[len(lit):]

	return nil
}

func (p *parser) checkFullyConsumed() error {
	if p.rest != "" {
		return &ParseError{
			Layout: p.layout, Value: p.value,
			LayoutElem: "", ValueElem: p.rest,
			Message: "extra text",
		}
	}

	return nil
}

func (p *parser) mismatch(elem string) error {
	return &ParseError{
		Layout: p.layout, Value: p.value,
		LayoutElem: elem, ValueElem: p.rest,
	}
}

func (p *parser) rangeError(what, elem string) error {
	return &ParseError{
		Layout: p.layout, Value: p.value,
		LayoutElem: elem, ValueElem: p.value,
		Message: what + " out of range",
	}
}

// expandTwoDigitYear maps a two digit Persian year onto the 1350-1449 window,
// which covers every year the calendar is realistically used for.
func expandTwoDigitYear(y int) int {
	if y < 50 {
		return 1400 + y
	}

	return 1300 + y
}

func daysInMonth(year int, month Month) int {
	leap := 0
	if isLeap(year) {
		leap = 1
	}

	return pMonthCount[month.index()][leap]
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isZoneNameByte(c byte) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		return true
	case c == '/' || c == '_' || c == '+' || c == '-':
		return true
	default:
		return false
	}
}

// decodeRune returns the first rune of s as a string, so that multi-byte
// literals in a layout are compared as a unit.
func decodeRune(s string) string {
	for i := 1; i <= len(s); i++ {
		if i == len(s) || isRuneStart(s[i]) {
			return s[:i]
		}
	}

	return s
}

func isRuneStart(c byte) bool {
	return c&0xC0 != 0x80
}
