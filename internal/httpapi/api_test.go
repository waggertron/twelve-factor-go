package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/waggertron/twelve-factor-go/internal/domain"
	"github.com/waggertron/twelve-factor-go/internal/readiness"
)

type memoryStore struct {
	mu    sync.Mutex
	byID  map[string]domain.Order
	byKey map[string]string
}

func newMemoryStore() *memoryStore {
	return &memoryStore{byID: map[string]domain.Order{}, byKey: map[string]string{}}
}

func (s *memoryStore) Create(_ context.Context, key string, input domain.OrderInput) (domain.Order, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id, ok := s.byKey[key]; ok {
		order := s.byID[id]
		if order.CustomerID != input.CustomerID || order.AmountCents != input.AmountCents {
			return domain.Order{}, false, domain.ErrConflict
		}
		return order, false, nil
	}
	now := time.Unix(1, 0).UTC()
	order := domain.Order{ID: "order-001", IdempotencyKey: key, CustomerID: input.CustomerID, AmountCents: input.AmountCents, Status: "accepted", CreatedAt: now, UpdatedAt: now}
	s.byID[order.ID], s.byKey[key] = order, order.ID
	return order, true, nil
}
func (s *memoryStore) Get(_ context.Context, id string) (domain.Order, error) {
	order, ok := s.byID[id]
	if !ok {
		return domain.Order{}, domain.ErrNotFound
	}
	return order, nil
}
func (s *memoryStore) Claim(_ context.Context, id string) (bool, error)    { return true, nil }
func (s *memoryStore) Complete(_ context.Context, id string) (bool, error) { return true, nil }
func (s *memoryStore) MarkEnqueued(_ context.Context, id string) error     { return nil }

func TestAPIAmountBoundaryAndIdempotency(t *testing.T) {
	state := &readiness.State{}
	state.MarkReady()
	orders, jobs := newMemoryStore(), &MemoryQueue{}
	handler := API{Orders: orders, Queue: jobs, Readiness: state}.Router()

	request := func(key string, amount int) *httptest.ResponseRecorder {
		body, _ := json.Marshal(domain.OrderInput{CustomerID: "customer-001", AmountCents: amount})
		req := httptest.NewRequest(http.MethodPost, "/v1/orders", strings.NewReader(string(body)))
		req.Header.Set("Idempotency-Key", key)
		req.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		return response
	}

	if status := request("boundary-99", 99).Code; status != http.StatusAccepted {
		t.Fatalf("got %d", status)
	}
	if status := request("boundary-100", 100).Code; status != http.StatusBadRequest {
		t.Fatalf("got %d", status)
	}
	if status := request("boundary-99", 99).Code; status != http.StatusAccepted {
		t.Fatalf("duplicate got %d", status)
	}
	if len(jobs.Jobs) != 1 {
		t.Fatalf("expected one job, got %d", len(jobs.Jobs))
	}
}

func TestDrainRejectsNewWork(t *testing.T) {
	state := &readiness.State{}
	state.MarkReady()
	state.BeginDrain()
	handler := API{Orders: newMemoryStore(), Queue: &MemoryQueue{}, Readiness: state}.Router()
	req := httptest.NewRequest(http.MethodPost, "/v1/orders", strings.NewReader(`{"customerId":"customer-001","amountCents":99}`))
	req.Header.Set("Idempotency-Key", "boundary-99")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("got %d", response.Code)
	}
}
