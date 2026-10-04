package repositories

import (
	"context"
	"errors"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/infrastructure/persistence/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	driver "go.mongodb.org/mongo-driver/v2/mongo"
)

type apiKeyRepository struct {
	coll mongo.Collection[domain.ApiKey, bson.ObjectID]
}

func NewApiKeyRepository(db *driver.Database) *apiKeyRepository {
	return &apiKeyRepository{
		coll: mongo.NewCollection[domain.ApiKey, bson.ObjectID](db, "api_keys"),
	}
}

func (r *apiKeyRepository) EnsureIndexes(ctx context.Context) error {
	return r.coll.EnsureIndex(ctx, []string{"service_account_id"}, false)
}

func (r *apiKeyRepository) Create(ctx context.Context, key *domain.ApiKey) error {
	return r.coll.Create(ctx, *key)
}

func (r *apiKeyRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.ApiKey, error) {
	key, err := r.coll.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &key, nil
}

func (r *apiKeyRepository) FindByPrefix(ctx context.Context, prefix string) (*domain.ApiKey, error) {
	key, err := r.coll.FindOne(ctx, mongo.Query{Filter: map[string]any{"key_prefix": prefix}})
	if err != nil {
		if errors.Is(err, mongo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &key, nil
}

func (r *apiKeyRepository) FindByServiceAccount(ctx context.Context, serviceAccountID bson.ObjectID) ([]domain.ApiKey, error) {
	result, err := r.coll.FindMany(ctx,
		mongo.Query{Filter: map[string]any{"service_account_id": serviceAccountID}},
		[]mongo.SortField{{Field: "created_at", Order: mongo.Descending}},
		mongo.Page{Limit: 100},
	)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (r *apiKeyRepository) Update(ctx context.Context, id bson.ObjectID, update domain.ApiKey) error {
	_, err := r.coll.Replace(ctx, id, update)
	return err
}

func (r *apiKeyRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	return r.coll.DeleteByID(ctx, id)
}