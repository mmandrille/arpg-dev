# v119 As-built: Live Unique Drops All Effects

Date: 2026-06-13
Spec: [`docs/specs/v119_spec-live-unique-drops-all-effects.md`](../specs/v119_spec-live-unique-drops-all-effects.md)
Plan: [`docs/plans/v119_2026-06-13-live-unique-drops-all-effects.md`](../plans/v119_2026-06-13-live-unique-drops-all-effects.md)

## What shipped

- Marked the named unique item catalog metadata as enabled/ready now that live unique behavior is
  represented by rolled equipment `effect_ids`.
- Updated shared validation so enabled unique item metadata must be `ready`, while disabled entries
  must remain `disabled_seed`.
- Added Go coverage proving every enabled unique effect can be selected by at least one compatible
  item template through the unique-effect roll helper.
- Added `unique_drop_lab` plus protocol scenario `57_live_unique_drops_all_effects`, which picks up a
  deterministic unique `cave_blade` and asserts `executioners_mark` in its `effect_ids`.
- Updated protocol bot rolled-item assertions to allow explicit unique `effect_ids` while still
  rejecting unexpected non-unique effects.
- Recorded a v119 maintenance exception for compact validator and Go coverage growth in existing
  grandfathered files; no production hotspot grew.

## Verification

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'TestUniqueEffectRollsRespectItemTypeCompatibility|TestAllEnabledUniqueEffectsReachACompatibleTemplateRoll'
ARPG_BOT_SCENARIO=live_unique_drops_all_effects make bot
make maintainability
make ci
```

## Deferred

Fixed hand-authored unique item stat packages, mystery-seller unique stock, market restrictions,
production unique art/audio, and unique-specific inspection UI remain deferred.
