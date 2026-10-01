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

type ServiceAccountHandler struct {
	applicationService ports.EnvironmentService
	accountService     ports.ServiceAccountService
}

func NewServiceAccountHandler(applicationService ports.EnvironmentService, accountService ports.ServiceAccountService) *ServiceAccountHandler {
	return &ServiceAccountHandler{applicationService: applicationService, accountService: accountService}
}

func (h *ServiceAccountHandler) Create(c *gin.Context) {
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

	var req dto.CreateServiceAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := h.accountService.CreateServiceAccount(c.Request.Context(), objUserID, objOrgID, req.Name, req.Description)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEntityHasNoRoles), errors.Is(err, services.ErrEntityHasInvalidRoles):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"serviceAccount": account,
		"message":        "service account created successfully",
	})
}

func (h *ServiceAccountHandler) List(c *gin.Context) {
	orgID := c.GetString(middleware.CtxKeyOrgID)

	objOrgID, err := bson.ObjectIDFromHex(orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization"})
		return
	}

	accounts, err := h.accountService.ListServiceAccounts(c.Request.Context(), objOrgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"serviceAccounts": accounts})
}

func (h *ServiceAccountHandler) Delete(c *gin.Context) {
	userID := c.GetString(middleware.CtxKeyUserID)
	orgID := c.GetString(middleware.CtxKeyOrgID)
	accountID := c.Param("accountID")

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
	objAccountID, err := bson.ObjectIDFromHex(accountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid service account id"})
		return
	}

	if err := h.accountService.DeleteServiceAccount(c.Request.Context(), objUserID, objOrgID, objAccountID); err != nil {
		switch {
		case errors.Is(err, services.ErrEntityHasNoRoles), errors.Is(err, services.ErrEntityHasInvalidRoles):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrServiceAccountNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "service account deleted successfully"})
}

func (h *ServiceAccountHandler) MintApiKey(c *gin.Context) {
	userID := c.GetString(middleware.CtxKeyUserID)
	orgID := c.GetString(middleware.CtxKeyOrgID)
	accountID := c.Param("accountID")

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
	objAccountID, err := bson.ObjectIDFromHex(accountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid service account id"})
		return
	}

	var req dto.MintApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	app, err := h.applicationService.GetApplicationBySlug(c.Request.Context(), objOrgID, req.ApplicationSlug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	env, err := h.applicationService.GetEnvironmentBySlug(c.Request.Context(), objOrgID, app.ID, req.EnvironmentSlug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	scopes := make([]domain.ApiKeyScope, 0, len(req.Scopes))
	for _, scope := range req.Scopes {
		scopes = append(scopes, domain.ApiKeyScope(scope))
	}

	key, raw, err := h.accountService.MintApiKey(c.Request.Context(), objUserID, objOrgID, objAccountID, env.ID, scopes, req.ExpiresAt)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEntityHasNoRoles), errors.Is(err, services.ErrEntityHasInvalidRoles):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrServiceAccountNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrEnvironmentNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrApiKeyScopesInvalid):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"apiKey":  raw,
		"key":     key,
		"message": "api key created successfully. save it now; it will not be shown again",
	})
}

func (h *ServiceAccountHandler) ListApiKeys(c *gin.Context) {
	orgID := c.GetString(middleware.CtxKeyOrgID)
	accountID := c.Param("accountID")

	objOrgID, err := bson.ObjectIDFromHex(orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization"})
		return
	}
	objAccountID, err := bson.ObjectIDFromHex(accountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid service account id"})
		return
	}

	keys, err := h.accountService.ListApiKeys(c.Request.Context(), objOrgID, objAccountID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrServiceAccountNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"apiKeys": keys})
}

func (h *ServiceAccountHandler) RevokeApiKey(c *gin.Context) {
	userID := c.GetString(middleware.CtxKeyUserID)
	orgID := c.GetString(middleware.CtxKeyOrgID)
	keyID := c.Param("keyID")

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
	objKeyID, err := bson.ObjectIDFromHex(keyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid api key id"})
		return
	}

	if err := h.accountService.RevokeApiKey(c.Request.Context(), objUserID, objOrgID, objKeyID); err != nil {
		switch {
		case errors.Is(err, services.ErrEntityHasNoRoles), errors.Is(err, services.ErrEntityHasInvalidRoles):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrApiKeyNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "api key revoked successfully"})
}