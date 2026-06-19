# v293 As Built: Bishop Badge Costs

Date: 2026-06-19
Spec: [`docs/specs/v293_spec-bishop-badge-costs.md`](../specs/v293_spec-bishop-badge-costs.md)
Plan: [`docs/plans/v293_2026-06-19-bishop-badge-costs.md`](../plans/v293_2026-06-19-bishop-badge-costs.md)

## What shipped

- Added data-driven bishop service costs in `main_config.v0.json`: respec costs
  `respec_badge` x1 and revive-all costs `resurrection_badge` x1.
- Extended shared schema, Python validation, and Go load-time validation so bishop resource costs
  require non-negative counts, non-empty IDs when positive, and non-equippable currency items.
- Extracted bishop respec handling from `handlers.go` into `bishop_respec.go`, added focused
  bishop wallet helpers, and lowered the `handlers.go` file-size ratchet baseline.
- Bishop service open now reports respec affordability from both gold and the configured respec
  badge balance, with resource-cost metadata on service events.
- Bishop respec and revive-all now reject with `missing_resource` when the matching badge is absent,
  consume one badge on success, emit `resource_wallet_update`, and attach resource metadata to the
  existing service events.
- Fixed the v292 badge reward wallet helper so active-player quest badge grants no longer restore a
  stale player snapshot and drop just-earned quest gold.
- The bishop panel now loads configured badge costs, displays respec/revive badge requirements,
  disables each action by its own wallet balance, and refreshes when the wallet changes.
- Updated protocol and client bishop scenarios to earn deterministic depth-125 quest badges before
  opening the bishop and spending the respec badge.
- Updated leftover upgrade-resource UI tests/scenarios from the old `Shard` label to
  `Upgrade Badge`.

## Proof

Focused verification:

```bash
make validate-shared
(cd server && go test ./internal/game -run 'Bishop|MainConfig|BadgeReward|QuestTurnIn' -count=1)
godot --headless --path client --script res://tests/test_bishop_panel.gd
godot --headless --path client --script res://tests/test_shop_panel.gd
godot --headless --path client --script res://tests/test_blacksmith_panel.gd
make bot scenario=45_town_bishop_respec
make bot-client scenario=32_town_bishop_respec_panel HEADLESS=1
make bot-client scenario=39_blacksmith_upgrade_ui HEADLESS=1
make bot-client scenario=54_material_wallet_details HEADLESS=1
make bot-client scenario=61_material_wallet_window HEADLESS=1
make maintainability
```

Result: green on 2026-06-19.

Full verification:

```bash
make ci
```

Result: red on 2026-06-19. In-scope v293/v292 badge gates passed inside the run
(`town_bishop_respec`, `badge_reward_foundation`, `town_bishop_respec_panel`,
`badge_reward_wallet`). Residual failures were outside this slice: `go test ./...` failed
`TestEliteMinionFollowsLeaderWithoutPassiveAggro`; broad protocol bot failed
`mercenary_hiring_board` and `mercenary_death_loss`; broad client bot failed older combat,
boss, mercenary, and pre-label-fix wallet scenarios; smoke failed pre-label-fix blacksmith tests.
The in-scope label-fix tests/scenarios listed above were rerun green after the full-CI attempt.

## Manual visual command

```bash
make bot-visual scenario=32_town_bishop_respec_panel
```

## Deferred

- Durable resurrection effects, character selection for revive, and death recovery UX remain
  deferred; revive-all only consumes the configured badge and keeps the existing no-op event shape.
- Stat-badge and skill-badge spending routes remain deferred.
- Production badge art/icons and a dedicated badge inventory UI remain deferred.
- Full-CI residuals outside the badge/bishop flow need a stabilization slice before the next green
  repo-wide gate.
