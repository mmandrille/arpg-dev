---
name: review
description: >-
  Generate a repo-wide engineering review set under docs/reviews, following the
  existing overview/backend/client/shared-tooling-process pattern. Use when the
  user runs /review or $review, asks for the next periodic engineering review,
  or asks for a full-repo architecture/maintainability review document set.
---

# /review — Repo-wide Engineering Review

**Trigger:** `/review` or `$review`, optionally with a slice number or focus.

Examples:

- `/review` — review the current repo baseline and write the next review set.
- `$review v60` — write `docs/reviews/YYYYMMDD_v60-*.md`.
- `/review backend only` — write or update the relevant review file, then summarize what was skipped.

**Announce at start:** "Using the **review** skill to audit the repo and write the engineering review set."

## Hard rules

1. **Do not implement fixes.** This skill writes review documentation and, when appropriate, updates `PROGRESS.md` review metadata. Do not change production code.
2. **Ground every finding in current evidence.** Read code and docs directly; do not copy prior review claims unless re-verified.
3. **Record the baseline.** Every review file must state date, scope, branch/commit, and whether the worktree was clean.
4. **Use file:line citations for concrete issues.** Prefer `path:line` references for bugs, drift, and refactor targets.
5. **Separate facts from recommendations.** Issue sections describe observed risk; ranked recommendation sections say what to do next.
6. **If tests or checks are not run, say so in the review and final response.**

## Phase 0 — Baseline

Read first:

1. [`PROGRESS.md`](../../PROGRESS.md) — latest completed slice, review cadence, open gaps.
2. [`CLAUDE.md`](../../CLAUDE.md) — commands, architecture, invariants.
3. Latest files in [`docs/reviews/`](../../docs/reviews/) — template, prior findings, unresolved themes.
4. Relevant ADRs in [`docs/adr/`](../../docs/adr/) — especially ADR-0001 and any area-specific ADRs.

Determine:

- `review_slice`: user-supplied `vN`, otherwise latest completed slice from `PROGRESS.md`.
- `date_prefix`: `YYYYMMDD` from the local date.
- `baseline`: branch, short commit, latest commit title, and `git status --short`.
- Whether `PROGRESS.md` says **Next engineering review** is due. If not due but the user explicitly asked, proceed and note it as an ad hoc review.

File names:

```text
docs/reviews/{YYYYMMDD}_v{N}-overview.md
docs/reviews/{YYYYMMDD}_v{N}-backend.md
docs/reviews/{YYYYMMDD}_v{N}-client.md
docs/reviews/{YYYYMMDD}_v{N}-shared-tooling-and-process.md
```

If the target files already exist, update them in place rather than creating duplicates.

## Phase 1 — Evidence Gathering

Collect enough evidence to make the review reproducible:

- Git: branch, HEAD, status, recent commits.
- Repo shape: largest files, LOC by major area, test/scenario counts, schema/rule/golden counts.
- Backend: `server/internal/game`, `server/internal/realtime`, `server/internal/store`, server tests, determinism hazards.
- Client: `client/scripts`, `client/tests`, UI panel wiring, thin-client authority boundary, bot/replay scaffolding.
- Shared/tooling/process: `shared/protocol`, `shared/rules`, `shared/golden`, `tools/`, `Makefile`, `make/`, `scripts/`, `docs/specs`, `docs/plans`, ADRs, `PROGRESS.md`.
- Prior review follow-up: compare the latest review's top findings with current code and mark resolved, still open, or changed.

Useful checks, adjusted to the current environment:

```bash
git status --short
git rev-parse --abbrev-ref HEAD
git rev-parse --short HEAD
git log -1 --oneline
rg --files
make validate-shared
cd server && go test ./... && go vet ./...
.venv/bin/pytest tools
make client-unit
make client-smoke
```

Run `make ci` when the review is the official cadence gate or when the user asks for a CI-backed review. If a command is unavailable, too slow, or fails for reasons unrelated to the review, continue and record the exact limitation.

If multi-agent/subagent tools are available, use independent specialist audits to reduce blind spots:

