package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/handlers"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/tests/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Helper to create test router
func setupUserRouter(mockService *mocks.MockUserService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	h := handlers.NewUserHandler(mockService)

	// User routes
	v2 := r.Group("/api/v2")
	me := v2.Group("/me")
	me.Use(mocks.MockAuthMiddleware())
	me.GET("", h.GetMe)

	return r
}

func TestUserHandler_GetMe(t *testing.T) {
	mockSvc := new(mocks.MockUserService)
	router := setupUserRouter(mockSvc)

	user := &domain.User{
		ID:    bson.NewObjectID(),
		Email: "me@example.com",
	}

	mockSvc.On("GetUser", mock.Anything, mocks.SampleObjectIDHex).
		Return(user, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "me@example.com")
	mockSvc.AssertExpectations(t)
}
