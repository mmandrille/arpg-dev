# v506 — Living town

- **Status:** Implemented and batch-CI verified; resident-memory comparison remains unverified and is tracked as an evidence gap.
- **Date:** 2026-10-01
- **Codename:** `living-town`
- **Area:** Graphics / client presentation
- **Base:** `5365832b9029e0d9e178d57f2a95a6a697b01acc`
- **Depends on:** v500 town terrain and landmarks; no other batch slice is a prerequisite.
- **ADR:** [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md) D1–D4, D8–D9; [ADR-0007](../adr/0007-animation-state-model.md).

## Purpose

Make the town feel occupied through a very small group of animated, non-interactive residents staged among the existing stalls and props. Reuse the KayKit character models and animation clips already in the repository. Keep town services, routes, silhouettes, and combat readability clear.

Repository inspection found five KayKit hero GLBs (`kaykit_hero_{knight,barbarian,mage,rogue,ranger}_v0`) and the existing `kaykit_hero` clip profile (`Idle_A` and `Walking_A`). There is no dedicated town-NPC model pack in the runtime asset tree. The hero bodies can serve as anonymous residents without equipment or gameplay identity; this is the bounded placeholder-art choice for this slice.

## Non-goals

- Server NPCs, dialogue, quests, commerce, collision, navigation authority, or protocol changes.
- Moving or changing any service, spawn, gate, monster, or world-preset anchor.
- New character silhouettes, equipment overlays, or edits to `class_presentations.v0.json` / class asset registrations.
- New art packs, runtime asset downloads, external plugins, buildings, or a generalized crowd system.
- Changing the player animation state machine or combat animation behavior.

## Acceptance criteria

- [ ] A small data-defined cast (target: two residents) appears only in the level-0 town dressing. At least one has a short, slow walk-and-pause loop; the other idles near an existing stall/prop. Activity, transform, path, and playback pacing live in schema-backed presentation data.
- [ ] Residents use existing manifest-registered KayKit character scenes and the existing `Idle_A` / `Walking_A` clip library. Initialization and loop selection are stable for the same catalog; no wall-clock or server simulation state affects gameplay or replay.
- [ ] Residents have no collision, targeting, interactable identity, server entity, or gameplay effect. Dungeon entry removes their town root with the rest of the dressing.
- [ ] Every resident footprint and every point on a movement route meets a documented clearance from town interactable/spawn/monster anchors, service paths, existing props, the gate approach, and the palisade. Existing service approach and click targets remain usable.
- [ ] Real play-camera views at normal and maximum zoom show readable silhouettes and visible idle/movement motion without hiding the player, service targets, or town route landmarks.
- [ ] The matched town render-route comparison on Balanced and Performance tiers stays within the v500 budget: frame/process p95 delta no greater than `max(0.5 ms, 5%)` and draw-call delta no greater than `max(10 calls, 5%)`. If needed, lower the cast count or presentation cost while retaining the acceptance above.
- [ ] No external assets or plugins are required. The existing `town-play` scene capture includes the live dressing and records the visual result.

## Scope and likely files

- **Shared data:** `shared/assets/town_presentation.v0.json` and `.schema.json` for actor IDs, transforms, activity cycles/routes, and presentation pacing.
- **Client:** a focused town ambient actor/presentation helper under `client/scripts/`, wired through `client/scripts/town_dressing.gd` and, if needed, `town_presentation_loader.gd`. Keep the feature client-only and do not modify `main.gd` unless the existing dressing hook proves insufficient.
- **Proof:** a focused Godot test under `client/tests/`, clearance coverage in `tools/test_town_dressing.py`, and a client scenario `town_living_life` under `tools/bot/scenarios/client/` that captures the town and opens a nearby service.
- **Docs:** update `docs/CODEMAP.md`; leave lifecycle, `PROGRESS.md`, and as-built closeout to the coordinator after integration.
- **Evidence:** retain focused town-play captures and matched render-route measurements with the implementation handoff.

## Focused verification and visual proof

- `make validate-shared`
- `godot --headless --path client --script res://tests/test_town_ambient_life.gd`
- `.venv/bin/pytest -q tools/test_town_dressing.py`
- `make bot-visual scenario=town_living_life`
- `make bot-client SCENARIO=15_town_vendor_shop_panel HEADLESS=1`
- `make regen-screenshots SUITE="scenes"`; inspect the `town-play` normal/max-zoom captures.
- Compare the candidate with `docs/performance/fixtures/v500_town_render_route.json` using the same route, renderer, tiers, sample method, and interleaved runs as its documentation. Do not infer a performance pass from screenshots or headless checks.
- For interactive proof, run `make play`, enter the town, inspect the plaza/stall/gate at normal and maximum zoom while approaching and opening a service. Confirm the walk loop is visible and no resident obscures or blocks the service approach.
- `make maintainability` for touched files. The coordinator runs combined `make ci` only after the batch is integrated.

## Asset and plugin decision

- **Adopt:** the existing in-repo KayKit hero GLBs and the existing KayKit `Idle_A` / `Walking_A` clips as lightweight anonymous town residents. Keep current proportions and do not add equipment.
- **Borrow:** `KitPieceLibrary`, the existing hero clip profile/animation plumbing, `TownDressing`'s shared live/capture construction path, and the v500 town render-route fixture.
- **Reject:** external character models, asset downloads, a new animation pack, and third-party Godot plugins. No asset-manifest change is expected.

## Security and open decisions

- The requested security-router check classifies this presentation-only change as out of scope: it introduces no user-controlled rendering, network/API, persistence, auth, file input, or other security-relevant surface.
- No material product question blocks planning. The bounded choice is to reuse existing hero bodies as anonymous residents because the repo has no dedicated town NPC models; implementation should keep this choice isolated in town presentation data and should not change playable class catalogs.
