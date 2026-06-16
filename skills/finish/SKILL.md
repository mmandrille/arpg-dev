---
name: finish
description: >-
  Close out a slice: update PROGRESS.md, verify, and commit with feat: vN:
  title format. Use when the user runs /finish or says the slice is done and
  wants a consolidated commit.
disable-model-invocation: true
---

# /finish — Slice Close-out & Commit

**Trigger:** `/finish`

**Announce at start:** "Using the **finish** skill to consolidate the slice, verify CI, and commit."

## Hard rules

1. **`make ci` must pass** before committing a standalone slice. No commit on red CI.
   In `$autoloop` batch mode, commit each slice after adequate focused verification, and let
   `$autoloop` run one final `make ci` after all requested slices are committed.
2. **Update `PROGRESS.md`** if not already current — **Current status**, open gaps, review cadence.
   Update [`docs/progress/slice-lifecycle.md`](../../docs/progress/slice-lifecycle.md) for the new row.
3. **Commit message format is fixed:**

   ```
   feat: v{slice_number}: {title of this slice}
   ```

   Examples:
   - `feat: v20: play session loop`
   - `feat: v19: teleporters and waypoint UI`

   Use the slice **execution number** (`vN` from spec/plan filename), lowercase title after the colon, no extra prefix types (`chore:`, `fix:`) for the primary slice commit unless the user explicitly overrides.

4. **Never commit** `.env`, credentials, or secrets. Warn if present in the diff.
5. **Never** `git push` unless the user explicitly asks.
6. **Never** `--no-verify`, `--amend` (unless user asks and amend rules apply), or destructive git commands.

## Phase 0 — Identify the slice

Determine `vN`, codename, and human title from (in order):

1. Uncommitted / staged changes — look for `docs/specs/vN_spec-*.md`, `docs/plans/vN_*.md`.
2. Current git branch (identify codename from branch name, spec, or plan — do not create branches).
3. [`PROGRESS.md`](../../PROGRESS.md) — active branch, in-progress rows.
4. Plan checkboxes — all or most marked `[x]`.

If ambiguous (multiple slices in diff, or number unclear), **ask the user** before proceeding.

Record:

- `slice_number` = N (integer)
- `codename` = kebab-case
- `title` = short human-readable phrase for commit message (from spec heading or plan goal)

## Phase 1 — Consolidate PROGRESS.md

Read [`PROGRESS.md`](../../PROGRESS.md) and verify it reflects the shipped slice.

### Required updates when missing or stale

| Section | Action |
|---------|--------|
| **Current status** | `Latest completed slice` → vN; `Active branch` → feature branch or TBD; `CI gate` → date + green for standalone slices, or focused verification with final batch CI pending for `$autoloop`; `Next slice` → TBD |
| **Slice lifecycle table** | Add/update row in [`docs/progress/slice-lifecycle.md`](../../docs/progress/slice-lifecycle.md): Status = Complete after required verification, Spec + Plan links |
| **As-built summary** | Add or update `docs/as-built/vN_<codename>.md`: what it proves, key decisions, scope limits |
| **Open gaps & deferred work** | Update deferred/autoloop tables when spec non-goals change — **do not** add inline "Recently closed" prose to `PROGRESS.md` |
| **Last updated** | Today's date |
| **Engineering review** | If the shipped slice hits the next ~10-slice milestone in `PROGRESS.md` → **Next engineering review**, stop after the slice commit and hand off to `$refactor` then `$review`; `$review` updates **Last/Next engineering review** and adds the review set |
| **Maintenance ratchet** | Confirm any touched over-600-line files were split, stayed within allowance, or have a documented exception |

Cross-check against:

- The slice spec acceptance criteria (all met or explicitly deferred in PROGRESS).
- The plan's final verification section.
- Actual code changes in the diff.

If PROGRESS is incomplete, **update it** before CI/commit.

Also verify:

- Plan file checkboxes marked `[x]` for completed tasks.
- Spec status line updated to complete/approved if the spec has a Status field.

## Phase 2 — Pre-commit verification

### Git state

Run in parallel:

```bash
git status
git diff
git diff --staged
git log -5 --oneline
```

Review:

- All slice-related files are accounted for (no stray debug files).
- No secrets or local-only artifacts staged.
- Changes match the identified slice scope.

### CI gate

For standalone `/finish`:

```bash
make maintainability
```

```bash
make ci
```

If CI fails:

1. Diagnose and fix.
2. Re-run `make ci` until green.
3. Include fixes in the same commit if they belong to this slice.

Use focused commands while iterating if helpful: `make validate-shared`, `make test-go`, `make bot`, `make client-unit`.

**Evidence before commit:** confirm `make ci` succeeded in chat (brief summary or exit code 0).

For `$autoloop` batch mode:

1. Confirm the slice's focused verification evidence from `$execute`.
2. Run additional focused commands only if close-out changes create new risk.
3. Do **not** run `make ci` for every slice when focused tests cover the changes. `$autoloop`
   must run one final `make ci` after all requested slices are committed and fix minor batch
   regressions there.
4. Before committing, state that final CI is deferred to the autoloop batch gate.

## Phase 3 — Commit

### Staging

Stage all files that belong to the slice:

```bash
git add <paths>
```

Include: code, shared contracts, golden fixtures, bot scenarios, specs, plans, PROGRESS.md,
`docs/as-built/vN_<codename>.md`, ADR updates if part of slice.

Do **not** stage unrelated local edits — ask user if mixed changes exist.

### Commit message

```bash
git commit -m "$(cat <<'EOF'
feat: v{N}: {title of this slice}

EOF
)"
```

Replace `{N}` and `{title}` with values from Phase 0. Keep title concise (3–8 words).

Optional: add one short body sentence after the title line if it clarifies **why** the slice matters — only when it adds value.

### Post-commit

```bash
git status
```

Confirm clean working tree (or only intentional unstaged files). Report commit hash and message to the user.

## Phase 4 — Handoff

Tell the user:

1. Slice vN committed with message `feat: vN: …`.
2. Verification evidence (`make ci` green for standalone finish, or focused commands with final
   batch CI pending for `$autoloop`).
3. PROGRESS.md updated (list what changed).
4. Suggested follow-ups:
   - `/next` to pick the following slice
   - merge PR / push (only if they ask)
   - manual `make play` check if plan listed one and it wasn't run

## When to stop and ask

- Cannot determine slice number or title.
- Uncommitted changes span multiple slices.
- `make ci` fails after reasonable fix attempts.
- PROGRESS.md conflicts with spec (acceptance criteria not actually met).
- User has unrelated dirty files mixed into the slice — ask what to include.

**Do not commit partial or unverified slice work.**

## Relationship to `/execute`

`/execute` implements but does not commit by default. `/finish` is the **explicit close-out**: docs consolidation + CI proof + single `feat: vN:` commit.

Typical flow:

```
/next → spec → /plan → /execute → /finish
```
