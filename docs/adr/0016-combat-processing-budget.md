# ADR-0016: Combat Processing Budget (Proposed)

Status: Proposed  
Date: 2026-07-08

## Context

Crowded combat (solo D-3 packs, future 6-player boss rooms) produces high per-tick event, persist, and fanout volume. Skill-cast disconnects traced to tick budget overruns dominated by persist and send-queue pressure (v453).

## Decision

1. **Authority vs presentation split for projectiles**
   - Server may resolve multi-hit and instant-ray skills without simulating player skill projectile entities.
   - Client flight VFX remain driven by `skill_cast` fields (`position`, `direction`, `range`, `projectile_def_id`) via `ProjectileFlightPresentation.spawn_from_skill_cast`.
   - Hit outcomes are authoritative on the server; trajectories are client presentation.

2. **`skill_damage_burst` events**
   - Multi-hit skills (e.g. Volley) emit one `skill_damage_burst` per cast with `hits[]` instead of O(targets) `monster_damaged` events on the wire.
   - Bot ingest expands bursts into synthetic `monster_damaged` for assertion compatibility during migration.

3. **Skill `resolution` enum in `skills.v0.json`**
   - `simulated_projectile` — spawn projectile entity, resolve on impact (default legacy behavior).
   - `instant_ray` — resolve first line-of-sight hit on cast tick; emit `skill_cast` flight fields; no projectile entity.
   - `instant_aoe` — resolve fan/line multi-hit on cast tick (ranger volley/pierce family).

4. **Per-tick skill budget** (v455) caps resolutions and damage-event fanout; overflow defers to tick+1 FIFO ordered by `(player_id, sequence)` without silent cast drops.

## Consequences

- Replay determinism requires burst aggregation and budget deferral to preserve same seed + input ordering.
- Client must handle `skill_damage_burst` for combat text/VFX while preserving `skill_cast` projectile presentation.
- Protocol stays v8 with additive optional `hits` on burst events.

## Non-goals

- Monster projectile migration.
- Async persist worker.
- Removing all per-hit `monster_damaged` from every skill in one slice.
