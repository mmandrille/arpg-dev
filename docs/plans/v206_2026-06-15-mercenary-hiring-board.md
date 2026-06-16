# v206 Plan — Mercenary Hiring Board

Status: Complete
Goal: Add a server-authored town mercenary board that sells one fixed `mercenary_guard` hire for configured gold.
Architecture: The Go sim owns board validation, gold spend, companion spawn, and replacement of existing hired mercenaries. Shared rules provide the board interactable and hire cost; `action_intent` on the board performs the first hire attempt and emits service/hire events. Client presentation remains primitive/no-panel, because the next selected slice owns roster UI.
Tech stack: shared JSON/schema, Go sim/rules/tests, protocol bot, docs.

## Baseline and shortcut decision

Builds on v198 mercenary foundation and v205 clean CI. Asset/plugin decision: reject external assets/plugins; the slice reuses the existing mercenary companion archetype and existing primitive service-board presentation patterns.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/main_config.v0.json` | Add configured mercenary hire cost. |
| Modify | `shared/rules/main_config.v0.schema.json` | Validate the hire cost. |
| Modify | `shared/rules/interactables.v0.json` | Add `town_mercenary_board` service interactable. |
| Modify | `shared/rules/interactables.v0.schema.json` | Allow service `mercenary`. |
| Modify | `shared/rules/worlds.v0.json` | Add a compact mercenary hiring lab with board, target, and enough gold setup if needed. |
| Modify | `shared/protocol/state_delta.v8.schema.json` | Add event schema coverage for mercenary service events. |
| Modify | `shared/protocol/session_snapshot.v8.schema.json` | Add recent-event schema coverage for mercenary service events. |
| Modify | `server/internal/game/rules.go` | Load/validate the configured hire cost. |
| Modify | `server/internal/game/types.go` | Add event fields if needed. |
| Modify | `server/internal/game/interactables.go` | Route mercenary service action to hiring. |
| Add | `server/internal/game/mercenary_hiring.go` | Focused board validation and hire/spawn behavior. |
| Add | `server/internal/game/mercenary_hiring_test.go` | Focused success/reject/replacement coverage. |
| Add | `tools/bot/scenarios/88_mercenary_hiring_board.json` | Protocol proof for hiring and combat assist. |
| Modify | lifecycle docs | Spec/plan/as-built/progress/scenario catalog updates. |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/rules.go`
- [x] `server/internal/game/types.go`
- [x] `server/internal/game/interactables.go`
- [x] `tools/bot/run.py` not expected; existing action steps should be enough.
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected.
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Keep mercenary hiring behavior/tests in focused new files.
- [x] Avoid expanding `tools/bot/run.py`.

Verification:
```bash
make maintainability
```

## Task 1 — Shared Board Config

Files:
- Modify: `shared/rules/main_config.v0.json`
- Modify: `shared/rules/main_config.v0.schema.json`
- Modify: `shared/rules/interactables.v0.json`
- Modify: `shared/rules/interactables.v0.schema.json`
- Modify: `shared/rules/worlds.v0.json`
- Modify: `server/internal/game/rules.go`

- [x] Step 1.1: Add `mercenary_hire_cost_gold` to main config and schema.
- [x] Step 1.2: Add `town_mercenary_board` with service `mercenary`.
- [x] Step 1.3: Add/adjust a compact hiring lab world containing the board and a combat target.
- [x] Step 1.4: Add Go rule validation for non-negative hire cost.
```bash
make validate-shared
```

## Task 2 — Protocol Contract

Files:
- Modify: `shared/protocol/state_delta.v8.schema.json`
- Modify: `shared/protocol/session_snapshot.v8.schema.json`
- Modify: `server/internal/game/types.go`

- [x] Step 2.1: Add event fields/requirements for `mercenary_board_opened` and `mercenary_hired`.
- [x] Step 2.2: Keep schema changes additive to v8.
```bash
make validate-shared
```

## Task 3 — Server Hiring Behavior

Files:
- Modify: `server/internal/game/interactables.go`
- Add: `server/internal/game/mercenary_hiring.go`
- Add: `server/internal/game/mercenary_hiring_test.go`

- [x] Step 3.1: Route `action_intent` on `town_mercenary_board` to a focused hire helper.
- [x] Step 3.2: Validate board state, player state, and available player gold.
- [x] Step 3.3: Spend gold, spawn an owned `mercenary_guard`, emit `mercenary_hired`, and update gold state.
- [x] Step 3.4: Replace an existing hired mercenary for the same owner/source instead of stacking.
- [x] Step 3.5: Add focused success, insufficient-gold, invalid-target, and replacement tests.
```bash
cd server && go test ./internal/game -run 'TestMercenaryHiring|TestMercenaryFoundation|TestCompanion'
```

## Task 4 — Bot Proof

Files:
- Add: `tools/bot/scenarios/88_mercenary_hiring_board.json`
- Modify: `docs/progress/scenario-catalog.md`

- [x] Step 4.1: Add a protocol scenario that hires from the board and asserts one owned mercenary companion.
- [x] Step 4.2: Move the owner and prove the hired mercenary damages a target.
- [x] Step 4.3: Keep the v198 foundation scenario green.
```bash
make bot scenario=mercenary_hiring_board
make bot scenario=mercenary_foundation
```

## Task 5 — Lifecycle Docs and CI

Files:
- Modify: `docs/specs/v206_spec-mercenary-hiring-board.md`
- Modify: `docs/plans/v206_2026-06-15-mercenary-hiring-board.md`
- Add: `docs/as-built/v206_mercenary-hiring-board.md`
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`

- [x] Step 5.1: Mark spec/plan complete and add as-built proof.
- [x] Step 5.2: Update progress and lifecycle docs when CI is green.
```bash
make ci
```

## Final Verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestMercenaryHiring|TestMercenaryFoundation|TestCompanion'`
- [x] `make bot scenario=mercenary_hiring_board`
- [x] `make bot scenario=mercenary_foundation`
- [x] `make ci`

## Deferred Scope

Player-character mercenary listings, durable hire records, roster UI, client hire panel, player-set pricing, mercenary equipment, commands/stances, death recovery, XP/loot/potion behavior, and multiple active hired mercenaries remain deferred.
