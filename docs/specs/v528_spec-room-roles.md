# v528 Spec: Dungeon Room Roles

**Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
- **Date:** 2026-10-01
- **Codename:** `room-roles`
- **Baseline:** `876d02c872db4b465e92f460266f83ea6245c3c8`
- **Dependencies:** v522 `rooms-first`, v523 `room-shapes`, v524 `topology-motifs`

## Purpose

Give generated rooms explicit gameplay meaning so level transitions, encounters, rewards, and boss
set pieces land in rooms suited to them. Role assignment and placement policy must be tuned through
validated shared rules, remain reproducible from the existing seed, and retain server-side
reachability guarantees.

## Scope and decisions

- Assign generated rooms one of the roles `entry`, `transition`, `combat`, `reward_objective`, or
  `boss_arena` when applicable. A room has one stable role; all rooms remain traversable regardless
  of role.
- Derive mandatory placement roles from level content: the up-stair/arrival and level teleporter
  use `entry`; ordinary down stairs use `transition`; chests and eligible elite objectives use
  `reward_objective`; boss entities and the boss-floor exit use `boss_arena` when that floor has
  generated rooms. `combat` labels rooms for encounter use without relocating monsters in this
  slice.
- Preserve boss-floor progression order (reward chest before boss, exit gated by boss defeat). For
  boss floors, the boss arena may own the down stairs as already required by the boss sequence.
- Use validated shared rules for role weights, required/optional role constraints, and role-to-
  placement constraints. Use a dedicated deterministic stream on the existing seeded PCG/RNG path
  and stable room ordering so role assignment does not shift unrelated generation draws.
- Keep collision, room connectivity, encounter authority, and progression gating on the server.
  Room roles are generator metadata and do not add protocol fields or client authority.
- Client asset decision: **adopt** the already vendored KayKit Dungeon kit where generated room
  shape needs visual representation; **borrow** existing wall/entity presentation and rectangle
  contracts; **reject** new external assets, plugins, or asset pipelines. No client visual change
  is required solely to ship the role metadata.

## Acceptance criteria

1. Every generated room receives exactly one valid role; assignments are stable for identical seed,
   level, and rules inputs.
2. Role weights and assignment/placement constraints live in shared rules, are schema-validated,
   and receive semantic validation for positive/valid weights, supported role identifiers, and
   constraints that can be satisfied by the configured room and floor model.
3. On ordinary dungeon floors, the up stair/arrival and teleporter occupy an entry room, the down
   stair occupies a transition room, and generated chests/objectives occupy a reward/objective room.
   The combat role is assigned but does not change monster placement in this slice. Optional
   content may be absent under existing content eligibility rules.
4. On boss floors where rooms apply, there is a boss arena role containing the boss and gated exit;
   the existing chest-before-boss progression sequence remains reachable and unchanged.
5. Placement and role assignment preserve obstacle clearance, room geometry, and server-validated
   reachability between player entry and every mandatory progression placement. Impossible role
   constraints fail generation clearly or use a deterministic, validated fallback; they never
   silently produce unreachable content.
6. Existing generation, replay, and world behavior retain the current wire contract. No protocol
   change is justified by this slice.
7. Focused tests cover rule validation, deterministic assignment, role-driven placement, boss-floor
   ordering, and reachable generated floors across representative seeds/depths. Shared validation
   and the relevant focused bot scenario pass.

## Non-goals

- Changing boss cadence, boss identity/content, encounter rules, chest loot, objective rewards, or
  progression gating. Monster placement/composition is unchanged.
- Adding client role labels, room-specific art, room-specific effects, or new assets.
- Changing room shapes, topology motifs, corridor routing, wall generation, monster composition,
  or protocol/persistence contracts owned by prerequisite or sibling slices.
- Replacing the existing seeded PCG flow, introducing nondeterministic gameplay randomness, or
  redesigning generation retry/fallback behavior outside what role feasibility requires.

## Likely surfaces

- Shared: `shared/rules/dungeon_generation.v0.json` and its schema.
- Server: room model/assignment and room-first level population in `server/internal/game/`, plus
  generation rule loading and semantic validation.
- Tests: focused Go generation/rule tests and shared-rule validation coverage.
- Bot/docs: a focused dungeon-generation scenario or existing generation lab; update
  `docs/CODEMAP.md`, this slice's plan, and as-built handoff evidence as warranted.
- Protocol/client: no expected changes.

## Verification and evidence

Use the focused Go package tests covering dungeon generation and validation, shared-rule
validation, and the selected focused bot scenario. A deterministic seed/level sweep should verify
role assignment and reachability without claiming exhaustive all-seed proof. If a role-driven visual
change is made, capture it through the existing Godot tooling and describe the capture as
presentation evidence only; no visual quality or performance claim is required by this spec.

## Dependencies and integration risks

This spec and plan can proceed on the assigned base. Implementation waits until v522–v524 are
integrated into this worktree. The final room model, stable room identity/order, topology access,
and generated placement API must be inspected after transfer before finalizing implementation
touch points. Likely shared-file overlaps include `dungeon_generation.v0.json` and its schema, room
generation/population Go files, associated tests, and `docs/CODEMAP.md`; coordinate these with
v525–v530 before editing shared files. Revalidate this spec against the exact integrated
prerequisites before `/execute`.
