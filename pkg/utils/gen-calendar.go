package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/samber/lo"
)

var liturgicalDataPath = "https://raw.githubusercontent.com/v-bible/static/refs/heads/main/liturgical"

//nolint:gomnd
func EasterDate(y int) time.Time {
	c := math.Floor(float64(y) / 100)
	n := float64(y) - 19*math.Floor(float64(y)/19)
	k := math.Floor((c - 17) / 25)
	i := c - math.Floor(c/4) - math.Floor((c-k)/3) + 19*n + 15
	i -= 30 * math.Floor(i/30)
	i -= math.Floor(i/28) * (1 - math.Floor(i/28)*math.Floor(29/(i+1))*math.Floor((21-n)/11))
	j := float64(y) + math.Floor(float64(y)/4) + i + 2 - c + math.Floor(c/4)
	j -= 7 * math.Floor(j/7)
	l := i - j
	m := 3 + math.Floor((l+40)/44)
	d := l + 28 - 31*math.Floor(m/4)

	return time.Date(y, time.Month(int(m)), int(d), 0, 0, 0, 0, time.UTC)
}

type CalendarEntry struct {
	FirstReading  any    `json:"firstReading"`
	Psalm         any    `json:"psalm"`
	SecondReading any    `json:"secondReading"`
	Gospel        any    `json:"gospel"`
	YearCycle     string `json:"yearCycle"`
	YearNumber    string `json:"yearNumber"`
	Season        string `json:"season"`
	WeekdayType   string `json:"weekdayType"`
	WeekOrder     string `json:"weekOrder"`
	PeriodOfDay   string `json:"periodOfDay"`
	Description   string `json:"description"`
	ExtraCalendarEntry
}

type CalendarEntryData struct {
	CalendarEntry
	Weekday string `json:"weekday"`
	Date    string `json:"date"`
}

type ExtraCalendarEntry struct {
	SecondPsalm    any `json:"secondPsalm,omitempty"`
	ThirdReading   any `json:"thirdReading,omitempty"`
	ThirdPsalm     any `json:"thirdPsalm,omitempty"`
	FourthReading  any `json:"fourthReading,omitempty"`
	FourthPsalm    any `json:"fourthPsalm,omitempty"`
	FifthReading   any `json:"fifthReading,omitempty"`
	FifthPsalm     any `json:"fifthPsalm,omitempty"`
	SixthReading   any `json:"sixthReading,omitempty"`
	SixthPsalm     any `json:"sixthPsalm,omitempty"`
	SeventhReading any `json:"seventhReading,omitempty"`
	SeventhPsalm   any `json:"seventhPsalm,omitempty"`
	EighthReading  any `json:"eighthReading,omitempty"`
	EighthPsalm    any `json:"eighthPsalm,omitempty"`
}

type Options struct {
	IsEpiphanyOn6thJan         *bool
	IsAscensionOfTheLordOn40th *bool
	AdditionalCalendar         func(year int, options *Options) ([][]CalendarEntryData, error)
}

var YearCycleMap = []string{"C", "A", "B"}

var WeekdayMap = []string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}

func nextWeekday(time time.Time, weekday time.Weekday) time.Time {
	time = time.AddDate(0, 0, 1)
	for time.Weekday() != weekday {
		time = time.AddDate(0, 0, 1)
	}

	return time
}

func previousWeekday(time time.Time, weekday time.Weekday) time.Time {
	time = time.AddDate(0, 0, -1)
	for time.Weekday() != weekday {
		time = time.AddDate(0, 0, -1)
	}

	return time
}

