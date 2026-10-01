package repositories

import (
	"context"
	"errors"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/infrastructure/persistence/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	driver "go.mongodb.org/mongo-driver/v2/mongo"
)

type serviceAccountRepository struct {
	coll mongo.Collection[domain.ServiceAccount, bson.ObjectID]
}

func NewServiceAccountRepository(db *driver.Database) *serviceAccountRepository {
	return &serviceAccountRepository{
		coll: mongo.NewCollection[domain.ServiceAccount, bson.ObjectID](db, "service_accounts"),
	}
}

func (r *serviceAccountRepository) EnsureIndexes(ctx context.Context) error {
	return r.coll.EnsureIndex(ctx, []string{"organization_id"}, false)
}

func (r *serviceAccountRepository) Create(ctx context.Context, account *domain.ServiceAccount) error {
	return r.coll.Create(ctx, *account)
}

func (r *serviceAccountRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.ServiceAccount, error) {
	account, err := r.coll.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

func (r *serviceAccountRepository) FindByOrganization(ctx context.Context, orgID bson.ObjectID) ([]domain.ServiceAccount, error) {
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

func (r *serviceAccountRepository) Update(ctx context.Context, id bson.ObjectID, update domain.ServiceAccount) error {
	_, err := r.coll.Replace(ctx, id, update)
	return err
}

func (r *serviceAccountRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	return r.coll.DeleteByID(ctx, id)
}