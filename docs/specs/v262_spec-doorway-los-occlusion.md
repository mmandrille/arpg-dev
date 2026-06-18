# v262 Spec - Doorway LOS Occlusion

Status: Complete
Date: 2026-06-18
Codename: doorway-los-occlusion

## Purpose

Make closed doorways participate in fog-of-war line-of-sight occlusion. A monster behind a closed
door should be hidden by authoritative fog filtering just like a monster behind a generated wall;
opening the door should clear that occluder.

## Non-goals

- No new door protocol fields, persistent fog snapshots, reconnect/resume behavior, secret doors,
  locked/keyed doors, destructible doors, or generalized polygon LOS.
- No client pathfinding change, door auto-open behavior, minimap marker change, or generated
  dungeon tuning change.
- No high-wall height system beyond the existing wall-shadow presentation.

## Acceptance Criteria

- Server fog visibility treats closed interactables with `barrier_when_closed` as LOS blockers.
- Open doors and non-barrier interactables do not block fog visibility.
- Existing wall occlusion behavior remains unchanged.
- Focused server tests prove a closed door hides a living monster inside light radius and opening the
  door reveals it through the existing fog transition flow.
- Client fog overlay can render a supplied closed-door occluder with the existing shadow pipeline.

## Scope and Likely Files

- Server:
  - `server/internal/game/fog_of_war.go`
  - `server/internal/game/fog_of_war_test.go`
- Client:
  - `client/scripts/fog_of_war_overlay.gd`
  - `client/tests/test_fog_of_war_overlay.gd`
- Docs:
  - `docs/plans/v262_2026-06-18-doorway-los-occlusion.md`
  - `docs/as-built/v262_doorway-los-occlusion.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

Asset/plugin decision: reject external plugins/assets. Borrow the existing server closed-door
barrier rules and the code-native fog shadow overlay.

## Test and Bot Proof

- `cd server && go test ./internal/game -run TestFogOfWar`
- `make client-unit`
- `make maintainability`

Manual visual proof, if desired after implementation:

```bash
make bot-visual scenario=68_fog_los_shadow_mask
```

## Open Questions and Risks

- No required questions.
- Risk: client presentation does not currently receive barrier sizes in protocol. This slice should
  avoid a protocol change and keep client proof limited to supplied overlay occluders unless a clean
  existing data path is available.
