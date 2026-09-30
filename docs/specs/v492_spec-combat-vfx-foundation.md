# v492 — Combat VFX foundation: particles + in-repo glow shader, hit sparks, death burst, heal rain (ADR-0018 P5)

- **Status:** Implemented (v492)
- **Date:** 2026-09-30
- **Codename:** `combat-vfx-foundation`
- **ADR:** [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md) D3 (in-repo `.gdshader`, GPU/CPU particles allowed), P5
- **Baseline:** v491 (`5e70b02e`)
- **Spec gate:** client presentation + presentation data only; exempt, written anyway.

## Purpose

The client has no particle systems and no shaders. Combat reads through a hit flash and a death
darken only, and the heal skill's "rain" is 64 box-mesh plus-signs animated by hand. After this
slice:

1. **Foundation.** An in-repo additive soft-glow billboard shader (`client/shaders/vfx_soft_glow.gdshader`),
   and a `CombatVfx` module that builds one-shot `GPUParticles3D` bursts from a schema-backed catalog
   (`shared/assets/vfx_presentation.v0.json`: amount, lifetime, velocities, spread, gravity, sizes,
   start/end colours, glow energy, spawn height). Bursts free themselves when done.
2. **Hit sparks** on every entity hit reaction (`monster_damaged`, and player hits through the same
   `GameplayFeedbackPresentation.play_entity_reaction` path), sprayed away from the attacker, coloured
   by the event's `damage_type` (`force`, `fire`, `cold`, `poison`, `lightning`) from the catalog.
3. **Death burst** (dust + rising wisps) on death reactions.
4. **Heal rain** rebuilt on particles (falling glow motes over the heal radius + a fading ground
   ring), keeping `HealRainEffect`'s class, `setup(radius)` API and lifetime.
5. **Quality scaling.** `render_presentation.v0.json` quality tiers gain `particle_scale`; the
   `performance` tier halves particle counts. `SceneLightingRig.sync` hands the active tier to `CombatVfx`.

## Non-goals

- Death dissolve (needs the monster node kept alive past `entity_remove`; changes removal flow).
- Per-skill projectile/impact VFX, outline/rim highlight, screen effects, audio.
- Any gameplay, protocol or server change.

## Acceptance criteria

- [ ] `CombatVfx.make_burst(id, …)` builds a one-shot `GPUParticles3D` whose amount, lifetime, gravity
  and colour ramp come from the catalog (test derives from the catalog), with a `ShaderMaterial` using
  the glow shader, and frees itself.
- [ ] A hit reaction spawns a `hit_spark` burst at the target, coloured by `damage_type` (unknown type →
  the effect's default colours), directed away from the source; a death spawns `death_burst`.
- [ ] Bursts attach to the target's parent (they outlive the entity node on death).
- [ ] `performance` quality scales burst amount by its `particle_scale` (≥ 1 particle).
- [ ] `HealRainEffect` has particles + ring and still frees after its lifetime; existing heal-rain
  tests (co-op, status-effect) stay green.
- [ ] `make validate-shared` accepts the new catalog and schema changes.
- [ ] Visual gate: captures of hit sparks per damage type, the death burst and heal rain.
- [ ] `make client-unit`, combat client scenarios, `make ci` green.

## Files

new `client/shaders/vfx_soft_glow.gdshader`; new `shared/assets/vfx_presentation.v0.json` + schema;
`shared/assets/render_presentation.v0.json` + schema (`particle_scale`); new `client/scripts/combat_vfx.gd`;
`client/scripts/gameplay_feedback_presentation.gd`; `client/scripts/scene_lighting_rig.gd`;
`client/scripts/heal_rain_effect.gd`; new `client/tests/test_combat_vfx.gd`; docs. `main.gd` untouched.

**Asset decision:** in-repo shader and Godot built-in particles only (ADR-0018 D3); no textures, no plugins.
