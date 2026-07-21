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

func (m *MockTeamService) Create(ctx context.Context, userID bson.ObjectID, orgID bson.ObjectID, name, slug string) (*domain.Team, error) {
	args := m.Called(ctx, userID, orgID, name, slug)
	return args.Get(0).(*domain.Team), args.Error(1)
}
func (m *MockTeamService) GetBySlug(ctx context.Context, orgID bson.ObjectID, slug string) (*domain.Team, error) {
	args := m.Called(ctx, orgID, slug)
	return args.Get(0).(*domain.Team), args.Error(1)
}
func (m *MockTeamService) ListByOrganization(ctx context.Context, orgID bson.ObjectID) ([]domain.Team, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]domain.Team), args.Error(1)
}
func (m *MockTeamService) Delete(ctx context.Context, userID bson.ObjectID, orgID bson.ObjectID, teamID bson.ObjectID) error {
	args := m.Called(ctx, userID, orgID, teamID)
	return args.Error(0)
}
