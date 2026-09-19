package mocks

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockIdentityRepository struct {
	mock.Mock
}

func (m *MockIdentityRepository) EnsureIndexes(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockIdentityRepository) Create(ctx context.Context, identity *domain.Identity) error {
	args := m.Called(ctx, identity)
	return args.Error(0)
}

func (m *MockIdentityRepository) FindByOAuth(ctx context.Context, method domain.AuthMethod, providerID string) (*domain.Identity, error) {
	args := m.Called(ctx, method, providerID)
	if id := args.Get(0); id != nil {
		return id.(*domain.Identity), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockIdentityRepository) FindByUserID(ctx context.Context, userID bson.ObjectID) ([]domain.Identity, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Identity), args.Error(1)
}

func (m *MockIdentityRepository) UpdatePassword(ctx context.Context, userID bson.ObjectID, passwordHash string) error {
	args := m.Called(ctx, userID, passwordHash)
	return args.Error(0)
}

func (m *MockIdentityRepository) RecordLogin(ctx context.Context, id bson.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
