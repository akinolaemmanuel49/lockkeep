package repositories

import (
	"context"
	"errors"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/infrastructure/persistence/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	driver "go.mongodb.org/mongo-driver/v2/mongo"
)

type environmentRepository struct {
	coll mongo.Collection[domain.Environment, bson.ObjectID]
}

func NewEnvironmentRepository(db *driver.Database) *environmentRepository {
	return &environmentRepository{
		coll: mongo.NewCollection[domain.Environment, bson.ObjectID](db, "environments"),
	}
}

func (r *environmentRepository) EnsureIndexes(ctx context.Context) error {
	return r.coll.EnsureIndex(ctx, []string{"application_id", "slug"}, true)
}

func (r *environmentRepository) Create(ctx context.Context, env *domain.Environment) error {
	return r.coll.Create(ctx, *env)
}

func (r *environmentRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Environment, error) {
	env, err := r.coll.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &env, nil
}

func (r *environmentRepository) FindBySlug(ctx context.Context, orgID, appID bson.ObjectID, slug string) (*domain.Environment, error) {
	env, err := r.coll.FindOne(ctx, mongo.Query{Filter: map[string]any{
		"organization_id": orgID,
		"application_id":  appID,
		"slug":            slug,
	}})
	if err != nil {
		if errors.Is(err, mongo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &env, nil
}

func (r *environmentRepository) FindByApplication(ctx context.Context, appID bson.ObjectID) ([]domain.Environment, error) {
	result, err := r.coll.FindMany(ctx,
		mongo.Query{Filter: map[string]any{"application_id": appID}},
		[]mongo.SortField{{Field: "name", Order: mongo.Ascending}},
		mongo.Page{Limit: 100},
	)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (r *environmentRepository) Update(ctx context.Context, id bson.ObjectID, update domain.Environment) error {
	_, err := r.coll.Replace(ctx, id, update)
	return err
}

func (r *environmentRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	return r.coll.DeleteByID(ctx, id)
}