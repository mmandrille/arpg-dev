# v131 Plan — Purple Town Unique Chest

Status: Ready for implementation
Goal: Add a purple town chest that grants one deterministic unique rolled item for every enabled unique effect.
Architecture: Shared rules define a ready `town_unique_chest` service interactable placed in town.
The Go sim owns activation, derives contents from enabled unique effects and compatible templates in
stable order, and emits existing inventory changes. The client remains presentation-only; bot proof
opens the chest and validates effect coverage.
Tech stack: Go sim, shared JSON/schema, Python bot, existing Godot interactable presentation.

## Baseline and shortcut decision

Builds on v103-v108 unique effect execution, v119 live unique-effect drop coverage, and v130 review
guidance. Godot plugin checklist reviewed: reject external adoption because this is server-authored
debug content, not a new inventory UI/art system.

Capacity decision: the unique test chest bypasses normal inventory capacity for this explicit
debug/test interactable. This keeps the chest deterministic as the enabled effect catalog grows and
does not alter natural loot, shop, stash, or pickup capacity behavior.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/interactables.v0.json` | Add `town_unique_chest` service interactable |
| Modify | `shared/rules/interactables.v0.schema.json` | Allow `unique_test_chest` service |
| Modify | `shared/rules/worlds.v0.json` | Place chest in town near services |
| Modify | `server/internal/game/rules.go` | Validate new service kind |
| Create | `server/internal/game/unique_chest.go` | Deterministic unique grant helpers |
| Create | `server/internal/game/unique_chest_test.go` | Server grant coverage |
| Modify | `server/internal/game/sim.go` | Route interactable activation |
| Modify | `tools/bot/run.py` | Add inventory unique-effect coverage assertion if needed |
| Modify | `tools/bot/test_protocol.py` | Discover and unit-test the scenario/assertion |
| Create | `tools/bot/scenarios/61_purple_town_unique_chest.json` | Bot proof |
| Create | `docs/as-built/v131_purple-town-unique-chest.md` | As-built summary |
| Modify | `PROGRESS.md` | Lifecycle and summary update |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/rules.go`
- [x] `server/internal/game/sim.go`
- [x] `tools/bot/run.py`
- [x] `tools/bot/test_protocol.py`
- [x] `client/scripts/main.gd`

Decision:
- [x] Extract focused helper/test file as part of this slice: use `unique_chest.go` and
  `unique_chest_test.go`.
- [x] Defer splitting `rules.go`, `sim.go`, and bot monoliths because this slice needs only small
  routing/assertion additions and broader extraction would add risk.
- [x] Update the `main.gd` baseline to 6718 with this documented exception. The overage is dominated
  by existing projectile-preview work in the same file; the unique chest adds only a small tint hook.
- [x] Update `rules.go` to 3238 and `game_test.go` to 9044 for current skill cooldown
  validation/test growth already present in the worktree; future skill rule slices should split
  those files before adding more broad coverage there.

Verification:
```bash
make maintainability
```

## Task 1 — Shared Rules

Files:
- Modify: `shared/rules/interactables.v0.json`
- Modify: `shared/rules/interactables.v0.schema.json`
- Modify: `shared/rules/worlds.v0.json`
- Modify: `server/internal/game/rules.go`

- [x] Step 1.1: Add `unique_test_chest` as an allowed ready service kind and validate it in Go.
- [x] Step 1.2: Add `town_unique_chest` with a clear name and `service: "unique_test_chest"`.
- [x] Step 1.3: Place one `town_unique_chest` in the default town level on `dungeon_levels`.

```bash
make validate-shared
```

## Task 2 — Server Unique Chest Grants

Files:
- Create: `server/internal/game/unique_chest.go`
- Create: `server/internal/game/unique_chest_test.go`
- Modify: `server/internal/game/sim.go`

- [x] Step 2.1: Route `unique_test_chest` activation before generic ready/closed handling.
- [x] Step 2.2: Build deterministic grant rows by sorting enabled ready unique effect ids, selecting
  the first compatible item template by sorted id, and creating a unique `ItemRollPayload` with that
  exact effect id.
- [x] Step 2.3: Add items directly to inventory with normal `inventory_add` changes and an
  activation event; bypass capacity only for this service.
- [x] Step 2.4: Mark the chest open after successful grant and reject repeat activation without
  duplicating items.
- [x] Step 2.5: Add focused Go tests for effect coverage, deterministic ordering, compatibility,
  and repeat activation.

```bash
cd server && go test ./internal/game -run 'TestUniqueChest|TestRules'
```

## Task 3 — Bot Proof

Files:
- Modify: `tools/bot/run.py`
- Modify: `tools/bot/test_protocol.py`
- Create: `tools/bot/scenarios/61_purple_town_unique_chest.json`

- [x] Step 3.1: Add an assertion for inventory coverage of all enabled unique effects.
- [x] Step 3.2: Add a protocol scenario that starts in town, opens `town_unique_chest`, and asserts
  effect coverage plus unique rarity inventory count.
- [x] Step 3.3: Add test_protocol discovery coverage for the new scenario.

```bash
make bot scenario=purple_town_unique_chest
```

## Task 4 — Docs And Closeout

Files:
- Modify: `docs/specs/v131_spec-purple-town-unique-chest.md`
- Create: `docs/as-built/v131_purple-town-unique-chest.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Mark the spec complete after implementation.
- [x] Step 4.2: Add the v131 lifecycle row and summary to `PROGRESS.md`.
- [x] Step 4.3: Write the as-built with the debug capacity bypass and manual visual command.

```bash
rg -n "v131|purple-town-unique-chest|purple_town_unique_chest|town_unique_chest" docs/specs docs/plans docs/as-built PROGRESS.md tools/bot/scenarios
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestUniqueTestChest|TestRules'`
- [x] `make bot scenario=purple_town_unique_chest`
- [x] `make ci`

Manual visual check:
```bash
make bot-visual scenario=purple_town_unique_chest
```
