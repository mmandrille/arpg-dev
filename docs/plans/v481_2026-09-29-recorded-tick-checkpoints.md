# v481 Plan: Recorded Tick Checkpoints

Spec: [`v481_spec-recorded-tick-checkpoints.md`](../specs/v481_spec-recorded-tick-checkpoints.md)

1. **Failing tests first.** Add `realtime/tick_checkpoint_replay_test.go` on the v476 in-memory repo
   harness:
   - solo long-move stop → the rebuilt loop resumes on the exact tick at the same position, and
     `Verify` matches,
   - a quiet stretch writes a checkpoint, and a dropped loop (simulated crash) rebuilds within the
     interval,
   - a negative control without checkpoints that resumes behind live,
   - `Hub.Shutdown`.
   Confirm the positive cases fail on v479.
2. **Types.** Add `game.SystemTickCheckpointInputType`.
3. **Decode.** Extend `inputdecode/system.go` with `EncodeStoredTickCheckpoint` and the stored-only
   decode.
4. **Runner.** Add `realtime/tick_checkpoint.go` with `durableThrough` tracking, the quiet
   checkpoint after load shed in `doTick`, and a stop path that waits for the tick goroutine and
   then records. `detach` and `loopForSession` use it. Add `Hub.Shutdown` and wire it in
   `cmd/arpg-server/main.go`.
5. **Replay.** Drop checkpoint rows before `TickResults`.
6. **Verify.** Run `go test ./...`, `make lint-determinism`, `make maintainability`, and `make ci`
   with `CI_ADDR`/`CI_BASE_URL`. Revert Godot `.import` churn. Update the as-built, PROGRESS,
   indexes, and CODEMAP.
