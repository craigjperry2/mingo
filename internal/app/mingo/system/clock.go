package system

import "time"

type Clock func() time.Time

func NewClock() Clock {
	return time.Now
}

func ClockForTesting(timespec string) Clock {
	return func() time.Time {
		instant, _ := time.Parse(time.RFC3339, timespec)
		return instant
	}
}
