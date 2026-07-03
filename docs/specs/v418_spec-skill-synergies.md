# v418 Spec — Skill Synergies

Status: Complete
Date: 2026-07-03
Codename: skill-synergies
Baseline: v417 `skill-rank-scaling-and-pacing` complete

Related:

- [`../adr/0014-core-progression-and-endgame-design-rules.md`](../adr/0014-core-progression-and-endgame-design-rules.md)
- Prior brief: prerequisite skills grant rank-scaled bonuses to dependent skills (41 catalog rows in `skills.v0.json`)

## Purpose

Make invested ranks in prerequisite skills improve dependent skills even when the prereq is not on
the hotbar. Data is already in `skills.v0.json` (`synergies[]`); this slice wires **authoritative
combat application** and **tooltip surfacing**.

Examples:

- `piercing_shot` rank → `snipe` damage %
- `volley` rank → `rain_of_arrows` volley spread %
- `ground_slam` rank → `rend` cone size %
- `teleport` rank → `energy_ward` buff duration %

Synergy source rank uses **allocated skill-tree points only** (`progression.SkillRanks`), not gear
`+N skill` effective rank (v416).

## Non-goals

- New synergy rows beyond the 41 prerequisite edges (tuning pass is a follow-up)
- Reverse synergies (dependent buffing prereq)
- Respec / refund UX
- Production VFX/audio
- Protocol version bump beyond optional `synergy_status` on existing v8 skill progression rows

## Acceptance criteria

### Shared (already landed)

- [x] Every skill with `requirements.skills` declares `synergies[]` (CI: `validate_skill_synergies`)
- [x] Schema validates modifier kinds and positive `percent_per_source_rank`

### Server

- [x] Go sim applies each modifier kind at cast/stat resolution using allocated source rank
- [x] `damage_percent`, `cone_size_percent`, `volley_spread_percent`, `projectile_range_percent`,
  `buff_duration_percent`, `buff_power_percent`, `area_radius_percent`, `root_duration_percent`,
  `mark_duration_percent`, `bleed_duration_percent`, `revive_power_percent`, `passive_stat_percent`,
  `execute_threshold_percent` covered
- [x] `SkillProgressionView` rows include optional `synergy_status` with accumulated bonus display
- [x] Focused Go tests prove snipe damage, rend cone, energy ward duration scaling

### Client

- [x] Skill tooltips list synergy sources and accumulated bonus (from `synergy_status` or rules fallback)
- [x] Headless unit test for synergy line formatting

### Bot

- [x] Extended `skill_synergy_lab` scenario: ranger snipe damage increases after maxing `piercing_shot`

## Client asset decision

**Borrow** existing skill tree tooltip layout; no new art/plugins.

## Test proof

```bash
make validate-shared
cd server && go test ./internal/game -run 'SkillSynergy' -count=1
make client-unit
make bot scenario=skill_synergy_lab
```

## Open questions

Resolved: full 41-row catalog in data; allocated rank only; beneficiary-only tooltip section.
