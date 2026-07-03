# v417 As-Built — Skill Rank Scaling and Pacing

Date: 2026-07-03

## What shipped

- Skill-point cadence: grants at levels 1, 2, then every even level ≥4 (51 points by level 100).
- `max_rank` raised to 10 for rankable actives; utilities stay at `max_rank: 1`.
- Global compound rank curves in `character_progression.v0.json`: 8% magnitudes, 10% mana.
- Server `skill_rank_scaling.go` centralizes curve evaluation; sim damage/mana/effects/passives wired.
- Softened buy requirements: `level_per_rank` steps cap at 5 for ranks 6–10.
- Client `skill_rank_scaling.gd` mirrors curves for skill panel, bar, and passive tooltips.
- Golden `skill_points_and_magic_bolt.json` + validator cadence/compound cross-checks updated.
- Extended bot proofs: `class_progression_pacing`, new `skill_rank_scaling_lab`.

## Proof

```bash
make validate-shared
make maintainability
cd server && go test ./internal/game -run 'SkillPoint|SkillRank|TestFiveAllocates|TestAllocateCast' -count=1
make bot scenario=class_progression_pacing scenario=skill_rank_scaling_lab
```

## Deferred

- Respec UI / bishop respec scenario updates for max_rank 10 (existing scenarios still pin 5 where contract-specific)
- Skill-tree breadth expansion (~30 skills/class) and specialization pressure tuning
- Engineering review at ~v418 milestone
