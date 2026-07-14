package repository

import (
	"context"
	"errors"
	"time"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/infrastructure/persistence/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	driver "go.mongodb.org/mongo-driver/v2/mongo"
)

type identityRepository struct {
	coll mongo.Collection[domain.Identity, bson.ObjectID]
}

func NewIdentityRepository(db *driver.Database) *identityRepository {
	return &identityRepository{
		coll: mongo.NewCollection[domain.Identity, bson.ObjectID](db, "identities"),
	}
}

func (r *identityRepository) EnsureIndexes(ctx context.Context) error {
	if err := r.coll.EnsureIndex(ctx, []string{"user_id"}, false); err != nil {
		return err
	}
	return r.coll.EnsurePartialIndex(ctx,
		[]string{"auth_method", "auth_provider_id"},
		map[string]any{
			"auth_method": map[string]any{
				"$in": []string{"oauth_google", "oauth_github"},
			},
		},
		true,
	)
}

func (r *identityRepository) Create(ctx context.Context, identity *domain.Identity) error {
	return r.coll.Create(ctx, *identity)
}

func (r *identityRepository) FindByUserID(ctx context.Context, userID bson.ObjectID) ([]domain.Identity, error) {
	result, err := r.coll.FindMany(ctx,
		mongo.Query{Filter: map[string]any{"user_id": userID}},
		[]mongo.SortField{{Field: "created_at", Order: mongo.Ascending}},
		mongo.Page{Limit: 10},
	)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (r *identityRepository) FindByOAuth(ctx context.Context, method domain.AuthMethod, providerID string) (*domain.Identity, error) {
	identity, err := r.coll.FindOne(ctx, mongo.Query{Filter: map[string]any{
		"auth_method":      method,
		"auth_provider_id": providerID,
	}})
	if err != nil {
		if errors.Is(err, mongo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &identity, nil
}

func (r *identityRepository) UpdatePassword(ctx context.Context, userID bson.ObjectID, passwordHash string) error {
	_, err := r.coll.UpdateOne(ctx,
		mongo.Query{Filter: map[string]any{
			"user_id":     userID,
			"auth_method": domain.AuthMethodLocal,
		}},
		mongo.Update{
			Set: map[string]any{"password_hash": passwordHash},
		},
	)
	return err
}

func (r *identityRepository) RecordLogin(ctx context.Context, identityID bson.ObjectID) error {
	_, err := r.coll.UpdateByID(ctx, identityID, mongo.Update{
		Set: map[string]any{"last_login_at": time.Now()},
	})
	return err
}
