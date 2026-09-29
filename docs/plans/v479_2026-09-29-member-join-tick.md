# v479 Plan: Member Join Tick

Spec: [`v479_spec-member-join-tick.md`](../specs/v479_spec-member-join-tick.md)

1. **Reproduce.** Add `realtime/member_join_tick_replay_test.go`, which reuses the v478
   `member_setup_replay_test.go` harness. The host attaches and descends, `repo.joinGuest()` runs
   without an admit, then the loop settles and the test runs `replay.Verify`. A second test joins
   the guest before the loop builds. Confirm both fail on v478.
2. **Store.** Add `store.SessionMemberNotJoinedTick = -1`. `SetSessionMemberPlayer(..., joinedTick)`
   writes `joined_tick`. Update the Postgres repo, the realtime in-memory repo, and the replay
   fake. Add a store test for the new write.
3. **Live.** `buildSessionSim` passes `0` for the host and for each guest.
   `playerIDForMember` passes `Sim.PlayerJoinedTick(playerID)`. The HTTP join handler uses the
   constant.
4. **Replay.** Add `sessionsetup.JoinedSim(member)`. `sessionStartSim` skips members that
   fail it before resolving them.
5. **Verify.** `go test ./...`, `make lint-determinism`, `make ci` on an isolated port. Then
   update the as-built, PROGRESS (gap closed, legacy caveat), and the indexes.
