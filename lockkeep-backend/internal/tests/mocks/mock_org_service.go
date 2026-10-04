package mocks

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockOrganizationService struct {
	mock.Mock
}

var _ ports.OrganizationService = (*MockOrganizationService)(nil)

func (m *MockOrganizationService) CreateOrganization(ctx context.Context, userID bson.ObjectID, name, slug string) (*domain.Organization, error) {
	args := m.Called(ctx, userID, name, slug)
	if v := args.Get(0); v != nil {
		return v.(*domain.Organization), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockOrganizationService) GetOrganizationBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	args := m.Called(ctx, slug)
	if v := args.Get(0); v != nil {
		return v.(*domain.Organization), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockOrganizationService) ListOrganizationsByUser(ctx context.Context, userID bson.ObjectID) ([]domain.Organization, error) {
	args := m.Called(ctx, userID)
	if v := args.Get(0); v != nil {
		return v.([]domain.Organization), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockOrganizationService) UpdateOrganization(ctx context.Context, userID bson.ObjectID, orgID bson.ObjectID, update dto.UpdateOrganizationDTO) (*domain.Organization, error) {
	args := m.Called(ctx, userID, orgID, update)
	if v := args.Get(0); v != nil {
		return v.(*domain.Organization), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockOrganizationService) DeleteOrganizations(ctx context.Context, userID bson.ObjectID, orgIDs []bson.ObjectID) error {
	args := m.Called(ctx, userID, orgIDs)
	return args.Error(0)
}