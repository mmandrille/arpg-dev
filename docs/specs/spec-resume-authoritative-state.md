# Spec: `resume-authoritative-state`

Status: Complete (2026-06-05)
Branch: `feature/resume-authoritative-state`
Slice: v5 - authoritative combat/world state survives session resume
Baseline: slice v4 `take-a-hit` (complete on `main`)
Related: ADR-0001 (authoritative server, deterministic replay, event-sourced
session log), ADR-0007 (snapshot HP drives terminal animation state).

## 1. Purpose

Fix the highest-signal correctness gap left after v4: reconnecting to an
existing session currently rebuilds the sim from `seed + inventory` only, so
authoritative combat/world state is lost. The killed training dummy respawns at
full HP and player HP resets to full even though v4 proved real player damage.

After this slice:

- A WebSocket resume reconstructs the authoritative sim from the recorded session
  input stream before sending the initial `session_snapshot`.
- The resumed snapshot reflects the same server-owned state the live session had
  reached: player HP, monster HP/death, loot entities, inventory, equipped item,
  `server_tick`, and deterministic `nextID` allocation.
- Client smoke no longer forces monster `hp = 0` to test snapshot death pose; it
  receives the real dead monster from the server resume snapshot.
- Replay remains the canonical determinism check. Resume uses the same
  seed-plus-ordered-input model rather than introducing a second mutable
  authoritative snapshot source.

This slice proves restart/reconnect durability for the v0-v4 solo session. It is
not a broader save-game, progression, or character-scoped inventory system.

## 2. Non-goals

- No character-scoped inventory. Inventory remains session-scoped.
- No new protocol message type, schema version, or envelope bump.
- No respawn, healing, checkpoints, zone transitions, or save slots.
- No multiplayer handoff, cross-process session migration, or session locking.
- No historical inspect API beyond the existing replay and `/state` endpoints.
- No client-side patching of resume state. The server snapshot is authoritative.
- No new snapshot table unless implementation proves replay-based restore cannot
  satisfy acceptance within this slice.

