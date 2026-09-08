package domain

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

const ContractVersion = "1.0.1"

var identifierPattern = regexp.MustCompile(`^[A-Za-z0-9-]+$`)

var (
	ErrInvalid  = errors.New("invalid order input")
	ErrConflict = errors.New("idempotency key conflict")
	ErrNotFound = errors.New("order not found")
)

type OrderInput struct {
	CustomerID  string `json:"customerId"`
	AmountCents int    `json:"amountCents"`
}

func (i OrderInput) Validate() error {
	if len(i.CustomerID) < 3 || len(i.CustomerID) > 64 || !identifierPattern.MatchString(i.CustomerID) {
		return ErrInvalid
	}
	if i.AmountCents < 1 || i.AmountCents > 99 {
		return ErrInvalid
	}
	return nil
}

func ValidateIdempotencyKey(key string) error {
	if len(key) < 8 || len(key) > 64 || !identifierPattern.MatchString(key) {
		return ErrInvalid
	}
	return nil
}

type Order struct {
	ID             string     `json:"id"`
	IdempotencyKey string     `json:"idempotencyKey"`
	CustomerID     string     `json:"customerId"`
	AmountCents    int        `json:"amountCents"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"createdAt"`
	CompletedAt    *time.Time `json:"completedAt"`
}

type Job struct {
	SchemaVersion  int       `json:"schemaVersion"`
	OrderID        string    `json:"orderId"`
	IdempotencyKey string    `json:"idempotencyKey"`
	AttemptedAt    time.Time `json:"attemptedAt"`
}

func (j Job) Validate() error {
	if j.SchemaVersion != 1 || uuid.Validate(j.OrderID) != nil || ValidateIdempotencyKey(j.IdempotencyKey) != nil || j.AttemptedAt.IsZero() {
		return ErrInvalid
	}
	return nil
}

type Fixture struct {
	ID             string
	IdempotencyKey string
	Input          OrderInput
}

var (
	ValidSmallOrder = Fixture{
		ID:             "00000000-0000-4000-8000-000000000001",
		IdempotencyKey: "order-submit-001",
		Input:          OrderInput{CustomerID: "customer-001", AmountCents: 59},
	}
	ValidBoundaryOrder = Fixture{
		ID:             "00000000-0000-4000-8000-000000000002",
		IdempotencyKey: "order-submit-002",
		Input:          OrderInput{CustomerID: "c-2", AmountCents: 99},
	}
)
