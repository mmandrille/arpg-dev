# v102 Plan — Class Bot-Visual Scenarios

Status: Ready for implementation
Goal: Add class-foundation protocol and bot-visual scenarios for Paladin, Barbarian, and Sorcerer, and validate that every class skill is covered by its class scenario.
Architecture: This is a tooling/scenario slice. The Go sim remains authoritative and unchanged unless an existing bot step cannot express an already-supported action. Class and skill coverage is derived from shared rule catalogs, while the scenarios prove behavior over the normal protocol bot and Godot replay path. Existing Rogue coverage is kept and normalized under the same validation rule.
Tech stack: shared JSON rules, Python protocol bot, declarative bot scenarios, Godot bot-visual replay, docs.

## Baseline and shortcut decision

Builds on v97 `class-starter-loadouts`, v98 `rogue-class-foundation`, v99 `rogue-skill-mechanics`,
v100 `damage-types-and-resistances`, and v101 `undead-skeleton-poison-immunity`.
Current branch: `main`. Do not create branches.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `tools/bot/scenarios/50_paladin_class_foundation.json` | Paladin starter gear, movement, basic attacks, `holy_shield`, `heal` proof |
| Create | `tools/bot/scenarios/51_barbarian_class_foundation.json` | Barbarian starter gear, movement, basic attacks, `rage`, `cleave` proof |
| Create | `tools/bot/scenarios/52_sorcerer_class_foundation.json` | Sorcerer starter gear, movement, basic attacks, `magic_bolt`, `ice_shard`, `ligthing` proof |
| Modify | `tools/bot/scenarios/47_rogue_class_foundation.json` | Keep Rogue compliant with class-foundation validation if needed |
| Create | `tools/bot/class_foundation_coverage.py` | Rule-derived validation helpers for class scenario and class skill coverage |
| Modify | `tools/bot/test_protocol.py` | Focused tests for scenario discovery and class-foundation coverage |
| Modify if needed | `tools/bot/run.py` | Only for tiny reusable scenario-step support if existing actions are insufficient |
| Modify if needed | `shared/rules/worlds.v0.json` | Add one deterministic class-demo lab only if existing `skill_progression_lab` is awkward |
| Modify if needed | `shared/rules/monsters.v0.json` | Add or tune lab-only target definitions only if existing targets are unsuitable |
| Create | `docs/as-built/v102_class-bot-visual-scenarios.md` | As-built summary during finish |
| Modify | `PROGRESS.md` | Lifecycle, scenario catalog, latest completed slice during finish |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `tools/bot/run.py`
- [x] `tools/bot/test_protocol.py`
- [ ] Other over-limit file from `.maintainability/file-size-baseline.tsv`: `client/scripts/skills_panel.gd` fails standalone `make maintainability` at 838 lines versus baseline 812 + allowance 25; this slice did not touch it.

Decision:
- [x] Extract focused helper/module/test file as part of this slice.
- [ ] Defer extraction with rationale: not selected.

Rationale: class/skill scenario coverage should live in a small Python helper rather than adding
more validation logic to the already-large bot runner. `tools/bot/test_protocol.py` may still gain
small tests that call that helper.

Verification:

```bash
make maintainability
```

## Task 1 — Catalog coverage helper

Files:
- Create: `tools/bot/class_foundation_coverage.py`
- Modify: `tools/bot/test_protocol.py`

- [x] Step 1.1: Add a small helper that loads class ids from `shared/rules/character_progression.v0.json`, skill ids/classes from `shared/rules/skills.v0.json`, and scenario ids/steps/assertions from loaded bot scenarios.

```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

- [x] Step 1.2: Add validation that every class id has one scenario id named `{class_id}_class_foundation`.

```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

- [x] Step 1.3: Add validation that every skill whose `class` is a playable class is referenced by that class-foundation scenario, either in `debug_progression.skill_ranks`, an `allocate_skill_point` step, a `cast_skill` step, or an assertion that names the skill.

