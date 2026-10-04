package mocks

import (
	"time"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// MockRequireOrgAccess returns a middleware that simulates successful org access
func MockRequireOrgAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set org context
		c.Set(middleware.CtxKeyOrgID, SampleOrgIDHex)

		// Set sample membership
		membership := &domain.Membership{
			ID:             bson.NewObjectID(),
			UserID:         mustObjectID(SampleObjectIDHex),
			OrganizationID: mustObjectID(SampleOrgIDHex),
			RoleID:         domain.RoleOrgOwner,
			JoinedAt:       time.Now(),
		}
		c.Set(middleware.CtxKeyMembership, membership)

		c.Next()
	}
}

// Helper to convert hex to ObjectID safely for mocks
func mustObjectID(hex string) bson.ObjectID {
	id, _ := bson.ObjectIDFromHex(hex)
	return id
}
