package mocks

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockOrganizationService struct {
	mock.Mock
}

var _ ports.OrganizationService = (*MockOrganizationService)(nil)

func (m *MockOrganizationService) Create(ctx context.Context, userID bson.ObjectID, name, slug string) (*domain.Organization, error) {
	args := m.Called(ctx, userID, name, slug)
	return args.Get(0).(*domain.Organization), args.Error(1)
}

func (m *MockOrganizationService) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	args := m.Called(ctx, slug)
	return args.Get(0).(*domain.Organization), args.Error(1)
}

func (m *MockOrganizationService) ListForUser(ctx context.Context, userID bson.ObjectID) ([]domain.Organization, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Organization), args.Error(1)
}

func (m *MockOrganizationService) Update(ctx context.Context, userID bson.ObjectID, slug string, name string) (*domain.Organization, error) {
	args := m.Called(ctx, userID, userID, slug, name)
	return args.Get(0).(*domain.Organization), args.Error(1)
}

func (m *MockOrganizationService) Delete(ctx context.Context, userID bson.ObjectID, slug string) error {
	args := m.Called(ctx, userID, slug)
	return args.Error(0)
}
