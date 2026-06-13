# v119 Plan — Live Unique Drops All Effects

Status: Complete
Goal: Make all enabled unique effects reachable through live rolled unique equipment drops.
Architecture: Keep the existing rolled-template item model. Unique item catalog entries are metadata
for named concepts; live behavior comes from `unique_effects.v0.json` and the existing durable
`effect_ids` field on rolled item payloads.
Tech stack: Shared JSON rules, Python shared validator, Go roll tests, Python protocol bot scenario,
SDD docs.

## Baseline

v104 made unique rarity rolls attach one compatible effect; v105-v108 implemented the live effect
hooks. v119 does not add another item authority path. It removes the stale disabled-seed validator
assumption and proves all enabled effects are selectable by compatible templates.

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/game_test.go`
- [x] `tools/bot/run.py`
- [x] `tools/validate_shared.py`

Decision:
- [x] Defer extraction with rationale: this slice adds compact coverage/assertion branches to
  existing test/tool hotspots already owning rolled-item and validator behavior. No new production
  code is added; update the grandfathered baseline if the ratchet requires it.

## Task 1 — Catalog Status and Validation

Files:
- Modify: `shared/rules/unique_items.v0.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Mark live unique item metadata as enabled/ready.
- [x] Step 1.2: Update validation from “v95 seeds must remain disabled” to consistent
  enabled/status checks.

```bash
make validate-shared
```

## Task 2 — Roll Reachability Coverage

Files:
- Modify: `server/internal/game/game_test.go`

- [x] Step 2.1: Add a rule-derived test that every enabled unique effect can be selected by at
  least one compatible item template.
- [x] Step 2.2: Keep compatibility checks semantic; do not pin exact current effect counts.

```bash
cd server && go test ./internal/game/... -run 'TestUniqueEffectRollsRespectItemTypeCompatibility|TestAllEnabledUniqueEffectsReachACompatibleTemplateRoll'
```

## Task 3 — Live Protocol Proof

Files:
- Modify: `tools/bot/run.py`
- Modify: `shared/rules/worlds.v0.json`
- Create: `tools/bot/scenarios/57_live_unique_drops_all_effects.json`

- [x] Step 3.1: Let rolled-item assertions check explicit `effect_ids` instead of requiring
  all rolled items to have none.
- [x] Step 3.2: Add a compact lab proof that picks up a deterministic unique rolled item and
  asserts rarity/effect metadata.

```bash
ARPG_BOT_SCENARIO=live_unique_drops_all_effects make bot
```

## Task 4 — Lifecycle Docs and CI

Files:
- Modify: `docs/specs/v119_spec-live-unique-drops-all-effects.md`
- Modify: `docs/plans/v119_2026-06-13-live-unique-drops-all-effects.md`
- Create: `docs/as-built/v119_live-unique-drops-all-effects.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Mark spec and plan complete after verification.
- [x] Step 4.2: Update lifecycle/current status for v119.
- [x] Step 4.3: Write as-built notes.

```bash
make maintainability
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] focused Go unique roll tests
- [x] `ARPG_BOT_SCENARIO=live_unique_drops_all_effects make bot`
- [x] `make maintainability`
- [x] `make ci`

## Deferred

Fixed hand-authored unique item stat packages, mystery-seller unique stock, market restrictions,
production unique art/audio, and unique-specific inspection UI remain deferred.
