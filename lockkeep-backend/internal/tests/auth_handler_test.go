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

	return r
}

func TestAuthHandler_Register(t *testing.T) {
	mockSvc := new(mocks.MockAuthService)
	router := setupAuthRouter(mockSvc)

	user := &domain.User{
		ID:       bson.NewObjectID(),
		Username: "testuser",
		Email:    "test@example.com",
	}

	mockSvc.On("Register", mock.Anything, mock.AnythingOfType("dto.RegisterRequestDTO")).
		Return(user, nil)

	body := `{"username":"testuser","email":"test@example.com","password":"supersecret123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v2/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "user created successfully")
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Login(t *testing.T) {
	mockSvc := new(mocks.MockAuthService)
	router := setupAuthRouter(mockSvc)

	user := &domain.User{ID: bson.NewObjectID(), Email: "test@example.com"}
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

	user := &domain.User{ID: bson.NewObjectID(), Email: "oauth@example.com", Username: "oauthuser"}
	tokens := &dto.TokenPair{AccessToken: "access.token", RefreshToken: "refresh.token"}

	t.Run("New User", func(t *testing.T) {
		mockSvc.On("OAuth", mock.Anything, "valid.token").Return(user, nil, true, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/v2/auth/oauth", nil)
		req.Header.Set("Authorization", "Bearer valid.token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Existing User", func(t *testing.T) {
		mockSvc.On("OAuth", mock.Anything, "valid.token2").Return(user, tokens, false, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/v2/auth/oauth", nil)
		req.Header.Set("Authorization", "Bearer valid.token2")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "access_token")
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
