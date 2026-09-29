# v485 As-Built — Remote player class on player entity views

Date: 2026-09-29
Status: Complete (`make ci` green, 6m09s)
Spec: [`v485_spec-remote-player-class.md`](../specs/v485_spec-remote-player-class.md). There is
no separate plan file (small cross-stack fix).

## What shipped

**Server.** Player entity views now carry `character_class`.
- `Sim.entityView` calls `playerEntityClass`, which reads the active member's class from
  `s.progression` and every other member's from `playerState.Progression`. Entity id == PlayerID.
- Nothing is cached on the entity, so the value cannot go stale.
- The view code moved from `sim.go` to `entity_view.go` in a pure-move commit first. The `sim.go`
  baseline went from 6731 to 6621.

**Protocol.** Optional `character_class` (`minLength: 1`) was added in place to the
`session_snapshot.v8` and `state_delta.v8` entity defs, with the examples updated.
- The brief said the property already existed. It did not: the entity def is
  `additionalProperties: false`.
- So the mercenary-companion payloads that already sent the field were off-contract until this
  slice.

**Client.**
- Remote players were already built via `_apply_character_class_model(root, class_id)`. They now
  get a real class ID.
- `_upsert_entity` swaps an existing remote player's model in place when a later payload carries a
  different class (`remote_player_class_sync.gd`), then rebinds the animation and reaction
  controllers.
- The duplicated corpse-key copy loop was folded into the main key loop. `character_class` is now
  recorded on entity records.
- Remote-player bot introspection moved out of `main.gd`. The `main.gd` baseline went from 6765 to
  6758.

**Bot.**
- `wait/assert_remote_player_count` now lives in `bot_remote_player_assertions.gd`, with optional
  `rendered_classes` and `local_rendered_class` expectations.
- Scenario `client/21_join_game_listed_session` pins a `rogue` protocol host and a `sorcerer` Godot
  guest, and asserts both render as their classes.

## What it proved

- Go: `TestPlayerEntityViewsCarryEachMembersClass` (both viewers), `TestPlayerEntityClassTracksLiveProgressionInNextDelta`, and `TestNonPlayerEntityViewsOmitCharacterClass`. Both class tests fail with the hook disabled.
- GDScript: `test_remote_player_class.gd` passes 22 assertions and fails 3 with the swap hook disabled. `test_coop_client.gd` is unchanged (279 assertions).
- The two-peer run `make bot-client SCENARIO=21_join_game_listed_session HEADLESS=1` passes in 4.2s. With the server hook disabled it times out at the `rendered_classes` wait.
- `make validate-shared`, `make lint-determinism` and `make maintainability` are green.

## Known limits / follow-ups

- **There is no in-session class change path.** Class is fixed at character creation, and the
  debug progression PUT has no class field. The client swap and the Go "next delta" test are
  defensive: they cover resync correctness and any future respec or class select, not a live
  gameplay path.
- The two-peer proof is headless and checks the applied `class_id`. It took no screenshot.
- `state_delta.v8` `entity` still lacks `combat_stats`, which `session_snapshot.v8` has and
  companion updates send. This is pre-existing schema drift, not fixed here.
- There is still no runtime schema validation of live payloads. That is how the
  companion `character_class` drift survived.
- v484 (P3c) edits `main.gd`, `client_smoke.sh`, and the `main.gd` baseline row. Whichever branch
  merges second reconciles the baseline line.
