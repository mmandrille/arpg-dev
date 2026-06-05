# Resume Authoritative State (Slice v5) - Implementation Plan

Status: Complete (2026-06-05)

Goal: Make same-session reconnect restore authoritative combat/world state by
replaying recorded inputs before the initial resume snapshot.

Architecture: Extend the existing `replay.Reconstruct` path (already used by
`GET /state` and `make replay`) so WebSocket resume uses the same algorithm.
No second persisted entity snapshot table; no `LoadInventory` when inputs exist.

Review findings closed:
- `replay.Reconstruct` already replays seed + ordered inputs with tick gaps;
  `/state` already returns the correct combat snapshot. The gap is only
  `hub.Run` (`NewSim + LoadInventory`).
- Applying `LoadInventory` on top of replay would duplicate inventory and
  corrupt `nextID` — resume with recorded inputs must be replay-only.
- `replay` imports `realtime` for `DecodeStored`; wiring `hub` to `replay`
  creates `realtime → replay → realtime`. Break the cycle before Task 2.
- `server/internal/replay/replay_test.go` does not exist yet; create it for
  package-level reconstruction tests (HTTP replay tests stay in `http/replay_test.go`).
- WebSocket attach has no separate "resume" flag; use `len(session_inputs) > 0`
  to choose reconstruction vs fresh `NewSim`.

Tech stack: Go deterministic sim/replay/realtime tests, Postgres-backed store
contracts, Python protocol bot, Godot headless smoke.

