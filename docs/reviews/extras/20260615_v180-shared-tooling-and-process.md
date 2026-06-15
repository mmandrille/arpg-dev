# Shared tooling and process review v180

Date: 2026-06-15
Baseline: v180 — upgrade resource drop

## What improved

- Shared validation became less monolithic: `tools/validate_item_presentations.py:8` now owns item-presentation cross-checks, and `tools/validate_shared.py:3027` delegates to it.
- Scenario coverage grew to 119 protocol JSON scenarios and 45 client JSON scenarios. v179 and v180 added `tools/bot/scenarios/71_mana_regeneration.json` and `tools/bot/scenarios/72_upgrade_resource_drop.json`.
- The official review gate passed after catching and fixing a stale test whitelist. This is exactly the kind of batch-level regression the cadence is supposed to catch.
- The large-file ratchet reports 33 grandfathered files and 64,672 grandfathered lines, down from the previous review baseline after the pre-review refactor work.

## Risks

1. `tools/bot/run.py` is 4,294 lines while the baseline entry is 4,269 lines. It is still within the allowed ratchet window, but the next protocol bot capability should extract a domain helper first.

2. Coupled helper injections remain at 37 occurrences. The typed context cleanup identified in the prior review is still useful, but it is a larger refactor than the minor pre-review pass should take on.

3. `tools/validate_shared.py` is still 3,115 lines after the item-presentation extraction. More validators can move out by domain when nearby rules change.

4. The full CI gate is now about 6m25s locally. That is acceptable, but future scenario additions should keep using focused targeted tests during iteration and reserve `make ci` for final proof.

## Recommended tooling/process work

- Continue splitting `tools/validate_shared.py` by content domain when a slice touches that domain.
- Extract protocol bot helpers before adding new assertion/action families to `tools/bot/run.py`.
- Keep review cadence at v190 and run `$refactor` before `$review` again if the next feature batch adds pressure to large files.

## Verification evidence

- `make ci` on 2026-06-15:
  - shared schema validation
  - asset manifest and GLB validation
  - determinism lint
  - Go tests
  - Python unit checks
  - protocol bot plus replay
  - 45 Godot client bot scenarios
  - Godot headless smoke
