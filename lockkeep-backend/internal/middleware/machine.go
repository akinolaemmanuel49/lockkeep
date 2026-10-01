package middleware

import (
	"net/http"
	"strings"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/services"
	"github.com/gin-gonic/gin"
)

const (
	// Machine context keys set by MachineAuth.
	CtxKeyMachineEnvID  = "machineEnvID"
	CtxKeyMachineSAID   = "machineSAID"
	CtxKeyMachineScopes = "machineScopes"
	CtxKeyBearerSecret  = "machineBearerSecret"
)

// MachineAuth authenticates a Bearer API key issued for an environment and
// resolves the environment it is scoped to. It is wired to a top-level route
// group (outside requireOrgAccess) because the key itself carries the scope.
func MachineAuth(serviceAccountService *services.ServiceAccountService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		apikey, _, err := serviceAccountService.VerifyApiKey(c.Request.Context(), strings.TrimSpace(parts[1]))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			return
		}
		if len(apikey.Scopes) == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "api key has no scopes"})
			return
		}

		c.Set(CtxKeyMachineEnvID, apikey.EnvironmentID.Hex())
		c.Set(CtxKeyMachineSAID, apikey.ServiceAccountID.Hex())
		c.Set(CtxKeyMachineScopes, apikey.Scopes)
		c.Set(CtxKeyBearerSecret, []byte(strings.TrimSpace(parts[1])))
		c.Next()
	}
}