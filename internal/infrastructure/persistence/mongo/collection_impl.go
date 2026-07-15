// internal/infrastructure/persistence/mongo/collection.impl.go

package mongo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// mongoCollection implements Collection using the MongoDB driver
type mongoCollection[T Record, K ID] struct {
	coll *mongo.Collection
	typ  reflect.Type // cached for zero-value creation
}

// NewCollection creates a business-language collection backed by MongoDB
func NewCollection[T Record, K ID](db *mongo.Database, name string) Collection[T, K] {
	var zero T
	return &mongoCollection[T, K]{
		coll: db.Collection(name),
		typ:  reflect.TypeOf(zero),
	}
}

// ─── Helpers ───

func (m *mongoCollection[T, K]) zeroValue() T {
	var v T
	return v
}

func toBSONM(update Update) bson.M {
	set := make(bson.M)

	if len(update.Set) > 0 {
		set["$set"] = update.Set
	}
	if len(update.Unset) > 0 {
		set["$unset"] = update.Unset
	}
	if len(update.Inc) > 0 {
		set["$inc"] = update.Inc
	}
	if len(update.Push) > 0 {
		set["$push"] = update.Push
	}
	if len(update.Pull) > 0 {
		set["$pull"] = update.Pull
	}
	if len(update.AddToSet) > 0 {
		set["$addToSet"] = update.AddToSet
	}

	// Auto-inject updated_at if $set exists
	if _, ok := set["$set"]; ok {
		set["$set"] = injectTimestamp(set["$set"].(map[string]any), "updated_at")
	}

	return set
}

func injectTimestamp(m map[string]any, field string) map[string]any {
	m[field] = time.Now()
	return m
}

func toSort(sort []SortField) bson.D {
	d := make(bson.D, len(sort))
	for i, s := range sort {
		d[i] = bson.E{Key: s.Field, Value: int(s.Order)}
	}
	return d
}

func toFindOptions(sort []SortField, page Page) *options.FindOptionsBuilder {
	opts := options.Find()
	if len(sort) > 0 {
		opts.SetSort(toSort(sort))
	}
	if page.Limit > 0 {
		opts.SetLimit(int64(page.Limit))
	}
	if page.Offset > 0 {
		opts.SetSkip(int64(page.Offset))
	}
	return opts
}

func idFilter(id any) bson.M {
	switch v := id.(type) {
	case bson.ObjectID:
		return bson.M{"_id": v}
	case string:
		oid, err := bson.ObjectIDFromHex(v)
		if err == nil {
			return bson.M{"_id": oid}
		}
		return bson.M{"_id": v}
	default:
		return bson.M{"_id": id}
	}
}

// Helper to get the appropriate context for operations
func (m *mongoCollection[T, K]) getCollection(ctx context.Context) *mongo.Collection {
	// If the context has a session, use it
	if session := mongo.SessionFromContext(ctx); session != nil {
		// Get the client from the session and create a new collection with the session
		client := session.Client()
		// Use the session's client to get a collection with session context
		return client.Database(m.coll.Database().Name()).Collection(m.coll.Name())
	}
	return m.coll
}

// ─── Create ───

func (m *mongoCollection[T, K]) Create(ctx context.Context, record T) error {
	// Inject timestamps if the struct has them
	record = m.injectTimestamps(record, true)
	_, err := m.getCollection(ctx).InsertOne(ctx, record)
	if mongo.IsDuplicateKeyError(err) {
		return fmt.Errorf("%w: %v", ErrAlreadyExists, err)
	}
	return err
}

func (m *mongoCollection[T, K]) CreateMany(ctx context.Context, records []T) error {
	if len(records) == 0 {
		return nil
	}

	docs := make([]any, len(records))
	for i, r := range records {
		docs[i] = m.injectTimestamps(r, true)
	}

	_, err := m.getCollection(ctx).InsertMany(ctx, docs)
	if mongo.IsDuplicateKeyError(err) {
		return fmt.Errorf("%w: %v", ErrAlreadyExists, err)
	}
	return err
}

// ─── Read ───

func (m *mongoCollection[T, K]) FindByID(ctx context.Context, id K) (T, error) {
	var result T
	err := m.getCollection(ctx).FindOne(ctx, idFilter(id)).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return m.zeroValue(), ErrNotFound
		}
		return m.zeroValue(), err
	}
	return result, nil
}

func (m *mongoCollection[T, K]) FindOne(ctx context.Context, query Query) (T, error) {
	var result T
	err := m.getCollection(ctx).FindOne(ctx, query.Filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return m.zeroValue(), ErrNotFound
		}
		return m.zeroValue(), err
	}
	return result, nil
}

func (m *mongoCollection[T, K]) FindMany(ctx context.Context, query Query, sort []SortField, page Page) (Result[T], error) {
	cursor, err := m.getCollection(ctx).Find(ctx, query.Filter, toFindOptions(sort, page))
	if err != nil {
		return Result[T]{}, err
	}
	defer cursor.Close(ctx)

	var items []T
	if err := cursor.All(ctx, &items); err != nil {
		return Result[T]{}, err
	}

	total, err := m.getCollection(ctx).CountDocuments(ctx, query.Filter)
	if err != nil {
		return Result[T]{}, err
	}

	return Result[T]{
		Data:       items,
		TotalCount: total,
		HasMore:    page.Limit > 0 && int64(len(items))+int64(page.Offset) < total,
	}, nil
}

