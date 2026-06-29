# v364 Plan: Monster Movement LOD Retune

## Tasks

- [x] Spec
- [x] Retune `navigation.v0.json` LOD fields
- [x] Sync navigation goldens
- [x] Add semantic crowded high-precision share test
- [x] Verify focused Go tests + crowded perf probe
- [x] As-built + PROGRESS + lifecycle + commit

## Verification

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'MovementLOD|CrowdedLightning' -count=1
ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe
```
