package mocks

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"github.com/stretchr/testify/mock"
)

type MockVaultService struct {
	mock.Mock
}

var _ ports.VaultService = (*MockVaultService)(nil)

func (m *MockVaultService) VerifyVaultPassword(ctx context.Context, userID string, verificationHash string) (bool, error) {
	args := m.Called(ctx, userID, verificationHash)
	return args.Bool(0), args.Error(1)
}

func (m *MockVaultService) Me(ctx context.Context, vaultID string) (*domain.Vault, error) {
	args := m.Called(ctx, vaultID)
	if v := args.Get(0); v != nil {
		return v.(*domain.Vault), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockVaultService) GetVaultByUserID(ctx context.Context, userID string) (*domain.Vault, error) {
	args := m.Called(ctx, userID)
	if v := args.Get(0); v != nil {
		return v.(*domain.Vault), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockVaultService) Update(ctx context.Context, vaultID string, input dto.UpdateVaultProfileRequestDTO) (*domain.Vault, error) {
	args := m.Called(ctx, vaultID, input)
	if v := args.Get(0); v != nil {
		return v.(*domain.Vault), args.Error(1)
	}
	return nil, args.Error(1)
}