func (m *mongoCollection[T, K]) Exists(ctx context.Context, query Query) (bool, error) {
	count, err := m.getCollection(ctx).CountDocuments(ctx, query.Filter, options.Count().SetLimit(1))
	return count > 0, err
}

func (m *mongoCollection[T, K]) Count(ctx context.Context, query Query) (int64, error) {
	return m.getCollection(ctx).CountDocuments(ctx, query.Filter)
}

// ─── Update ───

func (m *mongoCollection[T, K]) UpdateByID(ctx context.Context, id K, update Update) (Change, error) {
	result, err := m.getCollection(ctx).UpdateByID(ctx, id, toBSONM(update))
	if err != nil {
		return Change{}, err
	}
	if result.MatchedCount == 0 {
		return Change{}, ErrNotFound
	}
	return Change{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
	}, nil
}

func (m *mongoCollection[T, K]) UpdateOne(ctx context.Context, query Query, update Update) (Change, error) {
	result, err := m.getCollection(ctx).UpdateOne(ctx, query.Filter, toBSONM(update))
	if err != nil {
		return Change{}, err
	}
	if result.MatchedCount == 0 {
		return Change{}, ErrNotFound
	}
	return Change{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
	}, nil
}

func (m *mongoCollection[T, K]) UpdateMany(ctx context.Context, query Query, update Update) (Change, error) {
	result, err := m.getCollection(ctx).UpdateMany(ctx, query.Filter, toBSONM(update))
	if err != nil {
		return Change{}, err
	}
	return Change{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
	}, nil
}

func (m *mongoCollection[T, K]) Replace(ctx context.Context, id K, record T) (Change, error) {
	record = m.injectTimestamps(record, false)
	result, err := m.getCollection(ctx).ReplaceOne(ctx, idFilter(id), record)
	if err != nil {
		return Change{}, err
	}
	if result.MatchedCount == 0 {
		return Change{}, ErrNotFound
	}
	return Change{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		UpsertedID:    result.UpsertedID,
	}, nil
}

// ─── Delete ───

func (m *mongoCollection[T, K]) DeleteByID(ctx context.Context, id K) error {
	result, err := m.getCollection(ctx).DeleteOne(ctx, idFilter(id))
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (m *mongoCollection[T, K]) DeleteOne(ctx context.Context, query Query) error {
	result, err := m.getCollection(ctx).DeleteOne(ctx, query.Filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (m *mongoCollection[T, K]) DeleteMany(ctx context.Context, query Query) (int64, error) {
	result, err := m.getCollection(ctx).DeleteMany(ctx, query.Filter)
	return result.DeletedCount, err
}

// ─── Aggregation ───

func (m *mongoCollection[T, K]) Aggregate(ctx context.Context, pipeline []map[string]any, into any) error {
	bsonPipeline := make(mongo.Pipeline, len(pipeline))
	for i, stage := range pipeline {
		bsonPipeline[i] = bson.D{}
		for k, v := range stage {
			bsonPipeline[i] = append(bsonPipeline[i], bson.E{Key: k, Value: v})
		}
	}
	cursor, err := m.getCollection(ctx).Aggregate(ctx, bsonPipeline)
	if err != nil {
		return err
	}
	return cursor.All(ctx, into)
}

// ─── Distinct ───

func (m *mongoCollection[T, K]) Distinct(ctx context.Context, field string, query Query) ([]any, error) {
	result := m.getCollection(ctx).Distinct(ctx, field, query.Filter)

	var values []any
	if err := result.Decode(&values); err != nil {
		return nil, err
	}
	return values, nil
}

// ─── Schema ───

func (m *mongoCollection[T, K]) EnsureIndex(ctx context.Context, fields []string, unique bool) error {
	keys := make(bson.D, len(fields))
	for i, f := range fields {
		keys[i] = bson.E{Key: f, Value: 1}
	}
	_, err := m.getCollection(ctx).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    keys,
		Options: options.Index().SetUnique(unique),
	})
	return err
}

func (m *mongoCollection[T, K]) EnsurePartialIndex(ctx context.Context, fields []string, filter map[string]any, unique bool) error {
	keys := make(bson.D, len(fields))
	for i, f := range fields {
		keys[i] = bson.E{Key: f, Value: 1}
	}
	_, err := m.getCollection(ctx).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    keys,
		Options: options.Index().SetUnique(unique).SetPartialFilterExpression(filter),
	})
	return err
}

func (m *mongoCollection[T, K]) DropIndex(ctx context.Context, name string) error {
	return m.getCollection(ctx).Indexes().DropOne(ctx, name)
}

// ─── Timestamp injection via reflection ───

func (m *mongoCollection[T, K]) injectTimestamps(record T, isNew bool) T {
	// Use reflection to set CreatedAt/UpdatedAt if fields exist
	// This avoids requiring an interface on T
	v := reflect.ValueOf(&record).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		bsonTag := field.Tag.Get("bson")
		name := strings.Split(bsonTag, ",")[0]

		if name == "created_at" && isNew {
			if v.Field(i).IsZero() {
				v.Field(i).Set(reflect.ValueOf(time.Now()))
			}
		}
		if name == "updated_at" {
			v.Field(i).Set(reflect.ValueOf(time.Now()))
		}
	}

	return record
}
