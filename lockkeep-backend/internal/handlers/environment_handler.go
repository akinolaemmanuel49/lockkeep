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

type EnvironmentHandler struct {
	applicationService ports.EnvironmentService
}

func NewEnvironmentHandler(applicationService ports.EnvironmentService) *EnvironmentHandler {
	return &EnvironmentHandler{applicationService: applicationService}
}

func (h *EnvironmentHandler) Create(c *gin.Context) {
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

	var req dto.CreateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	env, err := h.applicationService.CreateEnvironment(c.Request.Context(), objUserID, objOrgID, app.ID, req.Name, req.Slug, req.IsProtected)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEntityHasNoRoles), errors.Is(err, services.ErrEntityHasInvalidRoles):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrEnvironmentSlugTaken):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrApplicationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"environment": env,
		"message":     "environment created successfully",
	})
}

func (h *EnvironmentHandler) List(c *gin.Context) {
	orgID := c.GetString(middleware.CtxKeyOrgID)
	appSlug := c.Param("appSlug")

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

	envs, err := h.applicationService.ListEnvironments(c.Request.Context(), app.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"environments": envs})
}

func (h *EnvironmentHandler) Get(c *gin.Context) {
	orgID := c.GetString(middleware.CtxKeyOrgID)
	appSlug := c.Param("appSlug")
	envSlug := c.Param("envSlug")

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

	env, err := h.applicationService.GetEnvironmentBySlug(c.Request.Context(), objOrgID, app.ID, envSlug)
	if err != nil {
		if errors.Is(err, services.ErrEnvironmentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"environment": env})
}

func (h *EnvironmentHandler) Update(c *gin.Context) {
	userID := c.GetString(middleware.CtxKeyUserID)
	orgID := c.GetString(middleware.CtxKeyOrgID)
	appSlug := c.Param("appSlug")
	envSlug := c.Param("envSlug")

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

	env, err := h.applicationService.GetEnvironmentBySlug(c.Request.Context(), objOrgID, app.ID, envSlug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var req dto.UpdateEnvironmentDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.applicationService.UpdateEnvironment(c.Request.Context(), objUserID, objOrgID, env.ID, req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEntityHasNoRoles), errors.Is(err, services.ErrEntityHasInvalidRoles):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrEnvironmentNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrEnvironmentSlugTaken):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"environment": updated,
		"message":     "environment updated successfully",
	})
}

func (h *EnvironmentHandler) Delete(c *gin.Context) {
	userID := c.GetString(middleware.CtxKeyUserID)
	orgID := c.GetString(middleware.CtxKeyOrgID)
	appSlug := c.Param("appSlug")
	envSlug := c.Param("envSlug")

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

	env, err := h.applicationService.GetEnvironmentBySlug(c.Request.Context(), objOrgID, app.ID, envSlug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if err := h.applicationService.DeleteEnvironment(c.Request.Context(), objUserID, objOrgID, env.ID); err != nil {
		switch {
		case errors.Is(err, services.ErrEntityHasNoRoles), errors.Is(err, services.ErrEntityHasInvalidRoles):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrEnvironmentNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "environment deleted successfully"})
}