# v417 Spec — Skill Rank Scaling and Pacing

Status: Approved for implementation
Date: 2026-07-03
Codename: skill-rank-scaling-and-pacing
Baseline: v416 `per-skill-affix-rolls` complete

Related:

- [`v412_spec-class-build-pacing.md`](v412_spec-class-build-pacing.md)
- [`v416_spec-per-skill-affix-rolls.md`](v416_spec-per-skill-affix-rolls.md)
- [`../adr/0014-core-progression-and-endgame-design-rules.md`](../adr/0014-core-progression-and-endgame-design-rules.md)

## Purpose

Reshape skill progression and rank scaling so every invested rank feels meaningful, gear-pushed
ranks above the buyable cap keep improving, and players face real depth-vs-breadth choices.

1. **Skill-point cadence** — grant at levels **1, 2, 4, 6, 8…** (~**51** points at level cap 100).
2. **Buyable depth** — raise spendable `max_rank` from **5 → 10** for rankable actives/passives.
3. **Unified rank curves** — data-driven `compound_percent` scaling (default **8%** per rank on
   magnitudes; **10%** on mana costs) shared by Go sim and GDScript tooltips.
4. **Uncapped effective scaling** — `effectiveSkillRank` (v416) uses the same curves with no
   artificial ceiling; every rank strictly improves at least one outcome field.
5. **Spend gates** — soften `level_per_rank` requirement growth for ranks **6–10**; keep stat and
   prerequisite gates for early ranks.

Design ratio (documented, not enforced in code): **~50 spendable points** vs a future
**30 skills × 10 ranks** catalog (300 rank slots) to reward specialization.

## Non-goals

- Adding skills toward 30/class catalog
- Player respec / skill refund UI
- Full combat balance pass across all dungeon tiers
- Production skill VFX/audio
- Protocol version bump
- Changing stat-point cadence or class `level_stat_growth` from v412

## Acceptance criteria

- [ ] `character_progression.v0.json` grants skill points at 1, 2, 4, 6…; total at L100 = 51
- [ ] Spendable skills have `max_rank: 10` (rank-1 utility skills unchanged)
- [ ] `skill_rank_scaling` and `skill_mana_scaling` in progression rules; schema validates
- [ ] Go `rankScaledInt` / `rankScaledFloat` drive damage, effects, passives, mobility, mana, marks,
  bleed, poison, execute, companion/revive scaling
- [ ] `rankScaledInt(rank+1) > rankScaledInt(rank)` for representative skills at ranks 1–9 and
  gear-pushed ranks up to 15
- [ ] `allocate_skill_point` rejects above `max_rank`; cast paths use `effectiveSkillRank`
- [ ] `skillRequirementsForRank` caps `level_per_rank` steps at 5
- [ ] GDScript `SkillRankScaling` mirrors Go for tooltip previews
- [ ] `make validate-shared` passes; focused Go + GDScript tests green
- [ ] Extended `class_progression_pacing` and new `skill_rank_scaling_lab` bot scenarios green

## Client asset decision

**Borrow** existing skill icons and cast-burst presentation; no new plugins.

## Test and bot proof

```bash
make validate-shared
cd server && go test ./internal/game -run 'SkillPoint|SkillRankScaling' -count=1
godot --headless --path client --script res://tests/test_skill_rank_scaling.gd
make bot scenario=class_progression_pacing
make bot scenario=skill_rank_scaling_lab
```

## Open questions

Resolved by autoloop defaults: 51-point cadence, 8% compound magnitudes, 10% mana compound,
framework + all actives, silent rank clamp on resume, soften level gates ranks 6–10.
