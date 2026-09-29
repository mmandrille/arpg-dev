# v476 Plan: Recorded Load Shedding

Spec: [`v476_spec-recorded-load-shed.md`](../specs/v476_spec-recorded-load-shed.md)

1. **Reproduce and prove the root cause.** Replay the benchmark sessions still in local Postgres.
   Diff derived against recorded events to find the first divergence (tick 26, before the join).
   Correlate with `session_tick_budget_overrun` log lines. Inject the logged degradations in a
   scratch harness and confirm a full event match.
2. **Sim.** Add `game/load_shed.go`: `SystemLoadShedInputType`, `LoadShedDirective`,
   `ApplyLoadShed`, `LoadShedChanges`, `splitLoadShedInputs`. `TickResultsProfiled` splits
   directives out and applies them after `finalizeResults`. Unexport the two raw hooks. Move the
   `Input` struct from `sim.go` to `game/input.go` and add the `LoadShed` field, which shrinks the
   grandfathered `sim.go`.
3. **Decode.** Add `inputdecode/system.go`: `EncodeStoredLoadShed` and a stored-only decode path
   in `DecodeStored`. The type stays out of `IsClientIntent`.
4. **Runner.** Add `realtime/load_shed.go`: `loadShedSample`, the `loadShedPolicy` seam
   (`wallClockLoadShed` default), and `applyLoadShed`, which returns the input row. `doTick`
   persists the row after unlocking.
5. **Join tick.** Add `playerState.JoinedTick` and `Sim.PlayerJoinedTick`. Split `attach` into
   connection wiring plus `admitMemberLocked`, and move admission and `playerIDForMember` to
   `realtime/session_admission.go` (this shrinks the grandfathered `session_loop.go`).
6. **Diagnostics.** Add `logReplayFailure` to the three replay-backed inspect endpoints.
7. **Tests.** `realtime/load_shed_replay_test.go` covers the live loop → `replay.Verify` (positive
   and negative control) and the joined-tick race. Add unit tests for decode gating and the sim
   directive split and timing. Mutation-check both fixes.
8. **Verify.** `go test ./...`, `make lint-determinism`, `make ci`. Lower the ratchet baselines for
   `sim.go` and `session_loop.go`. Update the as-built, PROGRESS, and the lifecycle and codename
   indexes.
