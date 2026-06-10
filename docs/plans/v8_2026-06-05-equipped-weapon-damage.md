# Equipped Weapon Damage (Slice v8) - Implementation Plan

Status: Complete (`make ci` green on 2026-06-05)

Goal: Make equipped weapons change authoritative player attack damage while keeping
the unarmed vertical-slice path, protocol, replay, and client UI unchanged.

Architecture: Weapon damage is rules-as-data on item definitions. The Go sim
resolves the equipped weapon at attack time and chooses either the item damage
range or `combat.player_damage`; Go and GDScript golden tests prove the shared
range formula stays aligned.

Tech stack: shared JSON rules + JSON Schema + Python validator, Go sim tests,
Python protocol bot scenarios, Godot headless golden smoke.

Spec: [`docs/specs/v8_spec-equipped-weapon-damage.md`](../specs/v8_spec-equipped-weapon-damage.md)
Baseline: slice v7 `gear-before-combat-scenario` (complete; `make ci` green)
Branch: `feature/equipped-weapon-damage`

## Evaluation

- The spec is implementable without a protocol schema bump. `state_delta`
  already carries monster HP updates and kill events; inventory/equipped state
  is already reconstructed through recorded inputs.
- The narrow server hook is `server/internal/game/sim.go::rollDamage()`.
  Equipped state already exists as `map[slot]instanceID`, and inventory items
  retain their `item_def_id`, so no new persistence field is needed.
- `items.v0.schema.json` already uses conditional slot validation from v7.
  Add `damage` there and mirror the same checks in `LoadRules`; do not rely on
  schema validation alone at server boot.
- The bot's current final-state assertions cannot prove "killed in one attack".
  That must be recorded as a runtime observation during `attack_until_event`,
  keyed by accepted attack message IDs, then checked after the scenario run.
- `intent_accepted` is a separate WebSocket envelope, not part of
  `state_delta`; the bot must ingest it while pumping messages.
- No Godot UI or asset work is required. Per the plugins checklist:
  **reject plugin adoption for v8** because this slice changes authoritative
  combat math and golden validation only; GLoot/Godot-Inventory remain future
  display-only candidates.

## Decisions

| Topic | Decision |
|-------|----------|
| Damage composition | Replace base range with equipped weapon `damage` when present. |
| Rusty sword tuning | `damage: { "min": 3, "max": 5 }`, so 3 HP dummies die in one successful hit. |
| Fallback path | No weapon, missing equipped item, non-weapon slot, or no item damage falls back to `combat.player_damage`. |
| Attack counting | Count acknowledged `attack_intent` messages sent by the `attack_until_event` step for the selected `monster_def_id`; with v8 hit chance `1.0`, this equals successful attacks. |
| Runtime vs final assertions | `monster_killed_in_attacks` is runtime-only; `/state`, reconnect, and replay keep final-state assertions. |
| Client plugins | Reject for this slice; no inventory UI, drag/drop, tooltip, or addon integration. |

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/items.v0.schema.json` | Allow optional weapon-only `damage` range |
| Modify | `shared/rules/items.v0.json` | Add `rusty_sword.damage` |
| Create | `shared/golden/equipped_weapon_damage.v0.schema.json` | Golden schema for weapon damage formula |
| Create | `shared/golden/equipped_weapon_damage.json` | Cross-language rusty sword damage cases |
| Modify | `tools/validate_shared.py` | Validate weapon damage cross-checks and new golden |
| Modify | `server/internal/game/rules.go` | Add `ItemDef.Damage`; validate weapon-only ranges |
| Modify | `server/internal/game/sim.go` | Resolve equipped weapon damage in `rollDamage` |
| Modify | `server/internal/game/game_test.go` | Golden and equipped/unarmed sim coverage |
| Modify | `client/tests/test_golden.gd` | Consume equipped weapon golden |
| Modify | `tools/bot/scenarios/02_gear_before_combat.json` | Add one-attack kill assertion |
| Modify | `tools/bot/run.py` | Track accepted attack counts and runtime assertions |
| Modify | `tools/bot/test_protocol.py` | Unit tests for attack-count observation/assertion |
| Modify | `PROGRESS.md` | Update only when v8 ships |

No expected changes: protocol schemas, store migrations, replay package,
realtime protocol, Godot `main.gd`, inventory UI, asset manifests, worlds.

## Task 1: Shared Rules and Golden Data

Files:
- Modify: `shared/rules/items.v0.schema.json`, `shared/rules/items.v0.json`
- Create: `shared/golden/equipped_weapon_damage.v0.schema.json`
- Create: `shared/golden/equipped_weapon_damage.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Extend item schema with a reusable `damage_range` object:
  required integer `min` and `max`, both `minimum: 0`.
