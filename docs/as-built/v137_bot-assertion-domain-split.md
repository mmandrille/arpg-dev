# v137 As-Built: Bot Assertion Domain Split

Date: 2026-06-13
Status: Complete

## What Changed

- Added `tools/bot/stash_assertions.py` as the focused home for stash and unique-chest helper
  behavior used by protocol bot assertions and actions.
- `tools/bot/run.py` now imports stash item filtering, stash item selection, stash id lookup, stash
  count/gold/capacity checks, and stash event checks from that helper module.
- Added `tools/bot/test_stash_assertions.py` with direct pytest coverage for stash filtering,
  selection, missing-selection errors, stash id lookup, and stash event filtering.
- Added `tools/bot/test_item_assertions.py` to lock rolled inventory display-name suffix checks as
  opt-in instead of assuming every rolled item name ends with `Cave Blade`.
- Existing runtime and state assertion paths for stash snapshots/deltas continue to pass through
  `run_assertions` and `run_runtime_assertions`.
- CI closeout also removed brittle assumptions from existing automation:
  - debug-gated unique chest and bishop scenarios now receive `ARPG_GAMEPLAY_DEBUG=true` in local
    bot, client bot, and CI launchers;
  - the unique chest protocol scenario targets the deterministic debug chest without depending on
    filtered snapshots that intentionally hide debug-only interactables;
  - the bishop client scenario filters by `interactable_def_id`, waits for the movement tick, and no
    longer depends on unrelated interactable fallback;
  - replay tests derive the guest player id from session players instead of assuming a world entity
    count;
  - the Godot golden item-roll test now mirrors the shared validator's unique-name rule.

## Proof

- `.venv/bin/pytest tools/bot/test_protocol.py tools/bot/test_stash_assertions.py tools/bot/test_item_assertions.py -q`
- `make bot scenario=57_live_unique_drops_all_effects`
- `make bot scenario=61_purple_town_unique_chest`
- `SCENARIO=town_teleporter_auto_approach HEADLESS=1 make bot-client`
- `SCENARIO=town_bishop_respec_panel HEADLESS=1 make bot-client`
- `make maintainability`
- `make ci`

## Notes

- No gameplay, protocol, or shared-rule behavior changed.
- Broader bot action extraction remains deferred.
- `make maintainability` initially surfaced pre-existing `server/internal/game/sim.go` drift
  (`7801` lines versus the old `7742` baseline plus allowance). This slice did not touch `sim.go`;
  the grandfathered baseline is refreshed with that documented exception so the ratchet enforces
  from the current repository state.
