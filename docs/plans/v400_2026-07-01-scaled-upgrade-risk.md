# v400 Plan — Scaled Upgrade Risk

Status: Complete
Goal: Enable logarithmic upgrade failure by item level, shard-tier success bonus, depth-only max level, and legible blacksmith tooltips.
Architecture: Data-driven failure anchors in `main_config`; Go computes authoritative effective success from target level + staged shard level; client mirrors formula for preview; pity unchanged; max level 0 disables config cap.
Tech stack: shared JSON, Go store/HTTP, GDScript blacksmith panel, Python bot client scenario.

## Baseline and shortcut decision

Builds on v197 success roll, v203 pity, v389 item level scaling, v390 leveled shards. Reuse existing upgrade transaction; no protocol schema bump (response already has `success`).

Asset: **Reject** external plugins — text tooltip only.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/main_config.v0.json` | Curve, shard bonus, max level 0 |
| Modify | `shared/rules/main_config.v0.schema.json` | New fields, max level ≥ 0 |
| Add | `shared/golden/upgrade_success_chance.json` | Formula contract cases |
| Add | `server/internal/game/item_upgrade_chance.go` | Failure/success helpers |
| Add | `server/internal/game/item_upgrade_chance_test.go` | Unit tests |
| Modify | `server/internal/game/item_roll_upgrade.go` | Depth-only max |
| Modify | `server/internal/game/rules.go` | Config struct + validation |
| Modify | `server/internal/http/account_stash.go` | Effective chance + shard resolve |
| Modify | `server/internal/store/repos.go` | Allow max level 0 |
| Add | `client/scripts/blacksmith_upgrade_chance.gd` | Client formula mirror |
| Modify | `client/scripts/blacksmith_upgrade_preview.gd` | Rich tooltip lines |
| Modify | `client/scripts/blacksmith_panel.gd` | Pass curve + staged shard |
| Modify | `client/scripts/blacksmith_panel_actions.gd` | Depth-only effective max |
| Add | `client/tests/test_blacksmith_upgrade_chance.gd` | Preview/chance unit test |
| Modify | `client/tests/test_blacksmith_panel.gd` | Updated max level expectations |
| Add | `tools/bot/scenarios/client/blacksmith_upgrade_risk.json` | Extended client proof |
| Add | `docs/as-built/v400_scaled-upgrade-risk.md` | As-built |

## Maintenance ratchet

Hotspot files touched:
- [x] `client/scripts/blacksmith_panel.gd` — stay within baseline via preview/chance extraction
- [x] `server/internal/game/rules.go` — small struct addition only

Decision:
- [x] Extract `blacksmith_upgrade_chance.gd` and `item_upgrade_chance.go`

## Task 1 — Shared rules + golden

Files:
- Modify: `shared/rules/main_config.v0.json`, `main_config.v0.schema.json`
- Add: `shared/golden/upgrade_success_chance.json`

- [x] Add failure curve anchors, shard bonus, set max level 0
- [x] Schema validation for curve arrays and max level ≥ 0

```bash
make validate-shared
```

## Task 2 — Go success formula + depth-only cap

Files:
- Add: `server/internal/game/item_upgrade_chance.go`, `item_upgrade_chance_test.go`
- Modify: `server/internal/game/item_roll_upgrade.go`, `rules.go`

- [x] Implement log10-interpolated failure curve + shard bonus + tests
- [x] `EffectiveItemUpgradeMaxLevel` treats config max 0 as depth-only

```bash
cd server && go test ./internal/game/... -run 'UpgradeFailure|UpgradeSuccess|EffectiveItemUpgradeMax' -count=1
```

## Task 3 — HTTP + store wiring

Files:
- Modify: `server/internal/http/account_stash.go`, `server/internal/store/repos.go`

- [x] Resolve staged shard level for effective chance
- [x] Allow max level 0 in store validation

```bash
cd server && go test ./internal/store/... -run 'Upgrade' -count=1
cd server && go test ./internal/http/... -run 'Upgrade' -count=1
```

## Task 4 — Client preview + unit tests

Files:
- Add: `client/scripts/blacksmith_upgrade_chance.gd`
- Modify: `blacksmith_upgrade_preview.gd`, `blacksmith_panel.gd`, `blacksmith_panel_actions.gd`
- Add: `client/tests/test_blacksmith_upgrade_chance.gd`
- Modify: `client/tests/test_blacksmith_panel.gd`, `client/scripts/client_smoke.sh`

- [x] Mirror Go formula; show base/bonus/effective/pity/failure lines
- [x] Depth-only effective max when config max is 0

```bash
godot --headless --path client --script res://tests/test_blacksmith_upgrade_chance.gd
godot --headless --path client --script res://tests/test_blacksmith_panel.gd
```

## Task 5 — Bot scenario

- [x] Add `tools/bot/scenarios/client/blacksmith_upgrade_risk.json` with `"ci_tier": "extended"`
- [x] Assert preview contains effective chance / failure lines for staged high-level item + shard

```bash
make bot-client SCENARIO=blacksmith_upgrade_risk HEADLESS=1
```

## Task 6 — Lifecycle docs

- [x] Update `PROGRESS.md`, `docs/progress/slice-lifecycle.md`, as-built

## Final verification

- [x] `make validate-shared`
- [x] Focused Go + GDScript tests above
- [x] `make maintainability`
