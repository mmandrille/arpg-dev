# v436 Plan — Weapons Tier-3 Melee

- [x] Add `tools/assets/import_equipment_glb.py` (Poly Pizza download + extent normalize)
- [x] Import melee weapons + shields; archive legacy procedural GLBs
- [x] Update manifest provenance + sha256
- [x] `make validate-assets`

## Verification

```bash
python3 tools/assets/import_equipment_glb.py assets/equipment/weapons/rusty_sword/rusty_sword.glb assets/equipment/weapons/long_sword/long_sword.glb assets/equipment/weapons/rapier/rapier.glb assets/equipment/weapons/starter_axe/starter_axe.glb assets/equipment/armor/shield/kite_shield.glb assets/equipment/armor/shield/tower_shield.glb
make validate-assets
```
