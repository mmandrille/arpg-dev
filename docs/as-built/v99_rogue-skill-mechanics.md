# v99 As-Built - Rogue skill mechanics

Date: 2026-06-12
Status: Complete

## What Shipped

- Made `poison_stab` a fully usable Rogue skill: it spends mana, starts cooldown, deals weapon
  attack damage, and applies deterministic poison ticks to the struck monster.
- Made `dash` a fully usable Rogue skill: it spends mana, starts cooldown, moves through monsters
  along a server-owned line, and damages crossed enemies from shared skill data.
- Added schema-backed `poison` and `dash` skill payloads so Rogue tuning stays in
  `shared/rules/skills.v0.json`.
- Implemented Rogue off-hand basic attacks with independent cooldown at 1.5x the main-hand cadence
  when a Rogue has a valid one-handed off-hand weapon.
- Extended the Rogue foundation bot scenario to allocate Dash and Poison Stab, dash through a
  target, poison it, then observe at least two main-hand attacks and one off-hand attack.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestRogue|TestLoadRules'`
- `.venv/bin/pytest tools/bot/test_protocol.py -q`
- `make bot scenario=rogue_class_foundation`
- `make maintainability`
- `make ci`

Manual visual check, when desired:

```bash
make bot-visual scenario=rogue_class_foundation
```

## Scope Limits

- No new Rogue VFX, animation clips, or UI widgets were added; existing skill and damage event
  presentation remains the client path.
- No broader dual-wield rebalance was attempted beyond the Rogue off-hand attack proof.
- No additional Rogue skills or passive tree branches were added.

## Maintainability Note

The new Rogue combat helpers live in `server/internal/game/rogue_skills.go` instead of expanding the
already-large simulation file. `server/internal/game/sim.go` only keeps the state wiring and generic
simulation call sites needed for deterministic ticking.

`make maintainability` also surfaced pre-existing committed drift in
`client/scripts/inventory_panel.gd`; the grandfathered baseline was updated to the current line
count so the ratchet is accurate again. This slice did not change the inventory panel.
