# v262 Plan - Doorway LOS Occlusion

Status: Complete
Goal: Closed barrier interactables block authoritative fog line-of-sight, and supplied door
occluders can use the existing client shadow pipeline.
Architecture: Reuse `InteractableDef.BarrierWhenClosed` for server LOS checks. Keep wall occlusion
unchanged and add a second pass over closed barrier interactables. On the client, extend
`FogOfWarOverlay` so its occluder input can include explicit non-wall blockers without changing the
network protocol.
Tech stack: Go fog filtering, Godot fog overlay unit tests, docs.

## Baseline and Shortcut Decision

Builds on v253 fog filtering, v255 wall shadow masks, and v261 generated door obstacles. This slice
does not add protocol-visible barrier metadata; server authority owns gameplay visibility, and the
client overlay gets a focused unit proof for supplied door occluders.

Asset/plugin decision: reject external plugins/assets. Borrow existing closed-door barrier rules
and the code-native fog shadow overlay.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/game/fog_of_war.go` | Include closed barrier interactables in LOS blocking |
| Modify | `server/internal/game/fog_of_war_test.go` | Add closed/open door fog tests |
| Modify | `client/scripts/fog_of_war_overlay.gd` | Accept supplied occluders in the shadow pipeline |
| Modify | `client/tests/test_fog_of_war_overlay.gd` | Prove a supplied door occluder casts one shadow |
| Modify | `docs/specs/v262_spec-doorway-los-occlusion.md` | Mark complete during close-out |
| Modify | `docs/progress/slice-lifecycle.md` | Add v262 lifecycle row |
| Add | `docs/as-built/v262_doorway-los-occlusion.md` | Record shipped behavior and proof |
| Modify | `PROGRESS.md` | Update current status and next selected autoloop item |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines and grandfathered files stay under their
ratchet allowance.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd` was not touched.
- [x] Did every touched grandfathered file stay at or below its baseline allowance?

Decision:
- [x] Keep server LOS logic in `fog_of_war.go`; no protocol or `main.gd` changes.
- [x] Keep client proof in `FogOfWarOverlay` unit tests.

Verification:
```bash
make maintainability
```

## Task 1 - Server Door LOS

Files:
- Modify: `server/internal/game/fog_of_war.go`
- Modify: `server/internal/game/fog_of_war_test.go`

- [x] Step 1.1: Include closed barrier interactables in fog LOS segment checks.
- [x] Step 1.2: Ignore open doors and non-barrier interactables.
- [x] Step 1.3: Add tests for closed-door hide and open-door reveal.

```bash
cd server && go test ./internal/game -run TestFogOfWar
```

## Task 2 - Client Occluder Pipeline

Files:
- Modify: `client/scripts/fog_of_war_overlay.gd`
- Modify: `client/tests/test_fog_of_war_overlay.gd`

- [x] Step 2.1: Rename/extend internal wall layout storage to occluder layout without changing
  existing `set_wall_layout` callers.
- [x] Step 2.2: Add `set_occluder_layout` for tests/future callers that supply door-sized blockers.
- [x] Step 2.3: Preserve existing wall debug keys and add a focused unit proof for a door occluder.

```bash
make client-unit
```

## Task 3 - Lifecycle Docs

Files:
- Modify: `docs/specs/v262_spec-doorway-los-occlusion.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v262_doorway-los-occlusion.md`
- Modify: `PROGRESS.md`

- [x] Step 3.1: Mark the spec complete.
- [x] Step 3.2: Add v262 lifecycle and as-built notes.
- [x] Step 3.3: Update `PROGRESS.md` current status and leave quest path minimap marker as the
  remaining selected autoloop scope.

```bash
make maintainability
```

## Final Verification

- [x] `cd server && go test ./internal/game -run TestFogOfWar`
- [x] `make client-unit`
- [x] `make maintainability`
- [ ] Autoloop final batch gate: `make ci` (pending after the selected batch)

Manual visual proof, if desired:

```bash
make bot-visual scenario=68_fog_los_shadow_mask
```
