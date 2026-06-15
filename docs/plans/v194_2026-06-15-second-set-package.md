# v194 Plan - Second Set Package

Status: Complete - `make ci` green on 2026-06-15
Goal: Add a second enabled five-piece set package through the existing data-driven set item pipeline.
Architecture: Set items remain fixed rolled payloads sourced from `set_items.v0.json`. The new set reuses the existing rule validation, unique chest exposure, item payload construction, and equipped set bonus aggregation. Tests should assert both catalog breadth and concrete new-package behavior without hard-coding that only one set exists.
Tech stack: Shared JSON rules, Go sim tests, Python protocol bot, lifecycle docs.

## Baseline and shortcut decision

Builds on v181 set item foundation and v193 clean state. Godot plugin adoption check: reject for v194 because this slice only adds shared item data plus server/bot proof; existing set rarity UI and unique chest presentation already render additional set pieces.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/set_items.v0.json` | Add the second enabled five-piece set catalog. |
| Modify | `server/internal/game/unique_chest_test.go` | Keep Verdant coverage, add second-set payload/bonus assertions, make total set count rule-derived. |
| Add | `tools/bot/scenarios/83_second_set_package.json` | Open the debug unique chest and take one new set piece. |
| Add | `docs/as-built/v194_second-set-package.md` | Record shipped behavior and verification. |
| Modify | `PROGRESS.md` | Mark v194 complete after verification. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/unique_chest_test.go`
- [x] `tools/validate_shared.py`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: `PROGRESS.md`
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Defer extraction with rationale: `unique_chest_test.go` already owns unique/set debug chest behavior; this slice adds focused assertions and keeps production code unchanged.

Verification:
```bash
make maintainability
```

## Task 1 - Shared second set data

Files:
- Modify: `shared/rules/set_items.v0.json`

- [x] Step 1.1: Add a second enabled five-piece set using existing base templates and distinct equipment slots.
- [x] Step 1.2: Give the set deterministic partial and full bonuses using existing stat keys.

```bash
make validate-shared
```

## Task 2 - Server set behavior tests

Files:
- Modify: `server/internal/game/unique_chest_test.go`

- [x] Step 2.1: Replace the one-set count assertion with rule-derived counts while keeping explicit Verdant payload and bonus checks.
- [x] Step 2.2: Add focused assertions for the new set payload, partial bonus, full-set bonus, and summary lines.
- [x] Step 2.3: Ensure unique chest tests expect both set packages via enabled rule counts.

```bash
cd server && go test ./internal/game -run 'TestUniqueTestChest|TestSetItem' -count=1
```

## Task 3 - Bot proof

Files:
- Add: `tools/bot/scenarios/83_second_set_package.json`

- [x] Step 3.1: Add a scenario that opens `town_unique_chest`, takes one new set item by display name, and asserts inventory count.

```bash
make bot scenario=83_second_set_package.json
```

## Task 4 - Lifecycle docs and CI

Files:
- Add: `docs/as-built/v194_second-set-package.md`
- Modify: `docs/plans/v194_2026-06-15-second-set-package.md`
- Modify: `docs/specs/v194_spec-second-set-package.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Update spec/plan status, add as-built notes, and update `PROGRESS.md`.
- [x] Step 4.2: Run final verification.

```bash
make maintainability
make validate-shared
cd server && go test ./internal/game -run 'TestUniqueTestChest|TestSetItem' -count=1
make bot scenario=83_second_set_package.json
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestUniqueTestChest|TestSetItem' -count=1`
- [x] `make bot scenario=83_second_set_package.json`
- [x] `make ci`
