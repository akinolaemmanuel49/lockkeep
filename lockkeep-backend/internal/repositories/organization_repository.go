package repositories

import (
	"context"
	"errors"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/infrastructure/persistence/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	driver "go.mongodb.org/mongo-driver/v2/mongo"
)

type organizationRepository struct {
	coll mongo.Collection[domain.Organization, bson.ObjectID]
}

func NewOrganizationRepository(db *driver.Database) *organizationRepository {
	return &organizationRepository{
		coll: mongo.NewCollection[domain.Organization, bson.ObjectID](db, "organizations"),
	}
}

func (r *organizationRepository) EnsureIndexes(ctx context.Context) error {
	return r.coll.EnsureIndex(ctx, []string{"slug"}, true)
}

func (r *organizationRepository) Create(ctx context.Context, org *domain.Organization) error {
	return r.coll.Create(ctx, *org)
}

func (r *organizationRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Organization, error) {
	org, err := r.coll.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &org, nil
}

func (r *organizationRepository) FindBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	org, err := r.coll.FindOne(ctx, mongo.Query{Filter: map[string]any{"slug": slug}})
	if err != nil {
		if errors.Is(err, mongo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &org, nil
}

func (r *organizationRepository) FindByMember(ctx context.Context, userID bson.ObjectID) ([]domain.Organization, error) {
	// This requires a join through memberships — implemented via aggregation
	// For now, placeholder: query organizations where owner is user
	// Full implementation needs membership lookup then org fetch
	result, err := r.coll.FindMany(ctx,
		mongo.Query{Filter: map[string]any{"owner_id": userID}},
		nil,
		mongo.Page{Limit: 100},
	)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (r *organizationRepository) Update(ctx context.Context, id bson.ObjectID, update domain.Organization) error {
	_, err := r.coll.Replace(ctx, id, update)
	return err
}

func (r *organizationRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	return r.coll.DeleteByID(ctx, id)
}
