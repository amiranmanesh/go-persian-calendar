package holiday

import (
	"sync"

	ptime "github.com/amiranmanesh/go-persian-calendar"
)

// A rule is a recurring event, fixed either in the Persian calendar (Solar) or
// in the Hijri calendar (Lunar). Month and day are read in whichever calendar
// the kind names.
type rule struct {
	id      string
	kind    Kind
	month   int
	day     int
	title   string
	holiday bool
}

// iranSolarRules are the days fixed in the Persian calendar. They are set by
// law, so they never move and are always Confirmed.
var iranSolarRules = []rule{
	{"nowruz", Solar, 1, 1, "جشن نوروز، آغاز سال نو", true},
	{"nowruz-2", Solar, 1, 2, "عید نوروز", true},
	{"nowruz-3", Solar, 1, 3, "عید نوروز", true},
	{"nowruz-4", Solar, 1, 4, "عید نوروز", true},
	{"islamic-republic-day", Solar, 1, 12, "روز جمهوری اسلامی ایران", true},
	{"nature-day", Solar, 1, 13, "روز طبیعت، سیزده به\u200cدر", true},
	{"khomeini-death", Solar, 3, 14, "رحلت حضرت امام خمینی", true},
	{"khordad-uprising", Solar, 3, 15, "قیام خونین ۱۵ خرداد", true},
	{"revolution-victory", Solar, 11, 22, "پیروزی انقلاب اسلامی", true},
	{"oil-nationalisation", Solar, 12, 29, "روز ملی شدن صنعت نفت ایران", true},
}

// iranLunarRules are the days fixed in the Hijri calendar. Their Persian date
// is computed and stays Estimated until the official calendar for the year is
// recorded in the data file.
var iranLunarRules = []rule{
	{"tasua", Lunar, 1, 9, "تاسوعای حسینی", true},
	{"ashura", Lunar, 1, 10, "عاشورای حسینی", true},
	{"arbaeen", Lunar, 2, 20, "اربعین حسینی", true},
	{"prophet-death", Lunar, 2, 28, "رحلت رسول اکرم؛ شهادت امام حسن مجتبی", true},
	{"reza-martyrdom", Lunar, 2, 30, "شهادت امام رضا", true},
	{"askari-martyrdom", Lunar, 3, 8, "شهادت امام حسن عسکری", true},
	{"prophet-birth", Lunar, 3, 17, "میلاد رسول اکرم و امام جعفر صادق", true},
	{"fatima-martyrdom", Lunar, 6, 3, "شهادت حضرت فاطمه زهرا", true},
	{"ali-birth", Lunar, 7, 13, "ولادت امام علی؛ روز پدر", true},
	{"mabath", Lunar, 7, 27, "مبعث رسول اکرم", true},
	{"mahdi-birth", Lunar, 8, 15, "ولادت حضرت قائم؛ جشن نیمه شعبان", true},
	{"ali-martyrdom", Lunar, 9, 21, "شهادت حضرت علی", true},
	{"eid-fitr", Lunar, 10, 1, "عید سعید فطر", true},
	{"eid-fitr-2", Lunar, 10, 2, "تعطیل به مناسبت عید سعید فطر", true},
	{"sadegh-martyrdom", Lunar, 10, 25, "شهادت امام جعفر صادق", true},
	{"eid-qorban", Lunar, 12, 10, "عید سعید قربان", true},
	{"eid-ghadir", Lunar, 12, 18, "عید سعید غدیر خم", true},
}

// iranObservances are commemorated but close nothing. The list is deliberately
// short: it covers the days most calendars print in red alongside the
// holidays, not every national day.
var iranObservances = []rule{
	{"world-nowruz-day", Solar, 1, 1, "روز جهانی نوروز", false},
	{"army-day", Solar, 1, 29, "روز ارتش جمهوری اسلامی ایران", false},
	{"teacher-day", Solar, 2, 12, "روز معلم", false},
	{"khalij-fars-day", Solar, 2, 10, "روز ملی خلیج فارس", false},
	{"khorramshahr-liberation", Solar, 3, 3, "آزادسازی خرمشهر", false},
	{"journalist-day", Solar, 5, 17, "روز خبرنگار", false},
	{"sacred-defence", Solar, 7, 1, "آغاز هفته دفاع مقدس", false},
	{"student-day", Solar, 9, 16, "روز دانشجو", false},
	{"yalda", Solar, 9, 30, "شب یلدا، جشن شب چله", false},
	{"revolution-decade", Solar, 11, 12, "بازگشت امام خمینی به ایران؛ آغاز دهه فجر", false},
	{"nuclear-technology-day", Solar, 1, 20, "روز ملی فناوری هسته\u200cای", false},
}

var iranCalendar = sync.OnceValue(func() *Calendar {
	rules := make([]rule, 0, len(iranSolarRules)+len(iranLunarRules)+len(iranObservances))
	rules = append(rules, iranSolarRules...)
	rules = append(rules, iranLunarRules...)
	rules = append(rules, iranObservances...)

	return &Calendar{
		name:      "iran",
		weekend:   []ptime.Weekday{ptime.Jomeh},
		rules:     rules,
		overrides: bundledIran(),
	}
})

// Iran returns the calendar of Iranian public holidays and observances.
//
// The weekly rest day is Friday. The result is shared and cached, so calling
// Iran repeatedly is cheap; it is safe for concurrent use.
func Iran() *Calendar {
	return iranCalendar()
}
