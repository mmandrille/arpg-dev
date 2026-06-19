# v289 Spec: Mercenary Offer Variants

Status: Implemented
Date: 2026-06-19
Codename: `mercenary-offer-variants`

## Purpose

Make the mercenary board able to offer more than the single fixed guard. This slice adds a small
server-authored offer catalog, introduces a second mercenary archetype, and proves that a pinned
session seed can hire that variant through the same authoritative board workflow.

## Non-goals

- Do not add player-character-derived mercenary listings, snapshots, persistence, or player-set
  prices.
- Do not add a multi-offer picker UI; the board still presents and hires one selected offer.
- Do not add multiple simultaneous hired mercenaries; hiring still replaces the previous board hire.
- Do not add per-offer pricing; variants use the existing configured mercenary hire cost.
- Do not add new art assets, external plugins, ranged mercenary AI, inventory, loot, XP, leveling, or
  per-companion commands.

## Acceptance Criteria

- Shared rules include a `mercenaries.v0.json` catalog with at least `fixed:mercenary_guard` and
  `fixed:mercenary_scout` offers.
- Shared validation and Go loading reject empty offer catalogs, duplicate offer ids, blank monster
  references, and offers that reference unknown monsters.
- Shared monster, visual, and i18n data define `mercenary_scout` as a no-drop/no-XP companion-style
  mercenary using existing in-repo monster presentation assets.
- The mercenary board deterministically selects one catalog offer from the session seed and board id.
- Existing guard-hire behavior remains valid for a pinned guard seed.
- A pinned scout seed emits `mercenary_board_opened` and `mercenary_hired` with
  `offer_id=fixed:mercenary_scout` and `monster_def_id=mercenary_scout`, spends the configured hire
  cost, and spawns an owned scout companion.
- `mercenary_lost` reports the selected hired offer and monster definition for variant hires.
- Protocol bot proof hires the scout variant and observes companion-sourced damage.
- Client bot proof shows the mercenary panel/companion HUD for the scout variant.

## Scope And Likely Files

- Shared rules/schema: `shared/rules/mercenaries.v0.json`,
  `shared/rules/mercenaries.v0.schema.json`, `shared/rules/monsters.v0.json`,
  `shared/assets/monster_visuals.v0.json`, `shared/assets/model_preview_catalog.v0.json`,
  `shared/i18n/en.json`, `shared/i18n/es.json`.
- Server: `server/internal/game/mercenary_rules.go`, `server/internal/game/mercenary_hiring.go`,
  `server/internal/game/monster_companion_combat.go`, `server/internal/game/rules.go`,
  `server/internal/game/mercenary_hiring_test.go`.
- Client: `client/scripts/mercenary_panel.gd`, `client/tests/test_mercenary_panel.gd` if display
  helpers need coverage.
- Bot: `tools/bot/scenarios/96_mercenary_offer_variants.json` and
  `tools/bot/scenarios/client/69_mercenary_offer_variant_ui.json`.
- Docs: v289 plan/as-built/lifecycle updates.

## Test And Bot Proof

Focused checks:

```bash
make validate-shared
make validate-assets
(cd server && go test ./internal/game -run 'TestMercenary' -count=1)
godot --headless --path client --script res://tests/test_mercenary_panel.gd
make bot scenario=96_mercenary_offer_variants
make bot-client scenario=69_mercenary_offer_variant_ui HEADLESS=1
make maintainability
```

Visual verification command for humans/agents:

```bash
make bot-visual scenario=96_mercenary_offer_variants
```

## Asset And Plugin Decision

- Adopt: existing companion AI, existing mercenary board workflow, existing companion HUD/panel, and
  existing monster dummy visual asset.
- Borrow: existing mercenary hiring protocol and client bot scenario structure.
- Reject: external assets/plugins, new model pipelines, offer-picker UI, and player-character
  snapshot persistence.

## Outcome

- Implemented with a server-authored guard/scout offer catalog, deterministic seed+board selection,
  variant-aware events/spawn/loss metadata, and protocol plus client bot proof for the scout offer.
- The maintainability ratchet stayed green by keeping mercenary catalog logic in
  `mercenary_rules.go` and moving the existing drop-rate helper out of `rules.go`.
