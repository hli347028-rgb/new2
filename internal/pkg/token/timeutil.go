package token

import "time"

// NowChina returns current time in China timezone.
func NowChina() time.Time {
	return time.Now().In(ChinaLocation())
}
