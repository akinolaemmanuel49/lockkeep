package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Environment struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	OrganizationID bson.ObjectID `bson:"organization_id" json:"organizationId"`
	ApplicationID  bson.ObjectID `bson:"application_id" json:"applicationId"`
	Name           string        `bson:"name" json:"name"`
	Slug           string        `bson:"slug" json:"slug"`
	IsProtected    bool          `bson:"is_protected" json:"isProtected"`
	CreatedBy      bson.ObjectID `bson:"created_by" json:"createdBy"`
	CreatedAt      time.Time     `bson:"created_at" json:"createdAt"`
	UpdatedAt      time.Time     `bson:"updated_at" json:"updatedAt"`
}