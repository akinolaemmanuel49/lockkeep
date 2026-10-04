package tests

import (
	"context"
	"testing"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/services"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func newTestVaultItemService() (*services.VaultItemService, *mocks.MockVaultItemRepository) {
	itemRepo := new(mocks.MockVaultItemRepository)
	svc := services.NewVaultItemService(itemRepo)
	return svc, itemRepo
}

func TestVaultItemService_CreateVaultItem(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, itemRepo := newTestVaultItemService()

		userID := bson.NewObjectID()
		item := &domain.VaultItem{Name: "db", Type: domain.ItemLogin}

		itemRepo.On("FindDuplicate", mock.Anything, userID, "db", domain.ItemLogin, mock.Anything).Return(false, nil).Once()
		itemRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.VaultItem")).Return(nil).Once()

		created, err := svc.CreateVaultItem(context.Background(), userID, item)

		assert.NoError(t, err)
		assert.Equal(t, userID, created.UserID)
		assert.False(t, created.ID.IsZero())
		itemRepo.AssertExpectations(t)
	})

	t.Run("duplicate rejected", func(t *testing.T) {
		svc, itemRepo := newTestVaultItemService()

		userID := bson.NewObjectID()
		item := &domain.VaultItem{Name: "db", Type: domain.ItemLogin}

		itemRepo.On("FindDuplicate", mock.Anything, userID, "db", domain.ItemLogin, mock.Anything).Return(true, nil).Once()

		_, err := svc.CreateVaultItem(context.Background(), userID, item)
		assert.ErrorIs(t, err, services.ErrDuplicateVaultItem)
		itemRepo.AssertExpectations(t)
	})
}

func TestVaultItemService_ListVaultItems(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, itemRepo := newTestVaultItemService()

		userID := bson.NewObjectID()
		items := []domain.VaultItem{{ID: bson.NewObjectID(), Name: "db"}}

		itemRepo.On("FindByUser", mock.Anything, userID).Return(items, nil).Once()

		result, err := svc.ListVaultItems(context.Background(), userID)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		itemRepo.AssertExpectations(t)
	})

	t.Run("empty list", func(t *testing.T) {
		svc, itemRepo := newTestVaultItemService()

		userID := bson.NewObjectID()
		itemRepo.On("FindByUser", mock.Anything, userID).Return(nil, nil).Once()

		result, err := svc.ListVaultItems(context.Background(), userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
		itemRepo.AssertExpectations(t)
	})
}

func TestVaultItemService_GetVaultItemByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, itemRepo := newTestVaultItemService()

		userID := bson.NewObjectID()
		item := &domain.VaultItem{ID: bson.NewObjectID(), UserID: userID, Name: "db"}

		itemRepo.On("FindByID", mock.Anything, item.ID).Return(item, nil).Once()

		result, err := svc.GetVaultItemByID(context.Background(), userID, item.ID)

		assert.NoError(t, err)
		assert.Equal(t, item.ID, result.ID)
		itemRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		svc, itemRepo := newTestVaultItemService()

		userID := bson.NewObjectID()
		itemID := bson.NewObjectID()
		itemRepo.On("FindByID", mock.Anything, itemID).Return(nil, nil).Once()

		_, err := svc.GetVaultItemByID(context.Background(), userID, itemID)
		assert.ErrorIs(t, err, services.ErrVaultItemNotFound)
		itemRepo.AssertExpectations(t)
	})

	t.Run("other users item not visible", func(t *testing.T) {
		svc, itemRepo := newTestVaultItemService()

		userID := bson.NewObjectID()
		otherID := bson.NewObjectID()
		item := &domain.VaultItem{ID: bson.NewObjectID(), UserID: otherID, Name: "private"}

		itemRepo.On("FindByID", mock.Anything, item.ID).Return(item, nil).Once()

		_, err := svc.GetVaultItemByID(context.Background(), userID, item.ID)
		assert.ErrorIs(t, err, services.ErrVaultItemNotFound)
		itemRepo.AssertExpectations(t)
	})
}

