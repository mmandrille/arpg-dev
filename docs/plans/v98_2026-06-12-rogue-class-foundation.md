# v98 Plan - Rogue class foundation

Status: Complete
Goal: Add Rogue as a selectable, persistent class with a slim model, starter dual swords, potions, and Rogue-only off-hand one-handed weapon equip support.
Architecture: Gameplay authority stays in shared rules and the Go server. The Godot client consumes class presentation metadata and generated deterministic GLB assets. Starter items remain durable inventory rows seeded by the HTTP character-creation path. v98 adds an equipment-rule exception for Rogue off-hand one-handed weapons but does not add off-hand attack timing yet.
Tech stack: shared JSON rules/assets, Go server/game/http tests, Godot client scripts/tests, deterministic GLB generation, Python protocol bot, lifecycle docs.

## Baseline and shortcut decision

Builds on v69-v71 class identity/picker/presentation and v97 starter loadouts. Godot plugin adoption:
reject external plugins and asset packs for this slice. Existing deterministic GLB generation, class
presentation metadata, and equipment presentation loaders already cover the required visual surface;
adopting a plugin would add dependency risk without solving a missing capability.

## Scope split

Included in v98:
- Rogue class rules, client picker visibility, class presentation, model, starter loadout, and
  Rogue-only off-hand one-handed weapon equip rules.

