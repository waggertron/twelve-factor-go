package integration

import (
	"context"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	rediscontainer "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/waggertron/twelve-factor-go/internal/domain"
	"github.com/waggertron/twelve-factor-go/internal/migrations"
	orderqueue "github.com/waggertron/twelve-factor-go/internal/queue"
	"github.com/waggertron/twelve-factor-go/internal/store"
)

func TestPostgresAndRedisWorkflow(t *testing.T) {
	ctx := context.Background()
	postgresContainer, err := postgres.Run(ctx, "postgres:18-alpine", postgres.WithDatabase("orders"), postgres.WithUsername("orders"), postgres.WithPassword("local_only"), postgres.BasicWaitStrategies())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = postgresContainer.Terminate(context.Background()) })
	databaseURL, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	pool, err := store.NewPool(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := migrations.Apply(ctx, pool, "001"); err != nil {
		t.Fatal(err)
	}
	if err := migrations.Ready(ctx, pool); err != nil {
		t.Fatal(err)
	}

	redisContainer, err := rediscontainer.Run(ctx, "redis:8-alpine")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = redisContainer.Terminate(context.Background()) })
	redisURL, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}

	orders := store.Postgres{Pool: pool}
	order, created, err := orders.Create(ctx, "integration-99", domain.OrderInput{CustomerID: "customer-001", AmountCents: 99})
	if err != nil || !created {
		t.Fatalf("create: %v, %v", created, err)
	}
	client, err := orderqueue.NewClient(redisURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	job := domain.Job{SchemaVersion: 1, OrderID: order.ID, Attempt: 1}
	if err := client.Enqueue(ctx, job); err != nil {
		t.Fatal(err)
	}
	if err := orders.MarkEnqueued(ctx, order.ID); err != nil {
		t.Fatal(err)
	}

	redisOptions, err := orderqueue.RedisOptions(redisURL)
	if err != nil {
		t.Fatal(err)
	}
	server := asynq.NewServer(redisOptions, asynq.Config{Concurrency: 1, Queues: map[string]int{"orders": 1}})
	mux := asynq.NewServeMux()
	mux.HandleFunc(orderqueue.TaskType, orderqueue.Processor{Orders: orders}.Handle)
	if err := server.Start(mux); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(server.Shutdown)

	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		stored, err := orders.Get(ctx, order.ID)
		if err == nil && stored.Status == "completed" {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	stored, err := orders.Get(ctx, order.ID)
	if err != nil || stored.Status != "completed" {
		t.Fatalf("completion: %+v, %v", stored, err)
	}
	if err := client.Enqueue(ctx, job); err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)
	stored, err = orders.Get(ctx, order.ID)
	if err != nil || stored.Status != "completed" {
		t.Fatalf("redelivery changed result: %+v, %v", stored, err)
	}
}