```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

- [x] Step 1.4: Confirm the current repo fails the new guard before adding the missing class scenarios, then keep the failure evidence in the working notes rather than committing a failing state.

```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

## Task 2 — Scenario lab decision

Files:
- Read: `shared/rules/worlds.v0.json`
- Modify if needed: `shared/rules/worlds.v0.json`
- Modify if needed: `shared/rules/monsters.v0.json`

- [x] Step 2.1: Try to script all three classes against existing `skill_progression_lab` targets first, using existing `combat_lab_soft_target`, `skill_xp_level6_dummy`, and `skill_xp_level7_dummy` where practical.

```bash
make validate-shared
```

- [x] Step 2.2: If target layout, camera framing, or multi-target skill proof is awkward, add one deterministic `class_foundation_lab` world with enough soft targets for Paladin, Barbarian, Sorcerer, and Rogue-style target separation.

```bash
make validate-shared
```

- [x] Step 2.3: Keep any lab target additions clearly test-only/content-lab scoped; do not rebalance existing monsters or combat formulas.

```bash
make validate-shared
```

## Task 3 — Paladin class-foundation scenario

Files:
- Create: `tools/bot/scenarios/50_paladin_class_foundation.json`

- [x] Step 3.1: Create `paladin_class_foundation` with `character_class: "paladin"`, visual metadata, deterministic seed, and debug progression sufficient for `heal` and `holy_shield`.

```bash
make bot scenario=paladin_class_foundation
```

- [x] Step 3.2: Assert starter `starter_paladin_sword` in `main_hand` and `starter_paladin_shield` in `off_hand`.

```bash
make bot scenario=paladin_class_foundation
```

- [x] Step 3.3: Add movement/repositioning, `holy_shield` cast, at least three main-hand basic attack damage events, and `heal` cast after damage is available or in a deterministic setup that emits visible heal state.

```bash
make bot scenario=paladin_class_foundation
```

- [x] Step 3.4: Do not wait for `holy_shield` expiration because the current effect duration is longer than 10 seconds.

```bash
make bot-visual scenario=paladin_class_foundation
```

## Task 4 — Barbarian class-foundation scenario

Files:
- Create: `tools/bot/scenarios/51_barbarian_class_foundation.json`

- [x] Step 4.1: Create `barbarian_class_foundation` with `character_class: "barbarian"`, visual metadata, deterministic seed, and debug progression sufficient for `rage` and `cleave`.

```bash
make bot scenario=barbarian_class_foundation
```

- [x] Step 4.2: Assert starter `starter_barbarian_axe` in `main_hand` and no incompatible offhand occupancy.

```bash
make bot scenario=barbarian_class_foundation
```

- [x] Step 4.3: Add movement/repositioning, `rage` cast, at least three axe basic attack damage events, and a `cleave` cast against at least one target, preferably multiple targets if deterministic in the chosen lab.

```bash
make bot scenario=barbarian_class_foundation
```

- [x] Step 4.4: Do not wait for `rage` expiration because the current effect duration is longer than 10 seconds.

```bash
make bot-visual scenario=barbarian_class_foundation
```

## Task 5 — Sorcerer class-foundation scenario

Files:
- Create: `tools/bot/scenarios/52_sorcerer_class_foundation.json`

- [x] Step 5.1: Create `sorcerer_class_foundation` with `character_class: "sorcerer"`, visual metadata, deterministic seed, and debug progression sufficient for `magic_bolt`, `ice_shard`, and `ligthing`.

```bash
make bot scenario=sorcerer_class_foundation
```

- [x] Step 5.2: Assert starter `starter_sorcerer_staff` in `main_hand` and no occupied offhand item.

```bash
make bot scenario=sorcerer_class_foundation
```

- [x] Step 5.3: Add movement/repositioning or kiting, at least three staff basic attack damage events, then cast `magic_bolt`, `ice_shard`, and `ligthing`.

