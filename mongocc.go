package mongocc

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// redactedURI is what Connect logs when the connection string cannot be parsed
// into a scheme and a host. An unparseable string may still carry credentials,
// so it is never echoed back.
const redactedURI = "[redacted]"

// redactMongoURI keeps only the part of a connection string that the log line
// is actually for — which server the client attached to — and drops everything
// else.
//
// It rebuilds the URI from scheme and host rather than removing the parts that
// are known to hold secrets today. Userinfo is the obvious one, but it is not
// the only one: the option string carries tlsCertificateKeyFilePassword, and
// the AWS session token inside authMechanismProperties. Rebuilding means an
// option added to the connection string later cannot leak by omission.
func redactMongoURI(mongoUri string) string {
	parsed, err := url.Parse(mongoUri)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return redactedURI
	}
	return parsed.Scheme + "://" + parsed.Host
}

// connectionLogLine builds the line Connect prints once the ping succeeds.
//
// It is a separate function so the redaction can be tested without a live
// MongoDB. A test that only covered redactMongoURI would not notice a later
// edit that formats the raw URI into this message again, which is exactly how
// the credentials got here in the first place.
func connectionLogLine(mongoUri string, dbName string) string {
	return fmt.Sprintf("You successfully connected to Mongo: %s db: %s", redactMongoURI(mongoUri), dbName)
}

func Connect(mongoUri string, dbName string, opts *ClientOptions) (*MongoQueries, error) {

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
	fmt.Println(connectionLogLine(mongoUri, dbName))

	return NewFromClient(client, dbName, opts), nil
}

// NewFromClient wraps a client the caller already owns.
//
// It is the constructor for the second and every later handle on one client.
// Connect opens a *mongo.Client on every call — its own connection pool, its
// own monitoring goroutines — so a service that needs N databases on the same
// server and calls Connect N times is running N pools against one deployment.
// The driver's client.Database(name) is a handle with no network I/O, and this
// is how a MongoQueries gets built on top of one.
//
// No ping and no log line: the client was already verified by whoever opened
// it. A nil opts means Debug off, same as Connect.
func NewFromClient(client *mongo.Client, dbName string, opts *ClientOptions) *MongoQueries {
	if opts == nil {
		opts = &ClientOptions{}
	}

	return &MongoQueries{
		db:    client.Database(dbName),
		Debug: opts.Debug,
	}
}

// WithDatabase returns a handle to another database on the SAME client.
//
// The new handle shares the receiver's connection pool and inherits its Debug
// flag; the receiver is left untouched. This is the multi-tenant shape — one
// server, one database per tenant — without one pool per tenant.
//
// Because the client is shared, a handle built here must never disconnect it:
// that would take every other handle down with it.
func (mongodb *MongoQueries) WithDatabase(dbName string) *MongoQueries {
	return &MongoQueries{
		db:    mongodb.db.Client().Database(dbName),
		Debug: mongodb.Debug,
	}
}

// Database exposes the handle this MongoQueries operates on. Name() says which
// database, and Client() is the way to Disconnect at shutdown — by the owner
// of the client, once, not by every handle derived from it.
func (mongodb *MongoQueries) Database() *mongo.Database {
	return mongodb.db
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
	if mongodb.Debug {
		fmt.Println("[Mongocc Log] Find", collectionName, query, opts)
	}
	return mongodb.db.Collection(collectionName).Find(ctx, query, opts)
}

func (mongodb *MongoQueries) FindOne(ctx context.Context, collectionName string, query interface{}, opts *options.FindOneOptionsBuilder) *mongo.SingleResult {
	if mongodb.Debug {
		fmt.Println("[Mongocc Log] FindOne", collectionName, query, opts)
	}
	return mongodb.db.Collection(collectionName).FindOne(ctx, query, opts)
}

func (mongodb *MongoQueries) FindOneAndUpdate(ctx context.Context, collectionName string, query interface{}, update interface{}, opts *options.FindOneAndUpdateOptionsBuilder) *mongo.SingleResult {
	if mongodb.Debug {
		fmt.Println("[Mongocc Log] FindOneAndUpdate", collectionName, query, update, opts)
	}
	return mongodb.db.Collection(collectionName).FindOneAndUpdate(ctx, query, update, opts)
}

