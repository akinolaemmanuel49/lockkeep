package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Organization struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string        `bson:"name" json:"name"`
	Slug      string        `bson:"slug" json:"slug"`
	OwnerID   bson.ObjectID `bson:"owner_id" json:"ownerId"`
	Settings  OrgSettings   `bson:"settings" json:"settings"`
	CreatedAt time.Time     `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time     `bson:"updated_at" json:"updatedAt"`
}

type OrgSettings struct {
	Require2FA      bool   `bson:"require_2fa" json:"require2FA"`
	DefaultTeamRole RoleID `bson:"default_team_role" json:"defaultTeamRole"`
}
