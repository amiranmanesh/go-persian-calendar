<div align="center">

# Go Persian Calendar

**A complete, dependency-free Persian (Solar Hijri / Jalali) calendar for Go — shaped like the standard `time` package.**

[![Go Reference](https://pkg.go.dev/badge/github.com/amiranmanesh/go-persian-calendar.svg)](https://pkg.go.dev/github.com/amiranmanesh/go-persian-calendar)
[![CI](https://github.com/amiranmanesh/go-persian-calendar/actions/workflows/ci.yml/badge.svg)](https://github.com/amiranmanesh/go-persian-calendar/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/amiranmanesh/go-persian-calendar)](https://goreportcard.com/report/github.com/amiranmanesh/go-persian-calendar)
[![Go Version](https://img.shields.io/github/go-mod/go-version/amiranmanesh/go-persian-calendar)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

[English](README.md) · [فارسی](README.fa.md)

</div>

---

`ptime` converts between the Persian and Gregorian calendars and gives you a `Time` type that behaves like `time.Time`: same method names, same value semantics, same layout-driven formatting. Conversions go through the Julian Day Number, so they stay exact on both sides of the 1582 Gregorian reform.

## Features

- **Familiar API** — `Now`, `Date`, `Unix`, `Add`, `AddDate`, `Sub`, `Before`, `After`, `Equal`, `Compare`, `Truncate`, `Round`.
- **Two layout languages** — the pattern-letter style (`yyyy/MM/dd`) and the standard library reference-time style (`2006/01/02`).
- **Parsing, not just formatting** — `Parse`, `ParseInLocation`, `ParseTimeFormat`, `ParseTimeFormatInLocation`.
- **Ready for JSON and SQL** — implements `json.Marshaler`, `encoding.TextMarshaler`, `driver.Valuer` and `sql.Scanner`.
- **Iranian and Dari names** — Dari month names are selected automatically for the `Asia/Kabul` location.
- **Calendar helpers** — leap years, week of month, week of year, first and last day of the week, month and year.
- **Public holidays** — the optional [`holiday`](#public-holidays) package answers whether a date is a day off, and says so honestly when a lunar date is still an estimate.
- **Zero dependencies**, allocation-conscious formatting, and a fuzz-tested parser.

## Installation

```bash
go get github.com/amiranmanesh/go-persian-calendar
```

Requires Go 1.21 or newer.

```go
import ptime "github.com/amiranmanesh/go-persian-calendar"
```

## Quick start

### Gregorian to Persian

```go
gt := time.Date(2016, time.January, 1, 12, 1, 1, 0, ptime.Iran())

pt := ptime.New(gt)

fmt.Println(pt.Date()) // 1394 دی 11
```

### Persian to Gregorian

```go
pt := ptime.Date(1394, ptime.Mehr, 2, 12, 59, 59, 0, ptime.Iran())

fmt.Println(pt.Time().Format(time.DateOnly)) // 2015-09-24
```

### The current moment

```go
pt := ptime.Now()

fmt.Println(pt.Date())                                  // 1394 بهمن 11
fmt.Println(pt.Clock())                                 // 21 54 30
fmt.Println(pt.Unix())                                  // 1454277270
fmt.Println(pt.Weekday())                               // یک‌شنبه
fmt.Println(pt.Yesterday().Weekday())                   // شنبه
fmt.Println(pt.BeginningOfMonth().Format(ptime.DateOnly)) // 1394-11-01
fmt.Println(pt.LastMonthDay().Day())                    // 30
fmt.Println(pt.IsLeap(), pt.YearWeek(), pt.MonthWeek())
```

### Formatting

```go
pt := ptime.Unix(1454277270, 0)

pt.Format("yyyy/MM/dd E hh:mm:ss a") // 1394/11/11 یک‌شنبه 09:54:30 ب.ظ
pt.Format(ptime.RFC3339)             // 1394-11-11T21:54:30+03:30
pt.Format(ptime.LongDate)            // یک‌شنبه 11 بهمن 1394
pt.TimeFormat("2 Jan 2006")          // 11 بهمن 1394
```

Formatting into an existing buffer avoids the string allocation:

```go
buf = pt.AppendFormat(buf[:0], ptime.DateTime)
```

### Parsing

```go
pt, err := ptime.Parse(ptime.DateTime, "1394-07-02 12:59:59")
if err != nil {
    // *ptime.ParseError reports which layout element failed and where
}

pt, err = ptime.ParseInLocation("d MMM yyyy", "2 مهر 1394", ptime.Iran())
pt, err = ptime.ParseTimeFormat("2006/01/02", "1394/07/02")
```

Without a time zone in the value, `Parse` returns a time in UTC — use `ParseInLocation` to pick a different default.

### JSON

```go
type Event struct {
    Name string     `json:"name"`
    At   ptime.Time `json:"at"`
}

json.Marshal(Event{
    Name: "نوروز",
    At:   ptime.Date(1404, ptime.Farvardin, 1, 0, 0, 0, 0, ptime.Iran()),
})
// {"name":"نوروز","at":"1404-01-01T00:00:00+03:30"}
```

The zero `Time` marshals to `null` and back.

### SQL

```go
var at ptime.Time

// Value() stores a Gregorian timestamp, Scan() reads one back.
db.QueryRow("SELECT created_at FROM events WHERE id = $1", id).Scan(&at)
db.Exec("INSERT INTO events (created_at) VALUES ($1)", at)
```

## Public holidays

```go
import "github.com/amiranmanesh/go-persian-calendar/holiday"

cal := holiday.Iran()

cal.IsHoliday(pt)                 // is this a day off, Fridays included?
cal.Lookup(pt).Title()            // "روز طبیعت، سیزده به‌در"
cal.NextWorkday(pt)               // the next working day
cal.Workdays(from, to)            // working days in a range
cal.Holidays(1404)                // every day off in a year
```

Iranian holidays come in three kinds, and the package treats each one honestly:

| Kind | Example | How it is resolved | Confidence |
|------|---------|--------------------|------------|
| Fixed in the Persian calendar | Nowruz, 22 Bahman | a rule in code, set by law | always `Confirmed` |
| Fixed in the Hijri calendar | Eid al-Fitr, Ashura | computed through the tabular Hijri calendar | `Estimated` until the year is settled |
| One-off | an air pollution closure | data only | `Confirmed` |

Iran fixes lunar dates by moon sighting, which no algorithm can predict: the
arithmetic calendar disagrees with the announced date about **43% of the time**,
almost always by a single day. So a lunar occurrence is reported as `Estimated`
until its year has passed and the announced date is recorded:

```go
for _, event := range cal.Lookup(pt).Events {
    if event.Confidence == holiday.Estimated {
        // Do not settle payroll on this date yet.
    }
}

cal.ConfirmedThrough() // the last year whose lunar dates are settled
```

The data is embedded at build time, so the package does no I/O and works
offline. `go get -u` brings newer data. A service that must pick up a correction
without rebuilding can load a fresher copy of the same file:

```go
resp, err := http.Get("https://amiranmanesh.github.io/go-persian-calendar/data/v1/iran.json")
// ...
overrides, err := holiday.Load(resp.Body)
cal = holiday.Iran().WithOverrides(overrides)
```

A monthly workflow reconciles the computed calendar against published Iranian
calendars and opens a pull request when a date moves, so every change to the
data is reviewed rather than applied silently.

## Predefined layouts

| Constant      | Layout                           | Example                              |
|---------------|----------------------------------|--------------------------------------|
| `RFC3339`     | `yyyy-MM-ddTHH:mm:ssZ`           | `1394-07-02T12:59:59+03:30`          |
| `RFC3339Nano` | `yyyy-MM-ddTHH:mm:ss.999999999Z` | `1394-07-02T12:59:59.05206509+03:30` |
| `DateTime`    | `yyyy-MM-dd HH:mm:ss`            | `1394-07-02 12:59:59`                |
| `DateOnly`    | `yyyy-MM-dd`                     | `1394-07-02`                         |
| `TimeOnly`    | `HH:mm:ss`                       | `12:59:59`                           |
| `Kitchen`     | `h:mm a`                         | `12:59 ب.ظ`                          |
| `LongDate`    | `E d MMM yyyy`                   | `پنج‌شنبه 2 مهر 1394`                |

## Layout reference

<details>
<summary><code>Format</code> and <code>Parse</code> — pattern letters</summary>

| Pattern | Meaning | Example |
|---------|---------|---------|
| `yyyy`, `yyy`, `y` | year | `1394` |
| `yy` | 2-digit year | `94` |
| `MMM` | Persian month name | `فروردین` |
| `MMI` | Dari month name | `حمل` |
| `MM` | 2-digit month | `01` |
| `M` | month | `1` |
| `dd` | 2-digit day of month | `01` |
| `d` | day of month | `1` |
| `E` | Persian weekday name | `شنبه` |
| `e` | short weekday name | `ش` |
| `A` | 12-hour marker | `قبل از ظهر` |
| `a` | short 12-hour marker | `ق.ظ` |
| `HH` / `H` | hour `[00-23]` / `[0-23]` | `09` / `9` |
| `kk` / `k` | hour `[01-24]` / `[1-24]` | `09` / `9` |
| `hh` / `h` | hour `[01-12]` / `[1-12]` | `09` / `9` |
| `KK` / `K` | hour `[00-11]` / `[0-11]` | `09` / `9` |
| `mm` / `m` | minute | `05` / `5` |
| `ss` / `s` | second | `05` / `5` |
| `n` | part of the day | `صبح` |
| `ns` | nanosecond, as a plain number | `52065090` |
| `S` | 3-digit millisecond | `052` |
| `.000` … `.000000000` | fractional second, fixed width | `.052` |
| `.999` … `.999999999` | fractional second, trailing zeros removed | `.052` |
| `D` / `RD` | day of year / remaining days of year | `186` |
| `w` / `rw` | week of year / remaining weeks of year | `46` |
| `W` / `rd` | week of month / remaining days of month | `3` |
| `z` | time zone name | `Asia/Tehran` |
| `Z` | time zone offset | `+03:30` |

Anything else is copied verbatim. Computed fields (`D`, `RD`, `w`, `rw`, `W`, `rd`, `E`, `e`, `n`) are matched and discarded when parsing.

</details>

<details>
<summary><code>TimeFormat</code> and <code>ParseTimeFormat</code> — reference time</summary>

| Layout | Meaning | Example |
|--------|---------|---------|
| `2006` / `06` | year | `1394` / `94` |
| `01` / `1` | month | `07` / `7` |
| `Jan`, `January` | month name | `مهر` |
| `02` / `2` / `_2` | day of month | `07` / `7` / `" 7"` |
| `Mon` / `Monday` | weekday | `ش` / `شنبه` |
| `Morning` | part of the day | `صبح` |
| `15` | hour `[00-23]` | `14` |
| `03` / `3` | hour `[01-12]` | `02` / `2` |
| `04` / `4` | minute | `07` / `7` |
| `05` / `5` | second | `08` / `8` |
| `.000` … `.999999999` | fractional second | `.052` |
| `PM` / `pm` | 12-hour marker | `بعد از ظهر` / `ب.ظ` |
| `MST` | time zone name | `Asia/Tehran` |
| `-0700`, `-07`, `-07:00`, `Z0700`, `Z07:00` | time zone offset | `+0330` |

</details>

## Locations

`ptime.Iran()` and `ptime.Afghanistan()` return `Asia/Tehran` and `Asia/Kabul`. Both are cached after the first lookup and fall back to a fixed offset when the host has no time zone database.

Month names follow the location: `TimeFormat` renders Dari names (`میزان`) in `Asia/Kabul` and Iranian names (`مهر`) everywhere else. `Format` chooses explicitly, with `MMM` for Iranian and `MMI` for Dari.

## Limitations

- The oldest representable Gregorian year is **1097**; `New` returns the zero `Time` for anything older.
- Leap years use the arithmetic 33-year cycle rule, which matches the official Iranian calendar over the range the package targets.

## Development

```bash
make test     # go test ./... -race -cover
make lint     # golangci-lint
make bench    # benchmarks
make fuzz     # a short fuzzing pass over the parser
make cover    # HTML coverage report
```

## Documentation

Full API documentation lives on [pkg.go.dev](https://pkg.go.dev/github.com/amiranmanesh/go-persian-calendar). Longer guides are in the [wiki](https://github.com/amiranmanesh/go-persian-calendar/wiki).

## Contributing

Issues and pull requests are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md).

## Credits

Originally created by [Navid Fathollahzade](https://github.com/yaa110) as [yaa110/go-persian-calendar](https://github.com/yaa110/go-persian-calendar), and maintained here with a modernized API, a parser, encoding support and a reworked toolchain.

## License

MIT — see [LICENSE](LICENSE).
