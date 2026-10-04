package crypto

import (
	"context"
	"fmt"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	// ScopeEnvironment identifies an EncryptedKey that wraps an environment DEK.
	ScopeEnvironment = "environment"
	// ExportInfo is the HKDF info tag for re-wrapping a DEK to a machine client.
	ExportInfo = "lockkeep:client-dek-wrap"
)

// KeyProvider is the key-management abstraction. Each method mirrors a cloud
// KMS operation (GenerateDataKey / Encrypt / Decrypt) so the rest of the system
// is agnostic to whether keys are managed locally, by AWS KMS, HashiCorp Vault,
// or GCP KMS.
type KeyProvider interface {
	// GenerateDataKey returns a fresh 32-byte DEK together with its wrapped form.
	GenerateDataKey(ctx context.Context, masterKeyID, scope string, scopeID bson.ObjectID) (*domain.EncryptedKey, []byte, error)
	// WrapDataKey wraps an existing 32-byte DEK for storage.
	WrapDataKey(ctx context.Context, masterKeyID, scope string, scopeID bson.ObjectID, plaintext []byte) (*domain.EncryptedKey, error)
	// UnwrapDataKey returns the plaintext DEK from a wrapped EncryptedKey.
	UnwrapDataKey(ctx context.Context, ek *domain.EncryptedKey) ([]byte, error)
}

// EnvWrapAAD binds a wrapped DEK to its environment.
func EnvWrapAAD(envID bson.ObjectID) []byte {
	return []byte(envID.Hex())
}

// EnvSecretAAD binds a secret's ciphertext to its environment, key and version
// so ciphertexts cannot be replayed across rows.
func EnvSecretAAD(envID bson.ObjectID, key string, version int) []byte {
	return []byte(fmt.Sprintf("%s|%s|%d", envID.Hex(), key, version))
}