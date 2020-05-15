package unixtime

import "time"

// ToMilli ..
func ToMilli(t time.Time) int64 {
	return t.Round(time.Millisecond).UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

// NowToMilli ....
func NowToMilli() int64 {
	return ToMilli(time.Now())
}

// MilliToTime ...
func MilliToTime(m int64) time.Time {
	return time.Unix(m/1e3, (m%1e3)*int64(time.Millisecond)/int64(time.Nanosecond))
}
