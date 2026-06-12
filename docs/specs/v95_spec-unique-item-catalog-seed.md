# v95 Spec: Unique Item Catalog Seed

Status: Approved for implementation from `$autoloop 3`
Date: 2026-06-12
Codename: `unique-item-catalog-seed`

## Purpose

Create the first schema-backed unique item catalog without making uniques live loot yet. The seed
catalog defines one disabled unique concept linked to an existing item template and records the
future behavior hook it is meant to unlock.

## Non-goals

- No unique drops, shop/mystery eligibility, equip behavior, combat effects, client UI, art, or
  market restrictions.
- No stat-only player-facing unique. ADR-0014 D5 requires future live uniques to change behavior.
- No unique-effect engine or skill mutation implementation.

## Acceptance criteria

- `shared/rules/unique_items.v0.schema.json` validates the unique item catalog shape.
- `shared/rules/unique_items.v0.json` defines one disabled unique with a stable id, base template,
  display name, minimum level, behavior hook summary, and explicit `enabled: false`.
- `tools/validate_shared.py` validates the catalog and cross-checks base templates.
- `make validate-shared`, `make maintainability`, `make test-go`, and `make ci` pass.

## Risks

| Risk | Mitigation |
|------|------------|
| A catalog-only unique could violate ADR-0014 D5 if treated as live content. | Keep `enabled: false`, exclude all loot/runtime paths, and document behavior-effect implementation as deferred. |

## ADR alignment

- ADR-0014 D5: unique items must change behavior; this slice records the behavior hook but does not
  make the item obtainable until an effect engine exists.
- ADR-0013: prepares future mystery-seller unique eligibility without changing current max rarity.
