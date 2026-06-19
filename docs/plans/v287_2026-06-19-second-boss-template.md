# v287 Plan — Second Boss Template

Status: Implemented
Goal: Add `crypt_matron` as a second data-authored boss template and make boss-floor template
selection deterministic from the configured pool.
Architecture: Keep all combat and presentation behavior on existing boss systems. Add data and a
small generator helper; avoid new pattern code or client presentation special cases.
Tech stack: Shared JSON/schema validation, Go dungeon generation tests, protocol bot scenario.

## Baseline and shortcut decision

Builds on the current Cave Warden boss template, v282 rectangle pattern, v283 summon-bat/HP tuning,
and existing boss health bar generic fallback. No new assets or pattern primitives are needed.

Asset/plugin decision:

- Adopt: existing boss templates schema, boss pattern definitions, monster visuals, boss loot tables,
  and generic boss portrait fallback.
- Borrow: existing `boss_floor_gate` scenario structure and deterministic RNG helper.
- Reject: external assets, plugins, production art, new reward tables, new arenas, and weighted pool
  design.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/boss_templates.v0.json` | Add `crypt_matron` template. |
| Modify | `shared/rules/dungeon_generation.v0.json` | Add `crypt_matron` to boss pool. |
| Modify | `server/internal/game/dungeon_gen.go` | Select boss template deterministically from the pool. |
| Create | `server/internal/game/boss_template_selection_test.go` | Add stable pool-selection coverage. |
| Create | `tools/bot/scenarios/95_second_boss_template.json` | Protocol proof for the new boss. |
| Create during finish | `docs/as-built/v287_second-boss-template.md` | Record proof and commands. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:

- [x] `server/internal/game/boss_template_selection_test.go` — added focused boss selection tests
  outside large existing test files.

Decision:

- [x] Extract helper/module: add a small local generator helper in `dungeon_gen.go`; no broader
  generation extraction for this slice.
- [x] Defer extraction with rationale: weighted boss selection and richer boss scheduling remain
  future design work.

Verification:

```bash
make maintainability
```

## Task 1 — Boss data

Files:

- Modify: `shared/rules/boss_templates.v0.json`
- Modify: `shared/rules/dungeon_generation.v0.json`

- [x] Step 1.1: Add `crypt_matron` with existing base monster, existing pattern deck, boss loot
  table, visual color/model/scale, and optional enrage.
- [x] Step 1.2: Add `crypt_matron` to `boss_template_pool` after `cave_warden`.
- [x] Step 1.3: Keep Cave Warden data unchanged.

Verify:

```bash
make validate-shared
```

## Task 2 — Deterministic selection

Files:

- Modify: `server/internal/game/dungeon_gen.go`
- Create: `server/internal/game/boss_template_selection_test.go`

- [x] Step 2.1: Add a helper that selects `boss_template_pool[0]` for a one-item pool and otherwise
  uses `NewRNG(SeedToUint64(seed + "|boss_template_selection|" + abs(level)))`.
- [x] Step 2.2: Prove `boss_floor_gate`, `boss_special_drops`, and `boss_enrage_phase` still select
  `cave_warden`.
- [x] Step 2.3: Prove seed `second_boss_template` selects `crypt_matron`.
- [x] Step 2.4: Prove generated boss entity fields match the selected template.

Verify:

```bash
(cd server && go test ./internal/game -run 'TestBoss' -count=1)
```

## Task 3 — Protocol proof

Files:

- Create: `tools/bot/scenarios/95_second_boss_template.json`

- [x] Step 3.1: Start `boss_floor_gate_lab` with seed `second_boss_template`.
- [x] Step 3.2: Assert exactly one boss with `boss_template_id: crypt_matron`.
- [x] Step 3.3: Wait for an authored `crypt_matron` pattern event.
- [x] Step 3.4: Kill the boss and assert `boss_killed` plus exit unlock.

Verify:

```bash
make bot scenario=second_boss_template
```

## Task 4 — Docs and lifecycle

Files:

- Existing: `docs/specs/v287_spec-second-boss-template.md`
- Existing: `docs/plans/v287_2026-06-19-second-boss-template.md`
- Create during finish: `docs/as-built/v287_second-boss-template.md`
- Modify during finish: `PROGRESS.md`

- [x] Step 4.1: Record focused checks and bot proof in the as-built note.
- [x] Step 4.2: Update lifecycle/current status during finish.

## Task 5 — Final verification

- [x] `make validate-shared`
- [x] `(cd server && go test ./internal/game -run 'TestBoss' -count=1)`
- [x] `make bot scenario=second_boss_template`
- [x] `make maintainability`
