# v212 Plan - Ranger Disengage

Status: Complete
Goal: Add Ranger Disengage as a mobility escape that snares pursuers at the starting position.
Architecture: Extend mobility impact origin for `mode: disengage`; add data, focused tests, class bot coverage, and docs.
Tech stack: Shared JSON/schema, Go simulation tests, Python protocol bot, lifecycle docs.

## Baseline and shortcut decision

Builds on v209 mobility support plus v210/v211 class mobility examples. Asset/plugin decision: borrow existing code-native skill visual/demo paths; reject external assets/plugins and production VFX.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/game/mobility_skills.go` | Apply Disengage impact at start position. |
| Modify | `shared/rules/skills.v0.json` | Add Disengage tuning. |
| Modify | `shared/assets/skill_presentations.v0.json` | Add Disengage icon/effect metadata. |
| Modify | `shared/i18n/en.json`, `shared/i18n/es.json` | Add Disengage text. |
| Modify | `server/internal/game/mobility_skills_test.go` | Prove Disengage movement and start snare. |
| Modify | `tools/bot/scenarios/58_ranger_class_foundation.json` | Include Disengage coverage. |
| Add/Modify | lifecycle docs | Spec, plan, as-built, progress, lifecycle row. |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Keep code change localized in existing mobility helper.

Verification:
```bash
make maintainability
```

## Task 1 - Mobility Behavior

- [x] Route Disengage impact to the cast start position while keeping other mobility skills endpoint-based.
```bash
cd server && go test ./internal/game -run 'TestRangerDisengage|TestLoadRules'
```

## Task 2 - Rules and Presentation

- [x] Add Disengage skill, i18n, and skill presentation metadata.
```bash
make validate-shared
```

## Task 3 - Bot Proof

- [x] Add Disengage to Ranger class foundation.
```bash
.venv/bin/pytest tools/bot/test_protocol.py::test_class_foundation_scenarios_cover_every_class_skill -q
make bot scenario=ranger_class_foundation
```

## Task 4 - Lifecycle Docs and CI

- [x] Update docs and progress.
```bash
make ci
```

## Final Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestRangerDisengage|TestLoadRules'`
- [x] `.venv/bin/pytest tools/bot/test_protocol.py::test_class_foundation_scenarios_cover_every_class_skill -q`
- [x] `make bot scenario=ranger_class_foundation`
- [x] `make maintainability`
- [x] `make ci`

## Deferred Scope

Trap inventory, deployable persistence, stealth, invulnerability, and production VFX/audio remain deferred.
