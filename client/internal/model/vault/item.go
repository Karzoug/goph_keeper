package vault

import (
	"bytes"
	"encoding/gob"
	"errors"

	"github.com/Karzoug/goph_keeper/client/pkg/crypto/chacha20poly1305"
	cvault "github.com/Karzoug/goph_keeper/common/model/vault"
	"github.com/Karzoug/goph_keeper/pkg/e"
)

var ErrUnknownVaultType = errors.New("unknown vault type")

type Item cvault.Item

func (item *Item) EncryptAndSetValue(data any, encrKey EncryptionKey) error {
	const op = "vault: encrypt and set value"

	br := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(br)
	if err := enc.Encode(data); err != nil {
		return e.Wrap(op, err)
	}

	bw := bytes.NewBuffer(nil)
	bw.Grow(chacha20poly1305.GetCapacityForEncryptedValue(br.Len()))

	if err := chacha20poly1305.Encrypt(br, bw, encrKey.Hash); err != nil {
		return e.Wrap(op, err)
	}

	item.Value = bw.Bytes()

	return nil
}

func (item Item) DecryptAnGetValue(encrKey EncryptionKey) (any, error) {
	const op = "vault: decrypt and get value"

	br := bytes.NewReader(item.Value)
	bw := bytes.NewBuffer(nil)

	if err := chacha20poly1305.Decrypt(br, bw, encrKey.Hash); err != nil {
		return nil, e.Wrap(op, err)
	}

	dec := gob.NewDecoder(bw)

	switch item.Type {
	case cvault.Password:
		psw := Password{}
		if err := dec.Decode(&psw); err != nil {
			return nil, e.Wrap(op, err)
		}
		return psw, nil
	case cvault.Card:
		crd := Card{}
		if err := dec.Decode(&crd); err != nil {
			return nil, e.Wrap(op, err)
		}
		return crd, nil
	case cvault.Text:
		txt := Text{}
		if err := dec.Decode(&txt); err != nil {
			return nil, e.Wrap(op, err)
		}
		return txt, nil
	case cvault.Binary:
		b := Binary{}
		if err := dec.Decode(&b); err != nil {
			return nil, e.Wrap(op, err)
		}
		return b, nil
	default:
		return nil, e.Wrap(op, ErrUnknownVaultType)
	}
}
