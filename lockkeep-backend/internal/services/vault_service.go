package services

import (
	"context"
	"crypto/subtle"
	"time"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/config"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type VaultService struct {
	cfg       *config.Config
	VaultRepo ports.VaultRepository
	userRepo  ports.UserRepository
}

func NewVaultService(cfg *config.Config, VaultRepo ports.VaultRepository, userRepo ports.UserRepository) *VaultService {
	return &VaultService{cfg: cfg, VaultRepo: VaultRepo, userRepo: userRepo}
}

var _ ports.VaultService = (*VaultService)(nil)

// VerifyVaultPassword reports whether the client-derived verification hash
// matches the one stored during vault setup. Constant-time comparison to avoid
// leaking hash equality through timing.
func (s *VaultService) VerifyVaultPassword(ctx context.Context, userID string, verificationHash string) (bool, error) {
	objID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}

	user, err := s.userRepo.FindByID(ctx, objID)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, ErrUserNotFound
	}
	if user.Vault == nil || user.Vault.VerificationHash == "" {
		return false, nil
	}

	return subtle.ConstantTimeCompare([]byte(user.Vault.VerificationHash), []byte(verificationHash)) == 1, nil
}

func (s *VaultService) GetVaultByUserID(ctx context.Context, userID string) (*domain.Vault, error) {
	objID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	vault, err := s.VaultRepo.FindByUser(ctx, objID)
	if err != nil {
		return nil, err
	}
	if vault != nil {
		return vault, nil
	}

	// No profile yet — create one on first access for users who have already
	// completed vault setup (their encryption metadata lives on the account).
	user, err := s.userRepo.FindByID(ctx, objID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	if user.Vault == nil || user.Vault.VerificationHash == "" {
		return nil, ErrVaultNotFound
	}

	vaultname := "My Vault"
	if user.Username != "" {
		vaultname = user.Username
	}

	vault = &domain.Vault{
		UserID:    objID,
		Vaultname: vaultname,
	}
	if err := s.VaultRepo.Create(ctx, vault); err != nil {
		return nil, err
	}

	return vault, nil
}

func (s *VaultService) Me(ctx context.Context, VaultID string) (*domain.Vault, error) {
	objID, err := bson.ObjectIDFromHex(VaultID)
	if err != nil {
		return nil, err
	}

	Vault, err := s.VaultRepo.FindByID(ctx, objID)
	if err != nil {
		return nil, err
	}
	if Vault == nil {
		return nil, ErrVaultNotFound
	}
	return Vault, nil
}

func (s *VaultService) Update(ctx context.Context, VaultID string, input dto.UpdateVaultProfileRequestDTO) (*domain.Vault, error) {
	objID, err := bson.ObjectIDFromHex(VaultID)
	if err != nil {
		return nil, err
	}

	Vault, err := s.VaultRepo.FindByID(ctx, objID)
	if err != nil {
		return nil, err
	}
	if Vault == nil {
		return nil, ErrVaultNotFound
	}

	if (input.Vaultname == nil || *input.Vaultname == "") && input.AvatarURL == nil {
		return nil, ErrInvalidProfileUpdate
	}

	if input.Vaultname != nil && *input.Vaultname != "" {
		Vault.Vaultname = *input.Vaultname
	}

	if input.AvatarURL != nil {
		Vault.AvatarURL = input.AvatarURL
	}

	Vault.UpdatedAt = time.Now()

	if err := s.VaultRepo.UpdateProfile(ctx, Vault.ID, *Vault); err != nil {
		return nil, err
	}

	return Vault, nil
}
