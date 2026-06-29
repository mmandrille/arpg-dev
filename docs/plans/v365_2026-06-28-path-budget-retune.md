# v365 Plan: Path Budget Retune

## Tasks

- [x] Spec
- [ ] Retune monster path budgets in navigation rules + goldens
- [ ] Add average path-node semantic test
- [ ] Verify focused tests + crowded perf probe
- [ ] As-built + lifecycle + commit

## Verification

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'CrowdedLightning|PathBudget' -count=1
ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe
```
