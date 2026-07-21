package mocks

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockOrganizationRepository struct {
	mock.Mock
}

func (m *MockOrganizationRepository) EnsureIndexes(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockOrganizationRepository) Create(ctx context.Context, org *domain.Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *MockOrganizationRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Organization, error) {
	args := m.Called(ctx, id)
	if u := args.Get(0); u != nil {
		return u.(*domain.Organization), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockOrganizationRepository) FindBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	args := m.Called(ctx, slug)
	if u := args.Get(0); u != nil {
		return u.(*domain.Organization), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockOrganizationRepository) FindByMember(ctx context.Context, userID bson.ObjectID) ([]domain.Organization, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) Update(ctx context.Context, id bson.ObjectID, update domain.Organization) error {
	args := m.Called(ctx, id, update)
	return args.Error(0)
}

func (m *MockOrganizationRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
