# v289 As Built: Mercenary Offer Variants

Date: 2026-06-19
Spec: [`docs/specs/v289_spec-mercenary-offer-variants.md`](../specs/v289_spec-mercenary-offer-variants.md)
Plan: [`docs/plans/v289_2026-06-19-mercenary-offer-variants.md`](../plans/v289_2026-06-19-mercenary-offer-variants.md)

## What shipped

- Added `mercenaries.v0.json` plus schema validation for authored mercenary offers.
- Added `mercenary_scout` as a no-drop/no-XP mercenary monster with existing in-repo presentation
  assets, localized names, and regenerated model preview catalog metadata.
- Server rules now load and validate mercenary offers, reject empty/duplicate/blank/unknown
  references, and select the board offer deterministically from session seed plus board entity id.
- Mercenary board open/hire events, spawned companion metadata, and `mercenary_lost` now use the
  selected offer and monster definition instead of hardcoded guard fields.
- The Godot mercenary panel displays the scout variant and focused panel tests cover scout offer,
  status, and stats-card text.
- Added protocol scenario `96_mercenary_offer_variants` and client scenario
  `69_mercenary_offer_variant_ui` to prove the pinned scout offer through server and UI flows.

## Proof

Focused verification:

```bash
make validate-shared
make validate-assets
(cd server && go test ./internal/game -run 'TestMercenary' -count=1)
godot --headless --path client --script res://tests/test_mercenary_panel.gd
make bot scenario=96_mercenary_offer_variants
make bot-client scenario=69_mercenary_offer_variant_ui HEADLESS=1
make maintainability
```

Result: green on 2026-06-19.

Full verification:

```bash
make ci
```

Result: deferred until the end of the selected autoloop queue.

## Manual visual command

```bash
make bot-visual scenario=96_mercenary_offer_variants
```

## Deferred

- Player-character-derived listings, snapshots, persistence, player-set prices, and offer picker UI
  remain deferred.
- Per-offer pricing, multiple simultaneous hired mercenaries, ranged mercenary AI, new art assets,
  loot, XP, leveling, and per-companion commands remain deferred.
