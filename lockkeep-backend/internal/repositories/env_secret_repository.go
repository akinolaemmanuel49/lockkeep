package repositories

import (
	"context"
	"errors"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/infrastructure/persistence/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	driver "go.mongodb.org/mongo-driver/v2/mongo"
)

type envSecretRepository struct {
	coll mongo.Collection[domain.EnvSecret, bson.ObjectID]
}

func NewEnvSecretRepository(db *driver.Database) *envSecretRepository {
	return &envSecretRepository{
		coll: mongo.NewCollection[domain.EnvSecret, bson.ObjectID](db, "env_secrets"),
	}
}

func (r *envSecretRepository) EnsureIndexes(ctx context.Context) error {
	return r.coll.EnsureIndex(ctx, []string{"environment_id", "key"}, true)
}

func (r *envSecretRepository) Create(ctx context.Context, secret *domain.EnvSecret) error {
	return r.coll.Create(ctx, *secret)
}

func (r *envSecretRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.EnvSecret, error) {
	secret, err := r.coll.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &secret, nil
}

func (r *envSecretRepository) FindByKey(ctx context.Context, environmentID bson.ObjectID, key string) (*domain.EnvSecret, error) {
	secret, err := r.coll.FindOne(ctx, mongo.Query{Filter: map[string]any{
		"environment_id": environmentID,
		"key":            key,
	}})
	if err != nil {
		if errors.Is(err, mongo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &secret, nil
}

func (r *envSecretRepository) FindByEnvironment(ctx context.Context, environmentID bson.ObjectID) ([]domain.EnvSecret, error) {
	result, err := r.coll.FindMany(ctx,
		mongo.Query{Filter: map[string]any{"environment_id": environmentID}},
		[]mongo.SortField{{Field: "key", Order: mongo.Ascending}},
		mongo.Page{Limit: 500},
	)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (r *envSecretRepository) Update(ctx context.Context, id bson.ObjectID, update domain.EnvSecret) error {
	_, err := r.coll.Replace(ctx, id, update)
	return err
}

func (r *envSecretRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	return r.coll.DeleteByID(ctx, id)
}