func getLiturgical(url string) ([]CalendarEntry, error) {
	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	fileData, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var data []CalendarEntry

	err = json.Unmarshal(fileData, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func GenerateAdvent(year int) ([][]CalendarEntryData, error) {
	adventSundayData, err := getLiturgical(liturgicalDataPath + "/sunday/1_advent.json")
	if err != nil {
		return nil, err
	}

	adventWeekdayData, err := getLiturgical(liturgicalDataPath + "/weekdays/1_advent.json")
	if err != nil {
		return nil, err
	}

	yearCycle := YearCycleMap[year%3]

	christmasEve := time.Date(year-1, time.December, 24, 0, 0, 0, 0, time.UTC)
	christmasDay := time.Date(year-1, time.December, 25, 0, 0, 0, 0, time.UTC)

	// NOTE: Advent 4 is always the Sunday before Christmas, unless Advent 4 is on
	// Dec 24th, the morning is Advent 4, and then the afternoon is Christmas
	// Eve
	advent4 := previousWeekday(christmasDay, time.Sunday)
	if christmasEve.Weekday() == time.Sunday {
		advent4 = christmasEve
	}

	// NOTE: If Advent 4 is on Sunday then there is only 3 weeks of weekdays
	// Ref: https://catholic-resources.org/Lectionary/Overview-Advent.htm
	advent1 := advent4.AddDate(0, 0, -7*3)

	var calendar [][]CalendarEntryData = make([][]CalendarEntryData, 0)

	weekOrder := 0

	for day := advent1; day.Before(christmasEve.AddDate(0, 0, 1)); day = day.AddDate(0, 0, 1) {
		if day.Weekday() == time.Sunday {
			weekOrder += 1

			calendar = append(calendar, lo.FilterMap(adventSundayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
				if d.WeekOrder == fmt.Sprint(weekOrder) && d.YearCycle == yearCycle {
					return CalendarEntryData{
						CalendarEntry: d,
						Weekday:       strings.ToLower(day.Weekday().String()),
						Date:          day.Format("02/01/2006"),
					}, true
				}

				return CalendarEntryData{}, false
			}))

			continue
		}

		adventWeekday := lo.Filter(adventWeekdayData, func(d CalendarEntry, _ int) bool {
			if d.WeekOrder == fmt.Sprint(weekOrder) && d.WeekdayType == strings.ToLower(day.Weekday().String()) {
				return true
			}

			return false
		})

		// NOTE: Pre-Christmas weekdays
		if day.After(time.Date(year-1, time.December, 17, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)) && day.Before(christmasEve.AddDate(0, 0, 1)) {
			adventWeekday = lo.Filter(adventWeekdayData, func(d CalendarEntry, _ int) bool {
				if d.WeekOrder == "preChristmas" && d.WeekdayType == day.Format("02/01") {
					return true
				}

				return false
			})
		}

		calendar = append(calendar, lo.Map(adventWeekday, func(d CalendarEntry, _ int) CalendarEntryData {
			return CalendarEntryData{
				CalendarEntry: d,
				Weekday:       strings.ToLower(day.Weekday().String()),
				Date:          day.Format("02/01/2006"),
			}
		}))
	}

	return calendar, nil
}

func GenerateChristmas(year int, isEpiphanyOn6thJan bool) ([][]CalendarEntryData, error) {
	christmasSundayData, err := getLiturgical(liturgicalDataPath + "/sunday/2_christmas.json")
	if err != nil {
		return nil, err
	}

	christmasWeekdayData, err := getLiturgical(liturgicalDataPath + "/weekdays/2_christmas.json")
	if err != nil {
		return nil, err
	}

	yearCycle := YearCycleMap[year%3]
	christmasDay := time.Date(year-1, time.December, 25, 0, 0, 0, 0, time.UTC)

	// NOTE: The Epiphany is always on Jan 6th, or on the first Sunday after Jan
	// 1st
	defaultEpiphany := time.Date(year, time.January, 6, 0, 0, 0, 0, time.UTC)
	alternateEpiphany := nextWeekday(time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC), time.Sunday)

	epiphany := alternateEpiphany
	if isEpiphanyOn6thJan {
		epiphany = defaultEpiphany
	}

	var calendar [][]CalendarEntryData = make([][]CalendarEntryData, 0)

	calendar = append(calendar, lo.FilterMap(christmasSundayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
		if d.WeekOrder == "nativityOfTheLord" {
			return CalendarEntryData{
				CalendarEntry: d,
				Weekday:       strings.ToLower(christmasDay.Weekday().String()),
				Date:          christmasDay.Format("02/01/2006"),
			}, true
		}

		return CalendarEntryData{}, false
	}))

	for day := time.Date(year-1, time.December, 26, 0, 0, 0, 0, time.UTC); day.Before(epiphany.AddDate(0, 0, -1+1)); day = day.AddDate(0, 0, 1) {
		calendar = append(calendar, lo.FilterMap(christmasWeekdayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
			if d.WeekdayType == day.Format("02/01") && (d.WeekOrder == "preEpiphany" || d.WeekOrder == "christmas") {
				return CalendarEntryData{
					CalendarEntry: d,
					Weekday:       strings.ToLower(day.Weekday().String()),
					Date:          day.Format("02/01/2006"),
				}, true
			}

			return CalendarEntryData{}, false
		}))
	}

	// NOTE: If Christmas Day is on a Sunday, then the Feast of the Holy Family
	// celebrated on Dec 30th, else it is celebrated on the Sunday after
	// Christmas
	calendar = append(calendar, lo.FilterMap(christmasSundayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
		if d.WeekOrder == "theHolyFamily" && d.YearCycle == yearCycle {
			newDate := nextWeekday(christmasDay, time.Sunday)
			if christmasDay.Weekday() == time.Sunday {
				newDate = time.Date(year-1, time.December, 30, 0, 0, 0, 0, time.UTC)
			}

			newWeekday := strings.ToLower(newDate.Weekday().String())

			return CalendarEntryData{
				CalendarEntry: d,
				Weekday:       newWeekday,
				Date:          newDate.Format("02/01/2006"),
			}, true
		}

		return CalendarEntryData{}, false
	}))

	// NOTE: The Solemnity of Mary, Mother of God is always on Jan 1st
	calendar = append(calendar, lo.FilterMap(christmasSundayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
		if d.WeekOrder == "maryMotherOfGod" {
			newDate := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
			newWeekday := strings.ToLower(newDate.Weekday().String())

			return CalendarEntryData{
				CalendarEntry: d,
				Weekday:       newWeekday,
				Date:          newDate.Format("02/01/2006"),
			}, true
		}

		return CalendarEntryData{}, false
	}))

	// NOTE: 2nd Sunday after Christmas only when Epiphany is on Jan 6th and
	// before the Epiphany
	if isEpiphanyOn6thJan {
		secondSundayAfterChristmas := nextWeekday(christmasDay, time.Sunday).AddDate(0, 0, 7)

		if secondSundayAfterChristmas.Before(epiphany) {
			calendar = append(calendar, lo.FilterMap(christmasSundayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
				if d.WeekOrder == "2ndAfterChristmas" {
					return CalendarEntryData{
						CalendarEntry: d,
						Weekday:       strings.ToLower(secondSundayAfterChristmas.Weekday().String()),
						Date:          secondSundayAfterChristmas.Format("02/01/2006"),
					}, true
				}

				return CalendarEntryData{}, false
			}))
		}
	}

	// NOTE: The Epiphany of the Lord
	calendar = append(calendar, lo.FilterMap(christmasSundayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
		if d.WeekOrder == "theEpiphanyOfTheLord" {
			return CalendarEntryData{
				CalendarEntry: d,
				Weekday:       strings.ToLower(epiphany.Weekday().String()),
				Date:          epiphany.Format("02/01/2006"),
			}, true
		}

		return CalendarEntryData{}, false
	}))

	isEpiphanyOn7thOr8thJan := epiphany.Equal(time.Date(year, time.January, 7, 0, 0, 0, 0, time.UTC)) || epiphany.Equal(time.Date(year, time.January, 8, 0, 0, 0, 0, time.UTC))

	baptismOfTheLord := nextWeekday(epiphany, time.Sunday)
	if !isEpiphanyOn6thJan && (isEpiphanyOn7thOr8thJan) {
		baptismOfTheLord = nextWeekday(epiphany, time.Monday)
	}

	// NOTE: Post Epiphany weekdays
	startPostEpiphany := epiphany.AddDate(0, 0, 1)
	endPostEpiphany := baptismOfTheLord.AddDate(0, 0, -1)

	// NOTE: Sometimes first Sunday of Jan 6th can be early as Jan 7th so the
	// start and end is not correct. Also no calculate post Epiphany weekdays if
	// Epiphany is on Jan 7th or Jan 8th
	if startPostEpiphany.Before(endPostEpiphany) {
		for day := startPostEpiphany; day.Before(endPostEpiphany.AddDate(0, 0, 1)); day = day.AddDate(0, 0, 1) {
			if isEpiphanyOn6thJan {
				calendar = append(calendar, lo.FilterMap(christmasWeekdayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
					if d.WeekOrder == "postEpiphanyFromJan6" && d.WeekdayType == day.Format("02/01") {
						return CalendarEntryData{
							CalendarEntry: d,
							Weekday:       strings.ToLower(day.Weekday().String()),
							Date:          day.Format("02/01/2006"),
						}, true
					}

					return CalendarEntryData{}, false
				}))
			} else if !isEpiphanyOn6thJan && !isEpiphanyOn7thOr8thJan {
				calendar = append(calendar, lo.FilterMap(christmasWeekdayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
					if d.WeekOrder == "postEpiphany" && d.WeekdayType == strings.ToLower(day.Weekday().String()) {
						return CalendarEntryData{
							CalendarEntry: d,
							Weekday:       strings.ToLower(day.Weekday().String()),
							Date:          day.Format("02/01/2006"),
						}, true
					}

					return CalendarEntryData{}, false
				}))
			}
		}
	}

	calendar = append(calendar, lo.FilterMap(christmasSundayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
		if d.WeekOrder == "baptismOfTheLord" && d.YearCycle == yearCycle {
			return CalendarEntryData{
				CalendarEntry: d,
				Weekday:       strings.ToLower(baptismOfTheLord.Weekday().String()),
				Date:          baptismOfTheLord.Format("02/01/2006"),
			}, true
		}

		return CalendarEntryData{}, false
	}))

	return calendar, nil
}

