# v84 As-built - Client Bot Step Registry

## What shipped

The Godot client bot runner now derives `ALL_STEP_TYPES` from the existing
`STEP_TYPES_WAIT`, `STEP_TYPES_ASSERT`, and `STEP_TYPES_ACTION` category arrays instead of keeping
a separate hand-maintained duplicate list.

## What it proves

- Adding a future step type to the proper category list automatically makes it known to validation.
- Unknown step types still reject through the existing validation path.
- Client bot unit coverage now verifies that every category entry is present in `ALL_STEP_TYPES`.

## Verification

```bash
make client-unit
make maintainability
make ci
```

## Deferred

- New scenario verbs.
- Broader `client/scripts/bot_scenario_runner.gd` decomposition.
