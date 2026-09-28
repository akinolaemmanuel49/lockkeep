package tests

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/handlers"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/middleware"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/services"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/tests/mocks"
	jwtpkg "github.com/akinolaemmanuel49/lockkeep-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupCryptoPolicyRouter(mockService *mocks.MockCryptoPolicyService, claims *jwtpkg.Claims) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := handlers.NewCryptoPolicyHandler(mockService)

	auth := func(c *gin.Context) {
		c.Set(middleware.CtxKeyUserID, claims.UserID)
		c.Set(middleware.CtxKeyClaims, claims)
		c.Next()
	}

	v2 := r.Group("/api/v2")
	cryptoPolicy := v2.Group("/crypto-policy")
	cryptoPolicy.Use(auth)
	{
		cryptoPolicy.GET("/current", h.GetCurrentPolicy)
		cryptoPolicy.POST("/current", h.SetCurrentPolicy)
	}

	return r
}

func testPolicy() *domain.CryptoPolicy {
	return &domain.CryptoPolicy{
		ID:      "current",
		Version: 1,
		KDFParams: domain.KDFParams{
			Algorithm:   "argon2id",
			Memory:      65536,
			Iterations:  3,
			Parallelism: 4,
		},
		UpdatedAt: time.Now(),
	}
}

func userClaims() *jwtpkg.Claims {
	return &jwtpkg.Claims{
		UserID:     mocks.SampleUserIDHex,
		Email:      "user@lockkeep.test",
		SystemRole: string(domain.RoleSystemUser),
	}
}

func adminClaims() *jwtpkg.Claims {
	return &jwtpkg.Claims{
		UserID:     mocks.SampleUserIDHex,
		Email:      "admin@lockkeep.test",
		SystemRole: string(domain.RoleSystemAdmin),
	}
}

func TestCryptoPolicyHandler_GetCurrentPolicy(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc := new(mocks.MockCryptoPolicyService)
		router := setupCryptoPolicyRouter(mockSvc, userClaims())

		mockSvc.On("GetCurrentPolicy", mock.Anything).Return(testPolicy(), nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v2/crypto-policy/current", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"version":1`)
		assert.Contains(t, w.Body.String(), `"algorithm":"argon2id"`)
		mockSvc.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockSvc := new(mocks.MockCryptoPolicyService)
		router := setupCryptoPolicyRouter(mockSvc, userClaims())

		mockSvc.On("GetCurrentPolicy", mock.Anything).Return(nil, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v2/crypto-policy/current", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "no crypto policy configured")
		mockSvc.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		mockSvc := new(mocks.MockCryptoPolicyService)
		router := setupCryptoPolicyRouter(mockSvc, userClaims())

		mockSvc.On("GetCurrentPolicy", mock.Anything).Return(nil, errors.New("db down")).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v2/crypto-policy/current", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestCryptoPolicyHandler_SetCurrentPolicy(t *testing.T) {
	body := `{"kdfParams":{"algorithm":"argon2id","memory":65536,"iterations":3,"parallelism":4}}`

	t.Run("admin success", func(t *testing.T) {
		mockSvc := new(mocks.MockCryptoPolicyService)
		router := setupCryptoPolicyRouter(mockSvc, adminClaims())

		mockSvc.On("SetCurrentPolicy", mock.Anything, mock.Anything).Return(nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/v2/crypto-policy/current", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "crypto policy updated")
		mockSvc.AssertExpectations(t)
	})

	t.Run("forbidden for non-admin", func(t *testing.T) {
		mockSvc := new(mocks.MockCryptoPolicyService)
		router := setupCryptoPolicyRouter(mockSvc, userClaims())

		req := httptest.NewRequest(http.MethodPost, "/api/v2/crypto-policy/current", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "system administrator privileges required")
		mockSvc.AssertNotCalled(t, "SetCurrentPolicy", mock.Anything, mock.Anything)
	})

	t.Run("malformed payload", func(t *testing.T) {
		mockSvc := new(mocks.MockCryptoPolicyService)
		router := setupCryptoPolicyRouter(mockSvc, adminClaims())

		req := httptest.NewRequest(http.MethodPost, "/api/v2/crypto-policy/current", bytes.NewBufferString(`{bad`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "SetCurrentPolicy", mock.Anything, mock.Anything)
	})

	t.Run("nothing to update", func(t *testing.T) {
		mockSvc := new(mocks.MockCryptoPolicyService)
		router := setupCryptoPolicyRouter(mockSvc, adminClaims())

		mockSvc.On("SetCurrentPolicy", mock.Anything, mock.Anything).
			Return(services.ErrNothingToUpdate).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/v2/crypto-policy/current", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertExpectations(t)
	})
}