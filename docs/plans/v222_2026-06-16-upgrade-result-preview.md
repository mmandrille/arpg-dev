# v222 Plan - Upgrade Result Preview

Status: Complete - focused `$autoloop` verification green; final batch CI pending
Goal: Make blacksmith upgrade attempts explain success, failure, spend, and after-attempt balances before the player clicks Upgrade.
Architecture: Presentation stays client-owned and server authority is unchanged. A new focused GDScript
helper computes preview lines from the current blacksmith config, staged item, gold, stash gold, and
resource wallet. The panel renders those lines and bot/debug state exposes them for automated proof.
Tech stack: Godot client, client bot scenario JSON, lifecycle docs.

## Baseline and shortcut decision

Builds on v197 success chance, v203 pity counters, and v221 resource wallet blacksmith spending.
No shared/server/protocol changes are planned.

Asset/plugin decision: reject external assets/plugins. Reuse existing in-repo blacksmith panel,
label layout, item icon drawing, and client bot infrastructure.

## Spec Review

- Baseline: v222 follows v221 on `main`.
- Scope: presentation-only; no upgrade formula, wallet persistence, or route changes.
- Contracts: no protocol/schema/golden changes.
- Determinism/server authority: random success and mutation remain server-owned; client only explains
  possible outcomes from existing state.
- Bot proof: existing `blacksmith_upgrade_ui` scenario will assert preview text.
- Maintainability: `client/scripts/blacksmith_panel.gd` is 596 lines, so preview logic must be
  extracted before adding behavior.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `client/scripts/blacksmith_upgrade_preview.gd` | Focused preview computation and stat helper logic. |
| Modify | `client/scripts/blacksmith_panel.gd` | Delegate preview computation and render the new lines. |
| Modify | `client/tests/test_blacksmith_panel.gd` | Unit proof for normal, failure-possible, and guaranteed previews. |
| Modify | `client/scripts/bot_scenario_runner.gd` | Match `preview_contains` against blacksmith panel debug lines. |
| Modify | `client/scripts/bot_step_catalog.gd` | Validate `preview_contains` as a blacksmith expectation. |
| Modify | `tools/bot/scenarios/client/39_blacksmith_upgrade_ui.json` | Assert new preview lines in client bot flow. |
| Add | `docs/as-built/v222_upgrade-result-preview.md` | Record proof and deferred scope. |
| Modify | `PROGRESS.md` | Advance current status after the slice. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v222 lifecycle row. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/blacksmith_panel.gd` is near the 600-line limit and must shrink via helper extraction.
- [x] `client/scripts/bot_scenario_runner.gd` is grandfathered; add only a tiny matcher hook.
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Extract focused helper/module/test file as part of this slice.

Verification:

```bash
make maintainability
```

## Task 1 - Extract preview computation

Files:
- Add: `client/scripts/blacksmith_upgrade_preview.gd`
- Modify: `client/scripts/blacksmith_panel.gd`

- [x] Move item level, cost, pity, stat-summary, and stat-delta preview logic into the new helper.
- [x] Add outcome lines for success, possible failure, spend, and after-attempt balances.
- [x] Keep existing preview labels and debug `preview_lines` sourced from the helper.

```bash
make client-unit
```

## Task 2 - Test preview behavior

Files:
- Modify: `client/tests/test_blacksmith_panel.gd`

- [x] Assert normal staged items show success, possible failure, spend, and after-attempt balances.
- [x] Assert pity-guaranteed items do not show a possible-failure outcome.
- [x] Preserve rolled, summary, and template-based stat delta expectations.

```bash
make client-unit
```

## Task 3 - Bot proof

Files:
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/scripts/bot_step_catalog.gd`
- Modify: `tools/bot/scenarios/client/39_blacksmith_upgrade_ui.json`

- [x] Add `preview_contains` support for blacksmith wait/assert steps.
- [x] Update the blacksmith scenario to assert result-preview text before clicking Upgrade.

```bash
make bot-client scenario=blacksmith_upgrade_ui
```

## Task 4 - Lifecycle docs and close-out

Files:
- Add: `docs/as-built/v222_upgrade-result-preview.md`
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `docs/specs/v222_spec-upgrade-result-preview.md`
- Modify: `docs/plans/v222_2026-06-16-upgrade-result-preview.md`

- [x] Mark this plan complete as tasks pass.
- [x] Mark the spec complete.
- [x] Update current status to v222 complete and next selected `$autoloop` slice.
- [x] Add the lifecycle row and as-built summary.

```bash
make maintainability
make client-unit
make bot-client scenario=blacksmith_upgrade_ui
```

## Final verification

- [x] `make maintainability`
- [x] `make client-unit`
- [x] `make bot-client scenario=blacksmith_upgrade_ui`

Batch-level `make ci` remains deferred to `$autoloop` after all selected slices are committed.
