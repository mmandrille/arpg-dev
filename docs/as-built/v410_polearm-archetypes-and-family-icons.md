# v410 As-built — Polearm Archetypes and Family Icons

## What shipped

- Four new weapon archetypes: `spear` (2H, reach 3.0), `short_spear` (1H, reach 2.2), `halberd` (2H, reach 2.5), `mace` (1H blunt).
- `hammer` and `morningstar` now use `item_type` `hammer` / `mace` with distinct icon shapes.
- Procedural inventory icons: `spear`, `halberd`, `mace`, `hammer`, `axe` shapes in `ItemIconDrawer`; `war_hammer` uses hammer shape.
- Depth-3+ treasure classes include new templates; extended bot `polearm_archetype_lab`.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run Polearm -count=1
godot --headless --path client --script res://tests/test_item_icon_drawer.gd
make bot scenario=polearm_archetype_lab
```

Visual: `make bot-visual scenario=polearm_archetype_lab`

## Deferred

- Production polearm 3D models; crossbow/thrown javelins; CI pack promotion.
