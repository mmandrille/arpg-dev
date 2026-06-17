# v253 As-Built - Fog of War Radius

Date: 2026-06-17

## What shipped

- Added a calculated `light_radius` derived stat, with class baselines from shared rules:
  Barbarian, Paladin, and Rogue start at 9; Sorcerer starts at 10; Ranger starts at 12.
- Added `light_radius` as an additive equipment stat for selected accessory roll pools.
- Exposed `derived_stats.light_radius` and stat-breakdown sources through v8 snapshot/delta
  protocol schemas, examples, shared validation, and the character stats UI.
- Enabled server-authoritative radial fog by default for recipient-scoped snapshots and deltas.
- Hidden living monsters outside the recipient hero's light radius are omitted from snapshots,
  update deltas, and referenced events.
- Rectangular wall obstacles now block living monster visibility inside the light radius; walls
  and other non-creature entities remain visible.
- Player movement can now emit fog transition deltas: hidden monsters spawn when scouted into
  light radius and visible monsters are removed when they leave it.
- Kept players, companions, loot, projectiles, interactables, wall layouts, and non-creature
  entities visible under the existing recipient filters.
- Added a code-native Godot overlay with clear light, a `light_radius * 1.25` gloom band, and
  opaque darkness outside the gloom radius.
- Added client bot debug state and an `assert_fog_of_war` client assertion for overlay radii.
- Added protocol and client bot scenarios for the authoritative visibility and overlay proofs.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestFogOfWar|TestLightRadius|TestCharacterStats'
cd server && go test ./internal/realtime -run 'TestFogOfWar|Test.*Deltas.*Scoped|TestShopDeltasAreActorScoped'
make client-unit
make bot scenario=92_fog_of_war_radius
HEADLESS=1 make bot-visual scenario=67_fog_of_war_overlay
make bot scenario=13_teleporter_lab,14_dungeon_monsters,24_boss_floor_gate,36_account_stash_storage,42_pack_aggro_and_dungeon_packs,59_ranger_piercing_and_pinning_shots
make ci
```

All focused checks and the final full `make ci` passed on 2026-06-17 during `$autoloop`. The final
CI pass required refreshing `81_magic_find_stat` to a new deterministic ring seed after adding
`light_radius` to the ring roll pool changed the old seed's rolled stats. A follow-up bug fix made
darkness fully opaque and added rectangular wall line-of-sight occlusion for living monsters; its
focused proof was:

```bash
cd server && go test ./internal/game -run 'Test(FogOfWar|LightRadius)'
cd server && go test ./internal/realtime -run 'Test.*Fog|TestSnapshot|TestDelta'
make client-unit
make bot scenario=92_fog_of_war_radius
make bot-visual scenario=67_fog_of_war_overlay
```

Manual visual proof, if desired:

```bash
make bot-visual scenario=67_fog_of_war_overlay
```

## Scope limits

- No doorway, high-obstacle, non-rectangular, destructible, or explored-map occlusion shipped.
  Rectangular walls block living monster visibility, but the wall layouts themselves remain
  visible.
- No durable explored-map memory, minimap memory, or session-persistent map reveal shipped.
- No monster AI awareness, stealth/scouting unit, aggro, PvP, or combat balance changes shipped.
- No imported fog art, production dungeon lighting, audio, or particle/VFX pass shipped.
