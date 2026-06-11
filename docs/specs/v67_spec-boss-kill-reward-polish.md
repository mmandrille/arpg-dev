# v67 Spec: Boss Kill Reward Polish

## Status

Approved for implementation in the current autoloop.

## Context

Boss floors already unlock exits and drop rewards when the boss dies, but the protocol only exposes
the death as a generic `monster_killed` event plus later `interactable_state_changed` events. That is
enough for correctness, but weak for client status and future boss reward hooks.

## Goals

- Emit a server-authored `boss_killed` event whenever an entity with `is_boss` dies.
- Include the boss entity id, source/target ids, and `boss_template_id` in the event.
- Keep the existing `monster_killed`, loot, XP, and exit-unlock behavior unchanged.
- Show a short Godot status message when the local client observes `boss_killed`.
- Extend protocol/client bot proof for the existing Cave Warden boss-floor flow.

## Non-goals

- No new loot tables, item rewards, XP tuning, boss balance, or additional boss templates.
- No production victory UI, chest animation, audio, portrait art, or quest completion flow.
- No protocol version bump unless the existing schema validation requires it.

## Acceptance Criteria

- Killing Cave Warden emits both `monster_killed` and `boss_killed`.
- `boss_killed` validates through current session snapshot/state delta schemas and carries
  `boss_template_id: "cave_warden"`.
- The Godot client exposes a visible/debuggable status message for the boss kill.
- Existing boss floor unlock and descent proof still passes.
- `make ci` remains green.
