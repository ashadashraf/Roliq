package database_test

import (
	"math"
	"testing"

	"github.com/pressly/goose/v3"
)

func TestMigrationsAreDiscoverable(t *testing.T) {
	migrations, err := goose.CollectMigrations("migrations", 0, math.MaxInt64)
	if err != nil {
		t.Fatalf("collect migrations: %v", err)
	}
	if len(migrations) == 0 {
		t.Fatal("expected at least one database migration")
	}
	if migrations[0].Version != 1 {
		t.Fatalf("expected first migration version 1, got %d", migrations[0].Version)
	}
}
