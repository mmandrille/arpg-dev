# v78 Plan — Main config drop profiles

Status: Complete

Architecture: keep authored treasure class entries in `treasure_classes.v0.json`, but let `main_config.gameplay.base_drop_rate_percent` own the success/no-drop split for dungeon monster primary attempts. This gives immediate tuning effect for the global drop rate while leaving richer entry-weight profiles for a later content-balancing slice.

## Tasks

- [x] Step 1: Add a loader pass that finds dungeon monster loot tables and applies the main-config drop rate to their primary attempts.
- [x] Step 2: Keep entry weights unchanged so the current ranking remains stable.
- [x] Step 3: Update shared validation messaging/checks for operational main-config drop rate.
- [x] Step 4: Add focused Go tests with a temp `main_config.v0.json` drop-rate override.
- [x] Step 5: Update `PROGRESS.md` and as-built notes.
- [x] Step 6: Run targeted checks, then `make ci` before commit.

## Verification

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestMainConfig|TestDungeonMonsterLootRate'
make ci
```
