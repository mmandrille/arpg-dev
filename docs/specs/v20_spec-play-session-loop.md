# Spec: `play-session-loop`

Status: Approved for implementation (gaps closed 2026-06-06)
Branch: `feature/play-session-loop`
Slice: v20 — default playable `dungeon_levels` world with town, stairs, teleporters, and unlimited session-only descent
Baseline: v19 `teleporters-and-waypoint-ui`
Related:

- [`v18_spec-dungeon-levels-and-stairs.md`](v18_spec-dungeon-levels-and-stairs.md)
- [`v19_spec-teleporters-and-waypoint-ui.md`](v19_spec-teleporters-and-waypoint-ui.md)
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) — D2 level 0 = town; D4 town waypoint always active
- [`../../PROGRESS.md`](../../PROGRESS.md)

Review status: Approved after gap review against v18/v19 as-built code. Closed gaps:

- Town is a **static** level `0` from `worlds.v0.json`, not PCG output.
- Multi-level sessions **no longer pre-generate level -1** at create; dungeon floors remain lazy on first visit.
- Protocol v1 allows **`target_level: 0`** and town rows in `discovered_teleporters`.
- Client wall rendering is **level-aware**: dungeon perimeter only when `current_level < 0`.
- Bot scenarios `12` / `13`, golden descend/ascend fixture, `play.sh`, and focused tests are in scope.

## 1. Purpose

Make `make play` start as an actual game loop rather than a lab scenario:

- The default interactive client creates a fresh `dungeon_levels` world session.
- The player starts in an empty town at level `0`.
- Town contains only a down stair and a teleporter.
- The town teleporter is enabled from the start.
- Descending from town generates dungeon level `-1`; each dungeon level has an up stair, down stair,
  and teleporter.
- The player can descend forever. Levels are generated on first visit from the session seed.
- Closing the game loses the run. No character persistence or durable progression is required.

Existing stored session/replay infrastructure may still exist for development and CI, but v20 does
not require or expose player resume for `make play`. Character and session persistence are deferred
to a future slice.

## 2. Non-goals

- No character-scoped inventory, waypoint, quest, or campaign persistence.
- No town NPCs, vendors, stash, combat safe-zone rules, or production town art.
- No monster or loot density generation beyond existing generated dungeon placeholders.
- No deletion of the existing resume/replay architecture.
- No player-facing resume flow for `make play`.
- No town perimeter art or production hub layout — movement bounds only.

## 3. Files to create or modify

```text
docs/specs/v20_spec-play-session-loop.md              - this slice contract
docs/plans/v20_2026-06-06-play-session-loop.md        - implementation plan
shared/rules/worlds.v0.json                           - town level 0 preset entities
shared/protocol/messages.v1.schema.json               - allow teleport target_level 0
shared/protocol/session_snapshot.v1.schema.json       - allow town teleporter discovery row
shared/protocol/state_delta.v1.schema.json            - allow town teleporter discovery update
shared/golden/dungeon_stairs.json                     - descend_then_ascend via town (0 -> -1 -> -2 -> -1 -> 0)
server/internal/game/sim.go                           - town bootstrap, entry level 0, travel to town
server/internal/game/game_test.go                     - play loop transition/teleporter coverage
tools/bot/scenarios/12_dungeon_levels.json            - start in town, descend before existing steps
tools/bot/scenarios/13_teleporter_lab.json            - start in town, descend before teleporter lab
scripts/play.sh                                       - default ARPG_WORLD_ID=dungeon_levels, fresh session
client/scripts/main.gd                                - default interactive world + town wall rendering
PROGRESS.md                                      - lifecycle update when v20 ships
```

## 4. Data shapes

### World preset

`shared/rules/worlds.v0.json` changes `dungeon_levels` from a level `-1` lab entry into the
default playable multi-level world with town level `0`:

```json
{
  "dungeon_levels": {
    "mode": "multi_level",
    "player": { "position": { "x": 4, "y": 10 } },
    "entities": [
      { "type": "interactable", "interactable_def_id": "stairs_down", "position": { "x": 8, "y": 10 } },
      { "type": "interactable", "interactable_def_id": "teleporter", "position": { "x": 4, "y": 13 } }
    ]
  }
}
```

Town uses **`navigation.v0.json`** bounds (same ~16×10 cage as legacy worlds), not
`dungeon_generation` 32×20 perimeter. Town has **no static wall entities** in v20.

### Protocol

Protocol v1 keeps the existing `teleport_intent { target_level }` shape. Relax constraints so
`target_level` and `discovered_teleporters[].level` allow **`0`** (town) in addition to negative
dungeon levels.

The initial snapshot for `dungeon_levels` includes `{ "level": 0, "discovered": true }` in
`discovered_teleporters`. Dungeon levels appear in the list as they are generated/visited (unchanged
v19 behavior).

### Sim entry and lazy generation

| Phase | `current_level` | Level contents |
|-------|-----------------|----------------|
| Session create | `0` | Town from world preset: player, `stairs_down`, `teleporter`; `discoveredTeleporters[0] = true` |
| First `descend_intent` from town | `-1` | Generated dungeon floor (stairs + teleporter + level -1 loot rules) |
| Further `descend_intent` | `-N` | Lazy `ensureDungeonLevel` (unchanged) |
| `ascend_intent` from `-1` | `0` | Town static level; player arrives at town `stairs_down` cell |
| `teleport_intent` to `0` | `0` | Requires current-floor teleporter discovered; player arrives at town teleporter cell |

