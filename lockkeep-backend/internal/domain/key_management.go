package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type KeyProvider string

const (
	KeyProviderAWSKMS         KeyProvider = "aws_kms"
	KeyProviderHashiCorpVault KeyProvider = "hashicorp_vault"
	KeyProviderLocal          KeyProvider = "local"
)

// MasterKeyRef — references external KMS, never stores actual key material
type MasterKeyRef struct {
	ID        string      `bson:"_id" json:"id"`
	Provider  KeyProvider `bson:"provider" json:"provider"`
	KeyID     string      `bson:"key_id" json:"keyId"`
	CreatedAt time.Time   `bson:"created_at" json:"createdAt"`
}

// EncryptedKey — actual encryption key, wrapped by master key
type EncryptedKey struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Scope       string        `bson:"scope" json:"scope"` // "org", "team" or "environment"
	ScopeID     bson.ObjectID `bson:"scope_id" json:"scopeId"`
	MasterKeyID string        `bson:"master_key_id" json:"masterKeyId"`
	Ciphertext  []byte        `bson:"ciphertext" json:"-"`
	Nonce       []byte        `bson:"nonce" json:"-"`
	Algorithm   string        `bson:"algorithm" json:"algorithm"`
	CreatedAt   time.Time     `bson:"created_at" json:"createdAt"`
	RotatedAt   *time.Time    `bson:"rotated_at,omitempty" json:"rotatedAt,omitempty"`
}
