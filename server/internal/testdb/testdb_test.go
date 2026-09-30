package testdb

import "testing"

func TestURLPrefersTestURLThenDatabaseURLThenDev(t *testing.T) {
	t.Setenv("ARPG_TEST_DATABASE_URL", "")
	t.Setenv("ARPG_DATABASE_URL", "")
	if got := URL(); got != DevURL {
		t.Fatalf("no env: URL() = %q, want dev URL", got)
	}
	t.Setenv("ARPG_DATABASE_URL", "postgres://ci")
	if got := URL(); got != "postgres://ci" {
		t.Fatalf("ARPG_DATABASE_URL: URL() = %q", got)
	}
	t.Setenv("ARPG_TEST_DATABASE_URL", "postgres://test")
	if got := URL(); got != "postgres://test" {
		t.Fatalf("ARPG_TEST_DATABASE_URL: URL() = %q", got)
	}
}
