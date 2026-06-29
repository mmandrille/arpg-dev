# v368 Spec: Remote Adaptive Smoothing

Status: Draft  
Date: 2026-06-28  
Codename: `remote-adaptive-smoothing`  
Baseline: v367 `camera-follow-damping`

## Purpose

Scale remote entity tick-smoothing duration by segment distance so co-op remotes and visible monsters
glide smoothly under movement variance without over-lengthening micro-corrections.

## Non-goals

- No local player anchor smoothing changes.
- No server/protocol changes.

## Acceptance criteria

- `movement_presentation.v0.json` adds `remote_adaptive` tuning block.
- Non-local `player`/`monster`/`companion` entities use distance-scaled segment duration.
- Unit test proves override duration is applied.
- v349 `entity_tick_smoothing` extended scenario still passes.

## Test proof

```bash
make validate-shared
make client-unit
HEADLESS=1 make bot-client SCENARIO=84_entity_tick_smoothing HEADLESS=1
```
