package handlers

import (
	"net/http"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type CryptoPolicyHandler struct {
	policyService *service.CryptoPolicyService
}

func NewCryptoPolicyHandler(policyService *service.CryptoPolicyService) *CryptoPolicyHandler {
	return &CryptoPolicyHandler{policyService: policyService}
}

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

func (h *CryptoPolicyHandler) SetCurrentPolicy(c *gin.Context) {
	userType, exists := c.Get("userType")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if userType != domain.ADMIN {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid user type"})
		return
	}

	var policy dto.SetCurrentPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.policyService.SetCurrentPolicy(c.Request.Context(), &policy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "crypto policy updated"})
}
