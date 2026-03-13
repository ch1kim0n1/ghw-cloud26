package services

import "time"

func TimestampNow() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}
