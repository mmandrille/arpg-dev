# v236 Plan - Mystery Seller Silhouettes

Status: Complete
Goal: Add safe slot-derived visual silhouettes to concealed mystery seller offers.
Architecture: Keep server contracts unchanged; derive a local silhouette key from visible
slot/category metadata and render it in the existing Godot shop panel.
Tech stack: Godot UI/client bot, docs.

## Baseline and Asset Decision

Builds on v64 mystery seller rerolls and v230 mystery set/unique eligibility. Asset/plugin decision:
reject external assets/plugins; silhouettes are code-drawn shapes based only on visible slot/category
metadata.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/shop_panel.gd` | Draw silhouettes, expose debug/summary clue |
| Add | `client/scripts/mystery_silhouette_drawer.gd` | Derive silhouette keys and render code-drawn shapes |
| Modify | `client/tests/test_shop_panel.gd` | Prove ring clue and identity hiding |
| Add | `tools/bot/scenarios/client/53_mystery_seller_silhouettes.json` | Client proof |
| Add | `docs/as-built/v236_mystery-seller-silhouettes.md` | Slice proof |

## Maintenance Ratchet

Target: touched source/test/tool files stay at or below their allowed baselines.

Hotspot / over-limit files touched:
- [x] `client/scripts/shop_panel.gd`
- [x] `client/tests/test_shop_panel.gd`

Decision:
- [x] Reuse existing shop debug state and bot `summary_contains` support instead of adding bot
  runner actions.
- [x] Extract the silhouette renderer into a small helper to keep the shop panel within ratchet.

Verification:
```bash
make maintainability
```

## Task 1 - Silhouette model and rendering

Files:
- Modify: `client/scripts/shop_panel.gd`

- [x] Add a local silhouette key derived from `slot` first, then `category`.
- [x] Render simple slot/category shapes in `_draw_mystery_icon`.
- [x] Add a visible detail line such as `Silhouette: Ring` for mystery offers.
- [x] Expose `mystery_silhouette` in debug offer rows.

## Task 2 - Focused client proof

Files:
- Modify: `client/tests/test_shop_panel.gd`
- Add: `tools/bot/scenarios/client/53_mystery_seller_silhouettes.json`

- [x] Assert the ring mystery row exposes the ring silhouette clue.
- [x] Assert the existing identity-hiding and no-preview checks still pass.
- [x] Add a bot scenario that opens the mystery seller and verifies visible summary rows include the
  silhouette clue.

```bash
godot --headless --path client --script res://tests/test_shop_panel.gd
make bot-client scenario=53_mystery_seller_silhouettes.json HEADLESS=1
```

## Task 3 - Lifecycle docs

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `docs/progress/scenario-catalog.md`
- Add: `docs/as-built/v236_mystery-seller-silhouettes.md`

- [x] Record verification and note that server contracts remain unchanged.

## Final Verification

- [x] `godot --headless --path client --script res://tests/test_shop_panel.gd`
- [x] `make bot-client scenario=53_mystery_seller_silhouettes.json HEADLESS=1`
- [x] `make maintainability`
