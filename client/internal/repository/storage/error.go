package storage

import "errors"

var (
	ErrRecordNotFound    = errors.New("record not found")
	ErrNoRecordsAffected = errors.New("no records affected")
)
