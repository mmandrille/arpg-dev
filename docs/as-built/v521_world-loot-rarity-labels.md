# v521 As-Built - World Loot Rarity Labels

## Shipped

Removed the floating geometric rarity meshes from ground loot. Revealed loot keeps its full
rarity/item name in the existing rarity text color, and ground/equipped item models retain their
existing rarity tint and lighting. Inventory/equipment slot rarity shapes remain intact. The
rarity-cue catalog and schema no longer carry unused world-marker geometry settings.

## Verification

| Command / evidence | Result |
|---|---|
| `godot --headless --path client --script res://tests/test_rarity_cues.gd` | PASS, 60 checks |
| `godot --headless --path client --script res://tests/test_loot_node_factory.gd` | PASS, 813 checks |
| `make client-unit` | PASS |
| `make validate-shared` | PASS, 2,267 checks and CODEMAP validation |
| `make validate-assets` | PASS, 464 checks |
| `make maintainability` | PASS, all ratchets |
| `make bot-visual scenario=inventory_lab_drop_item BOT_STEP_DELAY=0.05` | PASS, 1 scenario |
| `python3 skills/showme/scripts/render_focus.py --focus rarity-cues --ground-tone dark --reveal` | PASS; reviewed capture shows five tinted drops with colored names and no world markers |
| `make ci` | PASS in 7m43s; standalone final gate |
| `git diff --check` | PASS |

Renderer evidence: [five rarity drops](assets/v521/world-loot-rarity-labels.png), metadata at
[`world-loot-rarity-labels.json`](assets/v521/world-loot-rarity-labels.json).

## Limits

This verifies presentation and the existing pickup scenario; it makes no performance claim. The
capture reveals labels for inspection, while runtime reveal behavior is unchanged.
