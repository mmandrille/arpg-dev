# v410 Spec — Polearm Archetypes and Family Icons

Status: Complete
Date: 2026-07-02
Codename: polearm-archetypes-and-family-icons
Baseline: v409 `town-training-doll` complete

Related:

- [`v397_spec-item-archetype-library.md`](v397_spec-item-archetype-library.md)
- [`v394_spec-weapon-slot-families.md`](v394_spec-weapon-slot-families.md)
- [`../adr/0006-asset-pipeline.md`](../adr/0006-asset-pipeline.md)
- [`../adr/0014-core-progression-and-endgame-design-rules.md`](../adr/0014-core-progression-and-endgame-design-rules.md)

## Purpose

Extend the v397 archetype library with **reach-first melee weapons** and make inventory
icons **geometrically distinct per weapon `item_type`** (not only color/label overrides on
`blade`).

### New archetypes

| Template | Handedness | Reach | Attack speed | Identity |
|----------|------------|-------|--------------|----------|
| `spear` | 2H | 3.0 | ~0.80 | slow polearm, str-biased |
| `short_spear` | 1H | 2.2 | ~1.05 | fast skirmisher polearm |
| `halberd` | 2H | 2.5 | ~0.90 | mid reach polearm |
| `mace` | 1H | 1.5 | ~0.92 | blunt, str-biased |

### Presentation

- Add procedural icon shapes: `spear`, `halberd`, `mace`, `hammer` (1H), `axe`.
- Extend `item_type` enum with `spear`, `halberd`, `mace`, `hammer`.
- Re-type `hammer` → `hammer`, `morningstar` → `mace` (stats unchanged).
- Each weapon `item_type` maps to a presentation family with a unique `icon.shape`.

## Non-goals

- Production 3D models; borrow existing sword/axe GLBs.
- Crossbow, thrown javelins, piercing hits, new attack modes.
- Protocol schema bump; affix/name grammar changes; stash filters.
- CI pack promotion (extended scenario only).
- Treasure-class economy rebalance beyond adding new entries.

## Acceptance criteria

### Shared rules

- [ ] Four new templates with owner-requested spear reach/speed (`spear` 3.0 / ~0.80, `short_spear` 2.2 / ~1.05).
- [ ] `item_templates.v0.schema.json` `item_type` enum includes `spear`, `halberd`, `mace`, `hammer`.
- [ ] `hammer` and `morningstar` templates use `item_type` `hammer` and `mace`.
- [ ] New archetypes in depth-3+ treasure classes; `polearm_archetype_lab` world preset.
- [ ] `item_visuals.v0.json` and `item_presentations.v0.json` cover all new templates.
- [ ] `make validate-shared` passes.

### Combat

- [ ] Equipped `spear` resolves melee reach ≥ 2.9 from template data (Go test).
- [ ] Target at distance ~2.5 is in spear range but outside long-sword range (Go test).

### Client presentation

- [ ] `ItemIconDrawer` draws new shapes; schema allows them.
- [ ] Weapon families `axe`, `hammer`, `war_hammer`, `spear`, `halberd`, `mace` use distinct shapes (not `blade`/`greatsword` stand-ins).
- [ ] Headless unit test covers at least spear vs sword shape resolution.

### Bot

- [ ] Extended `polearm_archetype_lab`: common display names + spear-equipped hit on lab dummy at distance 2.0.

## Scope and likely files

- `shared/rules/item_templates.v0.json` + `.schema.json`
- `shared/rules/treasure_classes.v0.json`, `shared/rules/worlds.v0.json`
- `shared/assets/item_presentations.v0.json` + `.schema.json`
- `shared/assets/item_visuals.v0.json`
- `client/scripts/item_icon_drawer.gd`
- `client/tests/test_item_icon_drawer.gd`
- `server/internal/game/polearm_reach_test.go`
- `tools/bot/scenarios/119_polearm_archetype_lab.json`
- Docs: plan, as-built, lifecycle

## Test and bot proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run Polearm -count=1
godot --headless --path client --script res://tests/test_item_icon_drawer.gd
make bot scenario=polearm_archetype_lab
```

Visual: `make bot-visual scenario=polearm_archetype_lab`

## Asset decision

- **Adopt:** `ItemIconDrawer` procedural shapes, `item_presentations` family model.
- **Borrow:** `weapon_rusty_sword_v0` / `weapon_starter_axe_v0` GLBs for new archetypes.
- **Reject:** external art, texture pipeline.

## Open questions

None — defaults from `/next` brief apply.
