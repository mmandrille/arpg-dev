# v501 Plan — Monster Death Dissolve

- **Status:** Complete; integrated and combined batch `make ci` passed in 11m41s.
- **Date:** 2026-10-01
- **Base commit:** 5365832b9029e0d9e178d57f2a95a6a697b01acc
- **Goal:** Add a short configurable client-side material rim flash when the existing authoritative monster-death event enters the existing terminal reaction.
- **Prerequisites:** None. No sibling slice must integrate first. Execution was assigned by the batch coordinator after reviewing this plan.
- **Security router:** /meli-security-expert:meli-security-expert classifies pure UI/animation/presentation as out of scope. No network, user-controlled rendering, data access, or security-sensitive surface is planned.

## Baseline and design decisions

- The current client receives monster_killed, starts the monster's terminal death clip, applies the existing dark corpse tint, and spawns the existing v492 death burst. The server owns the death outcome and corpse/entity lifetime.
- Add one configurable material rim flash using built-in rim lighting plus a short configured emission tint. Do not dissolve/hide the corpse, add particles, or change its lifetime. The existing event/animation and death burst already cover those visual layers.
- **Adopt:** Godot's built-in StandardMaterial3D rim/emission properties, existing monster material copies, and the current monster_killed presentation path.
- **Borrow:** the existing client-only shared-presentation loader/schema pattern and the in-repo KayKit/Quaternius models and death clips.
- **Reject:** external art/plugins, new skill VFX, and a second particle treatment. No asset-manifest changes are expected.
- The event-time path may start the flash. Snapshot reconciliation of an already-dead monster continues to set the existing terminal pose without replaying a fresh flash.
- Keep schema-owned values limited to enabled, tint, peak_strength, rise_seconds, hold_seconds, and release_seconds. On missing/unreadable or invalid runtime data, the new effect is disabled; existing death presentation remains available.

## File map and ownership

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | shared/assets/monster_death_presentation.v0.json | Client-only enablement, tint, peak rim strength, and envelope durations. |
| Create | shared/assets/monster_death_presentation.v0.schema.json | Bounds and types for the death rim settings. |
| Create | client/scripts/monster_death_presentation_loader.gd | Lazy catalog load, validation/fallback, test reset, and immutable config accessor. |
| Modify | client/scripts/model_reaction_controller.gd | Add a monster death rim-flash helper; cache/restore existing rim material properties and stop/restore a flash on reset. Do not change hit, darken, death pose, or highlight semantics. |
| Modify | client/scripts/gameplay_feedback_presentation.gd | Invoke the helper only for a monster's event-driven death reaction; do not invoke for player/companion death or snapshot restoration. |
| Create | client/tests/test_monster_death_presentation.gd | Cover catalog loading and malformed config fallback, event trigger ownership, active values, release restoration, and early reset. |
| Modify | tools/bot/scenarios/client/103_combat_input_flow_polish.json | Capture immediately after the existing wait for an authoritative monster death reaction, before later movement steps. Use a unique v501 capture name and keep capture skippable in headless mode. |
| Modify | docs/CODEMAP.md | Index the new loader, catalog/schema, test, and capture scenario in the existing client/assets/bot rows. |
| Create | docs/as-built/v501_monster-death-dissolve.md | Record focused evidence, inspected real-camera capture, and remaining limits. |

No server, main.gd, protocol, VFX catalog, movement audit, CI pack, or asset manifest changes are planned. The capture extends an existing scenario without changing its movement/combat route, so no new movement-audit entry is needed. The scenario remains extended and is not promoted to tools/bot/ci_pack.json.

## Ordered tasks and checks

### 1. Add the isolated presentation catalog

Files: the new shared JSON/schema and client/scripts/monster_death_presentation_loader.gd.

- [x] Define strict, bounded schema values for the rim tint, strength, and rise/hold/release envelope.
- [x] Load once using the repository's existing client presentation path pattern. Missing, malformed, wrong-typed, or out-of-range config disables only this feature and emits a useful warning; no remote data or runtime fetch is introduced.
- [x] Expose a defensive config copy and a test reset/parser seam so invalid-data fallback can be exercised without rewriting shared files.

Verify: make validate-shared; include actual and malformed/default catalog cases in the focused GDScript test.

