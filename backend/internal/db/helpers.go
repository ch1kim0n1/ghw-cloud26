package db

import "errors"

var ErrNotFound = errors.New("resource not found")

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
