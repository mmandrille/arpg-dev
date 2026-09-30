// determinism-lint verifies that the game/ package honours the determinism
// invariants documented in CLAUDE.md:
//
//  1. No time.Now() calls — ALL non-test game/ files.
//  2. No math/rand imports — ALL non-test game/ files.
//  3. No os.Getenv / os.LookupEnv — ALL non-test game/ files. Process
//     environment is not a recorded replay input; callers must pass
//     configuration in explicitly.
//  4. No range over a map (key+value, key-only or value-only) — ALL non-test
//     game/ files. Go map iteration order is randomised, so any map range whose
//     result depends on visit order (first match, strict-< tie-break, append,
//     ID allocation) breaks replay.
//
// Suppressions: add "//nolint:determinism" with a one-line WHY on the same
// line as a flagged statement. For map ranges this documents a known
// order-independent iteration (output is a map, commutative sum, existence
// check, collect-then-sort); for os.Getenv it documents startup-only config.
//
// Grandfathered sites: pre-existing unsuppressed findings are counted per file
// in the baseline file (default: .maintainability/determinism-baseline.tsv).
// A file may not exceed its baseline, and a file that drops below it must
// lower the baseline in the same change. Run with -write-baseline to
// regenerate it after auditing sites.
//
// A type-check failure is an error (exit 2), never a silent pass.
//
// Exit 0 = clean. Exit 1 = violations (stderr). Exit 2 = usage/parse/type
// error. Run from server/:
//
//	go run ./cmd/determinism-lint -baseline ../.maintainability/determinism-baseline.tsv ./internal/game/...
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type finding struct {
	file string
	line int
	msg  string
	// counted findings are map ranges, which may be grandfathered by the
	// baseline. Uncounted findings (time.Now, math/rand, os.Getenv) always fail.
	counted bool
}

func main() {
	baselinePath := flag.String("baseline", "", "per-file baseline TSV for grandfathered map-range sites")
	writeBaseline := flag.Bool("write-baseline", false, "rewrite -baseline from the current findings and exit")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: determinism-lint [-baseline file] [-write-baseline] <dir>")
		os.Exit(2)
	}
	dir := strings.TrimSuffix(flag.Arg(0), "/...")

	findings, err := lintDir(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "determinism-lint:", err)
		os.Exit(2)
	}

	counts := map[string]int{}
	var hard []finding
	for _, f := range findings {
		if f.counted {
			counts[f.file]++
		} else {
			hard = append(hard, f)
		}
	}

	if *writeBaseline {
		if *baselinePath == "" {
			fmt.Fprintln(os.Stderr, "determinism-lint: -write-baseline requires -baseline")
			os.Exit(2)
		}
		if err := writeBaselineFile(*baselinePath, counts); err != nil {
			fmt.Fprintln(os.Stderr, "determinism-lint:", err)
			os.Exit(2)
		}
		fmt.Printf("determinism-lint: wrote baseline for %d files (%d sites)\n", len(counts), total(counts))
		return
	}

	baseline := map[string]int{}
	if *baselinePath != "" {
		baseline, err = readBaselineFile(*baselinePath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "determinism-lint:", err)
			os.Exit(2)
		}
	}

	var problems []string
	for _, f := range hard {
		problems = append(problems, fmt.Sprintf("%s:%d: %s", f.file, f.line, f.msg))
	}
	for _, f := range findings {
		if f.counted && counts[f.file] > baseline[f.file] {
			problems = append(problems, fmt.Sprintf("%s:%d: %s", f.file, f.line, f.msg))
		}
	}
	for _, file := range sortedKeys(counts, baseline) {
		got, want := counts[file], baseline[file]
		switch {
		case got > want:
			problems = append(problems, fmt.Sprintf(
				"%s: %d unsuppressed map ranges exceed the grandfathered baseline of %d (sites listed above)",
				file, got, want))
		case got < want:
			problems = append(problems, fmt.Sprintf(
				"%s: %d unsuppressed map ranges is below the baseline of %d — lower the baseline in the same change (-write-baseline)",
				file, got, want))
		}
	}

	if len(problems) == 0 {
		fmt.Printf("determinism-lint: OK (%d grandfathered map-range sites in %d files)\n", total(counts), len(counts))
		return
	}
	for _, p := range problems {
		fmt.Fprintln(os.Stderr, "DETERMINISM:", p)
	}
	os.Exit(1)
}

