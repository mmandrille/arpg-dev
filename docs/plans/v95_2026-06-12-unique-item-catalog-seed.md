# v95 Plan — Unique Item Catalog Seed

Status: Ready for implementation
Goal: Add a disabled, schema-backed unique item catalog seed with validator coverage.
Architecture: The catalog is shared data only. It is validated by tooling but not loaded into runtime loot, shops, mystery seller, market, or client presentation paths.
Tech stack: shared JSON/schema, Python validator, lifecycle docs.

## Baseline and shortcut decision

Builds on item templates, mystery-seller ADR direction, and ADR-0014 D5. No Godot plugin adoption applies because there is no UI/art work.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `shared/rules/unique_items.v0.schema.json` | Unique catalog schema |
| Create | `shared/rules/unique_items.v0.json` | Disabled first unique concept |
| Modify | `tools/validate_shared.py` | Catalog validation and template cross-checks |
| Modify | `PROGRESS.md` | Lifecycle close-out |
| Create | `docs/as-built/v95_unique-item-catalog-seed.md` | As-built summary |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `tools/validate_shared.py`

Decision:
- [x] Defer extraction with rationale: the validator already centralizes shared-data cross-checks;
  this slice adds a small catalog check and records the baseline if needed.

Verification:
```bash
make maintainability
```

## Task 1 — Shared Catalog

- [x] Add the schema.
- [x] Add one disabled unique item concept.

```bash
make validate-shared
```

## Task 2 — Validator

- [x] Load `unique_items.v0.json`.
- [x] Cross-check base templates and disabled seed status.

```bash
make validate-shared
```

## Task 3 — Lifecycle Docs And CI

- [x] Update plan checkboxes, `PROGRESS.md`, and as-built.

```bash
make maintainability
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `make maintainability`
- [x] `make test-go`
- [x] `make ci`

## Deferred scope

Unique drops, unique effects, skill/build behavior changes, client presentation, mystery-seller
unique eligibility, and market restrictions remain deferred.
