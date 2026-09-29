# v479 Plan: Recorded Co-op Member Lifecycle

Spec: [`v479_spec-recorded-member-lifecycle.md`](../specs/v479_spec-recorded-member-lifecycle.md)

1. **Failing test first.** Add `realtime/member_lifecycle_replay_test.go`, reusing the v476
   `loadShedMemRepo` harness plus disconnect/listing stubs. Cover:
   - join → fight → leave mid-fight → `replay.Verify` matches,
   - join → leave → rejoin → `replay.Verify` matches,
   - negative controls where the leave or rejoin is applied without a row, which must fail `Verify`,
   - a full leave, loop stop, and resume, which must restart on the exact live tick and verify.
   Confirm the positive cases fail on v476.
2. **Sim.** Add `game/member_lifecycle.go` with the three `SystemMember*InputType` constants, the
   `MemberLifecycle` payload, `ApplyMemberLeave`, `ApplyMemberRejoin`, and the
   `IsMemberLifecycleInput` helper. Add an `Input.Member` field.
3. **Decode.** In `inputdecode/system.go`, add `EncodeStoredMemberLifecycle` and extend the
   stored-only decode.
4. **Runner.** Add `realtime/member_lifecycle.go` with `recordMemberLifecycleLocked` (seq/seen/row)
   and `persistSystemInput`, and share `systemInputRowLocked` with the load-shed path. Wire it into
   `detach` (leave), `admitMemberLocked` (rejoin), and `playerIDForMember` (join). Rows persist under
   `l.mu`, next to the session_members write those paths already make there.
5. **Replay.** Move the member roster code (`sessionStartSim`, pending members, connectivity) out of
   `replay.go` into `replay/members.go`. Merge the duplicated reconstruct loops into one stepper
   (`replayTicks`) that applies membership rows after their stamped tick, in sequence order.
   `BuildTimeline` uses the same stepper. This shrinks `replay.go` below its 751 baseline.
6. **Unit tests.** Decode gating and round-trip (`inputdecode/system_test.go`), and sim apply
   semantics (`game/member_lifecycle_test.go`).
7. **Verify.** Run `cd server && go test ./...`, `make lint-determinism`, `make maintainability`, and
   `make ci` (with a unique `CI_ADDR`). Lower the ratchet baselines, then update the as-built,
   PROGRESS, the lifecycle and codename indexes, and CODEMAP.
