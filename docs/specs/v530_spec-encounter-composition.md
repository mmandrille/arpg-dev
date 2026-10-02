# v530 Spec — Encounter composition by room role

**Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
- **Date:** 2026-10-01
- **Codename:** `encounter-composition`
- **Base:** `876d02c872db4b465e92f460266f83ea6245c3c8`
- **Dependency:** v529 room-role generation must be integrated before implementation.

## Purpose

Make normal generated dungeon encounters use shared, validated composition rules keyed by the
room-role identifiers introduced by v529. Composition should make packs feel suited to their room
while keeping generated population bounded, deterministic, and server-authoritative. Elite leaders
and their guards must have an explicit generated relationship that remains compatible with the
existing elite side-objective behavior.

## Behavior

- Each generated normal-floor room has the v529 room role and receives zero or more packs according
  to data-driven encounter composition rules for that role. v529 supplies canonical room-index/role
  candidates; v530 assigns the original whole-pack multiset across positive-weight candidates using
  usable area × role weight and pack-aware remaining capacity, then records final per-room counts.
  Use v529's role vocabulary and room representation; do not invent a parallel role taxonomy.
- Rules describe weighted pack archetypes using the existing monster `pack_role` vocabulary
  (`frontline`, `ranged`, `flanker`, `swarm`) and constraints for member selection. Monster
  definitions remain selected from the configured pool and role constraints; no tuning value is
  hardcoded in Go.
- Every generated pack has a stable pack identity and is wholly assigned to one eligible room.
  Each member is placed within that room's legal interior using obstacle, body, and reachability
  checks. Members stay clear of corridors by configured body clearance; the configured pack radius
  bounds member spread within the room. Failed placement is bounded and cannot leave partial packs.
- Encounter composition preserves the existing area-derived floor total and exact original pack-size
  multiset. v530 assigns each complete pack occurrence to one eligible positive-weight v529 room
  candidate using usable area × role weight as a preference and pack-aware remaining capacity as a
  hard constraint. It never splits, drops, or adds packs. Existing minimum monster requirements
  remain honored; impossible capacity returns an explicit generation error.
- Elite placement and guard relationships are rule-driven. An elite leader is identified within its
  pack, and configured guards belong to the same pack and room, with valid member-role constraints.
  The selected elite and guards use the existing rarity/stat path and server-owned runtime pack
  metadata. Boss-floor population remains on its existing fixed path and is not re-composed.
- The existing elite-objective generation remains compatible: eligibility still depends on generated
  elite leaders and a reachable chest location is reserved before packs fill rooms when the floor roll
  and geometry permit. If that optional reservation prevents a complete assignment, generation retries
  the unchanged pack-size multiset without it. Chest presence and the kill gate remain server-owned.
- Identical seed, level, and shared rules produce identical room-to-pack assignments and member
  placements. New random choices use named deterministic PCG streams so encounter composition does
  not perturb unrelated dungeon, objective, loot, or boss generation rolls.
- No wire/protocol shape changes are expected. The client remains a renderer of server-authored
  entities and pack relationships.

## Security and authority

- Shared rule values are schema-validated and semantically bounded before generation; the server
  uses those validated limits for all member counts, retries, and placements.
- Clients cannot submit room roles, pack membership, elite status, guard relationships, or generation
  counts. The server derives and owns those values from its seeded rules and generated floor.
- Preserve server-side collision, reachability, objective eligibility, and gameplay authority.
  Do not add secrets or security-sensitive random values; all generation randomness remains on the
  existing deterministic PCG path.

## Out of scope

- v529 room-role taxonomy or room geometry changes. v529's candidate-row API remains intact; v530
  owns complete-pack allocation over those rows and writes the final per-room counts.
- Changes to total population formulas, monster balance, rarity/loot tuning, boss-floor population,
  elite objective completion/unlock rules, or quest behavior.
- New protocol fields, client-side encounter decisions, new monster definitions, or visual effects.
- Replacing existing monster pools, pack aggro, elite aura, objective reward tables, or seeded RNG.
- New external art, plugins, dependencies, or asset pipelines.

## Acceptance criteria