func (mongodb *MongoQueries) InsertOne(ctx context.Context, collectionName string, document interface{}) (*mongo.InsertOneResult, error) {
	if mongodb.Debug {
		fmt.Println("[Mongocc Log] InsertOne", collectionName, document)
	}
	return mongodb.db.Collection(collectionName).InsertOne(ctx, document)
}

func (mongodb *MongoQueries) InsertMany(ctx context.Context, collectionName string, documents []interface{}) (*mongo.InsertManyResult, error) {
	if mongodb.Debug {
		fmt.Println("[Mongocc Log] InsertMany", collectionName, documents)
	}
	return mongodb.db.Collection(collectionName).InsertMany(ctx, documents)
}

func (mongodb *MongoQueries) UpdateOne(ctx context.Context, collectionName string, query interface{}, update interface{}, opts *options.UpdateOneOptionsBuilder) (*mongo.UpdateResult, error) {
	if mongodb.Debug {
		fmt.Println("[Mongocc Log] UpdateOne", collectionName, query, update, opts)
	}
	return mongodb.db.Collection(collectionName).UpdateOne(ctx, query, update, opts)
}

func (mongodb *MongoQueries) UpdateMany(ctx context.Context, collectionName string, query interface{}, update interface{}, opts *options.UpdateManyOptionsBuilder) (*mongo.UpdateResult, error) {
	if mongodb.Debug {
		fmt.Println("[Mongocc Log] UpdateMany", collectionName, query, update, opts)
	}
	return mongodb.db.Collection(collectionName).UpdateMany(ctx, query, update, opts)
}

func (mongodb *MongoQueries) DeleteOne(ctx context.Context, collectionName string, query interface{}, opts *options.DeleteOneOptionsBuilder) (*mongo.DeleteResult, error) {
	if mongodb.Debug {
		fmt.Println("[Mongocc Log] DeleteOne", collectionName, query, opts)
	}
	return mongodb.db.Collection(collectionName).DeleteOne(ctx, query, opts)
}

func (mongodb *MongoQueries) DeleteMany(ctx context.Context, collectionName string, query interface{}, opts *options.DeleteManyOptionsBuilder) (*mongo.DeleteResult, error) {
	if mongodb.Debug {
		fmt.Println("[Mongocc Log] DeleteMany", collectionName, query, opts)
	}
	return mongodb.db.Collection(collectionName).DeleteMany(ctx, query, opts)
}

func (mongodb *MongoQueries) Aggregate(ctx context.Context, collectionName string, pipeline interface{}, opts *options.AggregateOptionsBuilder) (*mongo.Cursor, error) {
	if mongodb.Debug {
		fmt.Println("[Mongocc Log] Aggregate", collectionName, pipeline, opts)
	}
	return mongodb.db.Collection(collectionName).Aggregate(ctx, pipeline, opts)
}

func (mongodb *MongoQueries) CountDocuments(ctx context.Context, collectionName string, query interface{}) (int64, error) {
	if mongodb.Debug {
		fmt.Println("[Mongocc Log] CountDocuments", collectionName, query)
	}
	return mongodb.db.Collection(collectionName).CountDocuments(ctx, query)
}

func (mongodb *MongoQueries) CheckMongoError(err error) error {
	if err != nil {
		if mongodb.Debug {
			fmt.Println("[Mongocc Log] CheckMongoError", err)
		}
		if errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("NOT_FOUND: %s", err.Error())
		}
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("INDEX_DUPLICATED: %s", err.Error())
		}
		if mongo.IsNetworkError(err) {
			return fmt.Errorf("NETWORK_ERROR: %s", err.Error())
		}
		return errors.New(err.Error())
	}
	return err
}

func CheckMongoError(err error) error {
	fmt.Println("func CheckMongoError is deprecated")
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
		return errors.New(err.Error())
	}
	return err
}
