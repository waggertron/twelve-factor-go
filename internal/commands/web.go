package commands

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/waggertron/twelve-factor-go/internal/config"
	"github.com/waggertron/twelve-factor-go/internal/httpapi"
	"github.com/waggertron/twelve-factor-go/internal/migrations"
	"github.com/waggertron/twelve-factor-go/internal/queue"
	"github.com/waggertron/twelve-factor-go/internal/readiness"
	"github.com/waggertron/twelve-factor-go/internal/store"
	"github.com/waggertron/twelve-factor-go/internal/telemetry"
)

func webCommand(logger *slog.Logger) *cobra.Command {
	return &cobra.Command{Use: "web", Short: "Run the HTTP process", RunE: func(_ *cobra.Command, _ []string) error {
		return runWeb(logger)
	}}
}

func runWeb(logger *slog.Logger) error {
	cfg, err := config.Load("web")
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
	queueClient, err := queue.NewClient(cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("queue unavailable")
	}
	defer queueClient.Close()
	state := &readiness.State{}
	api := httpapi.API{Orders: store.Postgres{Pool: pool}, Queue: queueClient, Readiness: state}
	server := &http.Server{Addr: fmt.Sprintf(":%d", cfg.Port), Handler: api.Router(), ReadHeaderTimeout: 5 * time.Second}
	errorsCh := make(chan error, 1)
	go func() { errorsCh <- server.ListenAndServe() }()
	state.MarkReady()
	telemetry.Event(ctx, logger, "web", cfg.ReleaseID, "web.ready", "port", cfg.Port)
	select {
	case <-ctx.Done():
	case err := <-errorsCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}
	state.BeginDrain()
	telemetry.Event(context.Background(), logger, "web", cfg.ReleaseID, "web.draining")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ShutdownGraceMS)*time.Millisecond)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("web shutdown deadline: %w", err)
	}
	return nil
}
