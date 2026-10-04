package mocks

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockTeamService struct {
	mock.Mock
}

var _ ports.TeamService = (*MockTeamService)(nil)

func (m *MockTeamService) CreateTeam(ctx context.Context, userID bson.ObjectID, orgID bson.ObjectID, name, slug string) (*domain.Team, error) {
	args := m.Called(ctx, userID, orgID, name, slug)
	if v := args.Get(0); v != nil {
		return v.(*domain.Team), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTeamService) GetTeamBySlug(ctx context.Context, orgID bson.ObjectID, slug string) (*domain.Team, error) {
	args := m.Called(ctx, orgID, slug)
	if v := args.Get(0); v != nil {
		return v.(*domain.Team), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTeamService) ListTeamsByOrganization(ctx context.Context, orgID bson.ObjectID) ([]domain.Team, error) {
	args := m.Called(ctx, orgID)
	if v := args.Get(0); v != nil {
		return v.([]domain.Team), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTeamService) UpdateTeam(ctx context.Context, orgID bson.ObjectID, teamID bson.ObjectID) (*domain.Team, error) {
	args := m.Called(ctx, orgID, teamID)
	if v := args.Get(0); v != nil {
		return v.(*domain.Team), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTeamService) DeleteTeams(ctx context.Context, userID bson.ObjectID, orgID bson.ObjectID, teamIDs []bson.ObjectID) error {
	args := m.Called(ctx, userID, orgID, teamIDs)
	return args.Error(0)
}