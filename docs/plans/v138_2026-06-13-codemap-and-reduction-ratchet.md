# v138 Plan — CODEMAP & Reduction Ratchet

Status: Complete
Goal: Cut per-feature context for AI agents (a committed domain→files index) and convert the
maintainability ratchet from containment to reduction (lower-bound ratchet + CI enforcement + trend
metric + policy), with zero game-behavior change.
Architecture: Two cohesive docs/tooling deliverables. (1) `docs/CODEMAP.md` maps each gameplay
domain to the exact files it spans, validated by `tools/validate_codemap.py` so it cannot rot.
(2) `scripts/check-file-size-ratchet.sh` gains a lower-bound rule (a baseline may not sit more than
the growth allowance above the file's actual count → shrinks lock in) and a trend summary, and is
wired into `make ci` (today it is not). Touch-to-shrink and new-domain rules ship as documented
policy + an extended plan checklist; the lower-bound ratchet is the mechanical guarantee. No git-diff
enforcement (per decision: brittle on the work-on-`main` flow).
Tech stack: Bash ratchet script, Make fragments, Python validator + pytest, Markdown docs.

## Baseline and shortcut decision

- Builds on v137 (`bot-assertion-domain-split`); v138 is the next free slice number (verified: no
  existing `docs/specs/v138*` or `docs/plans/v138*`).
- Reuses existing patterns: the `tools/validate_*.py` + `tools/test_validate_*.py` convention
  (mirrors `validate_unique_items.py` / `validate_i18n.py`); the `scripts/check-file-size-ratchet.sh`
  + `.maintainability/file-size-baseline.tsv` ratchet; the `make/ci.mk` `ci:`/`maintainability:`
  targets.
- No Godot plugin adoption note required: no client UI/camera/art is in scope (that is downstream
  proposal #4).
- **Decisions locked with user:** (1) touch-to-shrink = policy + lower-bound ratchet, no git-diff;
  (2) downstream #2/#4/#5 stay roadmap-only in the v138 spec — no pre-created stubs, no folding #5 in.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `scripts/check-file-size-ratchet.sh` | Add lower-bound ratchet, trend summary, `ROOT`/`BASELINE` overrides for testability |
| Modify | `.maintainability/file-size-baseline.tsv` | Refresh to exact current line counts so both bounds pass on landing |
| Create | `tools/test_file_size_ratchet.py` | Prove upper-bound fail, lower-bound fail, and within-allowance pass |
| Modify | `make/ci.mk` | `ci:` depends on `maintainability` so the ratchet runs in `make ci` |
| Create | `docs/CODEMAP.md` | Domain → files index with a "how to use" header |
| Create | `tools/validate_codemap.py` | Assert every path referenced in CODEMAP exists on disk; assert no domain row is empty |
| Create | `tools/test_validate_codemap.py` | Validator passes on real CODEMAP; fails on a fabricated missing path |
| Modify | `make/shared.mk` | Run `validate_codemap.py` as part of `validate-shared` |
| Modify | `CLAUDE.md` | Rewrite "Maintainability Ratchet" for reduction; add CODEMAP pointer |
| Modify | `AGENTS.md` | Add CODEMAP as the canonical "what files do I load for domain X" index |
| Modify | `skills/plan/SKILL.md` | Extend the "Maintenance ratchet" checklist with a touch-to-shrink question |
| Modify | `PROGRESS.md` | Add v138 to the slice-numbering list + lifecycle table |
| Create | `docs/as-built/v138_codemap-and-reduction-ratchet.md` | Closeout note (written at `/finish`) |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [ ] `client/scripts/main.gd` — not touched.
- [ ] `server/internal/game/game_test.go` — not touched.
- [ ] `tools/bot/run.py` — not touched.
- [ ] `tools/validate_shared.py` — **not touched** (CODEMAP validation is wired via `make/shared.mk`,
  not added inside the 3324-line validator).
- [ ] Other over-limit file from `.maintainability/file-size-baseline.tsv`: **none**. New files
  (`validate_codemap.py`, `test_validate_codemap.py`, `test_file_size_ratchet.py`) are small;
  `check-file-size-ratchet.sh` grows from 68 to ~120 lines (well under 600).

Decision:
- [x] No extraction required — v138 adds only small files and edits docs (exempt) + the ratchet
  script itself. The baseline TSV is refreshed (not grown).

Verification:
```bash
make maintainability
```

## Task 1 — Reduction ratchet engine

Files:
- Modify: `scripts/check-file-size-ratchet.sh`
- Modify: `.maintainability/file-size-baseline.tsv`

- [x] Step 1.1: Make the script testable — allow `ROOT` and `BASELINE` to be overridden by env
  (`ROOT="${ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"`,
  `BASELINE="${BASELINE:-${ROOT}/.maintainability/file-size-baseline.tsv}"`). `MAX_LINES` and
  `GROWTH_ALLOWANCE` are already env-overridable.
- [x] Step 1.2: Add the **lower-bound ratchet**. For each path in the baseline that still exists,
  if `baseline_count - line_count > GROWTH_ALLOWANCE`, append a failure:
  `"<path>: <line_count> lines is far below grandfathered baseline <baseline_count> — lower the
  baseline to <line_count> (or drop the entry if <= ${MAX_LINES})."` This locks in shrinks: once a
  file is cut, its baseline must follow down.
- [x] Step 1.3: Add the **trend summary**. After the pass/fail decision, always print one line:
  `"grandfathered: <N> files, <total> lines (target: down)"` where `N` and `<total>` are computed
  from the baseline entries' current actual `wc -l`. Stdout only — no committed trend file (avoids
  per-slice churn; milestone reviews record the number).
- [x] Step 1.4: Keep the existing upper-bound (`> baseline + allowance`) and new-file (`> MAX_LINES`)
  rules unchanged.
- [x] Step 1.5: Refresh `.maintainability/file-size-baseline.tsv` to exact current counts so both
  bounds pass on landing (regenerate the second column from `wc -l` for each tracked baseline path;
  preserve the header comment and path order).
```bash
make maintainability   # must print "file-size ratchet passed" + the trend line, exit 0
```

## Task 2 — Ratchet behavior test

Files:
- Create: `tools/test_file_size_ratchet.py`

- [x] Step 2.1: In a `tmp_path`, `git init` + write a small fixture file and a fabricated baseline,
  then invoke the script via `subprocess` with `ROOT`/`BASELINE` overrides.
- [x] Step 2.2: Assert **upper-bound fail** — file at 700 lines, baseline says 600 → exit != 0,
  stderr mentions "exceeds".
- [x] Step 2.3: Assert **lower-bound fail** — file at 100 lines, baseline says 700 → exit != 0,
  stderr mentions "lower the baseline".
- [x] Step 2.4: Assert **within-allowance pass** — file at 610 lines, baseline 600 → exit 0.
```bash
.venv/bin/python -m pytest tools/test_file_size_ratchet.py -q
```

## Task 3 — Wire the ratchet into CI

Files:
- Modify: `make/ci.mk`

- [x] Step 3.1: Make `ci:` depend on `maintainability` (`ci: maintainability`) so the cheap ratchet
  runs fail-fast before `scripts/ci.sh`. (Do not edit `scripts/ci.sh` step numbering — the make
  dependency is the lower-risk wiring.)
- [x] Step 3.2: Confirm `make ci` invokes the ratchet (grep the run output for the trend line / "file-size ratchet passed").
```bash
make maintainability && echo "ratchet standalone OK"
# full proof deferred to Final verification (make ci)
```

## Task 4 — CODEMAP document

Files:
- Create: `docs/CODEMAP.md`

- [x] Step 4.1: Write a short "how to use" header: "To work on domain X, load these files first.
  For status read `PROGRESS.md`; for history read `docs/specs|plans|as-built/` and `docs/adr/`.
  Every path here is checked to exist by `tools/validate_codemap.py`."
- [x] Step 4.2: Write a domain → files table with columns: **Domain | Server | Client | Shared rules
  | Bot | Tests / Migrations**. Populate each cell with real, existing backtick-quoted paths
  (gather via `grep`/`ls`; do not invent). Minimum domains:
  market, unique-items/effects, classes (paladin/rogue/ranger), skills, stash, shop/vendor,
  dungeon-gen, combat & damage-types, elite-aura, corpse/permadeath, town-services,
  session & realtime, replay, persistence/store, protocol, bot/scenarios, assets, i18n/text.
- [x] Step 4.3: Keep it an index, not a tutorial — no per-feature how-it-works prose.
```bash
ls docs/CODEMAP.md
```

## Task 5 — CODEMAP validator + wiring

Files:
- Create: `tools/validate_codemap.py`
- Create: `tools/test_validate_codemap.py`
- Modify: `make/shared.mk`

- [x] Step 5.1: `validate_codemap.py` — read `docs/CODEMAP.md`, extract every backtick-quoted token
  that looks like a repo path (contains `/` and a file extension, or matches a tracked path), and
  assert each exists relative to repo root. Exit non-zero with a clear per-path message on any
  missing path. Also fail if any domain table row has an empty set of paths.
- [x] Step 5.2: `test_validate_codemap.py` — (a) run the validator against the real
  `docs/CODEMAP.md` and assert it passes; (b) run it against a fabricated CODEMAP string containing
  a bogus path and assert it reports failure.
- [x] Step 5.3: Wire into `validate-shared` in `make/shared.mk` (append a
  `$(PYTHON) tools/validate_codemap.py` invocation to the recipe) so a stale map fails CI step 1/9.
```bash
.venv/bin/python -m pytest tools/test_validate_codemap.py -q
make validate-shared
```

## Task 6 — Policy docs

Files:
- Modify: `CLAUDE.md`
- Modify: `AGENTS.md`
- Modify: `skills/plan/SKILL.md`

- [x] Step 6.1: Rewrite `CLAUDE.md` → "Maintainability Ratchet" for **reduction**: state the
  lower-bound ratchet (shrinks lock in), the **touch-to-shrink** rule (a slice editing a
  grandfathered file leaves it ≤ its baseline; the +25 allowance is for untouched incidental drift
  only), the **new-domain** rule (a new gameplay domain starts in its own file, never inside a
  coordinator), the trend expectation, and that the ratchet is now a `make ci` gate (not advisory).
- [x] Step 6.2: Add a CODEMAP pointer in `CLAUDE.md` (Architecture section) and `AGENTS.md`
  (entrypoint list): "To find which files a domain spans, read `docs/CODEMAP.md`."
- [x] Step 6.3: Extend the `skills/plan/SKILL.md` "Maintenance ratchet" plan-template checklist with:
  `- [ ] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?`
```bash
grep -n "touch-to-shrink" CLAUDE.md skills/plan/SKILL.md
grep -n "CODEMAP" AGENTS.md CLAUDE.md
```

## Task 7 — Lifecycle docs and CI

Files:
- Modify: `PROGRESS.md`
- Create: `docs/as-built/v138_codemap-and-reduction-ratchet.md` (at `/finish`)

- [x] Step 7.1: Add `v138_* = codemap-and-reduction-ratchet` to the slice-numbering list and a v138
  lifecycle-table row (spec + plan links; as-built at finish).
- [x] Step 7.2: Update **Latest completed slice** / **Next slice** at `/finish` time.
- [x] Step 7.3: Write the as-built note (what changed, proof commands) at `/finish`.
```bash
make ci
```

## Final verification

- [x] `make maintainability` (lower-bound + trend; green on refreshed baseline)
- [x] `.venv/bin/python -m pytest tools -q` (includes the two new test files)
- [x] `make validate-shared` (CODEMAP path validation runs)
- [x] `make ci` (now exercises the ratchet end-to-end via the `ci: maintainability` dependency)

No `make test-go` / `make client-unit` / `make bot` proof is required: v138 changes no Go source,
GDScript, protocol, or scenario. Existing CI bot/replay coverage remains the regression proof that
nothing moved.

## Deferred scope (downstream roadmap — separate specs v139+, scheduled via /next)

- **#2** Server coordinator splits: `repos.go` → `market_repo.go` (`repos.go:1709–2590`);
  `sim.go` → `sim_load.go` (`772–1078`) + `sim_players.go` (`1086–1364`); mirror test splits off
  `game_test.go`/`store_test.go`.
- **#4** Client `main.gd` decomposition into focused presenters/bridges (v127 `TownServiceBridge`
  pattern + headless tests).
- **#5** Market expiration freshness contract: funnel all market read paths through one expiry sweep
  and/or a deterministic periodic tick so refund/restore/audit side-effects can't be stranded behind
  `ListActiveMarketListings`; lands inside the `market_repo.go` boundary from #2.
```
