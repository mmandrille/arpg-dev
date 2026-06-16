# v211 Plan - Paladin Charge

Status: Complete
Goal: Add Paladin Charge as a mobility escape that shield-smashes and stuns endpoint targets.
Architecture: Reuse v209 mobility execution; add only data, focused tests, class bot coverage, and docs.
Tech stack: Shared JSON/schema, Go simulation tests, Python protocol bot, lifecycle docs.

## Baseline and shortcut decision

Builds on v209 mobility support and v210 Leap proof. Asset/plugin decision: borrow existing code-native skill visual/demo paths; reject external assets/plugins and production VFX.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/skills.v0.json` | Add Charge tuning. |
| Modify | `shared/assets/skill_presentations.v0.json` | Add Charge icon/effect metadata. |
| Modify | `shared/i18n/en.json`, `shared/i18n/es.json` | Add Charge text. |
| Modify | `server/internal/game/mobility_skills_test.go` | Prove Charge movement, damage, and stun. |
| Modify | `tools/bot/scenarios/50_paladin_class_foundation.json` | Include Charge coverage. |
| Add/Modify | lifecycle docs | Spec, plan, as-built, progress, lifecycle row. |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Reuse focused mobility files; no coordinator growth.

Verification:
```bash
make maintainability
```

## Task 1 - Rules and Presentation

- [x] Add Charge skill, i18n, and skill presentation metadata.
```bash
make validate-shared
```

## Task 2 - Server Proof

- [x] Add focused Charge movement/damage/stun test.
```bash
cd server && go test ./internal/game -run 'TestPaladinCharge|TestLoadRules'
```

## Task 3 - Bot Proof

- [x] Add Charge to Paladin class foundation.
```bash
.venv/bin/pytest tools/bot/test_protocol.py::test_class_foundation_scenarios_cover_every_class_skill -q
make bot scenario=paladin_class_foundation
```

## Task 4 - Lifecycle Docs and CI

- [x] Update docs and progress.
```bash
make ci
```

## Final Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestPaladinCharge|TestLoadRules'`
- [x] `.venv/bin/pytest tools/bot/test_protocol.py::test_class_foundation_scenarios_cover_every_class_skill -q`
- [x] `make bot scenario=paladin_class_foundation`
- [x] `make maintainability`
- [x] `make ci`

## Deferred Scope

Ranger Disengage, shield-equipment requirements, wall-breaking collision, and production VFX/audio remain deferred.
