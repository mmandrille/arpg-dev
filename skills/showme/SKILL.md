---
name: showme
description: Capture focused Godot client visuals for fast feedback while tuning presentation. Use when the user asks to show, render, preview, screenshot, or open the client for a specific visual under improvement, such as equipment models, paper-doll inventory UI, item icons, character facing, armor/helmet/boots/shield placement, or another isolated Godot client presentation detail.
---

# Show Me

Use this skill to produce quick visual proof for the exact client element being tuned, without running the full server game loop unless the request needs it.

## Workflow

1. Prefer a focused screenshot first. It is fast, deterministic, and easy to attach back to the user.
2. Use live mode only when the user asks to see the client window, needs rotation/interaction, or a screenshot is not enough.
3. Keep the capture scoped to the thing under review. Do not start `make play` unless the user specifically needs a full gameplay path.
4. After changing visuals, rerun the focused capture and inspect the image before asking for feedback.

## Script

Run from the repo root:

```bash
python3 skills/showme/scripts/render_focus.py --focus <focus>
```

The script prints the screenshot path under `.artifacts/showme/`.

### Modes and flags

| Flag | Default | Purpose |
|------|---------|---------|
| `--focus` | `gear` | What to render (see catalog below). |
| `--mode` | `screenshot` | `screenshot` saves PNG; `live` opens a Godot window. |
| `--duration` | `-1` (screenshot) / `45` (live) | Live timeout in seconds; `0` keeps the window open until closed. |
| `--items` | gear default set | Comma-separated `item_def_id`s. Used by `gear` and `floor-item`. |
| `--class-id` | (none) | Class model for `gear`, e.g. `paladin`, `barbarian`, `sorcerer`. |
| `--width` / `--height` | `640×480` | Override window size; many focuses auto-bump when left at default. |
| `--output` | timestamped under `.artifacts/showme/` | PNG path for screenshot mode. |
| `--godot` | `godot` | Godot binary when not on PATH. |

```bash
# Screenshot (default).
python3 skills/showme/scripts/render_focus.py --focus gear

# Live preview for 45 seconds (gear auto-rotates).
python3 skills/showme/scripts/render_focus.py --focus gear --mode live --duration 45

# Live preview until the window is closed.
python3 skills/showme/scripts/render_focus.py --focus gear --mode live --duration 0
```

Screenshot and live mode both use Godot's render-capable window path because the macOS headless/dummy renderer cannot produce viewport pixels. If sandbox GUI restrictions block either mode, rerun the command with escalation and a short approval prompt.

## Focus catalog

Canonical list lives in `skills/showme/scripts/render_focus.py` (`--focus` choices). When adding a new focus, register it there and in this table.

### Quick picker

| User wants to see… | `--focus` |
|--------------------|-----------|
| Equipped character / socket placement | `gear` |
| All three class models side by side | `classes` |
| Dropped loot on the ground | `floor-item` |
| Paper-doll inventory + tooltip | `inventory` |
| Hero corpse interactable (3D) | `corpse` |
| Player inventory + corpse loot panels | `corpse-inventory` |
| Skills panel + hover state | `skills` |
| Every item icon family (grid) | `item-icons` |
| Vendor buy/sell UI | `shop` |
| Bishop heal/resurrect panel | `bishop` |
| Market board interactable (3D labels) | `market-board` |
| Market publish tab | `market-publish` |
| Market offer/bid tab | `market-offer` |
| Character select / create screen | `character-menu` |
| Multiplayer session browser | `join-menu` |
| Player HP/mana HUD bar | `hud` |
| Up + down stair meshes | `stairs` |
| Town stash + treasure chest | `chests` |
| Town vendor + mystery seller | `vendors` |
| Monster lineup (dummy, wolf, bat, skeleton, boss) | `monsters` |
| Companions + revive corpse UI | `companions` |
| Heal-rain VFX over targets | `heal-rain` |
| Full town layout overview | `town` |

### Equipment and models

| Focus | Shows | Extra flags | Default size* |
|-------|-------|-------------|---------------|
| `gear` | Isolated `character.tscn` with equipment from `EquipmentResolver`. Default items: `long_sword,shield,helm,mail,boots`. | `--items`, `--class-id` | 640×480 |
| `classes` | Barbarian, sorcerer, paladin models in a row with labels. | — | 1120×640 |
| `floor-item` | Single loot node on grass (`LootNodeFactory`). First `--items` entry or `long_sword`. | `--items` | 640×480 |

