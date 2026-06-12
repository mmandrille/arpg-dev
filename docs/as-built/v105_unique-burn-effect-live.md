# v105 As-built: Unique Burn Effect Live

Date: 2026-06-12

## What Shipped

- Added `fire` as a supported damage type in shared schemas and Go canonicalization.
- Added deterministic unique burn DOT state to the sim/player snapshot lifecycle.
- Made equipped items with `effect_ids: ["everburning_wound"]` apply burn after successful hero
  damage.
- Burn ticks once per second for 10 seconds at 10% of the original hit damage and emits
  `monster_damaged` events with `skill_id: "everburning_wound"` and `damage_type: "fire"`.
- Added a Godot burning cue using an orange tint and additive flame marker on affected monsters.
- Added a deterministic `unique_burn_lab` world plus protocol and Godot client bot scenarios.

## Key Decisions

- No protocol bump was needed; existing `skill_effect_started`, `monster_damaged`, and
  `skill_effect_ended` event fields cover the live effect.
- The visual cue uses in-repo primitive meshes and existing status marker patterns; no addon or
  external asset was adopted.

## Verification

- `make validate-shared`
- `cd server && go test ./internal/game/... -run 'TestUniqueBurn|TestDamageType'`
- `make client-unit`
- `ARPG_BOT_SCENARIO=unique_burn_effect_live VERBOSE=1 make bot`
- `HEADLESS=1 make bot-client scenario=33_unique_burn_effect_live.json`

## Visual Verification

- `HEADLESS=1 make bot-client scenario=33_unique_burn_effect_live.json`
- `make bot-visual scenario=unique_burn_effect_live`

## Deferred

- Additional live unique effects.
- Dedicated fire resistance content and balance tuning.
- Tooltip copy that resolves unique effect ids into localized display names.
