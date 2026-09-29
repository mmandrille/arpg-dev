# v485 Spec — Remote player class on player entity views

Date: 2026-09-29
Status: Complete
Codename: `remote-player-class`
Found during: ADR-0018 P3c investigation (v484 `retire-legacy-hero`, in flight on its own branch).

## Problem

Every co-op partner renders as the class fallback model instead of their own KayKit hero.

- `EntityView.CharacterClass` (`server/internal/game/types.go`) exists with `omitempty`, but the
  server only fills it for mercenary companions (`companion_ai.go`). `(e *entity) view()` never sets
  it for `type:"player"`.
- The client's `_make_remote_player_node` resolves `ClassPresentationsLoader.resolve("")`, which is
  the fallback asset.
- **The brief's premise was wrong.** `character_class` is **not** a property of the v8
  `session_snapshot` / `state_delta` `entity` schema def. That def is `additionalProperties: false`,
  so the companion payloads that already carry the field are off-contract today. Nothing validates
  live traffic against the schemas; only `shared/protocol/examples/` are checked, which is why the
  drift went unnoticed.

## Goals

1. Every player entity view (snapshot, spawn, update) carries the owning member's
   `character_class`.
2. The value comes from live progression, not from a copy cached on the entity: the active
   member's is `s.progression`, every other member's is `playerState.Progression`, and the entity id
   is the `PlayerID`.
3. Add `character_class` (optional, `minLength: 1`) to both v8 entity defs in place. This follows
   the additive-optional precedent (v208 `companion_stance`, v448 `steward_hunt_target`); old clients
   ignore it. Put it in the protocol examples.
4. Client: remote players are built with their class model. A later snapshot or delta carrying a
   different class for an existing record swaps the model in place and rebinds the animation and
   reaction controllers.
5. A two-peer run proves a real protocol host and a real Godot guest of different classes each
   render the other as their own class.

## Non-goals

- **An in-session class change.** None exists: class is chosen at character creation and the
  debug progression PUT has no class field. The Go "next delta" test pins the property that such a
  change would rely on (views derived live). It does not exercise a gameplay path.
- Remote-player gear, hero corpse class models, and the legacy-hero retirement (v484).
- A protocol version bump. The change is additive and optional.

## Design

- `server/internal/game/entity_view.go` (new, pure move out of `sim.go`: `Sim.entityView`,
  `entity.view`, `entity.bossPhaseView`) gets `Sim.playerEntityClass(e)`, called from
  `entityView` for player entities. This is view-only: no sim state, input or replay change, and a
  map lookup by key rather than a range.
- `client/scripts/remote_player_class_sync.gd` (new): `class_changed`, `swap_model` (calls main's
  `_apply_character_class_model`, then rebuilds `AnimationController` and `ModelReactionController`,
  as the local-player swap does), and the bot-state helpers `remote_player_ids` / `rendered_classes`.
- `client/scripts/bot_remote_player_assertions.gd` (new): the `wait/assert_remote_player_count`
  matcher moves out of `bot_scenario_runner.gd` and gains the optional `rendered_classes` and
  `local_rendered_class` expectations.
- Scenario `client/21_join_game_listed_session`: the preflight host is pinned to `rogue` (new
  `preflight.character_class` → `client_join_preflight.py --character-class`), and the guest picks
  `sorcerer`. The final wait requires the remote player to render `rogue` and the local hero
  `sorcerer`. Neither is the server default class.

## Tests

- Go `entity_view_test.go`: both viewers' snapshots show each member's class; a guest move delta
  reflects live progression; non-player entities omit the field. The class IDs are derived from the
  loaded rules.
- GDScript `test_remote_player_class.gd` (registered in `client_smoke.sh`): the snapshot builds the
  class model; a delta without a class keeps the controller; a delta with a new class swaps the
  model in place (one `ModelRoot`, new controller); plus matcher unit cases. The classes are derived
  from the presentation catalog, excluding the fallback asset.
- Scenario 21 A/B: it fails with the server hook disabled and passes with it.
