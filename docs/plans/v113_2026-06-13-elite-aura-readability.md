# v113 Plan — Elite Aura Readability

Status: Ready for implementation
Goal: Show server-owned elite command aura state on buffed generated-pack follower monsters.
Architecture: Reuse the existing `effect_ids` entity field instead of adding a protocol version.
The server computes active aura state from v112 pack metadata/rules; the client only renders a
display marker from received ids.
Tech stack: Go sim, Godot client presentation, client bot, lifecycle docs.

## Baseline and shortcut decision

`player_status_effect_markers.gd` patterns.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/game/elite_aura.go` | Expose active aura ids for entity views. |
| Modify | `server/internal/game/sim.go` | Include aura effect ids in monster views. |
| Modify | `server/internal/game/elite_aura_test.go` | Prove view effect id appears/disappears. |
| Modify | `client/scripts/player_status_effect_markers.gd` | Add elite command marker helper. |
| Modify | `client/scripts/main.gd` | Sync marker from monster `effect_ids` and expose debug state. |
| Modify | `client/scripts/bot_scenario_runner.gd` | Add assertion support if existing presentation checks are insufficient. |
| Create | `tools/bot/scenarios/client/34_elite_aura_readability.json` | Client bot proof. |
| Modify | `PROGRESS.md` | Mark v113 complete during finish. |
| Create | `docs/as-built/v113_elite-aura-readability.md` | Record shipped behavior and proof. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/sim.go`
- [x] `client/scripts/main.gd`
- [x] `client/scripts/bot_scenario_runner.gd`

Decision:
- [x] Keep over-limit edits surgical and put reusable marker code in the focused marker helper file.
- [x] Run `make maintainability` before final CI.

Verification:
```bash
make maintainability
```

## Task 1 — Server aura effect ids

Files:
- Modify: `server/internal/game/elite_aura.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/elite_aura_test.go`

- [x] Step 1.1: Add helper returning active monster aura effect ids.
- [x] Step 1.2: Include `elite_command` in monster entity views only when v112 aura conditions hold.
- [x] Step 1.3: Cover in focused Go tests.
```bash
cd server && go test ./internal/game -run TestEliteAura -count=1
```

## Task 2 — Client marker presentation

Files:
- Modify: `client/scripts/player_status_effect_markers.gd`
- Modify: `client/scripts/main.gd`

- [x] Step 2.1: Add a small code-native elite command marker and sync helper.
- [x] Step 2.2: Sync marker from entity `effect_ids` for monster records.
- [x] Step 2.3: Expose marker presence in presentation debug state.
```bash
make client-unit
```

## Task 3 — Client bot proof

Files:
- Create: `tools/bot/scenarios/client/34_elite_aura_readability.json`
- Modify: `client/scripts/bot_scenario_runner.gd` if needed

- [x] Step 3.1: Add a dungeon client scenario that descends and waits for an elite command marker.
- [x] Step 3.2: Add/extend assertion support if current presentation matcher cannot assert it.
```bash
make bot-client scenario=34_elite_aura_readability
```

## Task 4 — Lifecycle docs and CI

Files:
- Modify: `docs/plans/v113_2026-06-13-elite-aura-readability.md`
- Modify: `PROGRESS.md`
- Create: `docs/as-built/v113_elite-aura-readability.md`

- [x] Step 4.1: Mark plan tasks complete as they pass.
- [x] Step 4.2: Update `PROGRESS.md` latest slice, next slice, lifecycle row, and recently closed note.
- [x] Step 4.3: Add the v113 as-built note.
```bash
make ci
```

## Final verification

- [x] `cd server && go test ./internal/game -run TestEliteAura -count=1`
- [x] `make client-unit`
- [x] `make bot-client scenario=34_elite_aura_readability`
- [x] `make maintainability`
- [x] `make ci`