```bash
make bot scenario=sorcerer_class_foundation
```

- [x] Step 5.4: Wait for cold slow ending only if the implementation verifies the current duration is under 10 seconds and the end state matters visually.

```bash
make bot-visual scenario=sorcerer_class_foundation
```

## Task 6 — Rogue normalization and validation proof

Files:
- Modify if needed: `tools/bot/scenarios/47_rogue_class_foundation.json`
- Modify: `tools/bot/test_protocol.py`

- [x] Step 6.1: Ensure `rogue_class_foundation` still references all Rogue skills (`poison_stab`, `dash`) in a way accepted by the new coverage helper.

```bash
make bot scenario=rogue_class_foundation
```

- [x] Step 6.2: Ensure Rogue still proves starter dual swords and at least three basic attack events, including offhand if current scenario support allows it.

```bash
make bot scenario=rogue_class_foundation
```

- [x] Step 6.3: Run focused Python tests proving all four classes and all current class skills are covered.

```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

## Task 7 — Bot scenario verification

Files:
- Create/Modify: `tools/bot/scenarios/*.json`
- Modify if needed: `tools/bot/run.py`

- [x] Step 7.1: Verify each new scenario individually over protocol.

```bash
make bot scenario=paladin_class_foundation
make bot scenario=barbarian_class_foundation
make bot scenario=sorcerer_class_foundation
```

- [x] Step 7.2: Verify Rogue remains green over protocol.

```bash
make bot scenario=rogue_class_foundation
```

- [x] Step 7.3: Verify all class-foundation scenarios in one selected protocol run.

```bash
make bot scenario=paladin_class_foundation,barbarian_class_foundation,sorcerer_class_foundation,rogue_class_foundation
```

- [x] Step 7.4: Run visual approval captures/replays for the three new class scenarios.

```bash
make bot-visual scenario=paladin_class_foundation
make bot-visual scenario=barbarian_class_foundation
make bot-visual scenario=sorcerer_class_foundation
```

## Task 8 — Lifecycle docs and close-out

Files:
- Create: `docs/as-built/v102_class-bot-visual-scenarios.md`
- Modify: `PROGRESS.md`
- Modify: `docs/plans/v102_2026-06-12-class-bot-visual-scenarios.md`

- [x] Step 8.1: Write the as-built summary with exact scenario ids and commands that proved the slice.

```bash
rg -n "v102|class-bot-visual|paladin_class_foundation|barbarian_class_foundation|sorcerer_class_foundation" docs/as-built PROGRESS.md
```

- [x] Step 8.2: Update `PROGRESS.md`: latest completed slice, next slice, lifecycle row, slice numbering note, scenario catalog, and any new/closed gaps.

```bash
rg -n "Latest completed slice|Next slice|v102|class-bot-visual|paladin_class_foundation|barbarian_class_foundation|sorcerer_class_foundation" PROGRESS.md
```

- [x] Step 8.3: Mark this plan's checklist complete as implementation finishes.

```bash
rg -n "\\[ \\]" docs/plans/v102_2026-06-12-class-bot-visual-scenarios.md
```

## Final verification

- [ ] `make maintainability` (blocked by unrelated `client/scripts/skills_panel.gd` ratchet failure)
- [x] `make validate-shared` if `shared/rules/` changes
- [x] `.venv/bin/pytest tools/bot/test_protocol.py -q`
- [x] `make bot scenario=paladin_class_foundation,barbarian_class_foundation,sorcerer_class_foundation,rogue_class_foundation`
- [x] `make bot-visual scenario=paladin_class_foundation`
- [x] `make bot-visual scenario=barbarian_class_foundation`
- [x] `make bot-visual scenario=sorcerer_class_foundation`
- [x] `make ci`

## Deferred scope

- Skill id cleanup for `ligthing` remains out of scope.
- New class skills, new classes, and combat balance remain out of scope.
- Rich cinematic camera choreography remains out of scope; class scenario visual metadata is enough.
