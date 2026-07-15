package mocks

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockMembershipRepository struct {
	mock.Mock
}

func (m *MockMembershipRepository) EnsureIndexes(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockMembershipRepository) Create(ctx context.Context, membership *domain.Membership) error {
	args := m.Called(ctx, membership)
	return args.Error(0)
}

func (m *MockMembershipRepository) FindByUserAndOrg(ctx context.Context, userID, orgID bson.ObjectID) (*domain.Membership, error) {
	args := m.Called(ctx, userID, orgID)
	return args.Get(0).(*domain.Membership), args.Error(1)
}

func (m *MockMembershipRepository) FindByOrganization(ctx context.Context, orgID bson.ObjectID) ([]domain.Membership, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]domain.Membership), args.Error(1)
}

func (m *MockMembershipRepository) UpdateRole(ctx context.Context, userID, orgID bson.ObjectID, roleID domain.RoleID) error {
	args := m.Called(ctx, userID, orgID, roleID)
	return args.Error(0)
}

func (m *MockMembershipRepository) UpdateTeamRoles(ctx context.Context, userID, orgID bson.ObjectID, teamRoles []domain.TeamRole) error {
	args := m.Called(ctx, userID, orgID, teamRoles)
	return args.Error(0)
}

func (m *MockMembershipRepository) Delete(ctx context.Context, userID, orgID bson.ObjectID) error {
	args := m.Called(ctx, userID, orgID)
	return args.Error(0)
}

func (m *MockMembershipRepository) FindByUser(ctx context.Context, userID bson.ObjectID) ([]domain.Membership, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Membership), args.Error(1)
}
