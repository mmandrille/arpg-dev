# v119 Spec: Live Unique Drops All Effects

Status: Complete
Date: 2026-06-13
Codename: `live-unique-drops-all-effects`

## Purpose

Make the existing unique-effect catalog fully reachable from real rolled unique equipment drops.
Generated template drops already use the shared rarity table and unique rolls already attach one
compatible enabled effect; v119 removes the remaining disabled unique-item seed assumption and adds
proof that every enabled compatible effect can be selected through the live roll path.

## Non-goals

- No fixed hand-authored unique item stat packages.
- No mystery-seller unique stock changes.
- No new unique-effect mechanics beyond the already implemented hooks.
- No unique-specific market restrictions.
- No production unique art/audio.

## Acceptance criteria

- `shared/rules/unique_items.v0.json` no longer marks catalog entries as disabled seeds when their
  referenced behavior model is now live.
- Shared validation accepts enabled/ready unique item metadata and rejects inconsistent enabled/status
  pairs.
- Go coverage proves every enabled unique effect is selectable for at least one compatible item
  template through `rollUniqueEffectForTemplate`.
- A live protocol bot scenario picks up a rolled unique template item from floor loot and asserts
  `rarity: unique` plus one expected `effect_id`.
- Existing rolled-drop assertions continue accepting normal non-unique items with empty `effect_ids`.

## Likely files

- `shared/rules/unique_items.v0.json`
- `tools/validate_shared.py`
- `server/internal/game/game_test.go`
- `tools/bot/run.py`
- `shared/rules/worlds.v0.json` or a focused existing lab world
- `tools/bot/scenarios/57_live_unique_drops_all_effects.json`
- `docs/as-built/v119_live-unique-drops-all-effects.md`
- `PROGRESS.md`

## Verification

- `make validate-shared`
- `cd server && go test ./internal/game/... -run 'TestUniqueEffectRollsRespectItemTypeCompatibility|TestAllEnabledUniqueEffectsReachACompatibleTemplateRoll'`
- `ARPG_BOT_SCENARIO=live_unique_drops_all_effects make bot`
- `make maintainability`
- `make ci`
