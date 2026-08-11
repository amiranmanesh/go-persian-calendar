// Command gen reconciles the computed holiday calendar with published Iranian
// calendars and rewrites holiday/data/iran.json.
//
// The rules in the holiday package already produce every recurring event. What
// they cannot produce is the officially announced date of a lunar event, which
// depends on moon sighting, and one-off closures. This command finds exactly
// those, by comparing what the rules compute against a reference calendar, and
// writes only the differences.
//
// Only facts are taken from the references: which date an event fell on. The
// resulting file is our own compilation and carries the module's license.
//
// Usage:
//
//	go run ./holiday/internal/gen -from 1390 -to 1419 -out holiday/data/iran.json
//
// The command fails rather than writing a smaller file when a reference cannot
// be reached, so a broken source can never quietly erase data.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	ptime "github.com/amiranmanesh/go-persian-calendar"
	"github.com/amiranmanesh/go-persian-calendar/holiday"
)

const referenceURL = "https://raw.githubusercontent.com/hasan-ahani/shamsi-holidays/master/holidays/%d.json"

// keywords match a reference title to a rule ID. The reference writes titles
// in prose, so matching is by distinctive substrings rather than equality.
// Every substring in a group must be present.
var keywords = map[string][][]string{
	"tasua":            {{"تاسوعا"}},
	"ashura":           {{"عاشورا"}},
	"arbaeen":          {{"اربعین"}},
	"prophet-death":    {{"رحلت", "رسول"}},
	"reza-martyrdom":   {{"شهادت", "امام رضا"}},
	"askari-martyrdom": {{"عسکری"}},
	"prophet-birth":    {{"میلاد", "رسول"}},
	"fatima-martyrdom": {{"فاطمه"}},
	"ali-birth":        {{"ولادت", "امام علی"}},
	"mabath":           {{"مبعث"}},
	"mahdi-birth":      {{"نیمه شعبان"}, {"ولادت", "قائم"}},
	"ali-martyrdom":    {{"شهادت", "حضرت علی"}},
	"eid-fitr":         {{"فطر"}},
	"sadegh-martyrdom": {{"شهادت", "جعفر صادق"}},
	"eid-qorban":       {{"قربان"}},
	"eid-ghadir":       {{"غدیر"}},
}

// lunarRuleIDs is the set of rules the reconciler pins, in a stable order.
var lunarRuleIDs = []string{
	"tasua", "ashura", "arbaeen", "prophet-death", "reza-martyrdom",
	"askari-martyrdom", "prophet-birth", "fatima-martyrdom", "ali-birth",
	"mabath", "mahdi-birth", "ali-martyrdom", "eid-fitr", "eid-fitr-2",
	"sadegh-martyrdom", "eid-qorban", "eid-ghadir",
}

func main() {
	var (
		from    = flag.Int("from", 1390, "first Persian year to reconcile")
		to      = flag.Int("to", 1419, "last Persian year to reconcile")
		out     = flag.String("out", "holiday/data/iran.json", "file to write")
		dry     = flag.Bool("dry-run", false, "report differences without writing")
		timeout = flag.Duration("timeout", 2*time.Minute, "total time allowed for fetching")
	)

	flag.Parse()

	if err := run(*from, *to, *out, *dry, *timeout); err != nil {
		fmt.Fprintln(os.Stderr, "gen:", err)
		os.Exit(1)
	}
}

func run(from, to int, out string, dry bool, timeout time.Duration) error {
	if from > to {
		return fmt.Errorf("empty year range %d..%d", from, to)
	}

	client := &http.Client{Timeout: timeout}
	calendar := holiday.Iran().WithOverrides(nil) // rules only, so nothing is pinned twice

	years := make(map[string][]holiday.Entry, to-from+1)
	confirmed := 0

	for year := from; year <= to; year++ {
		reference, err := fetchYear(client, year)
		if err != nil {
			// A year the reference does not publish yet is not an error; a
			// reference that cannot be reached at all is.
			if errors.Is(err, errNotPublished) {
				fmt.Fprintf(os.Stderr, "gen: %d is not published by the reference, skipping\n", year)

				continue
			}

			return fmt.Errorf("year %d: %w", year, err)
		}

		entries, err := reconcile(calendar, year, reference)
		if err != nil {
			return fmt.Errorf("year %d: %w", year, err)
		}

		if len(entries) > 0 {
			years[strconv.Itoa(year)] = entries
		}

		confirmed = year

		fmt.Fprintf(os.Stderr, "gen: %d reconciled, %d entries\n", year, len(entries))
	}

	if confirmed == 0 {
		return errors.New("no year could be reconciled, refusing to write")
	}

	// References publish predicted lunar dates for future years alongside
	// announced ones. Only a year that has already ended can be called settled,
	// so the confirmed mark never runs past the previous Persian year.
	if lastComplete := ptime.Now().Year() - 1; confirmed > lastComplete {
		confirmed = lastComplete
	}

	data := holiday.Overrides{
		Schema:           1,
		Calendar:         "iran",
		UpdatedAt:        time.Now().UTC().Format(time.DateOnly),
		ConfirmedThrough: confirmed,
		Sources: []string{
			"https://github.com/hasan-ahani/shamsi-holidays (dates only, sourced from time.ir)",
		},
		Years: years,
	}

	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding: %w", err)
	}

	encoded = append(encoded, '\n')

	if dry {
		fmt.Fprintf(os.Stderr, "gen: dry run, %d bytes not written\n", len(encoded))

		return nil
	}

	if err := os.WriteFile(out, encoded, 0o644); err != nil { //nolint:gosec // data file, world readable on purpose
		return fmt.Errorf("writing %s: %w", out, err)
	}

	fmt.Fprintf(os.Stderr, "gen: wrote %s, confirmed through %d\n", out, confirmed)

	return nil
}

