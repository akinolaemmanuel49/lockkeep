// internal/tests/auth_handler_test.go
package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/handlers"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/services"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/tests/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Helper to create test router
func setupAuthRouter(mockService *mocks.MockAuthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	h := handlers.NewAuthHandler(mockService)

	// Auth routes
	v2 := r.Group("/api/v2")
	auth := v2.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	auth.POST("/oauth", h.OAuth)
	auth.POST("/refresh", h.Refresh)

	// Authenticated account + vault-setup routes
	me := v2.Group("/me")
	me.Use(mocks.MockAuthMiddleware())
	{
		me.POST("/email", h.UpdateEmail)
		me.POST("/password", h.UpdateAccountPassword)
		me.POST("/vault/setup", h.SetVerificationHash)
		me.PUT("/vault/setup", h.SetVerificationHash)
		me.GET("/vault/kdfparams", h.GetKDFParams)
	}

	return r
}

func TestAuthHandler_SetVerificationHash(t *testing.T) {
	mockSvc := new(mocks.MockAuthService)
	router := setupAuthRouter(mockSvc)

	user := &domain.User{ID: mocks.MustObjectID(mocks.SampleObjectIDHex), Email: "vault@example.com"}

	mockSvc.On("SetVerificationHash", mock.Anything, mocks.MustObjectID(mocks.SampleObjectIDHex),
		mock.AnythingOfType("dto.SetVerificationHashRequest")).Return(user, nil).Once()

	body := `{"verification_hash":"hash123","kdf_params":{"algorithm":"scrypt","salt":"salt","memory":128,"iterations":17,"parallelism":1}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v2/me/vault/setup", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "vault@example.com")
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_GetKDFParams(t *testing.T) {
	mockSvc := new(mocks.MockAuthService)
	router := setupAuthRouter(mockSvc)

	kdf := &domain.KDFParams{Algorithm: "scrypt", Salt: "salt", Memory: 128, Iterations: 17, Parallelism: 1}

	mockSvc.On("GetKDFParams", mock.Anything, mocks.MustObjectID(mocks.SampleObjectIDHex)).Return(kdf, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v2/me/vault/kdfparams", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"algorithm":"scrypt"`)
	assert.Contains(t, w.Body.String(), `"salt":"salt"`)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_UpdateEmail(t *testing.T) {
	mockSvc := new(mocks.MockAuthService)
	router := setupAuthRouter(mockSvc)

	user := &dto.UserResponse{ID: mocks.MustObjectID(mocks.SampleObjectIDHex).Hex(), Email: "new@example.com"}
	tokens := &dto.TokenPair{AccessToken: "access.token", RefreshToken: "refresh.token"}

	mockSvc.On("UpdateEmail", mock.Anything, mocks.MustObjectID(mocks.SampleObjectIDHex),
		mock.AnythingOfType("dto.UpdateEmailRequest")).Return(user, tokens, nil).Once()

	body := `{"email":"new@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v2/me/email", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "access_token")
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_UpdateAccountPassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc := new(mocks.MockAuthService)
		router := setupAuthRouter(mockSvc)

		mockSvc.On("UpdateAccountPassword", mock.Anything, mocks.MustObjectID(mocks.SampleObjectIDHex),
			mock.AnythingOfType("dto.UpdateAccountPasswordRequest")).Return(nil).Once()

		body := `{"currentPassword":"old-pass","newPassword":"new-pass-123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v2/me/password", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "password successfully updated")
		mockSvc.AssertExpectations(t)
	})

	t.Run("invalid current password", func(t *testing.T) {
		mockSvc := new(mocks.MockAuthService)
		router := setupAuthRouter(mockSvc)

		mockSvc.On("UpdateAccountPassword", mock.Anything, mocks.MustObjectID(mocks.SampleObjectIDHex),
			mock.AnythingOfType("dto.UpdateAccountPasswordRequest")).Return(services.ErrInvalidPassword).Once()

		body := `{"currentPassword":"wrong","newPassword":"new-pass-123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v2/me/password", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestAuthHandler_Register(t *testing.T) {
	mockSvc := new(mocks.MockAuthService)
	router := setupAuthRouter(mockSvc)

	user := &dto.UserResponse{
		ID:                bson.NewObjectID().Hex(),
		Email:             "test@example.com",
		AuthMethod:        "local",
		HasMasterPassword: false,
	}
	tokens := &dto.TokenPair{AccessToken: "access.token", RefreshToken: "refresh.token"}

	mockSvc.On("Register", mock.Anything, mock.AnythingOfType("dto.RegisterRequestDTO")).
		Return(user, tokens, nil)

	body := `{"username":"testuser","email":"test@example.com","password":"supersecret123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v2/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "user created successfully")
	assert.Contains(t, w.Body.String(), "access_token")
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Login(t *testing.T) {
	mockSvc := new(mocks.MockAuthService)
	router := setupAuthRouter(mockSvc)

	user := &dto.UserResponse{ID: bson.NewObjectID().Hex(), Email: "test@example.com", AuthMethod: "local"}
	tokens := &dto.TokenPair{AccessToken: "access.token", RefreshToken: "refresh.token"}

	mockSvc.On("Login", mock.Anything, mock.AnythingOfType("dto.LoginRequestDTO")).
		Return(user, tokens, nil)

	body := `{"email":"test@example.com","password":"supersecret123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v2/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "access_token")
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_OAuth(t *testing.T) {
	mockSvc := new(mocks.MockAuthService)
	router := setupAuthRouter(mockSvc)

	user := &dto.UserResponse{ID: bson.NewObjectID().Hex(), Email: "oauth@example.com", AuthMethod: "oauth_google"}
	tokens := &dto.TokenPair{AccessToken: "access.token", RefreshToken: "refresh.token"}

	t.Run("New User", func(t *testing.T) {
		mockSvc2 := new(mocks.MockAuthService)
		r2 := setupAuthRouter(mockSvc2)
		mockSvc2.On("OAuth", mock.Anything, "valid.token").Return(user, tokens, true, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/v2/auth/oauth", nil)
		req.Header.Set("Authorization", "Bearer valid.token")

		w := httptest.NewRecorder()
		r2.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "access_token")
		mockSvc2.AssertExpectations(t)
	})

	t.Run("Existing User", func(t *testing.T) {
		mockSvc.On("OAuth", mock.Anything, "valid.token2").Return(user, tokens, false, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/v2/auth/oauth", nil)
		req.Header.Set("Authorization", "Bearer valid.token2")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "access_token")
		mockSvc.AssertExpectations(t)
	})
}

func TestAuthHandler_Refresh(t *testing.T) {
	mockSvc := new(mocks.MockAuthService)
	router := setupAuthRouter(mockSvc)

	tokens := &dto.TokenPair{AccessToken: "new.access.token", RefreshToken: "new.refresh.token"}

	mockSvc.On("Refresh", mock.Anything, "old.refresh.token").
		Return(tokens, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v2/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "old.refresh.token"})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "new.access.token")
}
