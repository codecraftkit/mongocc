package mongocc

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Connect(mongoUri string, dbName string) (*MongoQueries, error) {

	client, err := mongo.Connect(options.Client().ApplyURI(mongoUri))
	if err != nil {
		return nil, err
	}

	ctx := context.TODO()

	// Send a ping to confirm a successful connection
	var result bson.M
	if err = client.Database("admin").RunCommand(ctx, bson.D{{"ping", 1}}).Decode(&result); err != nil {
		return nil, err
	}
	fmt.Printf("You successfully connected to %s db: %s\n", mongoUri, dbName)

	db := client.Database(dbName)

	mongoQueries := MongoQueries{
		db:    db,
		Debug: true,
	}

	return &mongoQueries, nil

}

type MongoQueries struct {
	db    *mongo.Database
	Debug bool
}

type MongoFunctions interface {
	FindOne(ctx context.Context, collectionName string, query interface{}, opts *options.FindOneOptionsBuilder) *mongo.SingleResult
	Find(ctx context.Context, collectionName string, query interface{}, opts *options.FindOptionsBuilder) (*mongo.Cursor, error)
	InsertOne(ctx context.Context, collectionName string, document interface{}) (*mongo.InsertOneResult, error)
	UpdateOne(ctx context.Context, collectionName string, query interface{}, update interface{}, opts *options.UpdateOneOptionsBuilder) (*mongo.UpdateResult, error)
	UpdateMany(ctx context.Context, collectionName string, query interface{}, update interface{}, opts *options.UpdateManyOptionsBuilder) (*mongo.UpdateResult, error)
	DeleteOne(ctx context.Context, collectionName string, query interface{}, opts *options.DeleteOneOptionsBuilder) (*mongo.DeleteResult, error)
	DeleteMany(ctx context.Context, collectionName string, query interface{}, opts *options.DeleteManyOptionsBuilder) (*mongo.DeleteResult, error)
	Aggregate(ctx context.Context, collectionName string, pipeline interface{}, opts *options.AggregateOptionsBuilder) (*mongo.Cursor, error)
	CountDocuments(ctx context.Context, collectionName string, query interface{}) (int64, error)
}

func (mongodb *MongoQueries) GetCollection(collectionName string) *mongo.Collection {
	return mongodb.db.Collection(collectionName)
}

func (mongodb *MongoQueries) Find(ctx context.Context, collectionName string, query interface{}, opts *options.FindOptionsBuilder) (*mongo.Cursor, error) {
	return mongodb.db.Collection(collectionName).Find(ctx, query, opts)
}

func (mongodb *MongoQueries) FindOne(ctx context.Context, collectionName string, query interface{}, opts *options.FindOneOptionsBuilder) *mongo.SingleResult {
	return mongodb.db.Collection(collectionName).FindOne(ctx, query, opts)
}

func (mongodb *MongoQueries) FindOneAndUpdate(ctx context.Context, collectionName string, query interface{}, update interface{}, opts *options.FindOneAndUpdateOptionsBuilder) *mongo.SingleResult {
	return mongodb.db.Collection(collectionName).FindOneAndUpdate(ctx, query, update, opts)
}

func (mongodb *MongoQueries) InsertOne(ctx context.Context, collectionName string, document interface{}) (*mongo.InsertOneResult, error) {
	return mongodb.db.Collection(collectionName).InsertOne(ctx, document)
}

func (mongodb *MongoQueries) InsertMany(ctx context.Context, collectionName string, documents []interface{}) (*mongo.InsertManyResult, error) {
	return mongodb.db.Collection(collectionName).InsertMany(ctx, documents)
}

func (mongodb *MongoQueries) UpdateOne(ctx context.Context, collectionName string, query interface{}, update interface{}, opts *options.UpdateOneOptionsBuilder) (*mongo.UpdateResult, error) {
	return mongodb.db.Collection(collectionName).UpdateOne(ctx, query, update, opts)
}

func (mongodb *MongoQueries) UpdateMany(ctx context.Context, collectionName string, query interface{}, update interface{}, opts *options.UpdateManyOptionsBuilder) (*mongo.UpdateResult, error) {
	return mongodb.db.Collection(collectionName).UpdateMany(ctx, query, update, opts)
}

func (mongodb *MongoQueries) DeleteOne(ctx context.Context, collectionName string, query interface{}, opts *options.DeleteOneOptionsBuilder) (*mongo.DeleteResult, error) {
	return mongodb.db.Collection(collectionName).DeleteOne(ctx, query, opts)
}

func (mongodb *MongoQueries) DeleteMany(ctx context.Context, collectionName string, query interface{}, opts *options.DeleteManyOptionsBuilder) (*mongo.DeleteResult, error) {
	return mongodb.db.Collection(collectionName).DeleteMany(ctx, query, opts)
}

func (mongodb *MongoQueries) Aggregate(ctx context.Context, collectionName string, pipeline interface{}, opts *options.AggregateOptionsBuilder) (*mongo.Cursor, error) {
	return mongodb.db.Collection(collectionName).Aggregate(ctx, pipeline, opts)
}

func (mongodb *MongoQueries) CountDocuments(ctx context.Context, collectionName string, query interface{}) (int64, error) {
	return mongodb.db.Collection(collectionName).CountDocuments(ctx, query)
}

func CheckMongoError(err error) error {
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("NOT_FOUND: %s", err.Error())
		}
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("INDEX_DUPLICATED: %s", err.Error())
		}
		if mongo.IsNetworkError(err) {
			return fmt.Errorf("NETWORK_ERROR: %s", err.Error())
		}
		return fmt.Errorf(err.Error())
	}
	return err
}
