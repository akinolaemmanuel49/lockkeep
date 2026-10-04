package mocks

import "go.mongodb.org/mongo-driver/v2/bson"

// Primary sample identifiers for consistent test data
const (
	SampleObjectIDHex     = "507f1f77bcf86cd799439011"
	SampleUserIDHex       = "507f1f77bcf86cd799439012"
	SampleOrgIDHex        = "507f1f77bcf86cd799439013"
	SampleTeamIDHex       = "507f1f77bcf86cd799439014"
	SampleIdentityIDHex   = "507f1f77bcf86cd799439015"
	SampleMembershipIDHex = "507f1f77bcf86cd799439016"
)

// MustObjectID panics on invalid hex — use only in test setup
func MustObjectID(hex string) bson.ObjectID {
	id, err := bson.ObjectIDFromHex(hex)
	if err != nil {
		panic(err)
	}
	return id
}
