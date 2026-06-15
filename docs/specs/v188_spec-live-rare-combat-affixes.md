# v188 Spec: Live Rare Combat Affixes

Status: Complete - make ci green on 2026-06-15

Codename: live-rare-combat-affixes

## Summary

Make rare-tier combat affixes from the rarity roll pool affect authoritative combat when equipped:
`hit_chance`, `crit_chance`, `evade_chance`, and the already-live `attack_speed_percent`.

## Problem

v187 made rarity roll pools inherited and data-driven, but the rare combat pool is still incomplete.
`attack_speed_percent` is rolled and live, while `hit_chance`, `crit_chance`, and `evade_chance`
are not roll candidates or equipment-derived combat inputs. Players can therefore find higher-rarity
items without the full rare-affix identity requested for combat gear.

## Goals

- Add rare-gated roll candidates for `hit_chance`, `crit_chance`, and `evade_chance`.
- Keep `attack_speed_percent` rare-gated and live.
- Apply equipped `hit_chance` and `crit_chance` rolls as additive percentage-point bonuses to the
  player's effective combat stats.
- Apply equipped `evade_chance` as a defender-side chance that turns an otherwise successful enemy
  hit into an existing `attack_missed` outcome.
- Expose the new stats in character progression and stat breakdowns so client/UI paths can inspect
  the final values.
- Prove the behavior through focused Go tests and one protocol bot scenario.

## Non-goals

- No new protocol message version.
- No PvP, monster equipment, affix naming, crafting, or rarity tuning beyond narrow roll ranges.
- No visual redesign; existing miss/crit/combat feedback presentation is reused.
- No skill cooldown or mana-cost affixes; those are v189.

## Acceptance Criteria

1. Shared item template schema and validation accept `hit_chance`, `crit_chance`, and `evade_chance`
   as supported roll stats, constrained as percent-point integer rolls.
2. At least one deterministic rare-or-higher template can roll each new rare combat affix.
3. Equipping rolled `hit_chance` can make an otherwise missing player attack hit.
4. Equipping rolled `crit_chance` can make an otherwise non-critical player hit crit.
5. Equipping rolled `evade_chance` can make an otherwise hitting monster attack miss.
6. `attack_speed_percent` remains live and covered by existing/focused tests.
7. `make maintainability`, `make validate-shared`, focused Go tests, the new bot scenario, and
   `make ci` pass.

## Files

- `shared/rules/item_templates.v0.schema.json`
- `shared/rules/item_templates.v0.json`
- `shared/golden/item_rolls.v0.schema.json`
- `server/internal/game/rules.go`
- `server/internal/game/shop.go`
- `server/internal/game/sim.go`
- `server/internal/game/*test.go`
- `tools/validate_shared.py`
- `tools/bot/scenarios/79_live_rare_combat_affixes.json`
- `docs/plans/v188_2026-06-15-live-rare-combat-affixes.md`
- `docs/as-built/v188_live-rare-combat-affixes.md`
- `PROGRESS.md`
