# v153 As-Built: Loot Label Filter Core

Date: 2026-06-14
Codename: `loot-label-filter-core`
Spec: [`v153_spec-loot-label-filter-core.md`](../specs/v153_spec-loot-label-filter-core.md)
Plan: [`v153_2026-06-14-loot-label-filter-core.md`](../plans/v153_2026-06-14-loot-label-filter-core.md)

## What shipped

First player-facing feature after the v141–v152 maintenance arc: a client-side, display-only rarity
threshold filter for ground loot labels. Holding the reveal key (Alt) shows ground labels; the new
filter narrows that reveal to At-or-above a chosen rarity (All → Magic+ → Rare+ → Unique), cutting
visual noise on crowded drops.

- New `client/scripts/loot_label_filter.gd` (`class_name LootLabelFilter extends RefCounted`) owns
  the display policy: the rarity ladder, `allows(rarity)`, `cycle()`, `mode_label()`, and the
  reveal-dim `display_color()` (moved out of `main.gd`). Off-ladder loot (currency/quest/consumable)
  is never hidden.
- `main.gd` integration is minimal and net-negative: it holds a filter instance, gates non-hovered
  label visibility through `allows()` in `_refresh_loot_label_visibility()`, delegates dimming to
  `display_color()`, and cycles the filter on `]` (`KEY_BRACKETRIGHT`), logging the mode. The hovered
  label always shows regardless of filter, so you never lose the item under the cursor.
- Server, protocol, store, replay, and shared rules are untouched — the server still owns every loot
  roll and pickup.

## Proof

- `client/tests/test_loot_label_filter.gd` (registered in `scripts/client_smoke.sh`): ladder
  ordering, per-mode thresholds, cycle wraparound, off-ladder always-allowed, case-insensitivity, and
  `display_color` dimming. `make client-unit` PASS (incl. `test_delta_apply`, which preloads
  `main.gd`, confirming the integration loads).
- `make ci` green.

## Maintainability

- Touch-to-shrink honored: `main.gd` 6703 → 6699 (the dim logic + constant moved to the focused
  filter script); baseline lowered to 6699. New code landed in a focused file, not the coordinator.
- CODEMAP gained a `Loot presentation` row.

## Deferred

- Category (currency/quest/consumable) filtering; a persisted filter via `client_settings`; a
  dedicated always-on HUD indicator widget; a `make bot-visual` showcase scenario once a visible
  indicator lands.
