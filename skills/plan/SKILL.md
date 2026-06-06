---
name: plan
description: >-
  Turn an approved spec into an implementation plan under docs/plans. Use when
  the user runs /plan, says "create a plan from spec", or points at
  docs/specs/vN_spec-*.md and wants a plan before coding.
disable-model-invocation: true
---

# /plan — Spec → Implementation Plan

**Trigger:** `/plan {spec_file.md}` (e.g. `/plan docs/specs/v20_spec-play-session-loop.md`)

**Announce at start:** "Using the **plan** skill to review the spec and draft the implementation plan."

## Hard rules

1. **Do not write code.** This skill only reviews the spec and produces a plan file.
2. **Do not write the plan until the spec review passes** — unresolved gaps or open questions block plan creation.
3. **Ask the user** when anything is ambiguous, contradictory, or missing. Use `AskQuestion` when available.
4. **Bot scenarios are mandatory** when the slice touches gameplay, protocol, world presets, combat, inventory, movement, transitions, or replay — unless the spec explicitly defers bot proof with a written reason.

## Phase 0 — Baseline context (read first)

1. [`docs/PROGRESS.md`](../../docs/PROGRESS.md) — current slice, open gaps, invariants.
2. [`CLAUDE.md`](../../CLAUDE.md) — commands, architecture, SDD process.
3. [`AGENTS.md`](../../AGENTS.md) — Godot plugin adoption requirement for client work.
4. The spec file the user provided.
5. Related specs, ADRs, and as-built code cited by the spec — spot-check that claims match reality.

## Phase 1 — Spec review (gate)

Review the spec systematically. Produce a short **Spec Review** section in chat before writing the plan.

### Gap / error checklist

| Area | Check |
|------|-------|
| **Baseline** | Slice number, codename, branch, and "builds on vN" match `PROGRESS.md`. |
| **Scope** | Non-goals are explicit; no hidden work smuggled in acceptance criteria. |
| **Contracts** | Protocol/schema/golden changes are listed; version bumps noted if needed. |
| **Determinism** | Go `game/` changes avoid wall-clock, unseeded randomness, map iteration. |
| **Shared rules** | Formulas/items/combat stay data-driven; golden updates called out. |
| **Server authority** | Client is presentation only; outcomes owned by sim. |
| **Animation** | New reactions have a server event before client mapping (ADR-0007). |
| **World presets** | `worlds.v0.json` changes are concrete (entity ids, positions, `world_id`). |
| **Bot proof** | Acceptance criteria map to bot scenario steps or new scenario JSON. |
| **Replay** | Input shape changes mention replay test updates. |
| **Client** | Godot plugin adoption checklist required if UI/camera/art is in scope. |
| **Docs** | `PROGRESS.md` update is in scope when slice ships. |
| **As-built drift** | Spec assumptions match current code (grep/read cited files). |

### Outcomes

- **Blockers found** → List numbered questions; **stop**. Do not create the plan.
- **Minor gaps you can fix** → Propose spec edits; ask user to confirm before planning.
- **Spec OK** → Proceed to Phase 2.

## Phase 2 — Write the plan

### Output path and naming

```
docs/plans/vN_YYYY-MM-DD-<codename>.md
```

- `N` = slice execution order from the spec filename (e.g. `v20_spec-…` → `v20_…`).
- `<codename>` = kebab-case from spec codename.
- Date = today (user's timezone from context, else UTC date).

If a plan for the same slice already exists, **update it** rather than creating a duplicate — note what changed in the plan header.

### Plan structure (required sections)

```markdown
# vN Plan — <Human title>

Status: Ready for implementation
Goal: <one sentence>
Architecture: <2–4 sentences — key design decisions>
Tech stack: <Go sim, shared JSON, Godot client, Python bot, etc.>

## Baseline and shortcut decision
<What prior slices this reuses; Godot plugin adopt/borrow/reject if client work>

## File map
| Action | Path | Responsibility |
|--------|------|----------------|

## Task 1 — <name>
Files:
- Create/Modify: `path`

- [ ] Step 1.1: <action>
```bash
<verify command>
```

(repeat tasks in dependency order: shared → server → bot → client → docs)

## Task N — Bot scenarios   ← REQUIRED when gameplay/protocol in scope
- [ ] New or updated `tools/bot/scenarios/NN_<lab>.json`
- [ ] Migrate affected existing scenarios (`01`–`13`)
- [ ] `tools/bot/test_protocol.py` if new assertions/helpers needed
- [ ] `make bot` verification

## Task N — Lifecycle docs and CI
- [ ] Update `docs/PROGRESS.md` when slice completes
```bash
make ci
```

## Final verification
- [ ] `make validate-shared` (if shared/ touched)
- [ ] `make test-go` (if server/ touched)
- [ ] `make client-unit` (if client/ touched)
- [ ] `make bot` (if bot scenarios touched)
- [ ] `make ci`
```

### Task ordering (default)

1. Shared contracts + golden fixtures → `make validate-shared`
2. Go sim + unit tests → `cd server && go test ./internal/game/...`
3. Bot scenarios + `make bot`
4. Godot client + `make client-unit` / `make client-smoke`
5. `docs/PROGRESS.md` + `make ci`

### Bot scenario guidance

- Scenarios live in `tools/bot/scenarios/`. Number new labs sequentially after `13`.
- Each scenario needs a matching `world_id` preset in `shared/rules/worlds.v0.json` when the lab is world-specific.
- Prefer **new lab scenario** over overloading `01_vertical_slice` unless the change is a migration of existing steps.
- Map spec acceptance criteria → concrete bot steps (`move`, `action_entity`, `use_stair`, `teleport_intent`, assertions).
- Reference existing patterns: `04_door_lab.json`, `12_dungeon_levels.json`, `13_teleporter_lab.json`.

### Plan quality bar

- Every task has **files**, **checkbox steps**, and **verify commands**.
- No vague steps ("implement feature", "fix bugs").
- Cross-language golden changes list both Go tests and `client/tests/test_golden.gd`.
- Call out deferred scope explicitly in the plan footer.

## Phase 3 — Handoff

After writing the plan, tell the user:

1. Plan path created/updated.
2. Any spec edits you recommend (if not already applied).
3. Suggested branch: `feature/<codename>` per spec.
4. Next step: `/execute docs/plans/vN_YYYY-MM-DD-<codename>.md`

## Examples

```
/plan docs/specs/v20_spec-play-session-loop.md
→ Review spec → ask about open questions → write docs/plans/v20_2026-06-06-play-session-loop.md
```

```
/plan docs/specs/v10_spec-click-action-and-melee-range.md
→ Must include Task for 04_door_lab.json + migration of 01/02/03 scenarios
```