func GenerateOT(year int, isEpiphanyOn6thJan bool) ([][]CalendarEntryData, error) {
	otSundayData, err := getLiturgical(liturgicalDataPath + "/sunday/5_ot.json")
	if err != nil {
		return nil, err
	}

	otWeekdayData, err := getLiturgical(liturgicalDataPath + "/weekdays/5_ot.json")
	if err != nil {
		return nil, err
	}

	yearCycle := YearCycleMap[year%3]
	yearNumber := 1

	if year%2 == 0 {
		yearNumber = 2
	}

	// NOTE: The Epiphany is always on Jan 6th, or on the first Sunday after Jan
	// 1st
	defaultEpiphany := time.Date(year, time.January, 6, 0, 0, 0, 0, time.UTC)
	alternateEpiphany := nextWeekday(time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC), time.Sunday)

	epiphany := alternateEpiphany
	if isEpiphanyOn6thJan {
		epiphany = defaultEpiphany
	}

	isEpiphanyOn7thOr8thJan := epiphany.Equal(time.Date(year, time.January, 7, 0, 0, 0, 0, time.UTC)) || epiphany.Equal(time.Date(year, time.January, 8, 0, 0, 0, 0, time.UTC))

	baptismOfTheLord := nextWeekday(epiphany, time.Sunday)
	if !isEpiphanyOn6thJan && (isEpiphanyOn7thOr8thJan) {
		baptismOfTheLord = nextWeekday(epiphany, time.Monday)
	}

	easterDay := EasterDate(year)
	ashWednesday := easterDay.AddDate(0, 0, -(7*6)-4)

	pentecost := easterDay.AddDate(0, 0, 49)

	christmasEve := time.Date(year, time.December, 24, 0, 0, 0, 0, time.UTC)
	christmasDay := time.Date(year, time.December, 25, 0, 0, 0, 0, time.UTC)

	// NOTE: Advent 4 is always the Sunday before Christmas, unless Advent 4 is on
	// Dec 24th, the morning is Advent 4, and then the afternoon is Christmas
	// Eve
	advent4 := previousWeekday(christmasDay, time.Sunday)
	if christmasEve.Weekday() == time.Sunday {
		advent4 = christmasEve
	}

	advent1 := advent4.AddDate(0, 0, -7*3)

	var calendar [][]CalendarEntryData = make([][]CalendarEntryData, 0)

	// NOTE: IF the Baptism of the Lord is on Monday then count as the first week
	weekOrder := 1
	if baptismOfTheLord.Weekday() == time.Sunday {
		weekOrder = 0
	}

	// NOTE: First OT
	for day := baptismOfTheLord; day.Before(ashWednesday.AddDate(0, 0, -1+1)); day = day.AddDate(0, 0, 1) {
		if day.Weekday() == time.Sunday {
			weekOrder += 1

			if day.Equal(baptismOfTheLord) {
				continue
			}

			calendar = append(calendar, lo.FilterMap(otSundayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
				if d.WeekOrder == fmt.Sprint(weekOrder) && d.YearCycle == yearCycle {
					return CalendarEntryData{
						CalendarEntry: d,
						Weekday:       strings.ToLower(day.Weekday().String()),
						Date:          day.Format("02/01/2006"),
					}, true
				}

				return CalendarEntryData{}, false
			}))

			continue
		}

		if day.Equal(baptismOfTheLord) {
			continue
		}

		calendar = append(calendar, lo.FilterMap(otWeekdayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
			if d.WeekOrder == fmt.Sprint(weekOrder) && d.WeekdayType == strings.ToLower(day.Weekday().String()) && d.YearNumber == fmt.Sprint(yearNumber) {
				return CalendarEntryData{
					CalendarEntry: d,
					Weekday:       strings.ToLower(day.Weekday().String()),
					Date:          day.Format("02/01/2006"),
				}, true
			}

			return CalendarEntryData{}, false
		}))
	}

	// NOTE: Second OT have to calculate so the last week before the Advent 1 is
	// the 34th Sunday
	// Ref: Checked from year 2023 -> 2100 here:
	// https://catholic-resources.org/Lectionary/Calendar.htm
	weekOrder = 34 - int(math.Ceil(advent1.Sub(pentecost).Hours()/24/7))

	if pentecost.Weekday() == time.Sunday {
		weekOrder += 1
	}

	for day := pentecost.AddDate(0, 0, 1); day.Before(advent1.AddDate(0, 0, -1+1)); day = day.AddDate(0, 0, 1) {
		if day.Weekday() == time.Sunday {
			weekOrder += 1

			calendar = append(calendar, lo.FilterMap(otSundayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
				if d.WeekOrder == fmt.Sprint(weekOrder) && d.YearCycle == yearCycle {
					return CalendarEntryData{
						CalendarEntry: d,
						Weekday:       strings.ToLower(day.Weekday().String()),
						Date:          day.Format("02/01/2006"),
					}, true
				}

				return CalendarEntryData{}, false
			}))

			continue
		}

		calendar = append(calendar, lo.FilterMap(otWeekdayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
			if d.WeekOrder == fmt.Sprint(weekOrder) && d.WeekdayType == strings.ToLower(day.Weekday().String()) && d.YearNumber == fmt.Sprint(yearNumber) {
				return CalendarEntryData{
					CalendarEntry: d,
					Weekday:       strings.ToLower(day.Weekday().String()),
					Date:          day.Format("02/01/2006"),
				}, true
			}

			return CalendarEntryData{}, false
		}))
	}

	return calendar, nil
}

