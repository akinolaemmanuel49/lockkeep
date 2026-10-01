package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Application struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	OrganizationID bson.ObjectID `bson:"organization_id" json:"organizationId"`
	Name           string        `bson:"name" json:"name"`
	Slug           string        `bson:"slug" json:"slug"`
	Description    *string       `bson:"description,omitempty" json:"description,omitempty"`
	CreatedBy      bson.ObjectID `bson:"created_by" json:"createdBy"`
	CreatedAt      time.Time     `bson:"created_at" json:"createdAt"`
	UpdatedAt      time.Time     `bson:"updated_at" json:"updatedAt"`
}