---
name: showme
description: Capture focused Godot client visuals for fast feedback while tuning presentation. Use when the user asks to show, render, preview, screenshot, or open the client for a specific visual under improvement, such as equipment models, paper-doll inventory UI, item icons, character facing, armor/helmet/boots/shield placement, or another isolated Godot client presentation detail. After broad visual changes, use `make regen-screenshots` to batch-capture every class, skill icon, item icon, and item model for regression review.
---

# Show Me

Use this skill to produce quick visual proof for the exact client element being tuned, without running the full server game loop unless the request needs it.

## Workflow

1. Prefer a focused screenshot first. It is fast, deterministic, and easy to attach back to the user.
2. After broad presentation changes, run `make regen-screenshots` (or a `SUITE=` subset) and inspect `.artifacts/screenshots/latest/` for regressions before asking for feedback.
3. Use live mode only when the user asks to see the client window, needs rotation/interaction, or a screenshot is not enough.
4. Keep the capture scoped to the thing under review. Do not start `make play` unless the user specifically needs a full gameplay path.
5. After changing visuals, rerun the focused capture and inspect the image before asking for feedback.

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
| `--refresh` | `0` (off) | **Gear + live only:** reload shared JSON configs every N seconds (see below). |
| `--rotation-period` | `0` (auto) | Seconds for one 360° rotation in live mode; defaults to `--refresh` when set. |
| `--items` | gear default set | Comma-separated `item_def_id`s. Used by `gear` and `floor-item`. |
| `--class-id` | (none) | Class model for `gear` / `skeleton`, e.g. `paladin`. |
| `--skill-id` | (none) | Skill id for `skill-icon` focus. |
| `--family-id` | (none) | Item presentation family for `item-icon` focus. |
| `--asset-id` | (none) | Asset manifest id for `item-asset` focus. |
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

# Live gear tuning loop: auto-reload configs every 3s, one full rotation per cycle.
python3 skills/showme/scripts/render_focus.py --focus gear --class-id paladin --items barbarian_axe --mode live --refresh 3 --duration 0
```

Screenshot and live mode both use Godot's render-capable window path because the macOS headless/dummy renderer cannot produce viewport pixels. If sandbox GUI restrictions block either mode, rerun the command with escalation and a short approval prompt.

### Live gear tuning loop (`--refresh`)

Use this when iterating on equipped-gear placement without restarting Godot. **Requires `--mode live` and `--focus gear`.**

Every `--refresh N` seconds the preview:

1. Invalidates and reloads shared JSON caches (`item_visuals.v0.json`, `assets/manifests/assets.v0.json`, `gear_sockets.v0.json`, `equipment_display.v0.json`, plus item/class presentation loaders used by the gear resolver).
2. Rebuilds bone sockets on the character and remounts the `--items` set.
3. Resets the turntable to its starting angle.

Rotation speed is tied to the refresh interval by default: one full 360° per cycle (`--rotation-period` overrides). Override with `--rotation-period` if you want a different spin rate.

**Hot-reloads on save:** socket offsets, item mount transforms, equipment display multipliers, and manifest `runtime_path` entries.

**Still needs a restart:** class body `.glb` scene swaps (reload class mesh from disk) and regenerated runtime `.glb` binaries until Godot's resource cache is cleared or the process exits.

Typical workflow while editing `shared/assets/item_visuals.v0.json` or `shared/assets/gear_sockets.v0.json`:

```bash
python3 skills/showme/scripts/render_focus.py \
  --focus gear --class-id paladin --items barbarian_axe,shield,helm \
  --mode live --refresh 3 --duration 0
