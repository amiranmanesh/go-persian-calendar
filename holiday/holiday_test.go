package holiday_test

import (
	"strings"
	"testing"

	ptime "github.com/amiranmanesh/go-persian-calendar"
	"github.com/amiranmanesh/go-persian-calendar/holiday"
)

func date(year int, month ptime.Month, day int) ptime.Time {
	return ptime.Date(year, month, day, 0, 0, 0, 0, ptime.Iran())
}

func TestSolarHolidays(t *testing.T) {
	cal := holiday.Iran()

	tests := []struct {
		name  string
		date  ptime.Time
		title string
	}{
		{"nowruz", date(1404, ptime.Farvardin, 1), "نوروز"},
		{"nowruz fourth day", date(1404, ptime.Farvardin, 4), "نوروز"},
		{"islamic republic day", date(1404, ptime.Farvardin, 12), "جمهوری اسلامی"},
		{"nature day", date(1404, ptime.Farvardin, 13), "طبیعت"},
		{"khomeini death", date(1404, ptime.Khordad, 14), "رحلت"},
		{"revolution victory", date(1404, ptime.Bahman, 22), "انقلاب"},
		{"oil nationalisation", date(1404, ptime.Esfand, 29), "نفت"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			day := cal.Lookup(tc.date)

			if !day.Holiday {
				t.Fatalf("%s is not a holiday", tc.date.Format(ptime.DateOnly))
			}

			if !strings.Contains(day.Title(), tc.title) {
				t.Errorf("title = %q, want it to contain %q", day.Title(), tc.title)
			}

			for _, event := range day.Events {
				if event.Kind == holiday.Solar && event.Confidence != holiday.Confirmed {
					t.Errorf("solar event %q is %s, want confirmed", event.ID, event.Confidence)
				}
			}
		})
	}
}

// TestLunarHolidays pins dates that Iran announced, cross-checked against their
// Gregorian equivalents.
func TestLunarHolidays(t *testing.T) {
	cal := holiday.Iran()

	tests := []struct {
		name      string
		id        string
		date      ptime.Time
		gregorian string
	}{
		{"eid al-fitr 1404", "eid-fitr", date(1404, ptime.Farvardin, 11), "2025-03-31"},
		{"ashura 1404", "ashura", date(1404, ptime.Tir, 15), "2025-07-06"},
		{"tasua 1404", "tasua", date(1404, ptime.Tir, 14), "2025-07-05"},
		{"arbaeen 1404", "arbaeen", date(1404, ptime.Mordad, 23), "2025-08-14"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			day := cal.Lookup(tc.date)

			if !day.Holiday {
				t.Fatalf("%s is not a holiday", tc.date.Format(ptime.DateOnly))
			}

			var found bool

			for _, event := range day.Events {
				if event.ID == tc.id {
					found = true

					if event.Kind != holiday.Lunar {
						t.Errorf("%s is %s, want lunar", tc.id, event.Kind)
					}
				}
			}

			if !found {
				t.Errorf("%s does not carry %s, it carries %v", tc.date.Format(ptime.DateOnly), tc.id, day.Events)
			}

			if got := tc.date.Time().Format("2006-01-02"); got != tc.gregorian {
				t.Errorf("Gregorian equivalent = %s, want %s", got, tc.gregorian)
			}
		})
	}
}

func TestConfidence(t *testing.T) {
	cal := holiday.Iran()

	confirmedThrough := cal.ConfirmedThrough()
	if confirmedThrough < 1400 {
		t.Fatalf("ConfirmedThrough = %d, want at least 1400", confirmedThrough)
	}

	assertConfidence := func(year int, want holiday.Confidence) {
		t.Helper()

		var seen bool

		for _, day := range cal.Events(year) {
			for _, event := range day.Events {
				if event.Kind != holiday.Lunar {
					continue
				}

				seen = true

				if event.Confidence != want {
					t.Errorf("%s %s in %d is %s, want %s",
						day.Date.Format(ptime.DateOnly), event.ID, year, event.Confidence, want)
				}
			}
		}

		if !seen {
			t.Errorf("year %d has no lunar events", year)
		}
	}

	assertConfidence(confirmedThrough, holiday.Confirmed)
	assertConfidence(confirmedThrough+1, holiday.Estimated)
	assertConfidence(confirmedThrough+40, holiday.Estimated)
}

