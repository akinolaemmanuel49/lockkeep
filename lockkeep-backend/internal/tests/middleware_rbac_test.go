package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/middleware"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/tests/mocks"
	jwtpkg "github.com/akinolaemmanuel49/lockkeep-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRequirePermissionRouter(
	claims *jwtpkg.Claims,
	membership *domain.Membership,
	teamID string,
	perm domain.Permission,
) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	auth := func(c *gin.Context) {
		c.Set(middleware.CtxKeyUserID, mocks.SampleUserIDHex)
		c.Set(middleware.CtxKeyClaims, claims)
		c.Next()
	}

	setContext := func(c *gin.Context) {
		c.Set(middleware.CtxKeyOrgID, mocks.SampleOrgIDHex)
		if teamID != "" {
			c.Set(middleware.CtxKeyTeamID, teamID)
		}
		if membership != nil {
			c.Set(middleware.CtxKeyMembership, membership)
		}
		c.Next()
	}

	ok := func(c *gin.Context) {
		c.Status(http.StatusOK)
	}

	r.Use(auth, setContext)
	r.GET("/guard", middleware.RequirePermission(perm), ok)

	return r
}

func membershipWithRole(roleID domain.RoleID) *domain.Membership {
	return &domain.Membership{
		UserID:         mocks.MustObjectID(mocks.SampleUserIDHex),
		OrganizationID: mocks.MustObjectID(mocks.SampleOrgIDHex),
		RoleID:         roleID,
	}
}

func emptyClaims() *jwtpkg.Claims {
	return &jwtpkg.Claims{
		UserID:     mocks.SampleUserIDHex,
		SystemRole: string(domain.RoleSystemUser),
	}
}

func doGuard(router *gin.Engine) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/guard", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestRequirePermission_FreshlyCreatedWorkspace(t *testing.T) {
	claims := emptyClaims()
	router := setupRequirePermissionRouter(
		claims,
		membershipWithRole(domain.RoleOrgOwner),
		"",
		domain.PermTeamManage,
	)

	w := doGuard(router)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequirePermission_DeniesWhenMembershipLacksPermission(t *testing.T) {
	router := setupRequirePermissionRouter(
		emptyClaims(),
		membershipWithRole(domain.RoleOrgMember),
		"",
		domain.PermTeamManage,
	)

	w := doGuard(router)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "insufficient permissions")
}

func TestRequirePermission_AllowsTeamRoleOnlyWhenGranted(t *testing.T) {
	teamID := mocks.SampleTeamIDHex
	membership := membershipWithRole(domain.RoleOrgMember)
	membership.TeamRoles = []domain.TeamRole{
		{
			TeamID: mocks.MustObjectID(teamID),
			RoleID: domain.RoleTeamAdmin,
		},
	}

	t.Run("granted by team role", func(t *testing.T) {
		router := setupRequirePermissionRouter(
			emptyClaims(),
			membership,
			teamID,
			domain.PermMemberInvite,
		)

		w := doGuard(router)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("not granted by team role", func(t *testing.T) {
		router := setupRequirePermissionRouter(
			emptyClaims(),
			membership,
			teamID,
			domain.PermOrgRead,
		)

		w := doGuard(router)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestRequirePermission_FallsBackToClaimsWithoutMembership(t *testing.T) {
	t.Run("allowed by claim", func(t *testing.T) {
		claims := emptyClaims()
		claims.Orgs = []jwtpkg.OrgClaim{
			{OrgID: mocks.SampleOrgIDHex, Role: string(domain.RoleOrgAdmin)},
		}
		router := setupRequirePermissionRouter(claims, nil, "", domain.PermOrgManage)

		w := doGuard(router)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("denied without claim", func(t *testing.T) {
		router := setupRequirePermissionRouter(emptyClaims(), nil, "", domain.PermOrgManage)

		w := doGuard(router)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestRequirePermission_SystemAdminBypass(t *testing.T) {
	claims := emptyClaims()
	claims.SystemRole = string(domain.RoleSystemAdmin)
	router := setupRequirePermissionRouter(claims, nil, "", domain.PermOrgManage)

	w := doGuard(router)
	assert.Equal(t, http.StatusOK, w.Code)
}