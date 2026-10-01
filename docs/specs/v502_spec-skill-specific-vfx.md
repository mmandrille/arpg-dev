# v502 — Skill-specific impact VFX

- **Status:** Complete; focused checks, the final Godot client gates, and combined batch `make ci` passed.
- **Date:** 2026-10-01
- **Codename:** `skill-specific-vfx`
- **Batch base:** `5365832b9029e0d9e178d57f2a95a6a697b01acc`
- **Dependencies:** v492 combat VFX foundation is present at the batch base; no other slice is a prerequisite.

## Purpose

Give three frequently encountered Sorcerer projectile skills a distinct, readable impact burst when they damage a monster: Magic Bolt, Ice Shard, and Lightning. The client currently selects the same generic `hit_spark` burst for every successful hit, even though authoritative combat events already include the originating `skill_id`.

The slice adds small, one-shot, schema-backed particle presets for these skills and selects them from the existing client hit-reaction path. Keep the established low-poly presentation and in-repo soft-glow shader. The effects communicate which spell landed while preserving the existing generic feedback for basic attacks and every unmapped skill.

## Scope

- `magic_bolt` gets a compact blue arcane impact burst.
- `ice_shard` gets a cool cyan, wider, lower-energy shatter-like burst.
- `lightning` gets a short-lived, bright yellow-white burst with a wider spray.
- The presets own their values in `shared/assets/vfx_presentation.v0.json`; the existing catalog schema must validate their shape and the skill-to-effect mapping.
- The client uses the server-provided `skill_id` from the existing `monster_damaged` event to choose a preset. No event or protocol field is added.
- In the `skill_visual_lab` fixture, keep the soft target within cast range while placing it outside the one-hit XP dummies' Lightning chain radius, so unrelated level-up and death effects do not obscure the impact capture.

## Non-goals

- No server simulation, damage, replay, protocol, or gameplay rule changes.
- No skill-cast, projectile-flight, status, player-hit, or basic-attack presentation changes.
- No monster death burst, death timing, entity removal, or dissolve changes (owned by v501).
- No production monster placement or XP-reward tuning; the only world-data adjustment is isolated to the visual-test lab fixture.
- No generalized effect editor, new VFX framework, new external art, plugin, or performance claim.
- No changes to skill icons or their current rank-based cast ring.

## Acceptance criteria

1. For a successful `monster_damaged` event, `CombatVfx` selects the mapped catalog preset for exactly `magic_bolt`, `ice_shard`, and `lightning`; each burst is still a one-shot, self-freeing `GPUParticles3D` attached to the entity's parent.
2. The three presets have distinct catalog-owned profiles and tests verify their selected IDs and profile values against the catalog rather than copying tuning constants into tests.
3. A basic attack or any unmapped skill continues to use `hit_spark`; its existing damage-type color behavior remains unchanged. Death continues to use `death_burst` regardless of `skill_id`.
4. Shared asset validation and the focused Godot VFX/client checks pass. Existing combat event handling remains functional.
5. Capture each selected skill in the actual renderer through the reusable `skill_visual` scenario and inspect the images. The scenario's existing `visual` replay metadata gates one deterministic `capture_frame` per allowlisted `monster_damaged` event, applies its configured isometric focus/zoom only in visual replay, skips headless runs, and uses the existing viewport capture helper. The ranged showcase approaches its configured soft target before casting; the lab fixture also keeps its XP dummies outside Lightning's chain radius so progression/death feedback cannot cover the capture. Captures are named `.artifacts/bot-captures/v502_skill_<skill_id>.png`; commands and paths are recorded in the as-built note. Captures do not stand in for performance measurements.

## Likely surfaces

- Shared presentation data/schema: `shared/assets/vfx_presentation.v0.json`, `shared/assets/vfx_presentation.v0.schema.json`.
- Visual-test fixture geometry: `shared/rules/worlds.v0.json` (`skill_visual_lab` only).
- Client VFX selection and construction: `client/scripts/combat_vfx.gd`.
- Focused coverage: `client/tests/test_combat_vfx.gd`, `tools/bot/test_skill_visual.py`.
- Documentation: this spec, its plan, and `docs/as-built/v502_skill-specific-vfx.md`; update the existing `docs/CODEMAP.md` visuals row to index the touched VFX builder, catalog, schema, and test.
- Existing event contract evidence: `server/internal/game/training_doll.go` assigns `SkillID` to skill damage events via `appendMonsterCombatEvent`; this slice consumes the field without changing its producer.

## Asset decision

**Adopt** the existing v492 `GPUParticles3D` builder and `vfx_soft_glow.gdshader`. **Reject** external assets, plugins, and a new asset pipeline: the three effects can use the current in-repo presentation path. No new KayKit model or texture is needed; retain the project's existing stylized, readable palette.

## Dependencies and integration risks

The coordinator transferred v501's checksum-verified prerequisite paths into this worktree; HEAD remains the recorded batch base. The current v501 change set adds the monster death presentation loader/rim flash and updates its reaction controller, CODEMAP row, and visual scenario. It does not touch `client/scripts/combat_vfx.gd`, `client/tests/test_combat_vfx.gd`, or the vfx catalog/schema, so there is no current same-path conflict. Both features pass through the existing reaction dispatcher; v502 must leave v501's death selection, rim flash, and removal flow untouched and retain a death regression check.

## Focused verification

- `make validate-shared`
- `make client-unit`
- `make bot-client SCENARIO=client_skill_points_and_magic_bolt HEADLESS=1` (combat-path regression where applicable)
- Real-renderer captures for each selected skill:
  - `ARPG_SKILL_VISUAL_SKILL_ID=magic_bolt make bot-visual scenario=skill_visual`
  - `ARPG_SKILL_VISUAL_SKILL_ID=ice_shard make bot-visual scenario=skill_visual`
  - `ARPG_SKILL_VISUAL_SKILL_ID=lightning make bot-visual scenario=skill_visual`

The screenshot set proves visible presentation on those replay paths only. It does not prove GPU cost or frame-rate impact. The coordinator owns the combined `make ci` gate after integration.

## Open questions

None. The accepted brief explicitly calls for a bounded skill subset; these three existing Sorcerer projectile skills provide a clear force/cold/lightning comparison and already produce skill-tagged monster damage events.
