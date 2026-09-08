package telemetry

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

func New() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func Event(ctx context.Context, logger *slog.Logger, processType, releaseID, event string, attrs ...any) {
	fields := []any{"service", "twelve-factor-orders", "processType", processType, "releaseId", releaseID, "event", event}
	span := trace.SpanContextFromContext(ctx)
	if span.IsValid() {
		fields = append(fields, "traceId", span.TraceID().String())
	}
	logger.InfoContext(ctx, event, append(fields, attrs...)...)
}
