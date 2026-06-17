# v253 Plan - Fog of War Radius

Status: Complete
Goal: Add radial, server-authoritative fog of war with class/item light radius and client gloom/darkness presentation.
Architecture: Keep the authoritative sim truth unchanged, then derive recipient-scoped entity
views for snapshots and deltas. Living monster visibility is filtered by the recipient player's
effective `light_radius` and rectangular wall line-of-sight occlusion; walls and non-creature
entities remain visible so gloom can show obstacles without leaking creatures. Client fog is
presentation-only and driven from the server's derived stat.
Tech stack: shared JSON/schema, Go simulation and realtime fanout, Python protocol bot, Godot
client, docs.

## Baseline and Shortcut Decision

Builds on v252 expanded dungeon profiles, the current recipient-scoped co-op snapshot path, and
existing generated wall rendering. This slice intentionally avoids durable exploration memory and
richer doorway/high-obstacle line-of-sight blocking.

Asset/plugin decision: reject external assets/plugins for this slice. Borrow existing in-repo
Godot code-native overlay and primitive rendering patterns (`input_shadow_overlay.gd`, wall/entity
presentation, client bot debug state); implement a focused `fog_of_war_overlay.gd` instead of
adding an asset pipeline.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/character_progression.v0.json` | Add per-class `light_radius` baselines and derived stat formula |
| Modify | `shared/rules/character_progression.v0.schema.json` | Validate class light-radius values and derived stat |
| Modify | `shared/rules/item_templates.v0.json` | Add `light_radius` to selected rollable equipment pools |
| Modify | `shared/rules/item_templates.v0.schema.json` | Allow `light_radius` as an item stat |
| Modify | `shared/golden/character_progression.json` | Keep progression formula goldens aware of `light_radius` |
| Modify | `shared/golden/character_progression.v0.schema.json` | Validate the golden derived stat field |
| Modify | `shared/protocol/session_snapshot.v8.schema.json` | Allow `derived_stats.light_radius` and stat-breakdown key |
| Modify | `shared/protocol/state_delta.v8.schema.json` | Allow progression deltas containing `light_radius` |
| Modify | `shared/protocol/examples/session_snapshot.json` | Include representative `light_radius` value |
| Modify | `shared/protocol/examples/state_delta.json` | Include representative `light_radius` value when needed by validation |
| Modify | `server/internal/game/rules.go` | Load and validate class light-radius rules and item stat support |
| Modify | `server/internal/game/derived_stats.go` | Add `LightRadius` to the derived stat protocol view |
| Modify | `server/internal/game/sim.go` | Add effective stat wiring and minimal calls into focused visibility helpers |
| Add | `server/internal/game/fog_of_war.go` | Recipient-scoped radial visibility helpers and delta filtering |
| Add | `server/internal/game/fog_of_war_test.go` | Class baseline, item roll, snapshot, and delta visibility tests |
| Modify | `server/internal/realtime/session_loop.go` | Use game visibility filtering after existing private-change filtering |
| Modify | `server/internal/realtime/runner.go` | Use game visibility filtering for the solo runner path |
| Add | `server/internal/realtime/fog_of_war_test.go` | Realtime fanout regression tests for hidden monster deltas |
| Add | `tools/bot/scenarios/92_fog_of_war_radius.json` | Protocol proof for hidden/visible monster information |
| Add | `client/scripts/fog_of_war_overlay.gd` | Code-native radial light/gloom/darkness overlay |
| Modify | `client/scripts/main.gd` | Wire overlay with minimal net line growth and expose bot debug state |
| Modify | `client/scripts/stat_labels.gd` | Display `Light radius` in item/stat labels |
| Modify | `client/scripts/character_stats_panel.gd` | Display `light_radius` in derived stats |
| Add | `client/tests/test_fog_of_war_overlay.gd` | Headless unit coverage for radius state/debug values |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Add client fog debug assertion |
| Add | `tools/bot/scenarios/client/67_fog_of_war_overlay.json` | Client visual/debug proof for overlay radii |
| Modify | `docs/specs/v253_spec-fog-of-war-radius.md` | Mark complete when shipped |
| Modify | `docs/progress/slice-lifecycle.md` | Add v253 lifecycle row |
| Add | `docs/as-built/v253_fog-of-war-radius.md` | Record shipped behavior and deferrals |
| Modify | `PROGRESS.md` | Update latest slice and deferred fog/LOS scope |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [ ] `client/scripts/main.gd` baseline 5735, current 5759. Keep net change at or below +1, or
  extract enough existing logic so the file stays within the +25 allowance.
- [ ] `server/internal/game/sim.go` baseline 6572, current 6571. Keep net change non-positive by
  putting fog logic in `fog_of_war.go`.
- [ ] `server/internal/game/rules.go` baseline 3303, current 3308. Keep growth minimal and offset
  if needed; validation logic should stay narrow.
- [ ] `server/internal/realtime/runner.go` baseline 627, current 642. Keep net change at or below
  +10 and prefer helper calls.
- [ ] `server/internal/realtime/session_loop.go` baseline 924, current 943. Keep net change at or
  below +6 and prefer helper calls.
- [ ] `tools/validate_shared.py` baseline 3017, current 3030. Do not edit unless validation
  requires it; prefer JSON schema support first.
- [ ] Did every touched grandfathered file stay at or below its baseline allowance?

Decision:
- [ ] Extract focused helper/module/test files: `server/internal/game/fog_of_war.go`,
  `server/internal/game/fog_of_war_test.go`, `server/internal/realtime/fog_of_war_test.go`,
  `client/scripts/fog_of_war_overlay.gd`, and `client/tests/test_fog_of_war_overlay.gd`.

Verification:
```bash
make maintainability
```

## Task 1 - Shared Light Radius Data

Files:
- Modify: `shared/rules/character_progression.v0.json`
- Modify: `shared/rules/character_progression.v0.schema.json`
- Modify: `shared/rules/item_templates.v0.json`
- Modify: `shared/rules/item_templates.v0.schema.json`
- Modify: protocol examples/schemas as needed for `light_radius`

- [x] Add class-level `light_radius` values: barbarian/paladin/rogue 9, sorcerer 10, ranger 12.
- [x] Add a derived stat entry for `light_radius` with additive item support.
- [x] Allow equipment rollable/base stat `light_radius` and add conservative rolls to accessory or
  armor pools without changing fixed starter loadouts.
- [x] Update schema/example validation for `derived_stats.light_radius` and stat breakdown keys.

```bash
make validate-shared
```

## Task 2 - Server Effective Stat and Visibility Filtering

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/derived_stats.go`
- Modify: `server/internal/game/sim.go`
- Add: `server/internal/game/fog_of_war.go`
- Add: `server/internal/game/fog_of_war_test.go`

