package mongocc

import "go.mongodb.org/mongo-driver/v2/mongo"

type MongoDataStore struct {
	DB     *mongo.Database
	Client *mongo.Client
}