func TestWeekend(t *testing.T) {
	cal := holiday.Iran()

	// 1404-05-24 is a Friday.
	friday := date(1404, ptime.Mordad, 24)
	if friday.Weekday() != ptime.Jomeh {
		t.Fatalf("test date is %s, want Jomeh", friday.Weekday())
	}

	day := cal.Lookup(friday)
	if !day.Weekend || !day.Holiday {
		t.Errorf("Friday: weekend = %v, holiday = %v, want both true", day.Weekend, day.Holiday)
	}

	saturday := friday.Tomorrow()
	if got := cal.Lookup(saturday); got.Weekend {
		t.Errorf("%s is marked as weekend", saturday.Format(ptime.DateOnly))
	}
}

func TestEventsAreOrdered(t *testing.T) {
	cal := holiday.Iran()

	for _, year := range []int{1400, 1404, 1410} {
		days := cal.Events(year)
		if len(days) == 0 {
			t.Fatalf("year %d has no events", year)
		}

		for i := 1; i < len(days); i++ {
			if !days[i-1].Date.Before(days[i].Date) {
				t.Errorf("year %d: %s is not before %s", year,
					days[i-1].Date.Format(ptime.DateOnly), days[i].Date.Format(ptime.DateOnly))
			}
		}

		for _, day := range days {
			if day.Date.Year() != year {
				t.Errorf("Events(%d) returned %s", year, day.Date.Format(ptime.DateOnly))
			}
		}
	}
}

// TestEveryLunarRuleOccurs guards against a rule whose Hijri date never lands
// inside the Persian year window.
func TestEveryLunarRuleOccurs(t *testing.T) {
	cal := holiday.Iran()

	want := map[string]bool{
		"tasua": true, "ashura": true, "arbaeen": true, "prophet-death": true,
		"reza-martyrdom": true, "askari-martyrdom": true, "prophet-birth": true,
		"fatima-martyrdom": true, "ali-birth": true, "mabath": true,
		"mahdi-birth": true, "ali-martyrdom": true, "eid-fitr": true,
		"eid-fitr-2": true, "sadegh-martyrdom": true, "eid-qorban": true,
		"eid-ghadir": true,
	}

	for year := 1390; year <= 1440; year++ {
		seen := make(map[string]bool, len(want))

		for _, day := range cal.Events(year) {
			for _, event := range day.Events {
				seen[event.ID] = true
			}
		}

		for id := range want {
			if !seen[id] {
				t.Errorf("year %d is missing %s", year, id)
			}
		}
	}
}

func TestHolidaysAndWorkdays(t *testing.T) {
	cal := holiday.Iran()

	holidays := cal.Holidays(1404)
	if len(holidays) < 60 || len(holidays) > 110 {
		t.Errorf("1404 has %d days off, want somewhere between 60 and 110", len(holidays))
	}

	var offWork int

	for _, day := range holidays {
		if !day.Weekend {
			offWork++
		}
	}

	if offWork < 20 || offWork > 30 {
		t.Errorf("1404 has %d non-weekend days off, want somewhere between 20 and 30", offWork)
	}

	first := date(1404, ptime.Farvardin, 1)
	last := date(1404, ptime.Esfand, 29)

	workdays := cal.Workdays(first, last)
	if workdays+len(holidays) != 365 {
		t.Errorf("workdays (%d) plus days off (%d) = %d, want 365", workdays, len(holidays), workdays+len(holidays))
	}

	if cal.Workdays(last, first) != 0 {
		t.Error("Workdays with a reversed range is not zero")
	}
}

