# v187 Plan — Rarity Roll Pools

Status: Complete - make ci green on 2026-06-15
Goal: Make item roll counts and rollable stat pools data-driven by rarity.
Architecture: Shared item template rules define rarity roll-count ranges and minimum-rarity-gated
roll candidates. The Go sim remains the only roll executor, using the existing seeded item RNG to
choose rarity, roll count, candidate stats, and stat values. Existing rolled item payloads carry the
final stat map, so no protocol schema bump is required.
Tech stack: Shared JSON/schema, Go sim/rules loader, Python bot scenario, existing Godot tooltip labels.

## Baseline and shortcut decision

Builds on v186 and reuses v23 rolled item payloads, v104 unique roll attachment, and v181 set
rarity presentation. Godot plugin decision: reject new inventory/UI plugin for this slice; only
existing tooltip/stat label paths may need small label additions.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/item_templates.v0.schema.json` | Add rarity roll-count ranges and min-rarity roll candidates. |
| Modify | `shared/rules/item_templates.v0.json` | Configure magic 1-2, rare 2-4, unique/set 3-5, and gated pools. |
| Modify | `shared/golden/item_rolls.json` | Pin deterministic roll-count and inherited-pool examples. |
| Modify | `shared/golden/item_rolls.v0.schema.json` | Allow expected roll count metadata if needed. |
| Modify | `server/internal/game/rules.go` | Load and validate new rarity/candidate fields. |
| Modify | `server/internal/game/shop.go` | Roll count from range and filter candidates by rarity. |
| Modify | `server/internal/game/game_test.go` | Focused deterministic roll and validation tests. |
| Modify | `client/scripts/stat_labels.gd` | Add readable labels for any newly introduced stat ids. |
| Create | `tools/bot/scenarios/78_rarity_roll_pools.json` | End-to-end protocol proof. |
| Modify | `tools/bot/run.py` or helpers | Add narrow assertion only if scenario DSL lacks one. |
| Create | `docs/as-built/v187_rarity-roll-pools.md` | As-built summary. |
| Modify | `PROGRESS.md` | Lifecycle and backlog update. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [ ] `client/scripts/main.gd`
- [x] `server/internal/game/game_test.go`
- [x] `tools/bot/run.py`
- [x] `tools/validate_shared.py`
- [ ] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [ ] Extract focused helper/module/test file as part of this slice, or
- [x] Defer extraction with rationale: existing item-roll tests and bot runner are touched narrowly;
  extraction would be broader than the feature.

Verification:
```bash
make maintainability
```

## Task 1 — Shared contract and catalog

Files:
- Modify: `shared/rules/item_templates.v0.schema.json`
- Modify: `shared/rules/item_templates.v0.json`
- Modify: `shared/golden/item_rolls.v0.schema.json`
- Modify: `shared/golden/item_rolls.json`

- [x] Step 1.1: Add rarity roll-count range fields while preserving old catalog intent.
- [x] Step 1.2: Add `min_rarity` or equivalent gated roll-candidate field.
- [x] Step 1.3: Configure magic 1-2, rare 2-4, unique 3-5, and set 3-5 if the catalog declares set.
- [x] Step 1.4: Update item-roll golden shape and examples.
```bash
make validate-shared
```

## Task 2 — Server roll execution

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/shop.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 2.1: Parse and validate roll-count ranges and gated candidates.
- [x] Step 2.2: Roll count deterministically from the configured range.
- [x] Step 2.3: Filter roll candidates by current rarity rank so higher rarities inherit lower pools.
- [x] Step 2.4: Add focused tests for roll-count ranges, inherited pools, and invalid configs.
```bash
cd server && go test ./internal/game/... -run 'ItemRoll|ItemTemplate|Rarity'
```

## Task 3 — Tooltip labels and bot proof

Files:
- Modify: `client/scripts/stat_labels.gd`
- Create: `tools/bot/scenarios/78_rarity_roll_pools.json`
- Modify: `tools/bot/run.py` or helper modules only if required

- [x] Step 3.1: Add labels for new stat keys introduced by the shared catalog.
- [x] Step 3.2: Add a bot scenario that rolls or acquires a deterministic higher-rarity item.
- [x] Step 3.3: Assert rarity, roll count range, and expected inherited-pool stat presence.
```bash
make client-unit
make bot scenario=78_rarity_roll_pools.json
```

## Task 4 — Lifecycle docs and CI

Files:
- Create: `docs/as-built/v187_rarity-roll-pools.md`
- Modify: `docs/specs/v187_spec-rarity-roll-pools.md`
- Modify: `docs/plans/v187_2026-06-15-rarity-roll-pools.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Mark spec and plan complete after verification.
- [x] Step 4.2: Add the as-built summary and PROGRESS lifecycle row.
- [x] Step 4.3: Record deferred scope for affix names, crafting, and final tuning.
```bash
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/... -run 'ItemRoll|ItemTemplate|Rarity'`
- [x] `make client-unit`
- [x] `make bot scenario=78_rarity_roll_pools.json`
- [x] `make ci`