- Backend specialist: Go server, determinism, persistence, replay, tests.
- Client specialist: Godot client, UI architecture, protocol application, smoke/unit coverage.
- Shared/tooling/process specialist: contracts, Python tools, build, SDD docs.

If subagents are unavailable, perform the same split sequentially.

## Phase 2 — Write Companion Reports

Write the three detailed reports first. Use concise, evidence-heavy prose.

### Backend report

Required shape:

```markdown
# arpg-dev — Backend (Go server) review at slice **v{N}**

**Date:** YYYY-MM-DD
**Scope:** ...
**Baseline:** ...
**Stats:** ...
**Overview:** [`YYYYMMDD_vN-overview.md`](YYYYMMDD_vN-overview.md)

---

## Summary

## 1. Architecture
## 2. Technical
## 3. Maintainability
## 4. Documentation

## Top 5 backend refactors

*Evidence: ...*
```

Inspect at least: authoritative boundary, deterministic sim rules, input dispatch, persistence transactions, realtime locking/fanout, replay path, major tests.

### Client report

Required shape mirrors backend with:

```markdown
# arpg-dev — Client (Godot 4 / GDScript) review at slice **v{N}**
```

Inspect at least: thin-client authority, protocol snapshot/delta application, input/prediction/reconciliation, UI panel ownership, bot/replay scaffolding, client tests, scene/script organization.

### Shared/tooling/process report

Required shape mirrors backend with:

```markdown
# arpg-dev — Shared contracts, Python tooling & SDD process review at slice **v{N}**
```

Inspect at least: protocol versioning, schemas/examples, rules-as-data, golden coverage, Python bot/validator shape, asset tooling, Makefile/CI orchestration, spec/plan/as-built cadence, ADR drift.

## Phase 3 — Write Overview

Create the overview last so it accurately consolidates the subreports.

Required shape:

```markdown
# arpg-dev — Engineering review at slice **v{N}** (overview)

**Date:** YYYY-MM-DD
**Reviewer:** ...
**Scope:** Whole repo — ...
**Git baseline:** ...
**Companion reports:**
- [`YYYYMMDD_vN-backend.md`](YYYYMMDD_vN-backend.md) — Go authoritative server
- [`YYYYMMDD_vN-client.md`](YYYYMMDD_vN-client.md) — Godot thin client
- [`YYYYMMDD_vN-shared-tooling-and-process.md`](YYYYMMDD_vN-shared-tooling-and-process.md) — contracts, Python tooling, docs/SDD

---

## 1. Executive summary
## 2. Scorecard
## 3. Cross-cutting themes
## 4. Top 10 recommendations
## 5. What "done well" looks like here

*Evidence: ...*
```

Scorecard dimensions should be stable across reviews:

- Architecture (boundaries)
- Architecture (internal cohesion)
- Technical correctness
- Test quality
- Maintainability
- Documentation
- Process / SDD
- Overall

Scores are useful only when justified by concrete evidence. Do not inflate scores to be polite; do not penalize twice for the same root cause.

## Phase 4 — PROGRESS.md Updates

If this is an official periodic review or the user asked to land it as the current review, update [`PROGRESS.md`](../../PROGRESS.md):

- `Last engineering review` → the new overview file and date.
- `Next engineering review` → next intended milestone, usually `v{N+10}`.
- Add or refresh a short open-gap entry only for findings that should steer `/next`. Do not paste the whole review into `PROGRESS.md`.
- Update `Last updated`.

If this is an ad hoc or partial review, do not change the cadence fields unless the user explicitly wants it to supersede the prior review.

## Quality Bar

- Each report should include both strengths and risks.
- Findings tagged `[High]`, `[Med]`, `[Low]`, or `[Strength]` should be internally consistent.
- Top recommendations must be actionable as future slices or small maintenance tasks.
- Avoid vague advice like "clean up code"; name the seam, file, or check to add.
- Keep the overview readable as an executive artifact; put detailed evidence in companion files.
- The final chat response should list the written files, baseline, checks run, and any checks skipped or failed.

## Handoff

End with:

1. Files written or updated.
2. Baseline reviewed.
3. Verification commands run and outcomes.
4. Whether `PROGRESS.md` cadence fields were updated.
5. Suggested next action, usually `$next` using the review findings as input.
