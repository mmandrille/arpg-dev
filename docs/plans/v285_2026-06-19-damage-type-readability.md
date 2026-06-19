# v285 Plan — Damage Type Readability

Status: Implemented
Goal: Make existing client floating combat text visibly communicate authoritative combat
`damage_type` without changing combat math or protocol.
Architecture: Keep all behavior client-side. Derive label, color, and variant from the event's
existing `damage_type` field in `client/scripts/main.gd`, pass the type into `DamageNumber`, and
extend bot debug matching so scenarios can assert the rendered type.
Tech stack: Godot GDScript client, existing client bot scenarios, existing client unit smoke gate.

## Baseline and shortcut decision

Builds on v100 damage types and resistances plus the existing v11 combat feedback and v33 unique
burn client scenarios. No server data, shared schema, or tuning data is changing because this slice
only renders metadata already present in combat events.

Asset/plugin decision:

- Adopt: current combat event payloads, floating combat text, settings toggle, and client bot
  damage-number debug state.
- Borrow: existing poison presentation color conventions where they already map cleanly to poison
  damage.
- Reject: external assets, icon fonts, new render layers, server event changes, and balance tuning.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/damage_number.gd` | Store `combat_damage_type` alongside text/variant for debug assertions. |
| Create | `client/scripts/damage_type_combat_text.gd` | Own damage-type and special-outcome floating-text presentation mapping. |
| Modify | `client/scripts/main.gd` | Route event `damage_type` presentation into damage numbers. |
| Modify | `client/scripts/bot_scenario_runner.gd` | Match `damage_type` on damage-number expectations. |
| Modify | `client/tests/test_rogue_presentation.gd` | Add unit proof for poison/fire/crit damage-type text behavior. |
| Modify | `tools/bot/scenarios/client/33_unique_burn_effect_live.json` | Assert the fire damage number produced by Everburning Wound. |
| Create during finish | `docs/as-built/v285_damage-type-readability.md` | Record implementation proof and commands. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:

- [x] `client/scripts/main.gd` — already a large coordinator. Keep the change limited to the
  existing combat-text helper and bot debug state.
- [x] `client/scripts/bot_scenario_runner.gd` — touch only damage-number matching.

Decision:

- [x] Extract helper/module: `client/scripts/damage_type_combat_text.gd` owns the mapping so
  `main.gd` stays under the grandfathered file-size allowance.
- [x] Defer extraction with rationale: broader combat-presentation routing cleanup remains future
  work; only the type-label mapping moved out.

Verification:

```bash
make maintainability
```

## Task 1 — Damage-number metadata

Files:

- Modify: `client/scripts/damage_number.gd`
- Modify: `client/scripts/main.gd`

- [x] Step 1.1: Add a `combat_damage_type` string to `DamageNumber`.
- [x] Step 1.2: Add a small mapping for `physical`, `fire`, `cold`, `lightning`,
  `poison`, and `force` to label/color/variant.
- [x] Step 1.3: Use `damage_type` instead of `skill_id == poison_stab` for poison hit color.
- [x] Step 1.4: Keep miss/block/immune priority unchanged.
- [x] Step 1.5: Keep critical variant/scale while formatting critical hit text with the type label.

Verify:

```bash
CLIENT_UNIT_ONLY=1 ./scripts/client_smoke.sh
```

## Task 2 — Bot assertion surface

Files:

- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`

- [x] Step 2.1: Include `damage_type` in `_bot_damage_numbers()`.
- [x] Step 2.2: Allow `wait_damage_number` and `assert_damage_number` to filter by damage type.
- [x] Step 2.3: Keep existing text/variant assertions working.

Verify:

```bash
CLIENT_UNIT_ONLY=1 ./scripts/client_smoke.sh
```

## Task 3 — Focused unit coverage

Files:

- Modify: `client/tests/test_rogue_presentation.gd`

- [x] Step 3.1: Update the existing poison floating text test to pass `damage_type: poison` and
  assert the type label and debug metadata.
- [x] Step 3.2: Add a fire damage-type test that proves a normal hit renders fire-specific text,
  color, variant, and metadata.
- [x] Step 3.3: Add a critical damage-type test that proves critical display survives while carrying
  the type label and metadata.

Verify:

```bash
GODOT=godot godot --headless --path client --script res://tests/test_rogue_presentation.gd
```

## Task 4 — Client bot proof

Files:

- Modify: `tools/bot/scenarios/client/33_unique_burn_effect_live.json`

- [x] Step 4.1: After the existing fire damage event assertion, wait for a damage number with
  `damage_type: fire`.
- [x] Step 4.2: Keep the existing burning-effect proof intact.

Verify:

```bash
make bot-visual scenario=33_unique_burn_effect_live HEADLESS=1
```

## Task 5 — Docs and lifecycle

Files:

- Existing: `docs/specs/v285_spec-damage-type-readability.md`
- Existing: `docs/plans/v285_2026-06-19-damage-type-readability.md`
- Create during finish: `docs/as-built/v285_damage-type-readability.md`
- Modify during finish: `PROGRESS.md`

- [x] Step 5.1: Record focused test and bot proof in the as-built note during finish.
- [x] Step 5.2: Update lifecycle/current status in `/finish`.

## Task 6 — Final verification

- [x] `godot --headless --path client --script res://tests/test_rogue_presentation.gd`
- [x] `CLIENT_UNIT_ONLY=1 ./scripts/client_smoke.sh`
- [x] `make bot-visual scenario=33_unique_burn_effect_live HEADLESS=1`
- [x] `make maintainability`
