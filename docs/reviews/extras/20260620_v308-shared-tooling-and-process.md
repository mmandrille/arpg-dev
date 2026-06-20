# arpg-dev — Shared contracts, Python tooling & SDD process review at slice **v308**

**Date:** 2026-06-20
**Scope:** Shared protocol/rules/golden contracts, the v302–v308 World Detail / Navigation data model, Python validator/bot tooling, CI orchestration, model preview catalog, SDD cadence, ADR drift, and review/refactor handoff.
**Baseline:** `main` at `e5862466` (`Merge branch 'codex/world-detail-navigation'`). Worktree clean before this review. Ad-hoc review requested before `$refactor`; prior cadence review was v301.
**Stats:** 73 protocol schema files, 40 rules files, 70 golden files, 105 protocol bot scenarios (v301: 101), 87 client bot scenarios (v301: 85), `tools` 16,545 Python lines. `make validate-shared` = 1,435 checks (v301: 1,395). Batch contract footprint: `worlds.v0.json` +194, `dungeon_generation.v0.json` +6, `monsters.v0.json`/`skills.v0.json` +1 each, `dungeon_obstacles.json` golden 163 lines changed, protocol v8 schemas +3 each.
**Overview:** [`../20260620_v308-overview.md`](../20260620_v308-overview.md)

---

## Summary

The v302–v308 batch is the cleanest contract/tooling work in the repo. Every new gameplay knob
(water, holes, solid-kind weights, flying navigation, leap obstacle-ignore, line-of-sight blocking) is
data-driven and schema-backed with tight enums, ranges, `additionalProperties: false`, and
mutual-exclusion constraints; `validate_shared.py` passes 1,435 checks against the live tree. Protocol
v8 changes are genuinely additive/optional, SDD coverage is 100% (spec+plan+as-built+lifecycle for all
seven slices), and three of four prior v301 recs were addressed. The real defects are
documentation/process drift, not code: `CODEMAP.md` was never updated for the four new backend files,
ADR-0008 was not touched despite this batch materially extending the world-structure model, the
`dungeon_obstacles.json` golden is a Go-only contract (not cross-language as its framing implies), and
the "10-second budget" hard rule in CLAUDE.md now contradicts the 15s ceiling / 10s default in code and
`scenario-catalog.md`.

## 1. Architecture / contracts

- **[Strength] Rules-as-data is exemplary.** New world config (`shared/rules/worlds.v0.schema.json:41`),
  generation knobs (`solid_kind_weights`/`water`/`holes` in `dungeon_generation.v0.schema.json`), the
  flying trait (`shared/rules/monsters.v0.schema.json:55`, enum `["grounded","flying"]` +
  `preferred_min_range`), and leap (`shared/rules/skills.v0.schema.json:364`,
  `ignore_obstacle_kinds` enum `["water","hole"]`) are pure data. The worlds-schema mutual-exclusion
  rules (`shared/rules/worlds.v0.schema.json:53`) forbid `kind` / `blocks_line_of_sight` on
  monster/item/teleporter spawns — a constraint class that prevents malformed-content bugs.
- **[Med] Protocol v8 was edited in place where a v9 bump was arguably owed.** Both
  `shared/protocol/session_snapshot.v8.schema.json` and `state_delta.v8.schema.json` added `kind` and
  `blocks_line_of_sight` to the `wall` `$def`. Neither field is in `wall.required`
  (`shared/protocol/session_snapshot.v8.schema.json:60`), so an old client validating an old payload
  still passes — backward-compatible for consumers. But `wall` has `additionalProperties: false`, so a
  *new* server emitting the new fields against the *pre-change* v8 schema would have failed validation:
  v8 is now silently a second dialect. The "schema changes require a version bump" invariant reads as
  v9-owed; the batch took the coordinated client+server escape hatch instead, with no spec note
  justifying it. Defensible for a co-shipped tree, but undocumented.

## 2. Technical correctness

