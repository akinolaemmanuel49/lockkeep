package handlers

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/config"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/repository"
	"github.com/akinolaemmanuel49/lockkeep-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type VaultHandler struct {
	cfg        *config.Config
	client     *mongo.Client
	userRepo   *repository.UserRepository
	vaultRepo  *repository.VaultRepository
	jwtManager *jwt.Manager
}

func NewVaultHandler(cfg *config.Config, client *mongo.Client, userRepo *repository.UserRepository, vaultRepo *repository.VaultRepository, jwtManager *jwt.Manager) *VaultHandler {
	return &VaultHandler{
		cfg:        cfg,
		client:     client,
		userRepo:   userRepo,
		vaultRepo:  vaultRepo,
		jwtManager: jwtManager,
	}
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

	var req dto.VerifyVaultPasswordRequest
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vault password"})
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

	var req dto.CreateCredentialRequest
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

	c.JSON(http.StatusOK, credential)
}

func (h *VaultHandler) GetCredentials(c *gin.Context) {
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

	credentials, err := h.vaultRepo.FindByTenant(ctx, user.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch credentials"})
		return
	}

	if credentials == nil {
		credentials = []domain.Credential{}
	}

	c.JSON(http.StatusOK, credentials)
}

func (h *VaultHandler) UpdateCredential(c *gin.Context) {
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

	credentialID := c.Param("id")
	credObjectID, err := bson.ObjectIDFromHex(credentialID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential id"})
		return
	}

	var req dto.UpdateCredentialRequest
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

	// Check for duplicates if org/identifier changed
	if req.Organization != "" && req.Identifier != "" {
		isDup, err := h.vaultRepo.FindDuplicate(ctx, user.ID, user.TenantID, req.Organization, req.Identifier, &credObjectID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		if isDup {
			c.JSON(http.StatusConflict, gin.H{"error": "credential with this organization and identifier already exists"})
			return
		}
	}

	updates := bson.M{}
	if req.Organization != "" {
		updates["organization"] = req.Organization
	}
	if req.SiteURL != "" {
		updates["site_url"] = req.SiteURL
	}
	if req.Identifier != "" {
		updates["identifier"] = req.Identifier
	}
	if req.Notes != "" {
		updates["notes"] = req.Notes
	}
	if req.EncryptedPassword != "" {
		updates["encrypted_password"] = req.EncryptedPassword
	}
	if req.IV != "" {
		updates["iv"] = req.IV
	}
	if req.Tag != "" {
		updates["tag"] = req.Tag
	}

	if err := h.vaultRepo.Update(ctx, credObjectID, user.ID, user.TenantID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update credential"})
		return
	}

	// Fetch updated credential
	updated, err := h.vaultRepo.FindByID(ctx, credObjectID, user.ID, user.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated credential"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *VaultHandler) DeleteCredential(c *gin.Context) {
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

	credentialID := c.Param("id")
	credObjectID, err := bson.ObjectIDFromHex(credentialID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential id"})
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

	if err := h.vaultRepo.Delete(ctx, credObjectID, user.ID, user.TenantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete credential"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "credential deleted"})
}

func (h *VaultHandler) MigrateVault(c *gin.Context) {
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

	var req dto.MigrateVaultRequest
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

	KDFParams := domain.KDFParams{
		Algorithm:   req.KDFParams.Algorithm,
		Salt:        req.KDFParams.Salt,
		Memory:      req.KDFParams.Memory,
		Iterations:  req.KDFParams.Iterations,
		Parallelism: req.KDFParams.Parallelism,
	}

	session, err := h.client.StartSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to start session",
		})
		return
	}
	defer session.EndSession(ctx)

	err = session.StartTransaction()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to start transaction",
		})
		return
	}

	err = mongo.WithSession(ctx, session, func(ctx context.Context) error {

		if err := h.userRepo.UpdateKDF(
			ctx,
			objectID,
			req.ExpectedVersion,
			req.VerificationHash,
			KDFParams,
		); err != nil {
			return err
		}

		if err := h.vaultRepo.BulkUpdate(
			ctx,
			user.ID,
			user.TenantID,
			req.VaultUpdates,
		); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, repository.ErrVaultAlreadyMigrated) {
			_ = session.AbortTransaction(ctx)

			c.JSON(http.StatusConflict, gin.H{
				"error": "vault already migrated",
			})
			return
		}

		_ = session.AbortTransaction(ctx)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "migration failed",
		})
		return
	}

	if err := session.CommitTransaction(ctx); err != nil {
		_ = session.AbortTransaction(ctx)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to commit migration",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "vault migrated",
	})
}
