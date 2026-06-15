# v178 Spec — Boss Summoned Adds

Status: Approved for planning
Date: 2026-06-15
Codename: boss-summoned-adds

## Purpose

Give Cave Warden a deterministic summon pattern that adds temporary battlefield pressure beyond
direct boss attacks. The boss should telegraph a summon beat, then spawn a small number of ordinary
server-owned monsters near itself as combat adds.

## Non-goals

- No new monster family, model, animation, sound, or bespoke summon VFX.
- No random or weighted boss pattern selection; deterministic deck order remains.
- No client-only spawned entities; summoned adds are normal authoritative monsters in snapshots and
  deltas.
- No add despawn timer, leash cleanup, special loot table, XP exception, or post-boss cleanup in
  this slice.
- No new client UI; existing entity rendering and boss phase readability are reused.

## Acceptance Criteria

- `shared/rules/boss_patterns.v0.json` defines a `summon_wolves` Cave Warden pattern with
  telegraph, summon active phase, recovery, cooldown, summoned monster id, count, radius, and
  duration in data.
- Cave Warden's deterministic `pattern_deck` includes `summon_wolves` after `stone_lance`, before
  `ground_slam`, so the boss-floor protocol proof observes the summon in a bounded runtime.
- Boss pattern validation accepts summon phases only when they define a known monster, positive
  count, positive spawn radius, and active kind without damage shape requirements.
- The Go sim spawns the requested number of non-boss add monsters exactly once for the active summon
  phase, places them deterministically near the boss on walkable floor positions, and emits normal
  entity-add changes plus a `boss_summoned_adds` event.
- Summoned adds use existing monster behavior, stats, visuals, and entity views; they can aggro,
  chase, take damage, and be counted by existing bot assertions.
- Protocol bot proof observes `summon_wolves`, sees the `boss_summoned_adds` event, and verifies
  at least the configured number of live `dungeon_wolf` adds appear during the boss-floor scenario.
- Existing boss movement, ranged pattern, locked-exit, kill, and client boss readability checks
  remain green.

## Scope and Files Likely Touched

- Shared rules/schemas: `shared/rules/boss_patterns.v0.json`,
  `shared/rules/boss_templates.v0.json`, `shared/rules/boss_patterns.v0.schema.json`.
- Server sim: `server/internal/game/rules.go`, `server/internal/game/boss_patterns.go`,
  `server/internal/game/types.go` if event metadata needs an additive field.
- Server tests: focused summon validation/spawn coverage near existing boss pattern tests.
- Bot scenario: `tools/bot/scenarios/24_boss_floor_gate.json`.
- Docs: this spec, matching plan, as-built notes, and `PROGRESS.md`.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game/... -run 'TestBoss(PatternDeckCycles|SummonedAdds|StoneLance|PhaseTimingAndDodge|FloorExitsUnlock)' -count=1`
- `make bot scenario=24_boss_floor_gate.json`
- `make bot-client scenario=28_boss_phase_readability.json HEADLESS=1`
- `make maintainability`
- Final `make ci`

## Open Questions and Risks

- No blocking questions.
- Risk: adding another boss deck phase can stretch the boss-floor scenario. Keep `summon_wolves`
  second or third in the deck and use bounded wait steps before the boss kill.
- Risk: spawn placement near the boss must avoid walls and blocked tiles without introducing
  nondeterministic retries.
