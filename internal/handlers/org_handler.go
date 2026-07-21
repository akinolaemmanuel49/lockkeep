package handlers

import (
	"errors"
	"net/http"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/middleware"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type OrganizationHandler struct {
	orgService ports.OrganizationService
}

func NewOrganizationHandler(orgService ports.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{orgService: orgService}
}

type CreateOrgRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}

func (h *OrganizationHandler) Create(c *gin.Context) {
	userID := c.GetString(middleware.CtxKeyUserID)

	objID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	var req CreateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	org, err := h.orgService.Create(c.Request.Context(), objID, req.Name, req.Slug)
	if err != nil {
		if errors.Is(err, services.ErrOrgSlugTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"organization": org,
		"message":      "organization created successfully",
	})
}

func (h *OrganizationHandler) List(c *gin.Context) {
	userID := c.GetString(middleware.CtxKeyUserID)

	objID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	orgs, err := h.orgService.ListForUser(c.Request.Context(), objID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"organizations": orgs})
}

func (h *OrganizationHandler) Get(c *gin.Context) {
	slug := c.Param("orgSlug")

	org, err := h.orgService.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		if errors.Is(err, services.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"organization": org})
}

func (h *OrganizationHandler) Delete(c *gin.Context) {
	userID := c.GetString(middleware.CtxKeyUserID)
	slug := c.Param("orgSlug")

	objID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	if err := h.orgService.Delete(c.Request.Context(), objID, slug); err != nil {
		switch {
		case errors.Is(err, services.ErrOrganizationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrNotOrgOwner):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "organization deleted successfully"})
}
