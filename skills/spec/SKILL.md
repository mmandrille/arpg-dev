---
name: spec
description: Draft or update SDD slice specs under docs/specs. Use when the user runs /spec or $spec, asks to turn a next-slice brief or idea into a spec, or wants a docs/specs/vN_spec-*.md file before planning.
---

# Spec

Use this skill to write a focused SDD slice spec. Do not implement code or write an execution plan here.

## Workflow

1. Announce: "Using the **spec** skill to draft the slice spec."
2. Read `CLAUDE.md`, `PROGRESS.md`, and any next brief, ADR, plan, or prior spec the user references.
3. Determine the correct `vN` from existing `docs/specs/`, `docs/plans/`, and the lifecycle table in `PROGRESS.md`. If a draft already owns the next number, update it instead of creating a duplicate.
4. For client UI, inventory presentation, camera tooling, or placeholder art, read `docs/researchs/godot-plugins-and-shortcuts.md` and include an adopt / borrow / reject note in the spec or call out that the plan must record it.
5. Write `docs/specs/vN_spec-<codename>.md` with concise sections that are specific enough for `/plan`.
6. Stop after the spec and summarize the file path plus any open questions that block planning.

## Required Sections

- Title, status, date, and codename.
- Purpose: what user-visible or system behavior changes.
- Non-goals: what is intentionally deferred.
- Acceptance criteria: observable checks, including gameplay/UI behavior.
- Scope and files likely touched: contracts, client, server, tools, docs, tests.
- Test and bot proof: unit, smoke, bot, visual, or replay coverage expected.
- Open questions and risks: only unresolved items that matter before planning.

## Guardrails

- Keep specs small enough for one implementation slice.
- Prefer updating contracts, fixtures, tests, and docs together over preserving stale compatibility.
- Do not invent plugin or asset choices without checking the project plugin research when visual/client work is in scope.
- If the brief is too vague to produce acceptance criteria, ask the minimum blocking question before writing the file.
