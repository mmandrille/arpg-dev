package migrations

import (
	"io/fs"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestMigrationVersionsAreUnique(t *testing.T) {
	t.Helper()

	entries, err := fs.Glob(FS, "*.sql")
	if err != nil {
		t.Fatalf("list migrations: %v", err)
	}
	sort.Strings(entries)

	seen := make(map[int64]string, len(entries))
	for _, name := range entries {
		version, err := versionFromName(name)
		if err != nil {
			t.Fatalf("parse version from %s: %v", name, err)
		}
		if prev, ok := seen[version]; ok {
			t.Fatalf("duplicate migration version %d: %s and %s", version, prev, name)
		}
		seen[version] = name
	}
}

func versionFromName(name string) (int64, error) {
	base := name
	if i := strings.IndexByte(base, '_'); i >= 0 {
		base = base[:i]
	}
	if base == "" {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseInt(base, 10, 64)
}
