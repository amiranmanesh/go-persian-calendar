# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.5.0] - 2026-08-11

### Added

- A `holiday` subpackage that answers whether a Persian date is a public
  holiday and what is commemorated on it: `Iran()`, `Calendar.Lookup`,
  `IsHoliday`, `IsWorkday`, `Events`, `Holidays`, `NextHoliday`, `NextWorkday`
  and `Workdays`.
- Every occurrence carries a `Confidence`. Days fixed in the Persian calendar
  are always `Confirmed`; days fixed in the Hijri calendar are `Estimated`
  until their year has passed and the announced date has been recorded, because
  Iran determines them by moon sighting. Measured against 31 years of published
  calendars, the arithmetic Hijri calendar disagrees with the announced date
  43% of the time, almost always by one day — so the flag is not a formality.
- `holiday.Load` and `Calendar.WithOverrides` load a fresher copy of the data
  at runtime, for services that cannot rebuild to pick up a correction. The
  file is published at
  https://amiranmanesh.github.io/go-persian-calendar/data/v1/iran.json
- Bundled data covering Persian years 1390 to 1420, embedded at build time so
  the package performs no I/O and works offline.
- A monthly workflow reconciles the computed calendar against published Iranian
  calendars and opens a pull request when a date moves, so data changes are
  reviewed rather than applied silently.

### Changed

- Calendar conversions moved to an internal `jdn` package, shared by `ptime`
  and `holiday`. The public API of `ptime` is unchanged.
- The internal Julian Day machinery gained tabular Hijri conversions.

## [1.4.0] - 2026-08-11

First release published under `github.com/amiranmanesh/go-persian-calendar`. Earlier
tags in this repository were inherited from the upstream project and declared a
different module path, so they were never installable from here.

### Migrating from `github.com/yaa110/go-persian-calendar`

```diff
-import ptime "github.com/yaa110/go-persian-calendar"
+import ptime "github.com/amiranmanesh/go-persian-calendar"
```

The API is source compatible. Three behaviors changed, all of them bug fixes —
see **Fixed** below for the details.

### Added

- `Parse`, `ParseInLocation`, `ParseTimeFormat` and `ParseTimeFormatInLocation`,
  covering both layout languages, with a `*ParseError` that names the layout
  element that failed.
- `Time` now implements `json.Marshaler`, `json.Unmarshaler`,
  `encoding.TextMarshaler`, `encoding.TextUnmarshaler`, `driver.Valuer` and
  `sql.Scanner`. The zero `Time` round trips through JSON as `null` and through
  SQL as `NULL`.
- `Time.Sub` returning a signed `time.Duration`, plus `Time.Truncate` and
  `Time.Round`.
- `Time.AppendFormat` and `Time.AppendTimeFormat`, which write into a caller
  supplied buffer instead of allocating a string.
- `UnixMilli` and `UnixMicro`, both as constructors and as methods.
- Predefined layouts: `RFC3339`, `RFC3339Nano`, `DateTime`, `DateOnly`,
  `TimeOnly`, `Kitchen` and `LongDate`.
- Fractional second patterns `.000`/`.000000`/`.000000000` and
  `.999`/`.999999`/`.999999999` in the `Format` layout language.
- `Month.IsValid`, `Weekday.IsValid`.
- Runnable examples, benchmarks and three fuzz targets covering the parser and
  both round trips.

### Changed

- Module path is now `github.com/amiranmanesh/go-persian-calendar`.
- Minimum Go version is 1.21.
- The package is split into focused files (`ptime.go`, `format.go`, `parse.go`,
  `marshal.go`, `month.go`, `weekday.go`, `daypart.go`, `location.go`,
  `conversion.go`) instead of one 1400 line file.
- `Iran()` and `Afghanistan()` cache their lookup, so they no longer hit the file
  system on every call.
- `TimeFormat` is a single pass scanner rather than two `strings.Replacer`
  passes, which removes the risk of substituting inside already substituted text.
- `Format` writes through a byte buffer, so `Format` costs one allocation.
- `Time.String()` now emits RFC 3339: the fractional second is trimmed of
  trailing zeros and omitted entirely when zero, and the zero `Time` renders as
  `0000-00-00T00:00:00Z` instead of a value that depended on the host time zone.
- `Time.Since` is deprecated in favor of `Time.Sub`.

### Fixed

- Fractional seconds in `TimeFormat` were computed from the decimal digits of
  the nanosecond field rather than from its value. For 52065090 ns, `.000` now
  correctly renders `.052` instead of `.520`, and `.000000000` renders
  `.052065090` instead of `.52065090`.
- `ZoneOffset` produced malformed output for negative offsets with a non-zero
  minute part, for example `-04:-30` instead of `-04:30`.
- `Month.String`, `Weekday.String` and `AmPm.Short` clamped through overlapping
  bounds checks; the behavior is unchanged but the intent is now explicit.

[Unreleased]: https://github.com/amiranmanesh/go-persian-calendar/compare/v1.5.0...HEAD
[1.5.0]: https://github.com/amiranmanesh/go-persian-calendar/releases/tag/v1.5.0
[1.4.0]: https://github.com/amiranmanesh/go-persian-calendar/releases/tag/v1.4.0
