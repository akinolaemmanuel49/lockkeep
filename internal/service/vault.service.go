package service

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrVaultNotSet          = errors.New("vault password not set")
	ErrInvalidVaultPassword = errors.New("invalid vault password")
	ErrItemNotFound         = errors.New("vault item not found")
	ErrDuplicateItem        = errors.New("vault item with this name and type already exists")
	ErrMigrationConflict    = errors.New("vault already migrated")
)

type VaultService struct {
	client           *mongo.Client
	userRepo         *repository.UserRepository
	cryptoPolicyRepo *repository.CryptoPolicyRepository
	vaultRepo        *repository.VaultRepository
}

func NewVaultService(client *mongo.Client,
	userRepo *repository.UserRepository,
	cryptoPolicyRepo *repository.CryptoPolicyRepository,
	vaultRepo *repository.VaultRepository) *VaultService {
	return &VaultService{
		client:           client,
		userRepo:         userRepo,
		cryptoPolicyRepo: cryptoPolicyRepo,
		vaultRepo:        vaultRepo,
	}
}

func (s *VaultService) VerifyVaultPassword(ctx context.Context, userID bson.ObjectID, req dto.VerifyVaultPasswordRequest) ([]domain.VaultItem, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if user.Vault.VerificationHash == "" {
		return nil, ErrVaultNotSet
	}

	if subtle.ConstantTimeCompare([]byte(user.Vault.VerificationHash), []byte(req.VerificationHash)) != 1 {
		return nil, ErrInvalidVaultPassword
	}

	items, err := s.vaultRepo.FindByTenant(ctx, user.TenantID)
	if err != nil {
		return nil, err
	}

	if items == nil {
		items = []domain.VaultItem{}
	}

	return items, nil
}

func (s *VaultService) CreateVaultItem(ctx context.Context, userID bson.ObjectID, req dto.CreateVaultItemRequest) (*domain.VaultItem, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	item := &domain.VaultItem{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Type:     domain.VaultItemType(req.Type),
		Name:     req.Name,
		Metadata: req.Metadata,
		Secret: domain.Secret{
			Ciphertext: req.Secret.Ciphertext,
			IV:         req.Secret.IV,
			Tag:        req.Secret.Tag,
			Version:    req.Secret.Version,
		},
	}

	if err := s.vaultRepo.Create(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *VaultService) GetVaultItems(ctx context.Context, userID bson.ObjectID) ([]domain.VaultItem, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	items, err := s.vaultRepo.FindByTenant(ctx, user.TenantID)
	if err != nil {
		return nil, err
	}

	if items == nil {
		items = []domain.VaultItem{}
	}

	return items, nil
}

func (s *VaultService) UpdateVaultItem(
	ctx context.Context,
	userID bson.ObjectID,
	itemID bson.ObjectID,
	req dto.UpdateVaultItemRequest,
	updatePolicy string,
) (*domain.VaultItem, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if req.Name != "" {
		isDup, err := s.vaultRepo.FindDuplicate(ctx, user.ID, user.TenantID, req.Name, domain.VaultItemType(req.Type), &itemID)
		if err != nil {
			return nil, err
		}
		if isDup {
			return nil, ErrDuplicateItem
		}
	}

	updates := bson.M{}
	if req.Type != "" {
		updates["type"] = domain.VaultItemType(req.Type)
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Metadata != nil {
		updates["metadata"] = req.Metadata
	}
	if req.Secret.Ciphertext != "" {
		updates["secret.ciphertext"] = req.Secret.Ciphertext
	}
	if req.Secret.IV != "" {
		updates["secret.iv"] = req.Secret.IV
	}
	if req.Secret.Tag != "" {
		updates["secret.tag"] = req.Secret.Tag
	}

	if updatePolicy == "true" {
		if req.Secret.Ciphertext == "" {
			return nil, errors.New("policy update requires secret re-encryption")
		}

		currentPolicy, err := s.cryptoPolicyRepo.GetCurrent(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch current crypto policy: %w", err)
		}

		updates["secret.version"] = currentPolicy.Version
	}

	if err := s.vaultRepo.Update(ctx, itemID, user.ID, user.TenantID, updates); err != nil {
		return nil, err
	}

	updated, err := s.vaultRepo.FindByID(ctx, itemID, user.ID, user.TenantID)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *VaultService) DeleteVaultItem(ctx context.Context, userID bson.ObjectID, itemID bson.ObjectID) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	return s.vaultRepo.Delete(ctx, itemID, user.ID, user.TenantID)
}

func (s *VaultService) MigrateVault(ctx context.Context, userID bson.ObjectID, req dto.MigrateVaultRequest) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	kdfParams := domain.KDFParams{
		Algorithm:   req.KDFParams.Algorithm,
		Salt:        req.KDFParams.Salt,
		Memory:      req.KDFParams.Memory,
		Iterations:  req.KDFParams.Iterations,
		Parallelism: req.KDFParams.Parallelism,
	}

	session, err := s.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	if err := session.StartTransaction(); err != nil {
		return err
	}

	err = mongo.WithSession(ctx, session, func(ctx context.Context) error {
		if err := s.userRepo.UpdateKDF(ctx, userID, req.ExpectedVersion, req.VerificationHash, kdfParams); err != nil {
			return err
		}

		if err := s.vaultRepo.BulkUpdate(ctx, user.ID, user.TenantID, req.VaultUpdates); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		_ = session.AbortTransaction(ctx)

		if errors.Is(err, repository.ErrVaultAlreadyMigrated) {
			return ErrMigrationConflict
		}
		return err
	}

	if err := session.CommitTransaction(ctx); err != nil {
		_ = session.AbortTransaction(ctx)
		return err
	}

	return nil
}
