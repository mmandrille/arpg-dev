# v172 Spec - Loot Filter Persistence

Status: Complete
Date: 2026-06-14
Codename: `loot-filter-persistence`

## Purpose

Persist the client-side loot label rarity filter introduced in v153 so the player's chosen threshold
survives closing and reopening the Godot client. This remains display-only local preference state;
the server still owns all loot, item, and pickup behavior.

## Non-goals

- No server, protocol, store, replay, or shared-rule changes.
- No settings-panel control for the loot filter.
- No category filtering or new filter modes.
- No account-synced settings.

## Acceptance criteria

1. `ClientSettings` reads and writes a `loot_filter_mode` field in `user://settings.json`.
2. Missing, malformed, or unknown saved values normalize to `All`.
3. `LootLabelFilter` can be initialized from a saved mode label without cycling through modes.
4. Pressing the existing loot-filter cycle input saves the new mode through `ClientSettings`.
5. The top-right HUD still only shows the loot filter line when the restored mode is active.
6. Focused Godot unit tests cover settings parsing/saving and filter mode restore behavior.
7. `make client-unit`, `make maintainability`, and `make ci` pass.

## Scope and likely files

- `client/scripts/client_settings.gd`
- `client/scripts/loot_label_filter.gd`
- `client/scripts/main.gd`
- `client/tests/test_client_bot.gd`
- `client/tests/test_loot_label_filter.gd`
- `docs/as-built/v172_loot-filter-persistence.md`
- `PROGRESS.md`

## Test and bot proof

```bash
make client-unit
make maintainability
make ci
```

Manual visual verification:

```bash
make bot-visual scenario=01_click_to_kill.json
```

## Open questions and risks

- No blocking questions. The slice deliberately keeps persistence local and display-only.
