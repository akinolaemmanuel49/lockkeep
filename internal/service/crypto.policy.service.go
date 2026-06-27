package service

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/repository"
)

type CryptoPolicyService struct {
	policyRepo *repository.CryptoPolicyRepository
}

func NewCryptoPolicyService(policyRepo *repository.CryptoPolicyRepository) *CryptoPolicyService {
	return &CryptoPolicyService{policyRepo: policyRepo}
}

func (s *CryptoPolicyService) GetCurrentPolicy(ctx context.Context) (*domain.CryptoPolicy, error) {
	return s.policyRepo.GetCurrent(ctx)
}

func (s *CryptoPolicyService) SetCurrentPolicy(ctx context.Context, policy *domain.CryptoPolicy) error {
	return s.policyRepo.SetCurrent(ctx, policy)
}
