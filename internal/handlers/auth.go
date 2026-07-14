package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/middleware"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
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
		case errors.Is(err, service.ErrOAuthEmailRequired):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrEmailTaken):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid access token"})
		}
		return
	}

	if isNewUser {
		c.JSON(http.StatusCreated, gin.H{
			"user":    user,
			"message": "user created successfully",
		})
		return
	}

	c.SetCookie("refresh_token", tokens.RefreshToken, 7*24*60*60, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"user":          user,
		"message":       "user logged in successfully",
	})
}

func (h *AuthHandler) Register(ctx *gin.Context) {
	var req dto.RegisterRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Register(ctx.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"user":    user,
		"message": "user created successfully",
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
		if errors.Is(err, service.ErrInvalidCredentials) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.SetCookie("refresh_token", tokens.RefreshToken, 7*24*60*60, "/", "", false, true)

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

	ctx.SetCookie("refresh_token", tokens.RefreshToken, 7*24*60*60, "/", "", false, true)

	ctx.JSON(http.StatusOK, gin.H{
		"access_token": tokens.AccessToken,
	})
}

func (h *AuthHandler) GetMe(ctx *gin.Context) {
	userID, exists := ctx.Get(middleware.CtxKeyUserID)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user, err := h.authService.Me(ctx, userID.(string))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"user": user, "message": "user retrieved successfully"})
}

func (h *AuthHandler) Logout(ctx *gin.Context) {
	ctx.SetCookie("refresh_token", "", -1, "/", "", false, true)
	ctx.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// func (h *AuthHandler) SetVerificationHash(ctx *gin.Context) {
// 	userID, exists := ctx.Get("userID")
// 	if !exists {
// 		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
// 		return
// 	}

// 	objectID, err := bson.ObjectIDFromHex(userID.(string))
// 	if err != nil {
// 		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
// 		return
// 	}

// 	var req dto.SetVerificationHashRequest
// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	user, err := h.authService.SetVerificationHash(ctx.Request.Context(), objectID, req)
// 	if err != nil {
// 		if errors.Is(err, service.ErrUserNotFound) {
// 			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
// 			return
// 		}
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, gin.H{"user": user})
// }

// func (h *AuthHandler) GetKDFParams(ctx *gin.Context) {
// 	userID, exists := ctx.Get("userID")
// 	if !exists {
// 		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
// 		return
// 	}

// 	objectID, err := bson.ObjectIDFromHex(userID.(string))
// 	if err != nil {
// 		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
// 		return
// 	}

// 	kdf, err := h.authService.GetKDFParams(ctx.Request.Context(), objectID)
// 	if err != nil {
// 		if errors.Is(err, service.ErrUserNotFound) {
// 			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
// 			return
// 		}
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, gin.H{
// 		"algorithm":   kdf.Algorithm,
// 		"salt":        kdf.Salt,
// 		"memory":      kdf.Memory,
// 		"iterations":  kdf.Iterations,
// 		"parallelism": kdf.Parallelism,
// 	})
// }

// func (h *AuthHandler) UpdateEmail(ctx *gin.Context) {
// 	userID, exists := ctx.Get("userID")
// 	if !exists {
// 		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
// 		return
// 	}

// 	objectID, err := bson.ObjectIDFromHex(userID.(string))
// 	if err != nil {
// 		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
// 		return
// 	}

// 	var req dto.UpdateEmailRequest
// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	tokens, user, err := h.authService.UpdateEmail(ctx.Request.Context(), objectID, req)
// 	if err != nil {
// 		switch {
// 		case errors.Is(err, service.ErrUserNotFound):
// 			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
// 		case errors.Is(err, service.ErrCannotEditOAuth):
// 			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		case errors.Is(err, service.ErrEmailInUse):
// 			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
// 		default:
// 			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		}
// 		return
// 	}

// 	ctx.SetCookie("refresh_token", tokens.RefreshToken, 7*24*60*60, "/", "", false, true)

// 	ctx.JSON(http.StatusOK, gin.H{
// 		"access_token":  tokens.AccessToken,
// 		"refresh_token": tokens.RefreshToken,
// 		"user":          user,
// 	})
// }

// func (h *AuthHandler) UpdateAccountPassword(ctx *gin.Context) {
// 	userID, exists := ctx.Get("userID")
// 	if !exists {
// 		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
// 		return
// 	}

// 	objectID, err := bson.ObjectIDFromHex(userID.(string))
// 	if err != nil {
// 		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
// 		return
// 	}

// 	var req dto.UpdateAccountPasswordRequest
// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	if err := h.authService.UpdateAccountPassword(ctx.Request.Context(), objectID, req); err != nil {
// 		switch {
// 		case errors.Is(err, service.ErrUserNotFound):
// 			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
// 		case errors.Is(err, service.ErrCannotEditOAuth):
// 			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		case errors.Is(err, service.ErrInvalidPassword):
// 			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		default:
// 			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		}
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, gin.H{"message": "password successfully updated"})
// }
