package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/config"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/repository"
	"github.com/akinolaemmanuel49/lockkeep-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	cfg        *config.Config
	userRepo   *repository.UserRepository
	jwtManager *jwt.Manager
}

func NewAuthHandler(cfg *config.Config, userRepo *repository.UserRepository, jwtManager *jwt.Manager) *AuthHandler {
	return &AuthHandler{
		cfg:        cfg,
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// --- OAuth (Auth0) ---

type Auth0UserInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (h *AuthHandler) OAuth(c *gin.Context) {
	accessToken, _ := strings.CutPrefix(c.Request.Header.Get("Authorization"), "Bearer ")

	// Get user info from Auth0
	userInfo, err := h.getAuth0UserInfo(accessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid access token"})
		return
	}

	provider := extractProvider(userInfo.Sub)

	ctx := c.Request.Context()

	// Check if user exists
	user, err := h.userRepo.FindByOAuth(ctx, provider, userInfo.Sub)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if user == nil {
		// Check for email collision across providers
		existing, err := h.userRepo.FindByEmail(ctx, userInfo.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		if existing != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "An account with this email already exists. Please sign in instead."})
			return
		}

		// Create new user
		user = &domain.User{
			Email:          userInfo.Email,
			TenantID:       generateTenantID(),
			AuthMethod:     "oauth_" + provider,
			AuthProviderID: userInfo.Sub,
		}
		if err := h.userRepo.Create(ctx, user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}
	}

	// Generate tokens
	tokens, err := h.jwtManager.Generate(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	// Set refresh token as httpOnly cookie
	c.SetCookie("refresh_token", tokens.RefreshToken, 7*24*60*60, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"access_token": tokens.AccessToken,
		"user": gin.H{
			"id":                user.ID.Hex(),
			"email":             user.Email,
			"tenantId":          user.TenantID,
			"hasMasterPassword": user.MasterPasswordHash != "",
		},
	})
}

func (h *AuthHandler) getAuth0UserInfo(accessToken string) (*Auth0UserInfo, error) {
	req, err := http.NewRequest("GET", "https://"+h.cfg.Auth0Domain+"/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.New("auth0 userinfo failed: " + string(body))
	}

	var userInfo Auth0UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}
	return &userInfo, nil
}

// --- Local Auth ---

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Check for existing email
	existing, err := h.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "An account with this email already exists. Please sign in instead."})
		return
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user := &domain.User{
		Email:        req.Email,
		TenantID:     generateTenantID(),
		AuthMethod:   "local",
		PasswordHash: string(passwordHash),
	}

	if err := h.userRepo.Create(ctx, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	tokens, err := h.jwtManager.Generate(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	c.SetCookie("refresh_token", tokens.RefreshToken, 7*24*60*60, "/", "", false, true)

	c.JSON(http.StatusCreated, gin.H{
		"access_token": tokens.AccessToken,
		"user": gin.H{
			"id":                user.ID.Hex(),
			"email":             user.Email,
			"tenantId":          user.TenantID,
			"hasMasterPassword": false,
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	user, err := h.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if user == nil || user.AuthMethod != "local" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	tokens, err := h.jwtManager.Generate(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	c.SetCookie("refresh_token", tokens.RefreshToken, 7*24*60*60, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"access_token": tokens.AccessToken,
		"user": gin.H{
			"id":                user.ID.Hex(),
			"email":             user.Email,
			"tenantId":          user.TenantID,
			"hasMasterPassword": user.MasterPasswordHash != "",
		},
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
		return
	}

	userID, err := h.jwtManager.ValidateRefresh(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	ctx := c.Request.Context()
	objectID, err := h.jwtManager.ParseUserID(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.userRepo.FindByID(ctx, objectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	tokens, err := h.jwtManager.Generate(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	c.SetCookie("refresh_token", tokens.RefreshToken, 7*24*60*60, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"access_token": tokens.AccessToken,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func generateTenantID() string {
	return "tenant_" + bson.NewObjectID().Hex()[:8]
}

func extractProvider(sub string) string {
	if strings.HasPrefix(sub, "google-oauth2|") {
		return "google"
	}
	if strings.HasPrefix(sub, "github|") {
		return "github"
	}
	if strings.HasPrefix(sub, "auth0|") {
		return "local" // Auth0 database connection
	}
	return "oauth"
}
