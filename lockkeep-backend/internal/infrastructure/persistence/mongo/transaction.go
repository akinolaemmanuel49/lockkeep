package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// mongoTransaction wraps a MongoDB session for explicit commit/rollback control
type mongoTransaction struct {
	session mongo.Session
	ctx     context.Context
}

// NewTransaction starts a new session
func NewTransaction(client *mongo.Client) (Transaction, error) {
	session, err := client.StartSession()
	if err != nil {
		return nil, fmt.Errorf("start session: %w", err)
	}

	// Start the transaction
	ctx := context.Background()
	if err := session.StartTransaction(); err != nil {
		session.EndSession(ctx)
		return nil, fmt.Errorf("start transaction: %w", err)
	}
	return &mongoTransaction{
		session: *session,
		ctx:     ctx,
	}, nil
}

// Commit finalizes the transaction
func (t *mongoTransaction) Commit(ctx context.Context) error {
	return t.session.CommitTransaction(ctx)
}

// Rollback aborts the transaction
func (t *mongoTransaction) Rollback(ctx context.Context) error {
	return t.session.AbortTransaction(ctx)
}

// Context returns the session context for use in operations
func (t *mongoTransaction) Context() context.Context {
	return t.ctx
}

// Close ends the session. Always defer after creation.
func (t *mongoTransaction) Close(ctx context.Context) {
	t.session.EndSession(ctx)
}
