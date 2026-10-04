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

type EnvSecretHandler struct {
	applicationService ports.EnvironmentService
	secretService      ports.EnvSecretService
}

func NewEnvSecretHandler(applicationService ports.EnvironmentService, secretService ports.EnvSecretService) *EnvSecretHandler {
	return &EnvSecretHandler{applicationService: applicationService, secretService: secretService}
}

func (h *EnvSecretHandler) resolvers(c *gin.Context) (orgID bson.ObjectID, envID bson.ObjectID, ok bool) {
	orgIDHex := c.GetString(middleware.CtxKeyOrgID)
	appSlug := c.Param("appSlug")
	envSlug := c.Param("envSlug")

	objOrgID, err := bson.ObjectIDFromHex(orgIDHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization"})
		return bson.ObjectID{}, bson.ObjectID{}, false
	}

	app, err := h.applicationService.GetApplicationBySlug(c.Request.Context(), objOrgID, appSlug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return bson.ObjectID{}, bson.ObjectID{}, false
	}

	env, err := h.applicationService.GetEnvironmentBySlug(c.Request.Context(), objOrgID, app.ID, envSlug)
	if err != nil {
		if errors.Is(err, services.ErrEnvironmentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return bson.ObjectID{}, bson.ObjectID{}, false
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return bson.ObjectID{}, bson.ObjectID{}, false
	}

	return objOrgID, env.ID, true
}

func (h *EnvSecretHandler) List(c *gin.Context) {
	userID := c.GetString(middleware.CtxKeyUserID)

	objUserID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	_, envID, ok := h.resolvers(c)
	if !ok {
		return
	}

	secrets, err := h.secretService.ListSecrets(c.Request.Context(), objUserID, envID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"secrets": secrets})
}

func (h *EnvSecretHandler) Upsert(c *gin.Context) {
	userID := c.GetString(middleware.CtxKeyUserID)
	keyParam := c.Param("key")

	objUserID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	orgID, envID, ok := h.resolvers(c)
	if !ok {
		return
	}

	var req dto.SecretInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if keyParam != "" {
		req.Key = keyParam
	}

	secret, err := h.secretService.UpsertSecret(c.Request.Context(), objUserID, orgID, envID, req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEntityHasNoRoles), errors.Is(err, services.ErrEntityHasInvalidRoles):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrEnvironmentNotFound), errors.Is(err, services.ErrEnvironmentKeyNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrInvalidSecretKind):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"secret":  secret,
		"message": "secret saved successfully",
	})
}

func (h *EnvSecretHandler) Get(c *gin.Context) {
	userID := c.GetString(middleware.CtxKeyUserID)
	key := c.Param("key")

	objUserID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	_, envID, ok := h.resolvers(c)
	if !ok {
		return
	}

	secret, err := h.secretService.GetSecret(c.Request.Context(), objUserID, envID, key)
	if err != nil {
		if errors.Is(err, services.ErrSecretNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"secret": secret})
}

func (h *EnvSecretHandler) Delete(c *gin.Context) {
	userID := c.GetString(middleware.CtxKeyUserID)
	key := c.Param("key")

	objUserID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	orgID, envID, ok := h.resolvers(c)
	if !ok {
		return
	}

	if err := h.secretService.DeleteSecret(c.Request.Context(), objUserID, orgID, envID, key); err != nil {
		switch {
		case errors.Is(err, services.ErrEntityHasNoRoles), errors.Is(err, services.ErrEntityHasInvalidRoles):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrSecretNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "secret deleted successfully"})
}