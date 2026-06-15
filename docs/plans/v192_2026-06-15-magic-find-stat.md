# v192 Plan - Magic Find Stat

Status: Complete
Goal: Add a visible equipped Magic Find stat that biases monster rolled-equipment rarity.
Architecture: Magic Find remains server-authoritative. The stat is stored as a normal rolled item
stat, aggregated through the existing effective combat/derived stat path, and consumed only by
monster loot item-template rolls. Shop and authored payload roll paths keep baseline rarity weights.
Tech stack: shared JSON/schema, Go sim, Python bot scenario, client stat label, SDD docs.

## Baseline and shortcut decision

Builds on v191 affix names. No new Godot UI/art is introduced; existing inventory/stat panels render
server-authored fields, so the plugin adoption checklist is not applicable.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/item_templates.v0.schema.json` | Allow `magic_find_percent` roll stats. |
| Modify | `shared/rules/item_templates.v0.json` | Add a narrow Magic Find roll candidate. |
| Modify | `shared/protocol/session_snapshot.v8.schema.json` / `state_delta.v8.schema.json` | Add derived stat field if schemas require it. |
| Modify | `server/internal/game/derived_stats.go` | Expose Magic Find in derived stats. |
| Modify | `server/internal/game/sim.go` or focused helper | Aggregate equipped Magic Find and apply to monster item rolls. |
| Modify | `client/scripts/stat_labels.gd` | Add readable label for Magic Find. |
| Add/Modify | `server/internal/game/*magic_find*_test.go` | Focused formula and stat coverage. |
| Add | `tools/bot/scenarios/81_magic_find_stat.json` | Protocol proof. |
| Modify | `PROGRESS.md` | Mark v192 complete at finish. |
| Add | `docs/as-built/v192_magic-find-stat.md` | Record what shipped. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/sim.go`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: protocol schemas do not count as source; check touched Go/client/tool files.
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Prefer focused helper/test files. If `sim.go` needs only a call-site change under allowance,
  defer broader extraction to `$refactor`.

Verification:
```bash
make maintainability
```

## Task 1 - Shared stat contract

Files:
- Modify: `shared/rules/item_templates.v0.schema.json`
- Modify: `shared/rules/item_templates.v0.json`
- Modify: protocol schemas if required
- Modify: `client/scripts/stat_labels.gd`

- [x] Add `magic_find_percent` as a valid bounded roll stat.
- [x] Add a deterministic roll candidate to a jewelry item.
- [x] Add stat label and protocol schema coverage for derived stats.

```bash
make validate-shared
```

## Task 2 - Server authority and tests

Files:
- Modify: `server/internal/game/derived_stats.go`
- Modify: `server/internal/game/sim.go` or focused helper
- Add/Modify: `server/internal/game/*magic_find*_test.go`

- [x] Aggregate equipped Magic Find into effective stats and breakdowns.
- [x] Bias monster item-template rarity rolls with Magic Find while preserving shop baseline rolls.
- [x] Add focused deterministic tests for visible stat and rarity influence.

```bash
cd server && go test ./internal/game -run 'MagicFind|ItemRollsGolden|ShopGeneratedOfferGolden' -count=1
```

## Task 3 - Bot proof

Files:
- Add: `tools/bot/scenarios/81_magic_find_stat.json`

- [x] Add a compact scenario proving equip, derived stat, and stat breakdown over protocol. Focused
  Go coverage proves the monster loot rarity influence.

```bash
make bot scenario=81_magic_find_stat.json
```

## Task 4 - Lifecycle docs and CI

Files:
- Modify: `PROGRESS.md`
- Add: `docs/as-built/v192_magic-find-stat.md`

- [x] Add lifecycle row and as-built note.
- [x] Run final verification.

```bash
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'MagicFind|ItemRollsGolden|ShopGeneratedOfferGolden' -count=1`
- [x] `make bot scenario=81_magic_find_stat.json`
- [x] `make ci`
