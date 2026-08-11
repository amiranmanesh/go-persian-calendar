// Copyright (c) 2016 Navid Fathollahzade
// Copyright (c) 2026 Amir Iranmanesh
//
// This source code is licensed under the MIT license found in the LICENSE file.

// Package ptime provides a complete implementation of the Persian
// (Solar Hijri / Jalali) calendar, designed to mirror the standard library
// [time] package as closely as possible.
//
// # Overview
//
// The central type is [Time], an immutable value describing a moment in the
// Persian calendar together with its [time.Location]. Conversions between the
// Persian and Gregorian calendars go through the Julian Day Number, which keeps
// them exact for both the Julian and the Gregorian era.
//
//	pt := ptime.Date(1394, ptime.Mehr, 2, 12, 59, 59, 0, ptime.Iran())
//	fmt.Println(pt.Time().Format(time.DateOnly)) // 2015-09-24
//
// # Creating a Time
//
// Use [Now] for the current moment, [Date] to build one from Persian calendar
// components, [New] to convert an existing [time.Time], [Unix] to convert a Unix
// timestamp, or [Parse] to read one from text.
//
// # Formatting and parsing
//
// Two layout languages are supported:
//
//   - The ptime layout language used by [Time.Format] and [Parse], built from
//     repeated pattern letters such as "yyyy/MM/dd HH:mm:ss".
//   - The standard library reference-time language used by [Time.TimeFormat]
//     and [ParseTimeFormat], such as "2006/01/02 15:04:05".
//
// Predefined layouts such as [RFC3339], [DateTime] and [DateOnly] cover the
// common cases.
//
// # Encoding
//
// [Time] implements [encoding.TextMarshaler], [encoding.TextUnmarshaler],
// [json.Marshaler], [json.Unmarshaler], [driver.Valuer] and [sql.Scanner], so it
// can be stored in JSON documents and SQL databases without a wrapper type.
//
// # Localization
//
// Both Iranian and Dari month names are available. [Time.TimeFormat] picks the
// Dari names automatically when the location is [Afghanistan].
//
// # Limitations
//
// Gregorian years below 1097 are not representable; [New] returns the zero
// [Time] for them.
package ptime
