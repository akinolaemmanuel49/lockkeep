package repositories

import (
	"context"
	"errors"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/infrastructure/persistence/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	driver "go.mongodb.org/mongo-driver/v2/mongo"
)

type teamRepository struct {
	coll mongo.Collection[domain.Team, bson.ObjectID]
}

func NewTeamRepository(db *driver.Database) *teamRepository {
	return &teamRepository{
		coll: mongo.NewCollection[domain.Team, bson.ObjectID](db, "teams"),
	}
}

func (r *teamRepository) EnsureIndexes(ctx context.Context) error {
	return r.coll.EnsureIndex(ctx, []string{"organization_id", "slug"}, true)
}

func (r *teamRepository) Create(ctx context.Context, team *domain.Team) error {
	return r.coll.Create(ctx, *team)
}

func (r *teamRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Team, error) {
	team, err := r.coll.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &team, nil
}

func (r *teamRepository) FindBySlug(ctx context.Context, orgID bson.ObjectID, slug string) (*domain.Team, error) {
	team, err := r.coll.FindOne(ctx, mongo.Query{Filter: map[string]any{
		"organization_id": orgID,
		"slug":            slug,
	}})
	if err != nil {
		if errors.Is(err, mongo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &team, nil
}

func (r *teamRepository) FindByOrganization(ctx context.Context, orgID bson.ObjectID) ([]domain.Team, error) {
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

func (r *teamRepository) Update(ctx context.Context, id bson.ObjectID, update domain.Team) error {
	_, err := r.coll.UpdateByID(ctx, id, mongo.Update{
		Set: map[string]any{
			"name":       update.Name,
			"slug":       update.Slug,
			"updated_at": update.UpdatedAt,
		},
	})
	return err
}

func (r *teamRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	return r.coll.DeleteByID(ctx, id)
}
