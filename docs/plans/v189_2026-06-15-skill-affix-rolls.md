# v189 Plan — Skill Affix Rolls

Status: Complete - `make ci` green on 2026-06-15
Goal: Make higher-rarity skill utility affixes rollable and live in authoritative skill casts.
Architecture: Shared item templates own roll eligibility and value bounds. The Go sim owns equipped
stat aggregation and applies skill cost/cooldown modifiers at cast commit time. Existing protocol
events expose mana spend and cooldown ticks.

## Baseline and shortcut decision

Builds on v187/v188 rarity pools and live combat-affix aggregation. Godot plugin decision: reject
new client/plugin work; existing inventory/shop/stash stat-line rendering is sufficient.

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines. Avoid growing `sim.go` by placing skill
affix helpers in `item_skill_stats.go` where skill item-stat behavior already lives.

Verification:
```bash
make maintainability
```

## Task 1 — Shared skill affix data

Files:
- Modify: `shared/rules/item_templates.v0.schema.json`
- Modify: `shared/rules/item_templates.v0.json`
- Modify: `shared/golden/item_rolls.v0.schema.json`
- Modify: `shared/rules/shops.v0.json`
- Modify: `shared/rules/shops.v0.schema.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Add supported stat ids and bounds for cooldown and mana-cost skill affixes.
- [x] Step 1.2: Add rare-gated roll candidates to skill-oriented equipment.
- [x] Step 1.3: Add shop pricing coverage.
```bash
make validate-shared
```

## Task 2 — Server skill cost/cooldown behavior

Files:
- Modify: `server/internal/game/item_skill_stats.go`
- Modify: `server/internal/game/handlers.go`
- Modify/Create: focused `server/internal/game/*_test.go`

- [x] Step 2.1: Aggregate equipped skill cooldown and mana-cost reduction totals.
- [x] Step 2.2: Apply mana-cost reduction before mana checks and Blood Price fallback.
- [x] Step 2.3: Apply cooldown reduction when skill cooldowns are committed.
- [x] Step 2.4: Add focused tests for reduced mana, reduced cooldown, and existing all-skills/skill-damage continuity.
```bash
cd server && go test ./internal/game/... -run 'SkillAffix|SkillDamage|AllSkills|MagicBolt'
```

## Task 3 — Client/protocol proof

Files:
- Modify: client stat label/rendering scripts
- Create: `tools/bot/scenarios/80_skill_affix_rolls.json`

- [x] Step 3.1: Show new skill affix stat lines in inventory/shop/stash labels.
- [x] Step 3.2: Add compact protocol bot proof for rolled skill affix protocol payloads.
```bash
make bot scenario=80_skill_affix_rolls.json
```

## Task 4 — Lifecycle docs and CI

Files:
- Create: `docs/as-built/v189_skill-affix-rolls.md`
- Modify: this plan, v189 spec, and `PROGRESS.md`

- [x] Step 4.1: Mark spec/plan complete after verification.
- [x] Step 4.2: Add as-built summary and PROGRESS lifecycle row.
- [x] Step 4.3: Record deferred per-skill affix targeting and naming scope.
```bash
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/... -run 'SkillAffix|SkillDamage|AllSkills|MagicBolt'`
- [x] `make bot scenario=80_skill_affix_rolls.json`
- [x] `make ci`