var errNotPublished = errors.New("not published")

type referenceDay struct {
	Date   string `json:"date"`
	Events []struct {
		Description string `json:"description"`
		IsHoliday   bool   `json:"is_holiday"`
	} `json:"events"`
	IsHoliday bool `json:"is_holiday"`
}

func fetchYear(client *http.Client, year int) ([]referenceDay, error) {
	resp, err := client.Get(fmt.Sprintf(referenceURL, year)) //nolint:noctx // the client carries the timeout
	if err != nil {
		return nil, fmt.Errorf("fetching: %w", err)
	}

	defer resp.Body.Close() //nolint:errcheck // read-only body

	if resp.StatusCode == http.StatusNotFound {
		return nil, errNotPublished
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading: %w", err)
	}

	var days []referenceDay
	if err := json.Unmarshal(body, &days); err != nil {
		return nil, fmt.Errorf("decoding: %w", err)
	}

	if len(days) == 0 {
		return nil, errors.New("reference returned no days")
	}

	return days, nil
}

// reconcile returns the entries needed to turn the computed year into the
// reference year: pins for lunar events whose computed date is wrong or merely
// unconfirmed, and one-off events the rules know nothing about.
func reconcile(calendar *holiday.Calendar, year int, reference []referenceDay) ([]holiday.Entry, error) {
	referenceDates, err := indexReference(reference)
	if err != nil {
		return nil, err
	}

	computed := make(map[string]string, len(lunarRuleIDs))

	for _, day := range calendar.Events(year) {
		for _, event := range day.Events {
			if event.Kind == holiday.Lunar {
				computed[event.ID] = day.Date.Format(ptime.DateOnly)
			}
		}
	}

	entries := make([]holiday.Entry, 0, 8)

	for _, id := range lunarRuleIDs {
		want, found := referenceDates[id]
		if !found {
			continue
		}

		// eid-fitr-2 always trails eid-fitr by a day; the reference lists both
		// under the same title, so derive it rather than matching it.
		if id == "eid-fitr-2" {
			continue
		}

		entries = append(entries, holiday.Entry{ID: id, Date: want})

		if id == "eid-fitr" {
			next, err := dayAfter(want)
			if err != nil {
				return nil, err
			}

			entries = append(entries, holiday.Entry{ID: "eid-fitr-2", Date: next})
		}
	}

	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Date < entries[j].Date })

	return entries, nil
}

func indexReference(reference []referenceDay) (map[string]string, error) {
	dates := make(map[string]string, len(lunarRuleIDs))

	for _, day := range reference {
		for _, event := range day.Events {
			if !event.IsHoliday {
				continue
			}

			id := matchRule(event.Description)
			if id == "" {
				continue
			}

			// Keep the first occurrence: a Persian year can contain two
			// Muharrams, and the official calendar prints the first.
			if _, seen := dates[id]; !seen {
				dates[id] = day.Date
			}
		}
	}

	if len(dates) < len(keywords)/2 {
		return nil, fmt.Errorf("only %d of %d lunar events matched, the reference format probably changed",
			len(dates), len(keywords))
	}

	return dates, nil
}

func matchRule(description string) string {
	for id, groups := range keywords {
		for _, group := range groups {
			if containsAll(description, group) {
				return id
			}
		}
	}

	return ""
}

func containsAll(haystack string, needles []string) bool {
	for _, needle := range needles {
		if !strings.Contains(haystack, needle) {
			return false
		}
	}

	return true
}

func dayAfter(date string) (string, error) {
	pt, err := ptime.Parse(ptime.DateOnly, date)
	if err != nil {
		return "", fmt.Errorf("parsing %q: %w", date, err)
	}

	return pt.Tomorrow().Format(ptime.DateOnly), nil
}
