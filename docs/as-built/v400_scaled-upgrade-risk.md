# v400 as-built — Scaled Upgrade Risk

## What it proved

- Blacksmith upgrades use a data-driven log10 failure curve from target item level 2 upward;
  target level ≤1 stays 100% success.
- `item_upgrade_max_level: 0` removes the flat config cap; deepest dungeon depth is the only
  upgrade ceiling via existing tier rules.
- Higher-tier upgrade shards add configurable success bonus per tier above the minimum required
  shard level (default +10%/tier).
- Server computes authoritative effective success from current level + staged shard level; client
  preview mirrors the same formula from shared rules.
- Blacksmith tooltips show base success, shard bonus, effective chance, pity, spend line, failure
  outcome, and requirement-scaling note when risk > 0.
- Failed attempts still consume gold + shard, leave stats unchanged, and increment pity (v203).

## Key files

- `shared/rules/main_config.v0.json` — failure curve anchors, shard bonus, max level 0
- `shared/golden/upgrade_success_chance.json` — cross-language formula contract
- `server/internal/game/item_upgrade_chance.go` — failure/success helpers
- `server/internal/http/account_stash.go` — effective chance + shard resolve
- `client/scripts/blacksmith_upgrade_chance.gd` — client formula mirror
- `client/scripts/blacksmith_upgrade_preview.gd` — rich tooltip lines
- `tools/bot/scenarios/client/blacksmith_upgrade_risk.json` — extended client proof

## Verification

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'UpgradeFailure|UpgradeSuccess|EffectiveItemUpgradeMax' -count=1
godot --headless --path client --script res://tests/test_blacksmith_upgrade_chance.gd
godot --headless --path client --script res://tests/test_blacksmith_panel.gd
make bot-client SCENARIO=blacksmith_upgrade_risk HEADLESS=1
make maintainability
```

## Non-goals honored

- No affix add/improve-roll recipes, bricking, or market restrictions.
- No ci-full cluster triage; scenario is extended-only.
