# v490 Plan — Kit coins and potions

Status: Implemented
Spec: [`docs/specs/v490_spec-kit-coins-potions.md`](../specs/v490_spec-kit-coins-potions.md)

## Baseline and shortcut decision
Reuses v487 ground model resolution + `ground_pose`, v483 ArmorLook detail-layer technique
(extracted to `ModelDetailTint`), v473/v489 vendoring and validators.

## Maintenance ratchet
No grandfathered file touched. `loot_node_factory.gd` (444) must stay ≤ 600.

## Task 1 — Assets + data + validation
- [x] Vendor 7 GLBs (+ import); manifest entries (CC0 provenance, sha256).
- [x] `item_presentations`: potion family `3d_model` + `ground_tint`; gold `ground_model_tiers`; schema.
- [x] `ground_pose.assets` overrides (upright, rest_on_floor, max_extent).
- [x] Validator: tier/asset ids resolve; tiers strictly ascending. Tests.
```bash
make validate-shared && make validate-assets && .venv/bin/pytest -q tools/test_validate_item_presentations.py
```
## Task 2 — Client
- [x] `ModelDetailTint` extraction; ArmorLook uses it.
- [x] Factory: amount-based gold tier; `ground_tint` detail path.
- [x] Tests (catalog-derived) in `test_loot_node_factory.gd`; `test_armor_look.gd` green.
```bash
make client-unit
```
## Task 3 — Visual gate
- [x] Captures of gold tiers + potions before/after → `docs/as-built/assets/v490/`.
## Task 4 — Regression + docs
- [x] `make bot-client SCENARIO=click_to_kill HEADLESS=1`; docs; PROGRESS.

## Execution notes

- The gold/potion capture is a throwaway scratch script (not committed): the `floor-item` focus hardcodes `amount: 1`, and `showme/visual_capture.gd` is already over its ratchet baseline.
- Mana tint strength raised 0.65 → 0.8 after the first capture (washed out on the brown base).
