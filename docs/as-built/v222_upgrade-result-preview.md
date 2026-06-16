# v222 As-Built - Upgrade Result Preview

Date: 2026-06-16

## What shipped

- Extracted blacksmith preview computation into `client/scripts/blacksmith_upgrade_preview.gd`, keeping
  the blacksmith panel below the file-size ratchet while preserving rolled, summary, and template
  stat-delta previews.
- Added staged-item preview lines for success result, possible failure result, spend on accepted
  attempt, and after-attempt gold/resource balances.
- Kept guaranteed-pity attempts from showing a possible-failure outcome.
- Added a client bot `click_blacksmith_stage_item` action so the blacksmith scenario can stage an
  item, assert preview text, and only then click Upgrade.
- Extended blacksmith bot wait/assert matching with `preview_contains`.

## Proof

```bash
make maintainability
make client-unit
make bot-client scenario=blacksmith_upgrade_ui
```

All focused checks passed on 2026-06-16 during `$autoloop`. The enclosing batch-level `make ci`
remains deferred until the selected feature queue is complete.

Manual visual proof, if desired:

```bash
make bot-visual scenario=blacksmith_upgrade_ui
```

## Scope limits

- No server, shared rules, protocol, upgrade formula, resource wallet, or HTTP route behavior changed.
- No new recipes, rarity curves, refunds, bricking, binding rules, or multi-resource costs shipped.
