# v253 Spec - Fog of War Radius

Status: Complete
Date: 2026-06-17
Codename: fog-of-war-radius

## Purpose

Introduce the first server-authoritative fog-of-war mechanic so dungeon play no longer gives the
client perfect creature information for the entire level. Each hero receives a calculated
`light_radius` derived from class baseline plus equipped item rolls. The server only reveals
creatures inside the light radius, while the client presents three radial bands around the local
hero: clear light, darker gloom, and total darkness.

## Non-goals

- No wall, doorway, high-obstacle, or line-of-sight occlusion. Obstacles blocking enemy visibility
  inside the light radius are deferred to a later slice.
- No durable explored-map memory across sessions or level revisits. This slice is radius-based
  current visibility only.
- No stealth, scouting units, monster AI awareness changes, aggro tuning, minimap memory, or PvP
  information rules.
- No external art, asset pipeline, shader plugin, or Godot addon adoption. Use code-native Godot
  overlay/mesh primitives and existing wall/entity presentation patterns.
- No schema major-version bump unless required by validation; keep changes additive to v8 payload
  shapes.

## Acceptance Criteria

- Shared character progression rules define per-class base light radius: barbarian, paladin, and
  rogue use 9; sorcerer uses 10; ranger uses 12.
- Rolled equipment templates can roll `light_radius`, and equipped light-radius rolls increase the
  hero's effective calculated stat.
- `character_progression.derived_stats.light_radius` and stat breakdowns expose the effective
  value so UI, bots, and clients can inspect it.
- Recipient-scoped snapshots and state deltas include player, companion, loot, projectile, and
  interactable entities on the current level as before, but hide living monsters outside the
  recipient hero's light radius.
- A monster that moves from hidden to visible emits an entity spawn/update to that recipient; a
  monster that moves from visible to hidden emits an entity remove to that recipient.
- Gloom extends from `light_radius` through `light_radius * 1.25`. Obstacles/walls in this band
  remain visible through the existing wall layout, but creatures are not revealed there.
- The Godot client renders clear light around the local hero, a darker/greyed gloom band outside
  it, and near-black total darkness beyond the gloom radius without hiding UI.
- Bot/client debug state exposes enough visibility/fog data to prove that far monsters are absent,
  near monsters are visible, and the client knows the current light/gloom radii.

## Scope and Likely Files

- Shared rules/schema: `shared/rules/character_progression.v0.json`,
  `shared/rules/character_progression.v0.schema.json`, `shared/rules/item_templates.v0.json`,
  `shared/rules/item_templates.v0.schema.json`.
- Protocol schemas/examples: `shared/protocol/session_snapshot.v8.schema.json`,
  `shared/protocol/state_delta.v8.schema.json`, example snapshot/delta payloads if validation
  requires new `light_radius` fields.
- Server: focused visibility/stat helpers plus existing snapshot/delta routing in
  `server/internal/game/` and `server/internal/realtime/`.
- Client: a focused fog overlay/presentation script wired from `client/scripts/main.gd`, plus stat
  label/panel display updates.
- Bot proof: new protocol bot scenario for authoritative visibility and a client bot scenario for
  fog presentation/debug state.
- Docs: plan, as-built, lifecycle, and `PROGRESS.md` close-out.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestFogOfWar|TestLightRadius|TestCharacterStats'`
- `cd server && go test ./internal/realtime -run 'TestFogOfWar|Test.*Deltas.*Scoped|TestShopDeltasAreActorScoped'`
- `make client-unit`
- `make bot scenario=92_fog_of_war_radius`
- `HEADLESS=1 make bot-visual scenario=67_fog_of_war_overlay`
- `make maintainability`

## Open Questions and Risks

- No required questions for this run.
- Risk: adding visibility removes public entity updates for hidden monsters, so recipient-specific
  visibility state must be deterministic and must not leak hidden creature positions through
  events or debug payloads.
- Risk: this slice changes information visibility but not server combat/AI, so monsters may still
  aggro or attack based on existing rules when the server deems them in range. Any tuning change to
  monster awareness is deferred.
- Risk: client fog presentation should be proven through debug state and visual scenario commands;
  production-quality dungeon lighting/art remains deferred.
