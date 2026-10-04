package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/hkdf"
)

// DeriveKey expands material into a fixed-length key with HKDF-SHA256.
func DeriveKey(master, salt []byte, info string, length int) ([]byte, error) {
	key := make([]byte, length)
	r := hkdf.New(sha256.New, master, salt, []byte(info))
	if _, err := io.ReadFull(r, key); err != nil {
		return nil, err
	}
	return key, nil
}

// HMACSHA256 returns the HMAC-SHA256 digest of data under secret.
func HMACSHA256(secret, data []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write(data)
	return mac.Sum(nil)
}

// ConstantTimeEqual compares digests in constant time.
func ConstantTimeEqual(a, b []byte) bool {
	return hmac.Equal(a, b)
}