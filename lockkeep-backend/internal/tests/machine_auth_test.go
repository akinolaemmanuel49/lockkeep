package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/middleware"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/services"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/tests/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// mintMachineTestKey wires a ServiceAccountService whose VerifyApiKey will
// accept the freshly minted raw key.
func mintMachineTestKey(t *testing.T) (*services.ServiceAccountService, string, bson.ObjectID) {
	t.Helper()
	saRepo := new(mocks.MockServiceAccountRepository)
	apiKeyRepo := new(mocks.MockApiKeyRepository)
	envRepo := new(mocks.MockEnvironmentRepository)
	membershipRepo := new(mocks.MockMembershipRepository)
	svc := services.NewServiceAccountService(saRepo, apiKeyRepo, envRepo, membershipRepo, "hash-secret")

	userID, orgID := bson.NewObjectID(), bson.NewObjectID()
	accountID, envID := bson.NewObjectID(), bson.NewObjectID()

	membershipRepo.On("FindByUserAndOrg", mock.Anything, userID, orgID).Return(ownerMembership(orgID, userID), nil)
	saRepo.On("FindByID", mock.Anything, accountID).Return(&domain.ServiceAccount{ID: accountID, OrganizationID: orgID}, nil)
	envRepo.On("FindByID", mock.Anything, envID).Return(&domain.Environment{ID: envID, OrganizationID: orgID}, nil)

	var stored *domain.ApiKey
	apiKeyRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.ApiKey")).Return(nil).Run(func(args mock.Arguments) {
		stored = args.Get(1).(*domain.ApiKey)
	})

	_, raw, err := svc.MintApiKey(context.Background(), userID, orgID, accountID, envID, []domain.ApiKeyScope{domain.ApiKeyScopeRead}, nil)
	assert.NoError(t, err)

	apiKeyRepo.On("FindByPrefix", mock.Anything, stored.KeyPrefix).Return(stored, nil)
	apiKeyRepo.On("FindByPrefix", mock.Anything, mock.AnythingOfType("string")).Return(nil, nil).Maybe()
	saRepo.On("FindByID", mock.Anything, accountID).Return(&domain.ServiceAccount{ID: accountID, OrganizationID: orgID}, nil).Maybe()
	apiKeyRepo.On("Update", mock.Anything, stored.ID, mock.AnythingOfType("domain.ApiKey")).Return(nil).Maybe()

	return svc, raw, envID
}

func TestMachineAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid key passes and populates context", func(t *testing.T) {
		svc, raw, envID := mintMachineTestKey(t)

		var gotEnv any
		r := gin.New()
		r.GET("/secrets", middleware.MachineAuth(svc), func(c *gin.Context) {
			gotEnv, _ = c.Get(middleware.CtxKeyMachineEnvID)
			secrets, _ := c.Get(middleware.CtxKeyBearerSecret)
			assert.NotEmpty(t, secrets.([]byte))
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/secrets", nil)
		req.Header.Set("Authorization", "Bearer "+raw)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, envID.Hex(), gotEnv)
	})

	t.Run("missing header is rejected", func(t *testing.T) {
		svc, _, _ := mintMachineTestKey(t)

		r := gin.New()
		r.GET("/secrets", middleware.MachineAuth(svc), func(c *gin.Context) { c.Status(http.StatusOK) })

		req := httptest.NewRequest(http.MethodGet, "/secrets", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid key is rejected", func(t *testing.T) {
		svc, _, _ := mintMachineTestKey(t)

		r := gin.New()
		r.GET("/secrets", middleware.MachineAuth(svc), func(c *gin.Context) { c.Status(http.StatusOK) })

		req := httptest.NewRequest(http.MethodGet, "/secrets", nil)
		req.Header.Set("Authorization", "Bearer lk_live_invalidinvalidinvalidinvalidinvalidinvalidinvalid")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}