1. Shared encounter-composition rules are schema- and semantic-validated against the exact room-role
   and monster `pack_role` vocabularies; malformed, unknown, negative, or out-of-bound values are
   rejected.
2. Normal room-based floor generation assigns packs by room role and satisfies each selected pack's
   required/allowed member-role constraints. Every member in a pack shares its pack identity and
   room assignment.
3. Generated ordinary-floor monster totals stay within the preexisting area-derived population
   bounds and configured pack-size/count limits across representative rules/seed fixtures. Existing
   minimum monster requirements are retained where the rule set says they apply.
4. Elite leader and guard membership/placement follow shared rules, remain server-authored, and are
   deterministic. Failed attempts do not leave partial packs or bypass placement/reachability
   validation.
5. Existing elite-objective behavior remains valid: no eligible elite leader means no objective
   chest; eligible elite generation still permits the reachable objective chest under its configured
   floor chance. Existing elite kill gating is unchanged.
6. Repeated generation with the same seed, level, and rules yields identical roles, pack composition,
   elite/guard relationships, positions, and objective placement; unrelated named RNG streams remain
   stable where the implementation can preserve that contract.
7. Boss-floor generation, existing pack aggro, runtime rarity scaling, and protocol schemas remain
   unchanged.
8. A focused protocol bot scenario (new or carefully extended) proves a generated normal floor has
   room-role-specific encounter composition and can still progress through the elite objective.
   Assertions are semantic and derive limits from rules rather than pinning today's tuning values.

## Likely surfaces

| Area | Likely files |
|---|---|
| Shared rules and validation | `shared/rules/dungeon_generation.v0.json`, `shared/rules/dungeon_generation.v0.schema.json`, `server/internal/game/dungeon_generation_rules.go`, `server/internal/game/rules.go`, dungeon rule validation helpers |
| Server generation | v529 room/role types and generation files; `server/internal/game/dungeon_gen.go`, `server/internal/game/dungeon_generated_types.go`, `server/internal/game/dungeon_population.go` only if runtime projection needs an existing metadata field |
| Tests and proof | Dungeon generation/room tests, elite objective tests, `tools/bot/scenarios/` and focused protocol bot assertions if needed |
| Registries and evidence | `docs/CODEMAP.md`, `docs/as-built/v530_encounter-composition.md`; spec/plan lifecycle edits only if coordinator requests closeout in this worktree |

No client source change is expected. If future implementation requires client presentation changes,
inspect in-repo Godot scripts/scenes and asset manifests first. **Asset decision:** adopt the already
vendored KayKit Dungeon kit where visuals are needed; borrow existing wall/entity presentation and
rectangle contracts; reject new external assets, plugins, or pipelines.

## Focused verification

- `make validate-shared`
- Focused Go tests for rule validation, room-role composition, placement rollback/bounds,
  deterministic generation, unchanged boss floors, pack aggro, and elite-objective compatibility.
- Focused protocol bot scenario proving composition and objective compatibility.
- Run `make maintainability` if touched files are near or above the 600-line target and the focused
  extraction changes source/test ownership.
- No `make ci` or `make ci-full` in this batch slice worktree.

## Risks and integration notes

- The coordinator supplied v529 as an uncommitted overlay on base `876d02c`. Its internal API
  supplies canonical role-keyed room candidates; v530 performs complete-pack assignment because it
  owns the unchanged pack-size multiset and spatial placement. Final `MonsterCount` values are written
  after successful composition, preserving the floor total and original pack sizes.
- v529 and v530 may both touch `dungeon_generated_types.go`, `dungeon_room_corridors.go`,
  `dungeon_gen.go`, `dungeon_generation.v0.json`, and its schema. Integrate v529 first and rebase
  the file map to its final ownership boundaries.
- Room assignment can change generated positions and therefore deterministic goldens. Prefer
  dedicated encounter RNG streams and semantic tests over broad golden churn.
- Composition must not increase area-derived total population or accidentally make existing minimum
  monster requirements impossible.
- Guard proximity must work with current pack assist semantics while not coupling elite objective
  validity to client events or client-submitted state.

## Open questions

No open questions. Exact role identifiers and room-bound representation are inherited from v529.
