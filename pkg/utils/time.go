package utils

import (
	"time"
)

const (
	DateTimeFormat = "2006-01-02 15:04:05"
	DateFormat     = "2006-01-02"
	TimeFormat     = "15:04:05"
	RFC822Format   = time.RFC1123Z
)

func Now() time.Time {
	return time.Now().Local()
}

func NowPtr() *time.Time {
	t := Now()
	return &t
}

func ParseDateTime(s string) (time.Time, error) {
	return time.ParseInLocation(DateTimeFormat, s, time.Local)
}

func ParseDate(s string) (time.Time, error) {
	return time.ParseInLocation(DateFormat, s, time.Local)
}

func FormatDateTime(t time.Time) string {
	return t.Format(DateTimeFormat)
}

func FormatDate(t time.Time) string {
	return t.Format(DateFormat)
}

func FormatRFC822(t time.Time) string {
	return t.Format(RFC822Format)
}

func DurationString(seconds int) string {
	if seconds <= 0 {
		return "00:00"
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	if h > 0 {
		return time.Date(0, 1, 1, h, m, s, 0, time.UTC).Format("15:04:05")
	}
	return time.Date(0, 1, 1, 0, m, s, 0, time.UTC).Format("04:05")
}

func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

func StartOfWeek(t time.Time) time.Time {
	offset := int(t.Weekday() - time.Monday)
	if offset < 0 {
		offset += 7
	}
	return StartOfDay(t.AddDate(0, 0, -offset))
}

func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func DaysAgo(n int) time.Time {
	return StartOfDay(Now().AddDate(0, 0, -n))
}

func Today() time.Time {
	return StartOfDay(Now())
}

func AddDays(t time.Time, n int) time.Time {
	return t.AddDate(0, 0, n)
}

func Between(t, start, end time.Time) bool {
	return !t.Before(start) && !t.After(end)
}
