# v99 Spec — Rogue Skill Mechanics

Status: Draft
Date: 2026-06-12
Codename: `rogue-skill-mechanics`

## Purpose

Make Rogue's two starter skills fully usable through the authoritative skill pipeline.
`Poison Stab` should deal the Rogue's weapon hit and apply damage over time. `Dash` should move
the Rogue through enemies for a short line, damage enemies crossed by the dash, and work with the
existing skill rank, mana, cooldown, bot, replay, and client event paths.

## Non-goals

- No protocol schema version bump unless an existing event payload cannot carry the required facts.
- No class rebalance beyond the skill-specific data required by this slice.
- No multi-target off-hand combo skill; dual-wield base attacks remain the v98/v99 combat behavior.

## Acceptance Criteria

1. `Poison Stab` is a Rogue-only rankable skill that costs mana, starts cooldown, deals base weapon
   attack damage immediately, and applies poison ticks to the struck monster.
2. Poison DOT deals a rank-scaled percent of the original hit per second for a data-driven duration.
   Rank increases percent. Magic increases duration.
3. `Dash` is a Rogue-only rankable skill that costs mana, starts cooldown, moves the player up to
   the configured range through monsters, and damages enemies crossed by the dash.
4. Dash rank increases range and lowers cooldown through shared skill data. Magic increases damage
   percent.
5. Skill outcomes remain server-authoritative and deterministic: no wall-clock time, no unseeded
   randomness, stable entity ordering.
6. Client presentation consumes server events for dash and poison; it does not decide hit, movement,
   DOT, or damage outcomes.
7. `tools/bot/scenarios/47_rogue_class_foundation.json` creates a Rogue, uses Dash through an enemy,
   applies Poison Stab, and observes at least three Rogue attacks against the same target, with at
   least two main-hand hits and one off-hand hit.
8. Shared validation, focused Go skill tests, bot tests, and `make ci` pass.

## Scope And Likely Files

- `shared/rules/skills.v0.schema.json` — add closed payloads for poison DOT and dash movement.
- `shared/rules/skills.v0.json` — tune `poison_stab` and `dash` with data-owned percentages,
  durations, range, cooldown, and damage settings.
- `server/internal/game/rules.go` — load and validate the new skill payloads.
- `server/internal/game/handlers.go`, `server/internal/game/sim.go` — authoritative cast, dash
  movement, damage, poison tick scheduling, event emission.
- `server/internal/game/game_test.go` or focused helper test file — deterministic unit coverage.
- `tools/bot/run.py`, `tools/bot/test_protocol.py` — scenario action/assertion support if needed.
- `tools/bot/scenarios/47_rogue_class_foundation.json` — extend the Rogue proof.
- `client/scripts/main.gd` and tests if presentation state needs a new event mapping.
- `PROGRESS.md`, `docs/plans/v99_2026-06-12-rogue-skill-mechanics.md`,
  `docs/as-built/v99_rogue-skill-mechanics.md`.

## Test And Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestRogue'`
- `.venv/bin/pytest tools/bot/test_protocol.py -q`
- `make bot scenario=rogue_class_foundation`
- `make client-unit` if client event handling changes
- `make maintainability`
- `make ci`

Manual visual check, when desired:

```bash
make bot-visual scenario=rogue_class_foundation
```

## Open Questions And Risks

- Large touched files are already over the ratchet baseline (`server/internal/game/sim.go`,
  `server/internal/game/game_test.go`, `tools/bot/run.py`, `client/scripts/main.gd`). Plan must
  either keep growth minimal or extract small helpers where practical.
- Dash movement needs collision-safe endpoint resolution. The conservative default is to reuse
  existing movement resolution and stop at the farthest valid point on the dash line.
- Poison tick cadence uses the server's 10 Hz tick rate; the visible behavior should be one tick
  per second for the configured duration.
