package commands

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/waggertron/twelve-factor-go/internal/config"
	"github.com/waggertron/twelve-factor-go/internal/migrations"
	"github.com/waggertron/twelve-factor-go/internal/store"
	"github.com/waggertron/twelve-factor-go/internal/telemetry"
)

func adminCommand(logger *slog.Logger) *cobra.Command {
	admin := &cobra.Command{Use: "admin", Short: "Run one-off administrative processes"}
	var target string
	migrate := &cobra.Command{Use: "migrate", Short: "Apply a bounded schema migration", RunE: func(cmd *cobra.Command, _ []string) error {
		if target != "001" {
			return fmt.Errorf("unsupported migration target")
		}
		cfg, err := config.Load("admin")
		if err != nil {
			return err
		}
		ctx := context.Background()
		pool, err := store.NewPool(ctx, cfg.DatabaseURL)
		if err != nil {
			return fmt.Errorf("database unavailable")
		}
		defer pool.Close()
		telemetry.Event(ctx, logger, "admin", cfg.ReleaseID, "admin.migration_started", "target", target)
		if err := migrations.Apply(ctx, pool, target); err != nil {
			return err
		}
		telemetry.Event(ctx, logger, "admin", cfg.ReleaseID, "admin.migration_finished", "target", target)
		return nil
	}}
	migrate.Flags().StringVar(&target, "target", "", "migration target")
	_ = migrate.MarkFlagRequired("target")
	admin.AddCommand(migrate)
	return admin
}
