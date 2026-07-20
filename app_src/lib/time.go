package lib

import "time"

func TimeNow() time.Time {
	return time.Now()
}

func TimeSince(t time.Time) time.Duration {
	return time.Since(t)
}