### Inventory and icons

| Focus | Shows | Extra flags | Default size* |
|-------|-------|-------------|---------------|
| `inventory` | `InventoryPanel` with sample items, equipped slots, and item tooltip. | — | 960×640 |
| `corpse` | Hero corpse interactable on dungeon floor. | — | 640×480 |
| `corpse-inventory` | Side-by-side player inventory + corpse stash panel. | — | 1120×640 |
| `skills` | Full `SkillsPanel` with sample progression; `rage` hovered for disabled-tooltip state. | — | 960×640 |
| `item-icons` | Grid of all `item_presentations.v0.json` families via `ItemIconsCatalog`. | — | 960×720 |

### Town / economy UI

| Focus | Shows | Extra flags | Default size* |
|-------|-------|-------------|---------------|
| `shop` | `ShopPanel` + inventory sell context, sample offers, offer tooltip. | — | 1280×760 |
| `bishop` | `BishopPanel` heal/resurrect UI (debug enabled). | — | 640×520 |
| `market-board` | 3D market board with incoming/published count labels. | — | 960×640 |
| `market-publish` | `MarketPanel` on publish tab with sample listings/stash. | — | 1120×720 |
| `market-offer` | `MarketPanel` on offer tab. | — | 1120×720 |

### Menus and HUD

| Focus | Shows | Extra flags | Default size* |
|-------|-------|-------------|---------------|
| `character-menu` | `CharacterSelectPanel` choose/create with live + dead characters. | — | 960×640 |
| `join-menu` | `MultiplayerSessionsPanel` with sample coop sessions. | — | 960×640 |
| `hud` | `PlayerHealthBar` with sample name, level, HP, mana. | — | 640×480 |

### World interactables and VFX

| Focus | Shows | Extra flags | Default size* |
|-------|-------|-------------|---------------|
| `stairs` | `stairs_up` and `stairs_down` meshes side by side. | — | 960×640 |
| `chests` | Town stash + treasure chest interactables. | — | 960×640 |
| `vendors` | Town vendor + mystery seller interactables. | — | 960×640 |
| `monsters` | Dummy, wolf, bat, skeleton, unique boss lineup. | — | 1120×640 |
| `companions` | Live `main.gd` slice: wolf/undead/mercenary companions + dead wolf + revive corpse bar. | — | 960×640 |
| `heal-rain` | `HealRainEffect` over five target markers. | — | 960×640 |
| `town` | `make_town_preview_scene()` isometric overview. | — | 1120×720 |

\*Default size applies when `--width 640 --height 480` is left unchanged; `render_focus.py` bumps dimensions per focus.

## Common examples

```bash
# Full default equipment set.
python3 skills/showme/scripts/render_focus.py --focus gear

# Specific items and class body.
python3 skills/showme/scripts/render_focus.py --focus gear --items helm,mail,boots --class-id paladin

# One floor loot drop.
python3 skills/showme/scripts/render_focus.py --focus floor-item --items spear

# Paper-doll inventory.
python3 skills/showme/scripts/render_focus.py --focus inventory

# All item icon families.
python3 skills/showme/scripts/render_focus.py --focus item-icons

# Skills panel (hover on locked rage).
python3 skills/showme/scripts/render_focus.py --focus skills

# Vendor shop with tooltip.
python3 skills/showme/scripts/render_focus.py --focus shop

# Monster silhouettes.
python3 skills/showme/scripts/render_focus.py --focus monsters

# Town overview.
python3 skills/showme/scripts/render_focus.py --focus town
```

## Notes

- The renderer mirrors client presentation code and shared visual metadata; it must not mutate server rules or gameplay authority.
- Implementation: `skills/showme/scripts/render_focus.py` (CLI) → `skills/showme/scripts/visual_capture.gd` (`_setup_*` per focus).
- For visual-equipment changes, usually run `godot --headless --path client --script res://tests/test_item_visuals.gd` after a capture.
- For UI layout changes, usually run `make client-unit` after the focused iteration stabilizes.
- Add new focus values only when a repeated feedback loop appears; update both `render_focus.py` choices and this catalog.
