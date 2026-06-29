# v364 As-Built — Monster Movement LOD Retune

## What shipped

- Retuned `shared/rules/navigation.v0.json` crowd LOD:
  - `monster_movement_lod_min_live_monsters`: 24 → **20** (LOD engages earlier in medium crowds)
  - `monster_movement_lod_near_distance`: 6 → **8** (wider high-precision combat bubble)
  - `monster_movement_lod_update_interval_ticks`: 5 → **6** (far monsters repath less often)
- Synced navigation goldens; added `TestMovementLODRetuneKeepsMeaningfulHighPrecisionShare`.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'MovementLOD|CrowdedLightning' -count=1
ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe
```

## Deferred

- Client-side movement presentation; overload degradation expansion (v366/v365 batch siblings).
