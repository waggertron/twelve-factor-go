package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestConfigureSupportedModes(t *testing.T) {
	for _, mode := range []string{"console", "memory", "disabled"} {
		shutdown, err := Configure(mode)
		if err != nil {
			t.Fatalf("mode %q: %v", mode, err)
		}
		if err := shutdown(context.Background()); err != nil {
			t.Fatalf("shutdown mode %q: %v", mode, err)
		}
	}
	if _, err := Configure("remote"); err == nil {
		t.Fatal("unsupported telemetry mode should fail")
	}
}

func TestStructuredEventHasCorrelation(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	span := trace.NewSpanContext(trace.SpanContextConfig{TraceID: trace.TraceID{1}, SpanID: trace.SpanID{2}, TraceFlags: trace.FlagsSampled})
	Event(trace.ContextWithSpanContext(context.Background(), span), logger, "worker", "release-1", "order.completed", "orderId", "order-001")
	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"service", "processType", "releaseId", "event", "traceId", "orderId"} {
		if _, ok := record[key]; !ok {
			t.Fatalf("missing %s", key)
		}
	}
}
