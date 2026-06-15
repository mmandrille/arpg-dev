# v131 Spec — Purple Town Unique Chest

Status: Complete
Date: 2026-06-13
Codename: purple-town-unique-chest

## Purpose

Add a visible purple chest in town that grants one deterministic unique rolled item for every
enabled unique effect in the game. The chest is a development/test surface: it lets bots and humans
inspect and equip all current unique-effect payloads without waiting for live loot RNG.

The server remains authoritative. Chest contents are derived from `unique_effects.v0.json` and
compatible item templates at activation time; the client only renders the interactable and receives
normal inventory updates.

## Non-goals

- No fixed named unique item catalog or flavor-name packages.
- No new player-facing unique inspection UI beyond existing inventory item payloads/tooltips.
- No change to natural dungeon/drop rarity odds.
- No support for reopening the chest to duplicate contents in the same session.

## Acceptance Criteria

1. Town includes a `town_unique_chest` interactable with purple presentation and a stable position
   near existing town services.
2. Activating that chest in town grants inventory items whose rolled payloads have rarity `unique`.
3. The granted set covers every enabled, ready unique effect exactly once by `effect_ids`.
4. Grants are deterministic across seeds/builds by sorting effect ids and choosing compatible
   templates deterministically.
5. The chest emits an activation event and normal `inventory_add` changes so existing replay,
   persistence, and client reconciliation paths keep working.
6. Re-activating the same chest does not duplicate unique items.
7. Bot proof opens the chest and asserts all enabled unique effects are present in inventory.

## Scope And Likely Files

- Shared rules:
  - `shared/rules/interactables.v0.json`
  - `shared/rules/interactables.v0.schema.json`
  - `shared/rules/worlds.v0.json`
  - `shared/assets/item_presentations.v0.json` or existing client presentation mapping if required
- Server:
  - `server/internal/game/sim.go` or a focused helper file for unique chest grants
  - focused Go test file for deterministic grants and no-duplicate behavior
- Tools/bot:
  - new scenario under `tools/bot/scenarios/`
  - `tools/bot/run.py` assertion support if existing inventory assertions cannot match effect ids
  - `tools/bot/test_protocol.py`
- Docs:
  - `docs/plans/v131_2026-06-13-purple-town-unique-chest.md`
  - `docs/as-built/v131_purple-town-unique-chest.md`
  - `PROGRESS.md`

## Test And Bot Proof

- `make validate-shared` validates the new interactable/world data.
- Focused Go tests prove the unique chest grants one unique per enabled effect, keeps effect ids
  compatible with item type, and refuses duplicate grants.
- Bot scenario opens the town chest and asserts inventory covers every enabled unique effect.
- `make bot scenario=purple_town_unique_chest` is the targeted gameplay proof.
- `make ci` is the final closeout gate.

uses existing client interactable rendering; inventory authority and item payloads stay in Go and
shared JSON.

## Resolved Questions And Risks

- Inventory capacity is intentionally bypassed for this explicit debug/test chest so it can cover
  every enabled effect as the catalog grows.
- Bot proof uses a catalog-driven inventory unique-effect coverage assertion.