func GenerateLent(year int) ([][]CalendarEntryData, error) {
	lentSundayData, err := getLiturgical(liturgicalDataPath + "/sunday/3_lent.json")
	if err != nil {
		return nil, err
	}

	lentWeekdayData, err := getLiturgical(liturgicalDataPath + "/weekdays/3_lent.json")
	if err != nil {
		return nil, err
	}

	yearCycle := YearCycleMap[year%3]

	easterDay := EasterDate(year)
	ashWednesday := easterDay.AddDate(0, 0, -(7*6)-4)

	chrismMass := previousWeekday(easterDay, time.Thursday)

	var calendar [][]CalendarEntryData = make([][]CalendarEntryData, 0)

	calendar = append(calendar, lo.FilterMap(lentWeekdayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
		if d.WeekdayType == "ashWednesday" {
			return CalendarEntryData{
				CalendarEntry: d,
				Weekday:       strings.ToLower(ashWednesday.Weekday().String()),
				Date:          ashWednesday.Format("02/01/2006"),
			}, true
		}

		return CalendarEntryData{}, false
	}))

	calendar = append(calendar, lo.FilterMap(lentWeekdayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
		if d.WeekdayType == "chrismMass" {
			return CalendarEntryData{
				CalendarEntry: d,
				Weekday:       strings.ToLower(chrismMass.Weekday().String()),
				Date:          chrismMass.Format("02/01/2006"),
			}, true
		}

		return CalendarEntryData{}, false
	}))

	// NOTE: Post Ash Wednesday weekdays
	for day := ashWednesday.AddDate(0, 0, 1); day.Before(nextWeekday(ashWednesday, time.Sunday).AddDate(0, 0, -1+1)); day = day.AddDate(0, 0, 1) {
		calendar = append(calendar, lo.FilterMap(lentWeekdayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
			if d.WeekOrder == "postAshWednesday" && d.WeekdayType == strings.ToLower(day.Weekday().String()) {
				return CalendarEntryData{
					CalendarEntry: d,
					Weekday:       strings.ToLower(day.Weekday().String()),
					Date:          day.Format("02/01/2006"),
				}, true
			}

			return CalendarEntryData{}, false
		}))
	}

	weekOrder := 0

	for day := nextWeekday(ashWednesday, time.Sunday); day.Before(previousWeekday(easterDay, time.Sunday).AddDate(0, 0, 1)); day = day.AddDate(0, 0, 1) {
		if day.Weekday() == time.Sunday {
			weekOrder += 1

			// NOTE: Palm Sunday
			if weekOrder == 6 {
				calendar = append(calendar, lo.FilterMap(lentSundayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
					if d.WeekOrder == "palmSunday" && d.YearCycle == yearCycle {
						return CalendarEntryData{
							CalendarEntry: d,
							Weekday:       strings.ToLower(day.Weekday().String()),
							Date:          day.Format("02/01/2006"),
						}, true
					}

					return CalendarEntryData{}, false
				}))
			} else {
				calendar = append(calendar, lo.FilterMap(lentSundayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
					if d.WeekOrder == fmt.Sprint(weekOrder) && d.YearCycle == yearCycle {
						return CalendarEntryData{
							CalendarEntry: d,
							Weekday:       strings.ToLower(day.Weekday().String()),
							Date:          day.Format("02/01/2006"),
						}, true
					}

					return CalendarEntryData{}, false
				}))
			}

			continue
		}

		calendar = append(calendar, lo.FilterMap(lentWeekdayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
			if d.WeekOrder == fmt.Sprint(weekOrder) && d.WeekdayType == strings.ToLower(day.Weekday().String()) {
				return CalendarEntryData{
					CalendarEntry: d,
					Weekday:       strings.ToLower(day.Weekday().String()),
					Date:          day.Format("02/01/2006"),
				}, true
			}

			return CalendarEntryData{}, false
		}))
	}

	// NOTE: Holy week
	for day := previousWeekday(easterDay, time.Sunday).AddDate(0, 0, 1); day.Before(previousWeekday(easterDay, time.Wednesday).AddDate(0, 0, 1)); day = day.AddDate(0, 0, 1) {
		calendar = append(calendar, lo.FilterMap(lentWeekdayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
			if d.WeekOrder == "holyWeek" && d.WeekdayType == strings.ToLower(day.Weekday().String()) {
				return CalendarEntryData{
					CalendarEntry: d,
					Weekday:       strings.ToLower(day.Weekday().String()),
					Date:          day.Format("02/01/2006"),
				}, true
			}

			return CalendarEntryData{}, false
		}))
	}

	return calendar, nil
}

