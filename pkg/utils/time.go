package utils

import (
	"time"
)

const (
	Month_TIME_LAYOUT             = "2006-01"
	DEFAULT_TIME_LAYOUT           = "2006-01-02"
	DEFAULT_DB_TIME_LAYOUT        = "2006-01-02 15:04:05"
	DEFAULT_TIMESTAMP_TIME_LAYOUT = "20060102150405"
)

func GetTimeLocation(location string) (*time.Location, error) {
	return time.LoadLocation(location)
}
