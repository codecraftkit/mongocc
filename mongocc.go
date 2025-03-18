package mongocc

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Connect(mongoUri string, dbName string) (*mongo.Database, error) {

	client, err := mongo.Connect(options.Client().ApplyURI(mongoUri))
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	// Send a ping to confirm a successful connection
	var result bson.M
	if err := client.Database("admin").RunCommand(ctx, bson.D{{"ping", 1}}).Decode(&result); err != nil {
		return nil, err
	}
	fmt.Printf("You successfully connected to MongoDB://%s\n", dbName)

	return client.Database(dbName), nil

}
