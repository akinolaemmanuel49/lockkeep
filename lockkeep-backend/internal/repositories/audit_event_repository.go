package repositories

import (
	"context"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/infrastructure/persistence/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	driver "go.mongodb.org/mongo-driver/v2/mongo"
)

type auditEventRepository struct {
	coll mongo.Collection[domain.AuditEvent, bson.ObjectID]
}

func NewAuditEventRepository(db *driver.Database) *auditEventRepository {
	return &auditEventRepository{
		coll: mongo.NewCollection[domain.AuditEvent, bson.ObjectID](db, "audit_events"),
	}
}

func (r *auditEventRepository) EnsureIndexes(ctx context.Context) error {
	return r.coll.EnsureIndex(ctx, []string{"resource_id"}, false)
}

func (r *auditEventRepository) Log(ctx context.Context, event *domain.AuditEvent) error {
	return r.coll.Create(ctx, *event)
}

func (r *auditEventRepository) FindByResource(ctx context.Context, resourceType domain.AuditResourceType, resourceID bson.ObjectID, limit int) ([]domain.AuditEvent, error) {
	result, err := r.coll.FindMany(ctx,
		mongo.Query{Filter: map[string]any{
			"resource_type": resourceType,
			"resource_id":   resourceID,
		}},
		[]mongo.SortField{{Field: "timestamp", Order: mongo.Descending}},
		mongo.Page{Limit: limit},
	)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (r *auditEventRepository) FindByUser(ctx context.Context, userID bson.ObjectID, limit int) ([]domain.AuditEvent, error) {
	result, err := r.coll.FindMany(ctx,
		mongo.Query{Filter: map[string]any{"user_id": userID}},
		[]mongo.SortField{{Field: "timestamp", Order: mongo.Descending}},
		mongo.Page{Limit: limit},
	)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (r *auditEventRepository) FindByOrganization(ctx context.Context, orgID bson.ObjectID, limit int) ([]domain.AuditEvent, error) {
	result, err := r.coll.FindMany(ctx,
		mongo.Query{Filter: map[string]any{
			"organization_id": orgID,
		}},
		[]mongo.SortField{{Field: "timestamp", Order: mongo.Descending}},
		mongo.Page{Limit: limit},
	)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}