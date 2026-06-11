# v79 As-built — Elite Pack Roles

Status: Complete

## Shipped

- Generated dungeon monsters now declare internal pack roles: frontline, ranged, flanker, and swarm.
- Dungeon monster placement rules now define pack composition constraints and elite-pack chance.
- Pack generation composes each planned pack so it has at least one frontline member and no more than one ranged member.
- Elite packs mark one planned internal leader and force that member to champion rarity without spawning extra champion minions from the leader flag.
- Existing random champion rarity can still spawn best-effort common followers independently of the planned pack size.
- The pack aggro bot scenario now pulls from outside passive aggro before attacking, keeping the proof stable with tougher elite packs.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestDungeonMonsterGeneration|TestDungeonMonsterGenerationCreatesDeterministicPacks|TestDungeonMonsterGenerationCanForceElitePackLeaders'`
- `ARPG_ADDR=:8888 SCENARIO=pack_aggro_and_dungeon_packs ./scripts/bot_local.sh`

## Deferred

- Elite abilities, aura modifiers, names, rewards, and client-facing elite labels remain future work.
- Threat/readability presentation is still deferred to the selected combat-readability slice.
