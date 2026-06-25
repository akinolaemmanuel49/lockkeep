package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID               bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Username         string        `bson:"username" json:"username"`
	Email            string        `bson:"email" json:"email"`
	TenantID         string        `bson:"tenant_id" json:"tenantId"`
	AuthMethod       string        `bson:"auth_method" json:"authMethod"`    // "oauth_google", "oauth_github", "local"
	AuthProviderID   *string       `bson:"auth_provider_id" json:"-"`        // OAuth sub ID
	PasswordHash     string        `bson:"password_hash,omitempty" json:"-"` // For local auth only
	VerificationHash string        `bson:"verification_hash,omitempty" json:"-"`
	KDFParams        KDFParams     `bson:"kdf_params,omitempty" json:"-"`
	CreatedAt        time.Time     `bson:"created_at" json:"createdAt"`
	UpdatedAt        time.Time     `bson:"updated_at" json:"updatedAt"`
}

type KDFParams struct {
	Algorithm   string `bson:"algorithm" json:"algorithm"`
	Salt        string `bson:"salt" json:"salt"`
	Memory      uint32 `bson:"memory" json:"memory"`
	Iterations  uint32 `bson:"iterations" json:"iterations"`
	Parallelism uint8  `bson:"parallelism" json:"parallelism"`
}
