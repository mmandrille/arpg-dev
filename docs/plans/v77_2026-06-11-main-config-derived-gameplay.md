# v77 Plan — Main config derived gameplay

Status: Complete

Architecture: v76 introduced `main_config.v0.json` as a validated mirror. v77 makes attack cadence and player movement consume it directly while preserving older rule files as compatibility inputs for adjacent combat/navigation settings.

## Tasks

- [x] Step 1: Route server attack cadence through `Rules.MainConfig.Gameplay.BaseAttackIntervalTicks`.
- [x] Step 2: Route server direct and auto movement through `Rules.MainConfig.Gameplay.BaseMovementSpeed`.
- [x] Step 3: Relax shared validation mirror checks for attack interval and movement speed; keep sanity checks on old rule files.
- [x] Step 4: Add focused Go tests showing main-config-only changes affect attack interval and movement distance.
- [x] Step 5: Update `PROGRESS.md`, mark spec/plan complete, and write as-built notes.
- [x] Step 6: Run targeted checks, then `make ci` before commit.

## Verification

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestMainConfig'
make ci
```
