# v54 Plan - Character Select Summaries

Status: Complete
Goal: Show server-authored character progression summaries in the Godot character picker.
Architecture: The HTTP character list remains account-scoped and read-only. The store left-joins
existing durable `character_progression` rows and returns safe display defaults when progression is
missing, without creating rows during listing. Godot renders the summary as presentation-only text
and keeps start/rename/delete behavior unchanged.
Tech stack: Go store + HTTP API, Godot client UI/tests, Godot client bot scenario, docs.

## Baseline and shortcut decision

Builds on v24/v45 character selection, v26/v39 durable level/gold/deepest-depth progression, and
v53 client bot coverage style. `GET /v0/characters` is the affected public surface; no OpenAPI file
exists in this repo, so the SDD spec plus Go tests are the contract.

Godot plugin shortcut decision: **reject external plugin adoption for v54 implementation**. The
adoption checklist in `docs/researchs/godot-plugins-and-shortcuts.md` was reviewed. This is a small
extension of an existing in-repo `Control` and does not justify adding GLoot, Godot-Inventory, or a
menu/UI dependency.

The full autoloop idea menu is persisted in `PROGRESS.md` under curated autoloop candidates so
future autoloops can reuse it.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/store/models.go` | Add a character summary model including durable progression display fields. |
| Modify | `server/internal/store/interfaces.go` | Return the summary model from character listing. |
| Modify | `server/internal/store/repos.go` | Left-join progression and coalesce defaults while preserving account scope/order. |
| Modify | `server/internal/http/character.go` | Extend character list/create/rename response shape with summary fields. |
| Modify | `server/internal/http/auth_session_test.go` | Prove response fields, defaults, and account scoping. |
| Modify | `client/scripts/character_select_panel.gd` | Render summary text and expose row debug summaries. |
| Modify | `client/scripts/bot_scenario_runner.gd` | Add optional character summary assertions. |
| Modify | `client/tests/test_coop_client.gd` | Unit coverage for summary rendering and dead row behavior. |
| Modify | `client/tests/test_client_bot.gd` | Unit coverage for new bot assertion shape. |
| Add | `tools/bot/scenarios/client/27_character_select_summaries.json` | Headless menu proof for summary fields. |
| Modify | `PROGRESS.md` | Lifecycle row, summary, and curated candidate carry-forward when the slice ships. |

## Task 1 - Store and HTTP contract

Files:
- Modify: `server/internal/store/models.go`
- Modify: `server/internal/store/interfaces.go`
- Modify: `server/internal/store/repos.go`
- Modify: `server/internal/http/character.go`
- Modify: `server/internal/http/auth_session_test.go`

- [x] Step 1.1: Add a `CharacterSummary` store model with `Level`, `Gold`, and
  `DeepestDungeonDepth` alongside existing character identity fields.
- [x] Step 1.2: Change `ListCharacters` to return summaries from a `LEFT JOIN character_progression`
  query, coalescing missing progression to `level=1`, `gold=0`, and `deepest_dungeon_depth=0`.
- [x] Step 1.3: Extend `characterResponse` and `characterToResponse` to include the new fields for
  list, create, and rename responses. Create/rename should use default summary values unless the
  caller already loaded a durable progression row.
- [x] Step 1.4: Add HTTP tests proving listed characters include exact progression values after a
  session/progression update and safe defaults before progression exists.
- [x] Step 1.5: Add/adjust account-scoping tests so another account cannot see summary values.

```bash
cd server && go test ./internal/http/... -run 'Test.*Character.*' -count=1
cd server && go test ./internal/store/... -run 'Test.*Character.*' -count=1
```

## Task 2 - Godot character picker presentation

Files:
- Modify: `client/scripts/character_select_panel.gd`
- Modify: `client/tests/test_coop_client.gd`

- [x] Step 2.1: Render each row with compact summary text for status, level, gold, and depth while
  keeping the button/select behavior unchanged.
- [x] Step 2.2: Expose debug row state from `get_debug_state()` so bot/unit tests can assert summary
  fields and label text without pixel checks.
- [x] Step 2.3: Update dead-row unit coverage to prove the summary remains visible while the select
  action stays disabled.

```bash
make client-unit
```

## Task 3 - Client bot proof

Files:
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/tests/test_client_bot.gd`
- Add: `tools/bot/scenarios/client/27_character_select_summaries.json`

- [x] Step 3.1: Add optional `assert_character_panel` filters for character summary fields such as
  `min_level`, `min_gold`, `min_deepest_dungeon_depth`, or exact values when useful.
- [x] Step 3.2: Add unit coverage for the new assertion fields.
- [x] Step 3.3: Add a client bot scenario that enters the menu, opens Create Game / Choose
  Character, and asserts a row with level/gold/deepest-depth fields in `get_bot_state()`.

```bash
make client-unit
HEADLESS=1 make bot-client scenario=27_character_select_summaries.json
```

## Task 4 - Lifecycle docs and CI

Files:
- Modify: `docs/specs/v54_spec-character-select-summaries.md`
- Modify: `docs/plans/v54_2026-06-10-character-select-summaries.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Mark this plan complete as tasks finish.
- [x] Step 4.2: Update `PROGRESS.md`: latest completed slice v54, numbering note, lifecycle
  row, "What each slice proved", scripted scenario catalog, recently closed item, and keep the
  remaining curated autoloop candidates for reuse.
- [x] Step 4.3: Record any deferred scope: portraits, class summaries, equipment previews,
  old-session resume UI, and richer character detail panels remain out of v54.

```bash
rg -n 'v54|character-select-summaries|Curated autoloop candidates|Latest completed slice' PROGRESS.md
make ci
```

## Final verification

- [x] `cd server && go test ./internal/http/... -run 'Test.*Character.*' -count=1`
- [x] `cd server && go test ./internal/store/... -run 'Test.*Character.*' -count=1`
- [x] `make client-unit`
- [x] `HEADLESS=1 make bot-client scenario=27_character_select_summaries.json`
- [x] `make ci`

## Deferred scope

- Character portraits, classes, visual customization, equipment previews, old-session resume UI,
  and richer character detail panels.
- The other curated autoloop candidates remain in `PROGRESS.md` for future selection.
