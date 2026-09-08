package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func New() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func Configure(mode string) (func(context.Context) error, error) {
	serviceResource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("twelve-factor-orders"),
	)
	var provider *sdktrace.TracerProvider
	switch mode {
	case "console":
		exporter, err := stdouttrace.New(stdouttrace.WithWriter(os.Stdout))
		if err != nil {
			return nil, fmt.Errorf("create console trace exporter: %w", err)
		}
		provider = sdktrace.NewTracerProvider(sdktrace.WithResource(serviceResource), sdktrace.WithBatcher(exporter))
	case "memory":
		exporter := tracetest.NewInMemoryExporter()
		provider = sdktrace.NewTracerProvider(sdktrace.WithResource(serviceResource), sdktrace.WithSyncer(exporter))
	case "disabled":
		provider = sdktrace.NewTracerProvider(sdktrace.WithResource(serviceResource))
	default:
		return nil, fmt.Errorf("unsupported telemetry mode")
	}
	otel.SetTracerProvider(provider)
	return provider.Shutdown, nil
}

func Event(ctx context.Context, logger *slog.Logger, processType, releaseID, event string, attrs ...any) {
	fields := []any{"service", "twelve-factor-orders", "processType", processType, "releaseId", releaseID, "event", event}
	span := oteltrace.SpanContextFromContext(ctx)
	if span.IsValid() {
		fields = append(fields, "traceId", span.TraceID().String())
	}
	logger.InfoContext(ctx, event, append(fields, attrs...)...)
}
