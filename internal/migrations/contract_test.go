package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestMigrationMatchesSharedDatabaseContract(t *testing.T) {
	sqlBytes, err := os.ReadFile("001_orders.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(sqlBytes)
	for _, expected := range []string{
		"amount_cents integer NOT NULL CHECK (amount_cents BETWEEN 1 AND 99)",
		"completed_at timestamptz",
		"CREATE TABLE IF NOT EXISTS order_jobs",
		"schema_version integer NOT NULL CHECK (schema_version = 1)",
		"published_at timestamptz",
	} {
		if !strings.Contains(sql, expected) {
			t.Fatalf("migration missing %q", expected)
		}
	}
	for _, forbidden := range []string{"updated_at", "queue_intents", "enqueued_at"} {
		if strings.Contains(sql, forbidden) {
			t.Fatalf("migration contains stale field %q", forbidden)
		}
	}
}
