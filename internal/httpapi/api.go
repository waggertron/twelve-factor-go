package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/waggertron/twelve-factor-go/internal/domain"
	"github.com/waggertron/twelve-factor-go/internal/queue"
	"github.com/waggertron/twelve-factor-go/internal/readiness"
	"github.com/waggertron/twelve-factor-go/internal/store"
)

type API struct {
	Orders    store.Orders
	Queue     queue.Enqueuer
	Readiness *readiness.State
}

func (a API) Router() http.Handler {
	router := chi.NewRouter()
	router.Get("/health/live", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, 200, map[string]string{"status": "live"}) })
	router.Get("/health/ready", func(w http.ResponseWriter, _ *http.Request) {
		if !a.Readiness.IsReady() {
			writeError(w, 503, "not_ready", "service is not accepting work")
			return
		}
		writeJSON(w, 200, map[string]string{"status": "ready"})
	})
	router.Post("/v1/orders", a.create)
	router.Get("/v1/orders/{orderID}", a.get)
	return router
}

func (a API) create(w http.ResponseWriter, r *http.Request) {
	if !a.Readiness.IsReady() {
		writeError(w, 503, "not_ready", "service is not accepting work")
		return
	}
	key := r.Header.Get("Idempotency-Key")
	var input domain.OrderInput
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if !ValidContentType(r.Header.Get("Content-Type")) || domain.ValidateIdempotencyKey(key) != nil || decoder.Decode(&input) != nil || input.Validate() != nil {
		writeError(w, 400, "invalid_request", "request did not satisfy the order contract")
		return
	}
	order, created, err := a.Orders.Create(r.Context(), key, input)
	if errors.Is(err, domain.ErrConflict) {
		writeError(w, 409, "idempotency_conflict", "idempotency key was reused with different input")
		return
	}
	if err != nil {
		writeError(w, 500, "internal_error", "request failed")
		return
	}
	if created {
		if err := a.Queue.Enqueue(r.Context(), domain.Job{SchemaVersion: 1, OrderID: order.ID, Attempt: 1}); err != nil {
			writeError(w, 503, "queue_unavailable", "order was persisted for later dispatch")
			return
		}
		if err := a.Orders.MarkEnqueued(r.Context(), order.ID); err != nil {
			writeError(w, 500, "internal_error", "request failed")
			return
		}
	}
	writeJSON(w, http.StatusAccepted, order)
}

func (a API) get(w http.ResponseWriter, r *http.Request) {
	order, err := a.Orders.Get(r.Context(), chi.URLParam(r, "orderID"))
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, 404, "not_found", "order was not found")
		return
	}
	if err != nil {
		writeError(w, 500, "internal_error", "request failed")
		return
	}
	writeJSON(w, 200, order)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}

type MemoryQueue struct{ Jobs []domain.Job }

func (q *MemoryQueue) Enqueue(_ context.Context, job domain.Job) error {
	q.Jobs = append(q.Jobs, job)
	return nil
}

func ValidContentType(value string) bool { return strings.HasPrefix(value, "application/json") }
