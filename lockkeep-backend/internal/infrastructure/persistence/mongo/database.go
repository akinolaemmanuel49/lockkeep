package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// Database wraps a mongo.Database with connection lifecycle management
type Database struct {
	client *mongo.Client
	db     *mongo.Database
}

// Connect establishes a connection with retry logic
func Connect(ctx context.Context, uri string, dbName string) (*Database, error) {
	clientOpts := options.Client().ApplyURI(uri).SetServerSelectionTimeout(10 * time.Second).SetConnectTimeout(10 * time.Second)

	var client *mongo.Client
	var err error
	const maxRetries = 30

	for i := 0; i < maxRetries; i++ {
		client, err = mongo.Connect(clientOpts)
		if err == nil {
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err = client.Ping(pingCtx, readpref.Primary())
			cancel()

			if err != nil {
				fmt.Printf("ERROR, %s", err)
			}

			if err == nil {
				break
			}
			_ = client.Disconnect(ctx)
		}

		if i < maxRetries-1 {
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("unable to connect to MongoDB after %d retries: %w", maxRetries, err)
	}

	return &Database{
		client: client,
		db:     client.Database(dbName),
	}, nil
}

// Raw returns the underlying mongo.Database for repository construction
func (d *Database) Raw() *mongo.Database {
	return d.db
}

// Disconnect cleanly closes the connection
func (d *Database) Disconnect(ctx context.Context) error {
	return d.client.Disconnect(ctx)
}

// Ping verifies connectivity
func (d *Database) Ping(ctx context.Context) error {
	return d.client.Ping(ctx, readpref.Primary())
}
