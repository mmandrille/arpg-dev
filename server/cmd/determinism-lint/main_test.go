package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writePkg(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, src := range files { //nolint:determinism test fixture writes are order-independent
		if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func findingsAt(fs []finding) []string {
	var out []string
	for _, f := range fs {
		out = append(out, f.file+":"+strings.Split(f.msg, " ")[0])
	}
	return out
}

func TestLintFlagsEveryMapRangeFormInAnyFile(t *testing.T) {
	dir := writePkg(t, map[string]string{
		"a.go": `package game

func kv(m map[int]int) (n int) {
	for k, v := range m {
		n += k + v
	}
	return
}

func keyOnly(m map[int]int) (out []int) {
	for k := range m {
		out = append(out, k)
	}
	return
}

func valueOnly(m map[int]int) (out []int) {
	for _, v := range m {
		out = append(out, v)
	}
	return
}

func countOnly(m map[int]int) (n int) {
	for range m {
		n++
	}
	return
}

func slice(s []int) (n int) {
	for _, v := range s {
		n += v
	}
	return
}

func suppressed(m map[int]int) (n int) {
	for _, v := range m { //nolint:determinism commutative sum
		n += v
	}
	return
}
`,
	})
	got, err := lintDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	mapRanges := 0
	for _, f := range got {
		if f.counted {
			mapRanges++
		}
	}
	if mapRanges != 3 {
		t.Fatalf("want kv, key-only and value-only map ranges flagged (3), got %d: %v", mapRanges, findingsAt(got))
	}
}

func TestLintFlagsEnvAndClockAndRand(t *testing.T) {
	dir := writePkg(t, map[string]string{
		"b.go": `package game

import (
	"math/rand"
	"os"
	"time"
)

func a() string { return os.Getenv("X") }
func b() int64  { return time.Now().Unix() }
func c() int    { return rand.Int() }
func d() string { return os.Getenv("Y") } //nolint:determinism startup config
`,
	})
	got, err := lintDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var hard int
	for _, f := range got {
		if !f.counted {
			hard++
		}
	}
	if hard != 3 {
		t.Fatalf("want math/rand, os.Getenv and time.Now flagged (3), got %d: %v", hard, findingsAt(got))
	}
}

func TestLintFailsOnTypeCheckError(t *testing.T) {
	dir := writePkg(t, map[string]string{
		"c.go": "package game\n\nfunc broken() int { return undefinedName }\n",
	})
	if _, err := lintDir(dir); err == nil {
		t.Fatal("type-check failure must be an error, not a silent pass")
	}
}

func TestBaselineRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "baseline.tsv")
	want := map[string]int{"sim.go": 3, "rules.go": 1}
	if err := writeBaselineFile(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := readBaselineFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) || got["sim.go"] != 3 || got["rules.go"] != 1 {
		t.Fatalf("baseline round trip: want %v, got %v", want, got)
	}
}
