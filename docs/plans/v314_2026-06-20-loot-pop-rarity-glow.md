# v314 Plan - Loot Pop Rarity Glow

Status: Complete
Goal: Add a small rarity glow and spawn-pop affordance to ground loot nodes.
Architecture: Keep the change inside the existing `LootNodeFactory` presentation helper. The server
continues to own loot state; the client only renders additional mesh children from current rarity data.
Tech stack: Godot 4 GDScript client helper, focused headless unit test, client smoke registration.

## Baseline and Shortcut Decision

Builds on v313. Asset/plugin decision: adopt existing code-native mesh/material patterns in
`LootNodeFactory`; reject external assets, shaders, and plugins.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/loot_node_factory.gd` | Add rarity glow and spawn-pop marker children |
| Create | `client/tests/test_loot_node_factory.gd` | Prove rarity glow, spawn marker, labels, and model path coexist |
| Modify | `scripts/client_smoke.sh` | Register focused unit test |
| Create | `docs/as-built/v314_loot-pop-rarity-glow.md` | Completion proof |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [ ] None

Decision:
- [x] Keep changes in the existing focused helper; no large coordinator touched.

Verification:
```bash
make maintainability
```

## Tasks

- [x] Add code-native rarity glow and spawn marker children to ground loot nodes.
- [x] Add focused Godot unit coverage for common and rare/unique loot visuals.
- [x] Register the focused unit test in client smoke.
- [x] Update lifecycle docs and as-built proof.

## Verification

- [x] `godot --headless --path client --script res://tests/test_loot_node_factory.gd`
- [x] `make client-unit`
- [x] `make maintainability`

Final full `make ci` remains deferred to the enclosing `$autoloop` batch gate.
