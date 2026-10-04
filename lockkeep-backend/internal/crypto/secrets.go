package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// ErrCiphertextTampered is returned when a sealed secret fails authentication.
var ErrCiphertextTampered = errors.New("ciphertext failed authentication")

// Seal encrypts plaintext with AES-256-GCM under the given 32-byte DEK,
// binding aad to the ciphertext. Used for secret values and for re-wrapping a
// DEK to a machine client.
func Seal(dek, plaintext, aad []byte) (ciphertext, nonce []byte, err error) {
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	nonce = make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}
	return gcm.Seal(nil, nonce, plaintext, aad), nonce, nil
}

// Open decrypts a sealed secret, authenticating it against aad.
func Open(dek, ciphertext, nonce, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, aad)
	if err != nil {
		return nil, ErrCiphertextTampered
	}
	return plaintext, nil
}