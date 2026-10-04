package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AuthMethod string

const (
	AuthMethodLocal  AuthMethod = "local"
	AuthMethodGoogle AuthMethod = "oauth_google"
	AuthMethodGitHub AuthMethod = "oauth_github"
	AuthMethodAuth0  AuthMethod = "oauth_auth0"
)

// Identity holds authentication credentials. One user can have multiple identities.
type Identity struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID         bson.ObjectID `bson:"user_id" json:"userId"`
	AuthMethod     AuthMethod    `bson:"auth_method" json:"authMethod"`
	AuthProviderID *string       `bson:"auth_provider_id,omitempty" json:"-"`
	PasswordHash   string        `bson:"password_hash,omitempty" json:"-"`
	LastLoginAt    *time.Time    `bson:"last_login_at,omitempty" json:"lastLoginAt"`
	CreatedAt      time.Time     `bson:"created_at" json:"createdAt"`
}
