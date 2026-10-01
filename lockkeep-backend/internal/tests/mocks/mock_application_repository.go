package mocks

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockApplicationRepository struct {
	mock.Mock
}

func (m *MockApplicationRepository) EnsureIndexes(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockApplicationRepository) Create(ctx context.Context, app *domain.Application) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockApplicationRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Application, error) {
	args := m.Called(ctx, id)
	if app := args.Get(0); app != nil {
		return app.(*domain.Application), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockApplicationRepository) FindBySlug(ctx context.Context, orgID bson.ObjectID, slug string) (*domain.Application, error) {
	args := m.Called(ctx, orgID, slug)
	if app := args.Get(0); app != nil {
		return app.(*domain.Application), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockApplicationRepository) FindByOrganization(ctx context.Context, orgID bson.ObjectID) ([]domain.Application, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]domain.Application), args.Error(1)
}

func (m *MockApplicationRepository) Update(ctx context.Context, id bson.ObjectID, update domain.Application) error {
	args := m.Called(ctx, id, update)
	return args.Error(0)
}

func (m *MockApplicationRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}