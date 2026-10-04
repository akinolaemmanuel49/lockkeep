package mocks

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockAuthService struct {
	mock.Mock
}

var _ ports.AuthService = (*MockAuthService)(nil)

func (m *MockAuthService) Register(ctx context.Context, input dto.RegisterRequestDTO) (*dto.UserResponse, *dto.TokenPair, error) {
	args := m.Called(ctx, input)
	var user *dto.UserResponse
	if v := args.Get(0); v != nil {
		user = v.(*dto.UserResponse)
	}
	var tokens *dto.TokenPair
	if v := args.Get(1); v != nil {
		tokens = v.(*dto.TokenPair)
	}
	return user, tokens, args.Error(2)
}

func (m *MockAuthService) Login(ctx context.Context, input dto.LoginRequestDTO) (*dto.UserResponse, *dto.TokenPair, error) {
	args := m.Called(ctx, input)
	var user *dto.UserResponse
	if v := args.Get(0); v != nil {
		user = v.(*dto.UserResponse)
	}
	var tokens *dto.TokenPair
	if v := args.Get(1); v != nil {
		tokens = v.(*dto.TokenPair)
	}
	return user, tokens, args.Error(2)
}

func (m *MockAuthService) OAuth(ctx context.Context, accessToken string) (*dto.UserResponse, *dto.TokenPair, bool, error) {
	args := m.Called(ctx, accessToken)

	var user *dto.UserResponse
	if v := args.Get(0); v != nil {
		user = v.(*dto.UserResponse)
	}
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
	if v := args.Get(0); v != nil {
		return v.(*dto.TokenPair), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthService) SetVerificationHash(ctx context.Context, userID bson.ObjectID, req dto.SetVerificationHashRequest) (*domain.User, error) {
	args := m.Called(ctx, userID, req)
	if v := args.Get(0); v != nil {
		return v.(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthService) GetKDFParams(ctx context.Context, userID bson.ObjectID) (*domain.KDFParams, error) {
	args := m.Called(ctx, userID)
	if v := args.Get(0); v != nil {
		return v.(*domain.KDFParams), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthService) UpdateEmail(ctx context.Context, userID bson.ObjectID, req dto.UpdateEmailRequest) (*dto.UserResponse, *dto.TokenPair, error) {
	args := m.Called(ctx, userID, req)
	var user *dto.UserResponse
	if v := args.Get(0); v != nil {
		user = v.(*dto.UserResponse)
	}
	var tokens *dto.TokenPair
	if v := args.Get(1); v != nil {
		tokens = v.(*dto.TokenPair)
	}
	return user, tokens, args.Error(2)
}

func (m *MockAuthService) UpdateAccountPassword(ctx context.Context, userID bson.ObjectID, req dto.UpdateAccountPasswordRequest) error {
	args := m.Called(ctx, userID, req)
	return args.Error(0)
}
