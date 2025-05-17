package jst

import (
	"fmt"
	"log/slog"
	"time"
)

var jstLocation *time.Location

const (
	DateLayout = "2006-01-02"
	TimeLayout = "15:04"
)

func init() {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to load Asia/Tokyo timezone: %v. Falling back to UTC.", err))
		jstLocation = time.UTC
	} else {
		jstLocation = loc
	}
}

func Location() *time.Location                 { return jstLocation }
func Now() time.Time                           { return time.Now().In(jstLocation) }
func Format(t time.Time, layout string) string { return t.In(jstLocation).Format(layout) }
func FormatDate(t time.Time) string            { return Format(t, DateLayout) }
func FormatTime(t time.Time) string            { return Format(t, TimeLayout) }
func ParseInJST(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, jstLocation)
}
func ParseRFC3339AndConvertToJST(value string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, err
	}
	return t.In(jstLocation), nil
}
func GetAnnictSeason(t time.Time) string {
	jstTime := t.In(jstLocation)
	year := jstTime.Year()
	month := jstTime.Month()
	var season string
	switch {
	case month >= 1 && month <= 3:
		season = "winter"
	case month >= 4 && month <= 6:
		season = "spring"
	case month >= 7 && month <= 9:
		season = "summer"
	case month >= 10 && month <= 12:
		season = "autumn"
	}
	return fmt.Sprintf("%d-%s", year, season)
}
func IsSameDate(t1, t2 time.Time) bool {
	return t1.In(jstLocation).Truncate(24 * time.Hour).Equal(t2.In(jstLocation).Truncate(24 * time.Hour))
}