func GenerateEaster(year int, isAscensionOfTheLordOn40th bool) ([][]CalendarEntryData, error) {
	triduumSundayData, err := getLiturgical(liturgicalDataPath + "/sunday/4_triduum.json")
	if err != nil {
		return nil, err
	}

	easterWeekdayData, err := getLiturgical(liturgicalDataPath + "/weekdays/4_easter.json")
	if err != nil {
		return nil, err
	}

	yearCycle := YearCycleMap[year%3]

	easterDay := EasterDate(year)

	holyThursday := easterDay.AddDate(0, 0, -3)
	goodFriday := easterDay.AddDate(0, 0, -2)
	easterVirgil := easterDay.AddDate(0, 0, -1)

	pentecost := easterDay.AddDate(0, 0, 49)

	ascensionOfTheLord := easterDay.AddDate(0, 0, 39)

	var calendar [][]CalendarEntryData = make([][]CalendarEntryData, 0)

	calendar = append(calendar, lo.FilterMap(triduumSundayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
		if d.WeekOrder == "holyThursday" {
			return CalendarEntryData{
				CalendarEntry: d,
				Weekday:       strings.ToLower(holyThursday.Weekday().String()),
				Date:          holyThursday.Format("02/01/2006"),
			}, true
		} else if d.WeekOrder == "goodFriday" {
			return CalendarEntryData{
				CalendarEntry: d,
				Weekday:       strings.ToLower(goodFriday.Weekday().String()),
				Date:          goodFriday.Format("02/01/2006"),
			}, true
		} else if d.WeekOrder == "easter" && d.PeriodOfDay == "theEasterVirgil" && d.YearCycle == yearCycle {
			return CalendarEntryData{
				CalendarEntry: d,
				Weekday:       strings.ToLower(easterVirgil.Weekday().String()),
				Date:          easterVirgil.Format("02/01/2006"),
			}, true
		} else if d.WeekOrder == "easter" && d.PeriodOfDay == "day" {
			return CalendarEntryData{
				CalendarEntry: d,
				Weekday:       strings.ToLower(easterDay.Weekday().String()),
				Date:          easterDay.Format("02/01/2006"),
			}, true
		}

		return CalendarEntryData{}, false
	}))

	// NOTE: Octave of Easter weekdays
	for day := nextWeekday(easterDay, time.Monday); day.Before(nextWeekday(easterDay, time.Saturday).AddDate(0, 0, 1)); day = day.AddDate(0, 0, 1) {
		calendar = append(calendar, lo.FilterMap(easterWeekdayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
			if d.WeekOrder == "octaveOfEaster" && d.WeekdayType == strings.ToLower(day.Weekday().String()) {
				return CalendarEntryData{
					CalendarEntry: d,
					Weekday:       strings.ToLower(day.Weekday().String()),
					Date:          day.Format("02/01/2006"),
				}, true
			}

			return CalendarEntryData{}, false
		}))
	}

	// NOTE: Start with the Sunday after Easter, but first is already Easter
	weekOrder := 1

	for day := nextWeekday(easterDay, time.Sunday); day.Before(pentecost.AddDate(0, 0, 1)); day = day.AddDate(0, 0, 1) {
		if day.Weekday() == time.Sunday {
			weekOrder += 1

			// NOTE: Pentecost Sunday
			if weekOrder == 8 {
				calendar = append(calendar, lo.FilterMap(triduumSundayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
					// NOTE: Pentecost Sunday
					if d.WeekOrder == "pentecost" && (d.YearCycle == yearCycle || d.YearCycle == "") {
						return CalendarEntryData{
							CalendarEntry: d,
							Weekday:       strings.ToLower(day.Weekday().String()),
							Date:          day.Format("02/01/2006"),
						}, true
					}

					return CalendarEntryData{}, false
				}))
			} else if !isAscensionOfTheLordOn40th && weekOrder == 7 {
				calendar = append(calendar, lo.FilterMap(triduumSundayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
					// NOTE: Pentecost Sunday
					if d.WeekOrder == "ascensionOfTheLord" && d.YearCycle == yearCycle {
						return CalendarEntryData{
							CalendarEntry: d,
							Weekday:       strings.ToLower(day.Weekday().String()),
							Date:          day.Format("02/01/2006"),
						}, true
					}

					return CalendarEntryData{}, false
				}))
			} else {
				calendar = append(calendar, lo.FilterMap(triduumSundayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
					if d.WeekOrder == fmt.Sprint(weekOrder) && d.YearCycle == yearCycle {
						return CalendarEntryData{
							CalendarEntry: d,
							Weekday:       strings.ToLower(day.Weekday().String()),
							Date:          day.Format("02/01/2006"),
						}, true
					}

					return CalendarEntryData{}, false
				}))
			}

			continue
		}

		calendar = append(calendar, lo.FilterMap(easterWeekdayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
			if d.WeekOrder == fmt.Sprint(weekOrder) && d.WeekdayType == strings.ToLower(day.Weekday().String()) {
				return CalendarEntryData{
					CalendarEntry: d,
					Weekday:       strings.ToLower(day.Weekday().String()),
					Date:          day.Format("02/01/2006"),
				}, true
			}

			return CalendarEntryData{}, false
		}))
	}

	if isAscensionOfTheLordOn40th {
		calendar = append(calendar, lo.FilterMap(triduumSundayData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
			if d.WeekOrder == "ascensionOfTheLord" && d.YearCycle == yearCycle {
				return CalendarEntryData{
					CalendarEntry: d,
					Weekday:       strings.ToLower(ascensionOfTheLord.Weekday().String()),
					Date:          ascensionOfTheLord.Format("02/01/2006"),
				}, true
			}

			return CalendarEntryData{}, false
		}))
	}

	return calendar, nil
}

