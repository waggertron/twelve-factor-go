package domain

import (
	"errors"
	"regexp"
	"time"
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
	ID             string    `json:"id"`
	IdempotencyKey string    `json:"idempotencyKey"`
	CustomerID     string    `json:"customerId"`
	AmountCents    int       `json:"amountCents"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type Job struct {
	SchemaVersion int    `json:"schemaVersion"`
	OrderID       string `json:"orderId"`
	Attempt       int    `json:"attempt"`
}

func (j Job) Validate() error {
	if j.SchemaVersion != 1 || j.OrderID == "" || j.Attempt < 1 || j.Attempt > 3 {
		return ErrInvalid
	}
	return nil
}
