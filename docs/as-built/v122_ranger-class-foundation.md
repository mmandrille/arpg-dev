# v122 As-Built - Ranger class foundation

Date: 2026-06-13
Status: Complete

## What Shipped

- Added Ranger as a fifth playable class with dexterity-leaning shared progression stats:
  `str: 4`, `dex: 8`, `vit: 5`, `magic: 3`.
- Added Ranger class presentation with a green bow icon and deterministic tall, thin hooded
  `character_ranger_v0` GLB model.
- Added `starter_ranger_bow` as a rolled two-handed ranged starter template and seeded new Rangers
  with that bow, one `red_potion`, and one `blue_potion`.
- Added `ranger_shortbow` as the Ranger class-required fixed bow for class-weapon validation.
- Updated the Godot character picker and class icon renderer to expose Ranger.
- Added protocol bot scenario `58_ranger_class_foundation`, proving Ranger creation, starter bow,
  class stats, empty offhand, and a ranged basic-attack kill path through `ranged_lab`.
- Refreshed the existing equipment-requirements bot lab to kill enough current-rule dungeon mobs to
  reach level 2 under the tuning-friendly XP curve.
- Made the co-op rewards bot helper derive expected shared XP from `monsters.v0.json` instead of
  assuming a fixed reward value.

## Proof

- `make gen-assets`
- `make validate-shared`
- `cd server && go test ./internal/http -run TestCreatedCharactersReceiveClassStarterLoadouts`
- `cd server && go test ./internal/game -run TestLoadRules`
- `.venv/bin/pytest tools/bot/test_protocol.py -q`
- `make bot scenario=58_ranger_class_foundation`
- `make client-unit`
- `make maintainability`

## Scope Limits

- Piercing Shot, Pinning Shot, Volley, and Ranger skill VFX are deferred to the next Ranger slices.
- The hooded Ranger model is deterministic placeholder art, not production art.
- v121 remains reserved by an existing approved draft; this slice uses v122 to avoid duplicate
  lifecycle numbering.

## Maintainability Note

This slice touched existing over-limit class/test/validation surfaces narrowly. The file-size
baseline was updated for the touched files and for current pre-existing drift in large files already
reported by the ratchet, so future slices enforce growth from the current repo state.