func TestNextHolidayAndWorkday(t *testing.T) {
	cal := holiday.Iran()

	// 1404-01-14 is the first working day after the Nowruz break.
	afterNowruz := date(1404, ptime.Farvardin, 13)

	if next := cal.NextWorkday(afterNowruz); next.Day() != 14 {
		t.Errorf("first workday after 13 Farvardin = %s, want 1404-01-14", next.Format(ptime.DateOnly))
	}

	esfand := date(1403, ptime.Esfand, 28)
	if next := cal.NextHoliday(esfand); next.Date.Year() != 1403 || next.Date.Day() != 29 {
		t.Errorf("next holiday after %s = %s, want 1403-12-29",
			esfand.Format(ptime.DateOnly), next.Date.Format(ptime.DateOnly))
	}

	// The search must cross a year boundary without looping forever.
	if next := cal.NextHoliday(date(1404, ptime.Esfand, 29)); next.Date.Year() != 1405 {
		t.Errorf("next holiday after the last day of 1404 = %s, want a date in 1405",
			next.Date.Format(ptime.DateOnly))
	}
}

func TestWithOverridesNil(t *testing.T) {
	rulesOnly := holiday.Iran().WithOverrides(nil)

	if got := rulesOnly.ConfirmedThrough(); got != 0 {
		t.Errorf("ConfirmedThrough without data = %d, want 0", got)
	}

	// Solar holidays survive without any data at all.
	if !rulesOnly.IsHoliday(date(1404, ptime.Farvardin, 1)) {
		t.Error("Nowruz is not a holiday without the data file")
	}
}

func TestLoad(t *testing.T) {
	const doc = `{
	  "schema": 1,
	  "calendar": "iran",
	  "confirmedThrough": 1404,
	  "years": {
	    "1404": [
	      {"id": "eid-fitr", "date": "1404-01-11"},
	      {"title": "تعطیلی به دلیل آلودگی هوا", "date": "1404-09-10", "holiday": true},
	      {"id": "nature-day", "remove": true}
	    ]
	  }
	}`

	overrides, err := holiday.Load(strings.NewReader(doc))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	cal := holiday.Iran().WithOverrides(overrides)

	if !cal.IsHoliday(date(1404, ptime.Azar, 10)) {
		t.Error("the one-off closure is not a holiday")
	}

	if got := cal.Lookup(date(1404, ptime.Azar, 10)).Title(); got != "تعطیلی به دلیل آلودگی هوا" {
		t.Errorf("title = %q", got)
	}

	if day := cal.Lookup(date(1404, ptime.Farvardin, 13)); day.Holiday {
		t.Error("13 Farvardin survived a remove entry")
	}

	if day := cal.Lookup(date(1404, ptime.Farvardin, 11)); !day.Holiday {
		t.Error("the pinned Eid al-Fitr is not a holiday")
	}
}

func TestLoadRejectsBadInput(t *testing.T) {
	tests := []struct {
		name string
		doc  string
	}{
		{"not json", `{`},
		{"newer schema", `{"schema": 99, "years": {}}`},
		{"year is not a number", `{"schema":1,"years":{"soon":[{"id":"x","date":"1404-01-01"}]}}`},
		{"date outside its year", `{"schema":1,"years":{"1404":[{"id":"x","date":"1405-01-01"}]}}`},
		{"malformed date", `{"schema":1,"years":{"1404":[{"id":"x","date":"1404/01/01"}]}}`},
		{"month out of range", `{"schema":1,"years":{"1404":[{"id":"x","date":"1404-13-01"}]}}`},
		{"neither id nor title", `{"schema":1,"years":{"1404":[{"date":"1404-01-01"}]}}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := holiday.Load(strings.NewReader(tc.doc)); err == nil {
				t.Error("Load accepted invalid data")
			}
		})
	}
}

func TestLookupIsIndependentOfClockAndZone(t *testing.T) {
	cal := holiday.Iran()

	morning := ptime.Date(1404, ptime.Farvardin, 1, 0, 0, 0, 0, ptime.Iran())
	evening := ptime.Date(1404, ptime.Farvardin, 1, 23, 59, 59, 0, ptime.Afghanistan())

	if cal.Lookup(morning).Title() != cal.Lookup(evening).Title() {
		t.Error("the same calendar date resolved differently across clock and zone")
	}
}

func TestDayTitleOfAnOrdinaryDay(t *testing.T) {
	day := holiday.Iran().Lookup(date(1404, ptime.Mordad, 25))

	if day.Title() != "" {
		t.Errorf("Title of a day with no events = %q, want empty", day.Title())
	}
}
