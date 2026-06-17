# v245 Plan - Blacksmith Second Recipe

Status: Complete
Goal: Add a server-owned `Hone Weapon` recipe as the second blacksmith selector option.
Architecture: Keep the existing item-upgrade route and store mutation path, but route requests
through a recipe ID. `item_upgrade` remains the default. `weapon_honing` reuses existing
gold/resource/chance tuning and narrows server eligibility to weapon templates with damage stats.
Tech stack: Go HTTP/store boundary, Godot UI/client bot, docs.

## Baseline and Asset Decision

Builds on v221 wallet-backed upgrades, v222 previews, and v238 recipe selector.

Asset/plugin decision:
- Adopt the existing `OptionButton`, `BlacksmithPanel`, and shared item-template data.
- Borrow the existing upgrade route and bot blacksmith preview assertions.
- Reject external assets/plugins and recipe icons.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/http/account_stash.go` | Decode/validate recipe IDs and choose server eligibility |
| Modify | `server/internal/http/auth_session_test.go` | Prove weapon recipe accept/reject/default behavior |
| Modify | `client/scripts/net_client.gd` | Send selected recipe ID with upgrade requests |
| Modify | `client/scripts/main.gd` | Pass selected panel recipe into the client call |
| Modify | `client/scripts/blacksmith_panel.gd` | Add second selector option, eligibility, preview/debug state |
| Modify | `client/tests/test_blacksmith_panel.gd` | Prove second recipe selection and non-weapon disablement |
| Modify | `client/tests/test_shop_panel.gd` | Preserve default blacksmith flow |
| Modify | `client/scripts/bot_controller.gd` | Add recipe selection bot action |
| Modify | `client/scripts/bot_step_catalog.gd` | Register recipe selection action |
| Modify | `client/scripts/bot_action_step_validator.gd` | Validate recipe selection action |
| Add | `tools/bot/scenarios/client/62_blacksmith_second_recipe.json` | Client proof |
| Add | `docs/as-built/v245_blacksmith-second-recipe.md` | As-built proof |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `client/scripts/bot_controller.gd`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected

Decision:
- [x] Keep `main.gd` to same-line call-shape edits only.
- [x] Keep `bot_controller.gd` to one tiny dispatch/action helper.
- [x] Keep `blacksmith_panel.gd` below 600 lines; extract only if the panel grows beyond target.

Verification:
```bash
make maintainability
```

## Task 1 - Server recipe authority

Files:
- Modify: `server/internal/http/account_stash.go`
- Modify: `server/internal/http/auth_session_test.go`

- [x] Add optional `recipe_id` request decoding with default `item_upgrade`.
- [x] Add `weapon_honing` validation and server eligibility for weapon templates with damage stats.
- [x] Reject unknown recipe IDs and reject non-weapons for `weapon_honing`.
- [x] Preserve existing default upgrade behavior.

```bash
cd server && go test ./internal/http -run Upgrade -count=1
```

## Task 2 - Client selector and request wiring

Files:
- Modify: `client/scripts/net_client.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/blacksmith_panel.gd`
- Modify: `client/tests/test_blacksmith_panel.gd`
- Modify: `client/tests/test_shop_panel.gd`

- [x] Add `Hone Weapon` as the second recipe option.
- [x] Update preview/debug state when the selected recipe changes.
- [x] Disable non-weapons for `weapon_honing` and show clear preview/status text.
- [x] Send selected `recipe_id` through inventory/stash upgrade requests.
- [x] Preserve default `Upgrade Item` tests.

```bash
godot --headless --path client --script res://tests/test_blacksmith_panel.gd
godot --headless --path client --script res://tests/test_shop_panel.gd
```

## Task 3 - Bot proof

Files:
- Modify: `client/scripts/bot_controller.gd`
- Modify: `client/scripts/bot_step_catalog.gd`
- Modify: `client/scripts/bot_action_step_validator.gd`
- Add: `tools/bot/scenarios/client/62_blacksmith_second_recipe.json`

- [x] Add `select_blacksmith_recipe` client-bot action.
- [x] Add a scenario that selects `weapon_honing`, stages a weapon, verifies preview text, and
  completes an upgrade.

```bash
godot --headless --path client --script res://tests/test_client_bot.gd
make bot-client scenario=62_blacksmith_second_recipe.json HEADLESS=1
```

## Task 4 - Lifecycle docs

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v245_blacksmith-second-recipe.md`

- [x] Record focused verification and deferred scope.

## Final Verification

- [x] `cd server && go test ./internal/http -run Upgrade -count=1`
- [x] `godot --headless --path client --script res://tests/test_blacksmith_panel.gd`
- [x] `godot --headless --path client --script res://tests/test_shop_panel.gd`
- [x] `godot --headless --path client --script res://tests/test_client_bot.gd`
- [x] `make bot-client scenario=62_blacksmith_second_recipe.json HEADLESS=1`
- [x] `make maintainability`