## 3. Files to create or modify

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/realtime/protocol.go` | Move `DecodeStored` (or equivalent) to a neutral package so `replay` does not import `realtime` (breaks `realtime` → `replay` cycle) |
| Create or modify | `server/internal/inputdecode/` (or `game`) | Store-independent conversion from persisted input payload bytes to `game.Input` |
| Modify | `server/internal/replay/replay.go` | Extend existing `Reconstruct` to return restored `*game.Sim`, derived events, and resume metadata (`seen` message IDs, next sequence) |
| Modify | `server/internal/realtime/hub.go` | When the session has recorded inputs, build the sim via replay reconstruction only — **do not** call `LoadInventory` |
| Modify | `server/internal/realtime/runner.go` | Accept optional resume metadata; seed `seen`/`seq`; initialize empty `buffer` on resume |
| Modify | `server/internal/http/realtime.go` | Let hub load inputs through `store.Repository`; persisted inventory load is optional validation only, not applied to the sim on replay resume |
| Modify | `server/internal/http/inspect.go` | Keep `/state` on the same `replay.Reconstruct` path (already replay-based; verify parity with WebSocket resume) |
| Modify | `server/internal/store/interfaces.go` | Add narrow list/query methods only if existing `ListInputs`/`ListEvents` are insufficient |
| Modify | `server/internal/http/ws_test.go` | Add reconnect/resume integration coverage for persisted HP, monster death, `/state` parity, and duplicate rejection |
| Create | `server/internal/replay/replay_test.go` | Package-level tests: reconstruction restores combat state; resume metadata; snapshot matches `Verify` report |
| Modify | `client/scripts/smoke.gd` | Remove forced monster-death resume harness; assert real server snapshot carries dead monster and reduced player HP |
| Modify | `tools/bot/run.py` | Add resume probe after the slice flow and assert HP/monster death survive reconnect |
| Modify | `docs/specs/spec-take-a-hit.md` | Replace the v4 resume limitation note with a v5 as-built pointer after implementation |
| Modify | `docs/PROGRESS.md` | Mark v5 complete when shipped and remove this gap from open items |

## 4. Data shapes

### 4.1 No protocol shape change

`session_snapshot` already carries the fields required to render restored state:

```json
{
  "server_tick": 42,
  "session_id": "sess_...",
  "seed": "deadbeefdeadbeef",
  "entities": [
    { "id": "1001", "type": "player", "position": { "x": 12, "y": 5 }, "hp": 9, "max_hp": 10 },
    { "id": "1002", "type": "monster", "monster_def_id": "training_dummy", "position": { "x": 12, "y": 5 }, "hp": 0, "max_hp": 3 }
  ],
  "inventory": [],
  "equipped": { "weapon": null },
  "recent_events": []
}
```

The slice changes the **source** of that snapshot on resume, not its JSON shape.

### 4.2 Restore input

Restore uses the existing durable session data:

- `sessions.seed`
- `session_inputs` ordered by `(tick, sequence, message_id)` — **sole authoritative
  source** for inventory, equipment, combat, and world state on resume
- `inventory_items` — **not** loaded into the sim when recorded inputs exist.
  Pickup and equip intents in the input stream already reconstruct inventory.
  Persisted inventory may be compared offline or in tests for consistency; applying
  `LoadInventory` on top of replay would duplicate items and corrupt `nextID`.
- `session_events` for optional verification/mismatch reporting, not as the
  state source

**Rule:** if `len(session_inputs) > 0`, resume is replay-only. If there are no
recorded inputs (first WebSocket attach to a brand-new session), a fresh
`NewSim` is sufficient and equivalent to empty replay.

### 4.3 Runner resume metadata

The live runner must start with metadata derived from recorded inputs:

| Field | Source | Why |
|-------|--------|-----|
| `sim.CurrentTick()` | replayed sim | New inputs must be clamped after restored state |
| `runner.seq` | max recorded input sequence + 1 | New same-connection inputs stay ordered after restored inputs |
| `runner.seen` | recorded message IDs | Resent old intents are rejected as duplicates |
| `runner.buffer` | empty | Inputs through the restored tick have already been applied |

## 5. Architecture and flow

### 5.1 Current behavior

```text
resume websocket
  -> game.NewSim(session_id, seed, rules)
  -> sim.LoadInventory(persisted session inventory)
  -> runner sends session_snapshot
```

This restores equipped inventory but loses combat/world mutations.

### 5.2 Target behavior

There is no separate "resume" flag on the WebSocket path. Every attach uses the
same hub entrypoint; detection is whether the session has recorded input history.

```text
websocket attach (session_id)
  -> load session_inputs ordered by tick, sequence, message_id
  -> if inputs empty:
       game.NewSim(session_id, seed, rules)   # first connect to new session
     else:
       replay.Reconstruct (extend existing)    # seed + ordered inputs only
       DO NOT call sim.LoadInventory           # inventory comes from replay
  -> initialize runner dedupe/sequence metadata from recorded inputs (if any)
  -> runner sends session_snapshot from sim
