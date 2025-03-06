package mongocc

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Connect(mongoUri string, dbName string, dataStore *MongoDataStore) error {

	client, err := mongo.Connect(options.Client().ApplyURI(mongoUri))
	if err != nil {
		return err
	}

	// Send a ping to confirm a successful connection
	var result bson.M
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Decode(&result); err != nil {
		return err
	}
	fmt.Printf("You successfully connected to MongoDB://%s\n", dbName)

	dataStore.Client = client
	dataStore.DB = client.Database(dbName)

	return nil

}

func (ds *MongoDataStore) GetColl(collectionName string) *mongo.Collection {
	return ds.DB.Collection(collectionName)
}
