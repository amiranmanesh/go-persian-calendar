package holiday_test

import (
	"testing"

	ptime "github.com/amiranmanesh/go-persian-calendar"
	"github.com/amiranmanesh/go-persian-calendar/holiday"
)

func BenchmarkLookup(b *testing.B) {
	cal := holiday.Iran()
	day := ptime.Date(1404, ptime.Mordad, 20, 0, 0, 0, 0, ptime.Iran())

	cal.Lookup(day) // warm the year cache

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sinkDay = cal.Lookup(day)
	}
}

func BenchmarkIsHoliday(b *testing.B) {
	cal := holiday.Iran()
	day := ptime.Date(1404, ptime.Farvardin, 1, 0, 0, 0, 0, ptime.Iran())

	cal.IsHoliday(day)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sinkBool = cal.IsHoliday(day)
	}
}

// BenchmarkColdYear measures resolving a year that is not cached yet.
func BenchmarkColdYear(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sinkDays = holiday.Iran().WithOverrides(nil).Events(1404)
	}
}

var (
	sinkDay  holiday.Day
	sinkDays []holiday.Day
	sinkBool bool
)
