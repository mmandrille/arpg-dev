# v499 — Live targeting and correction feel

- **Status:** Complete (combined `make ci` passed)
- **Date:** 2026-09-30
- **Codename:** `live-targeting-corrections`
- **Area:** Gameplay smoothness
- **Depends on:** v496 attack timing and v497 frame-pacing work, to avoid mistaking render stalls for input latency.
- **ADR:** [ADR-0001](../adr/0001-technology-stack.md) D2–D3.

## Purpose

Keep the locally controlled hero responsive and visually continuous when the player rapidly switches between moving, chasing a target, attacking, and moving away during a live session. Build a repeatable Godot-client scenario with bounded transport delay/jitter, measure the current input-to-visible-action and correction behavior, then fix the worst reproducible problem in that path. Existing v461 smoothing and v464 input-flow work remain the baseline.

## Non-goals

- Changing server authority, damage, pathfinding rules, movement/attack speed, or the 10 Hz tick rate.
- Broad retuning without a reproduced player-visible defect.
- A general network simulator product or offline replay as a substitute for live client proof.

## Acceptance criteria

- [ ] A live `runner: godot_client` scenario repeats click-to-move, click-to-chase, attack, and floor retarget while entities move. The scenario runs at local transport and at one documented, bounded delay/jitter profile owned by the test fixture.
- [ ] Instrumentation records input receipt, command dispatch, first visible local response, authoritative acknowledgement, reconciliation displacement, and rendered-frame stalls without logging sensitive account data.
- [ ] Baseline traces identify one reproducible defect or tail-latency source; the implementation changes only its owner. For a latency defect, the selected p95 metric improves by at least 25% in matched runs; for a stale-action or snap defect, the new scenario reproduces it before the fix and passes afterward. If the baseline shows no defect, stop and replace this slice with a new owner-approved feel issue rather than changing tuning values speculatively.
- [ ] A floor retarget cancels stale chase/attack intent; a late result cannot restart a canceled local swing. The hero follows the intended path without visible snap or oscillation beyond the configured visual-smoothing bound, except for explicit teleport/reset actions.
- [ ] A/B live-client traces show improved input-to-visible-response or correction p95 for the selected defect, with no new stale actions, extra server intents, or replay divergence.

## Scope and likely files

- **Client:** the measured owner among `client/scripts/attack_move_input_coordinator.gd`, `client/scripts/command_retarget_grace.gd`, `client/scripts/player_movement_feel.gd`, `client/scripts/movement_visual_smoothing.gd`, `client/scripts/reconciliation_backpressure.gd`, and the local-player bridge in `client/scripts/main.gd`.
- **Tools/bot:** add a bounded test-only transport-delay fixture and one live Godot-client scenario; extend `docs/performance/tools.md` with its trace and analysis method.
- **Data:** any new smoothing or buffer tuning belongs in the existing schema-backed shared presentation catalog, not a hardcoded client literal.
- **Contracts:** no gameplay rule or wire-schema change expected; if the trace proves a server/transport cause, the plan must call out that scope before implementation.

## Test and bot proof

- Unit tests cover pending-intent cleanup, smoothing bounds, and deterministic delay-fixture behavior where applicable.
- Run `make client-unit`, the v464 input-flow and v461 movement-smoothing scenarios, and the new live Godot-client scenario at both transport profiles.
- Use real play-camera footage and the event/frame trace together; a smooth-looking offline replay is insufficient evidence for local input response.

## Asset/plugin decision

- **Borrow** existing Godot input, camera, and presentation paths. No outside asset or plugin is needed.
- **Reject** a new input framework or network middleware for this focused correction slice.

## Open questions and risks

- The currently shipped v461/v464 behavior may already meet the goal. This is a baseline-gated slice; a no-defect finding must be reported honestly and the slot re-scoped before implementation.
- Delay injection must be test-only and bounded; it cannot alter production timing or server-authoritative outcomes.
