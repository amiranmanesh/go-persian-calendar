package ptime

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"time"
)

// ErrScanSource is returned by [Time.Scan] when the database driver hands over
// a value that cannot be converted into a Time.
var ErrScanSource = errors.New("ptime: unsupported source type for Scan")

// gregorianScanLayouts are tried, in order, when a driver returns a timestamp
// as text rather than as a time.Time.
var gregorianScanLayouts = []string{
	time.RFC3339Nano,
	"2006-01-02 15:04:05.999999999 -0700 MST",
	"2006-01-02 15:04:05.999999999-07:00",
	"2006-01-02 15:04:05.999999999",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

// MarshalJSON implements [json.Marshaler].
//
// The time is written as a quoted string in [RFC3339Nano] format. The zero Time
// is written as null.
func (t Time) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}

	b := make([]byte, 0, len(RFC3339Nano)+2)
	b = append(b, '"')
	b = t.AppendFormat(b, RFC3339Nano)
	b = append(b, '"')

	return b, nil
}

// UnmarshalJSON implements [json.Unmarshaler].
//
// It expects a quoted string in [RFC3339Nano] format. A null sets t to the zero
// Time.
func (t *Time) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" {
		*t = Time{}

		return nil
	}

	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return &ParseError{
			Layout: RFC3339Nano, Value: s,
			LayoutElem: `"`, ValueElem: s,
			Message: "not a JSON string",
		}
	}

	return t.UnmarshalText([]byte(s[1 : len(s)-1]))
}

// MarshalText implements [encoding.TextMarshaler].
//
// The time is written in [RFC3339Nano] format. The zero Time is written as an
// empty string.
func (t Time) MarshalText() ([]byte, error) {
	if t.IsZero() {
		return []byte{}, nil
	}

	return t.AppendFormat(make([]byte, 0, len(RFC3339Nano)), RFC3339Nano), nil
}

// UnmarshalText implements [encoding.TextUnmarshaler].
//
// It expects [RFC3339Nano] format. An empty input sets t to the zero Time.
func (t *Time) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		*t = Time{}

		return nil
	}

	parsed, err := Parse(RFC3339Nano, string(data))
	if err != nil {
		return err
	}

	*t = parsed

	return nil
}

// Value implements [driver.Valuer].
//
// The value is handed to the driver as a Gregorian [time.Time], so that it is
// stored as an ordinary SQL timestamp. The zero Time becomes NULL.
func (t Time) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil //nolint:nilnil // a NULL column is represented by a nil driver.Value
	}

	return t.Time(), nil
}

// Scan implements [sql.Scanner].
//
// It accepts NULL, a [time.Time], and a Gregorian timestamp in text form. Text
// is read as Gregorian rather than Persian because that is what [Time.Value]
// hands to the driver, and the two are indistinguishable on the wire.
func (t *Time) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*t = Time{}

		return nil
	case time.Time:
		*t = New(v)

		return nil
	case string:
		return t.scanText(v)
	case []byte:
		return t.scanText(string(v))
	default:
		return fmt.Errorf("%w: %T", ErrScanSource, src)
	}
}

func (t *Time) scanText(s string) error {
	if s == "" {
		*t = Time{}

		return nil
	}

	for _, layout := range gregorianScanLayouts {
		if gt, err := time.Parse(layout, s); err == nil {
			*t = New(gt)

			return nil
		}
	}

	return &ParseError{
		Layout: time.RFC3339Nano, Value: s,
		LayoutElem: time.RFC3339Nano, ValueElem: s,
		Message: "unrecognized Gregorian timestamp",
	}
}
