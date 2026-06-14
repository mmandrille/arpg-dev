# v154 Plan - Class Third Skill Trio

Status: Ready for implementation
Goal: Add one new active higher-row skill for Barbarian, Rogue, and Ranger without changing current skills.
Architecture: Keep skill behavior data-driven through the existing shared skill catalog. Reuse
existing server skill kind handlers and client presentation loaders; no protocol or new mechanics
framework is required.
Tech stack: shared JSON, Go sim validation/tests, Godot client tests, Python bot scenario/tooling.

## Baseline and shortcut decision

Builds on v153. Godot plugin decision: reject external skill-tree or VFX plugins. Existing
server-owned skill rules and code-native Godot presentation are sufficient for three catalog-driven
skills.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/skills.v0.json` | Add `earthbreaker`, `shadow_flurry`, `split_arrow` |
| Modify | `shared/assets/skill_presentations.v0.json` | Add icon/summary/visual metadata |
| Modify | `shared/i18n/en.json`, `shared/i18n/es.json` | Add names/summaries |
| Modify | `tools/validate_skills.py` | Validate expected class skill expansion |
| Modify | `server/internal/game/game_test.go` | Add focused catalog/progression assertions |
| Add | `server/internal/game/class_third_skills_test.go` | Focused prerequisite spendability tests |
| Modify | `tools/bot/test_skill_demo.py` | Include new visual catalog entries |
| Modify | `client/tests/test_skill_rules_loader.gd` | Include new presentation/loader coverage |
| Add | `tools/bot/scenarios/62_barbarian_earthbreaker.json` | Protocol proof for Barbarian Earthbreaker |
| Add | `tools/bot/scenarios/63_rogue_shadow_flurry.json` | Protocol proof for Rogue Shadow Flurry |
| Add | `tools/bot/scenarios/64_ranger_split_arrow.json` | Protocol proof for Ranger Split Arrow |
| Add | `docs/as-built/v154_class-third-skill-trio.md` | Shipped proof |
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
- [x] Defer extraction with rationale: only focused catalog assertions are added; extracting
  `game_test.go` during a content slice would add risk without reducing the touched behavior.

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

- [x] Add three new skill definitions using existing supported kinds.
- [x] Add presentation and text metadata.
- [x] Extend validation coverage for the new class skills.

```bash
make validate-shared
```

## Task 2 - Server/client/tool tests

Files:
- Modify: `server/internal/game/game_test.go`
- Modify: `tools/bot/test_skill_demo.py`
- Modify: `client/tests/test_skill_rules_loader.gd`

- [x] Add focused Go rule/progression assertions.
- [x] Extend Python skill-demo coverage.
- [x] Extend Godot loader/presentation assertions.

```bash
cd server && go test ./internal/game -run 'TestLoadRules|TestSkill'
.venv/bin/pytest tools/bot/test_skill_demo.py -q
make client-unit
```

## Task 3 - Bot scenario

Files:
- Add: `tools/bot/scenarios/62_barbarian_earthbreaker.json`
- Add: `tools/bot/scenarios/63_rogue_shadow_flurry.json`
- Add: `tools/bot/scenarios/64_ranger_split_arrow.json`

- [x] Prove Barbarian `earthbreaker`, Rogue `shadow_flurry`, and Ranger `split_arrow` cast through
  focused protocol bot scenarios.

```bash
make bot scenario=62_barbarian_earthbreaker.json
make bot scenario=63_rogue_shadow_flurry.json
make bot scenario=64_ranger_split_arrow.json
```

## Task 4 - Lifecycle docs and CI

Files:
- Add: `docs/as-built/v154_class-third-skill-trio.md`
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
- [x] `cd server && go test ./internal/game -run 'TestLoadRules|TestSkill'`
- [x] `.venv/bin/pytest tools/bot/test_skill_demo.py -q`
- [x] `make bot scenario=62_barbarian_earthbreaker.json`
- [x] `make bot scenario=63_rogue_shadow_flurry.json`
- [x] `make bot scenario=64_ranger_split_arrow.json`
- [x] `make client-unit`
- [x] `make maintainability`
- [x] `make ci`

## Deferred scope

- Sorcerer and Paladin new higher-row skills.
- New skill capability types.
- Broader skill-tree prerequisite restructuring.