- [x] Load/validate class light radius from shared character progression rules.
- [x] Add effective `LightRadius` calculation and stat breakdown sources for class baseline and
  equipped item rolls.
- [x] Filter recipient snapshots so living monsters outside the light radius are omitted, while
  players, companions, loot, projectiles, interactables, and walls remain visible.
- [x] Filter living monsters inside the light radius when rectangular wall obstacles block line of
  sight from the recipient hero.
- [x] Filter recipient deltas so hidden living monster updates are suppressed, visible transitions
  spawn/update, and hidden transitions remove previously visible monsters.
- [x] Add focused tests for class baselines, equipped `light_radius`, hidden/visible snapshot
  contents, and visibility transition deltas.

```bash
cd server && go test ./internal/game -run 'TestFogOfWar|TestLightRadius|TestCharacterStats'
```

## Task 3 - Realtime Fanout Integration

Files:
- Modify: `server/internal/realtime/session_loop.go`
- Modify: `server/internal/realtime/runner.go`
- Add: `server/internal/realtime/fog_of_war_test.go`

- [x] Apply game-level recipient visibility filtering after existing owner/private-change filters
  in the co-op session loop.
- [x] Apply the same filtering in the solo runner path.
- [x] Add realtime tests proving far monster deltas do not reach a client and near monster deltas
  do.

```bash
cd server && go test ./internal/realtime -run 'TestFogOfWar|Test.*Deltas.*Scoped|TestShopDeltasAreActorScoped'
```

## Task 4 - Protocol Bot Proof

Files:
- Add: `tools/bot/scenarios/92_fog_of_war_radius.json`
- Reuse: existing `character_progression` and `entity_count` runtime assertions

- [x] Add assertions for `character_progression.derived_stats.light_radius`.
- [x] Add visible-entity count assertions that prove a far monster is absent from the current
  authoritative client view and a nearby monster is present after movement.
- [x] Keep the scenario under the 10 second protocol-bot budget.

```bash
make bot scenario=92_fog_of_war_radius
```

## Task 5 - Client Fog Presentation

Files:
- Add: `client/scripts/fog_of_war_overlay.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/stat_labels.gd`
- Modify: `client/scripts/character_stats_panel.gd`
- Add: `client/tests/test_fog_of_war_overlay.gd`
- Modify: `client/scripts/bot_assertion_handlers.gd`
- Add: `tools/bot/scenarios/client/67_fog_of_war_overlay.json`

- [x] Render code-native clear/gloom/opaque-darkness bands around the local hero using the
  server-derived light radius and `gloom_radius = light_radius * 1.25`.
- [x] Keep the overlay out of UI layers and expose bot debug state with `enabled`,
  `light_radius`, and `gloom_radius`.
- [x] Display `Light radius` in stat labels and the character stats derived panel.
- [x] Add client unit coverage and a client bot scenario that checks overlay debug state.

```bash
make client-unit
HEADLESS=1 make bot-visual scenario=67_fog_of_war_overlay
```

## Task 6 - Lifecycle Docs

Files:
- Modify: `docs/specs/v253_spec-fog-of-war-radius.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v253_fog-of-war-radius.md`
- Modify: `PROGRESS.md`
- Modify: `docs/progress/scenario-catalog.md` if the catalog is maintained for new scenarios

- [x] Mark the spec complete.
- [x] Add v253 lifecycle and as-built notes.
- [x] Update `PROGRESS.md` current status and defer durable explored map memory, richer
  doorway/high-obstacle line-of-sight blocking, and production lighting/art.

```bash
make maintainability
```

## Final Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestFogOfWar|TestLightRadius|TestCharacterStats'`
- [x] `cd server && go test ./internal/realtime -run 'TestFogOfWar|Test.*Deltas.*Scoped|TestShopDeltasAreActorScoped'`
- [x] `make client-unit`
- [x] `make bot scenario=92_fog_of_war_radius`
- [x] `HEADLESS=1 make bot-visual scenario=67_fog_of_war_overlay`
- [x] `make bot scenario=13_teleporter_lab,14_dungeon_monsters,24_boss_floor_gate,36_account_stash_storage,42_pack_aggro_and_dungeon_packs,59_ranger_piercing_and_pinning_shots`
- [x] `make maintainability`
- [x] `make ci`

Autoloop batch note: final batch `make ci` passed after the selected queue was committed.
