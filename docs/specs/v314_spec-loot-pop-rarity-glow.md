# v314 Spec - Loot Pop Rarity Glow

Status: Complete
Date: 2026-06-20
Codename: loot-pop-rarity-glow

## Purpose

Make newly visible ground loot feel more rewarding by adding a code-native rarity glow and spawn-pop
presentation to existing loot nodes while preserving server-owned loot authority.

## Non-goals

- No loot table, rarity, pickup, ownership, protocol, or server changes.
- No external art packs, shaders, plugins, or imported assets.
- No loot filter behavior changes.

## Acceptance Criteria

- Ground loot nodes include a rarity-colored glow/ring that is present for all rarities.
- Ground loot nodes include a small spawn-pop marker that can be asserted headlessly.
- Existing labels, filters, item presentation metadata, and 3D ground equipment models still render.
- The change is covered by a focused Godot unit test registered in `make client-unit`.

## Scope and Files

- Modify `client/scripts/loot_node_factory.gd`.
- Create `client/tests/test_loot_node_factory.gd`.
- Register the test in `scripts/client_smoke.sh`.
- Add lifecycle/as-built docs when the slice ships.

## Test and Bot Proof

```bash
godot --headless --path client --script res://tests/test_loot_node_factory.gd
make client-unit
make maintainability
```

Manual visual command:

```bash
make bot-visual scenario=01_click_to_kill
```

## Open Questions and Risks

- None. Asset/plugin decision: adopt existing code-native mesh/material loot presentation; reject
  external assets/plugins for this small readability pass.
