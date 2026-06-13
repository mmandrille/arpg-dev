# v135 Plan — Second Named Unique

Status: Complete
Goal: Add a second enabled named unique package and prove the deterministic chest path includes it.
Architecture: Keep named unique data in `shared/rules/unique_items.v0.json`; reuse existing Go
payload construction, validation, and purple unique chest wiring.
Tech stack: shared JSON rules, Go rule tests, Python validator tests, lifecycle docs.

## Baseline And Shortcut Decision

Builds on v134 unique inspection UI. Godot plugin adoption is not applicable because this slice is a
data/catalog expansion with server validation tests and no new client presentation code.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/unique_items.v0.json` | Add the second enabled named unique |
| Modify | `server/internal/game/unique_chest_test.go` | Assert both named unique payloads/chest rows |
| Modify | `tools/test_validate_unique_items.py` | Keep validator unit fixtures representative |
| Create | `docs/as-built/v135_second-named-unique.md` | As-built summary |
| Modify | `docs/specs/v135_spec-second-named-unique.md` | Status closeout |
| Modify | `docs/plans/v135_2026-06-13-second-named-unique.md` | Checkbox closeout |
| Modify | `PROGRESS.md` | Lifecycle and next-slice update |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] None expected.

Decision:
- [x] Keep rule-test edits inside the existing focused unique chest test file unless it crosses the
  ratchet threshold.

Verification:
```bash
make maintainability
```

## Task 1 — Add Catalog Entry

Files:
- Modify: `shared/rules/unique_items.v0.json`

- [x] Step 1.1: Add a second enabled ready named unique using an existing template and live effect.
- [x] Step 1.2: Give it fixed stats and a minimum level that are deterministic and conservative.
- [x] Step 1.3: Keep behavior hook text clear enough for validator and future maintainers.

## Task 2 — Rule Tests

Files:
- Modify: `server/internal/game/unique_chest_test.go`
- Modify: `tools/test_validate_unique_items.py`

- [x] Step 2.1: Assert Go payload construction for both named uniques.
- [x] Step 2.2: Assert deterministic chest rows include both enabled named uniques.
- [x] Step 2.3: Keep Python validator fixture coverage representative after catalog expansion.

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestNamedUnique|TestUniqueTestChest'
.venv/bin/python -m pytest tools/test_validate_unique_items.py -q
```

## Task 3 — Lifecycle Docs And CI

Files:
- Create: `docs/as-built/v135_second-named-unique.md`
- Modify: `docs/specs/v135_spec-second-named-unique.md`
- Modify: `docs/plans/v135_2026-06-13-second-named-unique.md`
- Modify: `PROGRESS.md`

- [x] Step 3.1: Mark the spec and plan complete.
- [x] Step 3.2: Record v135 completion and next slice in `PROGRESS.md`.
- [x] Step 3.3: Add the v135 as-built summary.

```bash
make maintainability
make ci
```

## Final Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestNamedUnique|TestUniqueTestChest'`
- [x] `.venv/bin/python -m pytest tools/test_validate_unique_items.py -q`
- [x] `make maintainability`
- [x] `make ci`
