# v530 As-built — Encounter composition by room role

- **Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
**Base:** `876d02c872db4b465e92f460266f83ea6245c3c8`
**Dependency:** v529 room candidates are present as an uncommitted overlay on the same base; there is no dependency commit SHA.

## Delivered behavior

- Validated, shared encounter-composition rules define pack/member roles, pack size and count limits, elite chance, and leader/guard requirements. They use the existing room-role and monster `pack_role` vocabularies.
- Normal generated floors consume the canonical v529 room-index/role candidate rows and assign the unchanged seeded pack-size multiset as complete room-local packs. Successful placement writes the final per-room `MonsterCount`; placement errors do not leave partial packs.
- Position checks honor the room interior, obstacles, body clearance from corridors, pack spread, and reachability. Named deterministic streams keep composition and objective placement independent from other generation rolls.
- The configured formation-attempt limit applies independently at each candidate center, so one unproductive center does not prevent the bounded search from checking later centers.
- The optional elite-objective chest location is reserved before packs fill the reward room. If that reservation makes assignment impossible, generation retries the same pack-size multiset without the optional reservation and suppresses the chest for that floor.
- Boss-floor population and protocol contracts remain unchanged. No client authority or new client surface was added.

## Changed and inherited paths

v530-owned changes add `server/internal/game/dungeon_encounter_composition.go` and its tests, `server/internal/game/dungeon_generation_reachability.go`, plus edits to `dungeon_gen.go`, `dungeon_room_layout.go`, `dungeon_elite_objective.go`, rules and validation, generation tests, shared generation rules/schema, CODEMAP, and `tools/bot/scenarios/14_dungeon_monsters.json`.

This detached checkout also contains the coordinator-supplied v522–v529 overlay. Its inherited paths include room generation/shape/topology/corridor/door/wall/role/population files, the v522–v529 specs/plans/as-built notes, affected goldens, scenario 124, and shared rules. The coordinator compared changed paths against its current overlay before integration. `PROGRESS.md` and the lifecycle index were updated during batch closeout.

## Verification

| Check | Result |
|---|---|
| Focused Go suite for room population, encounter composition, room roles, shapes, objectives, and monster generation | PASS |
| Deep-floor reachability/determinism and elite-objective focused suite | PASS |
| Shared validation using the existing workspace venv: `/Users/mmandrille/git/arpg-dev/.venv/bin/python tools/validate_shared.py` | PASS, 2,271 checks |
| CODEMAP validation | PASS |
| `make maintainability` | PASS |
| Bot `reachable_dungeon_obstacles` (scenario 28) | PASS |
| Bot `elite_minion_pack_ai` (scenario 77) | PASS after objective reservation/fallback changes |
| Bot `dungeon_monsters` (scenario 14) | PASS with test-only debug progression fixture |
| `git diff --check` | PASS |
| Combined `make ci` | Not run in slice worktree; coordinator-owned after integration |

The `make validate-shared` wrapper could not provision its virtual environment because the configured package index did not provide `setuptools>=68`; the exact underlying validator passed with the already provisioned workspace environment. Bot scenarios were run through `scripts/bot_local.sh` with the existing environment because the wrapper's package provisioning has the same limitation.

Bot evidence proves the named protocol flows complete under their fixtures. The scenario 14 debug progression was adjusted so the character survives the newly generated packs; it is not a balance or broad gameplay-quality claim. No visual capture was needed for this server-generation slice.

## Integration notes

- v530 overlaps v529 in generation rules, generated types, layout, and generation tests. Preserve the v529 canonical-candidate API; v530 owns assignment and final room counts.
- Objective reservation is optional and retries use the exact same pack-size multiset. If even the fallback placement fails, generation returns an explicit combined error.
- Focused checks are complete. The coordinator ran combined `make ci` after integrating all accepted batch slices; it passed.
