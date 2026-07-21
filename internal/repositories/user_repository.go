package repositories

import (
	"context"
	"errors"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/infrastructure/persistence/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	driver "go.mongodb.org/mongo-driver/v2/mongo"
)

type userRepository struct {
	coll mongo.Collection[domain.User, bson.ObjectID]
}

func NewUserRepository(db *driver.Database) *userRepository {
	return &userRepository{
		coll: mongo.NewCollection[domain.User, bson.ObjectID](db, "users"),
	}
}

func (r *userRepository) EnsureIndexes(ctx context.Context) error {
	if err := r.coll.EnsureIndex(ctx, []string{"email"}, true); err != nil {
		return err
	}
	return r.coll.EnsureIndex(ctx, []string{"username"}, true)
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	return r.coll.Create(ctx, *user)
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := r.coll.FindOne(ctx, mongo.Query{Filter: map[string]any{"email": email}})
	if err != nil {
		if errors.Is(err, mongo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	return r.coll.Exists(ctx, mongo.Query{Filter: map[string]any{"email": email}})
}

func (r *userRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.User, error) {
	user, err := r.coll.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UpdateVaultMetadata(ctx context.Context, userID bson.ObjectID, vault domain.VaultMetadata) error {
	_, err := r.coll.UpdateByID(ctx, userID, mongo.Update{
		Set: map[string]any{"vault": vault},
	})
	return err
}

func (r *userRepository) UpdateEmail(ctx context.Context, userID bson.ObjectID, email string) error {
	_, err := r.coll.UpdateByID(ctx, userID, mongo.Update{
		Set: map[string]any{"email": email},
	})
	return err
}