func GenerateCelebration(year int) ([][]CalendarEntryData, error) {
	saintData, err := getLiturgical(liturgicalDataPath + "/celebrations/1_saint.json")
	if err != nil {
		return nil, err
	}

	movableCelebrationData, err := getLiturgical(liturgicalDataPath + "/celebrations/2_movable_celebrations.json")
	if err != nil {
		return nil, err
	}

	var calendar [][]CalendarEntryData = make([][]CalendarEntryData, 0)

	calendar = append(calendar, lo.FilterMap(append(saintData, movableCelebrationData...), func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
		parsedDate, err := time.Parse("02/01", d.WeekdayType)
		if err != nil {
			return CalendarEntryData{}, false
		}

		return CalendarEntryData{
			CalendarEntry: d,
			Weekday:       strings.ToLower(parsedDate.AddDate(year, 0, 0).Weekday().String()),
			Date:          parsedDate.AddDate(year, 0, 0).Format("02/01/2006"),
		}, true
	}))

	return calendar, nil
}

func GenerateAnnunciationOfTheLord(year int) ([][]CalendarEntryData, error) {
	movableCelebrationData, err := getLiturgical(liturgicalDataPath + "/celebrations/2_movable_celebrations.json")
	if err != nil {
		return nil, err
	}

	easterDay := EasterDate(year)
	ashWednesday := easterDay.AddDate(0, 0, -(7*6)-4)

	annunciationOfTheLord := time.Date(year, time.March, 25, 0, 0, 0, 0, time.UTC)

	var calendar [][]CalendarEntryData = make([][]CalendarEntryData, 0)

	if annunciationOfTheLord.Weekday() == time.Sunday && annunciationOfTheLord.After(ashWednesday.AddDate(0, 0, -1)) && annunciationOfTheLord.Before(easterDay.AddDate(0, 0, -14+1)) {
		annunciationOfTheLord = time.Date(year, time.March, 26, 0, 0, 0, 0, time.UTC)
	} else if annunciationOfTheLord.After(previousWeekday(easterDay, time.Sunday).AddDate(0, 0, -1)) && annunciationOfTheLord.Before(nextWeekday(easterDay, time.Sunday).AddDate(0, 0, 1)) {
		annunciationOfTheLord = nextWeekday(nextWeekday(easterDay, time.Sunday), time.Monday)
	}

	calendar = append(calendar, lo.FilterMap(movableCelebrationData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
		if d.WeekdayType == "annunciationOfTheLord" {
			return CalendarEntryData{
				CalendarEntry: d,
				Weekday:       strings.ToLower(annunciationOfTheLord.Weekday().String()),
				Date:          annunciationOfTheLord.Format("02/01/2006"),
			}, true
		}

		return CalendarEntryData{}, false
	}))

	return calendar, nil
}

