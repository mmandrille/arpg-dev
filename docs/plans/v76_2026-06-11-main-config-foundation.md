# v76 Plan — Main config foundation

Status: Complete
Goal: Add a validated shared main config and expose it through the Go rules loader without changing gameplay behavior.
Architecture: `main_config.v0.json` becomes the first top-level tuning aggregate for values that currently live in narrower rules files or hardcoded constants. This slice keeps the existing authoritative consumers in place and adds drift guards so the new config starts as a safe mirror. Later slices can consume these values directly.
Tech stack: Shared JSON schema/data, Python validator, Go rules loader and focused Go tests.

## Baseline and shortcut decision

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `shared/rules/main_config.v0.json` | Designer-facing top-level gameplay tuning values. |
| Create | `shared/rules/main_config.v0.schema.json` | Validate the main config contract. |
| Modify | `server/internal/game/rules.go` | Load and expose main config on `Rules`. |
| Modify | `server/internal/game/game_test.go` | Focused loader coverage for the new config. |
| Modify | `tools/validate_shared.py` | Cross-check mirrored values against combat, navigation, and dungeon monster drop rate. |
| Modify | `docs/specs/v76_spec-main-config-foundation.md` | Mark complete if implementation matches. |
| Modify | `PROGRESS.md` | Record slice completion. |
| Create | `docs/as-built/v76_main-config-foundation.md` | Record what shipped. |

## Task 1 — Shared main config contract

Files:
- Create: `shared/rules/main_config.v0.json`
- Create: `shared/rules/main_config.v0.schema.json`

- [x] Step 1.1: Add `base_attack_interval_ticks`, `base_movement_speed`, and `base_drop_rate_percent` with current values.
- [x] Step 1.2: Add schema bounds for positive attack interval, positive movement speed, and drop rate `[0, 100]`.

```bash
make validate-shared
```

## Task 2 — Go rules loader

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 2.1: Add `MainConfig` types and `Rules.MainConfig`.
- [x] Step 2.2: Load and validate `main_config.v0.json` before dependent rules.
- [x] Step 2.3: Add a focused loader assertion for loaded values.

```bash
cd server && go test ./internal/game -run 'TestLoadRules|TestMainConfig'
```

## Task 3 — Drift guards

Files:
- Modify: `tools/validate_shared.py`

- [x] Step 3.1: Load `main_config.v0.json` in shared validation.
- [x] Step 3.2: Cross-check attack interval against `combat.v0.json`.
- [x] Step 3.3: Cross-check movement speed against `navigation.v0.json.cell_size`.
- [x] Step 3.4: Cross-check base drop rate against unique dungeon monster treasure-class drop rates.

```bash
make validate-shared
```

## Task 4 — Lifecycle docs and CI

Files:
- Modify: `docs/specs/v76_spec-main-config-foundation.md`
- Modify: `docs/plans/v76_2026-06-11-main-config-foundation.md`
- Modify: `PROGRESS.md`
- Create: `docs/as-built/v76_main-config-foundation.md`

- [x] Step 4.1: Mark spec and plan complete.
- [x] Step 4.2: Update `PROGRESS.md` latest slice and lifecycle table.
- [x] Step 4.3: Add as-built summary.

```bash
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestLoadRules|TestMainConfig'`
- [x] `make ci`

## Deferred scope

- v77 will make combat/movement calculations and assertions consume or derive from main config.
- v78 will replace repeated treasure-class drop weights with reusable drop-rate profiles.
- Class/skill/monster full config unification remains deferred until after the three-slice MVP.
