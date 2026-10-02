# v531 Plan — Multi-seed dungeon generation audit

**Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
- **Date:** 2026-10-01
- **Spec:** [v531 generation-audit](../specs/v531_spec-generation-audit.md)
- **Baseline commit:** `876d02c872db4b465e92f460266f83ea6245c3c8`
- **Prerequisites:** v522–v530 dungeon-structure slices. Documentation review can proceed now. Implementation starts only after the coordinator integrates those slices into this worktree and records the resulting commit.

## Goal

Produce a deterministic, report-only multi-seed audit of the integrated dungeon structure behavior, with explicit quality assertions for connectivity, room/corridor topology, anchor reachability, population bounds, and generation failures. No runtime generation or gameplay contract changes.

## Review against spec

- **Scope:** tooling/tests only; no runtime, shared-rule, protocol, or client behavior changes.
- **Acceptance mapping:** deterministic output and ordering → repeat-run report comparison; topology/connectivity/anchors → generated-level checks using integrated helpers/contracts; population → bounds derived from the active rules; failure accounting → per-seed/depth report entries and invariant findings.
- **Determinism:** audit seeds are fixed or deterministically generated from explicit parameters; reuse the existing seeded PCG-based `GenerateDungeonLevel`; do not add randomness. Sort all output collections and omit timestamps.
- **Authority/data:** generation stays server-authoritative; bounds derive from loaded shared generation/population rules. No tuning values are introduced.
- **Protocol/schema/golden:** none expected. Report is diagnostic audit output, not a gameplay golden.
- **Assets:** adopt none; borrow the vendored KayKit Dungeon kit and existing wall/entity presentation contracts only if documentation needs visual terms; reject external assets, plugins, and pipelines.
- **Maintainability:** keep implementation in a focused new Go audit file below 600 lines; extract report aggregation/formatting into a second focused file if needed. Do not grow the large `dungeon_gen.go` coordinator.
- **Security:** local test/report tooling only. The report path is fixed under repository `.artifacts/` and written with a root-scoped filesystem handle; no caller-supplied path or new dependency is introduced.

## Ordered tasks

### 1. Rebase audit design on integrated dependencies

- [x] Confirm the worktree contains the v522–v530 coordinator overlay on base `876d02c872db4b465e92f460266f83ea6245c3c8`; no separate integration commit SHA exists.
- [x] Read their as-built notes and changed generation/test surfaces; inspect the final room/corridor, door, role, population, and encounter contracts.
- [x] Confirm generated rooms, corridor edges/routes, anchors, role budgets, pack identities, and shared reachability helpers expose enough state to audit without production API changes.
- **Gate:** satisfied by the coordinator-supplied v522–v530 overlay; no separate integration commit SHA exists.

### 2. Add deterministic audit model and report-only runner

- [x] Add `server/internal/game/dungeon_generation_audit_test.go` plus `server/internal/game/dungeon_generation_audit_report_test.go`, keeping each below 600 lines.
- [x] Load current shared rules; use fixed `generation-audit-01` through `generation-audit-20` seeds and derive the depth span from finite profile endpoints, open-ended profile starts plus one configured boss cadence, and the configured boss-floor sample.
- [x] Generate each seed/depth pair through `GenerateDungeonLevel` and collect per-floor audit outcomes. Continue after individual failures so the report preserves the full failure picture.
- [x] Aggregate stable histograms/distributions for room/corridor counts, room roles/topology, anchor kinds and reachability, monster/encounter population, and per-seed/depth generation failures.
- [x] Derive admissible population bounds from active rules and report counts and bound outcomes without duplicating tuning defaults. Excluded the optional elite-objective chest from the ordinary chest bonus because it is appended after pack-size selection.
- [x] Ensure report ordering is deterministic; two full runs produced byte-identical JSON without timestamps/host-specific values.
- [x] Keep the seed set and depth derivation fixed and bounded; write to a fixed filename under the repository's ignored `.artifacts/` directory using root-scoped file operations.
- **Focused check:** `cd server && ARPG_DUNGEON_GENERATION_AUDIT=1 go test ./internal/game -run '^TestDungeonGenerationAuditReport$' -count=1`; this writes `.artifacts/dungeon-generation-audit.json` under the repository root.

### 3. Verify repeatability and invariant coverage

- [x] Run the same audit twice and compare the JSON bytes.
- [x] Add focused tests for numeric/label ordering, graph connectivity/invalid edges, and target-kind classification without pinning incidental distributions as gameplay goldens.
- [x] Run focused room/corridor/anchor and generation reachability tests against the integrated final state. Initial worker handoff failures were resolved during coordinator integration; the final focused suite and combined CI pass.
- [x] Inspect generated report counts/distributions. The initial worker handoff report covered 200 pairs and recorded 35 pack-placement failures; after coordinator solver/profile fixes, the final 240-pair audit reports zero generation failures and zero invariant findings. The earlier report is preserved at `.artifacts/dungeon-generation-audit-v531-worker-handoff.json`.
- **Focused checks:**
  - `cd server && go test ./internal/game -run '^TestDungeonGenerationAudit' -count=1`
  - `cd server && go test ./internal/game -run 'TestRoomCorridorLayout_|TestPlaceRoomCorridorLayout|Test.*Reachability' -count=1`
  - Repeat the opt-in report command, preserving the first output separately, and compare report bytes.

### 4. Document report and handoff

- [x] Add a small as-built report at `docs/as-built/v531_generation-audit.md` with exact inputs, aggregate distributions, commands/results, evidence limits, and resolved worker-handoff findings.
- [x] Add the audit ownership to the dungeon generation row in `docs/CODEMAP.md`.
- [x] Leave `PROGRESS.md`, lifecycle, and codename registries to the coordinator for batch completion and `/finish` closeout.
- [x] Run `git diff --check`; report all changed, deleted, untracked, and ignored evidence paths to the coordinator.

## File map and ownership

| Path | Change | Ownership / conflict |
|---|---|---|
| `server/internal/game/dungeon_generation_audit_test.go` | New report-only audit runner, aggregation, invariant attribution | v531-owned; no sibling expected to edit this file. |
| `server/internal/game/dungeon_generation_audit_report.go` | Optional new focused report serializer/helpers | Add only if needed to maintain the 600-line limit. |
| `Makefile` | Optional command alias | Avoid unless helpful; potential cross-slice/coordinator conflict. Direct focused Go test is acceptable. |
| `docs/CODEMAP.md` | Register the audit file | Likely shared doc conflict; integrate surgically after sibling updates. |
| `docs/as-built/v531_generation-audit.md` | Worker handoff evidence | New file; coordinator reconciles during final batch closeout. |

## Bot, visual, and performance evidence

No bot scenario, screenshot, real renderer capture, or performance comparison is required: this slice neither changes client presentation nor makes runtime performance claims. Existing generated geometry is inspected through the server-side audit report and focused tests.

## Integration and handoff

- Work only in this detached worktree; do not create a branch, commit, push, run `/finish`, or modify the coordinator checkout.
- The worktree starts at the assigned base, with the coordinator-supplied v522–v530 changes now copied in as an uncommitted overlay. Record that overlay state as the dependency baseline; no integration commit SHA exists yet.
- Coordinator runs combined `make ci` only after every accepted slice is integrated. v531's final gate is focused slice verification.
- Handoff must include the base and dependency commit, complete changed/deleted/untracked paths, ignored report artifact paths, exact commands/results, report distributions, evidence limits, unresolved acceptance criteria, and likely integration conflicts.
