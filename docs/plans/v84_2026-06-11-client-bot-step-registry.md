# v84 Plan - Client Bot Step Registry

Status: Ready for implementation
Goal: Remove duplicated client bot step registry maintenance by deriving `ALL_STEP_TYPES`.
Architecture: Keep the existing category arrays as the source of truth. `ALL_STEP_TYPES` becomes
their concatenation, and validation continues to use `stype not in ALL_STEP_TYPES`.
Tech stack: Godot GDScript client bot runner and unit tests.

## Baseline and shortcut decision

this is test tooling infrastructure, not UI, camera, inventory presentation, or art.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/bot_scenario_runner.gd` | Derive `ALL_STEP_TYPES` from category step lists. |
| Modify | `client/tests/test_client_bot.gd` | Add a registry derivation unit assertion. |
| Add | `docs/as-built/v84_client-bot-step-registry.md` | Summarize shipped behavior. |
| Modify | `PROGRESS.md` | Mark v84 complete and close the review finding. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/bot_scenario_runner.gd`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected

Decision:
- [x] Defer extraction with rationale: this slice removes registry duplication in-place; broader
  bot runner decomposition belongs to a separate tooling split.

Verification:
```bash
make maintainability
```

## Task 1 - Derive all step types

Files:
- Modify: `client/scripts/bot_scenario_runner.gd`

- [x] Step 1.1: Replace the hand-maintained `ALL_STEP_TYPES` contents with a derived registry.
- [x] Step 1.2: Preserve category arrays and existing validation code paths.

```bash
make client-unit
```

## Task 2 - Unit proof

Files:
- Modify: `client/tests/test_client_bot.gd`

- [x] Step 2.1: Add a test proving every category entry appears in `ALL_STEP_TYPES`.
- [x] Step 2.2: Keep the existing unknown-step rejection test unchanged.

```bash
make client-unit
```

## Task 3 - Lifecycle docs and CI

Files:
- Modify: `docs/specs/v84_spec-client-bot-step-registry.md`
- Modify: `docs/plans/v84_2026-06-11-client-bot-step-registry.md`
- Add: `docs/as-built/v84_client-bot-step-registry.md`
- Modify: `PROGRESS.md`

- [x] Step 3.1: Mark spec and plan complete.
- [x] Step 3.2: Add as-built and progress updates.
- [x] Step 3.3: Run final verification.

```bash
make maintainability
make ci
```

## Final verification

- [x] `make client-unit`
- [x] `make maintainability`
- [x] `make ci`

Deferred scope: new scenario verbs and broader bot runner decomposition.
