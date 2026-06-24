# v331 Spec — Hero Visibility Lighting

**Status:** Draft  
**Date:** 2026-06-23  
**Codename:** hero-visibility-lighting

---

## Purpose

Replace the isometric-only screen-radius fog overlay with a **camera-agnostic visibility
compositor** driven by a world-space hero light field. Dungeons render as true darkness outside the
hero's authoritative `light_radius`, with smooth physical falloff toward the edge, line-of-sight
shadow wedges from walls and closed doors, and ambient lighting suppression so geometry outside the
lit region is actually hidden — not merely tinted. The compositor is active in **isometric**,
**third_person**, and **chest_view** during gameplay, closing the v329 deferral for perspective fog.

Server monster visibility (`light_radius` + LOS) is unchanged.

---

## Non-goals

- No server gameplay visibility, `light_radius` formula, protocol, replay, or golden combat changes.
- No durable explored-map / minimap memory persistence.
- No monster fog-aware AI or aggro changes.
- No non-rectangular / destructible / polygon occluder expansion.
- No production fog art, particles, imported shaders, or Godot addons.
- No full merge with dungeon-depth lighting profiles beyond documented ambient suppression when fog
  is active (v318 profiles remain the base; fog scales ambient down in dungeons).
- No town fog (`level >= 0` stays fully lit with compositor off).
- No scroll-morph cameras or click-to-move in perspective.

---

## Acceptance criteria

### Shared presentation data

- `shared/assets/fog_presentation.v0.json` + schema define client-only tuning: falloff curve
  (`falloff_power`, `edge_feather_world`), shadow reach multiplier (replaces hardcoded gloom extend
  for LOS polygons), ambient suppression scales for dungeon fog, organic-edge flags per camera mode,
  and darkness color/alpha.
- `FogPresentationLoader` (`class_name` + `ensure_loaded()`) loads the catalog; gameplay code does
  not hardcode tuning literals.

### World-space visibility compositor

- `FogOfWarOverlay` (or successor compositor) samples **world floor position (y=0)** per screen pixel
  via camera unproject + ground-plane intersection, then computes visibility from distance to the
  hero and the data-driven falloff curve.
- Outside `light_radius` (+ configured feather), compositor reaches **full black** (`alpha = 1`).
- Inside radius, brightness falls off smoothly toward the edge (inverse-power falloff from shared
  data).
- Existing LOS shadow polygons (walls with `blocks_line_of_sight`, closed doors) still cast
  darkness wedges using the same occluder feed; water/holes/rubble remain non-occluding.
- Organic edge variation remains available; default **on** for isometric, **off** for perspective
  until tuned (data-driven per mode).

### Camera modes and lighting

- Fog compositor is **active in all three gameplay camera modes** when `light_radius > 0` and current
  level is a dungeon (`level < 0`).
- Dungeon ambient/directional lighting is scaled down while fog is active so outside-radius geometry
  is visually hidden, not only overlay-tinted.
- Town and zero `light_radius` disable the compositor (regression).

### Authority regression

- `make bot scenario=92_fog_of_war_radius` still passes (server visibility unchanged).
- `HEADLESS=1 make bot-visual scenario=67_fog_of_war_overlay` passes with updated falloff debug
  fields.
- `HEADLESS=1 make bot-visual scenario=68_fog_los_shadow_mask` passes (LOS shadows preserved).
- `HEADLESS=1 make bot-visual scenario=83_camera_mode_setting` asserts fog active in perspective
  modes when in a fog lab world.

### Tests

- `client/tests/test_fog_of_war_overlay.gd` covers world-falloff debug state, all-camera-mode
  activation hook, and existing LOS/occluder cases.
- `make client-unit` and `make maintainability` pass.

---

## Scope and likely files

| Area | Path |
|------|------|
| Shared | `shared/assets/fog_presentation.v0.json`, `fog_presentation.v0.schema.json` |
| Client loader | `client/scripts/fog_presentation_loader.gd` |
| Visibility field | `client/scripts/hero_visibility_field.gd` (occluder normalize, LOS shadow polygons) |
| Compositor | `client/scripts/fog_of_war_overlay.gd` (world-space shader, loader wiring) |
| Integration | `client/scripts/main.gd` (all-camera fog, ambient suppression) |
| Lighting | `client/scripts/dungeon_depth_lighting.gd` (optional helper for suppressed profile) |
| Tests | `client/tests/test_fog_of_war_overlay.gd`, `scripts/client_smoke.sh` |
| Bot | `tools/bot/scenarios/client/67_*.json`, `68_*.json`, `83_*.json`; `bot_assertion_handlers.gd` |
| Docs | plan, as-built, lifecycle on `/finish` |

### Asset / plugin decision

| Choice | Decision |
|--------|----------|
| `fog_presentation.v0.json` | **Adopt** — presentation tuning owner |
| `hero_visibility_field.gd` | **Adopt** — extracted LOS/occluder math |
| Existing `FogOfWarOverlay` + `Polygon2D` shadows | **Borrow** |
| External fog shaders / addons / art | **Reject** |

---

## Test and bot proof

```bash
make validate-shared   # if shared validators extended; schema file present regardless
make client-unit
HEADLESS=1 make bot-visual scenario=67_fog_of_war_overlay
HEADLESS=1 make bot-visual scenario=68_fog_los_shadow_mask
HEADLESS=1 make bot-visual scenario=83_camera_mode_setting
make bot scenario=92_fog_of_war_radius
make maintainability
make ci
```

Manual visual:

```bash
make play   # cycle V through modes in a dungeon; verify darkness + hero lamp in all modes
```

---

## Open questions and risks

| # | Item | Resolution |
|---|------|------------|
| Q-1 | Gloom middle band | **Resolved:** drop visible gloom band; lit → black with small data-driven edge feather; keep `shadow_reach_multiplier` for LOS polygon extent in debug |
| Q-2 | World-space compositor vs screen overlay | **Resolved:** world-space compositor (Option C) |
| Q-3 | Organic edge in perspective | Default off in data; isometric on |
| Risk | `main.gd` size | Narrow wiring only; extract field module |
| Risk | Bot asserts on `gloom_radius` | Map debug `gloom_radius` → `light_radius * shadow_reach_multiplier` for compatibility |
| Risk | Shader ground-plane assumption | Document y=0 floor; matches current dungeon layout |

---

## ADR alignment

- **ADR-0001 D2:** presentation-only; server owns monster visibility.
- **ADR-0007:** animation unaffected.
- **ADR-0008:** occluder kinds unchanged; no new server LoS table.
- **v329:** closes deferred perspective fog polish.
