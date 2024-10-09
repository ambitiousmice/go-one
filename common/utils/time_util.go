package utils

import "time"

func IsSameDay(secondTime int64, time2 time.Time) bool {
	if secondTime <= 0 {
		return false
	}
	time1 := time.Unix(secondTime, 0)
	return time1.Year() == time2.Year() && time1.Month() == time2.Month() && time1.Day() == time2.Day()
}
