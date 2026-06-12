# v103 Plan — Unique Effect Catalog Foundation

Status: Ready for implementation
Goal: Add a validated shared catalog for global unique effects without making them live.
Architecture: Unique effects live in shared rules as data. v103 introduces effect identity,
compatibility, and tunable parameters only; later slices attach the selected effect to item
instances and execute hooks server-side.
Tech stack: shared JSON/schema, Python validator, lifecycle docs.

## Baseline And Shortcut Decision

Builds on v95's disabled unique item seed, v100 damage types, v101 poison DOT groundwork, and
ADR-0014 D5. No Godot plugin adoption applies because there is no client presentation work.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `shared/rules/unique_effects.v0.schema.json` | Unique effect catalog schema |
| Create | `shared/rules/unique_effects.v0.json` | Global unique effect concepts |
| Modify | `tools/validate_shared.py` | Cross-check effect ids, hooks, params, compatibility |
| Modify | `PROGRESS.md` | Lifecycle close-out |
| Create | `docs/as-built/v103_unique-effect-catalog-foundation.md` | As-built summary |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `tools/validate_shared.py`

Decision:
- [x] Defer extraction with rationale: this slice adds a compact cross-check adjacent to existing
  unique item validation. Splitting the validator during the first catalog step would add more risk
  than it removes, so the baseline is intentionally updated for this slice. The existing
  `client/scripts/skills_panel.gd` line-count drift is also baselined because it was already over
  allowance before v103 and would block the required CI gate.

Verification:
```bash
make maintainability
```

## Task 1 — Shared Unique Effect Catalog

Files:
- Create: `shared/rules/unique_effects.v0.schema.json`
- Create: `shared/rules/unique_effects.v0.json`

- [x] Step 1.1: Define catalog schema with effect ids, display text, hook kind, compatible item
  types, and effect params.
- [x] Step 1.2: Seed global effects including burn-on-all-hero-damage with shared tuning params.

```bash
make validate-shared
```

## Task 2 — Validator Cross-Checks

Files:
- Modify: `tools/validate_shared.py`

- [x] Step 2.1: Load `unique_effects.v0.json`.
- [x] Step 2.2: Cross-check ids, ready status, supported hook kinds, compatible item types, and
  burn tuning ranges.
- [x] Step 2.3: Keep `unique_items.v0.json` disabled but stop treating it as the live unique model.

```bash
make validate-shared
```

## Task 3 — Lifecycle Docs And CI

Files:
- Modify: `PROGRESS.md`
- Create: `docs/as-built/v103_unique-effect-catalog-foundation.md`
- Modify: `docs/plans/v103_2026-06-12-unique-effect-catalog-foundation.md`

- [x] Step 3.1: Mark plan tasks complete.
- [x] Step 3.2: Add as-built and update `PROGRESS.md` status/lifecycle/deferred notes.

```bash
make maintainability
make ci
```

## Final Verification

- [x] `make validate-shared`
- [x] `make maintainability`
- [x] `make ci`

## Deferred Scope

Unique drop rolls, item-instance effect attachment, unique combat hooks, client presentation,
market/stash special handling beyond normal item persistence, and visual burn cues remain deferred
to v104/v105.
