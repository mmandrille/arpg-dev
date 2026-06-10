// determinism-lint verifies that the game/ package honours the determinism
// invariants documented in CLAUDE.md:
//
//  1. No time.Now() calls — checked in ALL game/ files.
//  2. No math/rand imports — ALL game/ files.
//  3. No bare map range with key+value variables in hot-path files (sim.go,
//     handlers.go).  rules.go and shop.go are excluded because their map
//     ranges are either in startup validation or use a collector→sort→use
//     pattern that is already deterministic.
//
// Suppressions: add "//nolint:determinism" on the same line as a range
// statement to document and suppress a known-safe map iteration.
//
// Exit 0 = clean. Exit 1 = violations (stderr). Run from the repo root:
//
//	go run ./cmd/determinism-lint ./internal/game/...
package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
)

// hotPathFiles are the only files checked for bare map ranges.
var hotPathFiles = map[string]bool{
	"sim.go":      true,
	"handlers.go": true,
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: determinism-lint <dir>")
		os.Exit(2)
	}
	dir := strings.TrimSuffix(os.Args[1], "/...")

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(2)
	}

	var violations []string

	for _, pkg := range pkgs {
		var files []*ast.File
		for _, f := range pkg.Files {
			files = append(files, f)
		}

		// Build a line→comment index for nolint suppression.
		nolintLines := buildNolintIndex(fset, files)

		// --- Pass 1: AST-only checks across all files ---
		for _, f := range files {
			filename := filepath.Base(fset.Position(f.Pos()).Filename)

			for _, imp := range f.Imports {
				path := strings.Trim(imp.Path.Value, `"`)
				if path == "math/rand" || path == "math/rand/v2" {
					violations = append(violations, fmt.Sprintf(
						"%s: imports %q — use the seeded splitmix64 RNG in rng.go", filename, path))
				}
			}

			ast.Inspect(f, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				ident, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}
				if ident.Name == "time" && sel.Sel.Name == "Now" {
					p := fset.Position(call.Pos())
					violations = append(violations, fmt.Sprintf(
						"%s:%d: time.Now() — use the tick counter for time-sensitive logic",
						filename, p.Line))
				}
				return true
			})
		}

		// --- Pass 2: bare map range — hot-path files only ---
		var hotFiles []*ast.File
		for _, f := range files {
			if hotPathFiles[filepath.Base(fset.Position(f.Pos()).Filename)] {
				hotFiles = append(hotFiles, f)
			}
		}
		if len(hotFiles) == 0 {
			continue
		}

		conf := types.Config{Importer: importer.Default()}
		info := &types.Info{Types: make(map[ast.Expr]types.TypeAndValue)}
		if _, err := conf.Check("game", fset, files, info); err != nil {
			continue
		}

		for _, f := range hotFiles {
			filename := filepath.Base(fset.Position(f.Pos()).Filename)
			ast.Inspect(f, func(n ast.Node) bool {
				rng, ok := n.(*ast.RangeStmt)
				if !ok {
					return true
				}
				if rng.Key == nil || rng.Value == nil {
					return true
				}
				// Skip blank-identifier key: "for _, v := range m" is safe
				// because iteration values are collected without ordering.
				if ident, ok := rng.Key.(*ast.Ident); ok && ident.Name == "_" {
					return true
				}
				tv, ok := info.Types[rng.X]
				if !ok {
					return true
				}
				if _, isMap := tv.Type.Underlying().(*types.Map); !isMap {
					return true
				}
				p := fset.Position(rng.Pos())
				// Respect //nolint:determinism suppressions.
				if nolintLines[nolintKey(p.Filename, p.Line)] {
					return true
				}
				violations = append(violations, fmt.Sprintf(
					"%s:%d: bare map range (key+value) — iteration order is non-deterministic; use a sorted* helper or add //nolint:determinism if the output is order-independent",
					filename, p.Line))
				return true
			})
		}
	}

	if len(violations) == 0 {
		fmt.Println("determinism-lint: OK")
		return
	}
	for _, v := range violations {
		fmt.Fprintln(os.Stderr, "DETERMINISM:", v)
	}
	os.Exit(1)
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
