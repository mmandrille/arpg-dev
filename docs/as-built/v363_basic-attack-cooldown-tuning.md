# v363 As-Built — Basic Attack Cooldown Tuning

Date: 2026-06-28

## Shipped

- `base_attack_interval_ticks` lowered from 14 → 12 in `main_config.v0.json` and `combat.v0.json`.
- Cross-language goldens updated for new interval math.

## Verification

```bash
make validate-shared
cd server && go test ./internal/game/... -run Golden
make client-unit
```
