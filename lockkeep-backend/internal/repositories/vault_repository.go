package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/dto"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const NAME_VAULT_COLLECTION = "vault"

type VaultRepository struct {
	collection *mongo.Collection
}

func NewVaultRepository(db *mongo.Database) *VaultRepository {
	return &VaultRepository{
		collection: db.Collection(NAME_VAULT_COLLECTION),
	}
}

func (r *VaultRepository) Create(ctx context.Context, item *domain.VaultItem) error {
	item.ID = bson.NewObjectID()
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, item)
	return err
}

func (r *VaultRepository) FindByTenant(ctx context.Context, tenantID string) ([]domain.VaultItem, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"tenant_id": tenantID,
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

func (r *VaultRepository) FindByUser(ctx context.Context, userID bson.ObjectID, tenantID string) ([]domain.VaultItem, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"user_id":   userID,
		"tenant_id": tenantID,
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

func (r *VaultRepository) FindByID(ctx context.Context, id bson.ObjectID, userID bson.ObjectID, tenantID string) (*domain.VaultItem, error) {
	var item domain.VaultItem
	err := r.collection.FindOne(ctx, bson.M{
		"_id":       id,
		"user_id":   userID,
		"tenant_id": tenantID,
	}).Decode(&item)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *VaultRepository) Update(ctx context.Context, id bson.ObjectID, userID bson.ObjectID, tenantID string, updates bson.M) error {
	updates["updated_at"] = time.Now()
	_, err := r.collection.UpdateOne(ctx,
		bson.M{
			"_id":       id,
			"user_id":   userID,
			"tenant_id": tenantID,
		},
		bson.M{"$set": updates},
	)
	return err
}

func (r *VaultRepository) BulkUpdate(
	ctx context.Context,
	userID bson.ObjectID,
	tenantID string,
	updates []dto.VaultItemUpdate,
) error {

	models := make([]mongo.WriteModel, 0, len(updates))

	for _, update := range updates {
		id, err := bson.ObjectIDFromHex(update.ID)
		if err != nil {
			return err
		}

		models = append(models,
			mongo.NewUpdateOneModel().
				SetFilter(bson.M{
					"_id":       id,
					"user_id":   userID,
					"tenant_id": tenantID,
				}).
				SetUpdate(bson.M{
					"$set": bson.M{
						"secret.ciphertext": update.Secret.Ciphertext,
						"secret.iv":         update.Secret.IV,
						"secret.tag":        update.Secret.Tag,
						"secret.version":    update.Secret.Version,
						"updated_at":        time.Now(),
					},
				}),
		)
	}

	_, err := r.collection.BulkWrite(ctx, models)
	return err
}

func (r *VaultRepository) Delete(ctx context.Context, id bson.ObjectID, userID bson.ObjectID, tenantID string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{
		"_id":       id,
		"user_id":   userID,
		"tenant_id": tenantID,
	})
	return err
}

func (r *VaultRepository) FindDuplicate(ctx context.Context, userID bson.ObjectID, tenantID string, name string, itemType domain.VaultItemType, excludeID *bson.ObjectID) (bool, error) {
	filter := bson.M{
		"user_id":   userID,
		"tenant_id": tenantID,
		"name":      name,
		"type":      itemType,
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

func (r *VaultRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "tenant_id", Value: 1},
			{Key: "name", Value: 1},
			{Key: "type", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	return err
}
