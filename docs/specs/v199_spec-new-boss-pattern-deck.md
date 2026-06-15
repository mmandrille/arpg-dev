# v199 Spec - New Boss Pattern Deck

Status: Approved for planning
Date: 2026-06-15
Codename: new-boss-pattern-deck

## Purpose

Expand the Cave Warden encounter with one more server-authored spatial attack pattern so the boss
deck exercises a broader set of readable threats before repeating. The new pattern should use
data-driven boss rules, deterministic deck order, existing boss phase events, and authoritative
server hit detection.

## Non-goals

- No new boss template, arena, monster family, asset, animation, sound, or bespoke VFX.
- No random or weighted pattern selection; the deck remains deterministic.
- No protocol version bump or new event type; existing boss phase metadata remains the surface.
- No client-only damage or telegraph logic.
- No boss tuning overhaul beyond the new pattern's data.

## Acceptance Criteria

- `shared/rules/boss_patterns.v0.json` defines a `shard_fan` pattern with telegraph, active cone
  damage, recovery, cooldown, range, cone width, and damage in data.
- `cave_warden.pattern_deck` includes `shard_fan` after `summon_wolves` and before `ground_slam`.
- Existing summon recovery/cooldown pacing stays rule-owned and is short enough that the expanded
  boss-floor proof remains inside the protocol bot scenario budget.
- Boss pattern validation requires cone telegraphs/active phases to provide positive width and
  rejects mismatched active hit predicates.
- The Go sim captures cone aim at telegraph start and applies authoritative active cone damage
  using range, angular width, and player radius.
- Focused Go tests prove cone hit/miss boundaries, missing/mismatched width validation, and the
  expanded deterministic deck order.
- The boss-floor protocol bot observes `shard_fan` before killing the boss.
- Existing boss ranged, summon, movement, locked-exit, kill, and client boss readability checks
  remain green.

## Scope and Files Likely Touched

- Shared rules: `shared/rules/boss_patterns.v0.json`, `shared/rules/boss_templates.v0.json`.
- Server sim/tests: `server/internal/game/boss_patterns.go`,
  `server/internal/game/boss_pattern_rules.go`, focused boss pattern tests, and deck-order tests.
- Bot: `tools/bot/scenarios/24_boss_floor_gate.json`.
- Docs: spec, plan, as-built, and `PROGRESS.md`.

## Test and Bot Proof

- `make maintainability`
- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestBoss(PatternDeckCycles|ShardFan|StoneLance|SummonedAdds|GroundSlam|PhaseTimingAndDodge|FloorExitsUnlock)' -count=1`
- `make bot scenario=24_boss_floor_gate.json`
- `make bot-client scenario=28_boss_phase_readability.json HEADLESS=1`
- Final `make ci`

## Open Questions and Risks

- No blocking questions.
- Risk: adding one more deck entry can stretch the boss-floor bot runtime. Keep the new pattern
  before `ground_slam` and use bounded waits before the kill step.
- Risk: cone width semantics must be explicit. This slice treats `width` as degrees for cone phases.