func TestVaultItemService_UpdateVaultItem(t *testing.T) {
	t.Run("success - rename", func(t *testing.T) {
		svc, itemRepo := newTestVaultItemService()

		userID := bson.NewObjectID()
		item := &domain.VaultItem{ID: bson.NewObjectID(), UserID: userID, Name: "old", Type: domain.ItemLogin}

		itemRepo.On("FindByID", mock.Anything, item.ID).Return(item, nil).Once()
		itemRepo.On("FindDuplicate", mock.Anything, userID, "new", domain.ItemLogin, mock.Anything).Return(false, nil).Once()
		itemRepo.On("Update", mock.Anything, item.ID, mock.AnythingOfType("*domain.VaultItem")).Return(nil).Once()

		result, err := svc.UpdateVaultItem(context.Background(), userID, item.ID, dto.VaultItemUpdate{
			Name: strPtr("new"),
		})

		assert.NoError(t, err)
		assert.Equal(t, "new", result.Name)
		itemRepo.AssertExpectations(t)
	})

	t.Run("success - re-encrypt secret", func(t *testing.T) {
		svc, itemRepo := newTestVaultItemService()

		userID := bson.NewObjectID()
		item := &domain.VaultItem{ID: bson.NewObjectID(), UserID: userID, Name: "db", Type: domain.ItemLogin}

		itemRepo.On("FindByID", mock.Anything, item.ID).Return(item, nil).Once()
		itemRepo.On("Update", mock.Anything, item.ID, mock.AnythingOfType("*domain.VaultItem")).Return(nil).Once()

		result, err := svc.UpdateVaultItem(context.Background(), userID, item.ID, dto.VaultItemUpdate{
			Secret: &dto.SecretDTO{Ciphertext: "c2", IV: "iv2", Tag: "tag2", Version: 2},
		})

		assert.NoError(t, err)
		assert.Equal(t, "c2", result.Secret.Ciphertext)
		assert.Equal(t, uint32(2), result.Secret.Version)
		itemRepo.AssertExpectations(t)
	})

	t.Run("duplicate name rejected", func(t *testing.T) {
		svc, itemRepo := newTestVaultItemService()

		userID := bson.NewObjectID()
		item := &domain.VaultItem{ID: bson.NewObjectID(), UserID: userID, Name: "db", Type: domain.ItemLogin}

		itemRepo.On("FindByID", mock.Anything, item.ID).Return(item, nil).Once()
		itemRepo.On("FindDuplicate", mock.Anything, userID, "taken", domain.ItemLogin, mock.Anything).Return(true, nil).Once()

		_, err := svc.UpdateVaultItem(context.Background(), userID, item.ID, dto.VaultItemUpdate{
			Name: strPtr("taken"),
		})
		assert.ErrorIs(t, err, services.ErrDuplicateVaultItem)
		itemRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		svc, itemRepo := newTestVaultItemService()

		userID := bson.NewObjectID()
		itemID := bson.NewObjectID()
		itemRepo.On("FindByID", mock.Anything, itemID).Return(nil, nil).Once()

		_, err := svc.UpdateVaultItem(context.Background(), userID, itemID, dto.VaultItemUpdate{
			Name: strPtr("new"),
		})
		assert.ErrorIs(t, err, services.ErrVaultItemNotFound)
		itemRepo.AssertExpectations(t)
	})
}

func TestVaultItemService_DeleteVaultItems(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, itemRepo := newTestVaultItemService()

		userID := bson.NewObjectID()
		item := &domain.VaultItem{ID: bson.NewObjectID(), UserID: userID, Name: "db"}

		itemRepo.On("FindByID", mock.Anything, item.ID).Return(item, nil).Once()
		itemRepo.On("Delete", mock.Anything, item.ID).Return(nil).Once()

		err := svc.DeleteVaultItems(context.Background(), userID, []bson.ObjectID{item.ID})
		assert.NoError(t, err)
		itemRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		svc, itemRepo := newTestVaultItemService()

		userID := bson.NewObjectID()
		itemID := bson.NewObjectID()
		itemRepo.On("FindByID", mock.Anything, itemID).Return(nil, nil).Once()

		err := svc.DeleteVaultItems(context.Background(), userID, []bson.ObjectID{itemID})
		assert.ErrorIs(t, err, services.ErrVaultItemNotFound)
		itemRepo.AssertExpectations(t)
	})
}