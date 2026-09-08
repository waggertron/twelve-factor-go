package commands

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

type fakeWorkerServer struct {
	calls []string
	block chan struct{}
}

func (f *fakeWorkerServer) Stop() { f.calls = append(f.calls, "stop-intake") }
func (f *fakeWorkerServer) Shutdown() {
	f.calls = append(f.calls, "drain-active")
	if f.block != nil {
		<-f.block
	}
}

func TestShutdownStopsIntakeBeforeDrain(t *testing.T) {
	server := &fakeWorkerServer{}
	if err := shutdownWorker(server, time.Second); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(server.calls, []string{"stop-intake", "drain-active"}) {
		t.Fatalf("unexpected shutdown order: %v", server.calls)
	}
}

func TestShutdownEnforcesDeadline(t *testing.T) {
	server := &fakeWorkerServer{block: make(chan struct{})}
	err := shutdownWorker(server, time.Millisecond)
	close(server.block)
	if err == nil || !strings.Contains(err.Error(), "deadline") {
		t.Fatalf("expected deadline error, got %v", err)
	}
}
