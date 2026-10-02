# v523 Spec — Room Shapes

**Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
- **Date:** 2026-10-01
- **Base:** `876d02c872db4b465e92f460266f83ea6245c3c8`
- **Dependency:** v522 rooms-first generation; specification and plan may proceed before integration, implementation may not.

## Purpose

Add a small, shared-data-driven vocabulary of room footprints to ordinary dungeon generation so generated rooms are not all rectangular. The server continues to generate and validate authoritative geometry using its seeded PCG path.

## Scope and decisions

- Add schema-backed shape definitions/weights to `shared/rules/dungeon_generation.v0.json` and its schema. Include rectangle, L, T, cross, and arena forms; keep each shape's geometry finite, bounded, and expressed as a small set of normalized connected regions or equivalent declarative geometry.
- Select shapes deterministically from the existing room-generation RNG stream. No unseeded randomness or per-request override is introduced.
- Apply each footprint to ordinary dungeon rooms after v522's rooms-first integration. Keep anchor placement, room connections, door placement, room bounds, and generated reachability valid for every selectable shape.
- Keep boss floors, legacy divider-only layout, wire protocol, encounter authority, and client rules interpretation unchanged.
- **Security:** generation remains server-authoritative. Validate the checked-in rule shape vocabulary and its numeric bounds; bound geometry complexity by the finite catalog and current room count. Do not introduce user-controlled geometry, dependencies, secrets, or nondeterministic randomness.

## Non-goals

- No new art, external assets, plugins, asset pipelines, client rendering features, or balance tuning beyond shape selection/footprint parameters.
- No protocol changes, new room roles, encounter/combat changes, obstacle motifs, door redesign, or general dungeon-generation rewrite.
- No attempt to claim generation performance, gameplay quality, or universal balance from visual captures.

## Asset decision

- **Adopt:** the already vendored KayKit Dungeon kit where existing dungeon visuals render.
- **Borrow:** existing wall/entity presentation and current rectangular room-bound contracts as the starting geometry/rendering conventions.
- **Reject:** new external assets, plugins, and asset pipelines. This slice adds no client art.

## Observable acceptance criteria

1. Shared schema accepts the intended finite shape vocabulary and rejects unknown shape IDs, malformed geometry, out-of-range normalized coordinates, invalid/non-positive weights, and unbounded/empty region lists.
2. Ordinary dungeon generation can deterministically select rectangle, L, T, cross, and arena footprints from schema-backed weights without adding any non-PCG randomness.
3. For deterministic test fixtures that force each shape in turn, every resulting room footprint stays within the configured floor and required perimeter margin; derived room walls/doors stay within the same bounds.
4. Every forced-shape fixture produces a navigable generated level: player spawn, stairs, teleporters and required anchors remain reachable under the existing authoritative reachability validator; every room's walkable footprint is connected to its valid room exits.
5. Repeating a seed, level, and rules fixture produces the same room shapes and geometry. Existing rectangle-only golden/layout behavior is preserved when a rectangle-only fixture is used; ordinary mixed-weight seeded generation is stable after the change.
6. Boss floors and legacy room-layout behavior remain unchanged; no wire protocol files change.
7. Focused tests and shared-rule validation pass. Run the existing `wall_floor_dungeon_rollout` visual scenario when v522 is available; report its bounds: it shows the client rendering a generated depth-1 wall/floor layout, but does not prove all shape variants, reachability, or player navigation.

## Likely surfaces

- Shared: `shared/rules/dungeon_generation.v0.json`, `shared/rules/dungeon_generation.v0.schema.json`.
- Server: `server/internal/game/dungeon_generation_rules.go`, `server/internal/game/dungeon_profiles.go`, `server/internal/game/dungeon_room_corridors.go`, room geometry helpers, and focused room-corridor tests; rules decode in `server/internal/game/rules.go` if the updated fields require it.
- Tooling/docs: `docs/CODEMAP.md`; `docs/as-built/v523_room-shapes.md` at handoff. Progress/lifecycle registry changes belong to coordinator closeout unless needed to keep an existing validator consistent.
- Client: no code expected; existing scenario `tools/bot/scenarios/client/79_wall_floor_dungeon_rollout.json` is visual verification only.

## Verification

- `cd server && go test ./internal/game -run 'Test.*Room.*(Shape|Reachability|Bounds|Determinism)' -count=1` (final exact names to match implementation).
- `make validate-shared`.
- `HEADLESS=1 make bot-visual scenario=wall_floor_dungeon_rollout` once the v522 dependency is integrated and its room-placement behavior is available.
- Add a deterministic seed/level/rules sweep for bounds and reachability only if focused shape fixtures do not adequately exercise generated combinations; keep sweep scope bounded and reproducible.

## Dependencies and integration risks

- v522 rooms-first generation must be integrated before implementation, because it relocates anchors into generated rooms and is the geometry baseline this slice extends.
- Expected overlap with v522 in `dungeon_generation.v0.json`/schema, `dungeon_generation_rules.go`, room packing and room tests. Rebase/compare the latest integrated state before coding; preserve the v522 anchor-in-room contract.
- Sibling slices may touch room topology, corridor routing, doors, wall continuity, room roles/population, and audit files. Keep this slice focused on footprint representation and deterministic selection; coordinate geometry helper ownership during integration.
- Adding shared rule fields must preserve strict schema validation and Go decode parity. No golden or wire change is expected unless existing generated-layout goldens intentionally encode default mixed shapes; document and narrowly update any such fixture if necessary.

## Open questions

None blocking at spec stage. Exact normalized-region representation and shape weights are implementation details for the plan, constrained by the acceptance criteria and bounded catalog above.
