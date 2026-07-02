# v407 Plan — Weapon Elemental Damage

Status: Complete
Goal: Split weapon elemental rolls from physical damage and prove typed elemental hits on basic attacks.
Architecture: Keep `bonus_*_damage` stat keys; exclude them from physical range; apply flat typed follow-up hit after successful physical basic attack. Lab worlds may pin loot via `loot_preset`.
Tech stack: Go sim, shared JSON rules, Python protocol bot.

## Baseline and shortcut decision

Reuse v100 resistances and v389 item-level scaling. Client stat labels borrow existing tooltip paths; no new art.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/item_templates.v0.json` | elemental affix 1–6 at ilvl 1 |
| Modify | `shared/rules/worlds.v0.json` + schema | lab world + loot preset |
| Modify | `shared/rules/monsters.v0.json` | cold-resistant lab monster |
| Create | `server/internal/game/weapon_elemental.go` | split damage + roll exclusivity helpers |
| Create | `server/internal/game/weapon_elemental_test.go` | focused proofs |
| Modify | `server/internal/game/sim.go` | apply elemental after physical hit; loot preset spawn |
| Modify | `server/internal/game/item_rolls.go`, `item_reroll.go` | mutually exclusive elemental affix rolls |
| Modify | `server/internal/game/shop.go` | "+X Element Damage" summary lines |
| Modify | `server/internal/game/affix_names.go` | physical-only roll contribution |
| Modify | `server/internal/game/mercenary_character.go` | physical-only roll contribution |
| Modify | `server/internal/game/rules.go` | `WorldLootPreset` on entities |
| Create | `tools/bot/scenarios/116_weapon_elemental_damage.json` | protocol proof |

## Maintenance ratchet

Hotspot files touched: `sim.go`, `shop.go` — extract logic to `weapon_elemental.go`; sync baselines for grandfathered files.

Verification:
```bash
make maintainability
```

## Task 1 — Shared rules and lab preset

- [x] Step 1.1: Bump elemental affix template min/max to 1–6; add cold-resistant monster; add `weapon_elemental_lab` with `loot_preset` cold sword
```bash
make validate-shared
```

## Task 2 — Server roll exclusivity and split damage

- [x] Step 2.1: Implement elemental affix exclusivity on item rolls/rerolls
- [x] Step 2.2: Exclude elemental from physical contributions; apply typed follow-up hit on basic attack
- [x] Step 2.3: Spawn world loot from `loot_preset`
```bash
cd server && go test ./internal/game/... -run 'WeaponElemental|ElementalAffix' -count=1
```

## Task 3 — Tooltip summary lines

- [x] Step 3.1: Show "+X Cold Damage" (etc.) in `itemSummaryLines`
```bash
cd server && go test ./internal/game/... -run 'TestItemSummary|WeaponElemental' -count=1
```

## Task 4 — Bot scenario

- [x] Step 4.1: Add `116_weapon_elemental_damage.json` (`ci_tier: extended`)
```bash
make bot scenario=116_weapon_elemental_damage
```

## Task 5 — Lifecycle docs

- [x] Update PROGRESS.md, slice lifecycle, as-built

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/... -run 'WeaponElemental|ElementalAffix' -count=1`
- [x] `make bot scenario=116_weapon_elemental_damage`
