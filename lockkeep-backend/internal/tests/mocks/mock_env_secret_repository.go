package mocks

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockEnvSecretRepository struct {
	mock.Mock
}

func (m *MockEnvSecretRepository) EnsureIndexes(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockEnvSecretRepository) Create(ctx context.Context, secret *domain.EnvSecret) error {
	args := m.Called(ctx, secret)
	return args.Error(0)
}

func (m *MockEnvSecretRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.EnvSecret, error) {
	args := m.Called(ctx, id)
	if secret := args.Get(0); secret != nil {
		return secret.(*domain.EnvSecret), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockEnvSecretRepository) FindByKey(ctx context.Context, environmentID bson.ObjectID, key string) (*domain.EnvSecret, error) {
	args := m.Called(ctx, environmentID, key)
	if secret := args.Get(0); secret != nil {
		return secret.(*domain.EnvSecret), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockEnvSecretRepository) FindByEnvironment(ctx context.Context, environmentID bson.ObjectID) ([]domain.EnvSecret, error) {
	args := m.Called(ctx, environmentID)
	return args.Get(0).([]domain.EnvSecret), args.Error(1)
}

func (m *MockEnvSecretRepository) Update(ctx context.Context, id bson.ObjectID, update domain.EnvSecret) error {
	args := m.Called(ctx, id, update)
	return args.Error(0)
}

func (m *MockEnvSecretRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}