func GeneratePostPentecostSolemnity(year int) ([][]CalendarEntryData, error) {
	movableCelebrationData, err := getLiturgical(liturgicalDataPath + "/celebrations/2_movable_celebrations.json")
	if err != nil {
		return nil, err
	}

	yearCycle := YearCycleMap[year%3]

	easterDay := EasterDate(year)

	pentecost := easterDay.AddDate(0, 0, 49)

	var calendar [][]CalendarEntryData = make([][]CalendarEntryData, 0)

	trinitySunday := nextWeekday(pentecost, time.Sunday)

	calendar = append(calendar, lo.FilterMap(movableCelebrationData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
		if d.WeekdayType == "trinitySunday" && d.YearCycle == yearCycle {
			return CalendarEntryData{
				CalendarEntry: d,
				Weekday:       strings.ToLower(trinitySunday.Weekday().String()),
				Date:          trinitySunday.Format("02/01/2006"),
			}, true
		}

		return CalendarEntryData{}, false
	}))

	bodyAndBloodOfChrist := nextWeekday(trinitySunday, time.Sunday)

	calendar = append(calendar, lo.FilterMap(movableCelebrationData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
		if d.WeekdayType == "bodyAndBloodOfChrist" && d.YearCycle == yearCycle {
			return CalendarEntryData{
				CalendarEntry: d,
				Weekday:       strings.ToLower(bodyAndBloodOfChrist.Weekday().String()),
				Date:          bodyAndBloodOfChrist.Format("02/01/2006"),
			}, true
		}

		return CalendarEntryData{}, false
	}))

	sacredHeart := nextWeekday(bodyAndBloodOfChrist, time.Friday)

	calendar = append(calendar, lo.FilterMap(movableCelebrationData, func(d CalendarEntry, _ int) (CalendarEntryData, bool) {
		if d.WeekdayType == "sacredHeart" && d.YearCycle == yearCycle {
			return CalendarEntryData{
				CalendarEntry: d,
				Weekday:       strings.ToLower(sacredHeart.Weekday().String()),
				Date:          sacredHeart.Format("02/01/2006"),
			}, true
		}

		return CalendarEntryData{}, false
	}))

	return calendar, nil
}

