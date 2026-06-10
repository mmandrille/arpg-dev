# v34 Plan — Universal Model Reactions and Co-op Player Tint

Status: Ready for implementation
Goal: Add client-only hit/death presentation reactions for local player, remote co-op players, and monsters while making remote players reuse the humanoid player visual path with a distinct tint.
Architecture: Keep ADR-0007 intact: authoritative server events drive client-only presentation, and animation state never crosses the wire. Add a small model reaction layer around existing `AnimationController` priority so terminal death can supersede hit/locomotion while transform/material tweens handle lean, blink, and downed corpse presentation. Remote co-op players should instantiate the existing character scene or equivalent humanoid path, get their own controller/reaction state, and remain fully server-authoritative.
Tech stack: Godot client/GDScript, existing shared protocol events, Godot client unit/smoke tests, client bot scenario, docs.

## Baseline and shortcut decision

Baseline is v33 `true-coop-session`: `main.gd` already distinguishes `local_player_id`, renders remote `type: "player"` entities under `entities_root`, and routes `player_damaged` / `player_killed` by entity id. Current drift to fix: remote players are primitive capsule placeholders and only monsters receive `AnimationController` setup in `_upsert_entity`.

Godot plugin adoption checklist outcome for v34:

| Candidate | Decision | Reason |
|-----------|----------|--------|
| Existing `AnimationController` + Godot Tween/AnimationPlayer | Borrow/reuse | Covers event-driven priority and simple transform/material reactions without a dependency. |
| Built-in `AnimationTree` / state machine | Reject for v34 | Existing terminal > one-shot > locomotion controller is sufficient. |
| LimboAI / external state-machine plugin | Reject | Too heavy for a small presentation-only reaction slice. |
| New model/asset pack | Reject | Slice explicitly reuses the current character/monster presentation. |

No shared JSON, Go sim, persistence, replay, or protocol schema work is expected. If implementation discovers a required server/protocol change, stop and revise the spec/plan before coding it.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `docs/specs/v34_spec-model-reaction-polish.md` | Keep client scenario numbering aligned with repo state |
| Create | `docs/plans/v34_2026-06-08-model-reaction-polish.md` | Implementation checklist |
| Create | `client/scripts/model_reaction_controller.gd` | Encapsulate hit/death transform and material reactions |
| Modify | `client/scripts/animation_controller.gd` | Expose/coordinate terminal debug state if needed |
| Modify | `client/scripts/main.gd` | Attach reaction controllers, route events, build remote player visuals and debug state |
| Modify | `client/tests/test_animation.gd` | Unit coverage for reaction priority and dead snapshot presentation |
| Modify | `client/tests/test_coop_client.gd` | Unit coverage for remote player model/controller/tint behavior |
| Modify | `client/scripts/bot_scenario_runner.gd` | Add client-bot assertions for entity presentation/reaction debug if needed |
| Modify | `client/scripts/bot_controller.gd` | Dispatch any new read-only bot assertions if needed |
| Create | `tools/bot/scenarios/client/12_model_reaction_polish.json` | Godot client proof |
| Modify | `PROGRESS.md` | Lifecycle update when v34 ships |

## Task 1 — Reaction Controller

Files:
- Create: `client/scripts/model_reaction_controller.gd`
- Modify: `client/tests/test_animation.gd`

- [x] Step 1.1: Add a small reaction controller that binds to a root `Node3D`, records base transform/material colors per mesh instance, and duplicates material overrides per instance before tinting.
- [x] Step 1.2: Implement `play_hit(source_position, fallback_direction)` as a short lean-away + dark blink that restores base transform/color when the entity survives.
- [x] Step 1.3: Implement `enter_death(source_position, fallback_direction)` as terminal state: kill active hit tween, rotate down onto the floor, apply persistent dark corpse tint, and ignore later hit calls.
- [x] Step 1.4: Expose `get_debug_state()` with at least `terminal`, `last_reaction`, `base_tint`, `current_tint`, and enough lean/death state for headless tests.
- [x] Step 1.5: Add unit tests proving hit restores color/transform, death overrides hit, and hit is ignored after death.

```bash
make client-unit
```

## Task 2 — Main Client Integration

