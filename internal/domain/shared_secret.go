package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type SharedSecretScope string

const (
	ScopeOrganization SharedSecretScope = "organization"
	ScopeTeam         SharedSecretScope = "team"
)

type SharedSecret struct {
	ID             bson.ObjectID     `bson:"_id,omitempty" json:"id"`
	OrganizationID bson.ObjectID     `bson:"organization_id" json:"organizationId"`
	TeamID         *bson.ObjectID    `bson:"team_id,omitempty" json:"teamId,omitempty"`
	Scope          SharedSecretScope `bson:"scope" json:"scope"`

	Name        string   `bson:"name" json:"name"`
	Description *string  `bson:"description,omitempty" json:"description,omitempty"`
	Category    string   `bson:"category" json:"category"` // "credential", "api_key", "note", etc.
	Tags        []string `bson:"tags" json:"tags"`

	// Server-managed encryption. Decrypted server-side, sent over TLS.
	EncryptedData []byte `bson:"encrypted_data" json:"-"`
	Nonce         []byte `bson:"nonce" json:"-"`
	KeyVersion    int    `bson:"key_version" json:"-"`

	CreatedBy bson.ObjectID  `bson:"created_by" json:"createdBy"`
	CreatedAt time.Time      `bson:"created_at" json:"createdAt"`
	UpdatedBy *bson.ObjectID `bson:"updated_by,omitempty" json:"updatedBy,omitempty"`
	UpdatedAt time.Time      `bson:"updated_at" json:"updatedAt"`
}
