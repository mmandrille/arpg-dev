# v501 Spec — Monster Death Dissolve

- **Status:** Complete; focused visual/runtime proof and the combined batch `make ci` passed.
- **Date:** 2026-10-01
- **Base commit:** `5365832b9029e0d9e178d57f2a95a6a697b01acc`
- **Batch:** v501–v510 graphics, feel, and gameplay

## Purpose

Make an authoritative monster defeat easier to read by adding a brief, configurable material rim flash to the existing client-side death pose. This slice chooses the rim-flash option from the accepted brief; it does not add a dissolve or new particles. The existing death clip, darkened corpse, and v492 death burst remain in place.

## User-visible behavior

- A monster's `monster_killed` event starts a short, warm rim flash on its visible meshes. The flash eases back to each material's prior rim state while the existing death animation and corpse presentation continue.
- The flash is limited to monster death events. Player, companion, boss authority, loot, death timing, entity lifetime, collision, and replay behavior keep their existing semantics.
- Flash enablement, tint, peak strength, rise time, hold time, and release time live in a schema-backed client presentation catalog. Missing, unreadable, malformed, or invalid data disables this new effect; the existing death presentation continues unchanged.

## Non-goals

- No server simulation, protocol/schema, persistence, replay, combat balance, monster lifecycle, corpse despawn, or loot changes.
- No replacement death clips, imported art, added particles, audio, skill-VFX presets, or changes to player/companion death presentation.
- No external assets or plugins are needed. **Adopt:** the existing Godot `StandardMaterial3D` rim feature and client-only event path. **Reject:** external art/plugins and another particle effect for this slice; existing KayKit/Quaternius monster models and v492 death burst already provide the other visual layers.

## Acceptance criteria

1. The rim flash starts only in response to the existing authoritative `monster_killed` event for an entity typed as a monster; no new event or wire field is introduced.
2. On meshes with standard materials, the flash uses the configured emission tint and rim strength, rises/releases over the configured durations, and restores each material's prior rim, rim-tint, and emission values on completion or reset.
3. Missing/unreadable configuration safely disables the new effect without breaking the existing death pose, darken, death burst, or terminal reaction.
4. Player and companion death paths do not receive the monster-specific rim flash. Monster death animation and existing server-owned corpse/lifecycle behavior remain unchanged.
5. Shared JSON schema validation and focused Godot tests cover catalog loading, invalid-data fallback, trigger ownership, and material restoration. A client bot scenario confirms the live killed monster reaches terminal death on the authoritative event.
6. A real-windowed-renderer capture from the player's isometric camera shows the flash on a live killed monster. Record the exact capture and inspect it; headless checks alone do not satisfy this criterion.

## Likely surfaces

- Client: `client/scripts/model_reaction_controller.gd`, `client/scripts/gameplay_feedback_presentation.gd`, new monster-death presentation loader, and focused client tests.
- Shared presentation data: new `shared/assets/monster_death_presentation.v0.json` and schema.
- Bot: extend the existing `combat_input_flow_polish` client scenario with a death-reaction assertion and immediate windowed capture, or add an equally focused scenario if that route cannot capture the active effect. Keep it outside the merge-blocking CI pack unless the coordinator identifies a coverage gap that merits promotion.
- Docs: update the relevant `docs/CODEMAP.md` client/assets/test indexes; add the v501 plan and as-built report.

## Focused verification

- `godot --headless --path client --script res://tests/test_monster_death_presentation.gd`
- `godot --headless --path client --script res://tests/test_death_pose_ownership.gd`
- `make validate-shared`
- `make bot-client SCENARIO=combat_input_flow_polish HEADLESS=1`
- `AUTOPLAY_STEP_DELAY=0.05 make bot-visual scenario=combat_input_flow_polish` (capture must be windowed; inspect `v501_monster_death_rim.png`)
- `make maintainability`
- `git diff --check`

No performance improvement claim is made; this short-lived material treatment will be visually checked on the existing player camera. If implementation changes draw/renderer behavior beyond a transient material property, add a matched route comparison to the plan before claiming no frame cost.

## Dependencies and integration risks

- Depends only on the accepted v501 batch brief and existing v500 base; no sibling slice is a prerequisite.
- Keep death-specific values and logic out of the shared skill VFX catalog owned by v502. The proposed shared material/data surfaces do not overlap v503–v510's planned item, dungeon, class, rarity, boss-pattern, or passive surfaces.
- The event handler must not replay the flash for already-dead monsters restored from a snapshot; snapshot reconciliation continues to establish the existing terminal pose only.
- The coordinator owns combined `make ci`, `/finish`, integration, commits, review, and refactor after the batch is ready.

## Open questions

None. The brief explicitly allows a dissolve **and/or** rim flash; this spec selects a short material rim flash to preserve corpse presentation and avoid changing entity lifetime.
