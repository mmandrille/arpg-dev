# v101 Plan — Undead Skeleton Poison Immunity

Status: Complete
Goal: Add an undead skeleton enemy that can be poisoned but takes zero poison damage through full poison resistance.
Architecture: Shared rules own the undead definition and poison resistance. The server continues to own poison application, status state, and damage mitigation, while the client only renders the catalog-selected skeleton model. Bot proof exercises the production protocol against a compact lab world.
Tech stack: Go sim, shared JSON/schema catalogs, deterministic GLB generator, Godot client scene registry, Python protocol bot.

## Baseline and Shortcut Decision

Builds on v100 `damage-types-and-resistances`, which added data-driven damage types, monster resistance maps, and combat event `damage_type`.

Godot plugin / asset decision:
- Adopt: none.
- Borrow: existing deterministic generated GLB pipeline, asset manifest, monster visual catalog, and low-poly monster scene pattern.
- Reject: external Godot plugins and third-party skeleton art packs for this narrow CI-backed proof.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/monsters.v0.json` | Add undead monster with `poison: 1.0` resistance |
| Modify | `shared/rules/worlds.v0.json` | Add undead poison immunity lab world |
| Modify | `shared/i18n/en.json`, `shared/i18n/es.json` | Localize undead name |
| Modify | `shared/assets/monster_visuals.v0.schema.json` | Allow `monster_skeleton` visual model |
| Modify | `shared/assets/monster_visuals.v0.json` | Map undead to skeleton visual |
| Modify | `assets/manifests/assets.v0.json` | Add skeleton runtime GLB manifest entry |
| Modify | `tools/assets/gen_glb.py` | Generate deterministic skeleton GLB |
| Create | `client/assets/monsters/skeleton/monster_skeleton.glb` | Runtime generated skeleton model |
| Create | `client/scenes/monster_skeleton.tscn` | Godot scene wrapping skeleton GLB |
| Modify | `client/scripts/monster_visuals_loader.gd` | Accept skeleton visual id |
| Modify | `client/scripts/main.gd` | Register skeleton monster scene |
| Modify | `client/tests/test_animation.gd` | Cover skeleton scene/catalog entry |
| Modify | `server/internal/game/rogue_skills.go` | Apply poison on connected hit even when final damage is zero |
| Modify | `server/internal/game/damage_types_test.go` | Prove poison immunity status plus zero poison damage |
| Modify | `tools/bot/run.py` | Add event matcher support for damage/damage_type if needed |
| Create | `tools/bot/scenarios/49_undead_skeleton_poison_immunity.json` | Protocol proof |
| Create | `docs/as-built/v101_undead-skeleton-poison-immunity.md` | As-built summary |
| Modify | `PROGRESS.md` | Mark v101 complete and set next slice |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [ ] `server/internal/game/game_test.go`
- [x] `tools/bot/run.py`
- [ ] `tools/validate_shared.py`
- [ ] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected

Decision:
- [ ] Extract focused helper/module/test file as part of this slice, or
- [x] Defer extraction with rationale: `main.gd` and `run.py` only receive narrow registry/assertion additions that are lower risk than a structural split during an asset/combat slice.

Verification:
```bash
make maintainability
```

## Task 1 — Shared Enemy and Lab Data

Files:
- Modify: `shared/rules/monsters.v0.json`
- Modify: `shared/rules/worlds.v0.json`
- Modify: `shared/i18n/en.json`
- Modify: `shared/i18n/es.json`

- [x] Add `dungeon_undead` with poison resistance `1.0` and ordinary chase combat stats.
- [x] Add a compact `undead_skeleton_poison_immunity` lab world containing an undead target in poison-stab range.
- [x] Add localized undead display names.

```bash
make validate-shared
```

## Task 2 — Skeleton Asset and Client Catalog

Files:
- Modify: `shared/assets/monster_visuals.v0.schema.json`
- Modify: `shared/assets/monster_visuals.v0.json`
- Modify: `assets/manifests/assets.v0.json`
- Modify: `tools/assets/gen_glb.py`
- Create: `client/assets/monsters/skeleton/monster_skeleton.glb`
- Create: `client/scenes/monster_skeleton.tscn`
- Modify: `client/scripts/monster_visuals_loader.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/tests/test_animation.gd`

- [x] Add the `monster_skeleton` visual model to schemas, catalog, manifest, generator, and client scene registry.
- [x] Regenerate deterministic runtime assets.
- [x] Cover the skeleton scene and catalog mapping in headless client tests.

```bash
make gen-assets
make validate-assets
make client-unit
```

## Task 3 — Poison Immunity Semantics

Files:
- Modify: `server/internal/game/rogue_skills.go`
- Modify: `server/internal/game/damage_types_test.go`

- [x] Change poison application to key off a connected poison-applying hit, not positive final damage.
- [x] Allow poison DOT setup from zero final hit damage so immunity can keep emitting zero poison ticks.
- [x] Add a focused Go test proving the undead becomes poisoned and poison damage resolves to zero.

```bash
cd server && go test ./internal/game/... -run 'TestDamageType|TestResistance|TestPoison'
```

## Task 4 — Bot Scenario Proof

Files:
- Modify: `tools/bot/run.py`
- Create: `tools/bot/scenarios/49_undead_skeleton_poison_immunity.json`

- [x] Add bot event matcher support for `damage_type` and `damage` if current assertions cannot count zero poison events.
- [x] Add a protocol scenario that uses Poison Stab on the undead, waits for poison ticks, and asserts at least two zero-damage poison events plus a poison effect start.

```bash
make bot scenario=undead_skeleton_poison_immunity
```

## Task 5 — Lifecycle Docs and CI

Files:
- Create: `docs/as-built/v101_undead-skeleton-poison-immunity.md`
- Modify: `PROGRESS.md`

- [x] Document the implemented behavior and visual proof path.
- [x] Mark v101 complete, add lifecycle links, and set v102 as the next slice placeholder.

```bash
make maintainability
make ci
```

## Final Verification

- [x] `make validate-shared`
- [x] `make validate-assets`
- [x] `cd server && go test ./internal/game/... -run 'TestDamageType|TestResistance|TestPoison'`
- [x] `make bot scenario=undead_skeleton_poison_immunity`
- [x] `make client-unit`
- [x] `make maintainability`
- [x] `make ci`

Visual verification command for manual review:

```bash
make bot-visual scenario=undead_skeleton_poison_immunity
```
