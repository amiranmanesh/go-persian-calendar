//nolint:testpackage // these tests exercise unexported conversion helpers
package jdn

import "testing"

// TestHijriAnchors pins the tabular calendar to dates that can be checked
// against published references. The last two are dates Iran announced by moon
// sighting, so they also show that the table matches observation for the
// current era.
func TestHijriAnchors(t *testing.T) {
	tests := []struct {
		name                string
		hYear, hMonth, hDay int
		gYear, gMonth, gDay int
	}{
		{"epoch", 1, 1, 1, 622, 7, 16},
		{"1 Muharram 1443", 1443, 1, 1, 2021, 8, 10},
		{"Eid al-Fitr 1446", 1446, 10, 1, 2025, 3, 31},
		{"Ashura 1447", 1447, 1, 10, 2025, 7, 6},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			want := FromGregorian(tc.gYear, tc.gMonth, tc.gDay)

			if got := FromHijri(tc.hYear, tc.hMonth, tc.hDay); got != want {
				gy, gm, gd := ToGregorian(got)
				t.Errorf("FromHijri(%d, %d, %d) = %d (%d-%02d-%02d), want %d (%d-%02d-%02d)",
					tc.hYear, tc.hMonth, tc.hDay, got, gy, gm, gd,
					want, tc.gYear, tc.gMonth, tc.gDay)
			}

			y, m, d := ToHijri(want)
			if y != tc.hYear || m != tc.hMonth || d != tc.hDay {
				t.Errorf("ToHijri(%d) = %d-%02d-%02d, want %d-%02d-%02d",
					want, y, m, d, tc.hYear, tc.hMonth, tc.hDay)
			}
		})
	}
}

// TestHijriRoundTrip walks about 1600 Hijri years a week at a time.
func TestHijriRoundTrip(t *testing.T) {
	for j := hijriEpoch; j < hijriEpoch+600000; j += 7 {
		year, month, day := ToHijri(j)

		if got := FromHijri(year, month, day); got != j {
			t.Fatalf("JDN %d became %d-%02d-%02d and converted back to %d", j, year, month, day, got)
		}

		if month < 1 || month > 12 {
			t.Fatalf("JDN %d produced month %d", j, month)
		}

		if day < 1 || day > HijriMonthDays(year, month) {
			t.Fatalf("JDN %d produced day %d of %d-%02d", j, day, year, month)
		}
	}
}

// TestHijriYearLength checks the 30 year cycle: 11 leap years of 355 days.
func TestHijriYearLength(t *testing.T) {
	leaps := 0

	for year := 1441; year < 1471; year++ {
		length := FromHijri(year+1, 1, 1) - FromHijri(year, 1, 1)

		want := 354
		if IsHijriLeap(year) {
			want, leaps = 355, leaps+1
		}

		if length != want {
			t.Errorf("Hijri year %d is %d days, want %d", year, length, want)
		}
	}

	if leaps != 11 {
		t.Errorf("got %d leap years in a 30 year cycle, want 11", leaps)
	}
}
