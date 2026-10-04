package mocks

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockKeyManagementRepository struct {
	mock.Mock
}

func (m *MockKeyManagementRepository) EnsureIndexes(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockKeyManagementRepository) Create(ctx context.Context, key *domain.EncryptedKey) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockKeyManagementRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.EncryptedKey, error) {
	args := m.Called(ctx, id)
	if key := args.Get(0); key != nil {
		return key.(*domain.EncryptedKey), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockKeyManagementRepository) FindByScope(ctx context.Context, scope string, scopeID bson.ObjectID) (*domain.EncryptedKey, error) {
	args := m.Called(ctx, scope, scopeID)
	if key := args.Get(0); key != nil {
		return key.(*domain.EncryptedKey), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockKeyManagementRepository) Update(ctx context.Context, id bson.ObjectID, update domain.EncryptedKey) error {
	args := m.Called(ctx, id, update)
	return args.Error(0)
}

func (m *MockKeyManagementRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}