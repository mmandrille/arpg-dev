# v210 Plan - Barbarian Leap

Status: Complete
Goal: Add Barbarian Leap as a mobility escape that damages and stuns landing targets.
Architecture: Reuse v209 mobility execution; add only data, focused tests, class bot coverage, and docs.
Tech stack: Shared JSON/schema, Go simulation tests, Python protocol bot, lifecycle docs.

## Baseline and shortcut decision

Builds on v209 mobility support. Asset/plugin decision: borrow existing code-native skill visual/demo paths; reject external assets/plugins and production VFX.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/skills.v0.json` | Add Leap tuning. |
| Modify | `shared/assets/skill_presentations.v0.json` | Add Leap icon/effect metadata. |
| Modify | `shared/i18n/en.json`, `shared/i18n/es.json` | Add Leap text. |
| Modify | `server/internal/game/mobility_skills_test.go` | Prove Leap movement, damage, and stun. |
| Modify | `tools/bot/scenarios/51_barbarian_class_foundation.json` | Include Leap coverage. |
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

- [x] Add Leap skill, i18n, and skill presentation metadata.
```bash
make validate-shared
```

## Task 2 - Server Proof

- [x] Add focused Leap movement/damage/stun test.
```bash
cd server && go test ./internal/game -run 'TestBarbarianLeap|TestLoadRules'
```

## Task 3 - Bot Proof

- [x] Add Leap to Barbarian class foundation.
```bash
.venv/bin/pytest tools/bot/test_protocol.py::test_class_foundation_scenarios_cover_every_class_skill -q
make bot scenario=barbarian_class_foundation
```

## Task 4 - Lifecycle Docs and CI

- [x] Update docs and progress.
```bash
make ci
```

## Final Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestBarbarianLeap|TestLoadRules'`
- [x] `.venv/bin/pytest tools/bot/test_protocol.py::test_class_foundation_scenarios_cover_every_class_skill -q`
- [x] `make bot scenario=barbarian_class_foundation`
- [x] `make maintainability`
- [x] `make ci`

## Deferred Scope

Paladin Charge, Ranger Disengage, richer airborne physics, and production VFX/audio remain deferred.
