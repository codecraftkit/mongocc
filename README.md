# mongocc

The `mongocc` package provides a simple and efficient way to manage multiple MongoDB connections in Go projects. This module abstracts the configuration and connection setup, allowing developers to focus on directly interacting with collections and documents.

#### Key Features:
* Enables the instantiation of multiple independent MongoDB connections.
* Simplifies connection management with configurable structures.
* Provides wrapper methods for all common MongoDB operations (Find, Insert, Update, Delete, Aggregate).
* Built-in debug mode for logging operations.
* Error classification for common MongoDB errors (not found, duplicate key, network error).

---
### Install

```bash
go get github.com/codecraftkit/mongocc
```

### Usage
Here's a practical example of how to use the `mongocc` package:
```go
package main

import (
	"context"
	"fmt"

	"github.com/codecraftkit/mongocc"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func main() {
	mq, err := mongocc.Connect("mongodb://localhost:27017", "my_db", &mongocc.ClientOptions{
		Debug: true,
	})
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	type User struct {
		ID    string `bson:"_id,omitempty"`
		Email string `bson:"email"`
		Name  string `bson:"name"`
	}

	// Insert
	_, err = mq.InsertOne(ctx, "users", User{
		ID:    "asdqwe123",
		Name:  "John Doe",
		Email: "johndoe@example.com",
	})
	if err != nil {
		panic(err)
	}

	// FindOne
	var user User
	err = mq.FindOne(ctx, "users", bson.M{"_id": "asdqwe123"}, nil).Decode(&user)
	if err != nil {
		panic(err)
	}

	fmt.Println(user)
}
```
### Several databases on one client

`Connect` opens a new `*mongo.Client` — its own connection pool — on every call.
When you need more than one database on the same server, open the client once
and derive the other handles from it:

```go
registry, err := mongocc.Connect("mongodb://localhost:27017", "registry", nil)
if err != nil {
	panic(err)
}

// No network I/O: same client, same pool, another database.
tenantA := registry.WithDatabase("platform:tenant-a")
tenantB := registry.WithDatabase("platform:tenant-b")

// Or wrap a *mongo.Client you already own.
mq := mongocc.NewFromClient(client, "billing", nil)

// Disconnect ONCE, from the owner — the derived handles share the client.
defer registry.Database().Client().Disconnect(context.Background())
```

---

#### Why Use mongocc?
Modularity: Ideal for projects requiring multiple connections to different databases.
Ease of Use: Reduces the initial complexity of setting up MongoDB connections.
Seamless Integration: Compatible with the official MongoDB driver for Go (mongo-driver/v2).

#### Best Suited For:
Developers seeking a straightforward solution to manage MongoDB connections in applications that need to efficiently and cleanly interact with multiple databases.
