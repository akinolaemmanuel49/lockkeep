package mocks

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

var _ ports.UserService = (*MockUserService)(nil)

func (m *MockUserService) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, userID string, input dto.UpdateUserProfileRequestDTO) (*domain.User, error) {
	args := m.Called(ctx, userID, input)
	return args.Get(0).(*domain.User), args.Error(1)
}
