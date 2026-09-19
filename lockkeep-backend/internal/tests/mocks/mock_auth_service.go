package mocks

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

var _ ports.AuthService = (*MockAuthService)(nil)

func (m *MockAuthService) Register(ctx context.Context, input dto.RegisterRequestDTO) (*domain.User, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, input dto.LoginRequestDTO) (*domain.User, *dto.TokenPair, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*domain.User), args.Get(1).(*dto.TokenPair), args.Error(2)
}

func (m *MockAuthService) OAuth(ctx context.Context, accessToken string) (*domain.User, *dto.TokenPair, bool, error) {
	args := m.Called(ctx, accessToken)

	user := args.Get(0).(*domain.User)
	var tokens *dto.TokenPair
	if arg1 := args.Get(1); arg1 != nil {
		tokens = arg1.(*dto.TokenPair)
	}
	isNewUser := args.Bool(2)
	err := args.Error(3)

	return user, tokens, isNewUser, err
}

func (m *MockAuthService) Refresh(ctx context.Context, refreshToken string) (*dto.TokenPair, error) {
	args := m.Called(ctx, refreshToken)
	return args.Get(0).(*dto.TokenPair), args.Error(1)
}
