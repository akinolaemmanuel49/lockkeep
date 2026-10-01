package mocks

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) EnsureIndexes(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAuditRepository) Log(ctx context.Context, event *domain.AuditEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockAuditRepository) FindByResource(ctx context.Context, resourceType domain.AuditResourceType, resourceID bson.ObjectID, limit int) ([]domain.AuditEvent, error) {
	args := m.Called(ctx, resourceType, resourceID, limit)
	return args.Get(0).([]domain.AuditEvent), args.Error(1)
}

func (m *MockAuditRepository) FindByUser(ctx context.Context, userID bson.ObjectID, limit int) ([]domain.AuditEvent, error) {
	args := m.Called(ctx, userID, limit)
	return args.Get(0).([]domain.AuditEvent), args.Error(1)
}

func (m *MockAuditRepository) FindByOrganization(ctx context.Context, orgID bson.ObjectID, limit int) ([]domain.AuditEvent, error) {
	args := m.Called(ctx, orgID, limit)
	return args.Get(0).([]domain.AuditEvent), args.Error(1)
}