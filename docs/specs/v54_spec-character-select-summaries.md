# Spec: `character-select-summaries`

Status: Complete
Date: 2026-06-10
Codename: `character-select-summaries`
Slice: v54 - character select summaries

## Purpose

The character picker should show useful progression at the moment a player chooses who to play.
Today `GET /v0/characters` and the Godot `CharacterSelectPanel` expose only name, dead state, and
created date. This slice adds server-authored character summary fields from existing durable
progression so rows can show level, gold, deepest dungeon depth, and status before starting a
session.

This is a public REST response change. There is no OpenAPI/Swagger file in the repo, so the
interface contract is captured in this SDD spec plus Go HTTP/store tests and Godot client tests.

## Non-goals

- No class selection, portraits, visual customization, character model preview, or equipment preview.
- No old-session resume UI or session history in the character picker.
- No new character progression fields beyond existing durable level, gold, deepest dungeon depth,
  and dead/live status.
- No new database tables or migrations.
- No WebSocket protocol schema bump; this is an HTTP character-list response and client UI slice.
- No production menu art/audio or broader menu redesign.

## Acceptance criteria

- `GET /v0/characters` returns each account-scoped character with:
  - `level`
  - `gold`
  - `deepest_dungeon_depth`
  - existing `dead`, `name`, `character_id`, and `created_at`
- Characters with no `character_progression` row are still listed and report safe defaults:
  `level: 1`, `gold: 0`, `deepest_dungeon_depth: 0`.
- Account scoping remains unchanged: one account cannot see another account's summary values.
- Create, rename, delete, and dead-character behavior remain unchanged.
- The Godot character picker renders compact summary text for live and dead rows without disabling
  existing create, rename, delete, or select affordances.
- Dead rows remain disabled and visibly marked as not startable.
- A headless Godot client bot scenario proves the menu path sees at least one character row with the
  summary fields populated in `get_bot_state()`.
- `make ci` passes.

## Scope and likely files

- `server/internal/store/models.go` - add a summary model or extend character listing shape.
- `server/internal/store/repos.go` - list characters with left-joined progression defaults.
- `server/internal/store/interfaces.go` - expose the chosen store method shape.
- `server/internal/http/character.go` - extend `characterResponse`.
- `server/internal/http/auth_session_test.go` - HTTP response and account-scoping coverage.
- `client/scripts/character_select_panel.gd` - render summary rows and expose debug rows.
- `client/scripts/bot_scenario_runner.gd` - optional assertion support for character summary fields.
- `client/tests/test_coop_client.gd` and/or `client/tests/test_client_bot.gd` - local UI coverage.
- `tools/bot/scenarios/client/27_character_select_summaries.json` - client bot proof.
- `docs/PROGRESS.md` - lifecycle and deferred candidate updates when the slice ships.

## Test and bot proof

- Go:
  - `cd server && go test ./internal/http/... -run 'Test.*Character.*' -count=1`
  - `cd server && go test ./internal/store/... -run 'Test.*Character.*' -count=1`
- Godot:
  - `make client-unit`
  - `HEADLESS=1 make bot-client scenario=27_character_select_summaries.json`
- Final:
  - `make ci`

## Open questions and risks

- Q1: Should the picker show exact XP or stat points?
  - Default: no. Keep v54 to level, gold, deepest dungeon depth, and status only.
- Q2: Should missing progression rows be created while listing characters?
  - Default: no. Listing is read-only and uses display defaults for missing rows.
- Q3: Should deepest depth display as negative dungeon level or positive depth?
  - Default: use the existing durable `deepest_dungeon_depth` integer as-is. The UI label can say
    `Depth N`, with `0` meaning no dungeon progress yet.
- Risk: Client rows are currently compact buttons. Keep summary text short and expose a debug row
  state so headless tests can verify content without relying on pixels.
