# v285 Spec: Damage Type Readability

Status: Implemented
Date: 2026-06-19
Codename: `damage-type-readability`

## Purpose

Make existing floating combat text communicate the authoritative `damage_type` already emitted by
combat events. Players should be able to distinguish fire, cold, lightning, poison, force, and
physical hits from normal numeric damage without opening debug logs.

This is a client presentation slice. Server combat math, damage event contracts, skill formulas,
status effects, and balance stay unchanged.

## Non-goals

- Do not add new damage types, resistances, skills, or balance tuning.
- Do not change protocol event shape or server-side combat calculations.
- Do not add new art assets, external fonts, icon packs, or VFX systems.
- Do not redesign the floating combat text system or settings toggle.
- Do not change miss, block, immune, threat, mana, heal, or inventory feedback semantics.

## Acceptance Criteria

- Normal damaging combat events with a known `damage_type` show a short readable type label beside
  the damage amount.
- Damage-type text uses distinct colors for at least `fire`, `cold`, `lightning`, `poison`,
  `force`, and `physical`.
- Existing special outcomes still take priority: `MISS`, `BLOCK`, and `IMMUNE` remain unchanged.
- Critical hits remain visibly critical while still carrying the damage-type label when the event
  has a known `damage_type`.
- Poison damage no longer depends on the `poison_stab` skill id for its green presentation; the
  authoritative `damage_type` drives that presentation.
- The floating combat text setting still suppresses new damage numbers.
- Client bot debug state exposes enough metadata to assert a damage number's type.

## Scope And Likely Files

- `client/scripts/damage_number.gd` stores the damage type used by bot debug state.
- `client/scripts/main.gd` maps event `damage_type` to label, color, and variant before spawning
  floating combat text.
- `client/scripts/bot_scenario_runner.gd` can match `damage_type` on `wait_damage_number` and
  `assert_damage_number` steps.
- `client/tests/test_rogue_presentation.gd` covers the damage-type display helper behavior.
- `tools/bot/scenarios/client/11_combat_feedback.json` can assert the normal hit carries a
  physical damage-type label if the existing combat lab emits that type.
- `tools/bot/scenarios/client/33_unique_burn_effect_live.json` can prove fire damage text from the
  existing Everburning Wound unique effect.

## Test And Bot Proof

Focused checks:

```bash
GODOT=godot ./scripts/client_smoke.sh
```

Targeted slice proof while iterating:

```bash
CLIENT_UNIT_ONLY=1 ./scripts/client_smoke.sh
make bot-visual scenario=33_unique_burn_effect_live HEADLESS=1
```

Visual verification command for humans/agents:

```bash
make bot-visual scenario=33_unique_burn_effect_live
```

## Asset And Plugin Decision

- Adopt: existing authoritative combat event `damage_type` fields and current floating combat text
  presentation.
- Borrow: existing client bot damage-number debug state and combat feedback scenarios.
- Reject: external icons, new art assets, new plugins, new protocol fields, and a new combat text
  rendering system.

## Open Questions And Risks

- Some older events may omit `damage_type`; those must keep the previous numeric presentation.
- Combat lab fixtures may not expose every damage type in one client scenario. Focused unit coverage
  should cover the mapping, and bot proof should cover at least one non-physical live type.