### 2. Implement material flash and event ownership

Files: client/scripts/model_reaction_controller.gd, client/scripts/gameplay_feedback_presentation.gd.

- [x] Add a death-rim helper that duplicates/uses the reaction's per-entity material copies, stores every affected material's baseline rim_enabled, rim, rim_tint, emission_enabled, emission color, and emission energy, and animates the configured peak/rise/hold/release.
- [x] Restore the exact baseline material properties after release or if terminal state resets early; kill any tween during reset/disposal. Preserve shared source materials and textures.
- [x] Call it only from the event-driven death path when the target record's type is monster. Do not change direct snapshot death handling or player/companion death.
- [x] Do not touch gameplay state, combat outcome, hit feedback, death animation, death burst, or corpse interactions.

Verify: godot --headless --path client --script res://tests/test_monster_death_presentation.gd and godot --headless --path client --script res://tests/test_death_pose_ownership.gd.

### 3. Add event and real-camera proof

Files: client/tests/test_monster_death_presentation.gd, tools/bot/scenarios/client/103_combat_input_flow_polish.json.

- [x] Test a killed monster's materials have the configured rim while the terminal death reaction is active, then return to their own baseline; test the event path does not trigger the flash for non-`monster_killed` events, players, or companions.
- [x] Test early reset restores the material state and cancels the transient effect.
- [x] Add a windowed capture_frame immediately after scenario 103 observes reaction: death; name it v501_monster_death_rim. Keep skip_if_headless: true.
- [x] Run the bot in headless mode for authoritative event/reaction flow, then run the same route through the real renderer and inspect the capture before deciding whether tint/strength/timing are visually legible.

Verify:

    make bot-client SCENARIO=combat_input_flow_polish HEADLESS=1
    AUTOPLAY_STEP_DELAY=0.05 make bot-visual scenario=combat_input_flow_polish

The bot-visual default `AUTOPLAY_STEP_DELAY=0.45` can consume nearly the full 0.53-second effect envelope after the death wait succeeds. Override only this run to 0.05 so the existing scenario-local capture samples the live isometric frame near the configured peak; do not change the runtime envelope or the global runner delay to accommodate capture timing. Record the actual PNG location, renderer, visible effect, and whether the capture caught the transient flash. If the scenario's follow-up step misses the window or corpse is removed too soon, adjust the scenario to capture directly after death observation or use a focused existing showme capture that retains the live isometric camera.

### 4. Maintainability and handoff evidence

Files: docs/CODEMAP.md, docs/as-built/v501_monster-death-dissolve.md.

- [x] Update CODEMAP's client, assets, bot/scenario, and test indexes for every new path.
- [x] Keep new source and test files within the ratchet; avoid expanding central bot or scene coordinator files.
- [x] Write as-built evidence with exact commands/results, the inspected capture, and proof limits. Do not claim server/replay changes or a performance improvement.

Verify:

    make validate-shared
    make maintainability
    git diff --check

The session handoff must also list the exact base, dependency state, changed/deleted/untracked paths, ignored evidence, shared-file overlaps, remaining acceptance gaps, and any required capture path. This batch worker does not run make ci/make ci-full, /finish, commit, push, or update lifecycle/PROGRESS.md; those belong to the coordinator after integration.

## Review and integration risks

- The direct trigger belongs in GameplayFeedbackPresentation.play_entity_reaction, which handles event-time reactions, rather than main.gd's general entity/snapshot reconciliation path.
- ModelReactionController also serves players/companions; the new helper must remain opt-in and the integration call must verify target type.
- Multiple MeshInstance3Ds may have different material baselines. Restoration must be per material, not a single cached strength/color.
- v502 owns shared skill-VFX presets, so this slice must add no entries to vfx_presentation.v0.json or modify combat_vfx.gd.
- The capture scenario is extended and independent of sibling gameplay/camera surfaces; no known prerequisite or shared-file conflict exists. Recheck actual batch integration state before any later execute assignment.
- A transient material change may affect a few visible corpses briefly. This plan makes no frame-time claim and does not require matched performance data unless the implementation expands beyond that bounded material effect.

## Execution gate

The coordinator approved the spec and plan and assigned execution. Focused client/shared/maintainability verification is complete; the coordinator owns sibling integration, combined batch CI, and closeout.
