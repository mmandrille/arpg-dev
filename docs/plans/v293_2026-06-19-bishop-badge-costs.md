# v293 Plan — Bishop Badge Costs

Status: Complete
Goal: Make bishop respec and revive-all consume badge wallet resources instead of being free.
Architecture: Shared main config owns the two bishop resource costs. The Go sim validates and
consumes wallet resources through a focused helper, emits `resource_wallet_update`, and keeps
existing bishop events with added resource metadata. The client bishop panel receives authoritative
service open state plus the current wallet and disables each action by its own badge requirement.
Tech stack: Shared rule catalogs, Go sim/service tests, Godot bishop panel tests, protocol/client bot
scenarios, SDD docs.

## Baseline and shortcut decision

Builds on v213 bishop revive-all, v221 resource wallet, v237/v244 wallet UI details/window, v291
quest town turn-in, and v292 badge reward supply.

Asset/plugin decision:

- Adopt: existing account `resource_wallet`, badge item definitions, bishop service intents, and
  bishop panel.
- Borrow: v292 depth-125 quest turn-in reward flow for deterministic bot badge acquisition.
- Reject: external assets/plugins, production badge icons, new protocol schema version, and a new
  dedicated badge inventory UI.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/main_config.v0.json` | Add bishop respec/revive badge cost rows. |
| Modify | `shared/rules/main_config.v0.schema.json` | Validate bishop badge cost shape. |
| Modify | `tools/validate_main_config.py` | Validate non-negative costs and required resource IDs. |
| Modify | `server/internal/game/rules.go` | Add typed bishop cost fields without growing the file beyond baseline. |
| Modify | `server/internal/game/main_config_validation.go` | Validate bishop resource IDs as non-equippable currency. |
| Create | `server/internal/game/bishop_respec.go` | Move and update bishop respec handling out of `handlers.go`. |
| Modify | `server/internal/game/handlers.go` | Remove the respec handler body after extraction. |
| Modify | `server/internal/game/bishop_revive.go` | Consume resurrection badge on revive-all. |
| Create | `server/internal/game/bishop_costs.go` | Shared bishop wallet affordability/consume helpers. |
| Modify | `server/internal/game/bishop_test.go` | Cover missing/consumed respec and resurrection badges. |
| Modify | `server/internal/game/types.go` | Add resource-cost metadata field if needed for bishop events. |
| Modify | `client/scripts/bishop_panel.gd` | Show and gate respec/revive by badge wallet balances. |
| Modify | `client/scripts/main.gd` | Pass wallet/config into bishop panel and refresh after wallet changes. |
| Create/modify | `client/tests/test_bishop_panel.gd` | Cover badge label/enabled state and post-consumption update. |
| Modify | `client/scripts/bot_scenario_runner.gd` and action catalog only if needed | Let client bot assert badge-cost panel state. |
| Modify | `tools/bot/runtime_queries.py` only if needed | Let protocol bot assert resource-cost event fields. |
| Modify | `tools/bot/scenarios/45_town_bishop_respec.json` | Acquire badges, respec, and prove wallet consumption. |
| Modify | `tools/bot/scenarios/client/32_town_bishop_respec_panel.json` | Acquire badges, open panel, prove action affordability/cost labels. |
| Create during finish | `docs/as-built/v293_bishop-badge-costs.md` | Record proof and deferred resurrection implementation. |
| Modify during finish | `PROGRESS.md`, `docs/progress/slice-lifecycle.md`, `docs/progress/slice-codename-index.md` | Lifecycle updates. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines. Grandfathered files must not grow beyond
their current allowance.

Hotspot / over-limit files touched:

- [x] `server/internal/game/handlers.go` is grandfathered (`1430` baseline, `1430` current); move
  `handleBishopRespec` into `bishop_respec.go` so the file shrinks.
- [x] `server/internal/game/rules.go` is grandfathered (`3303` baseline, `3298` current); add only
  typed config fields and keep it below baseline.
- [x] `client/scripts/bot_scenario_runner.gd` is grandfathered (`1629` baseline, `1647` current);
  avoid touching unless panel assertions cannot cover v293.
- [x] Avoid `tools/bot/run.py`; it is already close to its allowance.
- [x] New Go/GDScript/test files must stay under 600 lines.

Decision:

- [x] Put cost logic in new bishop/resource helper files and use existing bot actions where possible.

Verification:

```bash
make maintainability
```

## Task 1 — Shared bishop badge costs

Files:

- Modify: `shared/rules/main_config.v0.json`
- Modify: `shared/rules/main_config.v0.schema.json`
- Modify: `tools/validate_main_config.py`
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/main_config_validation.go`

- [x] Step 1.1: Add `bishop_respec_resource_item_def_id`, `bishop_respec_resource_count`,
  `bishop_revive_resource_item_def_id`, and `bishop_revive_resource_count` to shared config.
