package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

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
