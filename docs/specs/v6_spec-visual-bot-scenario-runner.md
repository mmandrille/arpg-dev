# Spec: `visual-bot-scenario-runner`

Status: Draft
Branch: `feature/visual-bot-scenario-runner`
Related: [`docs/PROGRESS.md`](../PROGRESS.md), [`v5_spec-resume-authoritative-state.md`](v5_spec-resume-authoritative-state.md)

## 1. Purpose

`make bot-visual` should make every bot scenario visually accessible in sequence. Bot scenarios must be discoverable, run through the same authoritative protocol as the real client, record their sessions for replay verification, and then play back visually without hardcoding scenario-specific behavior in Godot.

The local loop becomes:

```text
discover scenarios -> run protocol sessions -> verify replay -> launch Godot replay playlist
```

## 2. Non-goals

- No gameplay shortcut, seeded outcome override, or direct scene mutation that bypasses protocol-derived state.
- No production replay browser or long-term artifact retention policy.
- No new gameplay content beyond the existing vertical slice scenario.
- No protocol schema bump for live WebSocket messages.

## 3. Files to create or modify

```text
docs/specs/v6_spec-visual-bot-scenario-runner.md - feature contract
docs/plans/v6_2026-06-05-visual-bot-scenario-runner.md - implementation plan
tools/bot/scenarios/*.json - declarative scenario catalog
tools/bot/run.py - scenario discovery, execution, and manifest output
tools/bot/test_protocol.py - scenario catalog tests
server/internal/replay/replay.go - replay timeline reconstruction
server/internal/http/inspect.go - debug-gated timeline endpoint
server/internal/http/replay_test.go - timeline endpoint coverage
client/scripts/main.gd - visual replay playlist mode
scripts/bot_visual.sh - run scenarios, verify replay, launch playlist
make/agents.mk - target description
docs/PROGRESS.md - slice lifecycle update when complete
```

## 4. Data shapes

### Scenario catalog

Scenario files are JSON and live under `tools/bot/scenarios/`.

```json
{
  "id": "vertical_slice",
  "title": "Vertical slice",
  "description": "Move, kill the training dummy, pick up loot, and equip it.",
  "steps": [
    {"action": "move", "direction": {"x": 1, "y": 0}, "duration_ticks": 3},
    {"action": "attack_until_event", "target_id": "1002", "event_type": "monster_killed"},
    {"action": "pick_up_first_loot"},
    {"action": "equip_first_inventory_item", "slot": "weapon"}
  ],
  "assertions": [
    "equipped_rusty_sword",
    "player_damaged",
    "monster_dead"
  ]
}
```

### Run manifest

The bot writes `.artifacts/bot-runs/<timestamp>.json`.

```json
{
  "generated_at": "2026-06-05T00:00:00Z",
  "base_url": "http://localhost:8080",
  "scenarios": [
    {
      "id": "vertical_slice",
      "title": "Vertical slice",
      "session_id": "session-id",
      "seed": "seed",
      "status": "passed",
      "replay_match": true
    }
  ]
}
```

### Debug replay timeline endpoint

`GET /v0/sessions/{session_id}/replay/timeline`

Auth:

- `Authorization: Bearer <access-token>`
- `X-Debug-Token: <debug-token>`

Response:

```json
{
  "session_id": "session-id",
  "seed": "seed",
  "envelopes": [
    {
      "type": "session_snapshot",
      "message_id": "replay-snapshot",
      "session_id": "session-id",
      "tick": 0,
      "payload": {}
    },
    {
      "type": "state_delta",
      "message_id": "replay-tick-0",
      "session_id": "session-id",
      "tick": 0,
      "payload": {
        "server_tick": 0,
        "changes": [],
        "events": []
      }
    }
  ]
}
```

The endpoint is local/debug tooling only. It re-simulates from persisted seed + inputs and emits client-facing envelopes that Godot can apply through the same snapshot/delta handlers as live networking.

## 5. Architecture and flow

```text
make bot-visual
  scripts/bot_visual.sh
    start local server
    tools.bot.run --scenario all --write-manifest ...
      discover JSON scenarios
      execute each over auth + WebSocket
      assert authoritative state
      call /replay for verification
      write run manifest
    launch Godot with ARPG_VISUAL_REPLAY_MANIFEST=<manifest>
      main.gd fetches /replay/timeline for each session
      applies envelopes at ARPG_AUTOPLAY_STEP_DELAY cadence
      advances to next scenario when timeline ends
```

## 6. Visual playback rules

- Godot must not know scenario IDs or scenario-specific gameplay branching.
- Godot consumes only the run manifest and debug replay timeline envelopes.
- Client-only animation cues may be inferred from authoritative events already present in `state_delta.events`.
- The live interactive path remains unchanged when no replay manifest is provided.
- `make bot-visual` exits the Godot client normally after the playlist completes unless
  `ARPG_VISUAL_REPLAY_EXIT_ON_COMPLETE=0` is set.

## 7. Acceptance criteria

1. `make bot` runs all discovered scenarios by default.
2. `make bot-visual` records all scenarios, verifies replay for each, and launches Godot with a replay playlist.
3. Adding a new JSON scenario under `tools/bot/scenarios/` makes it available to `make bot` and `make bot-visual`.
4. `/v0/sessions/{session_id}/replay/timeline` is auth + debug-token gated.
5. Godot visual playlist applies protocol-shaped replay envelopes through existing render handlers.
6. Godot exits cleanly after `visual replay playlist complete` when launched by `make bot-visual`.
7. Existing `make ci` remains green.

## 8. Open questions

| # | Question | Status |
|---|----------|--------|
| 1 | Should replay timelines eventually include client presentation annotations for attack wind-up/facing? | Deferred |
| 2 | Should run manifests be persisted outside local `.artifacts/`? | Deferred |

## 9. Testing plan

1. `make tools && .venv/bin/python -m pytest -q tools`
2. `cd server && go test ./internal/replay ./internal/http`
3. `make bot` against a running server
4. `make bot-visual` manual visual inspection when Godot is available
5. `make ci`
