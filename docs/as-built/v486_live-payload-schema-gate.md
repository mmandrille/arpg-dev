# v486 As-Built: Live Payload Schema Gate

Date: 2026-09-29
Status: Complete
Commit: pending
Baseline: v483 `armor-look` (`5e891e5e`); v484/v485 in flight elsewhere
Spec: [`v486_spec-live-payload-schema-gate.md`](../specs/v486_spec-live-payload-schema-gate.md)

## What shipped

- **`tools/bot/payload_schema.py`.** A leaf module that validates every received `session_snapshot`
  and `state_delta` payload against the v8 schemas.
  - Hooked in `state_ingest.py`, the single path for deltas and snapshots across solo, co-op
    peers and wait pumps. `run.py` is untouched.
  - The mode comes from `ARPG_BOT_SCHEMA_VALIDATION`: `off` (default), `report`, or `strict`.
    When the mode is not `off`, the gate prints a startup line.
- **CI.** `scripts/ci.sh` step 9 runs the protocol pack with `strict`; an explicit env value still
  wins. Benchmark probes stay off.
- **Cheap by construction.**
  - Validators are compiled once per message type.
  - `changes[]` dispatch on `op`, and `events[]` / `recent_events[]` dispatch on `event_type`, to
    validators holding only the applicable `oneOf` branch or `allOf` rules.
  - The busy example delta dropped from 20.1 ms to 3.8 ms (5.3×). The results match the plain
    validator, which a test asserts on 8 mutations.
- **Readable violations.**
  - Each violation is tagged with its discriminator (`[event_type=stash_opened]`).
  - `oneOf` failures report from the nearest branch.
  - Keys for value-echoing rules dedupe by rule.
  - A broken schema, such as a dangling `$ref`, becomes a gate violation rather than an ingest
    crash.
- **`suspended()` perf window.** `combat_soak_runtime.py` pauses the gate only around the measured
  cast/pump loops of `crowded_skill_overlap_lab` and `six_player_boss_combat_soak`. Setup and
  initial snapshots stay validated. See Proof → perf.

### Drift fixed

Three report-mode surveys ran over `make bot scenario=all` (135 scenarios). They found 44
distinct path/rule drifts. The schema/server owner split is in the spec. In summary:

- **Schema, both v8 files where the def exists.**
  - New optional props: entity `character_class` (same line as v485, so it merges clean) and
    `combat_stats` (delta), plus a `companion_combat_stats` def; progression `character_class`;
    derived `block_percent`/`evade_chance`; item and stash `set_piece_id`/`named_unique_id`;
    `boss_phase.duration_ticks`; telegraph `hit_shape`; `resource_bag_item_remove.bag_item_id`;
    mystery offer `source_depth`.
  - Widened enums/patterns: wall kinds and sources, breakdown key `evade_chance`, breakdown
    source `class_affinity`, item slot `ring`, shop `offer_id`/`refresh_key` with `|` and upper-case
    ULIDs, and stock availability `mystery:` ids.
  - Dropped stale bounds: `party.maxItems 2`, the `rolled_stats.hotbar_slots` max of 10 (a tuning
    duplicate), and the `hex_seed` pattern.
  - `rolled_stats` is an open integer map.
  - The `state_delta` event def referenced `#/$defs/equipped` and `#/$defs/hotbar_slot`, which
    existed only in the snapshot schema. Any unique-chest or corpse event crashed the validator,
    and it had gone unnoticed because nothing validated live traffic.
- **Required-set corrections.**
  - `boss_phase` requires `duration_ticks`, not `remaining_ticks`.
  - Omitted-when-empty event lists are no longer required. This is documented on the `event`
    def, and `hits` stays required.
  - `skill_damage_burst.correlation_id` is optional.
  - `player_mana_restored` requires item or reason.
  - The examples now carry `duration_ticks`.
- **Server.** `protocolVisualTint` in `companion_ai.go` converts rules `RRGGBB` companion tints
  to protocol `#RRGGBB`, and the revived-corpse tint is now `#444441`. `companion_bar.gd` only
  honours `#` tints, so every summoned companion portrait had been silently using the
  per-family fallback colour.
- **Client.** `main.gd` stash and blacksmith panels treat an absent `stash_items` as empty.
  Before this, an empty stash kept the previous list.

## Proof

- `make ci`: **green** with the strict gate on the ci protocol pack (6m08s, then 6m00s on the final tree).
- **Negative control.** `character_progression.character_class` was temporarily removed from
  `state_delta.v8` and `make ci` rerun. Step 9 **failed**: every protocol scenario died with
  `PayloadSchemaError: schema violation: state_delta at changes/4/character_progression [op=character_progression_update]: Additional properties are not allowed ('character_class' was unexpected)`.
  The schema was then restored byte-identical.
- **Budget.** Compared with `ARPG_BOT_SCHEMA_VALIDATION=off make ci` on the same host, the 22 ci
  protocol scenarios took 114.2 s strict vs 111.5 s off, which is +2.7 s or 2.4% total. No
  scenario approached its budget. Per-scenario differences are inside run-to-run noise: repeated
  `vertical_slice` runs swing 1.9–5.0 s in either mode. Compile plus first validation costs
  under 10 ms, and a warm busy delta costs ~4 ms.
- **Survey 3, the post-fix report run over `all`.** It found the last 10 drifts (dangling refs,
  shop patterns, `hotbar_slots`, mystery `source_depth`), which are fixed and re-verified in
  targeted runs.
- **Perf A/B, same host and minutes apart:**
  - validation `off`: both soaks log `overruns=0`;
  - `strict` without suspension: both fail `tick_overrun_ms p95=54.7, want <= 50`;
  - `strict` with `suspended()`: they pass twice (overruns 0/1 and 0/0), with elapsed within
    ~0.6 s of `off`.

  Mechanism: a lagging reader fills `sendCh`, the tick goroutine then runs `coalesceOutbound`
  under the overflow mutex, and tick time grows. This is logged as a gap in PROGRESS.
- **Tests.**
  - `tools/bot/test_payload_schema.py`: 27 tests, including import independence, examples clean,
    dispatch vs plain equivalence, modes, dedupe, nearest `oneOf` branch, every `$ref` resolving,
    broken schema reported not raised, `suspended()`, and the `state_ingest` wiring.
  - `server/internal/game/companion_visual_tint_test.go`.
  - `ranger_skills_test.go` had pinned the buggy bare tint. It now derives the expected tint
    from rules.
- `go test ./...`, `pytest tools`, `make validate-shared`, `make lint-determinism` and
  `make maintainability` all pass. The new files are under 600 lines, and no grandfathered file
  grew.

## Gaps / follow-ups

- There is no static Go `json:` tag ↔ schema cross-check. A one-off version of that check found
  `boss_phase.duration_ticks` and entity `character_class` before any scenario did.
- A slow client inflates server tick time through send-overflow coalescing on the tick path.
  This is server robustness work and is not addressed here.
- The Godot client bot and the reconnect snapshots in `run.py` `check_persistence` are not
  validated. Wiring the latter would grow the frozen `run.py`.
- Scenario coverage bounds the gate: event types and fields that no protocol scenario produces
  are still unchecked.
