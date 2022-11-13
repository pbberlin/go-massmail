package main

import (
	"fmt"
)

var weekdayTranslations = []map[string]string{
	{
		"en": "Sunday",
		"de": "Sonntag",
	},
	{
		"en": "Monday",
		"de": "Montag",
	},
	{
		"en": "Tuesday",
		"de": "Dienstag",
	},
	{
		"en": "Wednesday",
		"de": "Mittwoch",
	},
	{
		"en": "Thursday",
		"de": "Donnerstag",
	},
	{
		"en": "Friday",
		"de": "Freitag",
	},
	{
		"en": "Saturday",
		"de": "Sonnabend",
	},
}

// WeekdayByInt maps 1 to January, 12 to December
func WeekdayByInt(i int, lang string) string {
	if i < 0 || i > 7 {
		return fmt.Sprintf("error_unknown_weekday_idx__%v", i)
	}
	return weekdayTranslations[i][lang]
}

var monthsTranslations = []map[string]string{
	{
		"en": "January",
		"de": "Januar",
	},
	{
		"en": "February",
		"de": "Februar",
	},
	{
		"en": "March",
		"de": "März",
	},
	{
		"en": "April",
		"de": "April",
	},
	{
		"en": "May",
		"de": "Mai",
	},
	{
		"en": "June",
		"de": "Juni",
	},
	{
		"en": "July",
		"de": "Juli",
	},
	{
		"en": "August",
		"de": "August",
	},
	{
		"en": "September",
		"de": "September",
	},
	{
		"en": "October",
		"de": "Oktober",
	},
	{
		"en": "November",
		"de": "November",
	},
	{
		"en": "December",
		"de": "Dezember",
	},
}

// MonthByInt maps 1 to January, 12 to December
func MonthByInt(i int, lang string) string {
	if i < 1 || i > 12 {
		return fmt.Sprintf("error_unknown_month_idx__%v", i)
	}
	return monthsTranslations[i-1][lang]
}
