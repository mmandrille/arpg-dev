# v265 As-Built - Door Fog And Toggle

Date: 2026-06-18
Spec: [`docs/specs/v265_spec-door-fog-and-toggle.md`](../specs/v265_spec-door-fog-and-toggle.md)
Plan: [`docs/plans/v265_2026-06-18-door-fog-and-toggle.md`](../plans/v265_2026-06-18-door-fog-and-toggle.md)

## Shipped Behavior

- Shared rules now make `wooden_door` barriers 1.6 world units wide, and generated dungeon door
  gaps use the same 1.6 width.
- Open barrier interactables are actionable, so clicking an open wooden door closes it again. The
  close emits `interactable_state_changed` with `state: "closed"`.
- Auto-approach now evaluates all in-range action goals and prefers same-side goals for closed
  barrier interactables, preventing closed-door clicks from walking around a wall to open from the
  far side when a same-side interaction point is reachable.
- Treasure chests and non-barrier interactables keep their existing one-shot/open behavior.
- The Godot client loads interactable barrier sizes from shared rules and uses them for door panel
  width, pick collider width, and fog occluder size.
- Closed barrier interactables are synced into `FogOfWarOverlay` as dynamic extra occluders. Opening
  a door removes that occluder; closing the door adds it back.
- The client bot fog assertion now supports `extra_occluder_count`, and event matching supports
  `event_state` without conflicting with entity-state selectors.
- `tools/bot/scenarios/client/73_door_fog_toggle.json` proves closed -> open -> closed fog occluder
  behavior in a fog-enabled `door_lab` session.

## Boundaries

- No imported door art, external asset pipeline, Godot addon, or protocol version bump shipped.
- No durable explored-map memory, reconnect/resume map memory, or server visibility formula change
  shipped.
- Existing server fog LOS semantics for closed doors were already present; this slice aligned client
  fog presentation with that behavior.

## Verification

```bash
make validate-shared
go test ./internal/game -run 'TestClosedDoorAutoApproachPrefersPlayerSide|TestDoorLabClosedDoorPreventsPassageUntilActivated|TestOpenDoorCanBeClosedAgain|TestGeneratedDungeonDoorGeneration|TestGeneratedDungeonDoorsPopulateAsClosedInteractables|TestFogOfWarDeltasRevealMonstersWhenClosedDoorOpens|TestTreasureChestOpensOnceAndDropsLoot' # from server/
go test ./internal/game # from server/
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_fog_of_war_overlay.gd
make client-unit
HEADLESS=1 make bot-visual scenario=73_door_fog_toggle
make maintainability
```

All focused commands passed on 2026-06-18. The v264-v265 batch `make ci` gate passed on
2026-06-18.
