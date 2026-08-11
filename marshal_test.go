package ptime_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	ptime "github.com/amiranmanesh/go-persian-calendar"
)

func TestMarshalJSON(t *testing.T) {
	type payload struct {
		At ptime.Time `json:"at"`
	}

	tests := []struct {
		name string
		in   ptime.Time
		want string
	}{
		{
			name: "with a fractional second",
			in:   ptime.Date(1394, ptime.Mehr, 2, 12, 59, 59, 52065090, ptime.Iran()),
			want: `{"at":"1394-07-02T12:59:59.05206509+03:30"}`,
		},
		{
			name: "without a fractional second",
			in:   ptime.Date(1394, ptime.Mehr, 2, 12, 59, 59, 0, ptime.Iran()),
			want: `{"at":"1394-07-02T12:59:59+03:30"}`,
		},
		{
			name: "in utc",
			in:   ptime.Date(1400, ptime.Farvardin, 1, 0, 0, 0, 0, time.UTC),
			want: `{"at":"1400-01-01T00:00:00+00:00"}`,
		},
		{
			name: "zero time",
			in:   ptime.Time{},
			want: `{"at":null}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := json.Marshal(payload{At: tc.in})
			if err != nil {
				t.Fatalf("json.Marshal returned error: %v", err)
			}

			if string(encoded) != tc.want {
				t.Errorf("json.Marshal = %s, want %s", encoded, tc.want)
			}

			var decoded payload
			if err := json.Unmarshal(encoded, &decoded); err != nil {
				t.Fatalf("json.Unmarshal returned error: %v", err)
			}

			if tc.in.IsZero() {
				if !decoded.At.IsZero() {
					t.Errorf("json.Unmarshal = %s, want the zero Time", decoded.At)
				}

				return
			}

			if !decoded.At.Equal(tc.in) {
				t.Errorf("json.Unmarshal = %s, want %s", decoded.At, tc.in)
			}
		})
	}
}

func TestUnmarshalJSONErrors(t *testing.T) {
	for _, input := range []string{`123`, `"not a date"`, `"1394-13-02T00:00:00Z"`} {
		var pt ptime.Time
		if err := json.Unmarshal([]byte(input), &pt); err == nil {
			t.Errorf("json.Unmarshal(%s) = %s, want an error", input, pt)
		}
	}
}

func TestMarshalText(t *testing.T) {
	in := ptime.Date(1394, ptime.Mehr, 2, 12, 59, 59, 0, ptime.Iran())

	encoded, err := in.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText returned error: %v", err)
	}

	if want := "1394-07-02T12:59:59+03:30"; string(encoded) != want {
		t.Fatalf("MarshalText = %s, want %s", encoded, want)
	}

	var got ptime.Time
	if err := got.UnmarshalText(encoded); err != nil {
		t.Fatalf("UnmarshalText returned error: %v", err)
	}

	if !got.Equal(in) {
		t.Errorf("UnmarshalText = %s, want %s", got, in)
	}

	if err := got.UnmarshalText(nil); err != nil {
		t.Fatalf("UnmarshalText(nil) returned error: %v", err)
	}

	if !got.IsZero() {
		t.Errorf("UnmarshalText(nil) = %s, want the zero Time", got)
	}
}

func TestValue(t *testing.T) {
	in := ptime.Date(1394, ptime.Mehr, 2, 12, 59, 59, 0, ptime.Iran())

	value, err := in.Value()
	if err != nil {
		t.Fatalf("Value returned error: %v", err)
	}

	gt, ok := value.(time.Time)
	if !ok {
		t.Fatalf("Value returned %T, want time.Time", value)
	}

	if want := in.Time(); !gt.Equal(want) {
		t.Errorf("Value = %s, want %s", gt, want)
	}

	value, err = ptime.Time{}.Value()
	if err != nil {
		t.Fatalf("Value returned error: %v", err)
	}

	if value != nil {
		t.Errorf("Value of the zero Time = %v, want nil", value)
	}
}

func TestScan(t *testing.T) {
	want := ptime.Date(1394, ptime.Mehr, 2, 12, 59, 59, 0, ptime.Iran())

	// Text is read as Gregorian, which is what Value hands to the driver.
	sources := []any{
		want.Time(),
		want.Time().Format(time.RFC3339Nano),
		[]byte(want.Time().Format(time.RFC3339Nano)),
		want.Time().Format("2006-01-02 15:04:05-07:00"),
		want.Time().UTC().String(),
	}

	for _, src := range sources {
		var got ptime.Time
		if err := got.Scan(src); err != nil {
			t.Fatalf("Scan(%v) returned error: %v", src, err)
		}

		if !got.Equal(want) {
			t.Errorf("Scan(%v) = %s, want %s", src, got, want)
		}
	}

	var got ptime.Time
	if err := got.Scan(nil); err != nil {
		t.Fatalf("Scan(nil) returned error: %v", err)
	}

	if !got.IsZero() {
		t.Errorf("Scan(nil) = %s, want the zero Time", got)
	}

	if err := got.Scan(42); !errors.Is(err, ptime.ErrScanSource) {
		t.Errorf("Scan(42) error = %v, want ErrScanSource", err)
	}

	if err := got.Scan("definitely not a timestamp"); err == nil {
		t.Error("Scan of an unparsable string returned no error")
	}
}
