package vault

import "time"

const (
	Password ItemType = iota
	Card
	Text
	Binary
	BinaryLarge
)

type ItemType int32

type Item struct {
	Name      string
	Type      ItemType
	Value     []byte
	UpdatedAt time.Time
}
