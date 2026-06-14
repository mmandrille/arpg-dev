package game

import "strconv"

const randomQuestRewardFloorRollRange = 10

func dungeonLevelHasRandomQuestReward(seed string, levelNum int, rules DungeonGenerationRules) bool {
	if levelNum >= 0 || isBossFloor(levelNum, rules) {
		return false
	}
	rng := NewRNG(SeedToUint64(seed + "|random_quest_reward|" + strconv.Itoa(absInt(levelNum))))
	return rng.IntN(randomQuestRewardFloorRollRange) == 0
}

func maybePlaceRandomQuestRewardChest(seed string, rules DungeonGenerationRules, lootBand DungeonLootBand, out *generatedDungeonLevel) error {
	placement := rules.ChestPlacement
	if !placement.Enabled || !dungeonLevelHasRandomQuestReward(seed, out.levelNum, rules) {
		return nil
	}
	rng := NewRNG(SeedToUint64(seed + "|random_quest_reward_chest|" + strconv.Itoa(absInt(out.levelNum))))
	pos, ok := randomChestPosition(rng, rules, out)
	if !ok {
		return nil
	}
	out.chests = append(out.chests, generatedChest{
		defID:       placement.InteractableDefID,
		lootTable:   lootBand.ChestLootTable,
		pos:         pos,
		questReward: true,
	})
	return nil
}
