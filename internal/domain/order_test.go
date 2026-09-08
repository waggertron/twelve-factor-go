package domain

import (
	"testing"
	"time"
)

func TestAmountBoundary(t *testing.T) {
	for _, amount := range []int{1, 99} {
		if err := (OrderInput{CustomerID: "customer-001", AmountCents: amount}).Validate(); err != nil {
			t.Fatalf("amount %d should be valid: %v", amount, err)
		}
	}
	for _, amount := range []int{0, 100} {
		if err := (OrderInput{CustomerID: "customer-001", AmountCents: amount}).Validate(); err == nil {
			t.Fatalf("amount %d should be rejected", amount)
		}
	}
}

func TestJobContract(t *testing.T) {
	if err := (Job{SchemaVersion: 1, OrderID: ValidSmallOrder.ID, IdempotencyKey: ValidSmallOrder.IdempotencyKey, AttemptedAt: time.Now().UTC()}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (Job{SchemaVersion: 2, OrderID: ValidSmallOrder.ID, IdempotencyKey: ValidSmallOrder.IdempotencyKey, AttemptedAt: time.Now().UTC()}).Validate(); err == nil {
		t.Fatal("unsupported schema should be rejected")
	}
}

func TestCanonicalFixtures(t *testing.T) {
	if ValidSmallOrder.ID != "00000000-0000-4000-8000-000000000001" || ValidSmallOrder.IdempotencyKey != "order-submit-001" || ValidSmallOrder.Input.CustomerID != "customer-001" || ValidSmallOrder.Input.AmountCents != 59 {
		t.Fatalf("small fixture drifted: %+v", ValidSmallOrder)
	}
	if ValidBoundaryOrder.ID != "00000000-0000-4000-8000-000000000002" || ValidBoundaryOrder.IdempotencyKey != "order-submit-002" || ValidBoundaryOrder.Input.CustomerID != "c-2" || ValidBoundaryOrder.Input.AmountCents != 99 {
		t.Fatalf("boundary fixture drifted: %+v", ValidBoundaryOrder)
	}
}