Files:
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/animation_controller.gd`
- Modify: `client/tests/test_animation.gd`
- Modify: `client/tests/test_coop_client.gd`

- [x] Step 2.1: Preload the reaction controller and attach one to the local `character_visual` during `_ready`, preserving existing `PLAYER_TINT`.
- [x] Step 2.2: Replace `_make_remote_player_node` primitive capsule construction with instantiation of the existing humanoid/character visual path where available, and apply a readable dark charcoal/blue-black remote tint.
- [x] Step 2.3: Attach both `AnimationController` and reaction controller to remote player records in `_upsert_entity`; do not attach input prediction or local-player-only equipment behavior to remote players.
- [x] Step 2.4: Attach reaction controllers to monster records while keeping rarity tint as the base color.
- [x] Step 2.5: Route `monster_damaged`, `monster_killed`, `player_damaged`, and `player_killed` through a common helper that resolves `source_entity_id` and target positions before calling hit/death presentation.
- [x] Step 2.6: Apply terminal death presentation from snapshot/render paths when local player, remote player, or monster `hp <= 0`, not only from live events.
- [x] Step 2.7: Keep existing animation clips (`hit`, `death`) playing through `AnimationController` while transform/material presentation is layered by the reaction controller.
- [x] Step 2.8: Extend `get_bot_state()` / `_bot_entities_debug()` with presentation debug for local player, remote players, and monsters.
- [x] Step 2.9: Extend unit tests so remote player snapshots create a humanoid-like node with controller/reaction metadata and distinct tint, and remote player damage/death deltas update reaction debug state without touching local prediction.

```bash
make client-unit
```

## Task 3 — Client Bot Proof

Files:
- Create: `tools/bot/scenarios/client/12_model_reaction_polish.json`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/scripts/bot_controller.gd` if runner support requires controller plumbing
- Modify: `client/scripts/main.gd` if additional read-only bot state is needed

- [x] Step 3.1: Add client-bot assertion helpers for presentation debug, such as `wait_entity_reaction`, `assert_entity_terminal_reaction`, `assert_remote_player_visual`, and `assert_local_player_reaction`, using `get_bot_state()` rather than pixels.
- [x] Step 3.2: Create a focused solo client scenario that attacks a monster, waits for `monster_damaged`, asserts monster hit reaction debug, kills the monster, and asserts terminal downed/darkened death debug.
- [x] Step 3.3: In the same scenario or a second phase, trigger a player damage event through an existing lab/combat flow and assert local player hit reaction debug.
- [x] Step 3.4: If reliable with existing automation, include a co-op/session visibility step that proves remote player visual model/tint debug. If not reliable in v34, keep remote player visual proof in `test_coop_client.gd` and document the deferral in the scenario comments/plan completion notes.
- [x] Step 3.5: Verify the new client scenario directly.

```bash
HEADLESS=1 make bot-client scenario=12_model_reaction_polish
```

## Task 4 — Regression Checks

Files:
- Modify: client files touched by Tasks 1-3 as needed

- [x] Step 4.1: Run focused client unit coverage.

```bash
make client-unit
```

- [x] Step 4.2: Run client smoke to prove existing solo client flows and snapshot death behavior still work.

```bash
make client-smoke
```

- [x] Step 4.3: Run existing combat feedback client proof because it shares combat event/damage-number paths.

```bash
HEADLESS=1 make bot-client scenario=11_combat_feedback
```

- [x] Step 4.4: Run existing protocol bot only if implementation unexpectedly touches bot scenario runner shared behavior used by protocol scenarios.

```bash
make bot
```

## Task 5 — Lifecycle Docs and CI

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/plans/v34_2026-06-08-model-reaction-polish.md`

- [x] Step 5.1: Update this plan checkbox state as tasks complete.
- [x] Step 5.2: Update `PROGRESS.md` lifecycle table and v34 summary when implementation is complete.
- [x] Step 5.3: Record any actual deviation from the expected plugin shortcut decision.
- [x] Step 5.4: Run full CI.

```bash
make ci
```

## Final verification

- [x] `make client-unit`
- [x] `make client-smoke`
- [x] `HEADLESS=1 make bot-client scenario=12_model_reaction_polish`
- [x] `HEADLESS=1 make bot-client scenario=11_combat_feedback`
- [x] `make ci`

## Deferred scope

- No protocol/schema bump.
- No Go sim, replay, persistence, or shared rules changes.
- No production character customization or monster art replacement.
- No external Godot animation/state-machine plugin.
- No corpse collision, despawn, revive, respawn, or physics ragdoll.
