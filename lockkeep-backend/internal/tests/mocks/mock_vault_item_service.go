package mocks

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockVaultItemService struct {
	mock.Mock
}

var _ ports.VaultItemService = (*MockVaultItemService)(nil)

func (m *MockVaultItemService) CreateVaultItem(ctx context.Context, userID bson.ObjectID, item *domain.VaultItem) (*domain.VaultItem, error) {
	args := m.Called(ctx, userID, item)
	if v := args.Get(0); v != nil {
		return v.(*domain.VaultItem), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockVaultItemService) ListVaultItems(ctx context.Context, userID bson.ObjectID) ([]domain.VaultItem, error) {
	args := m.Called(ctx, userID)
	if v := args.Get(0); v != nil {
		return v.([]domain.VaultItem), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockVaultItemService) GetVaultItemByID(ctx context.Context, userID bson.ObjectID, vaultItemID bson.ObjectID) (*domain.VaultItem, error) {
	args := m.Called(ctx, userID, vaultItemID)
	if v := args.Get(0); v != nil {
		return v.(*domain.VaultItem), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockVaultItemService) UpdateVaultItem(ctx context.Context, userID bson.ObjectID, vaultItemID bson.ObjectID, update dto.VaultItemUpdate) (*domain.VaultItem, error) {
	args := m.Called(ctx, userID, vaultItemID, update)
	if v := args.Get(0); v != nil {
		return v.(*domain.VaultItem), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockVaultItemService) DeleteVaultItems(ctx context.Context, userID bson.ObjectID, vaultItemIDs []bson.ObjectID) error {
	args := m.Called(ctx, userID, vaultItemIDs)
	return args.Error(0)
}