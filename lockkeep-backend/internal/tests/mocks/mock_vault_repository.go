package mocks

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockVaultRepository struct {
	mock.Mock
}

func (m *MockVaultRepository) EnsureIndexes(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockVaultRepository) Create(ctx context.Context, vault *domain.Vault) error {
	args := m.Called(ctx, vault)
	return args.Error(0)
}

func (m *MockVaultRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Vault, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*domain.Vault), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockVaultRepository) FindByUser(ctx context.Context, userID bson.ObjectID) (*domain.Vault, error) {
	args := m.Called(ctx, userID)
	if v := args.Get(0); v != nil {
		return v.(*domain.Vault), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockVaultRepository) UpdateProfile(ctx context.Context, id bson.ObjectID, update domain.Vault) error {
	args := m.Called(ctx, id, update)
	return args.Error(0)
}