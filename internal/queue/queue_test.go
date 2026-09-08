package queue

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/waggertron/twelve-factor-go/internal/domain"
)

func TestSharedQueueAndPayloadContract(t *testing.T) {
	if QueueName != "orders.v1" || TaskType != "complete-order" {
		t.Fatalf("queue contract drifted: %q %q", QueueName, TaskType)
	}
	job := domain.Job{
		SchemaVersion:  1,
		OrderID:        domain.ValidSmallOrder.ID,
		IdempotencyKey: domain.ValidSmallOrder.IdempotencyKey,
		AttemptedAt:    time.Date(2026, time.September, 8, 7, 0, 0, 0, time.UTC),
	}
	payload, err := json.Marshal(job)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(payload, &fields); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"schemaVersion", "orderId", "idempotencyKey", "attemptedAt"} {
		if _, ok := fields[key]; !ok {
			t.Fatalf("payload missing %q", key)
		}
	}
	if len(fields) != 4 {
		t.Fatalf("unexpected payload fields: %v", fields)
	}
}
