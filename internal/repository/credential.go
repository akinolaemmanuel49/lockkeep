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

type CredentialRepository struct {
	collection *mongo.Collection
}

func NewCredentialRepository(db *mongo.Database) *CredentialRepository {
	return &CredentialRepository{
		collection: db.Collection("credentials"),
	}
}

func (r *CredentialRepository) Create(ctx context.Context, cred *domain.Credential) error {
	cred.ID = bson.NewObjectID()
	cred.CreatedAt = time.Now()
	cred.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, cred)
	return err
}

func (r *CredentialRepository) FindByUser(ctx context.Context, userID bson.ObjectID, tenantID string) ([]domain.Credential, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"user_id":   userID,
		"tenant_id": tenantID,
	}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var creds []domain.Credential
	if err := cursor.All(ctx, &creds); err != nil {
		return nil, err
	}
	return creds, nil
}

func (r *CredentialRepository) FindByID(ctx context.Context, id bson.ObjectID, userID bson.ObjectID, tenantID string) (*domain.Credential, error) {
	var cred domain.Credential
	err := r.collection.FindOne(ctx, bson.M{
		"_id":       id,
		"user_id":   userID,
		"tenant_id": tenantID,
	}).Decode(&cred)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &cred, nil
}

func (r *CredentialRepository) Update(ctx context.Context, id bson.ObjectID, userID bson.ObjectID, tenantID string, updates bson.M) error {
	updates["updated_at"] = time.Now()
	_, err := r.collection.UpdateOne(ctx,
		bson.M{
			"_id":       id,
			"user_id":   userID,
			"tenant_id": tenantID,
		},
		bson.M{"$set": updates},
	)
	return err
}

func (r *CredentialRepository) Delete(ctx context.Context, id bson.ObjectID, userID bson.ObjectID, tenantID string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{
		"_id":       id,
		"user_id":   userID,
		"tenant_id": tenantID,
	})
	return err
}

func (r *CredentialRepository) FindDuplicate(ctx context.Context, userID bson.ObjectID, tenantID string, org, identifier string, excludeID *bson.ObjectID) (bool, error) {
	filter := bson.M{
		"user_id":      userID,
		"tenant_id":    tenantID,
		"organization": org,
		"identifier":   identifier,
	}
	if excludeID != nil {
		filter["_id"] = bson.M{"$ne": *excludeID}
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *CredentialRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "tenant_id", Value: 1},
			{Key: "organization", Value: 1},
			{Key: "identifier", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	return err
}
