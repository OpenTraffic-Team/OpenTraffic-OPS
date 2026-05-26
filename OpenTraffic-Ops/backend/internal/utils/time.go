package utils

import (
	"time"
)

const (
	TimeFormat         = "2006-01-02 15:04:05"
	TimeFormatDate     = "2006-01-02"
	TimeFormatTime     = "15:04:05"
	TimeFormatCompact  = "20060102150405"
	TimeFormatYearMonth = "2006-01"
)

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
	return t.Format(TimeFormat)
}

// FormatDate 格式化日期
func FormatDate(t time.Time) string {
	return t.Format(TimeFormatDate)
}

// ParseTime 解析时间字符串
func ParseTime(s string) (time.Time, error) {
	return time.ParseInLocation(TimeFormat, s, time.Local)
}

// ParseDate 解析日期字符串
func ParseDate(s string) (time.Time, error) {
	return time.ParseInLocation(TimeFormatDate, s, time.Local)
}

// NowStr 当前时间字符串
func NowStr() string {
	return time.Now().Format(TimeFormat)
}

// TodayStr 今天日期字符串
func TodayStr() string {
	return time.Now().Format(TimeFormatDate)
}

// StartOfDay 获取指定时间的当天开始时间
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay 获取指定时间的当天结束时间
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// StartOfMonth 获取指定时间的当月开始时间
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth 获取指定时间的当月结束时间
func EndOfMonth(t time.Time) time.Time {
	return StartOfMonth(t).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

// AddDays 增加天数
func AddDays(t time.Time, days int) time.Time {
	return t.AddDate(0, 0, days)
}

// DaysBetween 计算两个时间之间的天数差
func DaysBetween(t1, t2 time.Time) int {
	d := t2.Sub(t1)
	return int(d.Hours() / 24)
}

// TimestampMs 当前毫秒时间戳
func TimestampMs() int64 {
	return time.Now().UnixMilli()
}

// TimestampSec 当前秒时间戳
func TimestampSec() int64 {
	return time.Now().Unix()
}

// IsSameDay 判断是否为同一天
func IsSameDay(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() && t1.Month() == t2.Month() && t1.Day() == t2.Day()
}
