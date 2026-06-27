package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type VaultItemType string

const (
	ItemLogin       VaultItemType = "login"
	ItemEnvironment VaultItemType = "environment"
	ItemSSHKey      VaultItemType = "ssh_key"
	ItemSecureNote  VaultItemType = "secure_note"
	ItemCard        VaultItemType = "payment_card"
	ItemAPIKey      VaultItemType = "api_key"
)

type VaultItem struct {
	ID       bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID   bson.ObjectID `bson:"user_id" json:"userId"`
	TenantID string        `bson:"tenant_id" json:"tenantId"`

	Type     VaultItemType `bson:"type" json:"type"`
	Name     string        `bson:"name" json:"name"`
	Metadata bson.M        `bson:"metadata,omitempty" json:"metadata,omitempty"`
	Secret   Secret        `bson:"secret" json:"secret"`

	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time `bson:"updated_at" json:"updatedAt"`
}

type Secret struct {
	Ciphertext string `bson:"ciphertext" json:"ciphertext"`
	IV         string `bson:"iv" json:"iv"`
	Tag        string `bson:"tag" json:"tag"`
	Version    uint32 `bson:"version,omitempty" json:"version,omitempty"`
}
