package usecase

import "time"

func weekend(d time.Time) bool {
	return d.Weekday() == time.Saturday || d.Weekday() == time.Sunday
}

func monday() time.Weekday { return time.Monday }
