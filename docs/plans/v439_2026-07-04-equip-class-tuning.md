# v439 Plan — Equip Class Tuning

- [x] Extend `item_visuals.v0.schema.json` with `class_transforms`
- [x] Sample overrides on `long_sword` for rogue/barbarian
- [x] `EquipmentVisualResolver.set_character_class()` + merge path
- [x] Wire `main.gd` after model swap

## Verification

```bash
make validate-shared
make client-unit
```
