# v400 Spec — Scaled Upgrade Risk

Status: Complete
Date: 2026-07-01
Codename: scaled-upgrade-risk
Baseline: v399 `ws-reconnect-bot-proof` complete

## Purpose

Turn blacksmith item upgrades into a depth-gated, data-driven risk loop:

1. **Safe early tier** — upgrades targeting item level ≤1 stay 100% success.
2. **Logarithmic failure curve** — failure chance rises from target level 2; at target level 10 failure is ~75% (25% success) per configured anchors.
3. **Shard insurance** — spending an upgrade shard above the minimum required tier adds +10% success per tier (data-driven).
4. **Depth-only cap** — remove the flat `item_upgrade_max_level` gate; deepest dungeon depth is the only upgrade ceiling (e.g. floor 140 → ilvl 14 band allows ilvl 12 upgrades when depth and shard tier qualify).
5. **Legible tooltips** — blacksmith preview shows base success, shard bonus, effective chance, pity, spend line, failure outcome, and requirement scaling note.

Failed attempts consume gold + shard, leave stats/requirements unchanged, and increment per-item pity (existing v203 behavior).

## Non-goals

- Affix add/improve-roll recipes
- Item bricking, refunds, or durability loss
- Market/trade restrictions for upgraded items
- Full next-requirement numeric preview (client shows scaling note + current requirements when present)
- ci-full cluster triage (separate maintenance slice)
- Production blacksmith art

## Acceptance criteria

- [x] `item_upgrade_max_level: 0` means depth-only cap (no flat config ceiling)
- [x] Target level ≤1 upgrades remain 100% success
- [x] Target level 10 with minimum-tier shard shows ~25% success (~75% failure) per curve [2→10% fail, 10→75% fail] on log10 curve
- [x] Each shard tier above minimum adds configured bonus to effective success (default +10%/tier)
- [x] Server computes effective success from target level + selected shard level; client preview matches formula from shared rules
- [x] Failed attempt spends gold + shard, keeps item unchanged, increments pity
- [x] Pity still guarantees success at threshold
- [x] Successful upgrade rescales stats and requirements (existing behavior)
- [x] Blacksmith tooltip shows base chance, shard bonus, effective chance, pity, and failure outcome when risk > 0
- [x] Character at depth 140 can upgrade ilvl 11 → 12 with ilvl 11 shard when depth cap allows
- [x] Golden fixture + Go tests lock failure/success formula; client-side tests green
- [x] Client unit test for preview lines; extended bot scenario proves risk tooltip fields
- [x] `make validate-shared`, focused tests green

## Scope and likely files

- `shared/rules/main_config.v0.json` + schema — failure curve, shard bonus, max level 0
- `shared/golden/upgrade_success_chance.json` — cross-language formula cases
- `server/internal/game/item_upgrade_chance.go` — failure/success helpers
- `server/internal/game/item_roll_upgrade.go` — depth-only max level
- `server/internal/game/rules.go` — new config fields + validation
- `server/internal/http/account_stash.go` — effective chance from shard level
- `server/internal/store/repos.go` — allow max level 0
- `client/scripts/blacksmith_upgrade_chance.gd`, `blacksmith_upgrade_preview.gd`, `blacksmith_panel.gd`, `blacksmith_panel_actions.gd`
- `client/tests/test_blacksmith_upgrade_chance.gd`
- `tools/bot/scenarios/client/blacksmith_upgrade_risk.json` (extended)

## Test and bot proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'UpgradeFailure|UpgradeSuccess|EffectiveItemUpgradeMax' -count=1
cd server && go test ./internal/store/... -run 'Upgrade' -count=1
godot --headless --path client --script res://tests/test_blacksmith_upgrade_chance.gd
make bot-client SCENARIO=blacksmith_upgrade_risk HEADLESS=1
```

## Asset decision

- **Reject** — external plugins; text-only tooltip updates on existing blacksmith panel

## Open questions

Resolved:

- Q1: Safe through target level 1; log curve from level 2; ~75% failure at target level 10.
- Q5: `item_upgrade_max_level: 0` — depth-only cap.
