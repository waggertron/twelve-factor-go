package commands

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/waggertron/twelve-factor-go/internal/telemetry"
)

func New(logger *slog.Logger) *cobra.Command {
	root := &cobra.Command{
		Use:           "orders",
		Short:         "Twelve-Factor order service",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(webCommand(logger), workerCommand(logger), adminCommand(logger))
	return root
}

func Execute() error {
	logger := telemetry.New()
	if err := New(logger).Execute(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}