func lintDir(dir string) ([]finding, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	var out []finding
	for _, name := range sortedPkgNames(pkgs) {
		pkg := pkgs[name]
		var files []*ast.File
		for _, fname := range sortedFileNames(pkg) {
			files = append(files, pkg.Files[fname])
		}

		conf := types.Config{Importer: importer.ForCompiler(fset, "source", nil)}
		info := &types.Info{Types: make(map[ast.Expr]types.TypeAndValue)}
		if _, err := conf.Check(name, fset, files, info); err != nil {
			return nil, fmt.Errorf("type-check of package %s failed (the lint cannot see map types): %w", name, err)
		}

		nolintLines := buildNolintIndex(fset, files)
		for _, f := range files {
			out = append(out, lintFile(fset, f, info, nolintLines)...)
		}
	}
	return out, nil
}

func lintFile(fset *token.FileSet, f *ast.File, info *types.Info, nolintLines map[string]bool) []finding {
	filename := filepath.Base(fset.Position(f.Pos()).Filename)
	var out []finding

	for _, imp := range f.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if path == "math/rand" || path == "math/rand/v2" {
			out = append(out, finding{file: filename, line: fset.Position(imp.Pos()).Line,
				msg: fmt.Sprintf("imports %q — use the seeded splitmix64 RNG in rng.go", path)})
		}
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			sel, ok := node.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			ident, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}
			p := fset.Position(node.Pos())
			if nolintLines[nolintKey(p.Filename, p.Line)] {
				return true
			}
			line := p.Line
			switch {
			case ident.Name == "time" && sel.Sel.Name == "Now":
				out = append(out, finding{file: filename, line: line,
					msg: "time.Now() — use the tick counter for time-sensitive logic"})
			case ident.Name == "os" && (sel.Sel.Name == "Getenv" || sel.Sel.Name == "LookupEnv"):
				out = append(out, finding{file: filename, line: line,
					msg: "os." + sel.Sel.Name + "() — process env is not a recorded replay input; pass config in explicitly"})
			}
		case *ast.RangeStmt:
			if node.Key == nil && node.Value == nil {
				return true // "for range m" only counts iterations
			}
			tv, ok := info.Types[node.X]
			if !ok {
				return true
			}
			if _, isMap := tv.Type.Underlying().(*types.Map); !isMap {
				return true
			}
			p := fset.Position(node.Pos())
			if nolintLines[nolintKey(p.Filename, p.Line)] {
				return true
			}
			out = append(out, finding{file: filename, line: p.Line, counted: true,
				msg: "map range — iteration order is non-deterministic; use a sorted* helper or add //nolint:determinism with a WHY if the result is order-independent"})
		}
		return true
	})
	return out
}

func readBaselineFile(path string) (map[string]int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read baseline: %w", err)
	}
	out := map[string]int{}
	for i, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		var file string
		var n int
		if _, err := fmt.Sscanf(line, "%s\t%d", &file, &n); err != nil {
			return nil, fmt.Errorf("baseline %s:%d: want \"<file>\\t<count>\": %w", path, i+1, err)
		}
		out[file] = n
	}
	return out, nil
}

func writeBaselineFile(path string, counts map[string]int) error {
	var b strings.Builder
	b.WriteString("# Grandfathered unsuppressed map ranges in server/internal/game (determinism-lint).\n")
	b.WriteString("# Each site is unaudited debt: sort it or add //nolint:determinism with a WHY, then lower the count.\n")
	b.WriteString("# Regenerate: cd server && go run ./cmd/determinism-lint -baseline ../.maintainability/determinism-baseline.tsv -write-baseline ./internal/game/...\n")
	for _, file := range sortedKeys(counts, nil) {
		fmt.Fprintf(&b, "%s\t%d\n", file, counts[file])
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func total(counts map[string]int) int {
	n := 0
	for _, c := range counts { //nolint:determinism commutative sum
		n += c
	}
	return n
}

func sortedKeys(a, b map[string]int) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range []map[string]int{a, b} {
		for k := range m {
			if !seen[k] {
				seen[k] = true
				out = append(out, k)
			}
		}
	}
	sort.Strings(out)
	return out
}

func sortedPkgNames(pkgs map[string]*ast.Package) []string {
	var out []string
	for name := range pkgs {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func sortedFileNames(pkg *ast.Package) []string {
	var out []string
	for name := range pkg.Files {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func nolintKey(filename string, line int) string {
	return fmt.Sprintf("%s:%d", filename, line)
}

// buildNolintIndex returns a set of "filename:line" keys where a
// //nolint:determinism comment appears on that line.
func buildNolintIndex(fset *token.FileSet, files []*ast.File) map[string]bool {
	out := make(map[string]bool)
	for _, f := range files {
		for _, cg := range f.Comments {
			for _, c := range cg.List {
				if strings.Contains(c.Text, "nolint:determinism") {
					p := fset.Position(c.Pos())
					out[nolintKey(p.Filename, p.Line)] = true
				}
			}
		}
	}
	return out
}
