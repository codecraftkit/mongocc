# mongocc - Technical Review

## What is it

Go library that provides an abstraction layer over the official MongoDB driver (`mongo-driver/v2`). It simplifies connection management and CRUD operations, enabling multiple independent connections to different databases with debug logging support.

## Architecture

**Wrapper/Facade** pattern over the official MongoDB driver. The library exposes a central struct (`MongoQueries`) that encapsulates the connection and delegates each operation to the corresponding collection.

```
                +------------------+
                |   Application    |
                +--------+---------+
                         |
                    Connect(uri, db)
                         |
                +--------v---------+
                |   MongoQueries   |
                | - db: *Database  |
                | - Debug: bool    |
                +--------+---------+
                         |
           +-------------+-------------+
           |             |             |
        Find()      InsertOne()   UpdateOne() ...
           |             |             |
           v             v             v
        +------------------------------+
        |  mongo-driver/v2 (official)  |
        +------------------------------+
```

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.25 |
| MongoDB Driver | mongo-driver/v2 v2.2.3 |
| Module Type | Library (package) |
| Repository | github.com/codecraftkit/mongocc |

## Project Structure

```
mongocc-go/
  .github/          # GitHub configuration (CI/workflows)
  mongocc.go        # Main code: connection, CRUD operations, error handling
  structs.go        # Configuration structs (ClientOptions)
  go.mod            # Module definition and dependencies
  go.sum            # Dependency checksums
  LICENSE           # Project license
  README.md         # Usage documentation
```

## Public API

### Functions

| Function | Signature | Description |
|---|---|---|
| `Connect` | `Connect(mongoUri, dbName string, opts *ClientOptions) (*MongoQueries, error)` | Establishes a MongoDB connection, verifies with ping, and returns a queries instance |
| `CheckMongoError` (deprecated) | `CheckMongoError(err error) error` | MongoDB error classification (standalone version, deprecated) |

### MongoQueries Methods

| Method | Description |
|---|---|
| `GetCollection` | Returns a collection by name |
| `Find` | Multi-document search |
| `FindOne` | Single document search |
| `FindOneAndUpdate` | Atomically finds and updates a document |
| `InsertOne` | Inserts a single document |
| `InsertMany` | Inserts multiple documents |
| `UpdateOne` | Updates a single document |
| `UpdateMany` | Updates multiple documents |
| `DeleteOne` | Deletes a single document |
| `DeleteMany` | Deletes multiple documents |
| `Aggregate` | Executes an aggregation pipeline |
| `CountDocuments` | Counts documents matching a query |
| `CheckMongoError` | Classifies MongoDB errors into categories (NOT_FOUND, INDEX_DUPLICATED, NETWORK_ERROR) |

### MongoFunctions Interface

Defines the contract for basic operations: `FindOne`, `Find`, `InsertOne`, `UpdateOne`, `UpdateMany`, `DeleteOne`, `DeleteMany`, `Aggregate`, `CountDocuments`.

**Note**: The interface does not include `FindOneAndUpdate`, `InsertMany`, or `GetCollection`, which are implemented in `MongoQueries`.

### Structs

| Struct | Fields | Description |
|---|---|---|
| `ClientOptions` | `Debug bool` | Client configuration options |
| `MongoQueries` | `db *mongo.Database`, `Debug bool` | Main instance with the connection and debug flag |

## Authentication and Authorization

Not applicable. The library delegates authentication to the MongoDB URI (the connection string may include credentials). It does not implement its own authentication or authorization.

## Configuration

| Parameter | Type | Description |
|---|---|---|
| `mongoUri` | string | MongoDB connection URI (includes credentials if applicable) |
| `dbName` | string | Database name |
| `Debug` | bool | Enables operation logging to stdout |

## Observations and Potential Improvements

### Urgent

- **Outdated README**: The README shows an old API (`MongoDataStore`, `Connect` with a different signature) that does not match the current code. This confuses consumers.
- **Deprecated function without removal plan**: `CheckMongoError` (standalone, line 161) prints a deprecation message on every call but remains exported.
- **No tests**: There are no test files (`_test.go`). Zero coverage for any operation.

### Recommended

- **Incomplete interface**: `MongoFunctions` does not include `FindOneAndUpdate`, `InsertMany`, or `GetCollection`. Consumers depending on the interface won't have access to these methods.
- **Logging with `fmt.Println`**: Debug mode uses `fmt.Println` directly. Using `log/slog` or accepting a configurable logger would be preferable.
- **Connection without timeout**: `Connect` uses `context.TODO()` for the ping. A context with timeout would prevent indefinite blocking.
- **URI exposed in logs**: Line 27 prints the full connection URI, which may contain credentials.
- **`fmt.Errorf` without wrapping**: In `CheckMongoError` (line 156/173), `fmt.Errorf(err.Error())` loses the original error chain. It should use `fmt.Errorf("...: %w", err)`.
- **Unexported `db` field**: `MongoQueries.db` is private, which is fine for encapsulation, but prevents advanced consumers from accessing the database directly (only through `GetCollection`).

## Development

```bash
# Install as dependency
go get github.com/codecraftkit/mongocc

# Run tests (when available)
go test ./...

# Verify compilation
go build ./...
```
