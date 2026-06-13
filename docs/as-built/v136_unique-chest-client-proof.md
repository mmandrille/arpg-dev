# v136 As-Built: Unique Chest Client Proof

Date: 2026-06-13
Status: Complete

## What Changed

- Added client bot scenario `unique_chest_client_proof`.
- The scenario opens `town_unique_chest`, waits for the unique chest panel, and asserts the panel is
  in `unique_chest` mode.
- Stash row debug summaries now include the same unique-effect text resolved for tooltip display.
- Client bot stash row matching can filter by `display_name`, `summary_contains`, and
  `container_mode`.
- The scenario proves `Embercall Blade` exposes the Everburning Wound summary and `Stormstring Bow`
  exposes the Stormbound Echo summary.

## Proof

- `make client-unit`
- `SCENARIO=unique_chest_client_proof HEADLESS=1 make bot-client`
- `make maintainability`
- `make ci`

## Notes

- The visual verification command for the purple chest remains
  `make bot-visual scenario=61_purple_town_unique_chest.json`.
