# v62 Plan - Monster Depth Stat Scaling

Status: Complete
Goal: Generated dungeon monsters scale through combat stats, not just HP/damage/XP.
Architecture: Keep monster definitions as archetype bases and apply data-driven depth plus rarity
modifiers only during generated dungeon monster spawn. Boss templates remain bespoke. No protocol
shape changes are needed because combat resolution remains server-authoritative.
Tech stack: Go sim, shared JSON rules/schemas/goldens, existing bot and CI gates.

## Baseline and Shortcut Decision

Builds on v30 monster rarity, v31 combat stat effects, v48 co-op challenge scaling, and v56 attack
cadence. No Godot UI, camera, inventory presentation, or art is in scope, so the plugin adoption
checklist is not applicable.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/dungeon_generation.v0.json` | Add depth and rarity stat scaling data |
| Modify | `shared/rules/dungeon_generation.v0.schema.json` | Validate scaling data |
| Modify | `shared/golden/monster_rarity.json` | Expected effective monster stats |
| Modify | `shared/golden/monster_rarity.v0.schema.json` | Validate new fixture fields |
| Modify | `server/internal/game/rules.go` | Add structs and validation |
| Modify | `server/internal/game/sim.go` | Apply scaling during generated spawns and attacks |
| Modify | `server/internal/game/game_test.go` | Unit/golden coverage |
| Add | `docs/as-built/v62_monster-depth-stat-scaling.md` | Slice summary |
| Modify | `PROGRESS.md` | Mark v62 complete |

## Task 1 - Shared Rules and Fixture Shape

Files:
- Modify: `shared/rules/dungeon_generation.v0.json`
- Modify: `shared/rules/dungeon_generation.v0.schema.json`
- Modify: `shared/golden/monster_rarity.json`
- Modify: `shared/golden/monster_rarity.v0.schema.json`

- [x] Step 1.1: Add `monster_depth_scaling` for HP, damage, armor, hit chance, crit chance,
  block percent, attack cooldown multiplier, and minimum attack cooldown ticks.
- [x] Step 1.2: Extend each monster rarity with stat multipliers/bonuses for the same stats.
- [x] Step 1.3: Extend the golden fixture expected fields.
- [x] Step 1.4: Validate shared data.

```bash
make validate-shared
```

## Task 2 - Go Rules Loading and Validation

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 2.1: Add Go structs for depth stat scaling and rarity stat scaling.
- [x] Step 2.2: Validate positive multipliers, safe caps, and attack cooldown floor.
- [x] Step 2.3: Add focused tests for validation failures.

```bash
cd server && go test ./internal/game/... -run TestRules
```

## Task 3 - Generated Monster Scaling

Files:
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 3.1: Add a deterministic helper that computes effective generated monster stats from
  base definition, dungeon depth, and rarity.
- [x] Step 3.2: Apply effective HP, attack damage, XP, armor, hit chance, crit chance,
  block percent, and attack cooldown during generated monster spawn.
- [x] Step 3.3: Ensure monster attack resolution uses scaled stats and cadence.
- [x] Step 3.4: Preserve boss-template and static/lab monster behavior.

```bash
cd server && go test ./internal/game/... -run 'TestMonsterRarity|TestGeneratedDungeonMonster'
```

## Task 4 - Bot and Regression Gates

Files:
- Existing bot scenarios only

- [x] Step 4.1: Run the protocol bot to verify existing combat and dungeon flows still complete.
- [x] Step 4.2: Run Go tests after focused iteration.

```bash
make bot
make test-go
```

## Task 5 - Lifecycle Docs and CI

Files:
- Add: `docs/as-built/v62_monster-depth-stat-scaling.md`
- Modify: `PROGRESS.md`
- Modify: `docs/plans/v62_2026-06-11-monster-depth-stat-scaling.md`

- [x] Step 5.1: Record the as-built behavior.
- [x] Step 5.2: Add v62 to `PROGRESS.md` as the latest completed slice.
- [x] Step 5.3: Mark plan checkboxes complete after verification.

```bash
make ci
```

## Final Verification

- [x] `make validate-shared`
- [x] `make test-go`
- [x] `make bot`
- [x] `make ci`
