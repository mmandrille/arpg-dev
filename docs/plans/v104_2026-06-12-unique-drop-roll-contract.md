# v104 Plan — Unique Drop Roll Contract

Status: Ready for implementation
Goal: Make unique item rolls attach one compatible global unique effect id.
Architecture: The item roller remains server-authoritative and deterministic. Unique rarity uses
the existing rolled payload shape and `effect_ids`, so the protocol already carries the attachment
through loot, inventory, stash, and market paths.
Tech stack: shared JSON/schema, Go rules loader and item roller, golden fixture, lifecycle docs.

## Baseline And Shortcut Decision

Builds on v103's `unique_effects.v0.json`. No Godot plugin adoption applies because there is no
client presentation work in v104.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/item_templates.v0.json` | Add `unique` rarity |
| Modify | `server/internal/game/rules.go` | Load unique effects |
| Modify | `server/internal/game/shop.go` | Attach compatible unique effect ids during item rolls |
| Modify | `server/internal/game/game_test.go` | Golden/compatibility proof |
| Modify | `shared/golden/item_rolls.json` | Add unique roll case |
| Modify | `tools/validate_shared.py` | Allow unique item-roll golden effect ids |
| Modify | `PROGRESS.md` | Lifecycle close-out |
| Create | `docs/as-built/v104_unique-drop-roll-contract.md` | As-built summary |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/rules.go`
- [x] `server/internal/game/shop.go`
- [x] `server/internal/game/game_test.go`
- [x] `tools/validate_shared.py`

Decision:
- [x] Defer extraction with rationale: the touched code is existing item-roll and rules-loader
  code. v104 is a small deterministic extension and broad extraction would obscure the contract.
  `server/internal/game/rules.go` grows past its allowance to load the new catalog, so the
  grandfathered baseline is intentionally updated for this slice.

Verification:
```bash
make maintainability
```

## Task 1 — Shared Rarity Contract

- [x] Step 1.1: Add `unique` item rarity.
- [x] Step 1.2: Update shared validation expectations for unique item-roll effects.

```bash
make validate-shared
```

## Task 2 — Server Unique Effect Loading And Roll Attachment

- [x] Step 2.1: Load `unique_effects.v0.json` into Go rules.
- [x] Step 2.2: Attach exactly one compatible effect id when rarity is `unique`.
- [x] Step 2.3: Add deterministic tests for unique effect attachment and compatibility.

```bash
cd server && go test ./internal/game/...
```

## Task 3 — Lifecycle Docs And CI

- [x] Step 3.1: Mark plan tasks complete.
- [x] Step 3.2: Add as-built and update `PROGRESS.md` status/lifecycle/deferred notes.

```bash
make maintainability
make ci
```

## Final Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/...`
- [x] `make maintainability`
- [x] `make ci`

## Deferred Scope

Unique effect combat execution, burn DOT ticks, client burning VFX, and bot-visual proof remain
deferred to v105.
