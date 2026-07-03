# Spec: `class-build-pacing`

Status: Approved for implementation  
Date: 2026-07-03  
Codename: class-build-pacing  
Slice: v412 — class build pacing and seven skills

## Purpose

Tighten class identity and expand build variety by:

1. Changing skill-point grants to levels **1, 3, 6, 9, …** (level 1 plus every multiple of 3).
2. Granting automatic per-class base stat growth on each level-up (barbarian +STR, sorcerer +Magic,
   paladin +VIT, ranger/rogue +DEX).
3. Adding **seven** new active skills (two ranger, two sorcerer, one each barbarian/paladin/rogue)
   using existing skill kinds, plus `evade_chance` support on cast-time `area_stat_buff`.

Builds on v411 skill-tree graph layout.

## Non-goals

- Rend, Retribution, Predator's Mark (v413).
- Persistent ground zones (Smoke Screen is cast-time ally buff for 10s).
- Protocol/schema version bump.
- Production VFX/audio (borrow existing projectile/cone/buff presentations).
- Full combat balance pass.

## Acceptance criteria

- [ ] `character_progression.v0.json` skill cadence grants at 1, 3, 6, 9…; Go unit test owns formula.
- [ ] Each class gains configured `level_stat_growth` on level-up; respec restores class base + total growth for level.
- [ ] Existing characters resume with growth floor applied (no stat loss).
- [ ] Seven skills in `skills.v0.json` with tree placement, prerequisites, class lock, i18n + presentations.
- [ ] `area_stat_percent_buff` allows `evade_chance`; Smoke Screen grants allies+caster +30% +5%/rank evade for 100 ticks.
- [ ] War Cry buffs allies + caster (area stat buff).
- [ ] Extended bot scenarios cast each new skill; progression scenario proves cadence + growth.
- [ ] `make validate-shared` and focused Go tests green.

## Skills

| ID | Class | Kind | Tier/Col | Prereq |
|----|-------|------|----------|--------|
| `rain_of_arrows` | ranger | projectile + volley | T3 c1 | `volley` |
| `explosive_shot` | ranger | projectile | T3 c4 | `snipe` |
| `fireball` | sorcerer | projectile (fire) | T3 c1 | `ice_shard` |
| `energy_ward` | sorcerer | self_buff | T3 c3 | `teleport` |
| `war_cry` | barbarian | area_stat_buff | T3 c4 | `leap` |
| `hammer_of_light` | paladin | cone_attack | T3 c2 | `charge` |
| `smoke_screen` | rogue | area_stat_buff | T3 c4 | `shadowstep` |

## Scope and files

- `shared/rules/character_progression.v0.json` + schema
- `shared/rules/skills.v0.json` + schema (`evade_chance` on area buff)
- `shared/assets/skill_presentations.v0.json`, `shared/i18n/en.json`
- `server/internal/game/sim.go`, `rules.go`, `rules_area_buffs.go`, `skill_buffs.go`
- `server/internal/game/*_test.go`
- `tools/bot/scenarios/` — extended labs
- `docs/as-built/v412_class-build-pacing.md`

## Test and bot proof

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestSkillPoint|TestClassLevelStatGrowth|TestClassBuildPacingSkills' -count=1
make bot scenario=class_progression_pacing
make bot scenario=class_build_pacing_skills
```

## Open questions

None — batch clarification resolved (War Cry allies+caster; Smoke cast-time 10s; pacing defaults).

## Client asset decision

**Borrow** existing projectile, cone, and holy-shield-style buff presentations; no new plugins.
