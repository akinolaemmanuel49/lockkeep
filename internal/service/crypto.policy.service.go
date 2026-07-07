package service

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
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

func (s *CryptoPolicyService) SetCurrentPolicy(ctx context.Context, policy *dto.SetCurrentPolicy) error {
	currentPolicy, err := s.policyRepo.GetCurrent(ctx)
	if err != nil {
		return err
	}

	version := currentPolicy.Version + 1
	newPolicy := s.toPolicy(policy, version)

	return s.policyRepo.SetCurrent(ctx, newPolicy)
}

func (s *CryptoPolicyService) toPolicy(policyDTO *dto.SetCurrentPolicy, version uint32) *domain.CryptoPolicy {
	return &domain.CryptoPolicy{
		ID:      "current",
		Version: version,
		KDFParams: domain.KDFParams{
			Algorithm:   policyDTO.KDFParams.Algorithm,
			Memory:      policyDTO.KDFParams.Memory,
			Iterations:  policyDTO.KDFParams.Iterations,
			Parallelism: policyDTO.KDFParams.Parallelism,
		},
	}
}
