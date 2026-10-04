package services

import (
	"context"
	"time"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type VaultItemService struct {
	itemRepo ports.VaultItemRepository
}

func NewVaultItemService(itemRepo ports.VaultItemRepository) *VaultItemService {
	return &VaultItemService{itemRepo: itemRepo}
}

var _ ports.VaultItemService = (*VaultItemService)(nil)

// CreateVaultItem persists a client-encrypted item. Duplicates are rejected
// per (userID, name, type).
func (s *VaultItemService) CreateVaultItem(ctx context.Context, userID bson.ObjectID, item *domain.VaultItem) (*domain.VaultItem, error) {
	isDup, err := s.itemRepo.FindDuplicate(ctx, userID, item.Name, item.Type, nil)
	if err != nil {
		return nil, err
	}
	if isDup {
		return nil, ErrDuplicateVaultItem
	}

	item.UserID = userID
	item.ID = bson.NewObjectID()
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()

	if err := s.itemRepo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *VaultItemService) ListVaultItems(ctx context.Context, userID bson.ObjectID) ([]domain.VaultItem, error) {
	items, err := s.itemRepo.FindByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []domain.VaultItem{}
	}
	return items, nil
}

func (s *VaultItemService) GetVaultItemByID(ctx context.Context, userID bson.ObjectID, vaultItemID bson.ObjectID) (*domain.VaultItem, error) {
	item, err := s.itemRepo.FindByID(ctx, vaultItemID)
	if err != nil {
		return nil, err
	}
	if item == nil || item.UserID != userID {
		return nil, ErrVaultItemNotFound
	}
	return item, nil
}

// UpdateVaultItem applies partial updates. Replaced ciphertext fields flow
// through only when present; version bumps are the client's concern.
func (s *VaultItemService) UpdateVaultItem(ctx context.Context, userID bson.ObjectID, vaultItemID bson.ObjectID, update dto.VaultItemUpdate) (*domain.VaultItem, error) {
	item, err := s.itemRepo.FindByID(ctx, vaultItemID)
	if err != nil {
		return nil, err
	}
	if item == nil || item.UserID != userID {
		return nil, ErrVaultItemNotFound
	}

	if update.Name != nil && *update.Name != "" || update.Type != nil {
		name := item.Name
		if update.Name != nil && *update.Name != "" {
			name = *update.Name
		}
		itemType := item.Type
		if update.Type != nil {
			itemType = domain.VaultItemType(*update.Type)
		}
		isDup, err := s.itemRepo.FindDuplicate(ctx, userID, name, itemType, &item.ID)
		if err != nil {
			return nil, err
		}
		if isDup {
			return nil, ErrDuplicateVaultItem
		}
	}

	if update.Name != nil && *update.Name != "" {
		item.Name = *update.Name
	}
	if update.Type != nil {
		item.Type = domain.VaultItemType(*update.Type)
	}
	if update.Metadata != nil {
		item.Metadata = *update.Metadata
	}
	if update.Secret != nil {
		item.Secret = domain.Secret{
			Ciphertext: update.Secret.Ciphertext,
			IV:         update.Secret.IV,
			Tag:        update.Secret.Tag,
			Version:    update.Secret.Version,
		}
	}
	item.UpdatedAt = time.Now()

	if err := s.itemRepo.Update(ctx, vaultItemID, item); err != nil {
		return nil, err
	}
	return item, nil
}

// DeleteVaultItems removes the caller-owned items in batch.
func (s *VaultItemService) DeleteVaultItems(ctx context.Context, userID bson.ObjectID, vaultItemIDs []bson.ObjectID) error {
	for _, id := range vaultItemIDs {
		item, err := s.itemRepo.FindByID(ctx, id)
		if err != nil {
			return err
		}
		if item == nil || item.UserID != userID {
			return ErrVaultItemNotFound
		}
		if err := s.itemRepo.Delete(ctx, id); err != nil {
			return err
		}
	}
	return nil
}