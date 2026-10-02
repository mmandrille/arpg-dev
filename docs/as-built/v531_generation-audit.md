# v531 Handoff — Multi-seed dungeon generation audit

- **Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
- **Date:** 2026-10-02
- **Slice:** v531 `generation-audit`
- **Base:** `876d02c872db4b465e92f460266f83ea6245c3c8`
- **Dependencies:** Coordinator-supplied v522–v530 overlay on the assigned detached worktree; no separate dependency commit SHA.
- **Scope:** Report-only tests/tooling and documentation. No runtime generation, gameplay rules, protocols, or client code changed.

## Result

- Added an opt-in report-only audit at `server/internal/game/dungeon_generation_audit_test.go` with report types and stable histogram helpers in `server/internal/game/dungeon_generation_audit_report_test.go`.
- Audits 20 fixed seeds (`generation-audit-01` through `generation-audit-20`) across levels −1 through −12, including boss floors −5 and −10. Depths are selected from the configured floor profiles and boss cadence. Closed doors count as reachable when at least one cardinal approach is reachable from the generated start, matching the player's ability to open the door from that side.
- The report includes per-seed/level outcomes, generation errors, invariant findings, room/shape/role/area histograms, graph edge/loop/degree and route-segment distributions, target reachability by kind, and rule-derived population/pack distributions.
- It records generator and invariant failures but does not fail the test for those findings by default. Set `ARPG_DUNGEON_GENERATION_AUDIT_STRICT=1` alongside the report opt-in to make the audit test fail after writing the report when either list is non-empty. The existing fail-fast progression sweep remains unchanged.
- Report output uses the fixed repository-local `.artifacts/dungeon-generation-audit.json` path, a single accepted opt-in value (`ARPG_DUNGEON_GENERATION_AUDIT=1`), and root-scoped file operations. No caller-supplied path is accepted.

## Audit evidence

The integrated report covers 240 seed/level pairs: **240 generated successfully, with zero generation errors**. It recorded **zero invariant findings and zero unreachable generated targets**. The sample includes 40 boss-floor outcomes and 200 ordinary floors. It observed 6,415 monsters, 240 up stairs, 240 down stairs, 80 teleporters, 109 treasure chests, and 141 wooden doors; all 7,225 generated targets were reachable under the audit's structural reachability check.

The initial worker handoff report covered 200 pairs and recorded 35 pack-placement failures. Coordinator integration added complete-pack backtracking and corrected room-capacity/objective placement constraints and floor profiles. The final 240-pair audit then generated every floor; the initial report is preserved at `.artifacts/dungeon-generation-audit-v531-worker-handoff.json` for comparison.

Population targets matched shared rule-derived counts on all 240 floors; pack counts stayed within the configured profile ranges. The complete distributions are in the JSON report.

Selected distributions:

| Metric | Distribution |
|---|---|
| Rooms per generated floor | boss/no-room: 40; 5 rooms: 137; 6 rooms: 63 |
| Corridor edges per floor | boss/no-room: 40; 5: 58; 6: 106; 7: 36 |
| Corridor loop edges | boss/no-room: 40; 1: 85; 2: 115 |
| Room shapes | rectangle: 686; L: 180; T: 141; cross: 56 |
| Room roles | entry: 200; transition: 200; combat: 402; reward/objective: 261 |
| Reachable targets | 7,225 total across 240 floors; 0 unreachable |
| Population target per floor | boss: 3 × 40; ordinary target 20 × 43, 22 × 17, 30 × 116, 32 × 24 |

Both runs produced 278,452 bytes and were byte-identical. SHA-256 for both:
`3496933cb65c18a8f5f326d7abecde182f1acc31e1b3747b5b6bda707cf6f7af`.

## Verification

- `cd server && ARPG_DUNGEON_GENERATION_AUDIT=1 go test ./internal/game -run '^TestDungeonGenerationAudit' -count=1 -timeout 20m` — **PASS**, twice; 795.981s and 836.926s. Reports are preserved at `.artifacts/dungeon-generation-audit-run1.json` and `.artifacts/dungeon-generation-audit-run2.json`; `.artifacts/dungeon-generation-audit.json` contains the final report.
- Byte comparison and SHA-256 of the two JSON reports — **PASS**, identical length and hash `3496933cb65c18a8f5f326d7abecde182f1acc31e1b3747b5b6bda707cf6f7af`.
- Focused room-role, shape, topology, population, and encounter-composition tests — **PASS**: `go test ./internal/game -run '^Test(DungeonRoomRoleRulesValidation|AssignDungeonRoomRolesIsDeterministicAndRuleDriven|GeneratedDungeonAnchorsFollowRoomRoles|DungeonRoomShapes_.*|RoomConnectionEdges_.*|DungeonRoomPopulation.*|.*EncounterComposition.*)$' -count=1 -timeout 10m`.
- Refreshed obstacle, stair, and guarded-chest goldens and reran the selected structural/golden suite — **PASS**; the deterministic room-door generation/opening test also passed.
- `make bot scenario=124_room_doors` — **PASS**; the dedicated lab spawn is adjacent to the generated threshold door and reaches level −2.
- `make bot-visual scenario=124_room_doors` — **PASS**, replay matched. Godot emitted resource/ObjectDB leak warnings during shutdown.
- Combined `make ci` passed in 21m23s and `/finish` lifecycle closeout is complete. `make ci-full` was not run.

## Exact focused commands

```bash
cd server && ARPG_DUNGEON_GENERATION_AUDIT=1 go test ./internal/game -run '^TestDungeonGenerationAudit' -count=1 -timeout 20m
cd server && ARPG_DUNGEON_GENERATION_AUDIT=1 ARPG_DUNGEON_GENERATION_AUDIT_STRICT=1 go test ./internal/game -run '^TestDungeonGenerationAuditReport$' -count=1 -timeout 20m
cd server && go test ./internal/game -run '^Test(DungeonRoomRoleRulesValidation|AssignDungeonRoomRolesIsDeterministicAndRuleDriven|GeneratedDungeonAnchorsFollowRoomRoles|DungeonRoomShapes_.*|RoomConnectionEdges_.*|DungeonRoomPopulation.*|.*EncounterComposition.*)$' -count=1 -timeout 10m
```

## Handoff inventory

v531-owned changes:

- Added: `server/internal/game/dungeon_generation_audit_test.go`; `server/internal/game/dungeon_generation_audit_report_test.go`; `docs/as-built/v531_generation-audit.md`.
- Modified: `docs/specs/v531_spec-generation-audit.md`; `docs/plans/v531_2026-10-01-generation-audit.md`; `docs/CODEMAP.md`.
- Deleted: none.
- Ignored evidence to preserve: `.artifacts/dungeon-generation-audit.json`; `.artifacts/dungeon-generation-audit-run1.json`; `.artifacts/dungeon-generation-audit-run2.json`; `.artifacts/dungeon-generation-audit-v531-worker-handoff.json`.

The dependency overlay also contained uncommitted v522–v530 changes; all paths were compared and integrated. `PROGRESS.md` and lifecycle/codename registries were updated during batch closeout.

## Evidence limits and unresolved items

The audit is a fixed 20-seed × 12-level sample, not exhaustive generation or gameplay proof. Its reachability results cover generated targets under the structural graph/grid model; the room-door protocol bot adds one end-to-end open-and-descend example. The audit does not establish player-perceived map quality, combat balance, or performance.
