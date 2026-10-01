# v499 As-Built — Live targeting and correction feel

- **Date:** 2026-09-30
- **Spec:** [`v499_spec-live-targeting-corrections.md`](../specs/v499_spec-live-targeting-corrections.md) · **Plan:** [`v499_2026-09-30-live-targeting-corrections.md`](../plans/v499_2026-09-30-live-targeting-corrections.md)
- **Scope:** local-player visual smoothing and explicit reset semantics; server authority, movement rules, tick rate, and wire format are unchanged.

## Measured defect and correction

The live windowed `106_live_targeting_corrections` route in `combat_control_lab` repeated floor movement, monster chase, a confirmed attack result, and rapid floor retargets under local transport and deterministic `bounded_80_20` delayed transport. On the combined v494–v498 client before this correction, three ordinary coalesced player updates per run crossed the old 1.5-world-unit reset threshold. The corresponding visual node moved 1.85–1.90 units in one frame, despite a smaller raw prediction/server gap. The event was an ordinary authoritative upsert, not a teleport or level transition.

`MovementVisualSmoothing.preserve_after_anchor_move` now always retains a bounded offset for ordinary anchor changes. The obsolete reset-distance field was removed from the schema-backed presentation catalog. Snapshot application, level arrival, and local mobility explicitly reset the visual; a level-change event leaves the reset pending until the later player-position update if the two arrive in different envelopes. The existing 0.8-unit maximum offset and catch-up rate remain data-driven and unchanged. The trace distinguishes ordinary upserts from intentional resets and excludes identifiers, account data, and envelope payloads.

## Matched live A/B

Godot 4.7.2, Forward+ Metal, Apple M4 Pro, windowed 1280×720, the same `live_targeting_499` seed and scenario, and three runs per profile were used on the integrated client. Each run had five actions, five dispatch and acknowledgement samples, and six outbound targeting command dispatches before and after. Values are medians of each run's nearest-rank p95 or maximum; the snap count is per run for ordinary authoritative updates exceeding 1.5 world units in one visual frame.

| Profile | Input→visible p95 before | After | Max ordinary visual step before | After | Large steps before → after |
|---|---:|---:|---:|---:|---:|
| Local | 130.40 ms | 81.27 ms | 1.893 | 1.260 | 3 → 0 in each run |
| Bounded delay/jitter | 278.11 ms | 264.52 ms | 1.850 | 1.265 | 3 → 0 in each run |

The local input→visible median of run p95 values improved 37.7%; the delayed profile improved 4.9%, so the primary supported result is elimination of the visual jump. The largest ordinary step after the correction ranged 1.219–1.317 locally and 1.213–1.270 with delay. Both profiles stayed within the configured 0.8-unit visual **offset** bound; offset and world-frame step measure different things. There were zero late attack dispatches or late local swings in every before/after run. The frame-stall count remained noisy (4–7 after) and is not attributed to this correction. Explicit snapshot/reset signals were kept separate from ordinary correction counts.

Raw filtered traces and runner status are retained under `.artifacts/v499-integrated/before-{local,delayed}{,2,3}.jsonl` and `after-{local,delayed}{1,2,3}.jsonl`, with per-run filtered client/server logs. The direct before/after trace establishes the ordinary-upsert defect and its removal; it does not establish improved behavior on every lower-end computer or prove a server replay comparison. No server-side code or command count changed in this slice.

## Verification

`test_movement_visual_smoothing` passes its coalesced-step and explicit-reset cases (10 assertions). `test_live_targeting_trace` passes its stale-correction and delayed level-arrival cases. The six after-correction live runs passed the bot scenario. The v461 smoothing and v464 combat-input regressions passed; the final combined `make ci` passed in 11m13s. The exact visual route is `HEADLESS=0 SCENARIO=106_live_targeting_corrections ./scripts/bot_client_local.sh`; set `ARPG_BOT_TRANSPORT_PROFILE=bounded_80_20` for the delayed profile.
