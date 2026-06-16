# v222 Spec - Upgrade Result Preview

Status: Complete - focused `$autoloop` verification green; final batch CI pending
Date: 2026-06-16
Codename: `upgrade-result-preview`
Baseline: v221 `resource-wallet-foundation`

## Purpose

The blacksmith already shows item level, cost, success chance, pity progress, resource availability,
and stat deltas. This slice makes the attempted result explicit before the player clicks Upgrade:
the staged item preview must spell out the success outcome, the failure outcome when failure is
possible, and the post-attempt gold/resource balances that will remain after an accepted attempt.

The slice is presentation-only. The server remains authoritative for success rolls, cost/resource
spending, item mutation, and pity updates.

## Non-goals

- No change to upgrade odds, costs, pity rules, resource wallet persistence, item mutation, or HTTP
  response shape.
- No new blacksmith recipes, rarity curves, refunds, bricking, binding, market restrictions, or
  multi-resource recipes.
- No standalone material wallet UI or blacksmith art/audio polish.
- No protocol/schema bump; all data comes from existing panel inputs and current item payloads.

## Acceptance criteria

- When an upgradeable item is staged, the visible blacksmith preview includes:
  - success chance or guaranteed-next-upgrade state,
  - the success result (`level current -> next` plus stat increase lines where available),
  - the failure result when failure is possible (`item unchanged` plus pity progress),
  - the accepted-attempt spend and after-attempt gold/resource balances.
- When the staged item is guaranteed by pity, the preview does not present failure as a possible
  result and instead identifies the attempt as guaranteed.
- The panel debug state exposes the same preview lines so the client bot can assert them.
- Existing stat-delta previews from rolled stats, summary lines, and template base stats still work.
- `client/scripts/blacksmith_panel.gd` stays within the maintainability ratchet; preview calculation
  moves into a focused helper rather than growing the panel coordinator.
- The existing `blacksmith_upgrade_ui` client bot scenario asserts the new preview lines before the
  upgrade click.

## Scope and likely files

- `client/scripts/blacksmith_upgrade_preview.gd` - new focused preview calculator.
- `client/scripts/blacksmith_panel.gd` - delegate preview/debug computation to the helper and render
  the extra lines.
- `client/tests/test_blacksmith_panel.gd` and existing blacksmith-related client tests - focused
  coverage for failure, guaranteed, spend, and balance preview lines.
- `client/scripts/bot_scenario_runner.gd`, `client/scripts/bot_step_catalog.gd` - add
  `preview_contains` matching for blacksmith panel assertions.
- `tools/bot/scenarios/client/39_blacksmith_upgrade_ui.json` - assert the result preview in the
  player-facing blacksmith scenario.
- Lifecycle docs: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`,
  `docs/as-built/v222_upgrade-result-preview.md`.

## Test and bot proof

- `make maintainability`
- `make client-unit`
- `make bot-client scenario=blacksmith_upgrade_ui`

Manual visual proof, if desired:

```bash
make bot-visual scenario=blacksmith_upgrade_ui
```

## Client asset/plugin decision

Reject external assets/plugins. This slice reuses existing in-repo Godot panel, label, and item-icon
presentation code; it adds text preview behavior only.

## Open questions and risks

- No blocking questions. The default is to preview deterministic consequences of an accepted attempt
  while clearly keeping the actual random outcome server-authored.
- Risk: preview wording can drift from server behavior. Mitigation: derive costs, wallet balances,
  level, pity, and stat deltas from the same panel config/item payload already used by the blacksmith
  UI and cover the wording through unit and bot checks.
