# v365 As-Built — Path Budget Retune

## What shipped

- Retuned monster navigation budgets in `navigation.v0.json`:
  - `monster_path_requests_per_tick`: 40 → **34**
  - `monster_path_nodes_per_search`: 360 → **300**
  - `monster_path_nodes_per_tick`: 600 → **500**
- Synced goldens; added `TestCrowdedLightningAveragePathNodesStayBelowBudgetCeiling` (95% of per-tick cap).

## Proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'CrowdedLightning' -count=1
ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe
```
