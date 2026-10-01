package crypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"time"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	// LocalKeyWrappingAlgorithm is the wrap algorithm recorded on EncryptedKey.
	LocalKeyWrappingAlgorithm = "aes-256-gcm-v1"
	// LocalMasterKeyID identifies the deployment key-encryption key.
	LocalMasterKeyID = "local"
)

var (
	ErrInvalidKeySize  = errors.New("wrapped key must be 32 bytes")
	ErrTamperedKeyWrap = errors.New("wrapped key failed authentication")
	ErrKeyNotWrapped   = errors.New("cannot derive wrapping key from empty material")
)

// LocalKeyProvider wraps 32-byte DEKs with AES-256-GCM under a deployment KEK
// derived via HKDF from a server-held secret. A future AWS KMS / Vault provider
// swaps this implementation without changing callers.
type LocalKeyProvider struct {
	kek []byte
}

// NewLocalKeyProvider derives the stable KEK from the deployment secret. When
// secret is empty an ephemeral key is generated (wrapped keys become
// unreadable after a restart — dev only).
func NewLocalKeyProvider(secret []byte) (*LocalKeyProvider, error) {
	var source []byte
	if len(secret) == 0 {
		source = make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, source); err != nil {
			return nil, err
		}
	} else {
		source = secret
	}

	kek, err := DeriveKey(source, nil, LocalMasterKeyID, 32)
	if err != nil {
		return nil, err
	}

	return &LocalKeyProvider{kek: kek}, nil
}

func (p *LocalKeyProvider) scopeAAD(scope string, scopeID bson.ObjectID) []byte {
	return []byte(scope + ":" + scopeID.Hex())
}

func (p *LocalKeyProvider) wrap(plaintext, aad []byte) (ciphertext, nonce []byte, err error) {
	block, err := aes.NewCipher(p.kek)
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

func (p *LocalKeyProvider) unwrap(ciphertext, nonce, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(p.kek)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, aad)
	if err != nil {
		return nil, ErrTamperedKeyWrap
	}
	return plaintext, nil
}

func (p *LocalKeyProvider) GenerateDataKey(ctx context.Context, masterKeyID, scope string, scopeID bson.ObjectID) (*domain.EncryptedKey, []byte, error) {
	dek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, nil, err
	}
	ek, err := p.WrapDataKey(ctx, masterKeyID, scope, scopeID, dek)
	if err != nil {
		return nil, nil, err
	}
	return ek, dek, nil
}

func (p *LocalKeyProvider) WrapDataKey(ctx context.Context, masterKeyID, scope string, scopeID bson.ObjectID, plaintext []byte) (*domain.EncryptedKey, error) {
	if len(plaintext) != 32 {
		return nil, ErrInvalidKeySize
	}
	ciphertext, nonce, err := p.wrap(plaintext, p.scopeAAD(scope, scopeID))
	if err != nil {
		return nil, err
	}
	if masterKeyID == "" {
		masterKeyID = LocalMasterKeyID
	}
	return &domain.EncryptedKey{
		ID:          bson.NewObjectID(),
		Scope:       scope,
		ScopeID:     scopeID,
		MasterKeyID: masterKeyID,
		Ciphertext:  ciphertext,
		Nonce:       nonce,
		Algorithm:   LocalKeyWrappingAlgorithm,
		CreatedAt:   time.Now(),
	}, nil
}

func (p *LocalKeyProvider) UnwrapDataKey(ctx context.Context, ek *domain.EncryptedKey) ([]byte, error) {
	return p.unwrap(ek.Ciphertext, ek.Nonce, p.scopeAAD(ek.Scope, ek.ScopeID))
}