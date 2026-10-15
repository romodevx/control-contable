package domain

import clock "time"

func saturday(d clock.Time) bool {
	return d.Weekday() == clock.Saturday
}
