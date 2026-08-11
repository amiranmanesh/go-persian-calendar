package holiday_test

import (
	"fmt"

	ptime "github.com/amiranmanesh/go-persian-calendar"
	"github.com/amiranmanesh/go-persian-calendar/holiday"
)

func ExampleCalendar_IsHoliday() {
	cal := holiday.Iran()

	nowruz := ptime.Date(1404, ptime.Farvardin, 1, 0, 0, 0, 0, ptime.Iran())
	ordinary := ptime.Date(1404, ptime.Mordad, 25, 0, 0, 0, 0, ptime.Iran())

	fmt.Println(cal.IsHoliday(nowruz), cal.IsHoliday(ordinary))
	// Output: true false
}

func ExampleCalendar_Lookup() {
	cal := holiday.Iran()

	day := cal.Lookup(ptime.Date(1404, ptime.Farvardin, 13, 0, 0, 0, 0, ptime.Iran()))

	fmt.Println(day.Holiday, day.Title())
	// Output: true روز طبیعت، سیزده به‌در
}

// A lunar holiday in a year that has not been settled yet is reported as an
// estimate, because Iran fixes those dates by moon sighting.
func ExampleEvent_confidence() {
	cal := holiday.Iran()

	settled := cal.ConfirmedThrough()

	for _, year := range []int{settled, settled + 1} {
		for _, day := range cal.Events(year) {
			for _, event := range day.Events {
				if event.ID == "eid-fitr" {
					fmt.Printf("%d: %s is %s\n", year, event.ID, event.Confidence)
				}
			}
		}
	}
	// Output:
	// 1404: eid-fitr is confirmed
	// 1405: eid-fitr is estimated
}

func ExampleCalendar_NextWorkday() {
	cal := holiday.Iran()

	// The Nowruz break covers 1 to 4 Farvardin; 12 and 13 are separate
	// holidays, so work resumes on the 5th.
	from := ptime.Date(1404, ptime.Farvardin, 1, 0, 0, 0, 0, ptime.Iran())

	fmt.Println(cal.NextWorkday(from).Format(ptime.DateOnly))
	// Output: 1404-01-05
}

func ExampleCalendar_Workdays() {
	cal := holiday.Iran()

	from := ptime.Date(1404, ptime.Farvardin, 1, 0, 0, 0, 0, ptime.Iran())
	to := ptime.Date(1404, ptime.Farvardin, 31, 0, 0, 0, 0, ptime.Iran())

	fmt.Printf("Farvardin 1404 has %d working days\n", cal.Workdays(from, to))
	// Output: Farvardin 1404 has 20 working days
}

func ExampleCalendar_Holidays() {
	cal := holiday.Iran()

	var official int

	for _, day := range cal.Holidays(1404) {
		if !day.Weekend {
			official++
		}
	}

	fmt.Printf("1404 has %d days off that are not Fridays\n", official)
	// Output: 1404 has 21 days off that are not Fridays
}
