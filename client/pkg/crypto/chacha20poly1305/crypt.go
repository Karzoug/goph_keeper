package chacha20poly1305

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/crypto/chacha20poly1305"

	"github.com/Karzoug/goph_keeper/pkg/e"
)

const chunkSize = 1024 * 32 // chunkSize in bytes

func GetCapacityForEncryptedValue(len int) int {
	return (chacha20poly1305.NonceSizeX+chunkSize+chacha20poly1305.Overhead)*(len/chunkSize) +
		chacha20poly1305.NonceSizeX + (len % chunkSize) + chacha20poly1305.Overhead
}

func Encrypt(r io.Reader, w io.Writer, encrKey []byte) error {
	aead, err := chacha20poly1305.NewX(encrKey)
	if err != nil {
		return e.Wrap("creating cipher", err)
	}

	buf := make([]byte, chunkSize)
	var adCounter uint32 = 0
	adCounterBytes := make([]byte, 4)

	for {
		n, err := r.Read(buf)

		if n > 0 {
			nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+n+aead.Overhead())
			if m, err := rand.Read(nonce[0:aead.NonceSize()]); err != nil || m != aead.NonceSize() {
				return errors.New("generated ramdom nonce has wrong size")
			}

			binary.LittleEndian.PutUint32(adCounterBytes, adCounter)
			msg := buf[:n]
			encryptedMsg := aead.Seal(nonce, nonce, msg, adCounterBytes)
			w.Write(encryptedMsg)
			adCounter += 1
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return e.Wrap("error reading: ", err)
		}
	}
	return nil
}

func Decrypt(r io.Reader, w io.Writer, encrKey []byte) error {
	aead, err := chacha20poly1305.NewX(encrKey)
	if err != nil {
		return e.Wrap("creating cipher", err)
	}
	decbufsize := aead.NonceSize() + chunkSize + aead.Overhead()

	buf := make([]byte, decbufsize)
	var adCounter uint32 = 0
	adCounterBytes := make([]byte, 4)

	for {
		n, err := r.Read(buf)
		if n > 0 {
			encryptedMsg := buf[:n]
			if len(encryptedMsg) < aead.NonceSize() {
				return errors.New("ciphertext too short")
			}
			nonce, ciphertext := encryptedMsg[:aead.NonceSize()], encryptedMsg[aead.NonceSize():]
			binary.LittleEndian.PutUint32(adCounterBytes, adCounter)
			plaintext, err := aead.Open(nil, nonce, ciphertext, adCounterBytes)
			if err != nil {
				return errors.New("decrypt ciphertext error: wrong password or data")
			}
			w.Write(plaintext)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return e.Wrap("error reading: ", err)
		}
		adCounter += 1
	}
	return nil
}
