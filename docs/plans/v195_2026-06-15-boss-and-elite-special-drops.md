# v195 Plan - Boss And Elite Special Drops

Status: Complete - `make ci` green on 2026-06-15
Goal: Add data-driven authored unique/set drops for boss and elite reward sources.
Architecture: Extend loot/treasure drop descriptors with optional `unique_item_id` and `set_item_id` fields. Rule validation resolves those ids against the ready unique/set catalogs, and `spawnLootDrops` converts them through the same payload builders used by the unique chest. Boss and elite reward loot tables point to new special treasure classes while normal monster drop-rate math stays scoped to dungeon monster tables.
Tech stack: Shared JSON schemas/rules, Go sim, Python protocol bot, lifecycle docs.

## Baseline and shortcut decision

Builds on v194's second set package and v193's named unique payload. Godot plugin adoption check: reject for v195 because existing unique/set item payloads already render through loot labels/tooltips; this slice is server/shared/bot behavior.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/treasure_classes.v0.schema.json` | Permit authored unique/set treasure entries. |
| Modify | `shared/rules/loot_tables.v0.schema.json` | Keep direct loot table entries aligned with the expanded drop descriptor. |
| Modify | `shared/rules/loot_tables.v0.json` | Point boss and elite reward sources at special treasure classes. |
| Modify | `shared/rules/treasure_classes.v0.json` | Define boss and elite special reward attempts. |
| Modify | `server/internal/game/rules.go` | Validate and roll authored unique/set drop ids. |
| Modify | `server/internal/game/sim.go` | Spawn fixed unique/set payloads from loot drops. |
| Modify | `server/internal/game/game_test.go` or focused test file | Prove special drops and normal drop-rate stability. |
| Add | `tools/bot/scenarios/84_boss_special_drops.json` | Kill boss and assert special drops appear. |
| Add | `docs/as-built/v195_boss-and-elite-special-drops.md` | Record shipped behavior and verification. |
| Modify | `PROGRESS.md` | Mark v195 complete after verification. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/game_test.go`
- [x] `server/internal/game/sim.go`
- [x] `server/internal/game/rules.go`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: `PROGRESS.md`
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/test file where practical, or defer with rationale if editing existing loot helpers is the smallest safe change.

Verification:
```bash
make maintainability
```

## Task 1 - Shared contract and reward data

Files:
- Modify: `shared/rules/treasure_classes.v0.schema.json`
- Modify: `shared/rules/loot_tables.v0.schema.json`
- Modify: `shared/rules/loot_tables.v0.json`
- Modify: `shared/rules/treasure_classes.v0.json`

- [x] Step 1.1: Add `unique_item_id` and `set_item_id` to treasure/loot entry schemas as mutually exclusive alternatives.
- [x] Step 1.2: Add boss and elite special treasure classes using existing ready named unique/set pieces.
- [x] Step 1.3: Point `boss_drop_tier_1` and the elite objective reward table at those special classes.

```bash
make validate-shared
```

## Task 2 - Server authored drop support

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/sim.go`
- Add/Modify: focused server tests

- [x] Step 2.1: Extend `LootEntry`, `TreasureClassEntry`, and `LootDrop` with unique/set ids.
- [x] Step 2.2: Validate authored ids against enabled ready catalogs and preserve exactly-one entry semantics.
- [x] Step 2.3: Roll authored ids through treasure classes and convert them to fixed payloads at spawn time.
- [x] Step 2.4: Add focused tests for boss/elite special drops and normal dungeon monster drop-rate math.

```bash
cd server && go test ./internal/game -run 'TreasureClass|SpecialDrop|Boss|EliteObjective|DungeonMonsterLootRate' -count=1
```

## Task 3 - Bot proof

Files:
- Add: `tools/bot/scenarios/84_boss_special_drops.json`
- Modify: `tools/bot/run.py` if a generic dropped-loot assertion is missing.

- [x] Step 3.1: Add a boss-floor scenario that kills Cave Warden and asserts the special named unique/set drops spawn.
- [x] Step 3.2: Add only minimal reusable bot assertion support if existing assertions cannot inspect dropped loot payloads.

```bash
make bot scenario=84_boss_special_drops.json
```

## Task 4 - Lifecycle docs and CI

Files:
- Add: `docs/as-built/v195_boss-and-elite-special-drops.md`
- Modify: `docs/specs/v195_spec-boss-and-elite-special-drops.md`
- Modify: `docs/plans/v195_2026-06-15-boss-and-elite-special-drops.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Update lifecycle docs after verification.
- [x] Step 4.2: Run final CI.

```bash
make maintainability
make validate-shared
cd server && go test ./internal/game -run 'TreasureClass|SpecialDrop|Boss|EliteObjective|DungeonMonsterLootRate' -count=1
make bot scenario=84_boss_special_drops.json
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TreasureClass|SpecialDrop|Boss|EliteObjective|DungeonMonsterLootRate' -count=1`
- [x] `make bot scenario=84_boss_special_drops.json`
- [x] `make ci`
