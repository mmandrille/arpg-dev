# v15 Plan — Item visuals and loot presentation

Status: In progress (2026-06-06)

## 1. Adoption checklist

  and shared JSON metadata only.
- **Godot version:** compatible with existing Godot 4.6.x project; no addon import.
- **Authoritative boundary:** presentation metadata only; Go server/protocol unchanged.
- **Agent ergonomics:** text-friendly JSON + GDScript, no editor-only setup.
- **Maintenance:** no new third-party dependency.
- **Integration cost:** low; validate via existing `make validate-shared`, `make client-unit`,
  `make bot-client`, and `make ci`.
  item presentation, not a new inventory framework or production art pass.

## 2. File map

| Action | Path | Purpose |
|--------|------|---------|
| Add | `shared/assets/item_presentations.v0.schema.json` | Presentation-only item icon/loot schema |
| Add | `shared/assets/item_presentations.v0.json` | Current item presentation metadata |
| Modify | `tools/validate_shared.py` | Cross-check presentation keys against item rules |
| Modify | `client/scripts/inventory_panel.gd` | Render shaped item icons in slots |
| Modify | `client/scripts/main.gd` | Render shaped ground loot and expose presentation debug state |
| Modify | `client/tests/test_item_visuals.gd` | Assert all current items have presentation metadata |
| Modify | `client/scripts/bot_scenario_runner.gd` | Add client-bot assertions for item presentation |
| Modify | client scenario JSON | Exercise presentation assertions |
| Modify | `PROGRESS.md` | Mark v15 complete when shipped |

## 3. Implementation steps

- [x] Step 1: Add shared `item_presentations` schema/data and validator cross-checks.
- [x] Step 2: Add a Godot presentation loader/render helper in inventory and loot paths.
- [x] Step 3: Extend Godot tests and client-bot assertions.
- [x] Step 4: Run focused validation/tests, then `make ci`.
- [x] Step 5: Update `PROGRESS.md` with v15 lifecycle notes.

## 4. Verification commands

```bash
make validate-shared
make validate-assets
make client-unit
make bot-client
make ci
```
