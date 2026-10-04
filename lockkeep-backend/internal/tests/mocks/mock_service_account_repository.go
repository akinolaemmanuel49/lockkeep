package mocks

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockServiceAccountRepository struct {
	mock.Mock
}

func (m *MockServiceAccountRepository) EnsureIndexes(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockServiceAccountRepository) Create(ctx context.Context, account *domain.ServiceAccount) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *MockServiceAccountRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.ServiceAccount, error) {
	args := m.Called(ctx, id)
	if account := args.Get(0); account != nil {
		return account.(*domain.ServiceAccount), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockServiceAccountRepository) FindByOrganization(ctx context.Context, orgID bson.ObjectID) ([]domain.ServiceAccount, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]domain.ServiceAccount), args.Error(1)
}

func (m *MockServiceAccountRepository) Update(ctx context.Context, id bson.ObjectID, update domain.ServiceAccount) error {
	args := m.Called(ctx, id, update)
	return args.Error(0)
}

func (m *MockServiceAccountRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}