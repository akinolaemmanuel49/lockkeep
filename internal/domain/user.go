package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Username  string        `bson:"username" json:"username"`
	Email     string        `bson:"email" json:"email"`
	AvatarURL *string       `bson:"avatar_url,omitempty" json:"avatarUrl,omitempty"`

	// Personal vault metadata — zero-knowledge, client-derived master key
	Vault *VaultMetadata `bson:"vault,omitempty" json:"vault,omitempty"`

	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time `bson:"updated_at" json:"updatedAt"`
}
