// Package holiday answers whether a Persian date is a public holiday, and what
// is commemorated on it.
//
// # How a date is resolved
//
// Iranian public holidays come in three kinds, and each needs a different
// treatment:
//
//   - Days fixed in the Persian calendar — Nowruz, 22 Bahman, 13 Farvardin.
//     These are set by law and never move, so they are rules in code and are
//     always [Confirmed].
//   - Days fixed in the Hijri lunar calendar — Eid al-Fitr, Ashura, Arbaeen.
//     Their Persian date is computed through the tabular Hijri calendar, which
//     is an arithmetic approximation of a calendar Iran determines by moon
//     sighting. A computed occurrence is [Estimated] until the official date
//     for that year is recorded, at which point it becomes [Confirmed].
//   - One-off days — an air pollution closure, an election, a day of mourning.
//     These cannot be predicted at all and exist only as data.
//
// The consequence for callers: check [Event.Confidence] before acting on a
// lunar holiday that has real consequences, such as a payroll run or an SLA
// calculation. [Calendar.ConfirmedThrough] reports the last year whose official
// calendar has been reconciled.
//
// # Keeping the data current
//
// The bundled data is embedded at build time, so this package performs no I/O
// and works offline. Upgrading the module brings newer data. A service that
// must pick up a correction without rebuilding can load a fresher copy of the
// same file at runtime with [Load] and [Calendar.WithOverrides]; the current
// file is published at
// https://amiranmanesh.github.io/go-persian-calendar/data/v1/iran.json
package holiday

import (
	"slices"
	"sort"
	"sync"

	ptime "github.com/amiranmanesh/go-persian-calendar"
	"github.com/amiranmanesh/go-persian-calendar/internal/jdn"
)

// A Kind describes which calendar fixes an event in place.
type Kind int

const (
	// Solar events fall on the same Persian date every year.
	Solar Kind = iota
	// Lunar events fall on the same Hijri date every year, so their Persian
	// date moves by about eleven days each year.
	Lunar
	// Special events belong to a single year and come from the data file.
	Special
)

// String returns the name of the kind.
func (k Kind) String() string {
	switch k {
	case Solar:
		return "solar"
	case Lunar:
		return "lunar"
	case Special:
		return "special"
	default:
		return "unknown"
	}
}

// A Confidence says how much to trust the date of an occurrence.
type Confidence int

const (
	// Confirmed means the date comes from the official calendar, either
	// because it is fixed by law or because the announced date was recorded.
	Confirmed Confidence = iota
	// Estimated means the date was computed from the tabular Hijri calendar
	// and may be off by a day. Do not settle money on it.
	Estimated
)

// String returns the name of the confidence level.
func (c Confidence) String() string {
	if c == Confirmed {
		return "confirmed"
	}

	return "estimated"
}

// An Event is something commemorated on a date.
type Event struct {
	// ID is a stable identifier for recurring events, such as "nowruz" or
	// "eid-fitr". It is empty for one-off events.
	ID string
	// Title is the Persian name of the event.
	Title string
	// Holiday reports whether the event makes the day an official day off.
	// Many events are commemorated without closing anything.
	Holiday bool
	// Kind says which calendar fixes the event.
	Kind Kind
	// Confidence says how much to trust the date.
	Confidence Confidence
}

// A Day is the result of looking up a date.
type Day struct {
	// Date is the date that was looked up, with its clock reading reset.
	Date ptime.Time
	// Events are everything commemorated on the date, official days off first.
	Events []Event
	// Holiday reports whether the date is an official day off, either because
	// it falls on the weekend or because one of its events closes the country.
	Holiday bool
	// Weekend reports whether the date falls on the weekly rest day.
	Weekend bool
}

// Title returns the name of the most significant event of the day, preferring
// an official holiday, or the empty string when nothing is commemorated.
func (d Day) Title() string {
	if len(d.Events) == 0 {
		return ""
	}

	return d.Events[0].Title
}

