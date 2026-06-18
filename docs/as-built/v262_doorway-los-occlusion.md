# v262 As-Built - Doorway LOS Occlusion

Date: 2026-06-18
Spec: [`docs/specs/v262_spec-doorway-los-occlusion.md`](../specs/v262_spec-doorway-los-occlusion.md)
Plan: [`docs/plans/v262_2026-06-18-doorway-los-occlusion.md`](../plans/v262_2026-06-18-doorway-los-occlusion.md)

## Shipped Behavior

- Server fog-of-war line-of-sight now treats closed interactables with
  `barrier_when_closed` as rectangular occluders.
- Open doors and closed interactables without a barrier definition do not block fog visibility.
- Existing generated wall occlusion remains in the same fog segment-check path.
- Opening a closed door can reveal an idle hidden monster through the existing fog transition flow.
- `FogOfWarOverlay` now accepts supplied non-wall occluders via `set_occluder_layout` and renders
  them through the same shadow polygon pipeline used for walls.

## Boundaries

- No protocol-visible barrier metadata was added.
- Client integration remains a supplied-occluder path; generated door protocol and `main.gd`
  forwarding are intentionally unchanged.
- No persistent fog, reconnect/resume behavior, secret/locked/keyed/destructible door rules, or
  generalized polygon LOS was added.

## Verification

```bash
cd server && go test ./internal/game -run TestFogOfWar
make client-unit
make maintainability
```

All commands passed on 2026-06-18. The selected autoloop batch `make ci` gate remains pending.

Manual visual proof for the existing LOS shadow-mask scenario, if desired:

```bash
make bot-visual scenario=68_fog_los_shadow_mask
```
