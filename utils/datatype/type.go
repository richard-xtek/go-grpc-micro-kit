package datatype

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// FloatToString ...
func FloatToString(value float64) string {
	result := strconv.FormatFloat(value, 'f', -1, 64)
	return result
}

//IntToString ...
func IntToString(num int) string {
	return strconv.Itoa(num)
}

// StringToInt ...
func StringToInt(str string) (int, error) {
	rs, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	return rs, nil
}

// StringToInt64 ...
func StringToInt64(str string) (int64, error) {
	n, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// StringToFloat64 ...
func StringToFloat64(str string) (float64, error) {
	rs, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, err
	}
	return rs, nil
}

// Trim ...
func Trim(str string) string {
	s := strings.Trim(str, "!¡ ^%")
	return s
}

// SplitPairCoin ...
func SplitPairCoin(str string) []string {
	rs := strings.Split(str, "-")
	return rs
}

// GetDurationIntervalSecond ...
func GetDurationIntervalSecond(s string) (int64, int64, string) {
	var intervalInSecond int64
	i, _ := StringToInt64(s[0 : len(s)-1])
	var units string
	switch s[len(s)-1 : len(s)] {
	case "m":
		units = "min"
		intervalInSecond = i * 60
	case "h":
		units = "hour"
		intervalInSecond = i * 60 * 60
	case "d":
		units = "day"
		intervalInSecond = i * 24 * 60 * 60
	case "w":
		units = "week"
		intervalInSecond = i * 7 * 24 * 60 * 60
	case "M":
		units = "month"
		m := time.Date(time.Now().Year(), time.Now().Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
		intervalInSecond = i * int64(m) * 24 * 60 * 60
	}

	return intervalInSecond, i, units
}

// ParseValueFromDecimalToString ...
func ParseValueFromDecimalToString(f float64, decimal int32) string {
	s := "%." + fmt.Sprintf("%d", decimal) + "f"
	str := fmt.Sprintf(s, f)
	return str
}
