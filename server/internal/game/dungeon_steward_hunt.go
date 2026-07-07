package game

import "strconv"

const stewardHuntFloorRollRange = 10

func dungeonLevelHasStewardHuntQuest(seed string, levelNum int, rules DungeonGenerationRules, steward QuestStewardRules) bool {
	if !steward.HuntQuest.Enabled || levelNum >= 0 || isBossFloor(levelNum, rules) {
		return false
	}
	rng := NewRNG(SeedToUint64(seed + "|steward_hunt|" + strconv.Itoa(absInt(levelNum))))
	return rng.IntN(stewardHuntFloorRollRange) == 0
}

func maybeAssignStewardHuntQuest(seed string, rules DungeonGenerationRules, steward QuestStewardRules, out *generatedDungeonLevel) {
	if !dungeonLevelHasStewardHuntQuest(seed, out.levelNum, rules, steward) {
		return
	}
	rng := NewRNG(SeedToUint64(seed + "|steward_hunt_target|" + strconv.Itoa(absInt(out.levelNum))))
	candidates := make([]int, 0, len(out.monsters))
	for idx, monster := range out.monsters {
		if monster.isBoss {
			continue
		}
		if _, ok := steward.trophyByMonster[monster.defID]; !ok {
			continue
		}
		candidates = append(candidates, idx)
	}
	if len(candidates) == 0 {
		return
	}
	chosen := candidates[rng.IntN(len(candidates))]
	monster := out.monsters[chosen]
	trophy, ok := steward.trophyByMonster[monster.defID]
	if !ok {
		return
	}
	out.monsters[chosen].stewardHuntTarget = true
	out.stewardHunt = &generatedStewardHunt{
		MonsterIndex:    chosen,
		MonsterDefID:    monster.defID,
		TrophyItemDefID: trophy.ItemDefID,
		TrophyLabel:     trophy.TrophyLabel,
	}
}
