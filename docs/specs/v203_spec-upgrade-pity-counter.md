# v203 Spec: Upgrade Pity Counter

Status: Complete
Date: 2026-06-15
Codename: upgrade-pity-counter

## Purpose

Make blacksmith upgrade failures accumulate item-owned pity progress. After enough accepted failed
attempts, the next accepted attempt on that same item is guaranteed to succeed.

## Player-visible Behavior

- Blacksmith preview/debug state shows the configured pity threshold and the staged item's current
  failure count.
- Failed accepted attempts still spend gold and the configured resource, but increment pity progress.
- Once an item reaches the configured threshold, the next accepted attempt succeeds even if the
  success roll would otherwise fail.
- A successful upgrade resets the item's pity failure count.

## Scope

- Add `item_upgrade_pity_failure_threshold` to shared `main_config`.
- Store pity progress in rolled item metadata as `upgrade_pity.failures`; no database migration.
- Thread the threshold through HTTP/store upgrade calls.
- Update failed-upgrade persistence so returned items include the new failure count.
- Update the blacksmith panel/debug state and bot matcher for pity threshold/count/guaranteed state.
- Add focused deterministic store coverage for fail, fail, guaranteed success.

## Non-goals

- No per-rarity pity curves, pity UI styling polish, account-wide pity, resource refunds, or pity
  transfer outside the item metadata.
- No recipe redesign, resource wallet, or market restrictions.
- No change to direct stash-upgrade routing beyond using the same item-owned pity rules.

## Acceptance Criteria

- Shared validation rejects negative pity thresholds.
- A forced-failure store test increments pity and guarantees success once threshold is reached.
- Failed upgrade responses return the same item with persisted pity metadata.
- The blacksmith panel exposes `pity_failure_count`, `pity_threshold`, and `pity_guaranteed`.
- The blacksmith client bot proof asserts the default pity debug fields.
- `make validate-shared`, focused store/http/client checks, focused client bot proof, and `make ci`
  pass.

## Asset/plugin decision

Rejected. This slice only adds item metadata and existing blacksmith-panel text/debug state.
