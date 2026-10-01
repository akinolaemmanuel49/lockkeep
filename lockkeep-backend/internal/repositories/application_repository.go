package repositories

import (
	"context"
	"errors"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/infrastructure/persistence/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	driver "go.mongodb.org/mongo-driver/v2/mongo"
)

type applicationRepository struct {
	coll mongo.Collection[domain.Application, bson.ObjectID]
}

func NewApplicationRepository(db *driver.Database) *applicationRepository {
	return &applicationRepository{
		coll: mongo.NewCollection[domain.Application, bson.ObjectID](db, "applications"),
	}
}

func (r *applicationRepository) EnsureIndexes(ctx context.Context) error {
	return r.coll.EnsureIndex(ctx, []string{"organization_id", "slug"}, true)
}

func (r *applicationRepository) Create(ctx context.Context, app *domain.Application) error {
	return r.coll.Create(ctx, *app)
}

func (r *applicationRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Application, error) {
	app, err := r.coll.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &app, nil
}

func (r *applicationRepository) FindBySlug(ctx context.Context, orgID bson.ObjectID, slug string) (*domain.Application, error) {
	app, err := r.coll.FindOne(ctx, mongo.Query{Filter: map[string]any{
		"organization_id": orgID,
		"slug":            slug,
	}})
	if err != nil {
		if errors.Is(err, mongo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &app, nil
}

func (r *applicationRepository) FindByOrganization(ctx context.Context, orgID bson.ObjectID) ([]domain.Application, error) {
	result, err := r.coll.FindMany(ctx,
		mongo.Query{Filter: map[string]any{"organization_id": orgID}},
		[]mongo.SortField{{Field: "name", Order: mongo.Ascending}},
		mongo.Page{Limit: 100},
	)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (r *applicationRepository) Update(ctx context.Context, id bson.ObjectID, update domain.Application) error {
	_, err := r.coll.Replace(ctx, id, update)
	return err
}

func (r *applicationRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	return r.coll.DeleteByID(ctx, id)
}