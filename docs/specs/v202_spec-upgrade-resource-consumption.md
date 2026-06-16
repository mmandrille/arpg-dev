# v202 Spec: Upgrade Resource Consumption

Status: Complete
Date: 2026-06-15
Codename: upgrade-resource-consumption

## Purpose

Make blacksmith item upgrades spend the upgrade resource introduced in v180. A player must carry an
`upgrade_shard` when upgrading an inventory item, and the server consumes one shard per attempt in
addition to the existing gold cost.

## Non-goals

- No resource wallet, stash material storage, or account-wide material balance.
- No per-rarity recipe curves, multi-resource recipes, market restrictions, or binding rules.
- No change to direct/internal stash-item upgrade API behavior outside the inventory blacksmith path.
- No new art assets or plugins.

## Acceptance Criteria

- `shared/rules/main_config.v0.json` declares the upgrade resource item and count.
- Rules loading and shared validation reject invalid upgrade resource config.
- Inventory blacksmith upgrades require at least one configured resource item in character inventory.
- Successful and failed upgrade attempts consume the configured resource after the server accepts the attempt.
- Client blacksmith preview and debug state show the required resource and disable upgrade when missing it.
- Client state removes the consumed resource after an accepted upgrade response.
- A client bot proof picks up an `upgrade_shard`, upgrades an item, and observes the shard count drop to zero.
- `make validate-shared`, focused Go tests, `make client-unit`, focused client bot proof, and `make ci` pass.

## Scope and Files Likely Touched

- `shared/rules/main_config.v0.json`
- `shared/rules/main_config.v0.schema.json`
- `shared/rules/worlds.v0.json`
- `server/internal/game/rules.go`
- `server/internal/http/account_stash.go`
- `server/internal/http/auth_session_test.go`
- `client/scripts/blacksmith_panel.gd`
- `client/scripts/main.gd`
- `client/scripts/bot_scenario_runner.gd`
- `client/scripts/bot_step_catalog.gd`
- `client/tests/test_shop_panel.gd`
- `tools/bot/scenarios/client/39_blacksmith_upgrade_ui.json`

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/http -run 'Upgrade' -count=1`
- `make client-unit`
- `SCENARIO=blacksmith_upgrade_ui HEADLESS=1 ./scripts/bot_client_local.sh`
- `make ci`

## Open Questions and Risks

- Risk: consuming inventory resources is not yet an account material wallet. This is intentional for
  the thin slice; stash materials remain deferred.
- Risk: direct stash-upgrade API callers do not provide a character inventory to consume from. This
  slice keeps that path legacy/internal and gates the player-facing blacksmith inventory route.
- Asset/plugin decision: rejected. The existing upgrade shard item presentation is reused.
