# v185 As-Built: Companion Rank Scaling and Limits

Date: 2026-06-15

## Shipped

- Companion skill limits now use shared rule data:
  - `base`
  - `per_rank_step`
  - `ranks_per_step`
- Revive limit is data-configured as `1 + floor((rank - 1) / 3)`.
- Ranger wolf remains one active companion through data (`per_rank_step: 0`).
- Summon and revive spawning share deterministic limit enforcement:
  - count existing companions with same owner and source skill
  - remove oldest same-skill companions only when the new spawn would exceed the limit
- Revived monster HP/damage scaling stays rule-derived from Revive power percent.
- Bot entity assertions can filter by `hp` and `max_hp` for rank-scaling proof.
- Added `companion_rank_scaling_lab` and protocol scenario `76_companion_rank_scaling_and_limits.json`.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game/... -run 'Revive|Companion|RangerBlackWolf|RangerSkillRules' -count=1`
- `make bot scenario=76_companion_rank_scaling_and_limits.json`
- `make ci`

## Visual Check

```bash
make bot-visual scenario=76_companion_rank_scaling_and_limits.json
```
