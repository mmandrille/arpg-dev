# Spec: `model-reaction-polish`

Status: Draft
Branch: `main`
Slice: v34 - universal model reactions and co-op player tint
Baseline: v33 `true-coop-session`
Related:

- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - thin client, authoritative server, client-only presentation
- [`v3_spec-animate-and-react.md`](v3_spec-animate-and-react.md) - original client animation controller and monster reactions
- [`v4_spec-take-a-hit.md`](v4_spec-take-a-hit.md) - player damage/death event reactions
- [`v31_spec-combat-stat-effects-and-feedback.md`](v31_spec-combat-stat-effects-and-feedback.md) - combat outcome feedback metadata
- [`v33_spec-true-coop-session.md`](v33_spec-true-coop-session.md) - local and remote player entity handling

## 1. Purpose

Make every visible character-like model react consistently when damaged or killed. The server
already emits authoritative combat events for player and monster damage/death; this slice makes the
Godot client presentation clearer and more consistent without changing gameplay authority.

After v34:

- Monsters, the local player, and remote co-op players all get a visible hit reaction when they
  receive damage.
- A hit reaction briefly leans the target away from the attacker when the attacker can be resolved,
  flashes the model darker, then restores the target's normal tint.
- Monsters, the local player, and remote co-op players all get a visible death reaction when killed.
- A death reaction rotates the target down onto the floor and darkens the model as a persistent
  corpse presentation.
- Remote co-op players reuse the same humanoid/player visual path as the local player, with a
  distinct readable dark tint.
- The existing event-driven animation priority remains intact: terminal death beats hit and
  locomotion.

The proof is: authoritative combat event -> entity-specific client reaction -> terminal death state
persists across live deltas and snapshot-driven render paths -> remote co-op player presentation is
visually distinct but model-consistent.

## 2. Non-goals

- No server gameplay changes.
- No protocol schema/version bump.
- No new combat events.
- No ragdoll physics, corpse collision changes, corpse despawn, revive, respawn, or checkpoint
  behavior.
- No production character customization, cosmetics, dye system, or per-account visual loadout.
- No monster art replacement. Monsters may keep the current dummy/monster presentation, but must
  share the same reaction behavior.
- No requirement to author new skeletal clips if a small Godot-side transform/material reaction can
  satisfy the slice.
- No final art direction pass for player/monster models.
  worth adopting.

## 3. Files to create or modify

```text
docs/specs/v34_spec-model-reaction-polish.md          - this slice contract
docs/plans/v34_<YYYY-MM-DD>-model-reaction-polish.md  - implementation plan
PROGRESS.md                                      - lifecycle update when v34 ships

client/scripts/main.gd                                - event routing, remote player model creation, tint application
client/scripts/animation_controller.gd                - reaction/death presentation state if it belongs with animation priority
client/tests/test_animation.gd                        - reaction priority and terminal death behavior
client/tests/test_coop_client.gd                      - remote player model/tint behavior
client/scripts/smoke.gd                               - client-side event reaction proof if needed
tools/bot/scenarios/client/12_model_reaction_polish.json - Godot client proof if reliable
```

The exact client decomposition may differ in the plan. If reaction behavior is large enough to
justify a helper, prefer a small in-repo script such as `model_reaction_controller.gd` over mixing
all tween/material state into `main.gd`.

## 4. Presentation behavior

### 4.1 Hit reaction

When a visible entity receives a damage event:

- `monster_damaged` targets the monster entity.
- `player_damaged` targets the damaged player entity, which may be local or remote.

The target should:

1. Lean quickly away from the attacking source entity when `source_entity_id` can be resolved.
2. Use a deterministic fallback lean direction when the source cannot be resolved.
3. Darken all model materials briefly.
4. Restore the target's base tint/material color after the blink.
5. Return to the previous locomotion/idle animation if the target is still alive.

The reaction is presentation-only. It must not alter authoritative position, HP, collision, or
gameplay state.

### 4.2 Death reaction

When a visible entity is killed or is rendered from a snapshot with `hp <= 0`, the target should:

1. Enter terminal animation state.
2. Rotate down onto the floor, using a simple readable direction that does not need physics.
3. Darken its model as a persistent corpse.
4. Ignore later hit and locomotion reactions.

Death presentation must remain stable if a later snapshot or replay render path creates an already
dead entity without seeing the original kill event.

### 4.3 Tint and material ownership

Each entity needs a restorable base tint:

- Local player: keep the current player tint.
- Remote co-op player: use the same player/humanoid model path, tinted readable dark charcoal or
  blue-black rather than pure black so model shape remains visible.
- Monster: keep rarity tinting where present.

Hit blink and death darkening must be layered over the base tint instead of permanently losing
rarity/player/remote-player color metadata.

### 4.4 Remote co-op player model

Remote co-op players should no longer use a simple capsule placeholder as their normal
presentation. They should reuse the same humanoid/player model path as the local player where the
current client structure allows it.

Required behavior:

- A remote player is visually recognizable as the same class of model as the local player.
- A remote player has an `AnimationController` or equivalent reaction controller.
- A remote player is tinted differently from the local player.
- A remote player remains server-authoritative: no local prediction, no local input routing, no
  independent gameplay state.

## 5. Event mapping

The slice must preserve the existing client-only animation boundary. Events are read from
`state_delta.events` and mapped by entity id:

| Event | Target | Reaction |
|-------|--------|----------|
| `monster_damaged` | damaged monster | hit reaction |
| `monster_killed` | killed monster | death reaction |
| `player_damaged` | damaged local or remote player | hit reaction |
| `player_killed` | killed local or remote player | death reaction |

If both damage and kill events appear for the same entity in one delta, terminal death wins. If a
death event arrives while a hit tween is active, the hit tween must be stopped or superseded so the
corpse remains down and dark.
## 7. Acceptance criteria

1. Local player, remote co-op players, and monsters all show a visible hit reaction from their
   corresponding authoritative damage events.
2. Hit reaction leans away from the resolved attacker when `source_entity_id` is available.
3. Hit reaction dark-blinks the model and restores the entity's normal base tint if the entity
   survives.
4. Local player, remote co-op players, and monsters all show terminal death presentation from kill
   events.
5. Already-dead entities created from snapshots enter the terminal downed/darkened presentation
   without needing to replay the original kill event.
6. Death presentation overrides active or later hit/locomotion reactions.
7. Remote co-op players reuse the same humanoid/player visual path as the local player and use a
   readable dark alternate tint.
8. The change requires no protocol schema bump and no server gameplay changes.
9. `make client-unit` passes.
10. `make client-smoke` passes.
11. A focused Godot client bot proof passes, or the plan documents why existing client smoke covers
    the same observable behavior reliably.

## 8. Open questions

Resolved defaults for v34:

| # | Decision |
|---|----------|
| Q-1 | Remote co-op players use readable dark charcoal/blue-black rather than pure black. |
| Q-2 | Monsters may keep the current dummy/monster presentation, but share the same reaction behavior. |
| Q-3 | Dead bodies remain visible as downed, darkened corpses. |

## 9. Implementation notes for planning

- Prefer one reaction path that works for local player, remote player, and monster nodes.
- Keep client debug state rich enough for headless tests to assert reaction/terminal state without
  pixel inspection.
- Be careful with material overrides: duplicated materials should not accidentally tint every
  instance that shares an imported mesh resource.
- Snapshot render paths must apply terminal death from `hp <= 0`, not only from live events.
- This is client presentation work; Go sim determinism should be unaffected.
