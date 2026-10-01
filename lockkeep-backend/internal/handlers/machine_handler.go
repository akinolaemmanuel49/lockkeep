package handlers

import (
	"errors"
	"net/http"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/middleware"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MachineHandler struct {
	secretService ports.EnvSecretService
}

func NewMachineHandler(secretService ports.EnvSecretService) *MachineHandler {
	return &MachineHandler{secretService: secretService}
}

func hasScope(scopes []domain.ApiKeyScope, want domain.ApiKeyScope) bool {
	for _, scope := range scopes {
		if scope == want {
			return true
		}
	}
	return false
}

// Export serves the re-wrapped envelope to an authenticated API key client.
func (h *MachineHandler) Export(c *gin.Context) {
	scopes, _ := c.Get(middleware.CtxKeyMachineScopes)
	keyScopes, ok := scopes.([]domain.ApiKeyScope)
	if !ok || !hasScope(keyScopes, domain.ApiKeyScopeRead) {
		c.JSON(http.StatusForbidden, gin.H{"error": "api key requires the secrets:read scope"})
		return
	}

	envIDHex := c.GetString(middleware.CtxKeyMachineEnvID)
	envID, err := bson.ObjectIDFromHex(envIDHex)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key scope"})
		return
	}

	saIDHex := c.GetString(middleware.CtxKeyMachineSAID)
	saID, err := bson.ObjectIDFromHex(saIDHex)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key scope"})
		return
	}

	bearerSecret, _ := c.Get(middleware.CtxKeyBearerSecret)
	secret, ok := bearerSecret.([]byte)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key scope"})
		return
	}

	envelope, err := h.secretService.ExportEnvelope(c.Request.Context(), saID, envID, secret)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEnvironmentNotFound), errors.Is(err, services.ErrEnvironmentKeyNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, envelope)
}