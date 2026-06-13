# v131 As-Built — Purple Town Unique Chest

Date: 2026-06-13

## What shipped

- Added `town_unique_chest`, a purple ready interactable in town.
- Activating it grants one deterministic unique rolled item for every enabled ready unique effect.
- Chest contents are derived from `unique_effects.v0.json` and compatible item templates in sorted
  order; natural loot/drop odds are unchanged.
- The chest bypasses normal inventory capacity only for this debug/test surface, emits normal
  `inventory_add` changes, opens after use, and cannot be used again to duplicate items.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestUniqueTestChest|TestRules'`
- `.venv/bin/python -m pytest tools/bot/test_protocol.py -q`
- `make bot scenario=purple_town_unique_chest`
- `make maintainability`
- `make ci`

Manual visual check:

```bash
make bot-visual scenario=purple_town_unique_chest
```

## Maintenance note

`client/scripts/main.gd` baseline is refreshed to 6718 lines with this documented exception. The
overage is dominated by existing projectile-preview code in the worktree; v131 only adds the small
purple chest presentation hook. Future client slices should avoid adding more behavior to
`main.gd`.

`server/internal/game/rules.go` and `server/internal/game/game_test.go` baselines are also refreshed
for the current skill cooldown validation/test growth that was already in the worktree. Future skill
rule slices should split those files before adding more broad coverage there.