```

Save the JSON file; within one refresh cycle the window updates. Close the Godot window (or Ctrl+C the terminal) when done.

Implementation: `skills/showme/scripts/render_focus.py` passes `--refresh` / `--rotation-period` to `client/scripts/showme/visual_capture.gd`, which calls `EquipmentVisualResolver.reload_data_only()` and `refresh_gear_sockets()` on each cycle.

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
| Single skill icon | `skill-icon` |
| Single item icon family | `item-icon` |
| Isolated item 3D GLB | `item-asset` |
| Vendor buy/sell UI | `shop` |
| Bishop heal/resurrect panel | `bishop` |
| Skeleton bone positions + socket markers | `skeleton` |
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
| `skeleton` | Character in spread T-pose; red dot at every bone + colored labeled sphere at each equipment socket. | `--class-id` | 800×600 |
| `floor-item` | Single loot node on grass (`LootNodeFactory`). First `--items` entry or `long_sword`. | `--items` | 640×480 |

### Inventory and icons

| Focus | Shows | Extra flags | Default size* |
|-------|-------|-------------|---------------|
| `inventory` | `InventoryPanel` with sample items, equipped slots, and item tooltip. | — | 960×640 |
| `corpse` | Hero corpse interactable on dungeon floor. | — | 640×480 |
| `corpse-inventory` | Side-by-side player inventory + corpse stash panel. | — | 1120×640 |
| `skills` | Full `SkillsPanel` with sample progression; `rage` hovered for disabled-tooltip state. | — | 960×640 |
| `item-icons` | Grid of all `item_presentations.v0.json` families via `ItemIconsCatalog`. | — | 960×720 |
| `skill-icon` | One `SkillIcon` from `skill_presentations.v0.json`. | `--skill-id` | 480×480 |
| `item-icon` | One `ItemFamilyIconPreview` from presentation families. | `--family-id` | 480×480 |
| `item-asset` | Isolated weapon/equipment GLB from `assets/manifests/assets.v0.json`. | `--asset-id` | 640×480 |

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

# Live tuning loop while editing item_visuals / gear_sockets JSON.
python3 skills/showme/scripts/render_focus.py --focus gear --class-id paladin --items barbarian_axe --mode live --refresh 3 --duration 0

# Skeleton bone visualization (default class: paladin).
python3 skills/showme/scripts/render_focus.py --focus skeleton --class-id paladin

# Check another class skeleton.
python3 skills/showme/scripts/render_focus.py --focus skeleton --class-id barbarian

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

## Batch visual regression (`make regen-screenshots`)

Data-driven batch capture: loops shared catalogs and writes one PNG per element. No server
required. Use after presentation slices to catch socket drift, missing icons, broken models, or
class-specific gear fit issues.

```bash
make regen-screenshots-list
make regen-screenshots                              # all suites (~200+ captures)
make regen-screenshots SUITE="skeleton gear"        # class bodies + equipped gear
make regen-screenshots SUITE="skill-icon"           # every skill icon
make regen-screenshots SUITE="item-icon item-icons" # per-family + grid
make regen-screenshots SUITE="floor-item item-asset" # loot + raw 3D assets
make regen-screenshots DRY_RUN=1                    # print planned jobs only
make regen-screenshots OUT=.artifacts/screenshots/my-run SUITE=gear
```

| Suite | Focus | Source data | What to check |
|-------|-------|-------------|---------------|
| `skeleton` | `skeleton` | `character_progression.v0.json` classes | Bone spread, socket markers per class |
| `gear` | `gear` | same classes | Default equipped set fit per class body |
| `skill-icon` | `skill-icon` | `skills.v0.json` | Every skill icon shape/label/color |
| `item-icon` | `item-icon` | `item_presentations.v0.json` families | Every item icon family |
| `item-icons` | `item-icons` | same families | Full icon grid overview |
| `floor-item` | `floor-item` | `item_visuals.v0.json` | Ground loot model per item |
| `item-asset` | `item-asset` | unique `asset_id`s in item visuals | Isolated 3D GLB load |

Output directory: `.artifacts/screenshots/<timestamp>/` with `index.json` manifest.
Symlink `latest` points at the most recent run.

Implementation: `tools/showme/screenshot_catalog.py` (job discovery) →
`tools/showme/regen_screenshots.py` (orchestrator) → `render_focus.py` per job.

**Agent checklist after visual work:**

1. Run the smallest relevant `SUITE=` (not the full matrix unless the change is broad).
2. Open PNGs under `.artifacts/screenshots/latest/<suite>/`.
3. Note any missing renders, wrong sockets, or icon/model errors in the slice summary.
4. For a single element still in flux, use `render_focus.py` instead of the full batch.

## Notes

- The renderer mirrors client presentation code and shared visual metadata; it must not mutate server rules or gameplay authority.
- Implementation: `skills/showme/scripts/render_focus.py` (CLI) → `client/scripts/showme/visual_capture.gd` (`_setup_*` per focus). Batch path: `make regen-screenshots` → `tools/showme/regen_screenshots.py`.
- For visual-equipment changes, usually run `godot --headless --path client --script res://tests/test_item_visuals.gd` after a capture.
- For UI layout changes, usually run `make client-unit` after the focused iteration stabilizes.
- Add new focus values only when a repeated feedback loop appears; update both `render_focus.py` choices and this catalog.
