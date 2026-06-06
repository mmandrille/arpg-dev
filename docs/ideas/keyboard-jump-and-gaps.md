# Idea: keyboard jump and gap traversal

Status: idea / deferred

This note captures a possible movement feature to consider later. It is not an approved slice,
spec, or implementation plan.

## Concept

Add a `jump` action that is only available from the keyboard at first. The default binding would be
the space bar.

The initial gameplay use case is traversal over a gap, river, or hole that divides a contained test
arena into two parts. The player must jump from one side to the other to continue.

## Suggested first slice shape

- Add a server-authoritative `jump_intent`.
- Bind space bar in the Godot client to send the jump intent.
- Keep mouse-based jumping out of scope until the core mechanic is proven.
- Treat jump as deterministic traversal, not full free-form vertical physics.
- Allow jumping only across an explicit obstacle type such as `gap`, `river`, or `hole`.
- Keep walls, closed doors, monsters, and other solid obstacles non-jumpable.
- Cancel queued auto-navigation when a jump intent is accepted, similar to manual movement.
- Play client-only jump presentation after the authoritative result is accepted.

## Bot scenario idea

Create a world preset such as `jump_gap_lab`:

- A walled cage contains the whole test area.
- A river, gap, or hole divides the cage into two sections.
- Normal movement toward the divider is blocked.
- A space-bar jump crosses the divider and lands the player on the other side.
- The far side contains a simple objective, such as loot or a target dummy, to prove the player can
  continue acting after the jump.

The protocol bot should prove:

- walking into the divider does not cross it;
- `jump_intent` crosses it from a valid position and direction;
- invalid jumps are rejected or leave the player in place;
- reconnect resume and replay reconstruct the same final position;
- visual replay shows the divided arena and jump traversal.

## Open questions

- Should jump direction come from current movement input, current facing, or an explicit payload?
- Should jump have a fixed distance, rule-defined distance, or per-character/item modifiers?
- Should gaps be rectangular world obstacles, line segments, or a richer terrain type?
- Should pathfinding consider jumpable gaps in a future slice?
- What client animation should represent the first version if no production jump clip exists yet?

## Non-goals for the first version

- No mouse-click jump command.
- No platforming physics.
- No attacks, pickups, or interactions while jumping.
- No jump over monsters or closed doors.
- No character stat system for jump distance.
- No production art requirement.
