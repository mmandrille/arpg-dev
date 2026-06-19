# v282 As-Built - Rectangular Boss Pattern

Date: 2026-06-19
Spec: [`docs/specs/v282_spec-rectangular-boss-pattern.md`](../specs/v282_spec-rectangular-boss-pattern.md)
Plan: [`docs/plans/v282_2026-06-19-rectangular-boss-pattern.md`](../plans/v282_2026-06-19-rectangular-boss-pattern.md)

## Shipped

- Added `crystal_wall`, a rectangle-shaped Cave Warden boss pattern with telegraph, active, and
  recovery phases.
- Added `crystal_wall` to the Cave Warden pattern deck after the already bot-proven stone lance,
  summon wolves, and shard fan entries.
- Extended boss pattern validation to require positive width for rectangle telegraphs and to keep
  active damage predicates matched to their telegraphs.
- Added server rectangle hit detection using the boss's locked aim plus range/width checks.
- Updated the boss deck cycle test to derive expected order from `pattern_deck` data.
- Added focused rectangle-pattern tests for data, validation, and hit predicate behavior.
- Updated client boss telegraph rendering so rectangle hit shapes produce rectangular `BoxMesh`
  decals, with factory/unit coverage.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game/...
GODOT=/opt/homebrew/bin/godot make client-unit
```

All focused checks above passed on 2026-06-19.

`make maintainability` remains blocked by pre-existing unrelated ratchet debt identified during v278;
the user explicitly directed the autoloop to continue and run `$refactor` after all selected slices.

Manual visual verification command:

```bash
make bot-visual scenario=24_boss_floor_gate
```

## Boundaries

- No boss HP, damage multiplier, loot, model, summon count, or boss-floor progression changed.
- No protocol schema changed; existing boss phase telegraph/hit-shape metadata carries rectangle
  fields.
