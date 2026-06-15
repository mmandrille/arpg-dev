# v183 As-Built — Ranger black wolf companion

Date: 2026-06-15

## What shipped

- Added a `summon_companion` skill kind with a data-driven `companion` payload.
- Added Ranger skill `black_wolf_companion`, gated behind `piercing_shot`, with one active companion.
- Added `companion_black_wolf` monster data, visual presentation, text catalog entries, and skill presentation data.
- Implemented summon handling that spends mana, starts cooldown, removes the owner’s previous wolf for that skill, and spawns a server-owned `companion`.
- Added `ranger_wolf_lab` and protocol bot scenario `74_ranger_wolf_companion.json`.

## Key Decisions

- The wolf is server-owned and uses the v182 companion follow/assist/attack AI.
- Recast replacement removes the old wolf entity and spawns a fresh companion.
- The wolf uses the existing `monster_quadruped` scene and a black `visual_tint`; no plugin or new art dependency was adopted.
- Wolf count remains fixed at one active companion for v183. Rank scaling and broader companion tuning remain deferred to v185.

## Verification

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'Ranger|Summon|Companion' -count=1
make bot scenario=74_ranger_wolf_companion.json
make ci
```

Manual visual proof:

```bash
make bot-visual scenario=74_ranger_wolf_companion.json
```

## Deferred

- v184 Sorcerer revive using the companion system.
- v185 data-driven companion scaling and quantity limits.
- v186 elite minions using companion-like follow/assist behavior.
- Ranger multi-wolf scaling, companion UI, persistence, equipment, commands, and leveling.
