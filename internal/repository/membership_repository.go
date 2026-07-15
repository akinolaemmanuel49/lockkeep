package repository

import (
	"context"
	"errors"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/infrastructure/persistence/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	driver "go.mongodb.org/mongo-driver/v2/mongo"
)

type membershipRepository struct {
	coll mongo.Collection[domain.Membership, bson.ObjectID]
}

func NewMembershipRepository(db *driver.Database) *membershipRepository {
	return &membershipRepository{
		coll: mongo.NewCollection[domain.Membership, bson.ObjectID](db, "memberships"),
	}
}

func (r *membershipRepository) EnsureIndexes(ctx context.Context) error {
	if err := r.coll.EnsureIndex(ctx, []string{"user_id", "organization_id"}, true); err != nil {
		return err
	}
	if err := r.coll.EnsureIndex(ctx, []string{"organization_id"}, false); err != nil {
		return err
	}
	return r.coll.EnsureIndex(ctx, []string{"user_id"}, false)
}

func (r *membershipRepository) Create(ctx context.Context, membership *domain.Membership) error {
	return r.coll.Create(ctx, *membership)
}

func (r *membershipRepository) FindByUserAndOrg(ctx context.Context, userID, orgID bson.ObjectID) (*domain.Membership, error) {
	membership, err := r.coll.FindOne(ctx, mongo.Query{Filter: map[string]any{
		"user_id":         userID,
		"organization_id": orgID,
	}})
	if err != nil {
		if errors.Is(err, mongo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &membership, nil
}

func (r *membershipRepository) FindByUser(ctx context.Context, userID bson.ObjectID) ([]domain.Membership, error) {
	result, err := r.coll.FindMany(ctx,
		mongo.Query{Filter: map[string]any{"user_id": userID}},
		nil,
		mongo.Page{Limit: 100},
	)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (r *membershipRepository) FindByOrganization(ctx context.Context, orgID bson.ObjectID) ([]domain.Membership, error) {
	result, err := r.coll.FindMany(ctx,
		mongo.Query{Filter: map[string]any{"organization_id": orgID}},
		nil,
		mongo.Page{Limit: 1000},
	)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (r *membershipRepository) UpdateRole(ctx context.Context, userID, orgID bson.ObjectID, roleID domain.RoleID) error {
	_, err := r.coll.UpdateOne(ctx,
		mongo.Query{Filter: map[string]any{
			"user_id":         userID,
			"organization_id": orgID,
		}},
		mongo.Update{
			Set: map[string]any{"role_id": roleID},
		},
	)
	return err
}

func (r *membershipRepository) UpdateTeamRoles(ctx context.Context, userID, orgID bson.ObjectID, teamRoles []domain.TeamRole) error {
	_, err := r.coll.UpdateOne(ctx,
		mongo.Query{Filter: map[string]any{
			"user_id":         userID,
			"organization_id": orgID,
		}},
		mongo.Update{
			Set: map[string]any{"team_roles": teamRoles},
		},
	)
	return err
}

func (r *membershipRepository) Delete(ctx context.Context, userID, orgID bson.ObjectID) error {
	return r.coll.DeleteOne(ctx, mongo.Query{Filter: map[string]any{
		"user_id":         userID,
		"organization_id": orgID,
	}})
}
