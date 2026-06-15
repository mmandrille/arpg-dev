# v187 Spec: Rarity Roll Pools

Status: Complete - make ci green on 2026-06-15
Date: 2026-06-15
Codename: rarity-roll-pools

## Purpose

Make rolled equipment properties scale by item rarity instead of relying on one flat template pool
and one fixed `stat_rolls` count. Magic items roll a small number of base improvements, rare items
gain access to stronger combat-affix candidates, and set/unique items gain access to the full
higher-rarity pool while keeping server-authoritative deterministic rolling.

This slice establishes the shared-data contract that later slices can use for live combat effects
and skill-specific affixes. It should be visible through existing item tooltips, persisted rolled
payloads, deterministic item-roll goldens, and a protocol bot proof.

## Non-goals

- No prefix/suffix name grammar or procedural affix text.
- No blacksmith/crafting route that adds or improves rolls.
- No final balance pass for exact stat ranges, rarity weights, or drop rates.
- No new set or unique catalog content.
- No per-skill cooldown or mana-cost affixes; v189 handles skill-affix behavior.
- No new inventory UI plugin or tooltip redesign.

## Acceptance Criteria

- Shared `item_templates.v0` supports rarity roll-count ranges, with magic configured as 1-2 rolls,
  rare as 2-4 rolls, and set/unique as 3-5 rolls.
- Shared item templates support roll candidates gated by a minimum rarity, and higher rarities
  inherit lower-rarity roll candidates.
- Server item rolling chooses the roll count through the seeded item RNG and remains deterministic
  for the same seed and input order.
- Existing base/template stat rolls remain valid without duplicating every stat in every rarity.
- The deterministic item-roll golden includes cases proving magic, rare, and unique/set roll counts
  and inherited pools.
- Rolled item payloads continue to persist and reload through the existing `rolled_stats` payload
  without a protocol schema bump.
- Existing inventory/stash/shop/loot tooltip paths can display any newly introduced stat keys with
  readable labels.
- A bot scenario proves at least one deterministic higher-rarity item has a roll count in range and
  a stat drawn from the expected inherited pool.
- `make validate-shared`, focused Go tests, `make bot`, `make client-unit` if client labels change,
  and `make ci` pass.

## Scope And Likely Files

- `shared/rules/item_templates.v0.schema.json`
- `shared/rules/item_templates.v0.json`
- `shared/golden/item_rolls.json`
- `shared/golden/item_rolls.v0.schema.json`
- `server/internal/game/rules.go`
- `server/internal/game/shop.go`
- `server/internal/game/game_test.go` or focused item-roll tests
- `client/scripts/stat_labels.gd` if new labels are needed
- `tools/bot/run.py` and/or bot assertion helpers if existing assertions are insufficient
- `tools/bot/scenarios/78_rarity_roll_pools.json`
- `docs/as-built/v187_rarity-roll-pools.md`
- `PROGRESS.md`

## Test And Bot Proof

- Go item-roll tests prove the configured roll-count ranges and minimum-rarity gating.
- The item-roll golden pins deterministic examples for magic, rare, and unique/set outcomes.
- `make validate-shared` validates the schema and catalog.
- `make bot scenario=78_rarity_roll_pools.json` proves a rolled item exposes the expected rarity,
  roll count, and stat pool behavior through the player-facing protocol path.
- Manual visual check, if desired: `make bot-visual scenario=78_rarity_roll_pools.json`.

## Open Questions And Risks

| Risk | Mitigation |
|------|------------|
| Changing roll counts shifts deterministic golden payloads. | Update item-roll goldens deliberately and keep assertions focused on rule-derived ranges. |
| New stat keys might display as raw ids. | Add labels only for keys introduced by this slice. |
| `set` is currently fixed-catalog, not a normal random rarity. | Support the roll-count contract for set payload construction where used, but do not add random set drops. |
| Higher roll counts can duplicate the same stat. | Preserve current additive duplicate-stat behavior; affix uniqueness is deferred. |

## ADR Alignment

- ADR-0001 D2/D6/D8: server owns rolled outcomes, shared rules stay declarative, deterministic
  replay remains protected by seeded RNG.
- ADR-0012: prepares item progression and future add/improve-roll upgrades without implementing
  crafting in this slice.
