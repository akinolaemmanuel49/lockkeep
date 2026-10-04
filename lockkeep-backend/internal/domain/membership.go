package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Membership struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID         bson.ObjectID `bson:"user_id" json:"userId"`
	OrganizationID bson.ObjectID `bson:"organization_id" json:"organizationId"`
	RoleID         RoleID        `bson:"role_id" json:"roleId"`
	TeamRoles      []TeamRole    `bson:"team_roles" json:"teamRoles"`
	JoinedAt       time.Time     `bson:"joined_at" json:"joinedAt"`
	UpdatedAt      time.Time     `bson:"updated_at" json:"updatedAt"`
}

type TeamRole struct {
	TeamID bson.ObjectID `bson:"team_id" json:"teamId"`
	RoleID RoleID        `bson:"role_id" json:"roleId"`
}
