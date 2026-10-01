package mocks

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockApiKeyRepository struct {
	mock.Mock
}

func (m *MockApiKeyRepository) EnsureIndexes(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockApiKeyRepository) Create(ctx context.Context, key *domain.ApiKey) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockApiKeyRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.ApiKey, error) {
	args := m.Called(ctx, id)
	if key := args.Get(0); key != nil {
		return key.(*domain.ApiKey), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockApiKeyRepository) FindByPrefix(ctx context.Context, prefix string) (*domain.ApiKey, error) {
	args := m.Called(ctx, prefix)
	if key := args.Get(0); key != nil {
		return key.(*domain.ApiKey), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockApiKeyRepository) FindByServiceAccount(ctx context.Context, serviceAccountID bson.ObjectID) ([]domain.ApiKey, error) {
	args := m.Called(ctx, serviceAccountID)
	return args.Get(0).([]domain.ApiKey), args.Error(1)
}

func (m *MockApiKeyRepository) Update(ctx context.Context, id bson.ObjectID, update domain.ApiKey) error {
	args := m.Called(ctx, id, update)
	return args.Error(0)
}

func (m *MockApiKeyRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}