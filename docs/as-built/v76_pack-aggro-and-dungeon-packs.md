# v76 As-built — Pack Aggro and Dungeon Packs

Status: Complete

## Shipped

- Generated non-boss dungeon floors now distribute their base monster population into 5 to 7 internal packs with 2 to 5 planned members each.
- Pack placement preserves required monster definitions first, distributes guarded-chest bonus population into the same pack sizing rules, and keeps members close enough for assist activation.
- Chase monster rules now support `assist_radius`, separate from passive `aggro_radius`.
- Aggro-on-hit now uses assist radius for combat joins while passive proximity acquisition still uses aggro radius.
- Champion rarity followers remain best-effort extras and do not make floor generation fail if nearby legal slots are blocked.
- The protocol bot can count matching runtime events and uses that to prove one generated-dungeon pull emits multiple `monster_aggro` events after damage starts.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestDungeonMonsterGeneration|TestDungeonMonsterGenerationCreatesDeterministicPacks|TestAggroOnHit|TestChampionMonstersSpawnWithCommonMinions|TestGuardedChestGenerationGolden'`
- `ARPG_ADDR=:8888 SCENARIO=pack_aggro_and_dungeon_packs ./scripts/bot_local.sh`

## Deferred

- Elite leaders, named pack identity, aura modifiers, and monster role composition are deferred to the next selected autoloop slice.
- Client-side threat readability remains deferred to the selected combat-readability slice.
- Final dungeon density and combat balance remain tuning work after the encounter-shape proof.