- **[Strength] Validation is automatic and green.** `validate_instances` globs every `*.v0.json` and
  validates against its schema (`tools/validate_shared.py:140`), so the new blocks are checked without
  new Python; run result `VALIDATION OK: 1,435 checks passed`.
- **[Strength] No new stat-key collision,** so agent rule #5 correctly produced no new `cross_checks()`.
- **[Med] One missing semantic cross-check.** The golden gained `minimum_water_count`,
  `minimum_hole_count`, and `solid_kinds` (`shared/golden/dungeon_obstacles.v0.schema.json:30`), but the
  Python guard at `tools/validate_shared.py:1449` still only asserts `minimum_generated_wall_count > 0`.
  The new floors are presence/type-checked by schema but not cross-linked to
  `dungeon_generation.water.target_count` / `holes.target_count`, so a "generation says 0 water but
  golden demands 2" drift would pass.
- **[Med] The obstacles golden is Go-only, not cross-language.** `shared/golden/dungeon_obstacles.json`
  is consumed only by `server/internal/game/dungeon_obstacles_golden_test.go`; no GDScript test reads
  it. This is a legitimate Go determinism golden under the Test Locking Policy, but the CLAUDE.md
  "Golden fixtures are cross-language contracts" framing overstates this one — the label is wrong, not
  the pinning.

## 3. Test quality / bot scenarios

- **[Strength] Scenario assertions genuinely prove each mechanic.** 99 (flying) requires the bat to move
  ≥2.0 and land within 2.5 across blockers a grounded monster cannot cross; 100 (leap) asserts a landing
  coord derived from the authored `barbarian_leap_obstacle_lab` layout (not a tuning pin) with a 0.4
  tolerance; 101 gates rock+column+rubble present; 102 asserts a monster reads 0 behind a blocker then
  appears from a clear angle — real line-of-sight proof. None hardcode a rule value as an equality.
- **[Med] The "10s hard rule" is now a 15s ceiling / 10s default.** `tools/bot/run.py:54` sets
  `MAX_SCENARIO_ELAPSED_S = 15.0` with per-scenario `max_elapsed_s` override; commit `20327d91`
  reworded `docs/progress/scenario-catalog.md:17` to a "default 10s / larger only when the generated
  route is the behavior under test" model. CLAUDE.md rule #12 still reads as an absolute 10s cap — a
  live contradiction between two governance docs.
- **[Low] Scenario 102 decouples a 240-tick (24s sim) path from the 15s wall-clock budget** with no
  explicit `max_elapsed_s`; the most likely scenario to brush the ceiling if pathfinding slows.

## 4. Maintainability

- **[Med] Both Python monoliths drifted toward their caps without a split.**
  `tools/validate_shared.py` 3,017 → 3,038 (+21) and `tools/bot/run.py` 4,210 → 4,231 (+21)
  (`.maintainability/file-size-baseline.tsv:33`, `:35`) — within the +25 allowance, but the standing
  v301 split rec was not acted on and both moved the wrong direction.
- **[Strength] Extraction-coupling stayed at zero.** `grep` finds no `helpers=globals()` sites and the
  ratchet reports 0 — agent rule #6 respected; the four new scenarios added data-routed assertion types,
  not new global-laundered modules.
- **[Strength] New backend domains went to focused files,** not into `dungeon_gen.go`.

## 5. Documentation / SDD process

- **[Strength] SDD coverage is complete and clean.** All seven slices have spec + plan + as-built and a
  lifecycle row with consistent "isolated v302-v308 batch `make ci` green before merge" status; the
  scenario catalog was updated.
- **[High] `docs/CODEMAP.md` is stale.** The "Dungeon generation" row (`docs/CODEMAP.md:15`) lists only
  `dungeon_gen.go, level.go, pathfind.go`; none of `dungeon_water.go`, `dungeon_holes.go`,
  `dungeon_obstacle_variety.go`, `monster_navigation_traits.go` appear, nor do scenarios 99–102. The
  documented "domain → files index for loading focused context" is incomplete for the headline batch.
