package repositories

import (
	"context"
	"time"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const NAME_CRYPTO_POLICY_COLLECTION = "cryptopolicy"

type CryptoPolicyRepository struct {
	collection *mongo.Collection
}

func NewCryptoPolicyRepository(db *mongo.Database) *CryptoPolicyRepository {
	return &CryptoPolicyRepository{
		collection: db.Collection(NAME_CRYPTO_POLICY_COLLECTION),
	}
}

func (r *CryptoPolicyRepository) GetCurrent(ctx context.Context) (*domain.CryptoPolicy, error) {
	var policy domain.CryptoPolicy
	err := r.collection.FindOne(ctx, bson.M{"_id": "current"}).Decode(&policy)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &policy, nil
}

func (r *CryptoPolicyRepository) SetCurrent(ctx context.Context, policy *domain.CryptoPolicy) error {
	policy.ID = "current"
	policy.UpdatedAt = time.Now()

	_, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"_id": "current"},
		policy,
		options.Replace().SetUpsert(true),
	)
	return err
}