Spec: `docs/specs/spec-resume-authoritative-state.md`
Baseline: slice v4 `take-a-hit` (complete on `main`)
Branch: `feature/resume-authoritative-state`

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/realtime/protocol.go` | Move `DecodeStored` out of `realtime` to break import cycle |
| Create or modify | `server/internal/inputdecode/` (or `game`) | Persisted payload bytes → `game.Input` without `store` or `realtime` |
| Modify | `server/internal/replay/replay.go` | Extend `Reconstruct`: return `*game.Sim`, derived events, `ResumeMetadata` |
| Modify | `server/internal/realtime/hub.go` | Replay reconstruction when inputs exist; never `LoadInventory` on that path |
| Modify | `server/internal/realtime/runner.go` | `newRunner(..., meta *ResumeMetadata)` seeds `seen`/`seq`, empty `buffer` |
| Modify | `server/internal/http/realtime.go` | Hub loads inputs via store; inventory DB load optional for validation only |
| Modify | `server/internal/http/inspect.go` | Confirm `/state` stays on extended `Reconstruct` (already replay-based) |
| Create | `server/internal/replay/replay_test.go` | Package tests: combat restore, resume metadata, snapshot vs `Verify` |
| Modify | `server/internal/http/ws_test.go` | Resume integration: HP, monster death, `/state` parity, duplicate, dead-player |
| Modify | `client/scripts/smoke.gd` | Assert real resumed monster death/player HP; remove forced death workaround |
| Modify | `tools/bot/run.py` | Resume probe asserts HP and monster death, not just inventory |
| Modify | `docs/specs/spec-take-a-hit.md` | Replace resume limitation with v5 as-built pointer after implementation |
| Modify | `docs/PROGRESS.md` | Mark v5 complete after implementation and remove the open resume gap |

## Task 1: Break Import Cycle And Extend Reconstruct

Files:
- Create or modify: `server/internal/inputdecode/` (or move decode to `game`)
- Modify: `server/internal/realtime/protocol.go`
- Modify: `server/internal/replay/replay.go`
- Create: `server/internal/replay/replay_test.go`

- [x] Step 1.0: Move `DecodeStored` (persisted payload → `game.Input`) to a
  neutral package so `replay` no longer imports `realtime`.
- [x] Step 1.1: Extend existing `Reconstruct` (do not fork a second algorithm) to
  return the live `*game.Sim` in addition to snapshot and derived events.
- [x] Step 1.2: Add `ResumeMetadata` with `SeenMessageIDs` and `NextSequence`
  (max recorded sequence + 1); preserve tick gaps (already implemented).
- [x] Step 1.3: Add a store-independent `ReconstructFromInputs(sessionID, seed,
  rules, []RecordedInput)` helper; keep `store.SessionInput` → `game.Input`
  conversion outside `game`.
- [x] Step 1.4: Refactor `Verify` to call the extended helper; behavior unchanged.
- [x] Step 1.5: Add `replay_test.go` asserting reconstructed player `hp < 10`,
  monster `hp == 0` after v4 slice inputs, and metadata populated.

Verification:

```bash
cd server && go test ./internal/replay/...
```

## Task 2: WebSocket Resume Uses Reconstructed Sim

Files:
- Modify: `server/internal/realtime/hub.go`
- Modify: `server/internal/realtime/runner.go`
- Modify: `server/internal/http/realtime.go`

- [x] Step 2.1: Hub loads `session_inputs` for the session before creating the
  runner.
- [x] Step 2.2: If `len(inputs) == 0` (brand-new session, first connect), keep
  `game.NewSim` only — equivalent to empty replay.
- [x] Step 2.3: If `len(inputs) > 0`, call extended `Reconstruct`; use returned
  `*game.Sim` and `ResumeMetadata`. **Do not** call `sim.LoadInventory`.
- [x] Step 2.4: Change `newRunner` to accept optional `*ResumeMetadata`; when
  non-nil, seed `runner.seen` from historical message IDs, set `runner.seq` to
  `NextSequence`, leave `runner.buffer` empty.
- [x] Step 2.5: Ensure the first `session_snapshot` reflects restored
  `server_tick`, entity HP, inventory, equipped state, and `nextID` continuity.
- [x] Step 2.6: In `handleWS`, stop passing persisted inventory into hub for sim
  hydration (hub loads inputs itself). Optional: log or test-assert inventory
  parity between replay output and `ListInventory` without applying DB rows to sim.
- [x] Step 2.7: Malformed historical input payload: fail closed with server error;
  log `session_id`, input id, tick, and message id (spec §6).

Verification:

```bash
cd server && go test ./internal/http/... -run Resume
```

## Task 3: Inspection Endpoint Parity (Mostly Verification)

Files:
- Modify: `server/internal/http/inspect.go` (only if `Reconstruct` signature changes)
- Modify: `server/internal/http/ws_test.go`

`/state` already calls `replay.Reconstruct`. This task confirms it stays on the
extended helper and adds cross-path regression coverage.

- [x] Step 3.1: Update `handleSessionState` only if the extended `Reconstruct`
  return signature requires it; semantics unchanged.
- [x] Step 3.2: Add `TestResumeSnapshotMatchesStateEndpoint`: drive the v4 slice,
  close WebSocket, reconnect, decode initial `session_snapshot` payload, `GET
  /state` with debug token — assert entity HP, inventory, and equipped weapon
  match (acceptance #7–8).
- [x] Step 3.3: Keep replay verification behavior unchanged: mismatched derived
  events are reported in the replay report.

Verification:

```bash
cd server && go test ./internal/http/... -run 'Resume|State'
```

## Task 4: Server Integration Tests

Files:
- Modify: `server/internal/http/ws_test.go`
- Modify: `server/internal/replay/replay_test.go`

- [x] Step 4.1: Script the v4 flow through WebSocket until the dummy dies and the
  player takes retaliation damage.
- [x] Step 4.2: Close the socket, reconnect the same `session_id`, and assert the
  initial snapshot has monster `hp == 0` and player `hp < 10`.
- [x] Step 4.3: Assert equipped `rusty_sword` still appears after resume.
- [x] Step 4.4: Re-send a historical message ID after resume and assert
  `intent_rejected.reason == "duplicate"`.
- [x] Step 4.5: **Separate scenario** (not default slice flow): dead-player
  resume test via focused fixture or high-HP monster setup; assert post-resume
  move/attack/pickup/equip reject with `player_dead`.
- [x] Step 4.6: Assert post-resume entity/item allocation does not collide with
  historical IDs.

Verification:

```bash
cd server && go test ./...
```

## Task 5: Bot And Client Smoke

Files:
- Modify: `tools/bot/run.py`
- Modify: `client/scripts/smoke.gd`

- [x] Step 5.1: Update `check_persistence` docstring: reconnect asserts replay-
  reconstructed state, not DB inventory reload into a fresh sim.
- [x] Step 5.2: After the scripted flow, open a new WebSocket for the same session
  and assert resumed player `hp < 10`.
- [x] Step 5.3: In the bot resume probe, assert training dummy `hp == 0`.
- [x] Step 5.4: In Godot smoke, remove the forced `hp=0` monster workaround in
  the resume phase.
- [x] Step 5.5: In Godot smoke, assert the real server snapshot carries the dead
  monster and the mounted weapon from equipped snapshot state.
- [x] Step 5.6: Preserve existing v4 live assertions: player damaged, monster
  hit/death animation, equip visual, and move-after-equip.

Verification:

```bash
make bot
make client-smoke
```

## Task 6: Docs And Final Gate

Files:
- Modify: `docs/specs/spec-take-a-hit.md`
- Modify: `docs/specs/spec-resume-authoritative-state.md`
- Modify: `docs/plans/2026-06-05-resume-authoritative-state.md`
- Modify: `docs/PROGRESS.md`

- [x] Step 6.1: Update the v4 spec limitation note to point to v5 once the
  server resume gap is fixed.
- [x] Step 6.2: Add v5 to `docs/PROGRESS.md` lifecycle table as complete only
  after implementation and CI pass.
- [x] Step 6.3: Move the "combat/world state does not persist on resume" gap out
  of open gaps after acceptance is met.
- [x] Step 6.4: Record as-built notes: import-cycle resolution, no
  `LoadInventory` on replay resume, crash-window / input-without-event behavior.
- [x] Step 6.5: Run full CI.

Final verification:

```bash
make validate-shared
cd server && go test ./...
make bot
make client-smoke
make ci
```

## Self-Review Checklist

- [x] Spec §7 acceptance maps to concrete tests.
- [x] No protocol schema bump unless implementation changes message shapes.
- [x] `game` remains deterministic and store-independent.
- [x] Resume, replay, and `/state` use one reconstruction algorithm (`Reconstruct`).
- [x] No `LoadInventory` when `len(session_inputs) > 0`.
- [x] `realtime` does not import `replay` while `replay` imports `realtime`.
- [x] Client no longer carries a forced-death workaround for server resume.

## As-Built Notes

- `server/internal/inputdecode` owns protocol envelope-to-`game.Input` decoding;
  `realtime` delegates live decode and `replay` delegates stored decode there.
- `ReconstructFromInputs` accepts `[]RecordedInput` so tick gaps stay explicit
  without pushing persistence concerns into `game`.
- WebSocket resume chooses replay when `session_inputs` is non-empty; fresh first
  attach remains `game.NewSim`.
- Resume reconstruction fails closed before WebSocket upgrade on malformed
  historical input payloads and logs session/input/tick/message context.
- Recorded events are verification data only. If an input exists without a
  matching event after a crash window, replay still applies the input and
  `/replay` reports any event mismatch.
