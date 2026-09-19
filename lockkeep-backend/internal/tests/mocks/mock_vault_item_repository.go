package mocks

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockVaultItemRepository struct {
	mock.Mock
}

func (m *MockVaultItemRepository) EnsureIndexes(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockVaultItemRepository) Create(ctx context.Context, item *domain.VaultItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockVaultItemRepository) FindByUser(ctx context.Context, userID bson.ObjectID) ([]domain.VaultItem, error) {
	args := m.Called(ctx, userID)
	if v := args.Get(0); v != nil {
		return v.([]domain.VaultItem), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockVaultItemRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.VaultItem, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*domain.VaultItem), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockVaultItemRepository) FindDuplicate(ctx context.Context, userID bson.ObjectID, name string, itemType domain.VaultItemType, excludeID *bson.ObjectID) (bool, error) {
	args := m.Called(ctx, userID, name, itemType, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockVaultItemRepository) Update(ctx context.Context, id bson.ObjectID, item *domain.VaultItem) error {
	args := m.Called(ctx, id, item)
	return args.Error(0)
}

func (m *MockVaultItemRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}