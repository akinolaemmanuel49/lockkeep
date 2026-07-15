package mocks

import (
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

const SampleObjectIDHex = "507f1f77bcf86cd799439011"

func MockAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(middleware.CtxKeyUserID, SampleObjectIDHex)
		c.Next()
	}
}
