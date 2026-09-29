# v478 Plan: Shared Member Setup

Spec: [`v478_spec-shared-member-setup.md`](../specs/v478_spec-shared-member-setup.md)

1. **Reproduce.** `realtime/member_setup_replay_test.go` runs a live `sessionLoop` on
   `dungeon_levels`. The host (or a late-join guest) has a level -1 corpse and a resource bag
   item. The host walks to the town stairs (`move_to_intent`, `descend_intent`) and engages a
   monster. Then the test runs `replay.Verify` and compares live and replay entity IDs. Confirm
   both tests fail on v476.
2. **Shared setup.** Add `internal/sessionsetup` with `Members`, `Host`, `IsHost`, `Resolve`,
   `NewHostSim`, `AddGuest`, `DisplayName`, `WorldID`, and the converters (`convert.go`, which
   keeps `WeaponSet`).
3. **Replay.** `sessionStartSim` and `addMemberToSim` use `sessionsetup`. `pendingMember`
   becomes a resolved `sessionsetup.Member`. Delete replay's converters, member sort,
   `displayNameForMember`, and `normalizeWorldID`.
4. **Live.** `buildSessionSim` and `playerIDForMember` use `sessionsetup`, with I/O outside the
   lock and `AddGuest` under it. Delete `session_corpses.go`, `progressionStateForMember`,
   `displayNameForMember`, and the dead converters in `hub.go`.
5. **Corpse snapshot.** Migration `0031_session_start_character_corpses.sql`. Add
   `SessionStartSnapshot.Corpses`. Extract `CreateSessionStartSnapshot` from `repos.go` into
   `store/session_start_snapshot.go`, taking a struct and inserting corpses in the same
   transaction. `LoadSessionStartSnapshotForMember` loads corpses. Add the table to the
   delete-character and stale-session cleanup lists. The HTTP `createSessionStartSnapshot`
   captures `ListRecoverableCharacterCorpses`.
6. **Verify.** `go test ./...`, `make lint-determinism`, `make ci`. Remove `replay.go` from the
   ratchet baseline and lower `session_loop.go` and `repos.go`. Update the as-built, PROGRESS,
   and the indexes.
