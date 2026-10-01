# v501 As-Built — Monster death rim flash

- **Date:** 2026-10-01
- **Spec:** [`v501_spec-monster-death-dissolve.md`](../specs/v501_spec-monster-death-dissolve.md) · **Plan:** [`v501_2026-10-01-monster-death-dissolve.md`](../plans/v501_2026-10-01-monster-death-dissolve.md)
- **Base commit:** `5365832b9029e0d9e178d57f2a95a6a697b01acc`
- **Status:** Integrated; focused checks and combined batch `make ci` passed (11m41s, 2026-10-01).

## What changed

- Added a schema-backed client catalog for the optional monster death rim flash. Invalid, missing, unreadable, or wrong-typed catalog data disables only the new effect.
- Added a transient rim and warm emission treatment to the existing event-time death reaction. It is gated on `event_type == monster_killed` and entity type `monster`; ordinary death reaction behavior remains unchanged for other event types, players, and companions.
- Captured each affected mesh's rim and emission baseline, duplicated the reaction material per entity, and restored its properties after the configured envelope or an early reset/disposal. No authoritative state, replay, protocol, corpse lifetime, animation, or existing death-burst behavior changed.
- Added a death-time capture immediately after the existing combat scenario observes the monster death reaction. The capture name is `v501_monster_death_rim`.

## Verification to date

| Check | Result |
|---|---|
| `godot --headless --path client --script res://tests/test_monster_death_presentation.gd` | PASS: 32 checks, including configured peak, eventual baseline restoration, early reset, non-kill event, player, and companion cases. Exit 0. |
| `godot --headless --path client --script res://tests/test_death_pose_ownership.gd` | PASS. Exit 0; Godot reports 3 ObjectDB instances and 1 resource still in use at process exit. |
| `make bot-client SCENARIO=combat_input_flow_polish HEADLESS=1` | PASS: 1/1. Exit 0. |
| `AUTOPLAY_STEP_DELAY=0.05 make bot-visual scenario=combat_input_flow_polish` | PASS: 1/1 visible client route. Exit 0. Capture: `.artifacts/bot-captures/v501_monster_death_rim.png`; manifest: `.artifacts/bot-captures/v501_monster_death_rim.json`. |
| `make validate-shared` | PASS: 2210 checks; CODEMAP index passes. |
| `make validate-assets` | PASS: 451 checks. |
| `make maintainability` | PASS: file-size, extraction-coupling, and progress-dashboard checks. |
| `git diff --check` | PASS. |

The final windowed capture reports 1920×1080, Forward+, Metal, Performance tier, at scenario tick 21. Inspection shows a clearly visible pale amber/ivory glow on the corpse at the event-relative capture step. The still reads mostly as a bright full-body emission silhouette; it does not cleanly isolate a narrow rim contour from the emission and scene lighting. The scoped 0.05-second autoplay delay allowed the existing scenario-local capture to sample the configured flash without changing the global runner delay or effect envelope. The earlier `.artifacts/bot-captures/v501_monster_death_rim_balanced.{png,json}` capture is retained as preliminary evidence; it missed the transient flash.

## Proof limits and handoff

- The tests establish the event gate and material state/restoration behavior; the image demonstrates one rendered frame only. It does not establish visual consistency across all monster assets or a performance/frame-time result.
- No server, protocol, authoritative death, replay, or corpse-lifetime behavior changed.
- Combined `make ci` passed on the integrated batch; its Godot smoke includes the monster-death presentation test. See the coordinator closeout in `PROGRESS.md` for the review/refactor handoff.
