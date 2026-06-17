# v247 As-Built - Mercenary Combat Stats

Date: 2026-06-17

## What shipped

- Added optional companion `combat_stats` to the entity view and v8 snapshot schema.
- Populated companion combat stats from actual companion state, using shared monster rules as the
  fallback for damage, cooldown, armor, block, hit chance, and crit chance.
- Preserved `combat_stats` through the client entity record and companion UI sync without growing
  `main.gd`.
- Extended the mercenary stats card with compact combat lines: damage, attack cooldown, defense,
  and accuracy.
- Added `64_mercenary_combat_stats.json` as the client bot proof.
- Fixed stale-session cleanup to delete `session_start_account_resource_wallet` rows before
  deleting old sessions; this was required for local bot server startup on existing DB state.

## Proof

```bash
cd server && go test ./internal/game -run 'MercenaryFoundation|MercenaryHiring' -count=1
cd server && go test ./internal/store -run DeleteStaleEmptySessions -count=1
godot --headless --path client --script res://tests/test_mercenary_panel.gd
make validate-shared
make bot-client scenario=64_mercenary_combat_stats.json HEADLESS=1
make maintainability
```

All focused checks passed on 2026-06-17 during `$autoloop`. The selected v241-v250 batch-level
`make ci` also passed on 2026-06-17 after v250.

Manual visual proof, if desired:

```bash
make bot-visual scenario=64_mercenary_combat_stats.json
```

## Scope limits

- No mercenary balance changes, gear, level scaling, durable roster details, offer variants, new AI,
  new command UI, production portraits, external assets, or external plugins shipped.
