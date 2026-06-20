# v314 As Built - Loot Pop Rarity Glow

Date: 2026-06-20
Spec: [`docs/specs/v314_spec-loot-pop-rarity-glow.md`](../specs/v314_spec-loot-pop-rarity-glow.md)
Plan: [`docs/plans/v314_2026-06-20-loot-pop-rarity-glow.md`](../plans/v314_2026-06-20-loot-pop-rarity-glow.md)

## What Shipped

- Ground loot nodes now add a code-native `RarityGlow` ring from the existing rarity color map.
- Ground loot nodes now add a `SpawnPopRing` marker so newly visible loot has an assertable pop
  affordance.
- Existing primitive/model ground loot, labels, filters, pickup authority, and loot rules remain
  unchanged.
- Added and registered `test_loot_node_factory.gd` so the glow and marker gate in `make client-unit`.

## Proof

```bash
godot --headless --path client --script res://tests/test_loot_node_factory.gd
make client-unit
make maintainability
```

Result: green on 2026-06-20. Full `make ci` is deferred to the enclosing `$autoloop` batch gate.

## Manual Visual Command

```bash
make bot-visual scenario=01_click_to_kill
```

## Deferred

- Animated particle systems, imported loot art, audio cues, loot filter changes, and server-side
  loot behavior remain deferred.
