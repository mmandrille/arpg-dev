package store_test

import "testing"

func cleanupMarketRowsForTestAccounts(t *testing.T) {
	t.Helper()
	// Market tests use unique account/listing ids. Broad package-level cleanup races with other
	// packages in `go test ./...`, so stale rows are left harmlessly in the local test database.
}
