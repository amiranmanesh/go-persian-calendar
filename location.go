package ptime

import (
	"sync"
	"time"
)

// Fallback offsets, in seconds east of UTC, used when the IANA time zone
// database is not available on the host.
const (
	iranFallbackOffset        = 12600 // UTC+03:30
	afghanistanFallbackOffset = 16200 // UTC+04:30
)

// Locations are looked up once and cached: time.LoadLocation hits the file
// system, and these two are on the hot path of almost every call.
var (
	iranLocation = sync.OnceValue(func() *time.Location {
		return loadLocation("Asia/Tehran", iranFallbackOffset)
	})

	afghanistanLocation = sync.OnceValue(func() *time.Location {
		return loadLocation("Asia/Kabul", afghanistanFallbackOffset)
	})
)

// Iran returns the "Asia/Tehran" time zone.
//
// If the time zone database is unavailable, a fixed UTC+03:30 zone is returned
// instead. The result is cached, so calling Iran repeatedly is cheap.
func Iran() *time.Location {
	return iranLocation()
}

// Afghanistan returns the "Asia/Kabul" time zone.
//
// If the time zone database is unavailable, a fixed UTC+04:30 zone is returned
// instead. The result is cached, so calling Afghanistan repeatedly is cheap.
func Afghanistan() *time.Location {
	return afghanistanLocation()
}

func loadLocation(name string, fallbackOffset int) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.FixedZone(name, fallbackOffset)
	}

	return loc
}
