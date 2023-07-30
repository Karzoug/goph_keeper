package crypto

import (
	"crypto/subtle"
	"runtime"
)

// Wipe takes a buffer and wipes it with zeroes.
func Wipe(buf []byte) {
	for i := range buf {
		buf[i] = 0
	}

	// This should keep buf's backing array live and thus prevent dead store
	// elimination, according to discussion at
	// https://github.com/golang/go/issues/33325 .
	runtime.KeepAlive(buf)
}

// Copy is identical to Go's builtin copy function except the copying is done in constant time. This is to mitigate against side-channel attacks.
func Copy(dst, src []byte) {
	if len(dst) > len(src) {
		subtle.ConstantTimeCopy(1, dst[:len(src)], src)
	} else if len(dst) < len(src) {
		subtle.ConstantTimeCopy(1, dst, src[:len(dst)])
	} else {
		subtle.ConstantTimeCopy(1, dst, src)
	}
}

// Move is identical to Copy except it wipes the source buffer after the copy operation is executed.
func Move(dst, src []byte) {
	Copy(dst, src)
	Wipe(src)
}

// Equal does a constant-time comparison of two byte slices. This is to mitigate against side-channel attacks.
func Equal(x, y []byte) bool {
	return subtle.ConstantTimeCompare(x, y) == 1
}
