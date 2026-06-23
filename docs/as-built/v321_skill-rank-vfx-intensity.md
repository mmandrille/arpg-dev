# v321 As Built - Skill Rank VFX Intensity

Date: 2026-06-23
Spec: [`docs/specs/v321_spec-skill-rank-vfx-intensity.md`](docs/specs/v321_spec-skill-rank-vfx-intensity.md)

## What Shipped

- Added optional `rank_intensity` metadata to skill presentations schema/data.
- `SkillRankIntensity` resolves accent width, glow rings, and cast burst scale by rank.
- Skill icons and cast bursts scale with rank; sample entries on `magic_bolt` and `cleave`.

## Proof

```bash
make validate-shared
godot --headless --path client --script res://tests/test_skill_rank_intensity.gd
make client-unit
make maintainability
```
