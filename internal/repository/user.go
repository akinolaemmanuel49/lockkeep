package repository

import (
	"context"
	"errors"
	"time"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrVaultAlreadyMigrated = errors.New("vault already migrated")

const NAME_USER_COLLECTION = "users"

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection(NAME_USER_COLLECTION),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	user.ID = bson.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, user)
	return err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"email": email})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return true, nil
		}
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByOAuth(ctx context.Context, provider, providerID string) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{
		"auth_method":      "oauth_" + provider,
		"auth_provider_id": providerID,
	}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateVaultMetadata(ctx context.Context, userID bson.ObjectID, vault domain.VaultMetadata) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": userID},
		bson.M{
			"$set": bson.M{
				"vault":      vault,
				"updated_at": time.Now(),
			},
		},
	)
	return err
}

func (r *UserRepository) UpdateEmail(ctx context.Context, userID bson.ObjectID, email string) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": userID},
		bson.M{
			"$set": bson.M{
				"email":      email,
				"updated_at": time.Now(),
			},
		},
	)
	return err
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID bson.ObjectID, passwordHash string) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": userID},
		bson.M{
			"$set": bson.M{
				"password_hash": passwordHash,
				"updated_at":    time.Now(),
			},
		},
	)
	return err
}

func (r *UserRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	_, err = r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "auth_method", Value: 1},
			{Key: "auth_provider_id", Value: 1},
		},
		Options: options.Index().
			SetUnique(true).
			SetPartialFilterExpression(
				bson.M{
					"auth_method": bson.M{
						"$in": []string{"oauth_google", "oauth_github"},
					},
				},
			),
	})
	return err
}

func (r *UserRepository) UpdateKDF(
	ctx context.Context,
	userID bson.ObjectID,
	expectedVersion uint32,
	verificationHash string,
	kdfParams domain.KDFParams,
) error {

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{
			"_id":           userID,
			"vault.version": expectedVersion,
		},
		bson.M{
			"$set": bson.M{
				"vault.verification_hash": verificationHash,
				"vault.kdf.algorithm":     kdfParams.Algorithm,
				"vault.kdf.salt":          kdfParams.Salt,
				"vault.kdf.memory":        kdfParams.Memory,
				"vault.kdf.iterations":    kdfParams.Iterations,
				"vault.kdf.parallelism":   kdfParams.Parallelism,
				"updated_at":              time.Now(),
			},
			"$inc": bson.M{
				"vault.version": 1,
			},
		},
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrVaultAlreadyMigrated
	}

	return nil
}