**Constants:** replace multi-level `entryLevel = -1` with **`townLevel = 0`**. Reject
`ascend_intent` when `current_level >= 0` (`already_at_entry`). Reject `descend_intent` when
`current_level` has no reachable down stair.

Dungeon travel helpers (`ensureDungeonLevel`, `GenerateDungeonLevel`) remain **negative levels
only**. Town is created once from the world preset via a dedicated bootstrap path (reuse the
existing single-level entity spawn loop for preset interactables).

## 5. Architecture and flow

```text
make play
  -> play.sh exports ARPG_WORLD_ID=dungeon_levels and unsets ARPG_SESSION_ID
  -> Godot client requests world_id=dungeon_levels unless ARPG_WORLD_ID overrides it
  -> server creates fresh solo session in level 0 (town only; level -1 not generated yet)
  -> town teleporter is already discovered/enabled
  -> player clicks town stair_down
  -> descend_intent moves to generated level -1 at stairs_up
  -> player can use stairs or discover teleporters
  -> each further descend lazily generates current_level - 1
  -> ascend from -1 returns to town at the town down stair
  -> teleport to town (level 0) is allowed once the current dungeon teleporter is discovered
  -> process exit/window close discards the run from a player-facing perspective
```

`dungeon_levels` is now the playable world. Bot scenarios `12_dungeon_levels` and
`13_teleporter_lab` prepend a town descend step and update final assertions where needed.

**Client presentation:**

- Default `current_world_id` for interactive (non-bot) play is `dungeon_levels` when
  `ARPG_WORLD_ID` is unset.
- `_render_world_walls`: for `dungeon_levels`, render 32×20 dungeon perimeter **only when
  `current_level < 0`**; at town (`current_level == 0`) render **no walls** (open placeholder hub).
- Level HUD remains hidden at `current_level == 0` (existing v18 behavior).

**Resume/replay (dev/CI only):** unchanged architecture. Reconnect and replay must still
reconstruct town, visited dungeon levels, `current_level`, and `discovered_teleporters` including
level `0`. v20 simply does not expose resume in `make play`.

## 6. Acceptance criteria

1. Fresh `dungeon_levels` sessions start at `current_level: 0` with player, `stairs_down`, and
   `teleporter` from the world preset. Level `-1` is **not** present until first descend.
2. From town, `descend_intent` transitions to level `-1` and generates up/down stairs plus a
   teleporter.
3. From level `-1`, `ascend_intent` returns to town at the town `stairs_down` position
   (`{8, 10}`).
4. From dungeon floors, repeated `descend_intent` can generate at least levels `-1`, `-2`, and
   `-3` without special-case caps.
5. Town teleporter discovery appears in snapshots as level `0` with `discovered: true` from session
   attach.
6. Teleporting back to town (`target_level: 0`) is accepted once the current dungeon floor
   teleporter has been discovered and the player is in range of it.
7. `make play` creates a new session by default (`ARPG_SESSION_ID` unset); player-facing resume is
   not required in v20.
8. Interactive Godot defaults to `dungeon_levels`; bot/client scenarios can still override
   `ARPG_WORLD_ID`.
9. `make bot` scenarios `01`–`13` pass after scenario updates; `make ci` green.
10. Reconnect resume and replay for `dungeon_levels` still reconstruct town + dungeon state (dev/CI
    regression guard).

## 7. Testing plan

1. `make validate-shared`
2. `cd server && go test ./internal/game/... -run 'Play|Dungeon|Teleport|Town'`
3. `make client-unit`
4. `make bot` (scenarios `12`, `13` cover town entry)
5. `make ci` final gate
6. Manual: `make play` — confirm town start, descend, ascend back to town, waypoint panel shows
   town as discovered

## 8. Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | `dungeon_levels` becomes the default playable world. | Avoid a second near-identical world preset and evolve the existing dungeon slice into play. |
| 2 | "Without persistence" means no player-facing resume requirement for now. | Existing technical persistence may remain for replay/dev tooling; character/session persistence is a future slice. |
| 3 | Town level `0` has a ground teleporter that starts enabled. | ADR D4: town is always an active waypoint hub within the session. |
| 4 | `make play` sets `ARPG_WORLD_ID=dungeon_levels` and clears `ARPG_SESSION_ID`. | Keeps the first playable loop simple without changing global client default for bot/smoke paths that omit the env var. |
| 5 | Town is static from `worlds.v0.json`; dungeon floors stay lazy PCG. | Matches ADR D2/D3 split between hub layout and seeded dungeon generation. |
| 6 | Town has no walls; dungeon perimeter renders only below level `0`. | Avoids showing a 32×20 dungeon cage in the hub placeholder. |
| 7 | Bot scenarios `12`/`13` gain a town-descend preamble instead of a new scenario file. | Preserves CI coverage IDs while updating entry behavior. |

## 9. Open questions

| # | Question | Resolution |
|---|----------|------------|
| 1 | Should `main.gd` default to `dungeon_levels` globally, or only via `play.sh`? | **`play.sh` only** for v20 — avoids surprising bot-visual/smoke paths that expect `vertical_slice` when `ARPG_WORLD_ID` is unset. Interactive default in `main.gd` still switches to `dungeon_levels` when env is empty and not in bot-client mode. |
| 2 | Should bot scenario 12 end back in town after ascending from `-1`? | **No** — scenario keeps its loot + return-to-`-1` proof; acceptance #3 is covered by a dedicated Go unit test and manual `make play`. Scenario 12 only adds the initial town → `-1` descend step. |
