# v181 Spec: Set Item Foundation

Status: Complete - make ci green on 2026-06-15
Date: 2026-06-15
Codename: set-item-foundation

## Purpose

Add the first green set-item rarity as a five-piece fixed package that can be tested from the existing debug unique chest. Equipping more pieces from the set increases server-authoritative item power, and equipping all five grants a stronger full-set buff.

## Non-goals

- Random set drops, set-specific loot tables, boss rewards, or economy pricing.
- A separate production set collection UI.
- New art assets beyond existing item icons, colors, labels, and chest testing flow.

## Acceptance Criteria

- Shared rules define `rarity: set` plus one enabled five-piece set catalog.
- The debug unique chest offers all five set items in addition to current unique items when gameplay debug is enabled.
- Set items render with green rarity color in existing loot labels and item tooltips.
- Equipping two, three, four, and five pieces adds deterministic server-owned set bonuses.
- Equipping all five pieces grants a powerful full-set buff through existing stat/effective-rank paths.
- `make validate-shared`, focused Go tests, client unit tests, and `make ci` pass.

## Scope And Likely Files

- Shared rules: `shared/rules/item_templates.v0.json`, `shared/rules/set_items.v0.json`, schema.
- Server: rules loading/validation, set payload construction, set bonus aggregation, unique chest contents.
- Client: rarity colors for set items in existing inventory/stash/shop/loot surfaces.
- Tests/docs: focused Go tests, client item visual/shop panel assertions, plan, as-built, `PROGRESS.md`.

## Test And Bot Proof

- Focused Go tests cover catalog payloads, debug chest inclusion, deterministic order, partial bonuses, and full-set buff.
- Client unit tests cover green rarity color on labels/tooltips.
- Existing chest bot scenario remains the manual/e2e testing path: `make bot scenario=61_purple_town_unique_chest`.

## Risks

- Set bonuses must not mutate persisted item stats; they are derived from currently equipped set pieces.
- Full-set bonuses must flow through existing server stat/rank aggregation, not client-side display shortcuts.
