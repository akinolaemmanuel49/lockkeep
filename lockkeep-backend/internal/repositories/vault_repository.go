package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	VaultItemCollectionName = "vault"
	VaultCollectionName     = "vaults"
)

// VaultItemRepository — personal vault items (zero-knowledge: ciphertext only).
type VaultItemRepository struct {
	collection *mongo.Collection
}

func NewVaultItemRepository(db *mongo.Database) *VaultItemRepository {
	return &VaultItemRepository{
		collection: db.Collection(VaultItemCollectionName),
	}
}

func (r *VaultItemRepository) Create(ctx context.Context, item *domain.VaultItem) error {
	item.ID = bson.NewObjectID()
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, item)
	return err
}

func (r *VaultItemRepository) FindByUser(ctx context.Context, userID bson.ObjectID) ([]domain.VaultItem, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"user_id": userID,
	}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []domain.VaultItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *VaultItemRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.VaultItem, error) {
	var item domain.VaultItem
	err := r.collection.FindOne(ctx, bson.M{
		"_id": id,
	}).Decode(&item)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *VaultItemRepository) FindDuplicate(ctx context.Context, userID bson.ObjectID, name string, itemType domain.VaultItemType, excludeID *bson.ObjectID) (bool, error) {
	filter := bson.M{
		"user_id": userID,
		"name":    name,
		"type":    itemType,
	}
	if excludeID != nil {
		filter["_id"] = bson.M{"$ne": *excludeID}
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *VaultItemRepository) Update(ctx context.Context, id bson.ObjectID, item *domain.VaultItem) error {
	item.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx,
		bson.M{"_id": id},
		item,
	)
	return err
}

func (r *VaultItemRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *VaultItemRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "name", Value: 1},
			{Key: "type", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	return err
}

// VaultRepository — personal vault profile documents (one per user).
type VaultRepository struct {
	collection *mongo.Collection
}

func NewVaultRepository(db *mongo.Database) *VaultRepository {
	return &VaultRepository{
		collection: db.Collection(VaultCollectionName),
	}
}

func (r *VaultRepository) Create(ctx context.Context, vault *domain.Vault) error {
	vault.ID = bson.NewObjectID()
	vault.CreatedAt = time.Now()
	vault.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, vault)
	return err
}

func (r *VaultRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Vault, error) {
	var vault domain.Vault
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&vault)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &vault, nil
}

func (r *VaultRepository) FindByUser(ctx context.Context, userID bson.ObjectID) (*domain.Vault, error) {
	var vault domain.Vault
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&vault)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &vault, nil
}

func (r *VaultRepository) UpdateProfile(ctx context.Context, id bson.ObjectID, update domain.Vault) error {
	update.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx,
		bson.M{"_id": id},
		update,
	)
	return err
}

func (r *VaultRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}