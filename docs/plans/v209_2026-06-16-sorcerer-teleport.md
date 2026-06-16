# v209 Plan - Sorcerer Teleport

Status: Complete
Goal: Add a server-authoritative Sorcerer Teleport mobility skill and reusable mobility skill contract.
Architecture: Extend the shared skill catalog with a `mobility` kind and keep the authoritative endpoint resolution in Go. Reuse existing `cast_skill_intent`, `skill_cast`, cooldown, mana, and code-native skill visual paths.
Tech stack: Shared JSON/schema, Go simulation, Python bot visual tooling, lifecycle docs.

## Baseline and shortcut decision

Builds on the existing skill stack and Rogue Dash endpoint resolver. Asset/plugin decision: borrow existing code-native skill visual/demo paths; reject external assets/plugins and production VFX.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/skills.v0.schema.json` | Allow `mobility` skill kind and typed mobility payload. |
| Modify | `shared/rules/skills.v0.json` | Add Sorcerer `teleport` tuning. |
| Modify | `server/internal/game/rules.go` | Add mobility payload to `SkillDef` and validation dispatch. |
| Modify | `server/internal/game/rogue_rules.go` | Add mobility payload validation near existing dash validation. |
| Modify | `server/internal/game/handlers.go` | Dispatch mobility skills from `handleCastSkill`. |
| Add | `server/internal/game/mobility_skills.go` | Server-authoritative movement, cast event, and optional impact helper. |
| Add | `server/internal/game/mobility_skills_test.go` | Focused Teleport sim coverage. |
| Modify | `tools/bot/skill_demo.py` | Categorize mobility skills. |
| Modify | `tools/bot/skill_visual_runtime.py` | Cast mobility skills in visual replay without damage assertion. |
| Add/Modify | lifecycle docs | Spec, plan, as-built, progress, lifecycle row. |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/rules.go`
- [x] `server/internal/game/handlers.go`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Extract focused `mobility_skills.go` and test file.
- [x] Touch large coordinator files only for struct/validation/dispatch wiring.

Verification:
```bash
make maintainability
```

## Task 1 - Shared Skill Contract

Files:
- Modify: `shared/rules/skills.v0.schema.json`
- Modify: `shared/rules/skills.v0.json`

- [x] Step 1.1: Add `mobility` schema and Teleport rules.
- [x] Step 1.2: Validate shared data.
```bash
make validate-shared
```

## Task 2 - Server Mobility Execution

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/rogue_rules.go`
- Modify: `server/internal/game/handlers.go`
- Add: `server/internal/game/mobility_skills.go`
- Add: `server/internal/game/mobility_skills_test.go`

- [x] Step 2.1: Add typed mobility rules and validation.
- [x] Step 2.2: Implement authoritative Teleport movement using collision-safe endpoint resolution.
- [x] Step 2.3: Cover Teleport movement and no-damage behavior.
```bash
cd server && go test ./internal/game -run 'TestSorcererTeleport|TestLoadRules'
```

## Task 3 - Visual Tooling

Files:
- Modify: `tools/bot/skill_demo.py`
- Modify: `tools/bot/skill_visual_runtime.py`

- [x] Step 3.1: Categorize mobility and build a directional cast visual step.
- [x] Step 3.2: Keep damage assertions only for damaging skill kinds.
```bash
.venv/bin/pytest tools/bot/test_skill_visual.py -q
```

## Task 4 - Lifecycle Docs and CI

Files:
- Modify: `docs/specs/v209_spec-sorcerer-teleport.md`
- Modify: `docs/plans/v209_2026-06-16-sorcerer-teleport.md`
- Add: `docs/as-built/v209_sorcerer-teleport.md`
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`

- [x] Step 4.1: Mark spec/plan complete and write as-built.
- [x] Step 4.2: Update progress and lifecycle docs after verification.
```bash
make ci
```

## Final Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestSorcererTeleport|TestLoadRules'`
- [x] `.venv/bin/pytest tools/bot/test_skill_visual.py -q`
- [x] `make maintainability`
- [x] `make ci`

## Deferred Scope

Barbarian Leap, Paladin Charge, Ranger Disengage, richer VFX/audio, and Teleport combat side effects remain deferred.
