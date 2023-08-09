package token

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/rs/xid"
)

const (
	tokenVersion       byte = 1
	tokenSize               = 1 + 12 + 15 + 32
	MinSecretKeyLength      = 16
)

var (
	ErrInvalidTokenFormat = errors.New("invalid token format")
	ErrSecretKeyTooShort  = fmt.Errorf("secret key must be more or equal than %d bytes", MinSecretKeyLength)
)

// SecretKey is a key for signing tokens.
// It should be longer than MinSecretKeyLength bytes.
type SecretKey []byte

func (t *SecretKey) UnmarshalText(text []byte) error {
	if len(text) < 16 {
		return ErrSecretKeyTooShort
	}
	*t = text
	return nil
}

// token is a auth token
type token struct {
	id  string
	exp time.Time
	sub string
}

// New returns a new token with unique ID.
func New(exp time.Time, key SecretKey) *token {
	id := xid.New()

	t := &token{
		id:  id.String(),
		exp: exp,
	}

	b := make([]byte, tokenSize)

	b[0] = tokenVersion
	copy(b[1:13], id.Bytes())
	expBin, _ := exp.MarshalBinary()
	copy(b[13:], expBin)
	copy(b[28:], generateSign(b[:28], key))

	t.sub = hex.EncodeToString(b)

	return t
}

// FromString returns a token from a string.
func FromString(s string, key SecretKey) (*token, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, ErrInvalidTokenFormat
	}

	if len(b) != tokenSize {
		return nil, ErrInvalidTokenFormat
	}

	if b[0] != tokenVersion {
		return nil, ErrInvalidTokenFormat
	}

	id, err := xid.FromBytes(b[1:13])
	if err != nil {
		return nil, ErrInvalidTokenFormat
	}
	t := &token{
		id:  id.String(),
		sub: s,
	}
	err = t.exp.UnmarshalBinary(b[13:28])
	if err != nil {
		return nil, ErrInvalidTokenFormat
	}

	if !bytes.Equal(b[28:], generateSign(b[:28], key)) {
		return nil, ErrInvalidTokenFormat
	}

	return t, nil
}

// IsExpired returns true if token is expired.
func (t *token) IsExpired() bool {
	return t.exp.Before(time.Now())
}

// ID returns an ID of token to store on server.
func (t *token) ID() string {
	return t.sub
}

// String returns a string representation of token to send to client.
func (t *token) String() string {
	return t.sub
}

func generateSign(data []byte, key SecretKey) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
