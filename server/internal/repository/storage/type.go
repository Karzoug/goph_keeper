package storage

import "errors"

const (
	SQLite Type = iota
)

const (
	sqliteString = "sqlite"
)

var ErrUnknownType = errors.New("unknown storage type")

type Type int8

func (e Type) String() string {
	switch e {
	case SQLite:
		return sqliteString
	default:
		return ErrUnknownType.Error()
	}
}

var TypeParserFunc = func(v string) (interface{}, error) {
	switch v {
	case sqliteString:
		return SQLite, nil
	default:
		return nil, ErrUnknownType
	}
}
