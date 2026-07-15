package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// mongoUnitOfWork executes operations within a MongoDB transaction
type mongoUnitOfWork struct {
	client *mongo.Client
}

// NewUnitOfWork creates a UnitOfWork backed by MongoDB sessions
func NewUnitOfWork(client *mongo.Client) UnitOfWork {
	return &mongoUnitOfWork{client: client}
}

// Within runs fn inside a transaction. Automatically commits on nil error, rolls back on error.
func (u *mongoUnitOfWork) Within(ctx context.Context, fn func(ctx context.Context) error) error {
	session, err := u.client.StartSession()
	if err != nil {
		return fmt.Errorf("start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx context.Context) (any, error) {
		return nil, fn(sessCtx)
	})

	return err
}
