package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type SecretKind string

const (
	SecretKindEnv  SecretKind = "env"
	SecretKindNote SecretKind = "note"
	SecretKindFile SecretKind = "file"
	SecretKindJSON SecretKind = "json"
)

// EnvSecret — an environment-scoped secret stored under a server-managed DEK.
// The DEK is wrapped server-side (see EncryptedKey) so the server can serve
// human UI reads, and re-wrapped to a caller on machine export.
type EnvSecret struct {
	ID             bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	OrganizationID bson.ObjectID  `bson:"organization_id" json:"organizationId"`
	EnvironmentID  bson.ObjectID  `bson:"environment_id" json:"environmentId"`
	Key            string         `bson:"key" json:"key"`
	Kind           SecretKind     `bson:"kind" json:"kind"`
	Version        int            `bson:"version" json:"version"`
	EncryptedData  []byte         `bson:"encrypted_data" json:"-"`
	Nonce          []byte         `bson:"nonce" json:"-"`
	CreatedBy      bson.ObjectID  `bson:"created_by" json:"createdBy"`
	CreatedAt      time.Time      `bson:"created_at" json:"createdAt"`
	UpdatedBy      *bson.ObjectID `bson:"updated_by,omitempty" json:"updatedBy,omitempty"`
	UpdatedAt      time.Time      `bson:"updated_at" json:"updatedAt"`
}