Deferred:
- Poison Stab DOT skill.
- Dash skill.
- Off-hand independent attack cadence/damage at 1.5x regular hand speed.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/character_progression.v0.json` | Add Rogue class stats. |
| Modify | `shared/rules/item_templates.v0.json` | Add starter Rogue sword template. |
| Modify | `shared/assets/class_presentations.v0.json` and schema | Add Rogue icon/model mapping. |
| Modify | `shared/assets/item_presentations.v0.json` | Add starter Rogue sword presentation. |
| Modify | `shared/assets/manifest` / generated asset metadata | Register Rogue model asset if required by current tooling. |
| Modify | `tools/assets/gen_glb.py` | Generate a slim Rogue character GLB. |
| Modify | `server/internal/http/starter_loadout.go` and tests | Seed Rogue starter swords and potions. |
| Modify | `server/internal/game/sim.go` or equipment helper tests | Allow Rogue off-hand one-handed weapons and reject others. |
| Modify | `client/scripts/character_select_panel.gd` and tests | Expose fourth Rogue option and resolve row/tooltips. |
| Modify | `client/tests/test_animation.gd` | Assert Rogue model resolves and has required bones. |
| Add/Modify | `tools/bot/scenarios/*.json`, `tools/bot/test_protocol.py` | Protocol proof for Rogue starter equipment. |
| Add | `docs/as-built/v98_rogue-class-foundation.md` | As-built summary. |
| Modify | `PROGRESS.md` | Lifecycle close-out. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files likely touched:
- [x] `server/internal/game/sim.go`
- [x] `server/internal/game/game_test.go`
- [x] `client/tests/test_coop_client.gd`
- [x] `client/tests/test_animation.gd`
- [x] `tools/bot/test_protocol.py`

Decision:
- [ ] Extract focused helper/module/test file as part of this slice, or
- [x] Defer extraction with rationale: expected edits are narrow data-path additions to established
  class/equipment/model tests. Broad extraction while adding a new class would raise regression risk.
- [x] Documented maintenance exception: update the grandfathered baseline for focused Rogue edits in
  `client/tests/test_coop_client.gd`, `server/internal/game/game_test.go`, `server/internal/game/sim.go`,
  `tools/bot/test_protocol.py`, and `tools/validate_shared.py`. The ratchet also surfaced already-present
  growth in `client/scripts/main.gd`, `client/tests/test_item_visuals.gd`, `server/internal/game/rules.go`,
  and `skills/showme/scripts/visual_capture.gd`; this slice records the current baseline so final CI can
  enforce growth from here.

Verification:
```bash
make maintainability
```

## Task 1 - Shared class and starter data

Files:
- Modify: `shared/rules/character_progression.v0.json`
- Modify: `shared/rules/item_templates.v0.json`
- Modify: `shared/assets/class_presentations.v0.json`
- Modify: `shared/assets/class_presentations.v0.schema.json`
- Modify: `shared/assets/item_presentations.v0.json`
- Modify: `shared/i18n/en.json`, `shared/i18n/es.json`

- [x] Step 1.1: Add Rogue with dexterity-leaning base stats.
- [x] Step 1.2: Add `starter_rogue_sword` as a common one-handed sword template.
- [x] Step 1.3: Add Rogue class presentation, starter sword presentation, and text keys.
- [x] Step 1.4: Update schema support for the Rogue icon shape if needed.
```bash
make validate-shared
```

## Task 2 - Server starter kit and equipment rules

Files:
- Modify: `server/internal/http/starter_loadout.go`
- Modify: `server/internal/http/starter_loadout_test.go`
- Modify: `server/internal/game/sim.go`
- Modify/Create: `server/internal/game/*_test.go`

- [x] Step 2.1: Seed Rogue starter sword in both `main_hand` and `off_hand`, plus one red and one blue potion.
- [x] Step 2.2: Allow Rogue to equip one-handed weapon templates/defs into `off_hand`.
- [x] Step 2.3: Preserve non-Rogue off-hand weapon rejection and two-handed hand blocking.
- [x] Step 2.4: Add focused tests for starter loadout and off-hand equip rules.
```bash
cd server && go test ./internal/http -run TestCreatedCharactersReceiveClassStarterLoadouts
cd server && go test ./internal/game -run 'TestCharacterClassesAndStartingStats|TestRogueOffhandWeaponEquipRules'
```

## Task 3 - Rogue model and client picker

Files:
- Modify: `tools/assets/gen_glb.py`
- Add: `client/assets/characters/rogue/rogue.glb` and `.import` metadata if generated by Godot
- Modify: `assets/manifests/assets.v0.json`
- Modify: `client/scripts/character_select_panel.gd`
- Modify: `client/tests/test_coop_client.gd`
- Modify: `client/tests/test_animation.gd`

- [x] Step 3.1: Add deterministic Rogue GLB generation and asset manifest entry.
- [x] Step 3.2: Update the class picker to include Rogue.
- [x] Step 3.3: Add/adjust client tests for four options, Rogue selection, and Rogue model bones.
```bash
make gen-assets
make client-unit
```

## Task 4 - Bot proof

Files:
- Add: `tools/bot/scenarios/47_rogue_class_foundation.json`
- Modify: `tools/bot/test_protocol.py`

- [x] Step 4.1: Add protocol scenario that creates a Rogue and asserts main/off-hand swords plus potions.
- [x] Step 4.2: Add scenario discovery/unit assertions.
```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
make bot scenario=47_rogue_class_foundation
```

## Task 5 - Lifecycle docs and CI

Files:
- Modify: `docs/specs/v98_spec-rogue-class-foundation.md`
- Modify: `docs/plans/v98_2026-06-12-rogue-class-foundation.md`
- Add: `docs/as-built/v98_rogue-class-foundation.md`
- Modify: `PROGRESS.md`

- [x] Step 5.1: Mark spec/plan complete and write as-built notes.
- [x] Step 5.2: Update `PROGRESS.md` current status and lifecycle row.
```bash
make maintainability
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/http -run TestCreatedCharactersReceiveClassStarterLoadouts`
- [x] `cd server && go test ./internal/game -run 'TestLoadRules|TestRogueOffhandWeaponEquipRules|TestEquipmentWrongSlotRejects'`
- [x] `.venv/bin/pytest tools/bot/test_protocol.py -q`
- [x] `make client-unit`
- [x] `make bot scenario=47_rogue_class_foundation`
- [x] `make ci`

## Deferred scope

- Poison Stab DOT skill.
- Dash movement/damage skill.
- Off-hand independent attack cadence and damage events at 1.5x regular hand speed.