func GenerateCalendar(year int, options *Options) ([]CalendarEntryData, error) {
	isEpiphanyOn6thJan := false
	isAscensionOfTheLordOn40th := false

	var additionalCalendar func(year int, options *Options) ([][]CalendarEntryData, error) = nil

	if options != nil {
		if options.IsEpiphanyOn6thJan != nil {
			isEpiphanyOn6thJan = *options.IsEpiphanyOn6thJan
		}

		if options.IsAscensionOfTheLordOn40th != nil {
			isAscensionOfTheLordOn40th = *options.IsAscensionOfTheLordOn40th
		}

		if options.AdditionalCalendar != nil {
			additionalCalendar = options.AdditionalCalendar
		}
	}

	var calendar [][]CalendarEntryData = make([][]CalendarEntryData, 0)

	advent, err := GenerateAdvent(year)
	if err != nil {
		return nil, err
	}

	christmas, err := GenerateChristmas(year, isEpiphanyOn6thJan)
	if err != nil {
		return nil, err
	}

	ot, err := GenerateOT(year, isEpiphanyOn6thJan)
	if err != nil {
		return nil, err
	}

	lent, err := GenerateLent(year)
	if err != nil {
		return nil, err
	}

	easter, err := GenerateEaster(year, isAscensionOfTheLordOn40th)
	if err != nil {
		return nil, err
	}

	saintCelebration, err := GenerateCelebration(year)
	if err != nil {
		return nil, err
	}

	annunciationOfTheLord, err := GenerateAnnunciationOfTheLord(year)
	if err != nil {
		return nil, err
	}

	postPentecostSolemnity, err := GeneratePostPentecostSolemnity(year)
	if err != nil {
		return nil, err
	}

	calendar = append(calendar, advent...)
	calendar = append(calendar, christmas...)
	calendar = append(calendar, ot...)
	calendar = append(calendar, lent...)
	calendar = append(calendar, easter...)
	calendar = append(calendar, saintCelebration...)
	calendar = append(calendar, annunciationOfTheLord...)
	calendar = append(calendar, postPentecostSolemnity...)

	if additionalCalendar != nil {
		additional, err := additionalCalendar(year, options)
		if err != nil {
			return nil, err
		}

		calendar = append(calendar, additional...)
	}

	calendarFlat := lo.Flatten(calendar)

	slices.SortFunc(calendarFlat, func(a, b CalendarEntryData) int {
		dateA, _ := time.Parse("02/01/2006", a.Date)
		dateB, _ := time.Parse("02/01/2006", b.Date)

		return dateB.Compare(dateA)
	})

	slices.Reverse(calendarFlat)

	return calendarFlat, nil
}
