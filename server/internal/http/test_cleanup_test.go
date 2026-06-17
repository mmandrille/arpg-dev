package httpapi

import "testing"

func cleanupMarketRowsForTestAccounts(t *testing.T, databaseURL string) {
	t.Helper()
	_ = databaseURL
	// Market tests use unique account/listing ids. Broad package-level cleanup races with other
	// packages in `go test ./...`, so stale rows are left harmlessly in the local test database.
}
