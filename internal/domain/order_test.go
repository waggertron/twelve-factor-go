package domain

import "testing"

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
	if err := (Job{SchemaVersion: 1, OrderID: "order-001", Attempt: 1}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (Job{SchemaVersion: 2, OrderID: "order-001", Attempt: 1}).Validate(); err == nil {
		t.Fatal("unsupported schema should be rejected")
	}
}
