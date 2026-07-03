# v416 Spec — Per-Skill Affix Rolls

Status: Complete
Date: 2026-07-03
Codename: per-skill-affix-rolls
Baseline: v415 `class-specialist-expansion` complete

Related:

- [`v189_spec-skill-affix-rolls.md`](v189_spec-skill-affix-rolls.md)
- [`v405_spec-class-specialist-gear.md`](v405_spec-class-specialist-gear.md)
- [`../adr/0014-core-progression-and-endgame-design-rules.md`](../adr/0014-core-progression-and-endgame-design-rules.md)

## Purpose

Add **+N to a random skill** as a weighted affix roll on generic **helm** and **amulet**
(~5%, rare+, all skills in catalog). Add **+N to a class skill** on **`class_required`**
templates (~10%, that class only — no cross-class picks).

Persist bonuses on `ItemRollPayload.skill_level_bonuses`. Rework **effective skill rank** so gear
bonuses (`all_skills` and per-skill) stack without `MaxRank` clamp or stat-requirement clamp on
**use**. Requirements remain **spend-only** when allocating skill points.

Server-authored **`skill_bonus_status`** drives client tooltips: green when active, red when wrong
class, gray when skill not yet learned.

## Non-goals

- Skill formula scaling audit or passive `PerRank: 0` fixes (v417 follow-up)
- Crafting, affix name grammar, production art
- +skill on generic weapons/rings
- Changing skill-tree `MaxRank` for point spending
- Protocol version bump beyond optional `skill_bonus_status` on existing v8 item views

## Acceptance criteria

### Shared rules

- [ ] `helm` and `amulet` include ~5% weighted `random_skill_level` roll (rare+)
- [ ] `class_required` templates include ~10% weighted `random_class_skill_level` roll
- [ ] Class-locked items pick skills only where `skill.class == template.class_required`
- [ ] Bonus value rolls `1..item_level` inclusive
- [ ] Schema validates roll kinds and `skill_level_bonuses` payload shape
- [ ] `make validate-shared` passes

### Server

- [ ] `effectiveSkillRank` = `baseRank + all_skills + perSkillBonus(skill)` when `baseRank >= 1`
- [ ] No `MaxRank` clamp and no `skillRequirementsMet` clamp on effective rank
- [ ] `spend_skill_intent` still enforces requirements and `MaxRank`
- [ ] `cast_skill_intent` and other use paths do not reject casts for unmet stat requirements
- [ ] `skill_bonus_status` on item views with `active` false for wrong class or unlearned skill
- [ ] Deterministic skill pick via sorted skill IDs + seeded RNG
- [ ] Focused Go tests cover roll, rank, spend gate, cast gate, class filter, status

### Client

- [ ] Tooltips show per-skill bonus lines (green active, red wrong class, gray unlearned)
- [ ] Headless unit test for skill bonus line formatting

### Bot

- [ ] Extended lab scenario: equip item with +skill, assert effective rank / cast behavior

## Scope and files

- `shared/rules/item_templates.v0.json`, `item_templates.v0.schema.json`
- `shared/golden/item_rolls.v0.schema.json`
- `shared/protocol/state_delta.v8.schema.json`, `session_snapshot.v8.schema.json`
- `server/internal/game/item_skill_level_bonuses.go`, `item_skill_stats.go`, `item_rolls.go`,
  `item_reroll.go`, `handlers.go`, `sim.go`, `types.go`, tests
- `client/scripts/skill_bonus_tooltip.gd`, inventory/stash/shop/market panels, tests
- `tools/bot/scenarios/` — new extended lab scenario
- `tools/validate_shared.py`

## Test and bot proof

```bash
make validate-shared
make maintainability
cd server && go test ./internal/game/... -run 'SkillLevelBonus|SkillAffix|effectiveSkillRank' -count=1
godot --headless --path client --script res://tests/test_skill_bonus_tooltip.gd
make bot scenario=skill_level_affix_lab
```

## Asset decision

- **Adopt:** `ClassAffinityTooltip` color pattern (green/red)
- **Borrow:** existing stat tooltip sections
- **Reject:** external art

## Open questions

None — locked in `/next` brief and user clarification (2026-07-03).
