# v244 Spec - Material Wallet Window

Status: Complete
Date: 2026-06-17
Codename: material-wallet-window

## Purpose

Give players an explicit material wallet window from the existing HUD wallet readout. The window
should show every nonzero wallet-backed resource with readable catalog-backed names, counts,
category text, and account-wide storage context without changing wallet ownership or blacksmith
spending.

## Non-goals

- No server/protocol changes, wallet persistence changes, new resource types, drop tuning,
  resource trading, stash material tabs, blacksmith recipe changes, icons, or external assets.
- No drag/drop material inventory, market material listing, multi-resource recipe authoring, or
  material exchange.
- No full HUD redesign; the compact `Shard N` readout remains the quick-scan surface.

## Client Asset / Plugin Decision

- **Adopt:** Existing in-repo `DraggableWindow` UI shell and shared `ItemRulesLoader` catalog data.
- **Borrow:** Existing `CharacterBar` wallet rows/details and client-bot wallet assertion pattern.
- **Reject:** External plugins, generated art/icons, and new asset pipeline work for this text-first
  material window.

## Acceptance Criteria

- Clicking the existing nonzero HUD material wallet readout opens a draggable `Material Wallet`
  window.
- The window lists each nonzero resource with shared-rule display name, amount, category when
  available, and an account-wide storage note.
- Empty/zero balances keep the compact HUD wallet hidden and keep the wallet window closed.
- Updating wallet state while the window is open refreshes the listed rows.
- Character-bar debug state exposes wallet-window visibility and rows for tests/bot assertions.
- A focused Godot unit test proves open, empty, and refresh behavior.
- A client bot scenario auto-picks an upgrade shard, opens the wallet window, and asserts the
  readable row.

## Scope and Likely Files

- Client: `client/scripts/material_wallet_panel.gd`, `client/scripts/character_bar.gd`.
- Client bot: `client/scripts/bot_controller.gd`, `client/scripts/bot_step_catalog.gd`,
  `client/scripts/bot_action_step_validator.gd`, `client/scripts/bot_assertion_handlers.gd`.
- Unit tests: `client/tests/test_character_bar.gd`.
- Bot/scenario: `tools/bot/scenarios/client/61_material_wallet_window.json`.
- Docs: plan, as-built, progress lifecycle.

## Test and Bot Proof

- `godot --headless --path client --script res://tests/test_character_bar.gd`
- `godot --headless --path client --script res://tests/test_client_bot.gd`
- `make bot-client scenario=61_material_wallet_window.json HEADLESS=1`
- `make maintainability`

## Open Questions and Risks

- No blocking questions. The conservative default is a text-only draggable window backed by the
  current account wallet state.
- Risk: touching bot coordinator files can worsen maintainability ratchet pressure. Keep additions
  tiny, reuse existing assertion flow, and run `make maintainability`.
