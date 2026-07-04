# v439 As Built — Equip Class Tuning

Date: 2026-07-04

## Shipped

- `class_transforms` optional block on item visuals (schema + data).
- `EquipmentVisualResolver.set_character_class()` merges per-class transforms.
- `main.gd` calls resolver after class model swap.

## Verification

```bash
make validate-shared
make client-unit
```
