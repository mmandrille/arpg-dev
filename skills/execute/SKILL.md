---
name: execute
description: >-
  Implement an approved plan from docs/plans, task by task, until make ci is
  green. Use when the user runs /execute, says "implement the plan", or points
  at docs/plans/vN_*.md and wants code changes.
disable-model-invocation: true
---

# /execute — Plan → Implementation

**Trigger:** `/execute {plan_file.md}` (e.g. `/execute docs/plans/v20_2026-06-06-play-session-loop.md`)

**Announce at start:** "Using the **execute** skill to review the plan and implement it."

## Hard rules

1. **Do not start coding until the plan review passes** — unresolved gaps block implementation.
2. **Follow the plan task order** unless a blocker forces a justified reorder (explain in chat).
3. **`make ci` must pass** before claiming the slice is done.
4. **Do not create git commits** unless the user explicitly asks.
5. **Do not skip bot scenarios** when the plan includes them.
6. **Update checkbox status** in the plan file as tasks complete (`- [ ]` → `- [x]`).

## Phase 0 — Baseline context (read first)

1. [`PROGRESS.md`](../../PROGRESS.md) — **Current status**, **Open gaps**, and **Agent checklist**
   only unless the task needs more.
2. [`CLAUDE.md`](../../CLAUDE.md)
3. The plan file the user provided.
4. The linked spec (`docs/specs/vN_spec-*.md`) if referenced in the plan.
5. Relevant ADRs cited by spec/plan.
6. [`docs/CODEMAP.md`](../../docs/CODEMAP.md) when the plan touches a known domain.

## Phase 1 — Plan review (gate)

Review the plan before touching code. Produce a short **Plan Review** in chat.

### Gap / error checklist

| Area | Check |
|------|-------|
| **Completeness** | File map covers all spec acceptance criteria. |
| **Ordering** | Shared → server → bot → client → docs; no client-before-server for authoritative behavior. |
| **Verify commands** | Each task has runnable commands; final verification includes `make ci`. |
| **Bot proof** | Gameplay/protocol tasks include scenario JSON + `make bot`. |
| **Golden/contracts** | Shared changes have `make validate-shared`; golden lists Go + GDScript tests. |
| **Branch** | Work only on the current checkout; never create or switch branches. |
| **Drift** | Plan still matches as-built code (prior slices may have landed since plan was written). |
| **Plugin checklist** | Client tasks include adopt/borrow/reject if UI/art is in scope. |
| **Maintenance ratchet** | Plan includes over-600-line file decisions and `make maintainability`. |

### Outcomes

- **Blockers found** → Numbered questions; **stop**. Do not implement.
- **Plan stale vs code** → Propose plan edits; ask user to confirm.
- **Plan OK** → Proceed to Phase 2.

## Phase 2 — Implement

### Default implementation order

```
shared/ (schemas, rules, golden)
  → server/ (game sim, realtime, store, replay)
  → tools/bot/ (scenarios, run.py, test_protocol.py)
  → client/ (GDScript, smoke, test_golden.gd)
  → PROGRESS.md
```

### Per-task workflow

For each plan task:

1. Mark task in progress (todo list or plan checkboxes).
2. Execute steps in order; run verify commands after each logical chunk.
3. If a verify command fails → diagnose root cause; fix; re-run. Do not claim done with failing tests.
4. Mark task complete in the plan file.

### Invariants (never break)

- Go `game/` determinism: seeded RNG only, no `time.Now()`, stable entity ordering.
- Shared rules are **data** — no executable logic in JSON rule files.
- Server owns outcomes; client is renderer + input.
- Animation is client-only; wire new reactions via `state_delta.events`.
- Golden changes → update Go tests **and** `client/tests/test_golden.gd`.
- File-size ratchet: new source/test/tool files stay <=600 lines; grandfathered files do not grow
  beyond the `.maintainability/file-size-baseline.tsv` allowance without a documented exception.

### Bot scenarios

When the plan includes bot work:

- Edit or create JSON under `tools/bot/scenarios/`.
- Run `make db-up && make bot` (or `make bot` if db is already up).
- Add/update `tools/bot/test_protocol.py` when new assertion helpers are needed.

### Client work

- Follow the plan's asset/plugin adopt / borrow / reject decision and reuse existing in-repo Godot scripts, scenes, demos, and asset manifests before introducing new dependencies.
- Run `make client-unit`; run `make client-smoke` when server integration is required.

## Phase 3 — Final verification

Run the plan's **Final verification** section. Minimum gate:

```bash
make maintainability
make ci
```

If `make ci` fails, fix and re-run until green. Use focused commands while iterating (`make validate-shared`, `make test-go`, `make bot`, `make client-unit`).

**Evidence before claims:** paste or summarize command output showing success before reporting completion.

## Phase 4 — Close out

When `make ci` is green:

1. Mark all plan checkboxes complete.
2. Update [`PROGRESS.md`](../../PROGRESS.md) (**Current status**, open gaps) and
   [`docs/progress/slice-lifecycle.md`](../../docs/progress/slice-lifecycle.md) (lifecycle row).
3. Add or update `docs/as-built/vN_<codename>.md` with what the slice proved.
4. Report to user:
   - What was implemented (concise).
   - Verification evidence (`make ci` green).
   - Suggested manual check from plan (e.g. `make play`) if any.
   - Reminder: commits only when they ask.

## When to stop and ask

Stop immediately and ask the user when:

- Plan step is ambiguous or contradicts the spec.
- Spec/plan assumption is wrong vs as-built code.
- A verify command fails after a reasonable fix attempt.
- Scope creep appears (work not in plan/spec).
- You need to change a shared protocol schema in a way the plan did not anticipate.

**Do not guess through blockers.**

## Examples

```
/execute docs/plans/v20_2026-06-06-play-session-loop.md
→ Review plan → implement tasks 1–5 → make ci green → update PROGRESS.md
```

```
/execute docs/plans/v10_2026-06-05-click-action-and-melee-range.md
→ Must not finish until 04_door_lab.json and make bot pass
```