```

`replay.Reconstruct` already implements tick replay and powers `GET /state` and
`make replay` verification. v5 extends it to return the live `*game.Sim` and
resume metadata, then wires `hub.Run` to use that path instead of
`NewSim + LoadInventory`.

Replay must preserve original input tick boundaries. If there are gaps between
recorded input ticks, the sim advances empty ticks until the next recorded tick,
because movement and future tick-based systems may depend on elapsed ticks.

### 5.3 Replay sharing and import boundaries

The implementation must avoid separate "resume replay" and "verify replay"
algorithms. Extend the existing `server/internal/replay.Reconstruct` helper so it
can:

1. Build a restored `*game.Sim` from `session + ordered inputs` (already mostly
   implemented; today it discards the sim after snapshot).
2. Return derived events for replay verification (already implemented).
3. Return resume metadata (`seen message_id`s, max sequence + 1) for realtime
   (new).

**Import cycle (mandatory before hub wires replay):** `replay` currently imports
`realtime` for `DecodeStored`. `hub` lives in `realtime` and must import
`replay`, which would create `realtime → replay → realtime`. Move persisted-input
decoding to a neutral package (`inputdecode`, `game`, or similar) before Task 2.
Do not make `game` depend on `store`.

### 5.4 Input/event mismatch policy

`session_events` remain the replay verification source. Resume may log or count a
mismatch between derived and recorded events, but the restored sim state must come
from deterministic replay of recorded inputs. If implementation discovers a
persisted input without corresponding recorded outputs, treat it as a replay
consistency issue and document the chosen behavior in the plan/as-built notes.

## 6. Failure behavior

- Missing session: unchanged 404/authorization behavior from session create/resume.
- Malformed historical input payload: resume fails closed with a server error and
  logs `session_id`, input id, tick, and message id.
- Replay mismatch against recorded events: log with enough context; replay
  endpoint reports mismatch as today.
- Duplicate input after resume: rejected as `duplicate` because recorded message
  IDs seed `runner.seen`.
- Dead player after resume: server rejects new gameplay intents with
  `player_dead`; client gates input from snapshot HP as in v4.

## 7. Acceptance criteria

1. A completed v4 slice session resumes with the training dummy still at `hp == 0`.
2. The same resumed snapshot carries player `hp < 10`.
3. Equipped `rusty_sword` still restores from snapshot.
4. New input after resume continues from the restored tick and allocates no
   duplicate entity or item IDs.
5. Re-sending a pre-resume message ID is rejected as `duplicate`.
6. If the player is dead in the restored snapshot, move/attack/pickup/equip are
   rejected with `player_dead`.
7. `/v0/sessions/{id}/state` and WebSocket resume agree on the reconstructed
   authoritative snapshot for the same session (payload equality modulo envelope).
8. WebSocket resume `session_snapshot` matches `replay.Reconstruct` output for the
   same `session_id`.
9. `make replay SESSION_ID=<id>` still verifies a post-v5 recorded session.
10. Client smoke removes the forced-death resume workaround and asserts real
    server-restored monster death.
11. Python bot resumes the same session after the scripted flow and asserts
    player HP and monster death survive reconnect.
12. `make ci` green.

## 8. As-Built Notes

- `DecodeStored` moved from `realtime` to `server/internal/inputdecode`, so
  `replay` no longer imports `realtime` and the hub can call `replay`.
- `replay.Reconstruct` returns the restored `*game.Sim`, snapshot, derived
  events, and `ResumeMetadata`; `Verify` and `/state` consume the same path.
- WebSocket attach checks `session_inputs`: empty history uses fresh `NewSim`;
  non-empty history uses replay reconstruction and does not call
  `LoadInventory`.
- The runner seeds duplicate detection and next sequence from historical inputs,
  with an empty post-resume buffer.
- Malformed historical input payloads fail closed before WebSocket upgrade and
  are logged with session/input/tick/message context.
- Input-without-event crash windows are replayed from the durable input stream;
  recorded events remain verification data, not the state source.
- No protocol schema bump was required.

## 9. Open questions

| # | Question | Status |
|---|----------|--------|
| 1 | Should resume compare derived events against `session_events` synchronously before accepting the WebSocket? | Resolved for v5: no blocking compare in the hot path; replay tests and `/replay` report mismatches. |
| 2 | How should persisted inputs with no recorded output be handled after a crash between input persistence and tick persistence? | Resolved for v5: deterministic replay applies recorded inputs; recorded events remain verification output. |
| 3 | Should active movement continue across disconnect based on elapsed wall time? | No for v5; replay advances only recorded sim ticks, not wall-clock time while disconnected. |

## 10. Testing plan

1. `cd server && go test ./internal/replay/...` — extended `Reconstruct`
   restores player/monster state, returns resume metadata, snapshot matches
   `Verify` report.
2. `cd server && go test ./internal/http/... -run Resume` — WebSocket resume
   snapshot carries dead monster, reduced player HP, `/state` parity, and
   duplicate rejection.
3. Dead-player resume (acceptance #6) — separate integration test with a focused
   fixture or high-HP monster setup; not part of the default bot/smoke slice flow
   (which leaves the player alive at `hp == 9`).
4. `make bot` — protocol bot completes slice, resumes same session, asserts HP
   and monster death survived reconnect (not just inventory).
5. `make client-smoke` — Godot smoke validates real resume snapshot without forced
   monster HP.
6. `make ci` — full local gate.
