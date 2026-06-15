# v177 Spec — Boss Ranged Pattern

Status: Approved for planning
Date: 2026-06-15
Codename: boss-ranged-pattern

## Purpose

Give Cave Warden a server-authored ranged attack so boss fights are not only contact and radial danger checks. The new pattern should telegraph a locked line toward the player, then damage players who remain inside that authoritative line during the active phase.

## Non-goals

- No projectile entity, projectile art, bespoke VFX/audio, or new boss model.
- No random or weighted pattern selection; deterministic deck order remains.
- No new boss template, floor layout, loot, XP, HP, or progression gate changes.
- No client-only combat shortcut; the server owns hit detection and damage.
- No production line-decal rendering. Existing boss phase/telegraph debug presentation may remain generic.

## Acceptance Criteria

- `shared/rules/boss_patterns.v0.json` defines a Cave Warden ranged pattern, `stone_lance`, with telegraph, active, recovery, cooldown, range, width, and damage in rules data.
- Cave Warden's `pattern_deck` includes `stone_lance` after `charged_melee`, before `ground_slam`, so protocol proof observes the ranged pattern within the scenario budget.
- Boss pattern validation accepts line phases only when the active hit predicate matches the prior telegraph shape, range, and width.
- The Go sim captures a deterministic line aim direction at telegraph start and reuses that same direction during the active phase.
- The authoritative line hit predicate damages players inside range and width, and misses players outside the line width or beyond range.
- Existing boss phase events expose additive line/range/width shape metadata through current v8 protocol schemas without removing existing fields.
- Protocol bot proof observes `stone_lance` during the boss-floor scenario while existing boss movement, locked-exit, kill, and client readability checks remain green.

## Scope and Files Likely Touched

- Shared rules/schemas: `shared/rules/boss_patterns.v0.json`, `shared/rules/boss_templates.v0.json`, `shared/rules/boss_patterns.v0.schema.json`.
- Protocol schemas: latest v8 `session_snapshot` and `state_delta` boss telegraph/hit shape definitions if line width/range metadata is emitted.
- Server sim: `server/internal/game/rules.go`, `server/internal/game/types.go`, `server/internal/game/sim.go`.
- Server tests: focused boss pattern, validation, deck order, and line-hit coverage in `server/internal/game/game_test.go` or nearby focused tests.
- Bot scenario: `tools/bot/scenarios/24_boss_floor_gate.json`.
- Docs: this spec, matching plan, as-built notes, and `PROGRESS.md`.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game/... -run 'TestBoss(PatternDeckCycles|StoneLance|PhaseTimingAndDodge|FloorExitsUnlock)' -count=1`
- `make bot scenario=24_boss_floor_gate.json`
- `make bot-client scenario=28_boss_phase_readability.json HEADLESS=1`
- `make maintainability`
- Final `make ci`

## Open Questions and Risks

- No blocking questions.
- Risk: line telegraph presentation remains generic in the current Godot marker. This slice still exposes line metadata and proves server authority; richer line decals are a later presentation slice.
