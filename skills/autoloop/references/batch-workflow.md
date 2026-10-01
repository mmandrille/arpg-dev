# Coordinated slice batch

Use this workflow when the user approves more than one slice, or asks `$autoloop` to run a batch. The main chat is the coordinator; each accepted slice gets one user-visible session and one detached worktree. The coordinator owns `main`, integration, final CI, review, and refactor. A direct `/spec`, `/plan`, `/execute`, or `/finish` outside a batch keeps its standalone behavior.

## Before dispatch

1. Read `PROGRESS.md` current status, open gaps, and checklist; then `CLAUDE.md`, `docs/CODEMAP.md`, relevant ADRs, and the current Git state. Do not infer the baseline from old slice numbers. This workflow integrates into `main`, so confirm the coordinator is already on `main`; do not switch branches on its initiative.
2. `/next` offers a coherent set of numbered, spec-ready slices with acceptance checks, dependencies, likely file overlap, and a safe execution order. If the user supplied the set, validate and present that set. Reserve `vN` numbers only after the user accepts the batch. Approval of the batch authorizes creating its sessions and temporary detached worktrees; do not ask again for each slice.
3. Record the exact `main` base commit and existing dirty files. Preserve unrelated user changes. If an accepted draft spec exists only as an uncommitted file, transfer it to its assigned worktree as part of dispatch. Resolve any ambiguity that would make integration unsafe before dispatch.
4. Create one separate user-visible session per accepted slice and one detached worktree at the recorded base (`git worktree add --detach <slice-path> <base-commit>`). Do not create a branch or check out `main` in a second worktree. In Codex, create a local project session and give it the absolute slice path; direct every file and command operation there because the session initially opens at the project checkout. Do not use a worktree-session creation option that creates a branch. Use another host's equivalent only if it also preserves detached HEAD. If separate sessions or detached worktrees are unavailable, explain the blocker rather than silently substituting in-chat subagents or creating branches.
5. Give each session its `vN`, brief, base commit, absolute worktree path, dependencies, owned surfaces, expected handoff, and this workflow. Sessions may spec and plan concurrently. A slice may implement only when its prerequisites are available in that worktree; keep dependent sessions working on independent spec/plan tasks until then. When a prerequisite reaches `main`, transfer that exact integrated prerequisite into the dependent worktree without overwriting its spec/plan or other edits, then run its focused dependency checks. Do not claim all implementation can run simultaneously when dependencies conflict.

## Slice-session contract

Each slice session runs `/spec` → `/plan` → `/execute` in its own worktree. Spec and plan are separate review gates; execution needs an approved plan. The session runs the smallest meaningful checks for its changes, including named bot and real-camera visual scenarios when applicable. It writes its spec, plan, and as-built evidence and updates plan checkboxes. It does **not** run `make ci`, `make ci-full`, `/finish`, commit, push, modify the coordinator's `main`, or mark the batch complete. A blocked session reports the exact question or failed check while other sessions continue.

Before handoff, the session reports:

- base commit, slice number/codename, dependencies used, and whether its worktree is dirty;
- complete changed, deleted, and untracked file list, plus any ignored raw evidence that must be preserved;
- exact focused commands and outcomes, bot scenario names, visual captures, measured limits, and acceptance criteria not met;
- shared files likely to conflict, and any integration steps needed.

The session stops after the handoff and stays available for focused fixes. Do not clean its worktree until the coordinator has verified complete transfer.

## Coordinator loop while sessions run

Watch all sessions and integrate each ready slice into the coordinator's `main` as soon as its dependencies are satisfied. Keep a small ledger of `proposed → dispatched → ready → integrated → verified → closed` for every slice; report progress without treating a quiet session as complete.

For each ready handoff:

1. Inspect the session's actual Git diff, untracked files, base commit, and evidence. Compare with the handoff manifest. Do not assume a copied file list is complete.
2. Apply the complete change set to `main` with a three-way merge or equivalent careful transfer. Resolve overlapping files against the current integrated state. Never overwrite a newer integrated edit blindly or cherry-pick a partial file set. Preserve required ignored evidence locally and keep secrets/local-only logs out of commits.
3. Reconcile shared registries and fixtures such as `docs/CODEMAP.md`, scenario movement audit, asset manifest, CI pack, schemas, maintainability baseline, and `PROGRESS.md` only where touched. For every `shared/**/*.v0.json` path touched by more than one slice, list the contributors and diff the final merged file against each handoff; verify that every intended change and its schema validation survived. Run focused checks for the integration seam and regressions introduced by overlap. Do not run full `make ci` for every slice.
4. Before releasing a child worktree, compare **every** child changed/deleted/untracked path with the integrated result. For divergent shared files, verify that each intended behavior and test is represented, not merely that the path exists. Keep the session open for corrections if anything is missing.
5. Mark the slice integrated in the ledger. Continue monitoring the remaining sessions; do not wait idly for the whole batch before integrating ready work.

## Final gate and closeout

Once **every** accepted slice is integrated and its focused checks are green:

1. Audit all handoff manifests against `main`; verify specs, plans, as-built notes, screenshots/assets, licenses, scenario registrations, and lifecycle rows. Run `git diff --check` and relevant cheap validators before the expensive gate.
2. Run one combined `make ci` on the integrated `main` state. If it fails, diagnose with focused checks, fix the integration or slice defect, and rerun `make ci` until the final state is green. A failed attempt is not a final gate. Do not run CI per slice to satisfy this step.
3. Use `/finish` in coordinator mode to confirm the green gate on that exact integrated code state, mark accepted slices complete, update `PROGRESS.md` and the lifecycle index, and commit the verified work in logical slice or batch commits. Do not push unless asked. Documentation-only closeout after CI does not require a second full run; a code or contract change does.
4. Preserve any ignored evidence referenced by as-built notes. Save a recoverable snapshot of any remaining dirty child state, then remove or archive **only** the batch's temporary worktrees after verifying all their changes are integrated. Leave the user's existing worktrees and sidebar choices intact. Confirm `git worktree list` contains no batch leftovers and `main` is clean.
5. Run `/review` on the new `main` baseline even if the usual cadence was not due. It records the combined `make ci` result and whether `make ci-full` actually ran. Validate and commit the review reports and cadence update as one documentation commit so `main` is clean. Then run `/refactor` against that exact review. Refactor uses focused verification per minor change; if it changes code or contracts beyond those checks, run one additional final `make ci` for the post-refactor state before reporting completion. Confirm the final `main` state is clean.

If an accepted slice remains blocked, keep the integrated state and evidence intact, report the blocker, and do not claim the batch's final CI/review/refactor is complete. If CI or review finds a missing child change after cleanup, restore it from the saved snapshot or session before declaring success.
