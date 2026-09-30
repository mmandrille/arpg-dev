# v486 Spec: Live Payload Schema Gate

Status: Complete
Date: 2026-09-29
Codename: `live-payload-schema-gate`
Baseline: v483 `armor-look` (commit `5e891e5e`). v484 (`retire-legacy-hero`) and v485
(`remote-player-class`) are in flight in other worktrees.

## Problem

`tools/validate_shared.py` validates only the hand-written `shared/protocol/examples/*.json`
against the v8 `session_snapshot` / `state_delta` schemas. Nothing validates what the server
actually sends, so off-contract payloads ship silently. v485 found one by hand: mercenary and
player entities emit `character_class`, but the v8 `entity` def is `additionalProperties: false`.
It also noted that `state_delta` `$defs.entity` lacks `combat_stats`, which companion entity
updates send.

A report-mode survey over every protocol scenario (`make bot scenario=all`, 135 scenarios) found
~30 distinct drifts. There were far more than the two known ones, and they sort into four classes:

| Class | Examples | Owner |
|-------|----------|-------|
| Undocumented optional props | `entity.character_class`, `entity.combat_stats` (delta), `character_progression.character_class`, `derived_stats.block_percent`/`evade_chance`, `item.set_piece_id`/`named_unique_id`, `boss_phase.duration_ticks`, `boss_telegraph.hit_shape`, `resource_bag_item_remove.bag_item_id` | schema |
| Stale enums / bounds | wall `kind: wood`, wall `source: room_wall/room_divider/town_perimeter`, breakdown key `evade_chance`, breakdown source `class_affinity`, item `slot: ring`, `party` `maxItems: 2` (six-player co-op), `seed` hex pattern (local-dev custom seeds), `offer_id` pattern (uppercase ULID in `character:char_…`) | schema |
| Required fields the server legitimately omits | `boss_phase.remaining_ticks` (server sends `started_tick` + `duration_ticks`; client derives remaining); event list fields dropped by Go `omitempty` when empty (`stash_items`, `sell_appraisals`, `offers`, …); `skill_damage_burst.correlation_id` (the envelope's `correlation_id` is optional); level-up `player_mana_restored` has no `item_instance_id` | schema (+1 client reader) |
| Server payload bugs | companion `visual_tint` sent as bare `RRGGBB` (rules form) instead of protocol `#RRGGBB`; `companion_bar.gd` ignores non-`#` tints, so every rule-defined companion tint was dropped for a hardcoded fallback | server |
| `rolled_stats` closed list | server type is `map[string]int` keyed by data-driven affix stat ids (`item_level`, `max_mana`, `movement_speed_percent`, `bonus_*_damage`, `light_radius`, …) | schema |

## Decision

### 1. Live gate in the protocol bot (`tools/bot/payload_schema.py`)

- A new focused module. It imports only `jsonschema`, never `tools.bot.run`. `state_ingest.py`
  calls `check_payload("state_delta" | "session_snapshot", payload)` at the top of delta
  ingestion and `ingest_snapshot`. Every co-op peer, wait pump, and the initial snapshot already
  route through those two functions. `run.py` is untouched (split freeze; it is 14 lines over
  baseline).
- **Mode** comes from `ARPG_BOT_SCHEMA_VALIDATION`:
  - `off` (default): interactive and benchmark runs pay nothing.
  - `report`: log each distinct violation once, keyed by generic path plus rule.
  - `strict`: raise `PayloadSchemaError(AssertionError)` on the first violation, which fails the
    scenario.
- `scripts/ci.sh` step 9 runs the protocol pack (`ci`, or `all` for `ci-full`) with `strict`.
  An explicit `ARPG_BOT_SCHEMA_VALIDATION` in the environment still overrides it. Benchmark
  probes stay unvalidated because they are perf probes.
- **Cost.** Plain `Draft202012Validator` takes ~20 ms for the busy 17 KB example delta: every
  `change` is tried against 24 `oneOf` branches and every `event` against 57 `allOf` `if`s.
  `CompiledPayloadSchema` compiles once per message type (`functools.cache`). It dispatches each
  `changes[]` item on `op` and each `events[]` / `recent_events[]` item on `event_type`, to a
  validator holding only the branches that can apply. That brings the delta down to ~4 ms, about
  5×. It is equivalent to the plain validator:
  - a `oneOf` whose branches pin distinct `op` consts can only pass on the matching branch;
  - an `if` on a non-matching `event_type` is vacuous;
  - unknown discriminators fall back to the full definition.

  Dispatch is built only when the structure matches that exact shape. Otherwise the def stays in
  the plain shell validator.
- **Diagnostics:**
  - Violations inside dispatched items are tagged `[op=…]` / `[event_type=…]`.
  - `oneOf` failures report from the branch with the fewest errors (the one the payload was aiming
    at), not `best_match`'s pick.
  - Messages that echo instance values (patterns, bounds, types) dedupe by rule, not by value.

### 2. Fix every violation

The ownership rule is: **fix the schema when the server shape is intentional and the client consumes
it; fix the server when the payload breaks an explicit protocol convention.**

- **Schema (both v8 files where the def exists), additive/widening in place.** This follows the
  precedent of v208 `companion_stance`, v448 `steward_hunt_target` and v485 `character_class`,
  with no version bump:
  - New optional props: every row of the table above.
  - `rolled_stats` becomes `additionalProperties: { "type": "integer" }`. Listed keys keep their
    bounds.
  - Enum and pattern widenings mirror the Go writers and `worlds.v0.schema.json` obstacle kinds.
  - `party` has no `maxItems`. No server party cap exists; `six_player_boss_combat_soak` proves six.
  - `seed` is a non-empty string. `session.go` accepts any `req.Seed` in local dev, and the unused
    `hex_seed` def is removed.
- **Required-set corrections.** These are the only non-additive edits. Each one matches what the
  server has always sent, so no conforming producer breaks:
  - `boss_phase` now requires `duration_ticks` in place of `remaining_ticks`. `remaining_ticks`
    stays an optional prop, and the examples were updated.
  - **Event list fields omitted when empty.** The `event` def documents that a list-valued event
    field (the Go `Event` slices/maps tagged `omitempty`) is omitted when empty, and readers must
    treat absent as empty. Those fields leave every `then.required`. `hits` stays required, because
    `collapseSkillDamageBurst` never emits an empty burst. The alternative, pointer slices on the
    shared `Event` struct, touches 12+ construction sites and every reader, for no wire gain.
  - `skill_damage_burst` no longer requires `correlation_id`.
  - `player_mana_restored` requires `entity_id` + `mana`, plus `item_instance_id` or `reason`.
    This mirrors `player_healed`. `level_up_resources.go` emits both with `reason: level_up`.
- **Server.** `companion_ai.go` adds `protocolVisualTint`, which converts the rules form
  `RRGGBB` (as `skills.v0.schema.json` requires) to protocol `#RRGGBB`. The revived-corpse
  constant becomes `#444441`. This is view-only state, so replay and goldens are unaffected.
  Visible effect: the companion bar portrait now uses the rule tint instead of the per-family
  fallback.
- **Client.** `main.gd` `_show_stash_panel` / `_show_blacksmith_panel` read
  `ev.get("stash_items", [])` instead of keeping the previous list. `stash_opened` and
  `blacksmith_service_opened` always describe the full stash, so an omitted list means empty.

## Tests

- `tools/bot/test_payload_schema.py`:
  - the module imports without `run`;
  - the examples are clean;
  - the compiled cache is identity-stable and dispatch tables exist for `changes` / `events` /
    `recent_events`;
  - an off-contract entity prop gives a generic key and a concrete path;
  - dispatch agrees with the plain validator on 8 mutations: unknown op, missing entity,
    wrong-typed damage, a missing per-event required field, an enum-branch required field,
    top-level extras, and a snapshot event;
  - unknown event types match the plain verdict;
  - `off` / `strict` / `report` modes, with report dedupe;
  - value-echo dedupe;
  - the `oneOf` nearest-branch leaf;
  - the `state_ingest` wiring, both with a recording gate and with strict failing an off-contract
    delta.
- `server/internal/game/companion_visual_tint_test.go`: every rules companion tint converts to
  `#RRGGBB`, the revived tint is protocol form, and the conversion is idempotent on
  empty and prefixed values.
- `make ci` runs with strict validation, which is the regression gate.

## Out of scope / follow-ups

- **A static Go-json-tag ↔ schema cross-check.** The one-off survey script found
  `boss_phase.duration_ticks` before any scenario did. A `validate_shared` cross-check would catch
  drift that no scenario exercises, such as `set_piece_id` on items.
- **Client-bot (Godot) payload validation.** The client bot consumes the same wire and is
  indirectly covered because the protocol pack drives the same server paths.
- **Reconnect snapshots in `check_persistence`** read `recv_json` directly in `run.py` and are not
  validated. Wiring them would grow the frozen `run.py`.
