<div align="center" dir="rtl">

# تقویم فارسی برای Go

**پیاده‌سازی کامل و بدون وابستگی تقویم هجری شمسی (جلالی) برای زبان Go، با طراحی مشابه پکیج استاندارد `time`.**

[![Go Reference](https://pkg.go.dev/badge/github.com/amiranmanesh/go-persian-calendar.svg)](https://pkg.go.dev/github.com/amiranmanesh/go-persian-calendar)
[![CI](https://github.com/amiranmanesh/go-persian-calendar/actions/workflows/ci.yml/badge.svg)](https://github.com/amiranmanesh/go-persian-calendar/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/amiranmanesh/go-persian-calendar)](https://goreportcard.com/report/github.com/amiranmanesh/go-persian-calendar)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

[English](README.md) · [فارسی](README.fa.md)

</div>

---

<div dir="rtl">

پکیج `ptime` تبدیل بین تقویم هجری شمسی و میلادی را انجام می‌دهد و نوع `Time` را در اختیار شما می‌گذارد که مانند `time.Time` رفتار می‌کند: همان نام متدها، همان معنای مقداری (value semantics) و همان روش قالب‌بندی مبتنی بر layout. همهٔ تبدیل‌ها از مسیر «شمارهٔ روز جولیانی» انجام می‌شوند، بنابراین در دو سوی اصلاح تقویم گرگوری در سال ۱۵۸۲ میلادی نیز دقیق باقی می‌مانند.

## امکانات

- **API آشنا** — `Now` و `Date` و `Unix` و `Add` و `AddDate` و `Sub` و `Before` و `After` و `Equal` و `Compare` و `Truncate` و `Round`.
- **دو زبان قالب‌بندی** — سبک حروف الگو (`yyyy/MM/dd`) و سبک زمان مرجع کتابخانهٔ استاندارد (`2006/01/02`).
- **تجزیه (Parse) علاوه بر قالب‌بندی** — `Parse` و `ParseInLocation` و `ParseTimeFormat` و `ParseTimeFormatInLocation`.
- **آمادهٔ JSON و SQL** — پیاده‌سازی `json.Marshaler` و `encoding.TextMarshaler` و `driver.Valuer` و `sql.Scanner`.
- **نام‌های ایرانی و دری** — نام دری ماه‌ها برای موقعیت `Asia/Kabul` به‌صورت خودکار انتخاب می‌شود.
- **توابع کمکی تقویم** — سال کبیسه، هفتهٔ ماه، هفتهٔ سال، اولین و آخرین روزِ هفته و ماه و سال.
- **بدون هیچ وابستگی**، قالب‌بندی کم‌تخصیص (allocation-conscious) و تجزیه‌کنندهٔ آزموده‌شده با fuzz.

## نصب

</div>

```bash
go get github.com/amiranmanesh/go-persian-calendar
```

<div dir="rtl">

نیازمند Go نسخهٔ ۱٫۲۱ یا بالاتر.

</div>

```go
import ptime "github.com/amiranmanesh/go-persian-calendar"
```

<div dir="rtl">

## شروع سریع

### میلادی به شمسی

</div>

```go
gt := time.Date(2016, time.January, 1, 12, 1, 1, 0, ptime.Iran())

pt := ptime.New(gt)

fmt.Println(pt.Date()) // 1394 دی 11
```

<div dir="rtl">

### شمسی به میلادی

</div>

```go
pt := ptime.Date(1394, ptime.Mehr, 2, 12, 59, 59, 0, ptime.Iran())

fmt.Println(pt.Time().Format(time.DateOnly)) // 2015-09-24
```

<div dir="rtl">

### زمان کنونی

</div>

```go
pt := ptime.Now()

fmt.Println(pt.Date())                                    // 1394 بهمن 11
fmt.Println(pt.Clock())                                   // 21 54 30
fmt.Println(pt.Unix())                                    // 1454277270
fmt.Println(pt.Weekday())                                 // یک‌شنبه
fmt.Println(pt.Yesterday().Weekday())                     // شنبه
fmt.Println(pt.BeginningOfMonth().Format(ptime.DateOnly)) // 1394-11-01
fmt.Println(pt.LastMonthDay().Day())                      // 30
```

<div dir="rtl">

### قالب‌بندی

</div>

```go
pt := ptime.Unix(1454277270, 0)

pt.Format("yyyy/MM/dd E hh:mm:ss a") // 1394/11/11 یک‌شنبه 09:54:30 ب.ظ
pt.Format(ptime.RFC3339)             // 1394-11-11T21:54:30+03:30
pt.Format(ptime.LongDate)            // یک‌شنبه 11 بهمن 1394
pt.TimeFormat("2 Jan 2006")          // 11 بهمن 1394
```

<div dir="rtl">

نوشتن در بافر موجود، از تخصیص رشته جلوگیری می‌کند:

</div>

```go
buf = pt.AppendFormat(buf[:0], ptime.DateTime)
```

<div dir="rtl">

### تجزیه

</div>

```go
pt, err := ptime.Parse(ptime.DateTime, "1394-07-02 12:59:59")
if err != nil {
    // خطای *ptime.ParseError مشخص می‌کند کدام بخش از layout مطابقت نداشته است
}

pt, err = ptime.ParseInLocation("d MMM yyyy", "2 مهر 1394", ptime.Iran())
pt, err = ptime.ParseTimeFormat("2006/01/02", "1394/07/02")
```

<div dir="rtl">

اگر مقدار ورودی منطقهٔ زمانی نداشته باشد، `Parse` زمان را در UTC برمی‌گرداند؛ برای انتخاب پیش‌فرض دیگر از `ParseInLocation` استفاده کنید.

### JSON

</div>

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

<div dir="rtl">

مقدار صفرِ `Time` به `null` تبدیل می‌شود و برعکس.

### SQL

</div>

```go
var at ptime.Time

// متد Value یک زمان میلادی ذخیره می‌کند و Scan همان را بازمی‌خواند.
db.QueryRow("SELECT created_at FROM events WHERE id = $1", id).Scan(&at)
db.Exec("INSERT INTO events (created_at) VALUES ($1)", at)
```

<div dir="rtl">

## layoutهای از پیش تعریف‌شده

| ثابت | layout | نمونه |
|---|---|---|
| `RFC3339` | `yyyy-MM-ddTHH:mm:ssZ` | `1394-07-02T12:59:59+03:30` |
| `RFC3339Nano` | `yyyy-MM-ddTHH:mm:ss.999999999Z` | `1394-07-02T12:59:59.05206509+03:30` |
| `DateTime` | `yyyy-MM-dd HH:mm:ss` | `1394-07-02 12:59:59` |
| `DateOnly` | `yyyy-MM-dd` | `1394-07-02` |
| `TimeOnly` | `HH:mm:ss` | `12:59:59` |
| `Kitchen` | `h:mm a` | `12:59 ب.ظ` |
| `LongDate` | `E d MMM yyyy` | `پنج‌شنبه 2 مهر 1394` |

## مرجع کامل حروف الگو

جدول کامل حروف الگوی `Format` و `Parse` و همچنین زمان مرجعِ `TimeFormat` در [README انگلیسی](README.md#layout-reference) و در [مستندات pkg.go.dev](https://pkg.go.dev/github.com/amiranmanesh/go-persian-calendar) آمده است.

## موقعیت‌های زمانی

توابع `ptime.Iran()` و `ptime.Afghanistan()` مناطق `Asia/Tehran` و `Asia/Kabul` را برمی‌گردانند. هر دو پس از نخستین فراخوانی کش می‌شوند و اگر پایگاه‌دادهٔ منطقهٔ زمانی روی سیستم موجود نباشد، به یک اختلاف ثابت برمی‌گردند.

نام ماه‌ها از موقعیت پیروی می‌کند: متد `TimeFormat` در `Asia/Kabul` نام‌های دری (`میزان`) و در بقیهٔ موارد نام‌های ایرانی (`مهر`) را می‌نویسد. در `Format` این انتخاب صریح است: `MMM` برای ایرانی و `MMI` برای دری.

## محدودیت‌ها

- کوچک‌ترین سال میلادی قابل نمایش **۱۰۹۷** است؛ برای مقادیر قدیمی‌تر، `New` مقدار صفرِ `Time` را برمی‌گرداند.
- محاسبهٔ سال کبیسه بر پایهٔ قاعدهٔ حسابیِ چرخهٔ ۳۳ ساله است که در بازهٔ هدف این پکیج با تقویم رسمی ایران مطابقت دارد.

## توسعه

</div>

```bash
make test     # اجرای تست‌ها همراه با race و coverage
make lint     # اجرای golangci-lint
make bench    # بنچمارک‌ها
make fuzz     # یک دور کوتاه fuzz روی تجزیه‌کننده
make cover    # گزارش HTML پوشش تست
```

<div dir="rtl">

## مشارکت

گزارش مشکل و pull request پذیرفته می‌شود؛ راهنمای آن در [CONTRIBUTING.md](CONTRIBUTING.md) آمده است.

## سپاس‌گزاری

نسخهٔ اصلی این پکیج را [نوید فتح‌الله‌زاده](https://github.com/yaa110) با نام [yaa110/go-persian-calendar](https://github.com/yaa110/go-persian-calendar) نوشته است. این مخزن ادامهٔ همان کار است، با API به‌روزشده، تجزیه‌کننده، پشتیبانی از encoding و زنجیرهٔ ابزار بازنویسی‌شده.

## مجوز

MIT — متن کامل در [LICENSE](LICENSE).

</div>
