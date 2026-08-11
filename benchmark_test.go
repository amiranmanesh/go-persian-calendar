package ptime_test

import (
	"encoding/json"
	"testing"
	"time"

	ptime "github.com/amiranmanesh/go-persian-calendar"
)

var benchTime = ptime.Date(1394, ptime.Mehr, 2, 12, 59, 59, 52065090, ptime.Iran())

func BenchmarkNew(b *testing.B) {
	gt := time.Date(2015, time.September, 24, 12, 59, 59, 0, time.UTC)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sinkTime = ptime.New(gt)
	}
}

func BenchmarkTimeConversion(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sinkGoTime = benchTime.Time()
	}
}

func BenchmarkFormatRFC3339(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sinkString = benchTime.Format(ptime.RFC3339)
	}
}

func BenchmarkAppendFormat(b *testing.B) {
	buf := make([]byte, 0, 64)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf = benchTime.AppendFormat(buf[:0], ptime.RFC3339)
	}

	sinkBytes = buf
}

func BenchmarkTimeFormat(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sinkString = benchTime.TimeFormat("2006-01-02T15:04:05.999999999-07:00")
	}
}

func BenchmarkParse(b *testing.B) {
	const value = "1394-07-02T12:59:59+03:30"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sinkTime, errSink = ptime.Parse(ptime.RFC3339, value)
	}
}

func BenchmarkMarshalJSON(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sinkBytes, errSink = json.Marshal(benchTime)
	}
}

func BenchmarkIran(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sinkLocation = ptime.Iran()
	}
}

// Sinks keep the benchmarked results alive so the compiler cannot elide them.
var (
	sinkTime     ptime.Time
	sinkGoTime   time.Time
	sinkString   string
	sinkBytes    []byte
	errSink      error
	sinkLocation *time.Location
)
