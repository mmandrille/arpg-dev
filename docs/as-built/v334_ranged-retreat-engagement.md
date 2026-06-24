# v334 As-Built — ranged-retreat-engagement

**Date:** 2026-06-24  
**Status:** Complete

## What was proved

Ranged monsters with `preferred_min_range` only retreat after remaining in melee attack range of
their chase target for `ranged_retreat_min_melee_engagement_seconds` (default 2.0s at 10 Hz). This
prevents instant kite-on-approach while preserving eventual repositioning.

## Key decisions

- Engagement timer is server-owned per monster entity; leaving melee range resets it.
- Tuning lives in `main_config.v0.json`, not hardcoded ticks.
- Existing `ranged_monster_retreat_lab` proof kept with a longer eventual-retreat window.

## Verification

```bash
cd server && go test ./internal/game/... -run RangedMonsterRetreat
make validate-shared
```
