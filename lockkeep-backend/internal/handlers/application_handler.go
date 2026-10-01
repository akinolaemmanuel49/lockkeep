package handlers

import (
	"errors"
	"net/http"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/middleware"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ApplicationHandler struct {
	applicationService ports.EnvironmentService
}

func NewApplicationHandler(applicationService ports.EnvironmentService) *ApplicationHandler {
	return &ApplicationHandler{applicationService: applicationService}
}

func (h *ApplicationHandler) Create(c *gin.Context) {
	userID := c.GetString(middleware.CtxKeyUserID)
	orgID := c.GetString(middleware.CtxKeyOrgID)

	objUserID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}
	objOrgID, err := bson.ObjectIDFromHex(orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization"})
		return
	}

	var req dto.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	app, err := h.applicationService.CreateApplication(c.Request.Context(), objUserID, objOrgID, req.Name, req.Slug, req.Description)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEntityHasNoRoles), errors.Is(err, services.ErrEntityHasInvalidRoles):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrApplicationSlugTaken):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"application": app,
		"message":     "application created successfully",
	})
}

func (h *ApplicationHandler) List(c *gin.Context) {
	orgID := c.GetString(middleware.CtxKeyOrgID)

	objOrgID, err := bson.ObjectIDFromHex(orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization"})
		return
	}

	apps, err := h.applicationService.ListApplications(c.Request.Context(), objOrgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"applications": apps})
}

func (h *ApplicationHandler) Get(c *gin.Context) {
	orgID := c.GetString(middleware.CtxKeyOrgID)
	appSlug := c.Param("appSlug")

	objOrgID, err := bson.ObjectIDFromHex(orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization"})
		return
	}

	app, err := h.applicationService.GetApplicationBySlug(c.Request.Context(), objOrgID, appSlug)
	if err != nil {
		if errors.Is(err, services.ErrApplicationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"application": app})
}

func (h *ApplicationHandler) Update(c *gin.Context) {
	userID := c.GetString(middleware.CtxKeyUserID)
	orgID := c.GetString(middleware.CtxKeyOrgID)
	appSlug := c.Param("appSlug")

	objUserID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}
	objOrgID, err := bson.ObjectIDFromHex(orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization"})
		return
	}

	app, err := h.applicationService.GetApplicationBySlug(c.Request.Context(), objOrgID, appSlug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var req dto.UpdateApplicationDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.applicationService.UpdateApplication(c.Request.Context(), objUserID, objOrgID, app.ID, req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEntityHasNoRoles), errors.Is(err, services.ErrEntityHasInvalidRoles):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrApplicationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrApplicationSlugTaken):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"application": updated,
		"message":     "application updated successfully",
	})
}

func (h *ApplicationHandler) Delete(c *gin.Context) {
	userID := c.GetString(middleware.CtxKeyUserID)
	orgID := c.GetString(middleware.CtxKeyOrgID)
	appSlug := c.Param("appSlug")

	objUserID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}
	objOrgID, err := bson.ObjectIDFromHex(orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization"})
		return
	}

	app, err := h.applicationService.GetApplicationBySlug(c.Request.Context(), objOrgID, appSlug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if err := h.applicationService.DeleteApplication(c.Request.Context(), objUserID, objOrgID, app.ID); err != nil {
		switch {
		case errors.Is(err, services.ErrEntityHasNoRoles), errors.Is(err, services.ErrEntityHasInvalidRoles):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrApplicationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "application deleted successfully"})
}