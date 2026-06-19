# v282 Spec: Rectangular Boss Pattern

Status: Complete
Date: 2026-06-19
Codename: rectangular-boss-pattern

## Purpose

Add another Cave Warden attack pattern that is meaningfully different from the existing melee,
circle, line, cone, and summon patterns. The new `crystal_wall` pattern uses the already-declared
`rectangle` boss shape, with server hit detection and client telegraph rendering support.

## Non-goals

- No boss HP or summon tuning; the selected boss HP change ships in the next slice.
- No new boss model, animation, or asset pipeline.
- No protocol version bump; existing `hit_shape`/`width` metadata already covers the pattern.

## Asset Decision

- Adopt: existing boss telegraph marker path and `rectangle` schema vocabulary.
- Borrow: line/cone aimed-pattern scheduling and hit-testing structure.
- Reject: adding an external VFX asset or a one-off client-only visual pattern.

## Acceptance Criteria

- `crystal_wall` exists in `shared/rules/boss_patterns.v0.json` with telegraph, active, and recovery
  phases using `rectangle` hit/active shapes.
- `cave_warden` includes `crystal_wall` in its `pattern_deck`.
- Boss validation requires positive `width` for rectangle telegraphs and active phases match the
  telegraph predicate.
- Server hit detection applies rectangle range/width against the boss's locked aim.
- Client boss telegraph marker renders rectangle shapes as rectangular decals rather than circles.
- Focused server/client tests prove the pattern rules, hit predicate, validation guard, and marker
  mesh.

## Scope and Likely Files

- Shared data/schema: `shared/rules/boss_patterns.v0.json`,
  `shared/rules/boss_templates.v0.json`.
- Server: `server/internal/game/boss_pattern_rules.go`, `server/internal/game/boss_patterns.go`,
  `server/internal/game/boss_rectangle_pattern_test.go`.
- Client: `client/scripts/boss_visuals_controller.gd`, existing client unit coverage.
- Docs: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`, and as-built summary at finish.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game/...`
- `GODOT=/opt/homebrew/bin/godot make client-unit`

Visual scenario for manual verification:

```bash
make bot-visual scenario=24_boss_floor_gate
```

## Open Questions and Risks

- No blocking questions.
- Risk: a new deck entry can lengthen the boss cycle. The pattern is appended after current bot-proven
  required phases so existing scenario waits remain stable.
