package mocks

import (
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func MockAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(middleware.CtxKeyUserID, SampleObjectIDHex)
		c.Set(middleware.CtxKeyClaims, nil)
		c.Next()
	}
}
