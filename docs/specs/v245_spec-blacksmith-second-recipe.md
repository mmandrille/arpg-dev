# v245 Spec - Blacksmith Second Recipe

Status: Complete
Date: 2026-06-17
Codename: blacksmith-second-recipe

## Purpose

Turn the v238 blacksmith recipe selector into a real two-option selector. Add a server-owned
`Hone Weapon` recipe that uses the existing upgrade cost/resource tuning but only accepts weapon
templates, so players can explicitly choose a weapon-focused recipe instead of a single generic
upgrade entry.

## Non-goals

- No new material types, cost tuning fields, success formulas, crafting categories, durability,
  recipe persistence, recipe unlocks, icons, art, or external assets.
- No multi-resource recipes, success-chance changes, item-level redesign, or broad crafting window.
- No market/stash restrictions for recipe-modified items.

## Client Asset / Plugin Decision

- **Adopt:** Existing `OptionButton` recipe selector, `BlacksmithPanel`, and shared item-template
  metadata.
- **Borrow:** Existing blacksmith upgrade HTTP route, wallet spending, preview lines, and bot
  blacksmith assertions.
- **Reject:** External plugins, icons, and generated recipe art for this slice.

## Acceptance Criteria

- The blacksmith selector exposes two recipes: `Upgrade Item` and `Hone Weapon`.
- Selecting `Hone Weapon` updates debug state and staged preview text to show the selected recipe.
- `Hone Weapon` only enables eligible weapon items with damage stats; non-weapon equipment is
  disabled for that recipe in the client.
- Upgrade requests send the selected `recipe_id` to the server.
- The server accepts `item_upgrade` as the default recipe and `weapon_honing` as the second recipe,
  rejects unknown recipe IDs, and rejects non-weapon items for `weapon_honing`.
- Existing upgrade behavior and tests continue to pass for the default `Upgrade Item` recipe.
- Focused Go, Godot, and client-bot proofs cover the second recipe.

## Scope and Likely Files

- Server: `server/internal/http/account_stash.go`, store/http tests as needed.
- Client: `client/scripts/blacksmith_panel.gd`, `client/scripts/net_client.gd`,
  `client/scripts/main.gd`.
- Client bot: `client/scripts/bot_controller.gd`, `client/scripts/bot_step_catalog.gd`,
  `client/scripts/bot_action_step_validator.gd`.
- Tests: `client/tests/test_blacksmith_panel.gd`, existing shop panel regression.
- Bot/scenario: `tools/bot/scenarios/client/62_blacksmith_second_recipe.json`.
- Docs: plan, as-built, progress lifecycle.

## Test and Bot Proof

- `cd server && go test ./internal/http -run Upgrade -count=1`
- `godot --headless --path client --script res://tests/test_blacksmith_panel.gd`
- `godot --headless --path client --script res://tests/test_shop_panel.gd`
- `godot --headless --path client --script res://tests/test_client_bot.gd`
- `make bot-client scenario=62_blacksmith_second_recipe.json HEADLESS=1`
- `make maintainability`

## Open Questions and Risks

- No blocking questions. `Hone Weapon` intentionally reuses existing upgrade cost/resource tuning
  so this slice introduces no new balance constants.
- Risk: changing the HTTP request shape could affect existing clients. The server treats a missing
  or empty `recipe_id` as `item_upgrade`, and tests preserve the default route behavior.
