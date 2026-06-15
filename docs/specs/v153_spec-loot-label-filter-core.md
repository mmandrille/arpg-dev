# v153 Spec: Loot Label Filter Core

Status: Draft
Date: 2026-06-14
Codename: `loot-label-filter-core`

## Purpose

Resume player-facing feature work (after the v141–v152 maintenance arc) with a small, display-only
loot quality-of-life feature: a client-side rarity threshold filter for ground loot labels. The
client already reveals ground loot labels on hover and while a reveal key (Alt) is held
(`main.gd:_refresh_loot_label_visibility`), colored by rarity. This slice lets the player cycle a
filter so the held-reveal only shows labels at or above a chosen rarity (All → Magic+ → Rare+ →
Unique), cutting visual noise on crowded drops — the standard ARPG "loot filter" affordance.

It is deliberately **presentation-only**: the server still owns every loot roll, item, and pickup.
The filter changes only which already-revealed labels are drawn on this client.

## Non-goals

- No server, protocol, store, replay, or shared-rules change. The filter is pure client state.
- No change to loot rolls, item ownership, pickup, or which items exist on the ground.
- No hiding of the **hovered** label — the item under the cursor always shows its label regardless
  of filter, so the player never loses the item they are pointing at.
- No category filter (currency/quest/consumable) in this slice; the threshold acts on the rarity
  ladder only, and non-rarity loot (e.g. currency) is always shown. Category filtering is deferred.
- No persistence of the chosen filter across sessions (client runtime state only); persisting it via
  `client_settings` is a deferred polish.
- No new dedicated HUD widget/scene surgery; the current mode is surfaced through the existing
  `status_text` overlay readout, not a new panel.

## Acceptance Criteria

- A focused new script `client/scripts/loot_label_filter.gd` (`class_name LootLabelFilter extends
  RefCounted`) owns the filter state: an ordered rarity ladder (`common`, `magic`, `rare`,
  `unique`), a current threshold, `allows(rarity: String) -> bool`, `cycle() -> void`, and
  `mode_label() -> String`. `allows` returns true for any rarity at or above the threshold and true
  for rarities not on the ladder (so currency/quest loot is never hidden).
- `main.gd` holds a `LootLabelFilter` instance and, in `_refresh_loot_label_visibility()`, a
  non-hovered label is shown only when `loot_label_reveal_held and filter.allows(rarity)`; the
  hovered/highlighted label always shows. No change to label colors or the hover path.
- A keybind (held-reveal context) cycles the filter and immediately refreshes label visibility; the
  new mode is reflected in the `status_text` readout.
- `main.gd` does not grow past its grandfathered baseline (touch-to-shrink): the filter logic lives
  in the new script; `main.gd` only gains the instance, the one visibility gate, and the keybind.
- Headless unit test `client/tests/test_loot_label_filter.gd` proves: ladder ordering, `allows`
  thresholds at each mode, `cycle` wraparound, and that off-ladder rarities are always allowed.
- `make client-unit` and `make ci` pass.

## Scope and files likely touched

- `client/scripts/loot_label_filter.gd` (new) — filter model.
- `client/scripts/main.gd` — instance, visibility gate in `_refresh_loot_label_visibility`, keybind,
  `status_text` readout line.
- `client/tests/test_loot_label_filter.gd` (new) — headless unit coverage.
- `docs/CODEMAP.md` — add the new script + test (and a Loot/itemization presentation row if absent).
- `PROGRESS.md`, `docs/as-built/v153_loot-label-filter-core.md`.
## Test and bot proof

- `client/tests/test_loot_label_filter.gd` headless unit test (pure filter logic) via `make client-unit`.
- `make client-unit` confirms `main.gd` still parses/loads with the integration.
- `make ci` (client smoke + existing client bot scenarios remain the integration regression).
- No new protocol bot scenario: this is display-only client presentation with no gameplay, protocol,
  world, inventory-ownership, or movement change (per AGENTS.md, bot scenarios are required only when
  those surfaces change).

## Open questions and risks

- Keybind choice must not collide with existing client bindings (WASD move, Shift stationary-attack,
  Alt reveal, camera zoom). Plan picks a free key (candidate: `KEY_BRACKETLEFT`/`KEY_BRACKETRIGHT` or
  `KEY_F`) and verifies no collision in `_unhandled_input`.
- `status_text` is an existing overlay that may be toggled off; the filter still works without it
  (the live label changes are the primary feedback). A dedicated always-visible indicator is deferred.
- Rarity ladder is a client display constant mirroring `LOOT_LABEL_RARITY_COLORS` key order; it is UX
  ordering, not gameplay tuning, so code ownership in the client is appropriate.
