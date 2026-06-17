# v237 Plan - Material Wallet Details

Status: Complete
Goal: Add readable material wallet details to the existing compact HUD readout.
Architecture: Reuse the existing `resource_wallet` client state and shared item rules; do not change
server contracts or wallet ownership.
Tech stack: Godot UI/client bot, docs.

## Baseline and Asset Decision

Builds on v221 resource wallet persistence, v228 HUD wallet readout, and v229 material auto-pickup.
Asset/plugin decision: reject external assets/plugins; this is text detail in the existing
`CharacterBar`.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/character_bar.gd` | Build catalog-backed wallet detail tooltip/debug rows |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Let wallet bot assertions check tooltip/detail text |
| Modify | `client/tests/test_character_bar.gd` | Prove detail rows and hidden empty wallet state |
| Add | `tools/bot/scenarios/client/54_material_wallet_details.json` | Client proof |
| Add | `docs/as-built/v237_material-wallet-details.md` | Slice proof |

## Maintenance Ratchet

Target: touched source/test/tool files stay at or below their allowed baselines.

Hotspot / over-limit files touched:
- [x] None expected; files are below 600 lines.

Decision:
- [x] Reuse existing `assert_resource_wallet_panel` rather than adding a new bot action.
- [x] Keep details in the character bar instead of introducing a standalone wallet window.

Verification:
```bash
make maintainability
```

## Task 1 - HUD wallet details

Files:
- Modify: `client/scripts/character_bar.gd`

- [x] Resolve resource display names/categories from `ItemRulesLoader.item_definition`.
- [x] Keep compact `wallet_text` unchanged for scanability.
- [x] Populate tooltip/detail lines with name, count, category, and account-wide context.
- [x] Expose `wallet_tooltip` and `wallet_details` in debug state.

## Task 2 - Tests and bot proof

Files:
- Modify: `client/tests/test_character_bar.gd`
- Modify: `client/scripts/bot_assertion_handlers.gd`
- Add: `tools/bot/scenarios/client/54_material_wallet_details.json`

- [x] Assert `upgrade_shard` details include the shared display name, count, category, and storage
  context.
- [x] Extend resource wallet bot assertion with `tooltip_contains`.
- [x] Add a short scenario that auto-picks the shard and asserts the tooltip detail.

```bash
godot --headless --path client --script res://tests/test_character_bar.gd
make bot-client scenario=54_material_wallet_details.json HEADLESS=1
```

## Task 3 - Lifecycle docs

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `docs/progress/scenario-catalog.md`
- Modify: `docs/progress/slice-codename-index.md`
- Add: `docs/as-built/v237_material-wallet-details.md`

- [x] Record focused verification and deferred scope.

## Final Verification

- [x] `godot --headless --path client --script res://tests/test_character_bar.gd`
- [x] `make bot-client scenario=54_material_wallet_details.json HEADLESS=1`
- [x] `make maintainability`
