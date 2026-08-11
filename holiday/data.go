package holiday

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

//go:embed data/iran.json
var iranData []byte

// Errors reported when an override file cannot be used.
var (
	// ErrUnsupportedSchema means the file was written by a newer version of
	// this package than the one reading it.
	ErrUnsupportedSchema = errors.New("holiday: unsupported data schema")
	// ErrMalformedData means the file parsed as JSON but does not describe a
	// usable calendar.
	ErrMalformedData = errors.New("holiday: malformed data")
)

// schemaVersion is the highest data schema this package understands.
const schemaVersion = 1

// Overrides is the part of a calendar that rules cannot produce: the officially
// announced dates of lunar events, and one-off days such as an air pollution
// closure or an election.
//
// Load parses it from the JSON published alongside the package. The zero value
// is valid and contributes nothing, which makes every lunar event [Estimated].
type Overrides struct {
	// Schema is the version of the file format.
	Schema int `json:"schema"`
	// Calendar names the country the data belongs to, such as "iran".
	Calendar string `json:"calendar"`
	// UpdatedAt is the Gregorian date the file was last regenerated.
	UpdatedAt string `json:"updatedAt"`
	// ConfirmedThrough is the last Persian year reconciled with the official
	// calendar. Lunar events after it are estimates.
	ConfirmedThrough int `json:"confirmedThrough"`
	// Sources records where the data was checked against, for auditing.
	Sources []string `json:"sources,omitempty"`
	// Years maps a Persian year to its entries.
	Years map[string][]Entry `json:"years"`

	// indexed is built once by Load. A zero Overrides leaves it nil and
	// contributes nothing, which is what WithOverrides(nil) relies on.
	indexed map[int]yearOverrides
}

// An Entry adjusts one date of one year.
//
// It does one of three things, depending on which fields are set:
//
//   - ID and Date: the rule with that ID falls on Date in this year, which
//     also marks the occurrence confirmed.
//   - ID and Remove: the rule does not occur in this year.
//   - Title and Date: a one-off event on Date.
type Entry struct {
	// ID is the rule this entry adjusts. Empty for a one-off event.
	ID string `json:"id,omitempty"`
	// Date is the Persian date, as "yyyy-MM-dd".
	Date string `json:"date,omitempty"`
	// Title names a one-off event.
	Title string `json:"title,omitempty"`
	// Holiday reports whether a one-off event closes the country.
	Holiday bool `json:"holiday,omitempty"`
	// Remove drops the occurrence of a rule in this year.
	Remove bool `json:"remove,omitempty"`
	// Note records why the entry exists, for whoever reviews the next change.
	Note string `json:"note,omitempty"`
}

// Load reads an override file, as published at
// https://amiranmanesh.github.io/go-persian-calendar/data/v1/iran.json
//
// Pass the result to [Calendar.WithOverrides].
func Load(r io.Reader) (*Overrides, error) {
	var o Overrides

	if err := json.NewDecoder(r).Decode(&o); err != nil {
		return nil, fmt.Errorf("holiday: decoding data: %w", err)
	}

	if o.Schema > schemaVersion {
		return nil, fmt.Errorf("%w: file is version %d, this package understands %d",
			ErrUnsupportedSchema, o.Schema, schemaVersion)
	}

	indexed, err := buildIndex(o.Years)
	if err != nil {
		return nil, err
	}

	o.indexed = indexed

	return &o, nil
}

// bundledIran returns the data embedded at build time.
func bundledIran() *Overrides {
	o, err := Load(strings.NewReader(string(iranData)))
	if err != nil {
		// The bundled file is generated and verified in CI, so a failure here
		// is a build that should never have shipped.
		panic("holiday: bundled data is invalid: " + err.Error())
	}

	return o
}

type pin struct {
	month  int
	day    int
	remove bool
}

type extra struct {
	month   int
	day     int
	title   string
	holiday bool
}

type yearOverrides struct {
	pins   map[string]pin
	extras []extra
}

func (o *Overrides) forYear(year int) yearOverrides {
	return o.indexed[year]
}

func buildIndex(years map[string][]Entry) (map[int]yearOverrides, error) {
	indexed := make(map[int]yearOverrides, len(years))

	for key, entries := range years {
		year, err := strconv.Atoi(key)
		if err != nil {
			return nil, fmt.Errorf("%w: %q is not a year: %w", ErrMalformedData, key, err)
		}

		current := yearOverrides{pins: make(map[string]pin, len(entries))}

		for _, entry := range entries {
			if err := addEntry(&current, year, entry); err != nil {
				return nil, err
			}
		}

		indexed[year] = current
	}

	return indexed, nil
}

func addEntry(into *yearOverrides, year int, entry Entry) error {
	if entry.ID != "" && entry.Remove {
		into.pins[entry.ID] = pin{remove: true}

		return nil
	}

	month, day, err := parseDate(entry.Date, year)
	if err != nil {
		return err
	}

	if entry.ID != "" {
		into.pins[entry.ID] = pin{month: month, day: day}

		return nil
	}

	if entry.Title == "" {
		return fmt.Errorf("%w: the entry for %d has neither an id nor a title", ErrMalformedData, year)
	}

	into.extras = append(into.extras, extra{
		month: month, day: day, title: entry.Title, holiday: entry.Holiday,
	})

	return nil
}

// parseDate reads a "yyyy-MM-dd" Persian date and checks that it belongs to
// the year that holds it.
func parseDate(date string, year int) (int, int, error) {
	parts := strings.Split(date, "-")
	if len(parts) != 3 {
		return 0, 0, fmt.Errorf("%w: %q is not a yyyy-MM-dd date", ErrMalformedData, date)
	}

	numbers := make([]int, 3)

	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil {
			return 0, 0, fmt.Errorf("%w: %q is not a yyyy-MM-dd date: %w", ErrMalformedData, date, err)
		}

		numbers[i] = n
	}

	entryYear, month, day := numbers[0], numbers[1], numbers[2]

	switch {
	case entryYear != year:
		return 0, 0, fmt.Errorf("%w: date %q is listed under year %d", ErrMalformedData, date, year)
	case month < 1 || month > 12:
		return 0, 0, fmt.Errorf("%w: date %q has no month %d", ErrMalformedData, date, month)
	case day < 1 || day > 31:
		return 0, 0, fmt.Errorf("%w: date %q has no day %d", ErrMalformedData, date, day)
	}

	return month, day, nil
}
