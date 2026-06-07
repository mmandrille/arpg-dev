---
name: next
description: >-
  Propose the next development slice from PROGRESS.md backlog and ADRs, or
  evaluate a user idea. Use when the user runs /next, asks what to build next,
  or wants a slice brief before writing a spec.
disable-model-invocation: true
---

# /next — Next Slice Discovery

**Trigger:** `/next` or `/next {optional idea}`

Examples:

- `/next` — agent proposes candidates from backlog and trajectory.
- `/next play session loop` — agent evaluates the idea against baseline and produces a spec-ready brief.
- `/next character persistence` — same, focused on the user's theme.

**Announce at start:** "Using the **next** skill to identify the next slice and prepare a spec brief."

## Hard rules

1. **Do not write code.** Do not implement. Do not run `/execute`.
2. **Do not write the spec file** unless the user explicitly asks after the brief — this skill produces the **brief**, not `docs/specs/vN_spec-*.md`.
3. **Read before proposing** — baseline must come from `PROGRESS.md` and as-built code, not assumptions.
4. **Ask the user** when slice priority is unclear or multiple valid paths exist.

## Phase 0 — Baseline context (read first)

1. [`docs/PROGRESS.md`](../../docs/PROGRESS.md) — latest completed slice, active branch, open gaps, deferred backlog.
2. [`CLAUDE.md`](../../CLAUDE.md) — architecture, invariants, slice pattern.
3. [`docs/adr/0001-technology-stack.md`](../../docs/adr/0001-technology-stack.md) — if not already familiar.
4. Relevant ADRs for the candidate area (e.g. [`0008-world-structure-and-dungeon-progression.md`](../../docs/adr/0008-world-structure-and-dungeon-progression.md) for world/progression).
5. Existing specs/plans without a completed lifecycle row (in-progress or drafted slices).
6. If user provided an idea — treat it as the primary candidate; still validate against baseline.

### Determine next slice number

- Highest `vN` in `docs/specs/` and `docs/plans/` lifecycle table → next is **v(N+1)** unless a drafted spec already claims a number.
- If `v20_spec-…` exists but v20 is not in the lifecycle table, v20 may be the active in-flight slice — say so explicitly.

## Phase 1 — Candidate discovery

### Without user idea

Survey these sources and rank 2–4 candidates:

| Source | What to extract |
|--------|-----------------|
| `PROGRESS.md` → **Open gaps & deferred work** | Documented backlog; do not treat as automatic next slice |
| `PROGRESS.md` → **Current status** | `Next slice: TBD`, active branch, latest completed |
| ADR deferred decisions | e.g. ADR-0008 D1 character persistence, D3 PCG density |
| In-flight specs | Specs/plans written but not marked complete |
| Natural trajectory | What logically follows the latest completed slice (e.g. v19 teleporters → v20 play loop) |
| Bot/scenario gaps | Missing end-to-end proof for a feature area |

Score each candidate on:

- **Player value** — does it move `make play` or core loop forward?
- **Architectural leverage** — unblocks how many future slices?
- **Complexity** — S / M / L / XL with one-line rationale.
- **Risk** — protocol changes, determinism, cross-language golden, client+server coupling.
- **Dependencies** — must ship after/before which slice?

### With user idea

Evaluate the idea directly using the same scoring. Compare briefly to top backlog alternative(s) so the user can confirm or pivot.

## Phase 2 — Slice brief (output)

Produce a structured brief for the **recommended** slice (or the user's idea if viable). Use this template in chat:

```markdown
## Next slice recommendation

**Proposed:** v{N} — `{codename}` — {human title}
**Baseline:** v{X} `{prior-codename}` complete
**Branch:** current checkout (do not create branches)

### Why this slice now
{2–4 sentences: player value, what it proves, why not something else}

### Complexity
**Size:** S | M | L | XL
**Touch surfaces:** shared | server | client | bot | docs (check all that apply)
**Estimated tasks:** {rough count}

### Purpose (spec §1 draft)
{What the slice does in plain language}

### Non-goals (spec §2 draft)
- {explicit deferrals}

### Likely files / contracts
{Bullet list of paths — protocol, rules, sim, client, bot scenarios}

### Bot proof
{New scenario `NN_<lab>.json`? migrate existing? or explicit deferral reason}

### Requirements checklist
- [ ] {verifiable acceptance criterion}
- [ ] …

### Open questions / doubts
| # | Question | Default if unanswered |
|---|----------|---------------------|
| Q-1 | … | … |

### ADR alignment
{Which ADR decisions this honors or defers; flag conflicts}

### Alternatives considered
| Slice | Why not now |
|-------|-------------|
| … | … |
```

### Quality bar

- Every acceptance criterion must be **verifiable** (bot step, test command, or observable behavior).
- Call out **protocol/schema bumps** explicitly.
- Call out **determinism** risks for Go `game/` changes.
- For client UI/art — note Godot plugin adoption will be required in plan (`docs/researchs/godot-plugins-and-shortcuts.md`).
- If the idea is too large → propose a **thin vertical slice** and defer the rest to non-goals.

## Phase 3 — Handoff

End with:

1. **Recommendation** — one primary slice (or "idea needs refinement" with blockers).
2. **User decision** — ask which candidate to pursue if multiple remain viable.
3. **Next commands:**
   - Write spec: create `docs/specs/vN_spec-<codename>.md` from the brief (or ask agent to draft it).
   - `/plan docs/specs/vN_spec-<codename>.md`
   - `/execute docs/plans/vN_<date>-<codename>.md`
   - `/finish` when implementation is done.

If blockers exist in the brief, **stop** and resolve questions before spec writing.

## Examples

```
/next
→ Read PROGRESS (v19 complete) → propose v20 play-session-loop + 2 alternatives → brief for v20
```

```
/next town safe zone
→ Evaluate idea vs ADR-0008 → complexity M → doubts about combat rules in town → brief with Q-1..Q-3
```
