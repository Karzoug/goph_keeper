package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKey_New_Verify(t *testing.T) {
	tests := []struct {
		name    string
		hash    []byte
		wantErr bool
	}{
		{
			name:    "valid",
			hash:    []byte("test"),
			wantErr: false,
		},
		{
			name:    "valid",
			hash:    []byte("*07S7A7$V*ufxm!4NKgTlrhSI4gk3BTReVAegz4XL52j$v11l09HfUhW7UB#fXG&JHFIyUaEVm$kxyr2iCUuo7z#kd*&vvRKN92"),
			wantErr: false,
		},
		{
			name:    "invalid: empty hash",
			hash:    []byte(""),
			wantErr: true,
		},
		{
			name:    "invalid: nil hash",
			hash:    []byte(""),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, err := NewKey(tt.hash)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NotNil(t, k)
			assert.True(t, k.Verify(tt.hash))
		})
	}
}

func TestKey_Verify(t *testing.T) {
	h := []byte("*07S7A7$V*ufxm!4NKgTlrhSI4gk3BTReVAegz4XL52j$v11l09HfUhW7UB#fXG&JHFIyUaEVm$kxyr2iCUuo7z#kd*&vvRKN92")
	k, err := NewKey(h)
	require.NoError(t, err)
	assert.NotNil(t, k)

	assert.True(t, k.Verify(h))
	assert.False(t, k.Verify([]byte("test")))
	assert.False(t, k.Verify([]byte("")))
	assert.False(t, k.Verify(nil))
}
