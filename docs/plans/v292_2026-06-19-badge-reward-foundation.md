# v292 Plan — Badge Reward Foundation

Status: Complete
Goal: Add data-driven badge resources that boss kills and quest town turn-ins grant into the account wallet.
Architecture: Badge reward definitions live in `main_config.v0.json` as shared tuning rows. The Go
sim rolls every eligible badge independently with the seeded sim RNG, grants successful rolls through
the existing resource wallet, and emits reward events plus `resource_wallet_update` changes. Client
presentation reuses the material wallet and item-definition labels; no new protocol schema version
or asset dependency is needed.
Tech stack: Shared rule catalogs, Go sim reward helpers/tests, Python protocol bot scenario, Godot
client wallet scenario, SDD docs.

## Baseline and shortcut decision

Builds on v221 resource-wallet foundation, v237/v244 wallet UI details/window, v249 boss reward
presentation, v291 quest town turn-in, ADR-0009 boss rewards, ADR-0012 upgrade resources, and
ADR-0014 economy/resource constraints.

Asset/plugin decision:

- Adopt: existing account `resource_wallet`, material wallet UI, item-definition labels, boss kill
  and quest turn-in server paths, and bot wallet assertions.
- Borrow: upgrade-shard wallet tests, `boss_special_drops` setup, and `quest_town_turn_in` scenario
  steps.
- Reject: external assets/plugins, production badge icons, new asset pipeline, new protocol schema
  version, and a dedicated badge inventory UI.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/main_config.v0.json` | Add badge reward tuning rows. |
| Modify | `shared/rules/main_config.v0.schema.json` | Validate badge reward row shape. |
| Modify | `shared/rules/items.v0.json` | Present upgrade shard as a badge and add respec/stat/skill/resurrection badges. |
| Modify | `tools/validate_main_config.py` | Validate badge rows without growing `tools/validate_shared.py`. |
| Modify | `server/internal/game/rules.go` | Add the minimal typed config field and call focused validation without net line growth. |
| Modify | `server/internal/game/main_config_validation.go` | Validate badge tuning values. |
| Create | `server/internal/game/badge_rewards.go` | Compute chances, roll badge rewards, and grant wallet resources. |
| Create | `server/internal/game/badge_rewards_test.go` | Cover thresholds, scaling, boss grants, and quest grants. |
| Modify | `server/internal/game/resource_wallet.go` | Treat configured badge IDs as wallet resources. |
| Modify | `server/internal/game/quest_turn_in.go` | Grant quest turn-in badge rewards after item consumption. |
| Modify | `server/internal/game/sim.go` | Add one no-net-growth hook from boss loot to badge rewards. |
| Create | `client/tests/test_material_wallet_badges.gd` | Prove wallet labels/window render badge names from item definitions. |
| Create | `tools/bot/scenarios/98_badge_reward_foundation.json` | Protocol proof for high-depth quest badge rewards. |
| Create | `tools/bot/scenarios/client/76_badge_reward_wallet.json` | Client proof for wallet badge visibility. |
| Create during finish | `docs/as-built/v292_badge-reward-foundation.md` | Record proof and deferred scope. |
| Modify during finish | `PROGRESS.md`, `docs/progress/slice-lifecycle.md`, `docs/progress/slice-codename-index.md` | Lifecycle updates. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines. Grandfathered files must not grow beyond
their current allowance.

Hotspot / over-limit files touched:

- [x] `server/internal/game/rules.go` is grandfathered (`3303` baseline, `3302` current); keep edits
  to one config field plus one focused validation call and remove at least one stale/blank line if
  needed so the file remains at or below baseline.
- [x] `server/internal/game/sim.go` is grandfathered and currently at its +25 allowance; add only a
  single hook and remove an equal number of lines so there is no net growth.
- [x] Avoid `server/internal/game/game_test.go`; add `badge_rewards_test.go` instead.
- [x] Avoid `tools/bot/run.py`; existing action/assertion support is enough.
- [x] Avoid `tools/validate_shared.py`; extend `tools/validate_main_config.py` instead.
- [x] New Go/GDScript/test files must stay under 600 lines.

Decision:

- [x] Extract badge reward behavior into `badge_rewards.go`/`badge_rewards_test.go`; no coordinator
  growth beyond unavoidable hook/config lines.

Verification:

```bash
make maintainability
```

## Task 1 — Shared badge reward rules

Files:

- Modify: `shared/rules/main_config.v0.json`
- Modify: `shared/rules/main_config.v0.schema.json`
- Modify: `shared/rules/items.v0.json`
- Modify: `tools/validate_main_config.py`
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/main_config_validation.go`

- [x] Step 1.1: Add badge reward rows for `upgrade_shard`, `respec_badge`, `stat_badge`,
  `skill_badge`, and `resurrection_badge` with unlock depths `10/20/30/40/50`, base chance `25`,
  and per-depth increase `1`.
