package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/middleware"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// refreshCookieMaxAge is how long a refresh token cookie stays valid.
const refreshCookieMaxAge = 7 * 24 * 60 * 60

// setRefreshCookie writes the httpOnly refresh-token cookie. SameSite=None +
// Secure are required for the cookie to survive cross-site requests (the web
// console on one origin calling the API on another); browsers reject
// SameSite=None without Secure. The host-only cookie (empty Domain) is fine
// because the API sets and reads its own cookie.
func setRefreshCookie(c *gin.Context, token string, maxAge int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		Expires:  time.Now().Add(time.Duration(maxAge) * time.Second),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})
}

// clearRefreshCookie expires the refresh-token cookie (logout / failed refresh).
func clearRefreshCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})
}

type AuthHandler struct {
	authService ports.AuthService
}

func NewAuthHandler(authService ports.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) OAuth(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	accessToken, _ := strings.CutPrefix(authHeader, "Bearer ")
	accessToken = strings.TrimSpace(accessToken)

	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing access token"})
		return
	}

	user, tokens, isNewUser, err := h.authService.OAuth(c.Request.Context(), accessToken)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrOAuthEmailRequired):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrEmailTaken):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid access token"})
		}
		return
	}

	setRefreshCookie(c, tokens.RefreshToken, refreshCookieMaxAge)

	status := http.StatusOK
	message := "user logged in successfully"
	if isNewUser {
		status = http.StatusCreated
		message = "user created successfully"
	}

	c.JSON(status, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"user":          user,
		"message":       message,
	})
}

func (h *AuthHandler) Register(ctx *gin.Context) {
	var req dto.RegisterRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, tokens, err := h.authService.Register(ctx.Request.Context(), req)
	if err != nil {
		if errors.Is(err, services.ErrEmailTaken) {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	setRefreshCookie(ctx, tokens.RefreshToken, refreshCookieMaxAge)

	ctx.JSON(http.StatusCreated, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"user":          user,
		"message":       "user created successfully",
	})
}

func (h *AuthHandler) Login(ctx *gin.Context) {
	var req dto.LoginRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, tokens, err := h.authService.Login(ctx.Request.Context(), req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	setRefreshCookie(ctx, tokens.RefreshToken, refreshCookieMaxAge)

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"user":          user,
		"message":       "user logged in successfully",
	})
}

func (h *AuthHandler) Refresh(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
		return
	}

	tokens, err := h.authService.Refresh(ctx.Request.Context(), refreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	setRefreshCookie(ctx, tokens.RefreshToken, refreshCookieMaxAge)

	ctx.JSON(http.StatusOK, gin.H{
		"access_token": tokens.AccessToken,
	})
}

func (h *AuthHandler) Logout(ctx *gin.Context) {
	clearRefreshCookie(ctx)
	ctx.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *AuthHandler) SetVerificationHash(ctx *gin.Context) {
	userID := ctx.GetString(middleware.CtxKeyUserID)
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	var req dto.SetVerificationHashRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.SetVerificationHash(ctx.Request.Context(), objectID, req)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *AuthHandler) GetKDFParams(ctx *gin.Context) {
	userID := ctx.GetString(middleware.CtxKeyUserID)
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	kdf, err := h.authService.GetKDFParams(ctx.Request.Context(), objectID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrVaultNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, kdf)
}

func (h *AuthHandler) UpdateEmail(ctx *gin.Context) {
	userID := ctx.GetString(middleware.CtxKeyUserID)
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	var req dto.UpdateEmailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, tokens, err := h.authService.UpdateEmail(ctx.Request.Context(), objectID, req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrCannotEditOAuth):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrEmailInUse):
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	setRefreshCookie(ctx, tokens.RefreshToken, refreshCookieMaxAge)

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"user":          user,
	})
}

func (h *AuthHandler) UpdateAccountPassword(ctx *gin.Context) {
	userID := ctx.GetString(middleware.CtxKeyUserID)
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	var req dto.UpdateAccountPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.UpdateAccountPassword(ctx.Request.Context(), objectID, req); err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrCannotEditOAuth):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrInvalidPassword):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "password successfully updated"})
}
