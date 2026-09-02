package mongocc

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// offlineClient builds a *mongo.Client that never reaches a server.
//
// mongo.Connect in driver v2 does not dial: it starts the topology monitor in
// the background and returns, so a URI pointing at a closed port yields a
// usable client whose Database() handles can be inspected. Nothing here runs
// an operation against it — the assertions are about which client a handle
// points to and which database it names, and both are answered locally.
func offlineClient(t *testing.T) *mongo.Client {
	t.Helper()

	client, err := mongo.Connect(options.Client().ApplyURI(
		"mongodb://127.0.0.1:1/?serverSelectionTimeoutMS=100&connectTimeoutMS=100",
	))
	if err != nil {
		t.Fatalf("mongo.Connect: %v", err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = client.Disconnect(ctx)
	})

	return client
}

func TestNewFromClientBuildsAHandleOnTheGivenClient(t *testing.T) {
	client := offlineClient(t)

	mq := NewFromClient(client, "billing", &ClientOptions{Debug: true})

	if got := mq.Database().Name(); got != "billing" {
		t.Errorf("Database().Name() = %q, want %q", got, "billing")
	}
	if mq.Database().Client() != client {
		t.Error("the handle is not on the client it was given")
	}
	if !mq.Debug {
		t.Error("Debug was not taken from the options")
	}
}

func TestNewFromClientNilOptionsMeansDebugOff(t *testing.T) {
	mq := NewFromClient(offlineClient(t), "billing", nil)

	if mq.Debug {
		t.Error("nil options must mean Debug off, as in Connect")
	}
}

// TestWithDatabaseSharesTheClient is the reason these methods exist.
//
// A consumer that needs one database per tenant used to call Connect once per
// tenant, and each call opened a client with its own pool. The property that
// stops that is pointer equality on the client: every handle derived from one
// base must report the same *mongo.Client.
func TestWithDatabaseSharesTheClient(t *testing.T) {
	client := offlineClient(t)
	base := NewFromClient(client, "registry", nil)

	tenants := []string{"platform:dotribe", "factoring:nukocapital", "registry"}
	for _, name := range tenants {
		handle := base.WithDatabase(name)

		if got := handle.Database().Name(); got != name {
			t.Errorf("WithDatabase(%q).Database().Name() = %q", name, got)
		}
		if handle.Database().Client() != client {
			t.Errorf("WithDatabase(%q) is on a different client — that is a second pool", name)
		}
	}

	if got := base.Database().Name(); got != "registry" {
		t.Errorf("the receiver was modified: Database().Name() = %q, want %q", got, "registry")
	}
}

func TestWithDatabaseInheritsDebug(t *testing.T) {
	client := offlineClient(t)

	on := NewFromClient(client, "registry", &ClientOptions{Debug: true}).WithDatabase("tenant")
	if !on.Debug {
		t.Error("Debug on the base must carry over to the derived handle")
	}

	off := NewFromClient(client, "registry", nil).WithDatabase("tenant")
	if off.Debug {
		t.Error("Debug off on the base must stay off on the derived handle")
	}
}
