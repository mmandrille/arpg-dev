# v201 Plan - Item Level Tooltip

Status: Ready for implementation
Goal: Show authoritative item levels in shared item tooltips while preserving requirement-level visibility.
Architecture: Keep this as a client presentation slice. The server already sends `item_level`; `ItemTooltipPanel` will prefer explicit item-level metadata and only collapse level requirements into the footer when no item level exists.
Tech stack: Godot client scripts/tests, SDD docs.

## Baseline and shortcut decision

Builds on v196 item-level payload propagation and v197 blacksmith presentation. Asset/plugin decision: reject external assets/plugins; reuse the existing shared tooltip component.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/item_tooltip_panel.gd` | Prefer explicit item-level footer; preserve requirement display when both exist. |
| Modify | `client/tests/test_shop_panel.gd` | Cover item-level footer and requirement visibility. |
| Add | `docs/as-built/v201_item-level-tooltip.md` | Record shipped behavior and proof. |
| Modify | `PROGRESS.md` | Update current status after completion. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v201 lifecycle row. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [ ] `client/scripts/main.gd`
- [ ] `server/internal/game/game_test.go`
- [ ] `tools/bot/run.py`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: `client/tests/test_shop_panel.gd`
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [ ] Extract focused helper/module/test file as part of this slice, or
- [x] Defer extraction with rationale: tooltip assertions belong beside the existing shop/inventory/market tooltip coverage, and this slice adds only a few focused lines to a grandfathered test.

Verification:
```bash
make maintainability
```

## Task 1 - Tooltip behavior

Files:
- Modify: `client/scripts/item_tooltip_panel.gd`

- [x] Step 1.1: Detect explicit `item_level` from top-level or nested stats payloads.
- [x] Step 1.2: Show `Item level N` for explicit item-level metadata.
- [x] Step 1.3: Keep existing requirement-level footer behavior for items without explicit item level.
- [x] Step 1.4: Preserve level requirements in the requirements block when an item-level footer is shown.

```bash
make client-unit
```

## Task 2 - Focused client proof

Files:
- Modify: `client/tests/test_shop_panel.gd`

- [x] Step 2.1: Add/update tooltip assertions for item-level footer text.
- [x] Step 2.2: Assert level requirements remain visible when item level exists.
- [x] Step 2.3: Assert existing requirement-level fallback remains for items without item level.

```bash
make client-unit
```

## Task 3 - Lifecycle docs and CI

Files:
- Add: `docs/as-built/v201_item-level-tooltip.md`
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`

- [x] Step 3.1: Mark the spec complete after verification.
- [x] Step 3.2: Add as-built proof notes.
- [x] Step 3.3: Update progress status and lifecycle row.

```bash
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make client-unit`
- [x] `make ci`
