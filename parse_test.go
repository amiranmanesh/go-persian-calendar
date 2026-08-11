package ptime_test

import (
	"errors"
	"testing"
	"time"

	ptime "github.com/amiranmanesh/go-persian-calendar"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name   string
		layout string
		value  string
		want   ptime.Time
	}{
		{
			name:   "date only",
			layout: ptime.DateOnly,
			value:  "1394-07-02",
			want:   ptime.Date(1394, ptime.Mehr, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:   "date and time",
			layout: ptime.DateTime,
			value:  "1394-07-02 12:59:59",
			want:   ptime.Date(1394, ptime.Mehr, 2, 12, 59, 59, 0, time.UTC),
		},
		{
			name:   "rfc3339 with offset",
			layout: ptime.RFC3339,
			value:  "1394-07-02T12:59:59+03:30",
			want:   ptime.Date(1394, ptime.Mehr, 2, 12, 59, 59, 0, ptime.Iran()),
		},
		{
			name:   "rfc3339 nano",
			layout: ptime.RFC3339Nano,
			value:  "1394-07-02T12:59:59.05206509+03:30",
			want:   ptime.Date(1394, ptime.Mehr, 2, 12, 59, 59, 52065090, ptime.Iran()),
		},
		{
			name:   "rfc3339 nano without fraction",
			layout: ptime.RFC3339Nano,
			value:  "1394-07-02T12:59:59+03:30",
			want:   ptime.Date(1394, ptime.Mehr, 2, 12, 59, 59, 0, ptime.Iran()),
		},
		{
			name:   "persian month name",
			layout: "d MMM yyyy",
			value:  "2 مهر 1394",
			want:   ptime.Date(1394, ptime.Mehr, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:   "dari month name",
			layout: "d MMI yyyy",
			value:  "2 میزان 1394",
			want:   ptime.Date(1394, ptime.Mizan, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:   "weekday is matched and ignored",
			layout: ptime.LongDate,
			value:  "پنج\u200cشنبه 2 مهر 1394",
			want:   ptime.Date(1394, ptime.Mehr, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:   "two digit year below the pivot",
			layout: "yy/MM/dd",
			value:  "04/05/20",
			want:   ptime.Date(1404, ptime.Mordad, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			name:   "two digit year above the pivot",
			layout: "yy/MM/dd",
			value:  "94/07/02",
			want:   ptime.Date(1394, ptime.Mehr, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:   "12 hour clock, afternoon",
			layout: "yyyy-MM-dd hh:mm a",
			value:  "1394-07-02 02:07 ب.ظ",
			want:   ptime.Date(1394, ptime.Mehr, 2, 14, 7, 0, 0, time.UTC),
		},
		{
			name:   "12 hour clock, midnight",
			layout: "yyyy-MM-dd hh:mm a",
			value:  "1394-07-02 12:00 ق.ظ",
			want:   ptime.Date(1394, ptime.Mehr, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:   "12 hour clock, noon",
			layout: "yyyy-MM-dd hh:mm A",
			value:  "1394-07-02 12:00 بعد از ظهر",
			want:   ptime.Date(1394, ptime.Mehr, 2, 12, 0, 0, 0, time.UTC),
		},
		{
			name:   "zero based 12 hour clock",
			layout: "yyyy-MM-dd KK:mm a",
			value:  "1394-07-02 02:07 ب.ظ",
			want:   ptime.Date(1394, ptime.Mehr, 2, 14, 7, 0, 0, time.UTC),
		},
		{
			name:   "24 hour clock where 24 means midnight",
			layout: "yyyy-MM-dd kk:mm",
			value:  "1394-07-02 24:00",
			want:   ptime.Date(1394, ptime.Mehr, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:   "milliseconds",
			layout: "yyyy-MM-dd HH:mm:ss.S",
			value:  "1394-07-02 12:59:59.052",
			want:   ptime.Date(1394, ptime.Mehr, 2, 12, 59, 59, 52000000, time.UTC),
		},
		{
			name:   "zone name",
			layout: "yyyy-MM-dd HH:mm:ss z",
			value:  "1394-07-02 12:59:59 Asia/Tehran",
			want:   ptime.Date(1394, ptime.Mehr, 2, 12, 59, 59, 0, ptime.Iran()),
		},
		{
			name:   "utc designator",
			layout: ptime.RFC3339,
			value:  "1394-07-02T12:59:59Z",
			want:   ptime.Date(1394, ptime.Mehr, 2, 12, 59, 59, 0, time.UTC),
		},
		{
			name:   "leap day",
			layout: ptime.DateOnly,
			value:  "1403-12-30",
			want:   ptime.Date(1403, ptime.Esfand, 30, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ptime.Parse(tc.layout, tc.value)
			if err != nil {
				t.Fatalf("Parse(%q, %q) returned error: %v", tc.layout, tc.value, err)
			}

			if !got.Equal(tc.want) {
				t.Errorf("Parse(%q, %q) = %s, want %s", tc.layout, tc.value, got, tc.want)
			}

			if y, m, d := got.Date(); y != tc.want.Year() || m != tc.want.Month() || d != tc.want.Day() {
				t.Errorf("Parse(%q, %q) date = %d/%d/%d, want %d/%d/%d",
					tc.layout, tc.value, y, m, d, tc.want.Year(), tc.want.Month(), tc.want.Day())
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name   string
		layout string
		value  string
	}{
		{"month out of range", ptime.DateOnly, "1394-13-02"},
		{"day out of range", ptime.DateOnly, "1394-07-31"},
		{"day out of range in a common year", ptime.DateOnly, "1394-12-30"},
		{"hour out of range", ptime.DateTime, "1394-07-02 24:00:00"},
		{"minute out of range", ptime.DateTime, "1394-07-02 12:60:00"},
		{"not a number", ptime.DateOnly, "1394-ab-02"},
		{"missing separator", ptime.DateOnly, "13940702"},
		{"trailing text", ptime.DateOnly, "1394-07-02 extra"},
		{"unknown month name", "d MMM yyyy", "2 خرداذ 1394"},
		{"unknown zone", "yyyy-MM-dd z", "1394-07-02 Mars/Olympus"},
		{"empty value", ptime.DateOnly, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ptime.Parse(tc.layout, tc.value)
			if err == nil {
				t.Fatalf("Parse(%q, %q) = %s, want an error", tc.layout, tc.value, got)
			}

			var parseErr *ptime.ParseError
			if !errors.As(err, &parseErr) {
				t.Fatalf("Parse(%q, %q) error is %T, want *ptime.ParseError", tc.layout, tc.value, err)
			}

			if parseErr.Error() == "" {
				t.Error("ParseError.Error() is empty")
			}
		})
	}
}

func TestParseInLocation(t *testing.T) {
	got, err := ptime.ParseInLocation(ptime.DateTime, "1394-07-02 12:59:59", ptime.Iran())
	if err != nil {
		t.Fatalf("ParseInLocation returned error: %v", err)
	}

	if want := ptime.Date(1394, ptime.Mehr, 2, 12, 59, 59, 0, ptime.Iran()); !got.Equal(want) {
		t.Errorf("ParseInLocation = %s, want %s", got, want)
	}

	// An explicit zone in the value wins over the default location.
	got, err = ptime.ParseInLocation(ptime.RFC3339, "1394-07-02T12:59:59Z", ptime.Iran())
	if err != nil {
		t.Fatalf("ParseInLocation returned error: %v", err)
	}

	if _, offset := got.Zone(); offset != 0 {
		t.Errorf("ParseInLocation zone offset = %d, want 0", offset)
	}
}

func TestParseTimeFormat(t *testing.T) {
	tests := []struct {
		name   string
		layout string
		value  string
		want   ptime.Time
	}{
		{
			name:   "numeric layout",
			layout: "2006/01/02 15:04:05",
			value:  "1394/07/02 12:59:59",
			want:   ptime.Date(1394, ptime.Mehr, 2, 12, 59, 59, 0, time.UTC),
		},
		{
			name:   "month name",
			layout: "2 Jan 2006",
			value:  "2 مهر 1394",
			want:   ptime.Date(1394, ptime.Mehr, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:   "dari month name",
			layout: "2 January 2006",
			value:  "2 میزان 1394",
			want:   ptime.Date(1394, ptime.Mizan, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:   "space padded day",
			layout: "2006-01-_2",
			value:  "1394-07- 2",
			want:   ptime.Date(1394, ptime.Mehr, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:   "12 hour clock with marker",
			layout: "2006-01-02 03:04 pm",
			value:  "1394-07-02 02:07 ب.ظ",
			want:   ptime.Date(1394, ptime.Mehr, 2, 14, 7, 0, 0, time.UTC),
		},
		{
			name:   "fraction and offset",
			layout: "2006-01-02T15:04:05.999999999-07:00",
			value:  "1394-07-02T12:59:59.05206509+03:30",
			want:   ptime.Date(1394, ptime.Mehr, 2, 12, 59, 59, 52065090, ptime.Iran()),
		},
		{
			name:   "zone name",
			layout: "2006-01-02 15:04:05 MST",
			value:  "1394-07-02 12:59:59 Asia/Tehran",
			want:   ptime.Date(1394, ptime.Mehr, 2, 12, 59, 59, 0, ptime.Iran()),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ptime.ParseTimeFormat(tc.layout, tc.value)
			if err != nil {
				t.Fatalf("ParseTimeFormat(%q, %q) returned error: %v", tc.layout, tc.value, err)
			}

			if !got.Equal(tc.want) {
				t.Errorf("ParseTimeFormat(%q, %q) = %s, want %s", tc.layout, tc.value, got, tc.want)
			}
		})
	}
}

func TestParseTimeFormatInLocation(t *testing.T) {
	got, err := ptime.ParseTimeFormatInLocation("2006-01-02", "1394-07-02", ptime.Afghanistan())
	if err != nil {
		t.Fatalf("ParseTimeFormatInLocation returned error: %v", err)
	}

	if got.Location().String() != ptime.Afghanistan().String() {
		t.Errorf("location = %s, want %s", got.Location(), ptime.Afghanistan())
	}
}

// TestFormatParseRoundTrip checks that everything Format writes, Parse reads back.
func TestFormatParseRoundTrip(t *testing.T) {
	layouts := []string{
		ptime.RFC3339,
		ptime.RFC3339Nano,
		ptime.DateTime,
		ptime.DateOnly,
		ptime.LongDate,
		"yyyy/MM/dd HH:mm:ss.S Z",
		"E، d MMM yyyy ساعت hh:mm:ss a",
		"yy-M-d k:m:s",
	}

	moments := []ptime.Time{
		ptime.Date(1394, ptime.Mehr, 2, 12, 59, 59, 52065090, ptime.Iran()),
		ptime.Date(1403, ptime.Esfand, 30, 0, 0, 0, 0, ptime.Iran()),
		ptime.Date(1400, ptime.Farvardin, 1, 23, 59, 59, 0, time.UTC),
		ptime.Date(1399, ptime.Dey, 15, 8, 5, 3, 1000000, ptime.Afghanistan()),
	}

	for _, layout := range layouts {
		for _, want := range moments {
			t.Run(layout+"/"+want.String(), func(t *testing.T) {
				formatted := want.Format(layout)

				got, err := ptime.ParseInLocation(layout, formatted, want.Location())
				if err != nil {
					t.Fatalf("Parse(%q, %q) returned error: %v", layout, formatted, err)
				}

				if reformatted := got.Format(layout); reformatted != formatted {
					t.Errorf("round trip of %q: got %q, want %q", layout, reformatted, formatted)
				}
			})
		}
	}
}
