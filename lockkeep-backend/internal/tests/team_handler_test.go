package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/handlers"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/middleware"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/services"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/tests/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// baselineOrg is the constant organization available to all team tests
var baselineOrg = &domain.Organization{
	ID:   mocks.MustObjectID(mocks.SampleOrgIDHex),
	Name: "Acme Corp",
	Slug: "acme-corp",
}

func setupTeamRouter(mockService *mocks.MockTeamService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := handlers.NewTeamHandler(mockService)

	orgAccess := func(c *gin.Context) {
		c.Set(middleware.CtxKeyOrgID, mocks.SampleOrgIDHex)
		c.Set(middleware.CtxKeyUserID, mocks.SampleUserIDHex)
		c.Set("organization", baselineOrg)
		c.Next()
	}

	v2 := r.Group("/api/v2")
	org := v2.Group("/organizations/:orgSlug")
	org.Use(orgAccess)
	{
		org.POST("/teams", h.Create)
		org.GET("/teams", h.List)
		org.GET("/teams/:teamSlug", h.Get)
		org.DELETE("/teams/:teamID", h.Delete)
	}

	return r
}

func TestTeamHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc := new(mocks.MockTeamService)
		router := setupTeamRouter(mockSvc)

		team := &domain.Team{
			ID:             bson.NewObjectID(),
			OrganizationID: baselineOrg.ID,
			Name:           "Engineering",
			Slug:           "engineering",
		}

		mockSvc.On("CreateTeam", mock.Anything, mock.Anything, baselineOrg.ID, "Engineering", "engineering").
			Return(team, nil).Once()

		body := `{"name":"Engineering","slug":"engineering"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v2/organizations/acme-corp/teams", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "team created successfully")
		mockSvc.AssertExpectations(t)
	})

	t.Run("slug taken", func(t *testing.T) {
		mockSvc := new(mocks.MockTeamService)
		router := setupTeamRouter(mockSvc)

		mockSvc.On("CreateTeam", mock.Anything, mock.Anything, baselineOrg.ID, "Engineering", "engineering").
			Return(nil, services.ErrTeamSlugTaken).Once()

		body := `{"name":"Engineering","slug":"engineering"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v2/organizations/acme-corp/teams", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("forbidden - no roles", func(t *testing.T) {
		mockSvc := new(mocks.MockTeamService)
		router := setupTeamRouter(mockSvc)

		mockSvc.On("CreateTeam", mock.Anything, mock.Anything, baselineOrg.ID, "Engineering", "engineering").
			Return(nil, services.ErrEntityHasNoRoles).Once()

		body := `{"name":"Engineering","slug":"engineering"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v2/organizations/acme-corp/teams", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("forbidden - invalid roles", func(t *testing.T) {
		mockSvc := new(mocks.MockTeamService)
		router := setupTeamRouter(mockSvc)

		mockSvc.On("CreateTeam", mock.Anything, mock.Anything, baselineOrg.ID, "Engineering", "engineering").
			Return(nil, services.ErrEntityHasInvalidRoles).Once()

		body := `{"name":"Engineering","slug":"engineering"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v2/organizations/acme-corp/teams", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("invalid json - missing name", func(t *testing.T) {
		mockSvc := new(mocks.MockTeamService)
		router := setupTeamRouter(mockSvc)

		body := `{"slug":"engineering"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v2/organizations/acme-corp/teams", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("invalid json - empty name", func(t *testing.T) {
		mockSvc := new(mocks.MockTeamService)
		router := setupTeamRouter(mockSvc)

		body := `{"name":"","slug":"engineering"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v2/organizations/acme-corp/teams", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestTeamHandler_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc := new(mocks.MockTeamService)
		router := setupTeamRouter(mockSvc)

		teams := []domain.Team{
			{ID: bson.NewObjectID(), OrganizationID: baselineOrg.ID, Name: "Engineering", Slug: "engineering"},
			{ID: bson.NewObjectID(), OrganizationID: baselineOrg.ID, Name: "Design", Slug: "design"},
		}

		mockSvc.On("ListTeamsByOrganization", mock.Anything, baselineOrg.ID).
			Return(teams, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v2/organizations/acme-corp/teams", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Engineering")
		assert.Contains(t, w.Body.String(), "Design")
		mockSvc.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		mockSvc := new(mocks.MockTeamService)
		router := setupTeamRouter(mockSvc)

		mockSvc.On("ListTeamsByOrganization", mock.Anything, baselineOrg.ID).
			Return(nil, assert.AnError).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v2/organizations/acme-corp/teams", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestTeamHandler_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc := new(mocks.MockTeamService)
		router := setupTeamRouter(mockSvc)

		team := &domain.Team{
			ID:             bson.NewObjectID(),
			OrganizationID: baselineOrg.ID,
			Name:           "Engineering",
			Slug:           "engineering",
		}

		mockSvc.On("GetTeamBySlug", mock.Anything, baselineOrg.ID, "engineering").
			Return(team, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v2/organizations/acme-corp/teams/engineering", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Engineering")
		mockSvc.AssertExpectations(t)
	})

	t.Run("team not found", func(t *testing.T) {
		mockSvc := new(mocks.MockTeamService)
		router := setupTeamRouter(mockSvc)

		mockSvc.On("GetTeamBySlug", mock.Anything, baselineOrg.ID, "missing").
			Return(nil, services.ErrTeamNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v2/organizations/acme-corp/teams/missing", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		mockSvc := new(mocks.MockTeamService)
		router := setupTeamRouter(mockSvc)

		mockSvc.On("GetTeamBySlug", mock.Anything, baselineOrg.ID, "engineering").
			Return(nil, assert.AnError).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v2/organizations/acme-corp/teams/engineering", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestTeamHandler_Delete(t *testing.T) {
	teamID := mocks.MustObjectID(mocks.SampleTeamIDHex)

	t.Run("success", func(t *testing.T) {
		mockSvc := new(mocks.MockTeamService)
		router := setupTeamRouter(mockSvc)

		mockSvc.On("DeleteTeams", mock.Anything, mock.Anything, baselineOrg.ID, []bson.ObjectID{teamID}).
			Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/v2/organizations/acme-corp/teams/"+teamID.Hex(), nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "team deleted successfully")
		mockSvc.AssertExpectations(t)
	})

	t.Run("forbidden - no roles", func(t *testing.T) {
		mockSvc := new(mocks.MockTeamService)
		router := setupTeamRouter(mockSvc)

		mockSvc.On("DeleteTeams", mock.Anything, mock.Anything, baselineOrg.ID, []bson.ObjectID{teamID}).
			Return(services.ErrEntityHasNoRoles).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/v2/organizations/acme-corp/teams/"+teamID.Hex(), nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("forbidden - invalid roles", func(t *testing.T) {
		mockSvc := new(mocks.MockTeamService)
		router := setupTeamRouter(mockSvc)

		mockSvc.On("DeleteTeams", mock.Anything, mock.Anything, baselineOrg.ID, []bson.ObjectID{teamID}).
			Return(services.ErrEntityHasInvalidRoles).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/v2/organizations/acme-corp/teams/"+teamID.Hex(), nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("invalid team id", func(t *testing.T) {
		mockSvc := new(mocks.MockTeamService)
		router := setupTeamRouter(mockSvc)

		req := httptest.NewRequest(http.MethodDelete, "/api/v2/organizations/acme-corp/teams/invalid-id", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}