- [x] Step 1.2: Add conditionals:
  - `equippable: true` still requires `slot`
  - `equippable: false` forbids `slot` and `damage`
  - `damage` is only valid for `equippable: true` with `slot: "weapon"`
- [x] Step 1.3: Add `damage: { "min": 3, "max": 5 }` to `rusty_sword`;
  leave `training_badge` statless.
- [x] Step 1.4: Add `equipped_weapon_damage.json` with `item_def_id:
  "rusty_sword"`, copied damage range, and draw cases `0..3 -> 3,4,5,3`.
- [x] Step 1.5: Add a golden schema requiring `description`, `item_def_id`,
  `damage`, and non-empty `cases`.
- [x] Step 1.6: Extend `validate_shared.py`:
  - load `equipped_weapon_damage.json`
  - verify `item_def_id` resolves
  - verify the referenced item is equippable in the weapon slot
  - verify the golden `damage` equals the item rule damage
  - verify all cases satisfy `min + (draw mod span)`
  - fail if any non-weapon or non-equippable item declares `damage`

Verification:

```bash
make validate-shared
```

## Task 2: Server Rules Loader and Damage Resolution

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/sim.go`

- [x] Step 2.1: Add a `Damage *DamageRange` field to `ItemDef`, tagged as
  `json:"damage,omitempty"`.
- [x] Step 2.2: In `LoadRules`, validate item damage:
  - if `Damage != nil`, item must be `Equippable` and `Slot == weaponSlot`
  - call `validateDamageRange("items.<id>.damage", *def.Damage)`
  - preserve existing non-equippable slot rejection
- [x] Step 2.3: Add `resolvePlayerAttackDamage() DamageRange` on `Sim`:
  - get `instanceID := s.equipped[weaponSlot]`
  - if `instanceID == 0`, return `s.rules.Combat.PlayerDamage`
  - find the inventory item; if missing, return base damage
  - look up the item definition; if missing or no `Damage`, return base damage
  - return `*def.Damage`
- [x] Step 2.4: Change `rollDamage()` to call
  `s.rollRange(s.resolvePlayerAttackDamage())`.
- [x] Step 2.5: Do not change RNG draw order in `handleAttack`: hit draw first,
  then one damage draw, then existing damage/retaliation flow.

Verification:

```bash
cd server && go test ./internal/game -run 'LoadRules|Damage|Equipped|Slice'
```

## Task 3: Go Sim Tests

Files:
- Modify: `server/internal/game/game_test.go`

- [x] Step 3.1: Update `TestLoadRules` to assert
  `rusty_sword.Damage == {3,5}` and `training_badge.Damage == nil`.
- [x] Step 3.2: Add `TestEquippedWeaponDamageGolden` consuming
  `shared/golden/equipped_weapon_damage.json`.
- [x] Step 3.3: Add `TestEquippedWeaponOneShotsRewardDummy`:
  - build `NewSimWithWorld(..., "gear_before_combat")`
  - pick up initial `rusty_sword`
  - equip it
  - attack `training_dummy_reward` once
  - assert ack, `monster_damaged`, `monster_killed`, loot spawn, and HP `0`
- [x] Step 3.4: Add unarmed fallback coverage by cloning loaded rules in-memory:
  set `rusty_sword.Damage = nil`, run the same gear flow with a seed/draw where
  base damage can be `2`, and assert the reward dummy survives one hit.
- [x] Step 3.5: Keep `TestScriptedSliceMatchesGolden` unchanged except for any
  helper reuse; it must still kill before equip and preserve `final_player_hp: 9`.

Verification:

```bash
cd server && go test ./internal/game/... -run 'Golden|Slice|Equipped|Weapon|LoadRules'
```

## Task 4: GDScript Golden Coverage

Files:
- Modify: `client/tests/test_golden.gd`

- [x] Step 4.1: Load `rules/items.v0.json` and
  `golden/equipped_weapon_damage.json`.
- [x] Step 4.2: Assert the golden item exists, is a weapon, and its `damage`
  range matches the item definition.
- [x] Step 4.3: Reuse the existing range formula check for the new cases.
- [x] Step 4.4: Update the success log to include `equipped_weapon_damage`.

Verification:

```bash
make client-smoke
```

## Task 5: Bot Runtime One-Attack Assertion

Files:
- Modify: `tools/bot/scenarios/02_gear_before_combat.json`
- Modify: `tools/bot/run.py`
- Modify: `tools/bot/test_protocol.py`

- [x] Step 5.1: Add this assertion to `02_gear_before_combat.json`:

```json
{ "type": "monster_killed_in_attacks", "monster_def_id": "training_dummy_reward", "max_attacks": 1 }
```

- [x] Step 5.2: Extend `RuntimeState` with runtime observation fields:
  pending attack message id -> monster def id, accepted attack counts by monster
  def id, and killed entity ids if useful for diagnostics.
- [x] Step 5.3: When `attack_until_event` sends an attack, construct the
  envelope first, record its `message_id` with the selected `monster_def_id`,
  then send it.
- [x] Step 5.4: Update `ingest_message` to handle:
  - `intent_accepted`: if `accepted_message_id` is a pending attack, increment
    the count for that monster def id
  - `intent_rejected`: if it matches a pending attack, raise or record a clear
    failure instead of counting it
  - `state_delta`: existing entity/event ingestion remains unchanged
- [x] Step 5.5: Add a runtime assertion pass after `drive_scenario` and before
  `/state` checks. It evaluates only assertion types that need runtime history,
  currently `monster_killed_in_attacks`.
- [x] Step 5.6: Keep `run_assertions` final-state-only. Either ignore runtime
  assertion types there or split assertion lists before calling it, so `/state`
  and reconnect do not fail on observations they cannot derive.
- [x] Step 5.7: Add Python unit tests:
  - `intent_accepted` increments attack count for pending attack
  - `monster_killed_in_attacks` passes at `1 <= max_attacks`
  - it fails with a clear message when count exceeds max
  - scenario loader accepts the updated gear scenario

Verification:

```bash
.venv/bin/python -m pytest tools/bot/test_protocol.py -q -k 'scenario or assertion or attack'
```

## Task 6: End-to-End Verification and Docs

Files:
- Modify after implementation is complete: `PROGRESS.md`

- [x] Step 6.1: Run focused shared + Go + Python checks.
- [x] Step 6.2: Run `make bot` to prove both scenarios, `/state`, reconnect,
  and replay verification still pass.
- [x] Step 6.3: Run `make ci`.
- [x] Step 6.4: Update `PROGRESS.md` only after the slice passes:
  mark v8 complete, add summary, and keep deferred UI/plugin work in backlog.

Final verification:

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'Golden|Slice|Equipped|Weapon|LoadRules'
.venv/bin/python -m pytest tools/bot/test_protocol.py -q -k 'scenario or assertion or attack'
make bot
make ci
```

Optional visual inspection:

```bash
make bot-visual
```

## Questions

No blocking questions remain before implementation. I will proceed with these
defaults unless the product direction changes:

| # | Question | Default |
|---|----------|---------|
| 1 | Should weapon damage replace base damage instead of adding to it? | Yes, replace. |
| 2 | Should the one-attack bot assertion count sent attacks or acknowledged attacks? | Acknowledged `attack_intent` messages. |
| 3 | Should runtime-only assertions run against `/state` and reconnect snapshots? | No; final-state assertions continue there. |
| 4 | Should v8 adopt any Godot inventory plugin? | No; explicitly reject for this slice. |

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Existing vertical-slice golden changes unexpectedly | Keep resolver dependent on equipped state at attack time; vertical slice attacks before equip. |
| Replay drift from RNG stream changes | Preserve exactly one hit draw and one damage draw per accepted attack. |
| Rules drift between schema and server loader | Implement both JSON Schema conditionals and Go `LoadRules` validation. |
| Bot one-hit assertion cannot be checked from final state | Split runtime assertions from final-state assertions and store accepted attack counts during WebSocket play. |
| Future non-weapon items accidentally get damage | Schema, Python cross-check, and Go loader all reject it. |
