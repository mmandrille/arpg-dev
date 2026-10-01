# v496 — Attack contact timing

- **Status:** Complete (combined `make ci` passed)
- **Date:** 2026-09-30
- **Codename:** `attack-contact-timing`
- **Area:** Graphics and gameplay feel
- **Depends on:** v495 performance baseline; followed by the scheduled engineering review and focused refactor before v497.
- **ADRs:** [ADR-0001](../adr/0001-technology-stack.md) D2–D3; [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md) P3/P5.

## Purpose

Make a basic attack read as one coherent action: immediate local swing, visible weapon/contact motion, and authoritative hit/miss/block confirmation at the target. v464 input flow, v465 outcome punch, v475 KayKit clips, and v492 particles exist; this slice checks and aligns their timing in live play. Begin with one melee and one ranged weapon family, then apply the same data-owned timing mechanism to the other basic-attack clips only when the capture supports it.

## Non-goals

- Damage, attack speed, cooldown, or class balance changes.
- New skills, production audio packs, elaborate per-class animation sets, or new protocol event types.
- Predicting damage or a hit outcome before server confirmation.
- Death dissolve and target outline, which need their own visual scope if still desired.

## Acceptance criteria

- [ ] A live-client baseline records input-to-swing, clip contact, authoritative result arrival, impact VFX/audio, and return to locomotion for representative melee and ranged attacks, including a miss or block.
- [ ] The swing starts locally without waiting for the 10 Hz server tick. The hit/crit/miss/block display still follows the authoritative event and never implies a successful hit for a miss or block.
- [ ] Contact motion, existing attack cue, target-side reaction, and v492 particles form a legible sequence at normal play speed. Attack animation does not restart or double-play when the result arrives after a locally predicted start.
- [ ] A floor retarget during recovery clears the pending combat command and blends into movement without a stale swing or camera jump; held/rapid attack input remains bounded by existing buffer rules.
- [ ] Clip/contact offsets, blends, and effect timing are data-owned in `shared/assets/` presentation catalogs where tunable; tests do not pin arbitrary current values.
- [ ] Real play-camera before/after clips show the melee and ranged cases, while a live-client bot asserts action/result/retarget order.

## Scope and likely files

- **Client/data:** `client/scripts/combat_local_attack_presentation.gd`, `client/scripts/animation_controller.gd`, `client/scripts/attack_animation_scaling.gd`, `client/scripts/client_audio_bridge.gd`, `client/scripts/gameplay_feedback_presentation.gd` or their current owner paths; `shared/assets/attack_presentation.v0.json`, `shared/assets/combat_feel_presentation.v0.json`, and schemas as needed.
- **Bot/tests:** extend an existing live-client attack-flow scenario rather than duplicating v464/v465 proof; add timing/duplicate-start unit checks.
- **Contracts:** no gameplay-rule or protocol change expected. If an existing result event lacks the information required for honest presentation, resolve that in the plan before expanding scope.

## Test and bot proof

- `make validate-shared`, focused combat/animation/audio GDScript tests, `make client-unit`, and `make bot-visual scenario=103_combat_input_flow_polish` plus `104_combat_impact_confirmation` as regressions.
- A live-client scenario must cover one melee and one ranged attack with outcome classification; capture normal-speed video or an equivalent frame sequence from the real renderer for visual review.
- Check the same sequence at Balanced and Performance quality so VFX scaling does not erase the outcome cue.

## Asset/plugin decision

- **Borrow** the in-repo KayKit attack clips, existing cue generator, and v492 VFX catalog. Check existing scenes and clips before editing offsets.
- **Reject** new asset packs and plugins for this timing slice.

## Open questions and risks

- The live baseline may show that one weapon family already feels aligned. Change only the reproducible mismatch; record the untouched family as a control.
- Authoritative event latency varies with transport. Presentation may smooth a late result, but may not postpone or fabricate combat outcomes.
