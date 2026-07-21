// internal/tests/organization_handler_test.go
package tests

import (
	"bytes"
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

func setupOrgRouter(mockSvc *mocks.MockOrganizationService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	h := handlers.NewOrganizationHandler(mockSvc)

	v2 := r.Group("/api/v2")
	orgs := v2.Group("/organizations")
	orgs.Use(mocks.MockAuthMiddleware())

	{
		orgs.GET("", h.List)
		orgs.POST("", h.Create)

		org := orgs.Group("/:orgSlug")
		org.Use(mocks.MockRequireOrgAccess())
		{
			org.GET("", h.Get)
			org.DELETE("", h.Delete)
		}
	}

	return r
}

func TestOrganizationHandler_Create(t *testing.T) {
	mockSvc := new(mocks.MockOrganizationService)
	router := setupOrgRouter(mockSvc)

	org := &domain.Organization{
		ID:   bson.NewObjectID(),
		Name: "Test Org",
		Slug: "test-org",
	}

	// Use mock.Anything for userID since it's generated in middleware
	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("bson.ObjectID"), "Test Org", "test-org").
		Return(org, nil)

	body := `{"name":"Test Org","slug":"test-org"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v2/organizations", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "organization created successfully")
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_List(t *testing.T) {
	mockSvc := new(mocks.MockOrganizationService)
	router := setupOrgRouter(mockSvc)

	orgs := []domain.Organization{
		{ID: bson.NewObjectID(), Name: "Org 1"},
	}

	mockSvc.On("ListForUser", mock.Anything, mock.AnythingOfType("bson.ObjectID")).Return(orgs, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/organizations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Org 1")
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_Get(t *testing.T) {
	mockSvc := new(mocks.MockOrganizationService)
	router := setupOrgRouter(mockSvc)

	slug := "my-org"
	org := &domain.Organization{ID: bson.NewObjectID(), Slug: slug, Name: "My Org"}

	mockSvc.On("GetBySlug", mock.Anything, slug).Return(org, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/organizations/my-org", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "My Org")
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_Delete(t *testing.T) {
	mockSvc := new(mocks.MockOrganizationService)
	router := setupOrgRouter(mockSvc)

	slug := "my-org"

	mockSvc.On("Delete", mock.Anything, mock.AnythingOfType("bson.ObjectID"), slug).
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v2/organizations/my-org", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "organization deleted successfully")
	mockSvc.AssertExpectations(t)
}
