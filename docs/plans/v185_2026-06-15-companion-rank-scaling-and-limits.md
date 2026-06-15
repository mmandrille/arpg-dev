# v185 Plan: Companion Rank Scaling and Limits

Status: Complete
Date: 2026-06-15
Spec: `docs/specs/v185_spec-companion-rank-scaling-and-limits.md`

## Adoption Checklist

- Decision: reject new plugin/asset dependency.
- Reason: this slice is server/shared data behavior only; existing companion AI and monster visuals cover presentation.
- Borrow/adopt: reuse v182/v184 companion spawn paths and extend the existing shared skill schema.

## Tasks

- [x] Replace integer companion limits with a data-driven limit rule.
- [x] Configure Revive as `base=1`, `per_rank_step=1`, `ranks_per_step=3`.
- [x] Configure Ranger wolf as data-declared one-active companion.
- [x] Apply active limits deterministically by removing oldest same-owner/same-skill companions when spawning over limit.
- [x] Keep revived monster HP/damage scaling derived from shared Revive rules.
- [x] Add validation for limit base, per-rank step, and ranks-per-step tuning.
- [x] Add Go coverage for rank-4 multi-revive and scaled stats.
- [x] Add bot proof for rank-4 two-companion Revive and scaled max HP.
- [x] Update docs/as-built and `PROGRESS.md`.
- [x] Run `make ci`.

## Bot Proof

Scenario: `tools/bot/scenarios/76_companion_rank_scaling_and_limits.json`

Expected flow:

1. Start a Sorcerer with Revive rank 4.
2. Kill and revive one wolf; assert one revived companion with rank-scaled max HP.
3. Kill and revive a second wolf after cooldown.
4. Assert two active revived wolf companions with rank-scaled max HP.
5. Prove companion-sourced damage against a nearby lab target.

Visual verification command:

```bash
make bot-visual scenario=76_companion_rank_scaling_and_limits.json
```
