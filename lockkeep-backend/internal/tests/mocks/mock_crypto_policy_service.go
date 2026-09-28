package mocks

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"github.com/stretchr/testify/mock"
)

type MockCryptoPolicyService struct {
	mock.Mock
}

var _ ports.CryptoPolicyService = (*MockCryptoPolicyService)(nil)

func (m *MockCryptoPolicyService) GetCurrentPolicy(ctx context.Context) (*domain.CryptoPolicy, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CryptoPolicy), args.Error(1)
}

func (m *MockCryptoPolicyService) SetCurrentPolicy(ctx context.Context, policy *dto.SetCurrentPolicy) error {
	args := m.Called(ctx, policy)
	return args.Error(0)
}