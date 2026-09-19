package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/handlers"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/services"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/tests/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func setupVaultRouter(mockVault *mocks.MockVaultService, mockItems *mocks.MockVaultItemService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	h := handlers.NewVaultHandler(mockVault, mockItems)

	v2 := r.Group("/api/v2")
	v := v2.Group("/me/vault")
	v.Use(mocks.MockAuthMiddleware())
	{
		v.GET("", h.GetVault)
		v.PATCH("", h.UpdateVault)
		v.POST("/verify", h.VerifyVaultPassword)

		v.GET("/items", h.ListVaultItems)
		v.POST("/items", h.CreateVaultItem)
		v.GET("/items/:itemID", h.GetVaultItem)
		v.PUT("/items/:itemID", h.UpdateVaultItem)
		v.DELETE("/items/:itemID", h.DeleteVaultItem)
	}

	return r
}

func TestVaultHandler_VerifyVaultPassword(t *testing.T) {
	userID := mocks.MustObjectID(mocks.SampleObjectIDHex)

	t.Run("success", func(t *testing.T) {
		mockVault := new(mocks.MockVaultService)
		mockItems := new(mocks.MockVaultItemService)
		router := setupVaultRouter(mockVault, mockItems)

		items := []domain.VaultItem{
			{ID: bson.NewObjectID(), UserID: userID, Name: "db", Type: domain.ItemLogin},
		}

		mockVault.On("VerifyVaultPassword", mock.Anything, mocks.SampleObjectIDHex, "correcthash").
			Return(true, nil).Once()
		mockItems.On("ListVaultItems", mock.Anything, userID).Return(items, nil).Once()

		body := `{"verification_hash":"correcthash"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v2/me/vault/verify", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), "db")
		mockVault.AssertExpectations(t)
		mockItems.AssertExpectations(t)
	})

	t.Run("wrong password returns success false", func(t *testing.T) {
		mockVault := new(mocks.MockVaultService)
		mockItems := new(mocks.MockVaultItemService)
		router := setupVaultRouter(mockVault, mockItems)

		mockVault.On("VerifyVaultPassword", mock.Anything, mocks.SampleObjectIDHex, "wronghash").
			Return(false, nil).Once()

		body := `{"verification_hash":"wronghash"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v2/me/vault/verify", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
		mockVault.AssertExpectations(t)
		mockItems.AssertNotCalled(t, "ListVaultItems", mock.Anything, mock.Anything)
	})

	t.Run("user not found", func(t *testing.T) {
		mockVault := new(mocks.MockVaultService)
		mockItems := new(mocks.MockVaultItemService)
		router := setupVaultRouter(mockVault, mockItems)

		mockVault.On("VerifyVaultPassword", mock.Anything, mocks.SampleObjectIDHex, "correcthash").
			Return(false, services.ErrUserNotFound).Once()

		body := `{"verification_hash":"correcthash"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v2/me/vault/verify", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockVault.AssertExpectations(t)
	})
}

func TestVaultHandler_GetVault(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockVault := new(mocks.MockVaultService)
		mockItems := new(mocks.MockVaultItemService)
		router := setupVaultRouter(mockVault, mockItems)

		vault := &domain.Vault{ID: bson.NewObjectID(), Vaultname: "My Vault"}

		mockVault.On("GetVaultByUserID", mock.Anything, mocks.SampleObjectIDHex).
			Return(vault, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v2/me/vault", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "My Vault")
		mockVault.AssertExpectations(t)
	})

	t.Run("vault not found", func(t *testing.T) {
		mockVault := new(mocks.MockVaultService)
		mockItems := new(mocks.MockVaultItemService)
		router := setupVaultRouter(mockVault, mockItems)

		mockVault.On("GetVaultByUserID", mock.Anything, mocks.SampleObjectIDHex).
			Return(nil, services.ErrVaultNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v2/me/vault", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockVault.AssertExpectations(t)
	})
}

func TestVaultHandler_UpdateVault(t *testing.T) {
	mockVault := new(mocks.MockVaultService)
	mockItems := new(mocks.MockVaultItemService)
	router := setupVaultRouter(mockVault, mockItems)

	vault := &domain.Vault{ID: bson.NewObjectID(), Vaultname: "Old Vault"}

	mockVault.On("GetVaultByUserID", mock.Anything, mocks.SampleObjectIDHex).Return(vault, nil).Once()
	mockVault.On("Update", mock.Anything, vault.ID.Hex(), mock.AnythingOfType("dto.UpdateVaultProfileRequestDTO")).
		Return(&domain.Vault{ID: vault.ID, Vaultname: "New Vault"}, nil).Once()

	body := `{"vaultname":"New Vault"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v2/me/vault", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "New Vault")
	mockVault.AssertExpectations(t)
	mockItems.AssertNotCalled(t, "ListVaultItems", mock.Anything, mock.Anything)
}

func TestVaultHandler_ListVaultItems(t *testing.T) {
	mockVault := new(mocks.MockVaultService)
	mockItems := new(mocks.MockVaultItemService)
	router := setupVaultRouter(mockVault, mockItems)

	userID := mocks.MustObjectID(mocks.SampleObjectIDHex)
	items := []domain.VaultItem{
		{ID: bson.NewObjectID(), UserID: userID, Name: "db", Type: domain.ItemLogin},
	}

	mockItems.On("ListVaultItems", mock.Anything, userID).Return(items, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v2/me/vault/items", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "db")
	mockItems.AssertExpectations(t)
}

func TestVaultHandler_CreateVaultItem(t *testing.T) {
	mockVault := new(mocks.MockVaultService)
	mockItems := new(mocks.MockVaultItemService)
	router := setupVaultRouter(mockVault, mockItems)

	userID := mocks.MustObjectID(mocks.SampleObjectIDHex)
	created := &domain.VaultItem{
		ID:     bson.NewObjectID(),
		UserID: userID,
		Name:   "db",
		Type:   domain.ItemLogin,
	}

	mockItems.On("CreateVaultItem", mock.Anything, userID, mock.AnythingOfType("*domain.VaultItem")).
		Return(created, nil).Once()

	body := `{"type":"login","name":"db","secret":{"ciphertext":"abc","iv":"1234","tag":"t","version":1}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v2/me/vault/items", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "db")
	mockItems.AssertExpectations(t)
}

func TestVaultHandler_GetVaultItem(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockVault := new(mocks.MockVaultService)
		mockItems := new(mocks.MockVaultItemService)
		router := setupVaultRouter(mockVault, mockItems)

		userID := mocks.MustObjectID(mocks.SampleObjectIDHex)
		item := &domain.VaultItem{ID: bson.NewObjectID(), UserID: userID, Name: "db"}

		mockItems.On("GetVaultItemByID", mock.Anything, userID, item.ID).Return(item, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v2/me/vault/items/"+item.ID.Hex(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "db")
		mockItems.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockVault := new(mocks.MockVaultService)
		mockItems := new(mocks.MockVaultItemService)
		router := setupVaultRouter(mockVault, mockItems)

		userID := mocks.MustObjectID(mocks.SampleObjectIDHex)
		itemID := bson.NewObjectID()

		mockItems.On("GetVaultItemByID", mock.Anything, userID, itemID).
			Return(nil, services.ErrVaultItemNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v2/me/vault/items/"+itemID.Hex(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockItems.AssertExpectations(t)
	})
}

func TestVaultHandler_UpdateVaultItem(t *testing.T) {
	mockVault := new(mocks.MockVaultService)
	mockItems := new(mocks.MockVaultItemService)
	router := setupVaultRouter(mockVault, mockItems)

	userID := mocks.MustObjectID(mocks.SampleObjectIDHex)
	itemID := bson.NewObjectID()
	updated := &domain.VaultItem{ID: itemID, UserID: userID, Name: "renamed"}

	mockItems.On("UpdateVaultItem", mock.Anything, userID, itemID, mock.AnythingOfType("dto.VaultItemUpdate")).
		Return(updated, nil).Once()

	body := `{"id":"` + itemID.Hex() + `","name":"renamed"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v2/me/vault/items/"+itemID.Hex(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "renamed")
	mockItems.AssertExpectations(t)
}

func TestVaultHandler_DeleteVaultItem(t *testing.T) {
	mockVault := new(mocks.MockVaultService)
	mockItems := new(mocks.MockVaultItemService)
	router := setupVaultRouter(mockVault, mockItems)

	userID := mocks.MustObjectID(mocks.SampleObjectIDHex)
	itemID := bson.NewObjectID()

	mockItems.On("DeleteVaultItems", mock.Anything, userID, []bson.ObjectID{itemID}).Return(nil).Once()

	req := httptest.NewRequest(http.MethodDelete, "/api/v2/me/vault/items/"+itemID.Hex(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "deleted")
	mockItems.AssertExpectations(t)
}