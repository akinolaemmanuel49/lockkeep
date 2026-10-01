package repositories

import (
	"context"
	"errors"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/infrastructure/persistence/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	driver "go.mongodb.org/mongo-driver/v2/mongo"
)

type keyManagementRepository struct {
	coll mongo.Collection[domain.EncryptedKey, bson.ObjectID]
}

func NewKeyManagementRepository(db *driver.Database) *keyManagementRepository {
	return &keyManagementRepository{
		coll: mongo.NewCollection[domain.EncryptedKey, bson.ObjectID](db, "encrypted_keys"),
	}
}

func (r *keyManagementRepository) EnsureIndexes(ctx context.Context) error {
	return r.coll.EnsureIndex(ctx, []string{"scope", "scope_id"}, true)
}

func (r *keyManagementRepository) Create(ctx context.Context, key *domain.EncryptedKey) error {
	return r.coll.Create(ctx, *key)
}

func (r *keyManagementRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.EncryptedKey, error) {
	key, err := r.coll.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &key, nil
}

func (r *keyManagementRepository) FindByScope(ctx context.Context, scope string, scopeID bson.ObjectID) (*domain.EncryptedKey, error) {
	key, err := r.coll.FindOne(ctx, mongo.Query{Filter: map[string]any{
		"scope":    scope,
		"scope_id": scopeID,
	}})
	if err != nil {
		if errors.Is(err, mongo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &key, nil
}

func (r *keyManagementRepository) Update(ctx context.Context, id bson.ObjectID, update domain.EncryptedKey) error {
	_, err := r.coll.Replace(ctx, id, update)
	return err
}

func (r *keyManagementRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	return r.coll.DeleteByID(ctx, id)
}