// A Calendar resolves dates for one country.
//
// A Calendar is safe for concurrent use. Resolved years are cached, so
// repeated lookups within a year cost a map read.
type Calendar struct {
	name      string
	weekend   []ptime.Weekday
	rules     []rule
	overrides *Overrides
	years     sync.Map // int -> map[int][]Event, keyed by month*100+day
}

// Name returns the name of the calendar, such as "iran".
func (c *Calendar) Name() string {
	return c.name
}

// Weekend returns the weekly rest days.
func (c *Calendar) Weekend() []ptime.Weekday {
	return append([]ptime.Weekday(nil), c.weekend...)
}

// ConfirmedThrough returns the last Persian year whose lunar dates are settled,
// meaning the year has passed and its announced dates were recorded.
//
// Lunar events in later years are [Estimated] even when the data file carries a
// date for them, because published calendars predict future moon sightings
// rather than report them.
func (c *Calendar) ConfirmedThrough() int {
	return c.overrides.ConfirmedThrough
}

// WithOverrides returns a copy of c that resolves dates using o instead of the
// bundled data. Passing nil restores the bundled data.
//
// Use it to pick up a correction without rebuilding:
//
//	o, err := holiday.Load(resp.Body)
//	if err == nil {
//	    cal = holiday.Iran().WithOverrides(o)
//	}
func (c *Calendar) WithOverrides(o *Overrides) *Calendar {
	if o == nil {
		o = &Overrides{}
	}

	return &Calendar{
		name:      c.name,
		weekend:   c.weekend,
		rules:     c.rules,
		overrides: o,
	}
}

// Lookup returns everything known about a date.
func (c *Calendar) Lookup(t ptime.Time) Day {
	date := ptime.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	day := Day{
		Date:    date,
		Events:  c.eventsOn(t.Year(), t.Month(), t.Day()),
		Weekend: c.isWeekend(date.Weekday()),
	}

	day.Holiday = day.Weekend

	for _, e := range day.Events {
		if e.Holiday {
			day.Holiday = true

			break
		}
	}

	return day
}

// IsHoliday reports whether a date is an official day off, counting the weekly
// rest day.
func (c *Calendar) IsHoliday(t ptime.Time) bool {
	return c.Lookup(t).Holiday
}

// IsWorkday reports whether a date is a working day.
func (c *Calendar) IsWorkday(t ptime.Time) bool {
	return !c.IsHoliday(t)
}

// Events returns every day of a Persian year that commemorates something, in
// date order. Days that are merely weekends are not included.
func (c *Calendar) Events(year int) []Day {
	byDate := c.year(year)

	keys := make([]int, 0, len(byDate))
	for key := range byDate {
		keys = append(keys, key)
	}

	sort.Ints(keys)

	days := make([]Day, 0, len(keys))

	for _, key := range keys {
		days = append(days, c.Lookup(
			ptime.Date(year, ptime.Month(key/100), key%100, 0, 0, 0, 0, ptime.Iran()),
		))
	}

	return days
}

// Holidays returns every day off in a Persian year, in date order, including
// weekends. Filter on [Day.Weekend] to drop those.
func (c *Calendar) Holidays(year int) []Day {
	var (
		days = make([]Day, 0, 80)
		day  = ptime.Date(year, ptime.Farvardin, 1, 0, 0, 0, 0, ptime.Iran())
	)

	for day.Year() == year {
		if resolved := c.Lookup(day); resolved.Holiday {
			days = append(days, resolved)
		}

		day = day.Tomorrow()
	}

	return days
}

// NextHoliday returns the first day off strictly after t.
func (c *Calendar) NextHoliday(t ptime.Time) Day {
	for day := t.Tomorrow(); ; day = day.Tomorrow() {
		if resolved := c.Lookup(day); resolved.Holiday {
			return resolved
		}
	}
}

