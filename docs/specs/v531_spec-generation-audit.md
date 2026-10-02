# v531 — Multi-seed dungeon generation audit

**Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
- **Date:** 2026-10-01
- **Codename:** `generation-audit`
- **Accepted batch assignment:** v531
- **Baseline:** `876d02c872db4b465e92f460266f83ea6245c3c8`
- **Dependencies:** Audit the integrated v522–v530 dungeon-structure behavior before implementation. Spec and plan may be prepared against this baseline; generation measurements and implementation must use the integrated dependency state.
- **ADRs:** [ADR-0001](../adr/0001-technology-stack.md) (server authority and seeded determinism), [ADR-0008](../adr/0008-world-structure-and-dungeon-progression.md) (procedural dungeon generation).

## Purpose

Provide a repeatable multi-seed audit that describes the final dungeon-structure batch across floor profiles and depths. Report distributions for generation success, room/corridor topology, anchor reachability, and generated population bounds, while retaining deterministic output that can be compared across repeated runs. This is quality tooling and test coverage; it does not tune or change runtime generation.

## Scope

- Audit a deterministic, documented set of seeds and the rule-derived supported depth range, including ordinary and boss floors where applicable.
- Emit a stable JSON report containing the audit inputs and aggregate distributions for room counts/sizes, corridor topology, reachable anchors/targets, generated population counts, and generation failures.
- Check structural invariants: each expected floor generates; required anchors are reachable from the generated spawn; required room/corridor and anchor-containment contracts from v522–v530 hold; and population counts stay within bounds derived from loaded generation rules.
- Keep the report deterministic across identical inputs: stable seed/depth ordering, stable bucket ordering, no timestamps or nondeterministic map serialization, and repeat-run equality.
- Make the audit runnable on demand and useful as a bounded focused test. The report must preserve per-seed/depth failures and identify the failing metric rather than reducing failures to an aggregate count.

## Non-goals

- Changing dungeon generation algorithms, RNG streams, seed derivation, gameplay tuning, server authority, protocol payloads, or client presentation.
- Replacing existing focused golden, regression, or progression-blocker tests with statistical checks.
- Claiming exhaustive correctness or production gameplay/performance proof from a finite seed sample.
- Introducing new assets, plugins, or asset pipelines.

## Acceptance criteria

1. A documented opt-in command runs the multi-seed audit and writes a machine-readable JSON report to the fixed ignored path `.artifacts/dungeon-generation-audit.json`; the default seed/depth selection is deterministic and reproducible.
2. Report data records the exact ordered seed and depth inputs, generated-floor count, generation failures with seed/depth/error, and distributions for room/corridor topology, anchors/reachability, and population.
3. The audit checks each generated floor against the integrated v522–v530 structural contracts and rule-derived population bounds. Failed checks are attributed to seed, depth, and invariant.
4. Two runs with identical rules, code, and inputs produce byte-identical report output. A regression check covers ordering and serialization; no wall-clock metadata appears in the report.
5. Summary values are derived from generated output and loaded shared rules. No current tuning number is copied into an unrelated assertion; exact values are used only for report format or determinism contracts.
6. Existing runtime generation, seeded PCG/RNG ownership, protocols, client code, and gameplay data remain unchanged.
7. The focused Go audit/tests pass, the report is inspected and its distributions and limitations are recorded in the as-built handoff, and `git diff --check` passes.

## Likely surfaces

| Area | Likely files |
|---|---|
| Report-only audit and focused checks | `server/internal/game/dungeon_generation_audit_test.go` (new; split helpers into a second file if needed for the 600-line limit) |
| Reproducible command | `Makefile` (only if needed to expose the audit cleanly; otherwise document the direct `go test` invocation) |
| Registries and handoff | `docs/CODEMAP.md`, `docs/specs/v531_spec-generation-audit.md`, `docs/plans/v531_2026-10-01-generation-audit.md`, `docs/as-built/v531_generation-audit.md` |

## Asset decision

- **Adopt:** no new visual assets are required; this audit measures server-generated geometry and entities.
- **Borrow:** if presentation terminology is needed in documentation, reuse the existing vendored KayKit Dungeon kit and current wall/entity presentation contracts.
- **Reject:** new external assets, plugins, downloads, or asset pipelines.

## Focused verification

- Run the audit with the documented fixed seed/depth inputs and the fixed repository-local report path.
- Run the focused `server/internal/game` audit and relevant existing room/corridor/anchor tests.
- Run the report twice with identical inputs and compare bytes; inspect the report for stable ordering, distributions, and any generation or invariant failures.
- Run `git diff --check`.

## Dependencies and integration risks

- v522–v530 change room placement, shapes, motifs, routing, doors, wall continuity, room roles, population, and encounter composition. The audit's structural assertions and distribution report must be finalized against their integrated result, not inferred from this base commit.
- Keep v531 changes in a new focused audit file to avoid conflicts with sibling changes in generation implementation and tests. An optional Make target may overlap coordinator Makefile edits; prefer direct test invocation unless the target materially improves repeatability.
- If the final integrated structures do not expose enough state to verify a stated contract without changing production behavior, report the gap for coordinator review rather than broadening runtime APIs or protocol payloads.

## Security assessment

This slice is local deterministic test tooling. It adds no network/API, persistence, authorization, secret, or user-controlled rendering surface. The report is written only to a fixed repository-local filename through a root-scoped filesystem handle; callers cannot provide a path.
