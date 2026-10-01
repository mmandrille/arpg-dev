# v499 Plan — Live targeting and correction feel

Status: Complete; combined `make ci` passed in 10m33s.
Goal: Measure the local player's live move/chase/attack/retarget path under local and bounded delayed transport, then correct only a demonstrated owner.
Architecture: A test-only Godot transport fixture delays complete WebSocket envelopes in order, using a deterministic bounded schedule. A live client-bot scenario drives the existing input bridge and emits per-action, per-frame trace markers without account or session data. The server remains authoritative; no protocol, shared gameplay rule, or replay format changes are planned.
Tech stack: Godot 4 GDScript, client-bot scenario JSON, existing shell runner, Python trace analysis if needed.

## Baseline and shortcut decision

`PROGRESS.md` identifies v493 as the completed baseline and marks v496/v497 as To do. This worktree can prove the measurement apparatus against v493, but the final defect selection and A/B claim must wait for the integrated v496/v497 baseline. Reuse v461 smoothing and v464 input-flow paths. **Borrow** the in-repo Godot input, client-bot, WebSocket, and camera facilities. **Reject** external assets, plugins, input frameworks, and network middleware.

## File map

| Action | Path | Responsibility |
| --- | --- | --- |
| Create | `client/scripts/bot_transport_delay.gd` | Deterministic, bounded FIFO release schedule for test envelopes |
| Modify | `client/scripts/net_client.gd` | Test-only transport seam and dispatch/receive trace hooks |
| Create | `client/scripts/live_targeting_trace.gd` | Bounded action/frame timeline and summary metrics |
| Modify | `client/scripts/bot_controller.gd` | Start/end trace and label actions while the live client bot runs |
| Create | `client/scripts/bot_action_formatter.gd` | Independent extraction of action log formatting from the oversized bot controller |
| Modify | `scripts/bot_client.sh` | Retain privacy-safe targeting trace lines under `.artifacts/v499/` |
| Create | `tools/bot/scenarios/client/106_live_targeting_corrections.json` | Move/chase/attack/floor-retarget live sequence |
| Create | `client/tests/test_bot_transport_delay.gd` | Fixture bounds, ordering, and deterministic schedule proof |
| Modify | `docs/performance/tools.md` | Reproduction, trace fields, matched-run method, and interpretation limits |
| Conditional | Measured client owner and focused tests | Only after integrated baseline reproduces a defect |

## Maintenance ratchet

Target: new source/test/tool files remain at or below 600 lines. `main.gd` is deliberately excluded from the measurement seam. `bot_controller.gd` is grandfathered; keep additions minimal and extract cohesive tracing into its own file. `net_client.gd` is below 600 lines but close enough that transport scheduling belongs in a separate file. Verify with `make maintainability`; if the existing `bot_controller.gd` baseline is exceeded, lower its net line count through a genuine extraction before completion.

## Task 1 — Test-only transport and trace

Files: `client/scripts/bot_transport_delay.gd`, `client/scripts/net_client.gd`, `client/scripts/live_targeting_trace.gd`, `client/scripts/bot_controller.gd`, `client/scripts/bot_action_formatter.gd`, `client/tests/test_bot_transport_delay.gd`, `scripts/bot_client.sh`.

- [x] Parse one allowlisted test profile only when `ARPG_BOT_CLIENT=1`; reject invalid values and cap delay/jitter.
- [x] Delay outbound and inbound complete envelopes with stable FIFO ordering; clear queues on reconnect/close and never alter payloads or server outcomes.
- [x] Record bot input receipt, transport dispatch, first visual-node response, authoritative ack, reconciliation displacement, and frame stalls with monotonic timestamps and no account/session identifiers.
- [x] Test deterministic scheduling, bounds, ordering, and no-delay behavior.

Verify: `make client-unit` and `make maintainability`.

## Task 2 — Live scenario and baseline

Files: `tools/bot/scenarios/client/106_live_targeting_corrections.json`, `docs/performance/tools.md`.

- [x] Use a moving monster in the existing combat lab; repeat floor move, chase, attack, and floor retarget without incidental navigation setup.
- [x] Run the new scenario locally and with the documented bounded delay/jitter profile, collecting at least three matched runs per profile and retaining trace files in `.artifacts/`.
- [x] Run the v464 input-flow and v461 movement-smoothing scenarios as regressions. Record exact server/client commit, Godot version, display mode, scenario seed, sample counts, p50/p95, and frame-stall counts in `docs/performance/v499-provisional-baseline.md`.
- [ ] Review visible play-camera footage together with trace. Mark the v493 result provisional; rerun on the integrated v496/v497 baseline before selecting a correction.

Verify: `HEADLESS=1 make bot-visual scenario=106_live_targeting_corrections`; `HEADLESS=1 make bot-visual scenario=103_combat_input_flow_polish`; `HEADLESS=1 make bot-visual scenario=80_movement_visual_smoothing`; visible `make bot-visual scenario=106_live_targeting_corrections`.

## Task 3 — Defect-gated correction

Files: only the owner identified by the integrated trace, its focused unit tests, and the scenario assertion.

- [ ] Check whether v496/v497 are integrated. If absent, stop before this task.
- [ ] Reproduce one stale action, snap, oscillation, or latency tail against the integrated baseline; identify its owner from event/frame evidence.
- [ ] Change only that owner, keeping tuning in the existing schema-backed presentation catalog if tuning is actually required.
- [ ] Run matched A/B live client traces. Require the spec's p95 improvement or before/after regression proof, no extra server intents, and replay equality.

Verify: focused client unit test, both live profiles, v461/v464 regressions, `make replay SESSION_ID=<recorded-id>` when a session is available, and `make ci` only when the completed slice is ready for integration.

## Lifecycle and handoff

After the dependency and correction gates pass, write `docs/as-built/v499_live-targeting-corrections.md` and update `PROGRESS.md` plus lifecycle index on the integration branch. This isolated worktree must not create a branch, commit, transfer changes to `main`, or run `/finish`.

## Integrated result

The combined v494–v498 baseline exposed three ordinary local-player visual jumps above 1.5 world units in every run of both transport profiles. The selected owner was `movement_visual_smoothing.gd`: its distance threshold incorrectly reset presentation on coalesced authoritative upserts. Ordinary updates now preserve the configured bounded offset; snapshot, level arrival, and local mobility retain explicit reset behavior. Three windowed local and three bounded-delay runs after the change had zero such jumps, six outbound targeting command dispatches per run as before, and zero late swings or attack dispatches. [As-built values and limits](../as-built/v499_live-targeting-corrections.md) supersede the provisional v493 measurement above. The v461/v464 regressions and the owner's single combined `make ci` remain the final gate.
