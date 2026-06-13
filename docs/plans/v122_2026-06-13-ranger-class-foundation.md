# v122 Plan - Ranger class foundation

Status: Complete
Goal: Add Ranger as a selectable, persistent bow class with a tall hooded model, green bow icon, starter bow, potions, and bot proof.
Architecture: Gameplay authority stays in shared rules and the Go server. The Godot client consumes class presentation metadata and deterministic GLB assets. Starter items are durable inventory rows seeded by the HTTP character-creation path. Active Ranger skills are deferred to follow-up slices.
Tech stack: shared JSON rules/assets, Go server/game/http tests, Godot client scripts/tests, deterministic GLB generation, Python protocol bot, lifecycle docs.

## Baseline and shortcut decision

Builds on v69-v71 class identity/picker/presentation, v97 starter loadouts, v98 Rogue foundation,
and existing ranged basic-attack support. Godot plugin adoption: **reject external plugins and asset
packs** for this slice. Existing deterministic GLB generation, class presentation metadata, and item
visual loaders cover the required art and UI surface with less dependency risk.

## Scope split

Included in v122:
- Ranger class rules, picker visibility, class presentation, model, starter bow, potions, and bot
  proof of ranged basic attacks.

Deferred:
- Piercing Shot, Pinning Shot, Volley, and skill visual showcase.
- Production art beyond deterministic placeholder GLB and vector class icon.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/character_progression.v0.json` | Add Ranger class stats. |
| Modify | `shared/rules/items.v0.json`, `shared/rules/items.v0.schema.json` | Add Ranger class bow item and allow Ranger class requirements. |
| Modify | `shared/rules/item_templates.v0.json` | Add `starter_ranger_bow` template for starter roll payloads. |
| Modify | `shared/assets/class_presentations.v0.json` | Add Ranger icon/model mapping. |
| Modify | `shared/assets/item_presentations.v0.json`, `shared/assets/item_visuals.v0.json` | Add starter bow presentation/visuals. |
| Modify | `assets/manifests/assets.v0.json`, `tools/assets/gen_glb.py` | Register and generate hooded Ranger GLB. |
| Modify | `server/internal/http/starter_loadout.go`, `server/internal/http/starter_loadout_test.go` | Seed Ranger starter bow and potions. |
| Modify | `server/internal/game/game_test.go`, `tools/validate_shared.py` | Validate class/weapon coverage. |
| Modify | `client/scripts/character_select_panel.gd`, `client/scripts/class_icon.gd` | Expose Ranger and draw bow icon. |
| Modify | `shared/i18n/en.json`, `shared/i18n/es.json` | Add Ranger text. |
| Add/Modify | `tools/bot/scenarios/58_ranger_class_foundation.json`, `tools/bot/test_protocol.py` | Protocol proof. |
| Add | `docs/as-built/v122_ranger-class-foundation.md` | As-built summary. |
| Modify | `PROGRESS.md` | Lifecycle close-out. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/character_select_panel.gd`
- [x] `server/internal/game/game_test.go`
- [x] `tools/bot/test_protocol.py`
- [x] `tools/validate_shared.py`

Decision:
- [ ] Extract focused helper/module/test file as part of this slice, or
- [x] Defer extraction with rationale: edits are narrow class-data additions to existing class
  coverage surfaces; extraction would be larger than the feature.
- [x] Documented maintenance exception: update the grandfathered baseline for touched v122 files
  plus current drift in `client/scripts/inventory_panel.gd`, `client/scripts/main.gd`,
  `server/internal/game/handlers.go`, `server/internal/game/shop.go`, `server/internal/game/sim.go`,
  and `server/internal/replay/replay_test.go` so the ratchet enforces future growth from the current
  repo state.

Verification:
```bash
make maintainability
```

## Task 1 - Shared class and starter data

Files:
- Modify: `shared/rules/character_progression.v0.json`
- Modify: `shared/rules/items.v0.json`
- Modify: `shared/rules/items.v0.schema.json`
- Modify: `shared/rules/item_templates.v0.json`
- Modify: `shared/assets/class_presentations.v0.json`
- Modify: `shared/assets/item_presentations.v0.json`
- Modify: `shared/assets/item_visuals.v0.json`
- Modify: `shared/i18n/en.json`, `shared/i18n/es.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Add Ranger class stats and Ranger class-required starter bow item/template.
- [x] Step 1.2: Add Ranger class presentation, starter bow presentation, visuals, and text keys.
- [x] Step 1.3: Update validation coverage for five classes and Ranger class weapon.
```bash
make validate-shared
```

## Task 2 - Server starter kit and rule tests

Files:
- Modify: `server/internal/http/starter_loadout.go`
- Modify: `server/internal/http/starter_loadout_test.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 2.1: Seed Ranger starter bow in `main_hand`, plus one red and one blue potion.
- [x] Step 2.2: Add focused tests for Ranger stats, weapon class requirement, and starter loadout.
```bash
cd server && go test ./internal/http -run TestCreatedCharactersReceiveClassStarterLoadouts
cd server && go test ./internal/game -run TestLoadRules
```

## Task 3 - Ranger model and client picker

Files:
- Modify: `tools/assets/gen_glb.py`
- Modify: `assets/manifests/assets.v0.json`
- Add: `client/assets/characters/ranger/ranger.glb`
- Modify: `client/scripts/character_select_panel.gd`
- Modify: `client/scripts/class_icon.gd`

- [x] Step 3.1: Add deterministic tall, thin hooded Ranger GLB generation and manifest entry.
- [x] Step 3.2: Update the class picker and icon drawing to include Ranger.
```bash
make gen-assets
make client-unit
```

## Task 4 - Bot proof

Files:
- Add: `tools/bot/scenarios/58_ranger_class_foundation.json`
- Modify: `tools/bot/test_protocol.py`

- [x] Step 4.1: Add protocol scenario that creates a Ranger and asserts starter bow and class stats.
- [x] Step 4.2: Have the scenario attack with the starter bow and assert a ranged basic attack event.
- [x] Step 4.3: Add scenario discovery/unit assertions.
```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
make bot scenario=58_ranger_class_foundation
```

## Task 5 - Lifecycle docs and CI

Files:
- Modify: `docs/specs/v122_spec-ranger-class-foundation.md`
- Modify: `docs/plans/v122_2026-06-13-ranger-class-foundation.md`
- Add: `docs/as-built/v122_ranger-class-foundation.md`
- Modify: `PROGRESS.md`

- [x] Step 5.1: Mark spec/plan complete and add as-built summary.
- [x] Step 5.2: Update `PROGRESS.md` latest completed slice and lifecycle list.
```bash
make maintainability
make ci
```

## Final verification

- [x] `make gen-assets`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/http -run TestCreatedCharactersReceiveClassStarterLoadouts`
- [x] `cd server && go test ./internal/game -run TestLoadRules`
- [x] `.venv/bin/pytest tools/bot/test_protocol.py -q`
- [x] `make bot scenario=58_ranger_class_foundation`
- [x] `make client-unit`
- [x] `make maintainability`
- [x] `make ci`
