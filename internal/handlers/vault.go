package handlers

import (
	"crypto/subtle"
	"net/http"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/config"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/repository"
	"github.com/akinolaemmanuel49/lockkeep-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type VaultHandler struct {
	cfg        *config.Config
	userRepo   *repository.UserRepository
	vaultRepo  *repository.VaultRepository
	jwtManager *jwt.Manager
}

func NewVaultHandler(cfg *config.Config, userRepo *repository.UserRepository, vaultRepo *repository.VaultRepository, jwtManager *jwt.Manager) *VaultHandler {
	return &VaultHandler{
		cfg:        cfg,
		userRepo:   userRepo,
		vaultRepo:  vaultRepo,
		jwtManager: jwtManager,
	}
}

type VerifyVaultPasswordRequest struct {
	VerificationHash string `json:"verification_hash" binding:"required"`
}

func (h *VaultHandler) VerifyVaultPassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	objectID, err := bson.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	var req VerifyVaultPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	user, err := h.userRepo.FindByID(ctx, objectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if user.VerificationHash == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "vault password not set"})
		return
	}

	// Constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(user.VerificationHash), []byte(req.VerificationHash)) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid master password"})
		return
	}

	// Fetch credentials for this user
	credentials, err := h.vaultRepo.FindByTenant(ctx, user.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch credentials"})
		return
	}

	if credentials == nil {
		credentials = []domain.Credential{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"credentials": credentials,
	})
}

type CreateCredentialRequest struct {
	Organization      string `bson:"organization" json:"organization"`
	SiteURL           string `bson:"site_url" json:"siteUrl"`
	Identifier        string `bson:"identifier" json:"identifier"`
	Notes             string `bson:"notes" json:"notes"`
	EncryptedPassword string `bson:"encrypted_password" json:"encryptedPassword"`
	IV                string `bson:"iv" json:"iv"`
	Tag               string `bson:"tag" json:"tag"`
}

func (h *VaultHandler) CreateCredential(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	objectID, err := bson.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	var req CreateCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	user, err := h.userRepo.FindByID(ctx, objectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	credential := &domain.Credential{
		ID:                bson.NewObjectID(),
		UserID:            user.ID,
		TenantID:          user.TenantID,
		Organization:      req.Organization,
		SiteURL:           req.SiteURL,
		Identifier:        req.Identifier,
		Notes:             req.Notes,
		EncryptedPassword: req.EncryptedPassword,
		IV:                req.IV,
		Tag:               req.Tag,
	}

	err = h.vaultRepo.Create(ctx, credential)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create credential"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"credential": credential,
	})
}
