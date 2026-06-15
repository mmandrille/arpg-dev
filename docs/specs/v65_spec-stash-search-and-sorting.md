# Spec: `stash-search-and-sorting`

Status: Accepted
Date: 2026-06-11
Branch: `main`
Codename: `stash-search-and-sorting`
Slice: v65 - stash search and sorting
Baseline: v64 `mystery-seller-paid-reroll`
Related:

- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`v50_spec-account-stash-storage.md`](v50_spec-account-stash-storage.md)

## 1. Purpose

The account stash is durable, but the client always shows raw server order. This slice adds local
stash search and sorting so a player can quickly find stored equipment without changing server
ownership, item identity, persistence, or transfer semantics.

## 2. Non-goals

- No protocol, schema, server, persistence, or replay contract changes.
- No tabs, item stacks, materials, capacity upgrades, market delivery, or stash overflow.
- No new item mutation, direct equip/use/sell from stash, or drag rules.

## 3. Acceptance Criteria

1. The Godot stash panel exposes a search field that filters visible stash rows by item name,
   item definition/template id, rarity, slot, and tooltip summary text.
2. Empty search shows all stash items.
3. Search is case-insensitive and trims leading/trailing whitespace.
4. The stash panel exposes a sort control with at least: acquired/default, name, rarity, and slot.
5. Sorting and searching are display-only: deposit/withdraw intents still use server-authored
   `stash_item_id` and existing item payloads.
6. Debug state reports search text, selected sort mode, filtered count, and rendered stash rows.
7. Existing stash deposit/withdraw/gold UI remains enabled/disabled as before.
8. Client bot scenarios can set stash search text, choose a sort mode, assert filtered counts, and
   withdraw a filtered item.
9. `make client-unit`, targeted client bot scenario, and `make ci` pass.

## 4. Scope

```text
client/scripts/stash_panel.gd
client/scripts/bot_controller.gd
client/scripts/bot_scenario_runner.gd
client/scripts/main.gd
client/tests/test_stash_panel.gd
client/tests/test_client_bot.gd
tools/bot/scenarios/client/30_stash_search_and_sorting.json
docs/plans/v65_2026-06-11-stash-search-and-sorting.md
docs/as-built/v65_stash-search-and-sorting.md
PROGRESS.md
```

## 5. Verification

- `make client-unit`
- `make bot-client scenario=30_stash_search_and_sorting`
- `make ci`
