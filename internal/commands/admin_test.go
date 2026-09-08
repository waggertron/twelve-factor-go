package commands

import (
	"io"
	"log/slog"
	"strings"
	"testing"
)

func TestAdminRejectsUnsupportedTargetBeforeConfig(t *testing.T) {
	command := New(slog.New(slog.NewTextHandler(io.Discard, nil)))
	command.SetArgs([]string{"admin", "migrate", "--target", "999"})
	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "unsupported migration target") {
		t.Fatalf("expected bounded target error, got %v", err)
	}
}
