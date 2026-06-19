# v292 Spec: Badge Reward Foundation

Status: Implemented
Date: 2026-06-19
Codename: `badge-reward-foundation`
Baseline: v291 `quest-town-turn-in`

## Purpose

Add the first server-authoritative badge reward layer for boss kills and town quest turn-ins.
Badges are account-wallet resources, not bag items, and are rolled from shared rules based on
dungeon depth. This establishes the supply side for upgrade, respec, stat, skill, and resurrection
resources before later slices consume every badge type.

## Non-goals

- Do not implement paid bishop respec/revive costs yet; that is the next selected slice.
- Do not add stat-badge or skill-badge consumption routes, passive tree changes, item mutation
  changes, resurrection character selection, or corpse/death persistence changes.
- Do not add source-depth metadata to quest items yet. Quest turn-in badge rolls use the
  character's current `deepest_dungeon_depth` as the conservative depth signal.
- Do not add new protocol schema versions; reuse existing `resource_wallet_update` changes and
  permissive event payload fields.
- Do not add production icon art, external assets, plugins, or a new wallet UI panel.

## Acceptance Criteria

- Shared rules define badge reward rows for:
  - upgrade badge at depth 10,
  - respec badge at depth 20,
  - stat badge at depth 30,
  - skill badge at depth 40,
  - resurrection badge at depth 50.
- Each badge row is data-driven with 25% base chance at its unlock depth and +1 percentage point per
  depth after unlock, capped at 100%.
- The existing `upgrade_shard` resource remains the blacksmith upgrade resource ID but is presented
  as an upgrade badge; new badge item definitions exist for respec, stat, skill, and resurrection.
- Boss kills roll every eligible badge independently from the active dungeon depth and grant won
  badges directly into the owning account wallet.
- Quest town turn-in rolls every eligible badge independently from `deepest_dungeon_depth` and grants
  won badges alongside the existing configured gold reward.
- Badge grants emit resource wallet changes and stable reward events with `resource_id`, `amount`,
  and reward source metadata.
- Existing wallet UI displays all positive badge balances using item definitions, and the client
  wallet window can show the new badge names without new asset dependencies.
- Focused tests prove depth thresholds, chance scaling, deterministic boss badge grants, quest
  turn-in badge grants, and wallet auto-pickup compatibility.
- Protocol and client bot proof show a quest turn-in granting a badge into the wallet and rendering
  it in the material wallet UI.

## Scope And Likely Files

- Shared: `shared/rules/main_config.v0.json`, `shared/rules/main_config.v0.schema.json`,
  `shared/rules/items.v0.json`, `shared/assets/item_presentations.v0.json`, and
  `shared/assets/item_presentations.v0.schema.json`.
- Server: `server/internal/game/rules.go`, `server/internal/game/main_config_validation.go`,
  `server/internal/game/resource_wallet.go`, `server/internal/game/quest_turn_in.go`,
  `server/internal/game/badge_rewards.go`, focused Go tests.
- Client: existing `client/scripts/character_bar.gd` and `client/scripts/material_wallet_panel.gd`
  should work from item definitions; only touch them if test proof exposes a display gap.
- Bot/debug: new or updated protocol and client scenarios proving badge wallet updates, plus debug
  progression seed support for `deepest_dungeon_depth`.
- Docs: v292 plan/as-built/lifecycle updates.

## Test And Bot Proof

Focused checks:

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

Manual visual verification command:

```bash
make bot-visual scenario=76_badge_reward_wallet
```

## Asset And Plugin Decision

- Adopt: existing account `resource_wallet`, material wallet UI, item-definition labels, boss kill
  events, quest town turn-in service, and bot wallet assertions.
- Borrow: upgrade-shard wallet pickup tests and boss special drop scenario patterns.
- Reject: external assets/plugins, production badge icons, new asset pipelines, and a new dedicated
  badge inventory UI for this slice.

## ADR Alignment

- ADR-0001: rewards stay server-authoritative and deterministic under the sim RNG.
- ADR-0009: boss rewards remain tied to boss kills without weakening the boss-floor gate.
- ADR-0012: upgrade resources remain wallet resources and stay data-driven.
- ADR-0014: this adds several resources, so each badge has a distinct player-visible purpose and
  later consumption is split into explicit small slices instead of one opaque multi-resource system.

## Open Questions And Risks

- No blocking questions. The slice intentionally creates badge supply before consuming respec,
  stat, skill, and resurrection badges.
