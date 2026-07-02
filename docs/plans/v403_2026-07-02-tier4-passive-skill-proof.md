# v403 Plan — Tier-4 Passive Skill Proof

Spec: [`docs/specs/v403_spec-tier4-passive-skill-proof.md`](../specs/v403_spec-tier4-passive-skill-proof.md)

## Task 1 — Go tier-4 passive tests

- [x] Add `server/internal/game/tier4_passive_skills_test.go` with unlock gating + stat bonus table for all four tier-4 passives.
- [x] Reuse `assertPassiveSpendable`, `skillProgressionRow`, and `passiveSkillStatTotal` patterns from `passive_skill_column_test.go`.

**Verify:** `cd server && go test ./internal/game/... -run TestTier4Passive -count=1`

## Task 2 — Protocol bot scenario

- [x] Add extended scenario `tools/bot/scenarios/tier4_passive_proof.json` (skill lab, sorcerer, `debug_progression` level 15 + prerequisite ranks).
- [x] Steps: assert locked at L14 sim state via progression payload, allocate `arcane_reservoir`, assert rank event + `max_mana` increased (range/semantic).

**Verify:** `make bot scenario=tier4_passive_proof`

## Task 3 — Docs and finish

- [x] Write `docs/as-built/v403_tier4-passive-skill-proof.md`.
- [x] Update lifecycle + PROGRESS at `/finish`.

**Verify:** `make ci`