- [x] Step 1.2: Rename the `upgrade_shard` display name to `Upgrade Badge` and add the four new
  badge item definitions as non-equippable currency resources.
- [x] Step 1.3: Validate badge row item IDs, positive unlock depths, chance ranges, and non-negative
  per-depth increases in schema, Python validation, and Go load-time validation.

Verify:

```bash
make validate-shared
```

## Task 2 — Server badge grant behavior

Files:

- Create: `server/internal/game/badge_rewards.go`
- Create: `server/internal/game/badge_rewards_test.go`
- Modify: `server/internal/game/resource_wallet.go`
- Modify: `server/internal/game/quest_turn_in.go`
- Modify: `server/internal/game/sim.go`

- [x] Step 2.1: Add chance calculation helpers that return no chance below unlock depth and cap
  scaled chances at 100%.
- [x] Step 2.2: Add a wallet grant helper that increments the current player's account wallet,
  appends `resource_wallet_update`, and emits a `badge_rewarded` event with source metadata.
- [x] Step 2.3: Roll boss badge rewards from `abs(levelNum)` after boss kill rewards and before the
  result is returned.
- [x] Step 2.4: Roll quest badge rewards from `progression.DeepestDungeonDepth` after the quest
  turn-in consumes the item and awards gold.
- [x] Step 2.5: Treat every configured badge reward item ID as a wallet resource for explicit or
  automatic pickup compatibility.
- [x] Step 2.6: Cover threshold/scaling math, forced deterministic boss badge grant, forced quest
  turn-in badge grant, and legacy upgrade-shard wallet pickup.

Verify:

```bash
(cd server && go test ./internal/game -run 'BadgeReward|QuestTurnIn|ResourceWallet|BossSpecial' -count=1)
```

## Task 3 — Client wallet proof

Files:

- Create: `client/tests/test_material_wallet_badges.gd`
- Modify only if needed: `client/scripts/character_bar.gd`
- Modify only if needed: `client/scripts/material_wallet_panel.gd`

- [x] Step 3.1: Add a headless client test that feeds badge balances into `CharacterBar`, opens the
  material wallet, and asserts compact label text plus detailed badge names.
- [x] Step 3.2: Patch wallet label fallback only if the existing item-definition path does not show
  badge names clearly.

Verify:

```bash
godot --headless --path client --script res://tests/test_material_wallet_badges.gd
```

## Task 4 — Bot proof

Files:

- Create: `tools/bot/scenarios/98_badge_reward_foundation.json`
- Create: `tools/bot/scenarios/client/76_badge_reward_wallet.json`

- [x] Step 4.1: Add protocol proof in `quest_turn_in_lab` with debug `deepest_dungeon_depth: 125`,
  turn in `Quest Leaf`, and assert positive badge wallet counts plus existing gold reward.
- [x] Step 4.2: Add client proof for the same turn-in, then open the material wallet and assert
  badge text is visible.

Verify:

```bash
make bot scenario=98_badge_reward_foundation
make bot-client scenario=76_badge_reward_wallet HEADLESS=1
```

Manual visual command:

```bash
make bot-visual scenario=76_badge_reward_wallet
```

## Task 5 — Docs and lifecycle

Files:

- Existing: `docs/specs/v292_spec-badge-reward-foundation.md`
- Existing: `docs/plans/v292_2026-06-19-badge-reward-foundation.md`
- Create during finish: `docs/as-built/v292_badge-reward-foundation.md`
- Modify during finish: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`,
  `docs/progress/slice-codename-index.md`

- [x] Step 5.1: Mark the spec/plan complete after focused checks.
- [x] Step 5.2: Record proof, manual visual command, and deferred badge spending/source-depth scope
  in as-built/progress docs.

## Final verification

For this `$autoloop` slice, final per-slice verification is focused and batch-level `make ci` stays
owned by `$autoloop` after the selected queue completes.

- [x] `make validate-shared`
- [x] `(cd server && go test ./internal/game -run 'BadgeReward|QuestTurnIn|ResourceWallet|BossSpecial' -count=1)`
- [x] `(cd server && go test ./internal/http -run TestDebugCharacterProgression -count=1)`
- [x] `godot --headless --path client --script res://tests/test_material_wallet_badges.gd`
- [x] `godot --headless --path client --script res://tests/test_character_bar.gd`
- [x] `make bot scenario=98_badge_reward_foundation`
- [x] `make bot-client scenario=76_badge_reward_wallet HEADLESS=1`
- [x] `make maintainability`

Deferred scope:

- v293 will make bishop respec/revive consume matching badges when gameplay debug is disabled.
- Stat and skill badge spending, source-depth quest-item metadata, and production badge icons remain
  future slices unless v293 discovers a small safe extension.
