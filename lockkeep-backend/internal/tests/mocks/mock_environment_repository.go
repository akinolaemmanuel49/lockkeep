package mocks

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockEnvironmentRepository struct {
	mock.Mock
}

func (m *MockEnvironmentRepository) EnsureIndexes(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockEnvironmentRepository) Create(ctx context.Context, env *domain.Environment) error {
	args := m.Called(ctx, env)
	return args.Error(0)
}

func (m *MockEnvironmentRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Environment, error) {
	args := m.Called(ctx, id)
	if env := args.Get(0); env != nil {
		return env.(*domain.Environment), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockEnvironmentRepository) FindBySlug(ctx context.Context, orgID, appID bson.ObjectID, slug string) (*domain.Environment, error) {
	args := m.Called(ctx, orgID, appID, slug)
	if env := args.Get(0); env != nil {
		return env.(*domain.Environment), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockEnvironmentRepository) FindByApplication(ctx context.Context, appID bson.ObjectID) ([]domain.Environment, error) {
	args := m.Called(ctx, appID)
	return args.Get(0).([]domain.Environment), args.Error(1)
}

func (m *MockEnvironmentRepository) Update(ctx context.Context, id bson.ObjectID, update domain.Environment) error {
	args := m.Called(ctx, id, update)
	return args.Error(0)
}

func (m *MockEnvironmentRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}