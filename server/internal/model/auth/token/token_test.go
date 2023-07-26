package token

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromString(t *testing.T) {
	b := make([]byte, 20)
	_, err := rand.Read(b)
	require.NoError(t, err)

	sk := SecretKey(b)

	tkn := New(time.Now().Add(time.Hour), sk)
	s := tkn.String()

	tkn2, err := FromString(s, sk)
	require.NoError(t, err)

	assert.Equal(t, tkn.id, tkn.id)
	assert.WithinDuration(t, tkn.exp, tkn2.exp, 0)
}

func TestTimeMarshalBinary(t *testing.T) {
	b, _ := time.Now().MarshalBinary()
	require.Len(t, b, 15, "token format version 1 expected 15 bytes for time")
}
