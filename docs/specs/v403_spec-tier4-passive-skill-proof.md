# v403 Spec — Tier-4 Passive Skill Proof

Status: Complete  
Date: 2026-07-02  
Codename: tier4-passive-skill-proof  
Baseline: v401 `class-skill-decuple` complete

## Purpose

Authoritatively prove the four v401 **tier-4 passives** unlock at level 15 with prerequisite chains and apply rule-defined passive stat bonuses. Close the E2E gap left after the decuple catalog landed (panel visibility only).

| Class | Skill | Prerequisite |
|-------|-------|--------------|
| Sorcerer | `arcane_reservoir` | `spell_dynamo` rank 1 |
| Barbarian | `unstoppable_heart` | `crushing_force` rank 1 |
| Paladin | `oathbound_resolve` | `consecrated_vitality` rank 1 |
| Ranger | `wildborn_endurance` | `deadeye` rank 1 |

## Non-goals

- Skill-tree UI layout rewrite.
- Rank-5 skill visual matrix / buff stat-delta reporting.
- New skills or balance retuning.
- Full four-class bot walkthrough (Go table owns all four; bot proves one representative path).

## Acceptance criteria

- [ ] Go table test: tier-4 passive **blocked below L15** and **spendable at L15** with prerequisite met (all four classes).
- [ ] Go table test: each tier-4 passive applies **rule-derived** passive stat totals (semantic minimum from `passive_stats.stats`).
- [ ] Extended bot scenario: sorcerer learns `arcane_reservoir` at L15 and shows increased `max_mana` breakdown vs pre-rank baseline.
- [ ] `make validate-shared`, focused Go tests, and bot scenario green.

## Scope and files

- `server/internal/game/tier4_passive_skills_test.go` — unlock + stat bonus tests
- `tools/bot/scenarios/tier4_passive_proof.json` — extended protocol proof (sorcerer)
- `docs/plans/v403_2026-07-02-tier4-passive-skill-proof.md`
- `docs/as-built/v403_tier4-passive-skill-proof.md`, lifecycle row at finish

No protocol/schema bump.

## Test and bot proof

```bash
cd server && go test ./internal/game/... -run TestTier4Passive -count=1
make bot scenario=tier4_passive_proof
make validate-shared
```

## Open questions

- Q-1: Bot representative class? **Default: sorcerer** (`max_mana_percent` is easy to observe).
- Q-2: Persistence/resume in bot? **Defer** — Go tests cover rank retention via fresh sim progression state.

## Asset decision

Reject external assets; no client presentation change beyond existing skills panel.
