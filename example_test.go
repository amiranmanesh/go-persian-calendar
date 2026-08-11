package ptime_test

import (
	"encoding/json"
	"fmt"
	"time"

	ptime "github.com/amiranmanesh/go-persian-calendar"
)

func ExampleNew() {
	gt := time.Date(2016, time.January, 1, 12, 1, 1, 0, ptime.Iran())

	pt := ptime.New(gt)

	fmt.Println(pt.Date())
	// Output: 1394 دی 11
}

func ExampleDate() {
	pt := ptime.Date(1394, ptime.Mehr, 2, 12, 59, 59, 0, ptime.Iran())

	fmt.Println(pt.Time().Format(time.DateOnly))
	// Output: 2015-09-24
}

func ExampleTime_Format() {
	pt := ptime.Date(1394, ptime.Bahman, 11, 21, 54, 30, 0, ptime.Iran())

	fmt.Println(pt.Format("yyyy/MM/dd E hh:mm:ss a"))
	// Output: 1394/11/11 یک‌شنبه 09:54:30 ب.ظ
}

func ExampleTime_TimeFormat() {
	pt := ptime.Date(1394, ptime.Mehr, 2, 14, 7, 8, 0, ptime.Iran())

	fmt.Println(pt.TimeFormat("2 Jan 2006"))
	// Output: 2 مهر 1394
}

func ExampleParse() {
	pt, err := ptime.Parse("yyyy/MM/dd HH:mm:ss", "1394/07/02 12:59:59")
	if err != nil {
		panic(err)
	}

	fmt.Println(pt.Format(ptime.LongDate))
	// Output: پنج‌شنبه 2 مهر 1394
}

func ExampleParseInLocation() {
	pt, err := ptime.ParseInLocation(ptime.DateOnly, "1403-12-30", ptime.Iran())
	if err != nil {
		panic(err)
	}

	fmt.Println(pt.IsLeap(), pt.Format(ptime.RFC3339))
	// Output: true 1403-12-30T00:00:00+03:30
}

func ExampleTime_AddDate() {
	pt := ptime.Date(1394, ptime.Esfand, 29, 0, 0, 0, 0, ptime.Iran())

	fmt.Println(pt.AddDate(0, 0, 1).Format(ptime.DateOnly))
	// Output: 1395-01-01
}

func ExampleTime_MarshalJSON() {
	type event struct {
		Name string     `json:"name"`
		At   ptime.Time `json:"at"`
	}

	encoded, err := json.Marshal(event{
		Name: "نوروز",
		At:   ptime.Date(1404, ptime.Farvardin, 1, 0, 0, 0, 0, ptime.Iran()),
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(string(encoded))
	// Output: {"name":"نوروز","at":"1404-01-01T00:00:00+03:30"}
}

func ExampleTime_Sub() {
	from := ptime.Date(1394, ptime.Mehr, 2, 8, 0, 0, 0, ptime.Iran())
	to := ptime.Date(1394, ptime.Mehr, 3, 12, 30, 0, 0, ptime.Iran())

	fmt.Println(to.Sub(from))
	// Output: 28h30m0s
}

func ExampleMonth_Dari() {
	fmt.Println(ptime.Mehr.String(), ptime.Mehr.Dari())
	// Output: مهر میزان
}

func ExampleTime_BeginningOfMonth() {
	pt := ptime.Date(1394, ptime.Mehr, 17, 13, 45, 12, 0, ptime.Iran())

	fmt.Println(pt.BeginningOfMonth().Format(ptime.DateTime))
	fmt.Println(pt.LastMonthDay().Format(ptime.DateOnly))
	// Output:
	// 1394-07-01 00:00:00
	// 1394-07-30
}
