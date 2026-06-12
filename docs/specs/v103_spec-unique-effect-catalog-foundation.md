# v103 Spec: Unique Effect Catalog Foundation

Status: Approved for implementation from `$autoloop 3`
Date: 2026-06-12
Codename: `unique-effect-catalog-foundation`

## Purpose

Create the shared, schema-backed catalog for unique effects. A unique item is a normal rolled
equipment item with one selected unique effect attached; it is not a fixed unique base item. This
slice defines that effect catalog and validation boundary without making unique drops live yet.

## Non-goals

- No unique drop rolls, item-instance effect attachment, shop/mystery eligibility, market behavior,
  combat execution, client UI, VFX, or bot scenario.
- No fixed one-unique-per-item-template model. Effects are globally selectable and may later be
  filtered by compatibility.
- No final unique effect balance.

## Acceptance Criteria

- `shared/rules/unique_effects.v0.schema.json` validates a global unique-effect catalog.
- `shared/rules/unique_effects.v0.json` defines at least three ready effect concepts, including
  a burn effect where all hero damage applies a 10-second burn ticking once per second for 10% of
  the original hit damage.
- Unique effects define behavior hooks and compatibility metadata, not rolled stats.
- `tools/validate_shared.py` cross-checks unique-effect ids, hook kinds, numeric params, and
  compatible item types against `item_templates.v0.json`.
- The legacy `unique_items.v0.json` seed remains valid as disabled concept data, but new runtime
  work targets `unique_effects.v0.json`.
- `make validate-shared`, `make maintainability`, and `make ci` pass.

## Scope And Likely Files

- `shared/rules/unique_effects.v0.schema.json`
- `shared/rules/unique_effects.v0.json`
- `tools/validate_shared.py`
- `docs/plans/v103_2026-06-12-unique-effect-catalog-foundation.md`
- `docs/as-built/v103_unique-effect-catalog-foundation.md`
- `PROGRESS.md`

## Test And Bot Proof

- Shared validation proves the new catalog structure and cross-references.
- No bot scenario is required because the slice does not make effects obtainable or executable.
- Final proof is `make validate-shared`, `make maintainability`, and `make ci`.

## Open Questions And Risks

| Risk | Mitigation |
|------|------------|
| The old unique item seed could imply fixed unique bases. | Keep it disabled and document `unique_effects.v0.json` as the forward runtime contract. |
| Effect tuning may become hardcoded later. | Store burn cadence and percent in shared data now. |

## ADR Alignment

- ADR-0014 D5: unique effects change behavior, not just stats.
- ADR-0012: leaves ownership and market eligibility untouched until item instances carry effects.
- ADR-0013: prepares future mystery-seller unique eligibility without making uniques purchasable.
