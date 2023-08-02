package chacha20poly1305

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	encr := strings.NewReader("Hello, gophers!")
	encw := bytes.NewBuffer(nil)
	key := make([]byte, 32)
	keyTooSmall := make([]byte, 14)
	rand.Read(key)
	rand.Read(keyTooSmall)

	err := Encrypt(encr, encw, keyTooSmall)
	require.Error(t, err)

	err = Encrypt(encr, encw, key)
	require.NoError(t, err)

	encBytes := encw.Bytes()

	decr := bytes.NewReader(encBytes)
	decw := bytes.NewBuffer(nil)
	err = Decrypt(decr, decw, key)
	require.NoError(t, err)

	assert.Equal(t, "Hello, gophers!", decw.String())

	err = Decrypt(decr, decw, keyTooSmall)
	assert.Error(t, err)
}

func TestGetCapacityForEncryptedValue(t *testing.T) {
	encr := strings.NewReader("Hello, gophers!")
	encw := bytes.NewBuffer(nil)
	key := make([]byte, 32)
	rand.Read(key)

	err := Encrypt(encr, encw, key)
	require.NoError(t, err)

	enccap := GetCapacityForEncryptedValue(int(encr.Size()))
	assert.LessOrEqual(t, encw.Len(), enccap)
}
