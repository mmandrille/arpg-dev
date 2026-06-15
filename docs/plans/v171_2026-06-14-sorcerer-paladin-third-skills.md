# v171 Plan - Sorcerer Paladin Third Skills

Status: Complete
Goal: Add one higher-row active skill for Sorcerer and Paladin using existing skill systems.
Architecture: Keep skill behavior data-driven through the shared skill catalog. Reuse existing
server skill kind handlers and Godot presentation loaders; no protocol schema or new mechanics
framework is required.
Tech stack: shared JSON, Go sim validation/tests, Godot client tests, Python bot scenario/tooling.

## Baseline and shortcut decision

presentation are sufficient for two catalog-driven skills.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/skills.v0.json` | Add `arcane_barrage` and `sanctuary` |
| Modify | `shared/assets/skill_presentations.v0.json` | Add icon/summary/visual metadata |
| Modify | `shared/i18n/en.json`, `shared/i18n/es.json` | Add names/summaries |
| Modify | `tools/validate_skills.py` | Validate expected class skill expansion |
| Modify | `server/internal/game/game_test.go` | Add focused catalog assertions |
| Modify | `server/internal/game/class_third_skills_test.go` | Add prerequisite spendability cases |
| Modify | `tools/bot/test_skill_demo.py` | Include new visual catalog entries |
| Modify | `client/tests/test_skill_rules_loader.gd` | Include new presentation/loader coverage |
| Add | `tools/bot/scenarios/69_sorcerer_arcane_barrage.json` | Protocol proof for Sorcerer Arcane Barrage |
| Add | `tools/bot/scenarios/70_paladin_sanctuary.json` | Protocol proof for Paladin Sanctuary |
| Add | `docs/as-built/v171_sorcerer-paladin-third-skills.md` | Shipped proof |
| Modify | `PROGRESS.md` | Lifecycle and deferred scope |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [ ] `client/scripts/main.gd`
- [x] `server/internal/game/game_test.go`
- [ ] `tools/bot/run.py`
- [ ] `tools/validate_shared.py`
- [ ] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [ ] Extract focused helper/module/test file as part of this slice, or
- [x] Defer extraction with rationale: only focused catalog assertions are expected; extracting
  broad skill tests during a content slice would add risk without reducing touched behavior.

Verification:
```bash
make maintainability
```

## Task 1 - Shared catalog

Files:
- Modify: `shared/rules/skills.v0.json`
- Modify: `shared/assets/skill_presentations.v0.json`
- Modify: `shared/i18n/en.json`
- Modify: `shared/i18n/es.json`
- Modify: `tools/validate_skills.py`

- [x] Add `arcane_barrage` as a Sorcerer projectile skill requiring `ligthing`.
- [x] Add `sanctuary` as a Paladin area defense skill requiring `holy_shield`.
- [x] Add presentation and localized text metadata.
- [x] Extend validation coverage for the expanded class skill set.

```bash
make validate-shared
```

## Task 2 - Server/client/tool tests

Files:
- Modify: `server/internal/game/game_test.go`
- Modify: `server/internal/game/class_third_skills_test.go`
- Modify: `tools/bot/test_skill_demo.py`
- Modify: `client/tests/test_skill_rules_loader.gd`

- [x] Add focused Go rule/progression assertions.
- [x] Extend Python skill-demo coverage.
- [x] Extend Godot loader/presentation assertions.

```bash
cd server && go test ./internal/game -run 'TestLoadRules|TestThirdClassSkillsRequirePrerequisites|TestClassSkillGates' -count=1
.venv/bin/pytest tools/bot/test_skill_demo.py -q
make client-unit
```

## Task 3 - Bot scenarios

Files:
- Add: `tools/bot/scenarios/69_sorcerer_arcane_barrage.json`
- Add: `tools/bot/scenarios/70_paladin_sanctuary.json`

- [x] Prove Sorcerer `arcane_barrage` and Paladin `sanctuary` cast through focused protocol bot scenarios.

```bash
make bot scenario=69_sorcerer_arcane_barrage.json
make bot scenario=70_paladin_sanctuary.json
```

## Task 4 - Lifecycle docs and CI

Files:
- Add: `docs/as-built/v171_sorcerer-paladin-third-skills.md`
- Modify: `PROGRESS.md`
- Modify: this plan

- [x] Record shipped proof and update lifecycle docs.
- [x] Run final verification.

```bash
make maintainability
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestLoadRules|TestThirdClassSkillsRequirePrerequisites|TestClassSkillGates' -count=1`
- [x] `.venv/bin/pytest tools/bot/test_skill_demo.py -q`
- [x] `make client-unit`
- [x] `make bot scenario=69_sorcerer_arcane_barrage.json`
- [x] `make bot scenario=70_paladin_sanctuary.json`
- [x] `make maintainability`
- [x] `make ci`

## Deferred scope

- New skill capability types and more distinctive Sorcerer/Paladin mechanics.
- Passive skill trees, mana regeneration, and final skill balance.
