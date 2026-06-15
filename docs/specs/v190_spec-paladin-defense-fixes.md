# Spec: `paladin-defense-fixes`

Status: Complete
Date: 2026-06-15
Codename: `paladin-defense-fixes`
Slice: v190 - paladin defense fixes
Baseline: v184 `revived-monster-companion`

## Purpose

Fix Paladin defensive skill behavior reported from play: shield block must appear in derived stats,
Holy Shield must visibly refresh armor/block while active, and Sanctuary must become a short
server-authoritative immunity dome around the caster and allies.

## Non-goals

- No final Paladin balance pass beyond the requested 5 range, 6 second duration, and 60 second cooldown.
- No external Godot plugin or asset dependency; the dome is a small code-native marker.
- No new skill tree layout or class-gate changes.

## Acceptance Criteria

1. Effective shield block is present in `derived_stats.block_percent` and the character stats panel can display it.
2. Holy Shield start/end refreshes the authoritative character progression payload so derived armor/block update immediately.
3. Sanctuary applies a radius 5 `sanctuary` effect for 60 ticks and starts a 600 tick cooldown.
4. Sanctuary prevents incoming player damage while active across monster melee/projectile, retaliation, and boss phase damage paths.
5. Sanctuary emits observable zero-damage `outcome: "immune"` combat events and renders a yellow dome from server-owned `effect_ids`.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestRulesLoad|TestSanctuary|TestHolyShield|TestCombatStatBreakdownsIncludeEquipmentAndCap' -count=1`
- `make client-unit`
- `make bot scenario=70_paladin_sanctuary.json`
- `make ci`

## Open Questions and Risks

- Production VFX/audio remains deferred. The current dome is intentionally primitive but readable.
