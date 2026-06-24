package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Credential struct {
	ID                bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID            bson.ObjectID `bson:"user_id" json:"userId"`
	TenantID          string        `bson:"tenant_id" json:"tenantId"`
	Organization      string        `bson:"organization" json:"organization"`
	SiteURL           string        `bson:"site_url" json:"siteUrl"`
	Identifier        string        `bson:"identifier" json:"identifier"`
	Notes             string        `bson:"notes" json:"notes"`
	EncryptedPassword string        `bson:"encrypted_password" json:"encryptedPassword"`
	IV                string        `bson:"iv" json:"iv"`
	Tag               string        `bson:"tag" json:"tag"`
	CreatedAt         time.Time     `bson:"created_at" json:"createdAt"`
	UpdatedAt         time.Time     `bson:"updated_at" json:"updatedAt"`
}
