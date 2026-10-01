# v496 attack-contact baseline — pre-v495 checkout

Date: 2026-09-30. Checkout: v493 current status; v494 and v495 remain To do. Local runtime: Godot 4.7.2 (the repo pins 4.6.3). This is working evidence for the v496 plan, not slice acceptance.

## Current sequence and measurements

| Point | Melee (`attack`) | Bow (`attack_ranged`) | Evidence / limit |
|-------|------------------|----------------------|------------------|
| Local start | `present_local_start` plays the cue and one-shot in the input dispatch, before `_send_action_intent` | Same in the equipped-bow path | Code path; no frame timestamp yet |
| Source clip duration | 1.067 s | 1.333 s | Godot 4.7.2 read of `KitHeroClips.library("kaykit_hero")`; these are unscaled clip lengths |
| Result | A lethal basic hit emits `monster_damaged`, then `monster_killed` | 104 bot observes hit and kill in live play | Server event contract / existing live bot; transport delay not measured |
| Target feedback | Damage event drives combat text; confirmed hit/crit drives hit audio, reaction and v492 particles; kill drives death feedback | Same | Code path and headless bot; headless renderer cannot prove particles or audio quality |
| Return to locomotion | `AnimationController` returns after the one-shot finishes | Same | Code path; real play-camera return frame not measured |

The original 103 client bot passed, but its player starts at `(2, 5)` and the monster at `(13, 5)`. Its immediate floor retarget happened during attack-move, before a melee swing. The amended scenario moves the player near the monster and asserts the local `attack` clip. Before the code correction, the live bot showed **two** melee-lunge starts after one local click and a lethal result. The second start was caused by `monster_killed` replaying the swing after `monster_damaged` had consumed the local prediction. After the correction, the same scenario passes with `count_min=1` and `count_max=1`, then dispatches floor movement.

The original 104 bow client bot passed before the correction. It confirms an equipped bow, authoritative damage reaction and kill, but does not expose clip contact frames, result latency, miss/block, or quality-tier particle visibility. Two later runs timed out at `wait_event item_picked_up` even though the pickup event was already in the pending list: the event arrived before that step's start boundary. Replacing that wait with `wait_inventory_item` made the bow hit/kill regression pass again. A later renderer recording exposed the same race at `wait_event item_equipped`, so the scenario now waits for equipped state; the revised bow hit/kill bot passes. The first windowed `--write-movie` file had no readable duration; the second captured only 120 frames / 2 s before equip, requiring 71 s wall time (2% real-time speed). Neither is combat visual or real-time timing proof. Keep normal-speed play-camera before/after capture, Balanced/Performance checks, and exact input/contact/result/effect/locomotion timestamps open for the v495-based timing pass.

The combat-stat blocking target exposed a separate contact classification defect. In an initial live run of focused scenario 109, the server reported `outcome=block` and the client displayed `BLOCK`, but the target's hit-feedback counter reached 2 after retries that included an ordinary hit. The client had treated every `monster_damaged` event as hit contact, including a blocked one. The correction preserves combat text and gates hit audio, reaction, and its VFX on a confirmed hit. Scenario 109 now records the feedback-counter delta for the matched damage event; its block step passed with zero added feedback. Earlier successful hits can legitimately raise the target's lifetime counter, so that lifetime count alone cannot classify the block. The rapid-input extension is written but awaits a headless run after the v495 timing comparison releases the host.

## Pending measurement after v495

- Capture one valid melee and bow attempt plus a miss or block, with timestamps or frame indices for input, local clip start, visible contact, authoritative result, target effect/audio, and locomotion return.
- Repeat matched captures on the v495 baseline before selecting data-owned clip/contact offsets or blend durations. Preserve an already-aligned family as a control.
- Verify both graphics quality tiers in the real renderer. Headless client bot success establishes event/order behavior only.
