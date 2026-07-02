# v403 As-built — Tier-4 Passive Skill Proof

## What it proved

- All four tier-4 passives (`arcane_reservoir`, `unstoppable_heart`, `oathbound_resolve`, `wildborn_endurance`) require level 15 + tier-3 prerequisite before spend.
- Rule-derived passive stat totals apply when tier-4 rank is learned (Go table across all classes).
- Extended bot scenario `tier4_passive_proof` proves sorcerer `arcane_reservoir` rank-up and passive `max_mana` breakdown contribution.
- Skill projectile `monster_damaged` events now carry `skill_id` when emitted through `damageMonsterByPlayerSkillTypedWithID`.

## Key files

- `server/internal/game/tier4_passive_skills_test.go`
- `tools/bot/scenarios/tier4_passive_proof.json`
- `server/internal/game/sim.go` — skill_id on skill damage events

## Verification

```bash
cd server && go test ./internal/game/... -run TestTier4Passive -count=1
make bot scenario=tier4_passive_proof
make validate-shared
make ci
```

## Non-goals honored

- No skill-tree UI rewrite, rank-5 visual matrix, or four-class bot matrix.
- Resume/persist bot proof deferred (Go progression state covers rank retention shape).
