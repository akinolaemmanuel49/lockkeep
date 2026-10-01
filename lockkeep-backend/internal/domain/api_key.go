package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ApiKeyScope string

const (
	ApiKeyScopeRead  ApiKeyScope = "secrets:read"
	ApiKeyScopeWrite ApiKeyScope = "secrets:write"
)

// ApiKey — a machine credential scoped to a single environment. Only the HMAC
// hash of the raw key is stored; the raw key is shown once at mint time.
type ApiKey struct {
	ID               bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	ServiceAccountID bson.ObjectID  `bson:"service_account_id" json:"serviceAccountId"`
	EnvironmentID    bson.ObjectID  `bson:"environment_id" json:"environmentId"`
	KeyPrefix        string         `bson:"key_prefix" json:"keyPrefix"`
	KeyHash          []byte         `bson:"key_hash" json:"-"`
	Scopes           []ApiKeyScope  `bson:"scopes" json:"scopes"`
	CreatedBy        bson.ObjectID  `bson:"created_by" json:"createdBy"`
	ExpiresAt        *time.Time     `bson:"expires_at,omitempty" json:"expiresAt,omitempty"`
	LastUsedAt       *time.Time     `bson:"last_used_at,omitempty" json:"lastUsedAt,omitempty"`
	RevokedAt        *time.Time     `bson:"revoked_at,omitempty" json:"revokedAt,omitempty"`
	CreatedAt        time.Time      `bson:"created_at" json:"createdAt"`
	UpdatedAt        time.Time      `bson:"updated_at" json:"updatedAt"`
}