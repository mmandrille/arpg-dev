# v80 Plan - Combat Threat Readability

Status: Ready for implementation
Goal: Make existing aggro and combat outcome events easier to read in the Godot client.
Architecture: Reuse authoritative `monster_aggro` and combat outcome events. The client maps those
events to display-only `DamageNumber` variants and bot debug state; no server or protocol contract
changes are needed.
Tech stack: Godot GDScript client, client bot scenario, existing Python protocol bot regression,
SDD docs.

## Baseline and shortcut decision

Baseline is v79 `elite-pack-roles` on `main`.

Adoption checklist:
- Adopt: existing in-repo `DamageNumber`, floating combat text setting, `get_bot_state()`, and
  client bot damage-number assertions.
- Borrow: existing block/miss/crit text styling pattern in `damage_number.gd`.
- Reject: external plugins/assets. The slice is a narrow presentation mapping over existing
  authoritative events.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/main.gd` | Show threat floating text for `monster_aggro` |
| Modify | `client/scripts/damage_number.gd` | Threat variant sizing/motion |
| Modify | `client/tests/test_coop_client.gd` | Unit proof for aggro text and setting gate |
| Add | `tools/bot/scenarios/client/31_combat_threat_readability.json` | Client-bot proof |
| Add | `docs/specs/v80_spec-combat-threat-readability.md` | Slice spec |
| Add | `docs/plans/v80_2026-06-11-combat-threat-readability.md` | This plan |
| Add | `docs/as-built/v80_combat-threat-readability.md` | As-built proof |
| Modify | `PROGRESS.md` | Lifecycle close-out |

## Task 1 - Threat floating text

Files:
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/damage_number.gd`
- Modify: `client/tests/test_coop_client.gd`

- [x] Step 1.1: Map `monster_aggro` events to `AGGRO` text with variant `threat`.
```bash
make client-unit
```

- [x] Step 1.2: Give the `threat` variant a distinct readable style while preserving existing
  miss/block/crit behavior.
```bash
make client-unit
```

- [x] Step 1.3: Add unit coverage that `monster_aggro` creates one `threat` text item and that
  disabling floating text suppresses it.
```bash
make client-unit
```

## Task 2 - Client bot proof

Files:
- Add: `tools/bot/scenarios/client/31_combat_threat_readability.json`

- [x] Step 2.1: Add a client scenario using the existing pack aggro world and seed.
- [x] Step 2.2: Wait for `monster_aggro`, then assert a `threat` damage number with text `AGGRO`.
- [x] Step 2.3: Run the focused client scenario.
```bash
SCENARIO=31_combat_threat_readability HEADLESS=1 ./scripts/bot_client_local.sh
```

## Task 3 - Regression and lifecycle

Files:
- Modify: `docs/plans/v80_2026-06-11-combat-threat-readability.md`
- Modify: `PROGRESS.md`
- Add: `docs/as-built/v80_combat-threat-readability.md`

- [x] Step 3.1: Run the protocol pack aggro regression.
```bash
ARPG_ADDR=:8888 SCENARIO=pack_aggro_and_dungeon_packs ./scripts/bot_local.sh
```

- [x] Step 3.2: Update lifecycle docs and as-built summary.
- [x] Step 3.3: Run full CI.
```bash
make ci
```

## Final verification

- [x] `make client-unit`
- [x] `SCENARIO=31_combat_threat_readability HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `ARPG_ADDR=:8888 SCENARIO=pack_aggro_and_dungeon_packs ./scripts/bot_local.sh`
- [x] `make ci`

## Deferred scope

- Threat markers, target indicators, sound, minimap danger hints, and richer elite leader
  readability.
