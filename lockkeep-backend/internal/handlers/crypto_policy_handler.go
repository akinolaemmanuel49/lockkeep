package handlers

import (
	"net/http"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/middleware"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/services"
	"github.com/akinolaemmanuel49/lockkeep-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
)

type CryptoPolicyHandler struct {
	policyService ports.CryptoPolicyService
}

func NewCryptoPolicyHandler(policyService ports.CryptoPolicyService) *CryptoPolicyHandler {
	return &CryptoPolicyHandler{policyService: policyService}
}

// GetCurrentPolicy returns the system-wide crypto policy used by clients for
// KDF derivation (and upgrade detection). Requires authentication.
func (h *CryptoPolicyHandler) GetCurrentPolicy(c *gin.Context) {
	policy, err := h.policyService.GetCurrentPolicy(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if policy == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no crypto policy configured"})
		return
	}

	c.JSON(http.StatusOK, policy)
}

// SetCurrentPolicy bumps the system-wide crypto policy to a new version. Only
// users carrying the system admin role in their JWT claims may write a policy.
func (h *CryptoPolicyHandler) SetCurrentPolicy(c *gin.Context) {
	claims, exists := c.Get(middleware.CtxKeyClaims)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	jc, ok := claims.(*jwt.Claims)
	if !ok || jc.SystemRole != string(domain.RoleSystemAdmin) {
		c.JSON(http.StatusForbidden, gin.H{"error": "system administrator privileges required"})
		return
	}

	var policy dto.SetCurrentPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.policyService.SetCurrentPolicy(c.Request.Context(), &policy); err != nil {
		if err == services.ErrNothingToUpdate {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "crypto policy updated"})
}