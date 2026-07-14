package mongo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrNotFound      = errors.New("record not found")
	ErrAlreadyExists = errors.New("record already exists")
	ErrConflict      = errors.New("conflict detected")
)

// ID is any type that can serve as a document identifier
type ID interface {
	bson.ObjectID | string
}

// Record is any struct that can be stored and retrieved
type Record interface {
	any
}

// Query represents a filter condition
type Query struct {
	Filter map[string]any
}

// Update represents a modification to apply
type Update struct {
	Set      map[string]any
	Unset    map[string]any
	Inc      map[string]any
	Push     map[string]any
	Pull     map[string]any
	AddToSet map[string]any
}

// SortOrder defines result ordering
type SortOrder int

const (
	Ascending  SortOrder = 1
	Descending SortOrder = -1
)

// SortField pairs a field name with its order
type SortField struct {
	Field string
	Order SortOrder
}

// Pagination bounds for list operations
type Page struct {
	Limit  int
	Offset int
}

// Result wraps a query outcome with metadata
type Result[T Record] struct {
	Data       []T
	TotalCount int64
	HasMore    bool
}

// Change captures what happened during a write
type Change struct {
	MatchedCount  int64
	ModifiedCount int64
	UpsertedID    any
}

// Collection is the business-language wrapper around MongoDB operations
type Collection[T Record, K ID] interface {
	// ─── Create ───
	Create(ctx context.Context, record T) error
	CreateMany(ctx context.Context, records []T) error

	// ─── Read ───
	FindByID(ctx context.Context, id K) (T, error)
	FindOne(ctx context.Context, query Query) (T, error)
	FindMany(ctx context.Context, query Query, sort []SortField, page Page) (Result[T], error)
	Exists(ctx context.Context, query Query) (bool, error)
	Count(ctx context.Context, query Query) (int64, error)

	// ─── Update ───
	UpdateByID(ctx context.Context, id K, update Update) (Change, error)
	UpdateOne(ctx context.Context, query Query, update Update) (Change, error)
	UpdateMany(ctx context.Context, query Query, update Update) (Change, error)
	Replace(ctx context.Context, id K, record T) (Change, error)

	// ─── Delete ───
	DeleteByID(ctx context.Context, id K) error
	DeleteOne(ctx context.Context, query Query) error
	DeleteMany(ctx context.Context, query Query) (int64, error)

	// ─── Aggregation / Advanced ───
	Aggregate(ctx context.Context, pipeline []map[string]any, into any) error
	Distinct(ctx context.Context, field string, query Query) ([]any, error)

	// ─── Schema ───
	EnsureIndex(ctx context.Context, fields []string, unique bool) error
	EnsurePartialIndex(ctx context.Context, fields []string, filter map[string]any, unique bool) error
	DropIndex(ctx context.Context, name string) error
}

// Transaction wraps multiple operations in an atomic unit
type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// UnitOfWork executes operations within a transaction
type UnitOfWork interface {
	Within(ctx context.Context, fn func(ctx context.Context) error) error
}
