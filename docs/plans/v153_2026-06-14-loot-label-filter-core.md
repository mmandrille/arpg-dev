# v153 Plan — Loot Label Filter Core

Status: Ready for implementation
Goal: Add a client-side rarity threshold filter for ground loot labels (All → Magic+ → Rare+ →
Unique), display-only, with the logic in a focused new script so `main.gd` does not grow.
Architecture: A new `LootLabelFilter` (RefCounted) owns the rarity ladder + threshold + `allows()`.
`main.gd` holds an instance and gates non-hovered label visibility through it in the existing
`_refresh_loot_label_visibility()` chokepoint; a keybind cycles it and refreshes labels; the current
mode shows in the existing `status_text` overlay. No server/protocol/shared change.
Tech stack: Godot 4 GDScript, headless client unit test, full CI.

## Baseline and shortcut decision

Builds on v152 (`make ci` green at commit `982ccfd3`). First feature slice after the v141–v152
maintenance arc; proves the architecture flows for player features again and that new code lands in
a focused file under the touch-to-shrink rule. Godot plugin adopt/borrow/reject: **reject** — no
loot-filter plugin exists in `docs/researchs/godot-plugins-and-shortcuts.md`; this is a few lines of
custom presentation over existing `Label3D` labels.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `client/scripts/loot_label_filter.gd` | Rarity ladder + threshold; `allows`, `cycle`, `mode_label`. |
| Modify | `client/scripts/main.gd` | Hold filter instance; gate reveal visibility; keybind; status readout. |
| Create | `client/tests/test_loot_label_filter.gd` | Headless unit coverage of the filter logic. |
| Modify | `docs/CODEMAP.md` | Add the new script + test. |
| Modify | `PROGRESS.md` | Lifecycle closeout. |
| Create | `docs/as-built/v153_loot-label-filter-core.md` | Close-out proof. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [ ] `client/scripts/main.gd` — must stay at or below baseline; only an instance field, one
  visibility gate, and one keybind branch are added. Filter logic lives in the new script.
- [ ] Other over-limit file: none.
- [ ] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)? Verify
  via `make maintainability` after the edits; if `main.gd` would exceed, extract an equivalent
  helper in the same slice.

Decision:
- [x] Extract focused helper as part of this slice (`loot_label_filter.gd`).

Verification:
```bash
make maintainability
```

## Task 1 — Filter model + unit test

Files:
- Create: `client/scripts/loot_label_filter.gd`, `client/tests/test_loot_label_filter.gd`

- [ ] Step 1.1: `class_name LootLabelFilter extends RefCounted` with `RARITY_ORDER := ["common","magic","rare","unique"]`, a `_threshold := 0`, `allows(rarity)`, `cycle()`, `mode_label()`.
- [ ] Step 1.2: `allows` true when `rank(rarity) >= _threshold`; off-ladder rarities (currency/quest) always true.
- [ ] Step 1.3: Unit test ordering, per-mode thresholds, `cycle` wraparound, off-ladder always allowed.
```bash
make client-unit
```

## Task 2 — Integrate into main.gd

Files:
- Modify: `client/scripts/main.gd`

- [ ] Step 2.1: Add `var _loot_filter := LootLabelFilter.new()` (preload the script).
- [ ] Step 2.2: In `_refresh_loot_label_visibility()`, non-hovered label shows only when
  `loot_label_reveal_held and _loot_filter.allows(rarity)`; hovered/highlighted always shows.
- [ ] Step 2.3: Add a non-colliding keybind in `_unhandled_input` (verify against WASD/Shift/Alt/zoom)
  that calls `_loot_filter.cycle()`, refreshes labels, and updates the `status_text` readout.
- [ ] Step 2.4: Confirm `main.gd` stays ≤ its grandfathered baseline.
```bash
make client-unit
make maintainability
```

## Task 3 — CODEMAP, lifecycle, CI

Files:
- Modify: `docs/CODEMAP.md`, `PROGRESS.md`
- Create: `docs/as-built/v153_loot-label-filter-core.md`

- [ ] Step 3.1: Add `loot_label_filter.gd` + `test_loot_label_filter.gd` to CODEMAP.
- [ ] Step 3.2: PROGRESS lifecycle row + as-built.
```bash
make ci
```

## Final verification

- [ ] `make client-unit` (new test + main.gd loads)
- [ ] `make maintainability` (main.gd ≤ baseline)
- [ ] `make ci`

## Deferred scope

- Category (currency/quest/consumable) filtering and a min-rarity-by-category matrix.
- A dedicated always-on HUD indicator widget and persisting the chosen filter via `client_settings`.
- Bot-visual scenario to showcase the filter (`make bot-visual scenario=...`) once a visible
  indicator lands.
