// Package testdb resolves the Postgres database used by DB-backed tests and
// decides whether an unreachable database skips or fails the test.
//
// A configured URL (ARPG_TEST_DATABASE_URL or ARPG_DATABASE_URL, which CI sets
// to this checkout's arpg_test_* database) means the caller expects DB
// coverage, so an unreachable database fails. With no URL configured the tests
// fall back to the local dev database and skip when it is down.
package testdb

import (
	"os"
	"testing"
)

// DevURL is the local `make db-up` database used when no URL is configured.
const DevURL = "postgres://arpg:arpg@localhost:5432/arpg?sslmode=disable"

// URL returns the database URL for DB-backed tests.
func URL() string {
	if v := configuredURL(); v != "" {
		return v
	}
	return DevURL
}

// SkipOrFail stops the test after a failed connect: a failure when a URL was
// configured explicitly, a skip otherwise.
func SkipOrFail(t testing.TB, what string, err error) {
	t.Helper()
	if configuredURL() != "" {
		t.Fatalf("%s: configured test database is unreachable (ARPG_TEST_DATABASE_URL/ARPG_DATABASE_URL set): %v", what, err)
	}
	t.Skipf("skipping %s: no Postgres at the dev URL: %v", what, err)
}

func configuredURL() string {
	if v := os.Getenv("ARPG_TEST_DATABASE_URL"); v != "" {
		return v
	}
	return os.Getenv("ARPG_DATABASE_URL")
}
