package ptime_test

import (
	"testing"
	"time"

	ptime "github.com/amiranmanesh/go-persian-calendar"
)

// fuzzLayouts are the layouts FuzzParse exercises. They are unambiguous: no two
// adjacent numeric fields run together, so a value that parses must reformat to
// something that parses back to the same instant.
var fuzzLayouts = []string{
	ptime.RFC3339,
	ptime.RFC3339Nano,
	ptime.DateOnly,
	ptime.DateTime,
	ptime.TimeOnly,
	ptime.LongDate,
	ptime.Kitchen,
	"yyyy/MM/dd HH:mm:ss.S Z",
	"d MMM yyyy",
	"E، d MMI yyyy، hh:mm a",
}

// FuzzParse checks that Parse never panics and that a successful parse is
// stable: formatting the result and parsing it again yields the same instant.
//
// Parse is deliberately more lenient than Format about digit counts, so the
// formatted text itself need not be byte identical to the input.
func FuzzParse(f *testing.F) {
	seeds := []string{
		"1394-07-02T12:59:59+03:30",
		"1394-07-02T12:59:59.05206509+03:30",
		"1403-12-30",
		"1394-07-02 12:59:59",
		"12:59:59",
		"2 مهر 1394",
		"پنج\u200cشنبه 2 مهر 1394",
		"9:54 ب.ظ",
		"",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, value string) {
		for _, layout := range fuzzLayouts {
			pt, err := ptime.Parse(layout, value)
			if err != nil {
				continue
			}

			formatted := pt.Format(layout)

			again, err := ptime.ParseInLocation(layout, formatted, pt.Location())
			if err != nil {
				t.Fatalf("Parse(%q, %q) produced %q, which no longer parses: %v", layout, value, formatted, err)
			}

			if !again.Equal(pt) {
				t.Errorf("Parse(%q, %q) is not stable: %s became %s via %q", layout, value, pt, again, formatted)
			}
		}
	})
}

// FuzzParseNeverPanics feeds arbitrary layouts and values through both parsers.
// Pathological layouts may fail to parse, but they must never panic.
func FuzzParseNeverPanics(f *testing.F) {
	f.Add(ptime.RFC3339, "1394-07-02T12:59:59+03:30")
	f.Add("2006-01-02 15:04:05", "1394-07-02 12:59:59")
	f.Add("yyyMMMdddSSnsz", "\x00\x00\x00")
	f.Add("", "")

	f.Fuzz(func(_ *testing.T, layout, value string) {
		if pt, err := ptime.Parse(layout, value); err == nil {
			_ = pt.Format(layout)
		}

		if pt, err := ptime.ParseTimeFormat(layout, value); err == nil {
			_ = pt.TimeFormat(layout)
		}
	})
}

// FuzzRoundTrip checks that every representable moment survives a
// Format/Parse round trip.
func FuzzRoundTrip(f *testing.F) {
	f.Add(1394, 7, 2, 12, 59, 59, 52065090)
	f.Add(1403, 12, 30, 0, 0, 0, 0)
	f.Add(1300, 1, 1, 23, 59, 59, 999999999)
	f.Add(1500, 6, 31, 0, 0, 0, 1)

	f.Fuzz(func(t *testing.T, year, month, day, hour, minute, sec, nsec int) {
		// Keep the inputs inside the range the calendar can represent.
		if year < 1100 || year > 1600 ||
			month < 1 || month > 12 ||
			day < 1 || day > 29 ||
			hour < 0 || hour > 23 ||
			minute < 0 || minute > 59 ||
			sec < 0 || sec > 59 ||
			nsec < 0 || nsec > 999999999 {
			t.Skip()
		}

		want := ptime.Date(year, ptime.Month(month), day, hour, minute, sec, nsec, time.UTC)

		formatted := want.Format(ptime.RFC3339Nano)

		got, err := ptime.Parse(ptime.RFC3339Nano, formatted)
		if err != nil {
			t.Fatalf("Parse(%q) returned error: %v", formatted, err)
		}

		if !got.Equal(want) || got.Nanosecond() != want.Nanosecond() {
			t.Errorf("round trip of %q produced %q", formatted, got.Format(ptime.RFC3339Nano))
		}
	})
}

// FuzzGregorianRoundTrip checks that converting to the Gregorian calendar and
// back is lossless.
func FuzzGregorianRoundTrip(f *testing.F) {
	f.Add(int64(1443000000))
	f.Add(int64(0))
	f.Add(int64(-2208988800))

	f.Fuzz(func(t *testing.T, sec int64) {
		gt := time.Unix(sec, 0).UTC()
		if gt.Year() < 1097 || gt.Year() > 9999 {
			t.Skip()
		}

		if got := ptime.New(gt).Time(); !got.Equal(gt) {
			t.Errorf("New(%s).Time() = %s", gt, got)
		}
	})
}
