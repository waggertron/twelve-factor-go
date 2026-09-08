package commands

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/spf13/cobra"
	"github.com/waggertron/twelve-factor-go/internal/config"
	"github.com/waggertron/twelve-factor-go/internal/domain"
	"github.com/waggertron/twelve-factor-go/internal/migrations"
	orderqueue "github.com/waggertron/twelve-factor-go/internal/queue"
	"github.com/waggertron/twelve-factor-go/internal/store"
	"github.com/waggertron/twelve-factor-go/internal/telemetry"
)

func workerCommand(logger *slog.Logger) *cobra.Command {
	return &cobra.Command{Use: "worker", Short: "Run the background worker", RunE: func(_ *cobra.Command, _ []string) error {
		return runWorker(logger)
	}}
}

type workerServer interface {
	Stop()
	Shutdown()
}

func shutdownWorker(server workerServer, grace time.Duration) error {
	server.Stop()
	done := make(chan struct{})
	go func() {
		server.Shutdown()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-time.After(grace):
		return fmt.Errorf("worker shutdown deadline exceeded")
	}
}

func runWorker(logger *slog.Logger) error {
	cfg, err := config.Load("worker")
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	pool, err := store.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("database unavailable")
	}
	defer pool.Close()
	if err := migrations.Ready(ctx, pool); err != nil {
		return err
	}
	redisOptions, err := orderqueue.RedisOptions(cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("queue unavailable")
	}
	server := asynq.NewServer(redisOptions, asynq.Config{Concurrency: cfg.WorkerConcurrency, Queues: map[string]int{"orders": 1}})
	mux := asynq.NewServeMux()
	processor := orderqueue.Processor{
		Orders: store.Postgres{Pool: pool},
		OnResult: func(ctx context.Context, job domain.Job, result string) {
			retryCount, _ := asynq.GetRetryCount(ctx)
			telemetry.Event(ctx, logger, "worker", cfg.ReleaseID, "order."+result, "orderId", job.OrderID, "attempt", retryCount+1)
		},
	}
	mux.HandleFunc(orderqueue.TaskType, processor.Handle)
	if err := server.Start(mux); err != nil {
		return err
	}
	telemetry.Event(ctx, logger, "worker", cfg.ReleaseID, "worker.ready", "concurrency", cfg.WorkerConcurrency)
	<-ctx.Done()
	telemetry.Event(context.Background(), logger, "worker", cfg.ReleaseID, "worker.draining")
	if err := shutdownWorker(server, time.Duration(cfg.ShutdownGraceMS)*time.Millisecond); err != nil {
		return err
	}
	telemetry.Event(context.Background(), logger, "worker", cfg.ReleaseID, "worker.stopped")
	return nil
}
