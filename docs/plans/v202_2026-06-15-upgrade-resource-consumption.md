# v202 Plan - Upgrade Resource Consumption

Status: Complete
Goal: Require and consume one configured upgrade shard for inventory blacksmith attempts.
Architecture: Keep `upgrade_shard` as a normal inventory item for this slice. The server remains authoritative: the inventory upgrade route checks and consumes the configured resource, while the client only previews availability and reconciles the accepted response.
Tech stack: Shared JSON rules/schema, Go HTTP/store-facing route logic, Godot blacksmith UI, client bot scenario, SDD docs.

## Baseline and shortcut decision

Builds on v180 `upgrade_shard`, v197 upgrade success chance, and v201 tooltip display. Asset/plugin decision: reject external assets/plugins; reuse the existing upgrade shard item and blacksmith panel.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/main_config.v0.json` | Configure upgrade resource item/count. |
| Modify | `shared/rules/main_config.v0.schema.json` | Validate resource config. |
| Modify | `shared/rules/worlds.v0.json` | Add an upgrade shard to the compact blacksmith client lab. |
| Modify | `server/internal/game/rules.go` | Load and validate resource config. |
| Modify | `server/internal/http/account_stash.go` | Require/consume resource on inventory upgrades and expose consumed resource in response. |
| Modify | `server/internal/http/auth_session_test.go` | Prove resource requirement and consumption. |
| Modify | `client/scripts/blacksmith_panel.gd` | Preview resource requirement and disable when missing. |
| Modify | `client/scripts/main.gd` | Remove consumed resource from local inventory after accepted upgrade. |
| Modify | `client/scripts/bot_scenario_runner.gd` | Match blacksmith resource debug expectations. |
| Modify | `client/scripts/bot_step_catalog.gd` | Validate resource expectations in blacksmith bot steps. |
| Modify | `client/tests/test_shop_panel.gd` | Unit-test blacksmith resource preview/enabling. |
| Modify | `tools/bot/scenarios/client/39_blacksmith_upgrade_ui.json` | Pick up shard and assert it is consumed. |
| Add | `docs/as-built/v202_upgrade-resource-consumption.md` | Record proof. |
| Modify | `PROGRESS.md` | Update current status after completion. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v202 lifecycle row. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [ ] `server/internal/game/game_test.go`
- [ ] `tools/bot/run.py`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: `client/scripts/bot_scenario_runner.gd`, `client/tests/test_shop_panel.gd`
- [ ] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?
- [x] Documented maintenance exception: `client/scripts/inventory_panel.gd` and `client/scripts/stash_panel.gd` were already over the old baseline on `main` after `70d5a4de fix: color unique chest item tooltips`; v202 refreshes those two baseline counts without touching the files.

Decision:
- [ ] Extract focused helper/module/test file as part of this slice, or
- [x] Defer extraction with rationale: this slice adds narrow route/UI wiring to existing blacksmith surfaces; broader splits would be unrelated maintenance.

Verification:
```bash
make maintainability
```

## Task 1 - Shared resource config

Files:
- Modify: `shared/rules/main_config.v0.json`
- Modify: `shared/rules/main_config.v0.schema.json`
- Modify: `shared/rules/worlds.v0.json`
- Modify: `server/internal/game/rules.go`

- [x] Step 1.1: Add `item_upgrade_resource_item_def_id` and `item_upgrade_resource_count`.
- [x] Step 1.2: Validate non-empty resource item id when the configured count is positive.
- [x] Step 1.3: Add one `upgrade_shard` to `vendor_lab` for focused blacksmith proof.

```bash
make validate-shared
```

## Task 2 - Server authority

Files:
- Modify: `server/internal/http/account_stash.go`
- Modify: `server/internal/http/auth_session_test.go`

- [x] Step 2.1: Reject inventory upgrades when the configured resource is missing.
- [x] Step 2.2: Consume the configured resource after an accepted inventory upgrade attempt.
- [x] Step 2.3: Include consumed resource metadata in upgrade responses.
- [x] Step 2.4: Cover success consumption and missing-resource rejection.

```bash
cd server && go test ./internal/http -run 'Upgrade' -count=1
```

## Task 3 - Client presentation and reconciliation

Files:
- Modify: `client/scripts/blacksmith_panel.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/tests/test_shop_panel.gd`

- [x] Step 3.1: Show required shard count in the blacksmith preview/debug state.
- [x] Step 3.2: Disable upgrade and show a clear status when the resource is missing.
- [x] Step 3.3: Remove consumed resource items from local inventory after accepted server response.
- [x] Step 3.4: Unit-test blacksmith resource count and enable/disable behavior.

```bash
make client-unit
```

## Task 4 - Client bot proof

Files:
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/scripts/bot_step_catalog.gd`
- Modify: `tools/bot/scenarios/client/39_blacksmith_upgrade_ui.json`

- [x] Step 4.1: Extend blacksmith bot matching for resource item/count.
- [x] Step 4.2: Update the blacksmith scenario to pick up `upgrade_shard`.
- [x] Step 4.3: Assert the shard count is one before upgrade and zero after upgrade.

```bash
SCENARIO=blacksmith_upgrade_ui HEADLESS=1 ./scripts/bot_client_local.sh
```

## Task 5 - Lifecycle docs and CI

Files:
- Add: `docs/as-built/v202_upgrade-resource-consumption.md`
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`

- [x] Step 5.1: Mark the spec complete after verification.
- [x] Step 5.2: Add as-built proof notes.
- [x] Step 5.3: Update progress status and lifecycle row.

```bash
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/http -run 'Upgrade' -count=1`
- [x] `make client-unit`
- [x] `SCENARIO=blacksmith_upgrade_ui HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `make ci`