- **[Med] ADR-0008 drift.** `docs/adr/0008-world-structure-and-dungeon-progression.md` has no mention of
  water/holes/flying/leap-crossing/LoS-gated fog. This batch materially extends the world-structure
  model (obstacle kinds, a navigation-trait axis, LoS gating fog reveal) — exactly what ADR-0008 exists
  to record. An addendum or a new ADR (e.g. "0015 obstacle kinds and navigation traits") is owed.

## Prior v301 extras recs — confirmation

- **"Make local CI db/server bootstrap safe for isolated worktrees" — RESOLVED.** `make/db.mk:3` checks
  for an already-ready `arpg-postgres` before starting compose; `Makefile:11` pins/exports
  `COMPOSE_PROJECT_NAME ?= arpg-dev` (commit `b2374b04`). Both claimed commits present and correct.
- **"Split next touched validation domain out of `validate_shared.py`" — OPEN (regressed slightly).**
  Grew +21, no extraction.
- **"Resolve user-provided-unverified class GLB provenance" — OPEN (unchanged).**
  `assets/manifests/assets.v0.json:25` still carries `"license": "user-provided-unverified"` for class
  models; untouched by this batch. Still a pre-production blocker.
- **"Continue tuning-friendly rule-test audit" — CHANGED (incremental).** New scenarios follow the
  tuning-friendly style and the golden check stays semantic, but no systematic sweep of existing tests
  happened and the new golden floors are not cross-linked to generation config. Audit not closed.

## Top 5 shared/tooling/process refactors

1. **Update `docs/CODEMAP.md:15`** to list the four new backend files and scenarios 99–102. Highest
   leverage, lowest risk.
2. **Reconcile the budget contract** — bring CLAUDE.md rule #12 in line with `tools/bot/run.py:54` and
   `scenario-catalog.md:17` (10s default / 15s ceiling), or restore a true 10s cap. Two governance docs
   contradict each other on a CI-enforced number.
3. **Write an ADR addendum (amend ADR-0008 or add 0015)** covering the obstacle `kind` taxonomy, the
   grounded/flying navigation axis, leap `ignore_obstacle_kinds`, and LoS-gated fog reveal.
4. **Add a `cross_checks()` semantic guard** linking `dungeon_obstacles.json` golden floors to
   `dungeon_generation.v0.json` (`water.target_count`, `holes.target_count`, `solid_kind_weights`),
   mirroring the existing navigation cross-check.
5. **Extract the obstacle/dungeon golden-floor validation out of `tools/validate_shared.py`** into a
   focused, independently-importable `validate_dungeon_goldens.py` (typed context, not `globals()`) —
   finally acts on the standing split rec and reverses the +21 drift.

## Suggested scores (0–10)

- **Architecture-boundaries: 8** — server-owns-outcome boundary respected, new domains in focused files;
  ding for in-place v8 schema reuse where a v9 bump or documented justification was owed.
- **Architecture-cohesion: 9** — every new mechanic is a constrained schema knob with mutual-exclusion
  logic; new Go files domain-scoped.
- **Technical-correctness: 8** — `validate_shared` green at 1,435, tight auto-applied constraints; one
  missing semantic cross-check on the new golden floors.
- **Test-quality: 8** — scenario assertions prove each mechanic with no accidental tuning pins; docked
  for the 15s-vs-10s budget ambiguity and scenario 102's unguarded long path.
- **Maintainability: 7** — zero `helpers=globals()` and new domains isolated, but both Python monoliths
  drifted +21 and the standing split rec was not acted on.
- **Documentation: 6** — flawless per-slice SDD, dragged down by stale CODEMAP and ADR-0008 not
  reflecting the new world model.
- **Process-SDD: 9** — full per-slice ceremony, isolated-batch CI gate documented, prior db/worktree fix
  shipped; the un-updated CLAUDE.md budget rule and unactioned split keep it from 10.

*Evidence: direct reads of the cited schema/rule/golden/scenario files, `make validate-shared` (1,435
checks) green, `git show` for `b2374b04`/`20327d91`, `.maintainability/file-size-baseline.tsv`, and
`assets/manifests/assets.v0.json`.*
