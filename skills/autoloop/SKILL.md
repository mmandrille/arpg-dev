---
name: autoloop
description: >-
  Coordinate an approved batch of game slices through separate sessions and
  detached worktrees, integrate each finished slice, then run combined CI,
  review, and refactor. Use for $autoloop, /autoloop, or a request to develop
  several slices in parallel. In a slice session, resume that slice's SDD steps.
disable-model-invocation: true
---

# $autoloop — Coordinated slice batch

Read [the batch workflow](references/batch-workflow.md) when coordinating a batch or working in one of its slice sessions. It is the shared contract for worktrees, handoffs, integration, CI, and cleanup.

## Main chat

1. Start with `PROGRESS.md` current status, open gaps, and checklist, then `CLAUDE.md`, `docs/CODEMAP.md`, the relevant ADRs, and Git state.
2. Use `/next` to offer a coherent batch of spec-ready slices. Show dependencies, likely overlap, acceptance checks, and execution order. If the user already proposed the slices, validate that set instead of replacing it. Wait for batch acceptance before creating sessions or worktrees.
3. After acceptance, reserve slice numbers and create one separate user-visible session and detached worktree per slice at the same recorded base. The batch approval authorizes these worktrees. Do not create branches. Give each session the brief and `/spec` → `/plan` → `/execute` assignment with focused verification and a complete handoff manifest.
4. Keep coordinating while sessions run: monitor them, answer shared questions, integrate ready slices to `main` as soon as dependencies allow, and run focused checks at overlap points. A dependent session can spec/plan while it waits for its implementation prerequisite.
5. After every accepted slice is integrated and accounted for, run the combined final gate, `/finish`, worktree cleanup, `/review`, and `/refactor` in the order defined by the batch workflow. A failed `make ci` is repaired and rerun; it is not reported as a completed gate.
6. Report slice outcomes, exact verification, commits, cleanup, and remaining limits. Do not claim completion while an accepted session is blocked or its files are missing from `main`.

## Slice session

When the session was created for one accepted `vN`, use the assigned brief and base commit. Run `/spec`, `/plan`, then `/execute` with their review gates. Work only in the assigned detached worktree. Run focused tests and visual/bot proof; write as-built evidence. Send the coordinator the complete handoff described in the batch workflow. Do not run `make ci`, `/finish`, commit, push, transfer to `main`, or delete the worktree.

## Standalone request

When the user explicitly asks for only one slice, or invokes `/spec`, `/plan`, `/execute`, or `/finish` outside a batch, use the standalone behavior of that skill. If `$autoloop` resumes an existing single slice, detect the latest approved checkpoint in the conversation: brief → spec, spec → plan, plan → execute, or verified implementation → finish. Continue the remaining gates for that slice. If ownership of dirty files or the active slice is ambiguous, ask before changing them. Do not create extra sessions merely because the skills support batch mode.

## Stops

Ask the smallest blocking question when product intent, a spec/plan gate, or safe integration cannot be resolved from the accepted batch. Continue independent sessions while one is blocked. Preserve uncommitted work and report exact failed checks; never bypass hooks, invent a branch, or discard a worktree to make the ledger look complete.
