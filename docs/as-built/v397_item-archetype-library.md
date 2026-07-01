# v397 As-built — Item Archetype Library

## What shipped

- Replaced all `cave_*` equipment templates with generic archetype IDs (`long_sword`, `mail`, `bow`, …) and
  added `equipment_category` (`weapon_1h`, `weapon_2h`, `off_hand`, `gear`, `jewelry`).
- Seven new archetypes in treasure classes: `dagger`, `hammer`, `morningstar`, `wand`, `warhammer`, `buckler`, `sash`.
- Unified display-name grammar: common = archetype only; magic/rare = affix + archetype; unique/set =
  affix + archetype + `of` + effect/set name.
- Weapon elemental rolls (`bonus_cold_damage`, etc.) affect affix words, damage range, and basic-attack
  `damage_type`.
- Extended bot scenario `item_archetype_lab` proves naming + cold affix path.

## Proof

```bash
make validate-shared
make validate-assets
cd server && go test ./internal/game/... -run 'AffixName|Archetype|ElementalWeapon|SetItem' -count=1
make bot scenario=item_archetype_lab
```

Local play after merge: `make db-reset` (breaking template ID rename).

## Deferred

- Off-hand book archetype, jewelry subtypes, full affix suffix grammar, stash filters by category.
