package vault

import (
	"bytes"
	"encoding/gob"

	"github.com/Karzoug/goph_keeper/common/model/vault"
)

type Item vault.Item

func (item *Item) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	dec := gob.NewDecoder(r)
	return dec.Decode(item)
}

func (item Item) MarshalBinary() ([]byte, error) {
	b := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(b)
	err := enc.Encode(item)
	return b.Bytes(), err
}