// NextWorkday returns the first working day strictly after t.
func (c *Calendar) NextWorkday(t ptime.Time) ptime.Time {
	for day := t.Tomorrow(); ; day = day.Tomorrow() {
		if c.IsWorkday(day) {
			return day
		}
	}
}

// Workdays counts the working days in the inclusive range [from, to].
// It returns 0 when to is before from.
func (c *Calendar) Workdays(from, to ptime.Time) int {
	count := 0

	for day := from; !day.After(to); day = day.Tomorrow() {
		if c.IsWorkday(day) {
			count++
		}
	}

	return count
}

func (c *Calendar) isWeekend(wd ptime.Weekday) bool {
	return slices.Contains(c.weekend, wd)
}

func (c *Calendar) eventsOn(year int, month ptime.Month, day int) []Event {
	events := c.year(year)[int(month)*100+day]
	if len(events) == 0 {
		return nil
	}

	return append([]Event(nil), events...)
}

// year resolves and caches every event of a Persian year, keyed by
// month*100+day.
func (c *Calendar) year(year int) map[int][]Event {
	if cached, ok := c.years.Load(year); ok {
		if events, ok := cached.(map[int][]Event); ok {
			return events
		}
	}

	built := c.buildYear(year)

	if actual, _ := c.years.LoadOrStore(year, built); actual != nil {
		if events, ok := actual.(map[int][]Event); ok {
			return events
		}
	}

	return built
}

func (c *Calendar) buildYear(year int) map[int][]Event {
	var (
		events    = make(map[int][]Event, 64)
		overrides = c.overrides.forYear(year)
		confirmed = year <= c.overrides.ConfirmedThrough
	)

	add := func(month, day int, e Event) {
		events[month*100+day] = append(events[month*100+day], e)
	}

	for _, r := range c.rules {
		month, day, ok := c.resolve(r, year)
		if !ok {
			continue
		}

		confidence := Confirmed
		if r.kind == Lunar && !confirmed {
			confidence = Estimated
		}

		// A pin carries a better date than the tabular calendar can compute,
		// but it only becomes confirmed once the year has actually been
		// announced: published calendars predict future lunar dates too.
		if pin, found := overrides.pins[r.id]; found {
			if pin.remove {
				continue
			}

			month, day = pin.month, pin.day
		}

		add(month, day, Event{
			ID:         r.id,
			Title:      r.title,
			Holiday:    r.holiday,
			Kind:       r.kind,
			Confidence: confidence,
		})
	}

	for _, extra := range overrides.extras {
		add(extra.month, extra.day, Event{
			Title:      extra.title,
			Holiday:    extra.holiday,
			Kind:       Special,
			Confidence: Confirmed,
		})
	}

	for key := range events {
		sortEvents(events[key])
	}

	return events
}

// resolve returns the Persian month and day a rule falls on in a Persian year.
// A lunar rule can miss a year entirely, or fall in it twice; the first
// occurrence wins, which matches how the official calendar prints them.
func (c *Calendar) resolve(r rule, year int) (int, int, bool) {
	if r.kind == Solar {
		return r.month, r.day, true
	}

	var (
		first = jdn.FromPersian(year, 1, 1)
		last  = jdn.FromPersian(year+1, 1, 1) - 1
	)

	// A Persian year spans parts of two Hijri years, so both are candidates.
	hijriYear, _, _ := jdn.ToHijri(first)

	for _, hy := range [2]int{hijriYear, hijriYear + 1} {
		day := jdn.FromHijri(hy, r.month, r.day)
		if day < first || day > last {
			continue
		}

		_, month, dayOfMonth := jdn.ToPersian(day)

		return month, dayOfMonth, true
	}

	return 0, 0, false
}

// sortEvents puts official days off first, then longer titles, so that
// Day.Title picks the headline occasion rather than a minor observance.
func sortEvents(events []Event) {
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].Holiday != events[j].Holiday {
			return events[i].Holiday
		}

		return events[i].Kind < events[j].Kind
	})
}
