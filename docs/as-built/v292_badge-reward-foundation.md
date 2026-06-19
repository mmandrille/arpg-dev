# v292 As Built: Badge Reward Foundation

Date: 2026-06-19
Spec: [`docs/specs/v292_spec-badge-reward-foundation.md`](../specs/v292_spec-badge-reward-foundation.md)
Plan: [`docs/plans/v292_2026-06-19-badge-reward-foundation.md`](../plans/v292_2026-06-19-badge-reward-foundation.md)

## What shipped

- Added data-driven badge reward rows in `main_config.v0.json` for upgrade, respec, stat, skill,
  and resurrection resources at depths 10/20/30/40/50 with 25% base chance and +1% per depth,
  capped at 100%.
- Kept `upgrade_shard` as the blacksmith resource ID but renamed it to `Upgrade Badge`, and added
  `respec_badge`, `stat_badge`, `skill_badge`, and `resurrection_badge` as non-equippable currency
  item definitions.
- Added badge presentation metadata using the existing code-native badge primitive for ground loot
  and item icons, with no external assets or plugins.
- Added Go load-time and Python validation for badge reward rows and badge item references.
- Added server badge reward helpers that roll every eligible badge independently, grant successful
  rolls into the owning account wallet, emit `resource_wallet_update`, and append `badge_rewarded`
  events with source metadata.
- Boss kills roll badge rewards from the active dungeon depth; quest town turn-ins roll from the
  character's `deepest_dungeon_depth`.
- Extended debug progression seeding for protocol and client bots so `deepest_dungeon_depth` can be
  seeded deterministically.
- Updated the character-bar material wallet labels and added headless proof that compact labels,
  tooltips, and the wallet window show the new badge names.
- Added protocol bot scenario `98_badge_reward_foundation` and client bot scenario
  `76_badge_reward_wallet`. Both use depth 125 so every configured badge reaches a deterministic
  100% proof chance.

## Proof

Focused verification:

```bash
make validate-shared
(cd server && go test ./internal/game -run 'BadgeReward|QuestTurnIn|ResourceWallet|BossSpecial' -count=1)
(cd server && go test ./internal/http -run TestDebugCharacterProgression -count=1)
godot --headless --path client --script res://tests/test_material_wallet_badges.gd
godot --headless --path client --script res://tests/test_character_bar.gd
make bot scenario=98_badge_reward_foundation
make bot-client scenario=76_badge_reward_wallet HEADLESS=1
make maintainability
```

Result: green on 2026-06-19.

Full verification:

```bash
make ci
```

Result: not run for this focused slice. The selected-batch full CI remains pending after the current
autoloop queue; the previous v291 full-CI attempt was red on non-v291 gates.

## Manual visual command

```bash
make bot-visual scenario=76_badge_reward_wallet
```

## Deferred

- v293 will make bishop respec and resurrection consume matching badges when gameplay debug is
  disabled.
- Stat-badge and skill-badge spending routes remain deferred.
- Quest source-depth metadata remains deferred; quest turn-in badge rolls currently use
  `deepest_dungeon_depth`.
- Production badge art/icons and a richer dedicated badge UI remain deferred.
