# v118 Plan — Blacksmith Upgrade UI

Status: Complete
Goal: Add a town blacksmith NPC and client path for account-stash item upgrades.
Architecture: The existing HTTP/store upgrade route remains authoritative. Shared rules add a
`town_blacksmith` service interactable so town state remains server-authored. The Godot client renders
the blacksmith and opens an upgrade UI that calls the HTTP route, then refreshes local stash state from
the response.
Tech stack: Shared JSON rules, Go rule validation only, Godot client UI/networking, Godot client bot,
SDD docs.

## Baseline and shortcut decision

HTTP client, and town NPC presentation patterns.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/interactables.v0.json` | Add `town_blacksmith` service metadata. |
| Modify | `shared/rules/worlds.v0.json` | Place blacksmith in town and vendor lab. |
| Modify | `server/internal/game/rules.go` | Allow `blacksmith` service validation. |
| Modify | `client/scripts/net_client.gd` | Add account-stash item upgrade HTTP helper. |
| Create | `client/scripts/blacksmith_panel.gd` | Focused upgrade UI/debug/actions. |
| Modify | `client/scripts/main.gd` | Render/open blacksmith and apply upgrade responses. |
| Modify | `client/scripts/bot_controller.gd` | Add blacksmith bot action forwarding. |
| Modify | `client/scripts/bot_scenario_runner.gd` | Add blacksmith wait/assert/action step support. |
| Create | `tools/bot/scenarios/client/39_blacksmith_upgrade_ui.json` | Focused client bot proof. |
| Create | `docs/as-built/v118_blacksmith-upgrade-ui.md` | As-built summary. |
| Modify | `PROGRESS.md` | Lifecycle and open-gaps close-out. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `server/internal/game/sim.go`
- [ ] `tools/bot/run.py`
- [ ] `tools/validate_shared.py`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: `client/scripts/bot_scenario_runner.gd`, `server/internal/store/store_test.go`

Decision:
- [x] Extract focused helper/module/test file as part of this slice: new `blacksmith_panel.gd`.
- [x] Defer extraction with rationale: the remaining growth is narrow service dispatch/UI-bot plumbing
  in existing integration hotspots plus focused store coverage for a real generated-item payload bug
  found by the bot proof. A larger extraction would exceed the slice scope, so v118 records a
  maintenance exception and updates the grandfathered baseline for those touched files.

Verification:
```bash
make maintainability
```

## Task 1 — Shared Blacksmith Service

Files:
- Modify: `shared/rules/interactables.v0.json`
- Modify: `shared/rules/worlds.v0.json`
- Modify: `server/internal/game/rules.go`

- [x] Step 1.1: Add `town_blacksmith` with service `blacksmith`.
- [x] Step 1.2: Place the blacksmith in `dungeon_levels` town and `vendor_lab`.
- [x] Step 1.3: Allow service validation without adding gameplay authority.

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'TestLoadRules|TestDefaultRules'
```

## Task 2 — Client Upgrade UI

Files:
- Modify: `client/scripts/net_client.gd`
- Create: `client/scripts/blacksmith_panel.gd`
- Modify: `client/scripts/main.gd`

- [x] Step 2.1: Add `upgrade_account_stash_item(stash_item_id)` to `net_client.gd`.
- [x] Step 2.2: Implement blacksmith panel rows for stash item name, level, cost, gold, and upgrade action.
- [x] Step 2.3: Render `town_blacksmith` as a distinct code-native NPC and open the panel on interaction.
- [x] Step 2.4: Apply upgrade response by updating the matching `stash_items` row and `stash_gold`.

```bash
make client-unit
```

## Task 3 — Client Bot Proof

Files:
- Modify: `client/scripts/bot_controller.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Create: `tools/bot/scenarios/client/39_blacksmith_upgrade_ui.json`

- [x] Step 3.1: Add bot actions/assertions for waiting on blacksmith panel and clicking one upgrade.
- [x] Step 3.2: Add scenario that sells/deposits enough gold, deposits a rolled item to stash, opens the blacksmith, upgrades once, and asserts level/gold changes.

```bash
make bot-client scenario=39_blacksmith_upgrade_ui
```

## Task 4 — Lifecycle Docs and CI

Files:
- Modify: `docs/specs/v118_spec-blacksmith-upgrade-ui.md`
- Modify: `docs/plans/v118_2026-06-13-blacksmith-upgrade-ui.md`
- Create: `docs/as-built/v118_blacksmith-upgrade-ui.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Mark the spec and plan complete after verification.
- [x] Step 4.2: Update `PROGRESS.md` lifecycle/current status and deferred scope.
- [x] Step 4.3: Write the as-built summary.

```bash
make maintainability
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `make client-unit`
- [x] `make bot-client scenario=39_blacksmith_upgrade_ui`
- [x] `make ci`

## Deferred scope

Crafting resources, upgrade failure/success chance, recipe tiers, market restrictions for upgraded
items, production blacksmith art/audio, and inventory/equipped item upgrade paths remain deferred.
