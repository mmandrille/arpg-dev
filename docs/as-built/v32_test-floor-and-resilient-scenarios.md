# v32 — Test floor and resilient scenarios

**Proves:** CI distinguishes intentional contract locks from mutable tuning details, so normal
dungeon, population, movement, loot-weight, and presentation tuning can proceed without weakening
replay, schema, formula, persistence, or protocol coverage.

- `CLAUDE.md` now documents the durable Test Locking Policy for future slices.
- The v32 plan audit record classifies exact assertions as contract locks, behavior proofs, or
  tuning details before changing tests.
- Python bot assertions now support semantic entity filters, range comparators, inventory filters,
  eventual assertions, and generated dungeon walk budgets derived from map size.
- Protocol scenarios now prove chase/leash/dungeon behavior through eventual or semantic assertions
  instead of fixed tick waits and incidental total population counts.
- Character leveling keeps formula, level, max-HP, stat allocation, event, replay, reconnect, and
  persistence locks while avoiding an exact generated-XP tuning total.
- Go, shared validation, and GDScript golden tests keep schema and formula contracts exact while
  deriving or structurally validating generated population, rarity tuning, and loot-depth offsets.
- Client bot scenarios can target entities by debug metadata such as monster definition,
  interactable definition, item definition, rarity, and state instead of fragile entity indexes.
- Local reverted probes changed dungeon floor size, generated monster population, and movement speed;
  the focused checks stayed green with no committed tuning changes.

**Explicit non-goals:** no gameplay, balance, protocol, or UI feature work; no committed tuning
changes; no broad test framework migration; no `tuning_sensitive` metadata.
