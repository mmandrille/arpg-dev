# Spec: `class-survival-skills`

Status: Complete  
Date: 2026-07-03  
Codename: class-survival-skills  
Slice: v414 — auto-proc Survival skills (one per class)

## Purpose

Add a **Survival** skill branch (column 6) with one **`survival_autocast`** skill per class. Each skill:

- Requires **character level 10** and **rank 1 paid with a skill point** (no auto-grant).
- Costs **0 mana**; only a **120s cooldown** (1200 ticks @ 10 Hz).
- **Auto-procs on PvE lethal damage** when off cooldown: incoming hit floors HP to **1**, then the survival effect activates.
- **Never procs on PvP** (player-sourced lethal damage).

## Non-goals

- Manual cast / hotbar assignment.
- Multi-tier Survival tree.
- Protocol/schema version bump.
- Production VFX/audio (borrow existing families).
- PvP survival proc (explicitly forbidden).

## Skills

| ID | Class | Duration | Effect summary |
|----|-------|----------|----------------|
| `second_wind` | barbarian | 10s | VIT +100%, 2× HP regen |
| `arcane_barrier` | sorcerer | 10s | Damage drains mana first (rank 1: 2 mana/HP; cheaper per rank) |
| `divine_protection` | paladin | 4s | Damage immunity + 5× all outgoing player damage |
| `evasive_stance` | rogue | 5s | Cleanse debuffs, 100% evade, mark all monsters in radius 5 |
| `spectral_path` | ranger | 5s | Redirect damage to attacker; walk through monsters |

Placement: `tree.branch: survival`, column 6, tier 2. Max rank 5.

## Acceptance criteria

- [ ] Schema: `tree.branch`, `kind: survival_autocast`, closed survival effect types.
- [ ] Five skills in rules + i18n + presentations.
- [ ] Lethal PvE hit → HP 1 → proc → cooldown; no proc on cooldown or rank 0.
- [ ] Player-sourced lethal → no proc (PvP guard).
- [ ] Each class effect behaves per table (Go tests).
- [ ] Skills panel renders column 6 (~580px width); Survival branch visible.
- [ ] Extended bot scenario proves barbarian proc; other classes in Go tests.
- [ ] `make validate-shared` and focused tests green.

## Test proof

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestSurvival' -count=1
make bot scenario=class_survival_skills
make client-unit
```

## Client asset decision

**Borrow** existing buff/heal/sanctuary/teleport presentation families; Survival icons use green-teal accent in `skill_presentations.v0.json`. No external plugins.

## Open questions

None — batch clarification resolved in `/next` thread.
