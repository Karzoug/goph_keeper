package vault

const (
	Unknown ItemType = iota
	Password
	Card
	Text
	Binary
	BinaryLarge
)

type ItemType int32

type Item struct {
	ID              string
	Name            string
	Type            ItemType
	Value           []byte
	ServerUpdatedAt int64
	ClientUpdatedAt int64
	IsDeleted       bool
}
