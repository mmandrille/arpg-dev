# Spec: `boss-pattern-variety`

Status: Accepted
Date: 2026-06-10
Codename: `boss-pattern-variety`
Slice: v58 - boss pattern variety
Baseline: v57 `boss-phase-readability`

## Purpose

The first boss floor is readable after v57, but Cave Warden still repeats a single
`charged_melee` pattern. This slice adds a second authoritative boss attack pattern and makes the
boss cycle its pattern deck deterministically, so the fight has a small amount of variety without
changing the protocol shape or moving combat authority to the client.

## Non-goals

- No new boss template, boss floor layout, loot, XP, HP, normal monster tuning, or progression gate
  changes.
- No random pattern picker, weighted deck, enrage phases, adds, ranged boss patterns, or projectile
  boss attacks.
- No protocol schema version bump. Existing `boss_phase_started` / `boss_phase_ended` payload fields
  remain sufficient.
- No Godot UI/art changes beyond existing v57 boss phase/telegraph rendering.
- No production VFX/audio or shape-specific decal art.

## Acceptance criteria

1. `shared/rules/boss_patterns.v0.json` defines a second Cave Warden-compatible pattern,
   `ground_slam`.
2. `ground_slam` is telegraph-first and data-driven:
   - telegraph phase has at least the configured minimum telegraph duration,
   - active damage phase uses the same `circle` hit predicate and radius as its telegraph,
   - recovery and cooldown are declared in rules data.
3. `shared/rules/boss_templates.v0.json` includes both `charged_melee` and `ground_slam` in the
   `cave_warden` pattern deck.
4. The Go sim cycles boss deck entries deterministically in deck order after each full pattern and
   cooldown, starting with `charged_melee`, then `ground_slam`, then repeating.
5. Circle boss active phases damage players inside the authoritative radius and do not damage
   players outside it.
6. Existing boss movement, telegraph-first damage, locked-exit, boss health bar, and v57 phase
   readability behavior remain green.
7. The protocol bot can prove that both boss pattern ids are observed during the first boss floor
   flow.

## Scope and likely files

- `shared/rules/boss_patterns.v0.json` - add `ground_slam`.
- `shared/rules/boss_templates.v0.json` - add `ground_slam` to Cave Warden deck.
- `server/internal/game/sim.go` - deterministic deck cycling and circle hit predicate.
- `server/internal/game/game_test.go` - focused deck-cycle and circle hit tests.
- `tools/bot/bot_types.py` / `tools/bot/run.py` - generic event payload assertions.
- `tools/bot/test_protocol.py` - unit coverage for event payload assertions.
- `tools/bot/scenarios/24_boss_floor_gate.json` - assert both boss pattern ids appear.
- `docs/specs/v58_spec-boss-pattern-variety.md` - this spec.
- `docs/plans/v58_2026-06-10-boss-pattern-variety.md` - implementation plan.
- `PROGRESS.md` and `docs/as-built/v58_boss-pattern-variety.md` - lifecycle close-out.

## Test and bot proof

- `make validate-shared`
- `cd server && go test ./internal/game/... -run 'TestBoss(Pattern|GroundSlam|Phase|Floor)' -count=1`
- `.venv/bin/pytest tools/bot/test_protocol.py -q`
- `make bot scenario=24_boss_floor_gate.json`
- `make bot-client scenario=28_boss_phase_readability.json HEADLESS=1`
- `make ci`

## Open questions and risks

- Q1: Should pattern choice be random?
  - Decision: no for this slice. Deck-order cycling is deterministic, readable, and enough to prove
    multi-pattern infrastructure. Weighted deterministic RNG can be a later balance slice.
- Risk: Existing client marker is a primitive radius indicator, not a shape-specific production decal.
  - Mitigation: `ground_slam` uses a circle radius, which v57 already renders clearly.