- [x] Step 1.2: Set defaults to `respec_badge` x1 and `resurrection_badge` x1.
- [x] Step 1.3: Validate non-negative counts, required IDs when count is positive, known item IDs,
  and non-equippable currency category.

Verify:

```bash
make validate-shared
```

## Task 2 — Server bishop consumption

Files:

- Create: `server/internal/game/bishop_costs.go`
- Create: `server/internal/game/bishop_respec.go`
- Modify: `server/internal/game/handlers.go`
- Modify: `server/internal/game/bishop_revive.go`
- Modify: `server/internal/game/bishop_test.go`
- Modify if needed: `server/internal/game/types.go`

- [x] Step 2.1: Add helpers to read configured bishop costs, check wallet affordability, and consume
  wallet resources with `resource_wallet_update`.
- [x] Step 2.2: Extract `handleBishopRespec` from `handlers.go`, then require and consume
  `respec_badge` before resetting the build.
- [x] Step 2.3: Require and consume `resurrection_badge` in `handleBishopReviveAll` while preserving
  the existing revived-count event semantics.
- [x] Step 2.4: Add resource cost metadata to service open, respec, and revive events.
- [x] Step 2.5: Update tests for missing badge rejection, successful consumption, wallet update, and
  unchanged build/reset behavior.

Verify:

```bash
(cd server && go test ./internal/game -run 'Bishop|MainConfig|BadgeReward|QuestTurnIn' -count=1)
```

## Task 3 — Client bishop panel

Files:

- Modify: `client/scripts/bishop_panel.gd`
- Modify: `client/scripts/main.gd`
- Create/modify: `client/tests/test_bishop_panel.gd`

- [x] Step 3.1: Pass current wallet and bishop cost config into `BishopPanel`.
- [x] Step 3.2: Display respec and revive badge requirements and disable each button when its
  specific badge balance is missing.
- [x] Step 3.3: Refresh the panel after `resource_wallet_update` and show successful service status
  after consumption.
- [x] Step 3.4: Cover the panel with a headless test.

Verify:

```bash
godot --headless --path client --script res://tests/test_bishop_panel.gd
```

## Task 4 — Bot proof

Files:

- Modify: `tools/bot/scenarios/45_town_bishop_respec.json`
- Modify: `tools/bot/scenarios/client/32_town_bishop_respec_panel.json`
- Modify only if needed: `tools/bot/runtime_queries.py`, `client/scripts/bot_scenario_runner.gd`,
  `client/scripts/bot_step_catalog.gd`, `client/scripts/bot_action_step_validator.gd`,
  `client/scripts/bot_controller.gd`, `client/scripts/bot_facade.gd`
- Modify if needed: `shared/rules/worlds.v0.json`

- [x] Step 4.1: Make protocol bishop respec proof acquire badges through a deterministic depth-125
  quest turn-in before respec.
- [x] Step 4.2: Assert `respec_badge` is consumed by respec and that missing badges reject in Go.
- [x] Step 4.3: Update client bishop panel proof to show badge requirements and enabled buttons
  after acquiring badges.

Verify:

```bash
make bot scenario=45_town_bishop_respec
make bot-client scenario=32_town_bishop_respec_panel HEADLESS=1
```

Manual visual command:

```bash
make bot-visual scenario=32_town_bishop_respec_panel
```

## Task 5 — Docs and lifecycle

Files:

- Existing: `docs/specs/v293_spec-bishop-badge-costs.md`
- Existing: `docs/plans/v293_2026-06-19-bishop-badge-costs.md`
- Create during finish: `docs/as-built/v293_bishop-badge-costs.md`
- Modify during finish: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`,
  `docs/progress/slice-codename-index.md`

- [x] Step 5.1: Mark the spec/plan complete after focused checks.
- [x] Step 5.2: Record proof, manual visual command, and deferred durable resurrection/stat/skill
  badge spending scope.

## Final verification

For this `$autoloop` slice, final per-slice verification is focused and batch-level `make ci` stays
owned by `$autoloop` after the selected queue completes.

- [x] `make validate-shared`
- [x] `(cd server && go test ./internal/game -run 'Bishop|MainConfig|BadgeReward|QuestTurnIn' -count=1)`
- [x] `godot --headless --path client --script res://tests/test_bishop_panel.gd`
- [x] `make bot scenario=45_town_bishop_respec`
- [x] `make bot-client scenario=32_town_bishop_respec_panel HEADLESS=1`
- [x] `make maintainability`

Batch-level `make ci` was attempted after the v292-v293 queue and was red on residual broad-suite
gates outside the bishop badge flow; see the v293 as-built for the exact failure set.

Deferred scope:

- Durable resurrection, character selection for revive, stat-badge spending, skill-badge spending,
  and production badge icons remain future slices.
