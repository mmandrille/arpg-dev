# Spec: `class-skill-engine-expansion`

Status: Approved for implementation  
Date: 2026-07-03  
Codename: class-skill-engine-expansion  
Slice: v413 — Rend, Retribution, Predator's Mark

## Purpose

Complete the ten-skill class build variety batch by adding three engine-backed actives deferred from v412:

1. **Rend** — barbarian cone that applies bleed DoT.
2. **Retribution** — paladin self-buff that reflects damage on block.
3. **Predator's Mark** — rogue projectile that applies a damage amp mark.

## Non-goals

- New protocol/schema version bump.
- Production VFX/audio (borrow existing families).
- Additional skills beyond these three.

## Acceptance criteria

- [ ] Shared schema supports `bleed` on skills, `mark` on skills, and `reflect_on_block_buff` self-buff effect.
- [ ] Go sim applies cone bleed, projectile mark, and block-triggered reflect.
- [ ] Three skills in `skills.v0.json` with tree placement, prerequisites, i18n, presentations.
- [ ] Unit tests + extended bot scenario prove each behavior.
- [ ] `make validate-shared` and focused tests green.

## Skills

| ID | Class | Kind | Tier/Col | Prereq |
|----|-------|------|----------|--------|
| `rend` | barbarian | cone_attack + bleed | T3 c3 | `ground_slam` |
| `retribution` | paladin | self_buff (reflect on block) | T3 c1 | `hammer_of_light` |
| `predators_mark` | rogue | projectile_attack + mark | T3 c1 | `eviscerate` |

## Test proof

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestRend|TestRetribution|TestPredatorsMark' -count=1
make bot scenario=class_skill_engine_expansion
```
