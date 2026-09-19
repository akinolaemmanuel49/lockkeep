package tests

import (
	"context"
	"testing"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/config"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/services"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func newTestVaultService() (*services.VaultService, *mocks.MockVaultRepository, *mocks.MockUserRepository) {
	cfg := &config.Config{}
	vaultRepo := new(mocks.MockVaultRepository)
	userRepo := new(mocks.MockUserRepository)
	svc := services.NewVaultService(cfg, vaultRepo, userRepo)
	return svc, vaultRepo, userRepo
}

func TestVaultService_VerifyVaultPassword(t *testing.T) {
	baseUserID := bson.NewObjectID()

	t.Run("success", func(t *testing.T) {
		svc, _, userRepo := newTestVaultService()

		user := &domain.User{
			ID: baseUserID,
			Vault: &domain.VaultMetadata{
				Version:          1,
				VerificationHash: "abc123",
			},
		}
		userRepo.On("FindByID", mock.Anything, baseUserID).Return(user, nil).Once()

		ok, err := svc.VerifyVaultPassword(context.Background(), baseUserID.Hex(), "abc123")
		assert.NoError(t, err)
		assert.True(t, ok)
		userRepo.AssertExpectations(t)
	})

	t.Run("mismatch returns false", func(t *testing.T) {
		svc, _, userRepo := newTestVaultService()

		user := &domain.User{
			ID: baseUserID,
			Vault: &domain.VaultMetadata{
				Version:          1,
				VerificationHash: "abc123",
			},
		}
		userRepo.On("FindByID", mock.Anything, baseUserID).Return(user, nil).Once()

		ok, err := svc.VerifyVaultPassword(context.Background(), baseUserID.Hex(), "different")
		assert.NoError(t, err)
		assert.False(t, ok)
		userRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		svc, _, userRepo := newTestVaultService()

		userRepo.On("FindByID", mock.Anything, baseUserID).Return(nil, nil).Once()

		_, err := svc.VerifyVaultPassword(context.Background(), baseUserID.Hex(), "abc123")
		assert.ErrorIs(t, err, services.ErrUserNotFound)
		userRepo.AssertExpectations(t)
	})
}

func TestVaultService_GetVaultByUserID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, vaultRepo, _ := newTestVaultService()

		vault := &domain.Vault{ID: bson.NewObjectID(), UserID: bson.NewObjectID(), Vaultname: "My Vault"}

		vaultRepo.On("FindByUser", mock.Anything, vault.UserID).Return(vault, nil).Once()

		result, err := svc.GetVaultByUserID(context.Background(), vault.UserID.Hex())

		assert.NoError(t, err)
		assert.Equal(t, vault.ID, result.ID)
		vaultRepo.AssertExpectations(t)
	})

	t.Run("vault not found", func(t *testing.T) {
		svc, vaultRepo, _ := newTestVaultService()

		userID := bson.NewObjectID()
		vaultRepo.On("FindByUser", mock.Anything, userID).Return(nil, nil).Once()

		_, err := svc.GetVaultByUserID(context.Background(), userID.Hex())
		assert.ErrorIs(t, err, services.ErrVaultNotFound)
		vaultRepo.AssertExpectations(t)
	})

	t.Run("invalid user id", func(t *testing.T) {
		svc, _, _ := newTestVaultService()

		_, err := svc.GetVaultByUserID(context.Background(), "not-a-valid-id")
		assert.Error(t, err)
	})
}

func TestVaultService_Me(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, vaultRepo, _ := newTestVaultService()

		vault := &domain.Vault{ID: bson.NewObjectID(), Vaultname: "My Vault"}
		vaultRepo.On("FindByID", mock.Anything, vault.ID).Return(vault, nil).Once()

		result, err := svc.Me(context.Background(), vault.ID.Hex())

		assert.NoError(t, err)
		assert.Equal(t, vault.ID, result.ID)
		vaultRepo.AssertExpectations(t)
	})

	t.Run("vault not found", func(t *testing.T) {
		svc, vaultRepo, _ := newTestVaultService()

		vaultID := bson.NewObjectID()
		vaultRepo.On("FindByID", mock.Anything, vaultID).Return(nil, nil).Once()

		_, err := svc.Me(context.Background(), vaultID.Hex())
		assert.ErrorIs(t, err, services.ErrVaultNotFound)
		vaultRepo.AssertExpectations(t)
	})
}

func TestVaultService_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, vaultRepo, _ := newTestVaultService()

		vault := &domain.Vault{ID: bson.NewObjectID(), Vaultname: "Old Vault"}
		vaultRepo.On("FindByID", mock.Anything, vault.ID).Return(vault, nil).Once()
		vaultRepo.On("UpdateProfile", mock.Anything, vault.ID, mock.AnythingOfType("domain.Vault")).Return(nil).Once()

		result, err := svc.Update(context.Background(), vault.ID.Hex(), dto.UpdateVaultProfileRequestDTO{
			Vaultname: strPtr("New Vault"),
		})

		assert.NoError(t, err)
		assert.Equal(t, "New Vault", result.Vaultname)
		vaultRepo.AssertExpectations(t)
	})

	t.Run("no fields to update", func(t *testing.T) {
		svc, vaultRepo, _ := newTestVaultService()

		vault := &domain.Vault{ID: bson.NewObjectID(), Vaultname: "My Vault"}
		vaultRepo.On("FindByID", mock.Anything, vault.ID).Return(vault, nil).Once()

		_, err := svc.Update(context.Background(), vault.ID.Hex(), dto.UpdateVaultProfileRequestDTO{})

		assert.ErrorIs(t, err, services.ErrInvalidProfileUpdate)
		vaultRepo.AssertExpectations(t)
	})

	t.Run("vault not found", func(t *testing.T) {
		svc, vaultRepo, _ := newTestVaultService()

		vaultID := bson.NewObjectID()
		vaultRepo.On("FindByID", mock.Anything, vaultID).Return(nil, nil).Once()

		_, err := svc.Update(context.Background(), vaultID.Hex(), dto.UpdateVaultProfileRequestDTO{
			Vaultname: strPtr("New Vault"),
		})

		assert.ErrorIs(t, err, services.ErrVaultNotFound)
		vaultRepo.AssertExpectations(t)
	})
}