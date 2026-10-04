package handlers

import (
	"errors"
	"net/http"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/middleware"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type VaultHandler struct {
	vaultService     ports.VaultService
	vaultItemService ports.VaultItemService
}

func NewVaultHandler(vaultService ports.VaultService, vaultItemService ports.VaultItemService) *VaultHandler {
	return &VaultHandler{
		vaultService:     vaultService,
		vaultItemService: vaultItemService,
	}
}

func (h *VaultHandler) VerifyVaultPassword(c *gin.Context) {
	userID, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.VerifyVaultPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	success, err := h.vaultService.VerifyVaultPassword(c.Request.Context(), userID.(string), req.VerificationHash)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if !success {
		c.JSON(http.StatusOK, gin.H{"success": false, "credentials": []domain.VaultItem{}})
		return
	}

	objID, ok := h.userObjectID(c)
	if !ok {
		return
	}

	items, err := h.vaultItemService.ListVaultItems(c.Request.Context(), objID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "credentials": items})
}

func (h *VaultHandler) GetVault(c *gin.Context) {
	userID, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	vault, err := h.vaultService.GetVaultByUserID(c.Request.Context(), userID.(string))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrVaultNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"vault":   vault,
		"message": "vault retrieved successfully",
	})
}

func (h *VaultHandler) UpdateVault(c *gin.Context) {
	userID, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.UpdateVaultProfileRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vault, err := h.vaultService.GetVaultByUserID(c.Request.Context(), userID.(string))
	if err != nil {
		if errors.Is(err, services.ErrVaultNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.vaultService.Update(c.Request.Context(), vault.ID.Hex(), req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrVaultNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrInvalidProfileUpdate):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"vault":   updated,
		"message": "vault updated successfully",
	})
}

func (h *VaultHandler) userObjectID(c *gin.Context) (bson.ObjectID, bool) {
	userID, exists := c.Get(middleware.CtxKeyUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return bson.ObjectID{}, false
	}

	objID, err := bson.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return bson.ObjectID{}, false
	}
	return objID, true
}

func (h *VaultHandler) ListVaultItems(c *gin.Context) {
	objID, ok := h.userObjectID(c)
	if !ok {
		return
	}

	items, err := h.vaultItemService.ListVaultItems(c.Request.Context(), objID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *VaultHandler) CreateVaultItem(c *gin.Context) {
	objID, ok := h.userObjectID(c)
	if !ok {
		return
	}

	var req dto.CreateVaultItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item := &domain.VaultItem{
		Type:     domain.VaultItemType(req.Type),
		Name:     req.Name,
		Metadata: req.Metadata,
		Secret: domain.Secret{
			Ciphertext: req.Secret.Ciphertext,
			IV:         req.Secret.IV,
			Tag:        req.Secret.Tag,
			Version:    req.Secret.Version,
		},
	}

	created, err := h.vaultItemService.CreateVaultItem(c.Request.Context(), objID, item)
	if err != nil {
		if errors.Is(err, services.ErrDuplicateVaultItem) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"item":    created,
		"message": "vault item created successfully",
	})
}

func (h *VaultHandler) GetVaultItem(c *gin.Context) {
	objID, ok := h.userObjectID(c)
	if !ok {
		return
	}

	itemID, err := bson.ObjectIDFromHex(c.Param("itemID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	item, err := h.vaultItemService.GetVaultItemByID(c.Request.Context(), objID, itemID)
	if err != nil {
		if errors.Is(err, services.ErrVaultItemNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *VaultHandler) UpdateVaultItem(c *gin.Context) {
	objID, ok := h.userObjectID(c)
	if !ok {
		return
	}

	itemID, err := bson.ObjectIDFromHex(c.Param("itemID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	var req dto.VaultItemUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.vaultItemService.UpdateVaultItem(c.Request.Context(), objID, itemID, req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrVaultItemNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrDuplicateVaultItem):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"item":    updated,
		"message": "vault item updated successfully",
	})
}

func (h *VaultHandler) DeleteVaultItem(c *gin.Context) {
	objID, ok := h.userObjectID(c)
	if !ok {
		return
	}

	itemID, err := bson.ObjectIDFromHex(c.Param("itemID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	if err := h.vaultItemService.DeleteVaultItems(c.Request.Context(), objID, []bson.ObjectID{itemID}); err != nil {
		if errors.Is(err, services.ErrVaultItemNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "vault item deleted successfully"})
}
