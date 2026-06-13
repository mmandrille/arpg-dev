# v123 Spec - Ranger Piercing And Pinning Shots

Status: Draft
Date: 2026-06-13
Codename: `ranger-piercing-and-pinning-shots`

## Purpose

Make Ranger's first two active bow skills fully usable through the authoritative skill pipeline.
`Piercing Shot` should fire a physical arrow that can pass through multiple monsters in a line.
`Pinning Shot` should fire a physical arrow that damages the first monster hit and roots it briefly.
Both skills must use shared skill data for costs, cooldowns, projectile tuning, pierce/root behavior,
and visual metadata.

## Non-goals

- No third Ranger skill; `Volley` is reserved for v124.
- No production animation set or external Godot plugin dependency.
- No new ground-targeting intent shape beyond the existing direction/target cast payload.
- No broad class balance pass outside the new skill data.
- No PvP behavior.

## Acceptance Criteria

1. `shared/rules/skills.v0.json` contains Ranger-only `piercing_shot` and `pinning_shot` skills
   with data-owned mana cost, cooldown, rank requirements, projectile range/speed/visual, physical
   damage, and per-skill mechanics.
2. Skill schemas and Go rule validation reject malformed pierce/root payloads.
3. `Piercing Shot` is server-authoritative and deterministic: it spawns/fires one physical arrow
   along a cast direction or target vector, damages every live monster intersected along its path up
   to a data-owned maximum hit count, emits one damage event per hit, and does not let the client
   choose hits.
4. `Pinning Shot` is server-authoritative and deterministic: it damages the first hit monster,
   applies a root effect for data-owned ticks, blocks rooted monster movement while the effect is
   active, emits start/end effect events, and clears the root from snapshots/state on expiry.
5. Ranger can learn both skills from the skill panel and class skill gates; other classes cannot.
6. Client presentation consumes server events/state only, rendering distinguishable arrow VFX for
   the two skills and a visible root marker/tint on pinned monsters.
7. Protocol bot coverage proves a Ranger can learn/cast both skills, Piercing Shot hits at least
   two lined-up monsters, and Pinning Shot roots a chase monster long enough to prevent movement
   before the root expires.
8. Focused Go tests, shared validation, bot tests, client tests, maintainability, and `make ci` pass.

## Scope And Likely Files

- `shared/rules/skills.v0.schema.json`
- `shared/rules/skills.v0.json`
- `shared/assets/skill_presentations.v0.json`
- `shared/i18n/en.json`
- `shared/i18n/es.json`
- `server/internal/game/rules.go`
- `server/internal/game/ranger_skills.go`
- `server/internal/game/ranger_skills_test.go`
- `server/internal/game/sim.go` and `handlers.go` only for thin dispatch/state hooks
- `client/scripts/projectile_visuals.gd`
- `client/scripts/player_status_effect_markers.gd`
- `client/tests/test_projectile_visuals.gd`
- `client/tests/test_status_effect_presentation.gd`
- `tools/bot/scenarios/59_ranger_piercing_and_pinning_shots.json`
- `tools/bot/test_protocol.py`
- `tools/validate_shared.py`
- `PROGRESS.md`, plan, and as-built docs

## Test And Bot Proof

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestRanger|TestLoadRules'
.venv/bin/pytest tools/bot/test_protocol.py -q
make bot scenario=59_ranger_piercing_and_pinning_shots
make client-unit
make maintainability
make ci
```

Manual visual verification:

```bash
make bot-visual scenario=59_ranger_piercing_and_pinning_shots
```

## Open Questions And Risks

- The current projectile resolver removes skill projectiles on first hit, so Piercing Shot may need
  a focused direct line-resolution helper rather than mutating the generic projectile loop.
- Root behavior must alter monster movement without changing attack cadence or unrelated slow
  mechanics.
- `server/internal/game/sim.go`, `server/internal/game/game_test.go`, `client/scripts/main.gd`, and
  `tools/bot/run.py` are over the maintainability target. This slice should prefer focused helper
  files and only add thin wiring to existing large files.

