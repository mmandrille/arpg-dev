package game

import (
	"fmt"
	"testing"
)

func TestRandomQuestRewardFloorRollDeterministicAndSparse(t *testing.T) {
	rules := loadRules(t).DungeonGeneration
	const sampleSize = 200

	first := 0
	second := 0
	for depth := 1; depth <= sampleSize; depth++ {
		levelNum := -depth
		if dungeonLevelHasRandomQuestReward("v155_distribution", levelNum, rules) {
			first++
		}
		if dungeonLevelHasRandomQuestReward("v155_distribution", levelNum, rules) {
			second++
		}
	}
	if first != second {
		t.Fatalf("random quest reward roll changed between passes: first=%d second=%d", first, second)
	}
	if first < 10 || first > 30 {
		t.Fatalf("random quest reward count = %d/%d, want roughly 10%%", first, sampleSize)
	}
}

func TestRandomQuestRewardFloorsExcludeBossFloors(t *testing.T) {
	rules := loadRules(t).DungeonGeneration
	for depth := rules.BossFloor.Cadence; depth <= rules.BossFloor.Cadence*4; depth += rules.BossFloor.Cadence {
		levelNum := -depth
		if !isBossFloor(levelNum, rules) {
			t.Fatalf("level %d should be a boss floor for cadence %d", levelNum, rules.BossFloor.Cadence)
		}
		if dungeonLevelHasRandomQuestReward("v155_boss_exclusion", levelNum, rules) {
			t.Fatalf("boss level %d rolled a random quest reward", levelNum)
		}
		level, err := GenerateDungeonLevel("v155_boss_exclusion", levelNum, rules)
		if err != nil {
			t.Fatalf("level %d generate: %v", levelNum, err)
		}
		for _, chest := range level.chests {
			if chest.questReward {
				t.Fatalf("boss level %d generated quest reward chest: %+v", levelNum, chest)
			}
		}
	}
}

func TestRandomQuestRewardFloorPlacesReachableChest(t *testing.T) {
	rules := loadRules(t).DungeonGeneration
	seed, levelNum := findRandomQuestRewardFloorForTest(t, rules, 40)

	level, err := GenerateDungeonLevel(seed, levelNum, rules)
	if err != nil {
		t.Fatalf("generate level %d for seed %s: %v", levelNum, seed, err)
	}
	questChest := findGeneratedQuestRewardChest(level)
	if questChest == nil {
		t.Fatalf("level %d for seed %s missing quest reward chest: %+v", levelNum, seed, level.chests)
	}
	start := generatedReachabilityStart(rules, level)
	if !generatedTargetReachableFrom(rules, level, start, questChest.pos) {
		t.Fatalf("quest reward chest at %+v unreachable from %+v", questChest.pos, start)
	}
	if err := validateGeneratedDungeonReachability(rules, level); err != nil {
		t.Fatalf("level %d reachability: %v", levelNum, err)
	}
}

func TestRandomQuestRewardChestOpensAndDropsLoot(t *testing.T) {
	rules := loadRules(t)
	seed, levelNum := findRandomQuestRewardFloorForTest(t, rules.DungeonGeneration, rules.DungeonGeneration.BossFloor.Cadence-1)
	generated, err := GenerateDungeonLevel(seed, levelNum, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate target level: %v", err)
	}
	questChest := findGeneratedQuestRewardChest(generated)
	if questChest == nil {
		t.Fatalf("generated target level missing quest reward chest")
	}

	sim, err := NewSimWithWorld("sess_random_quest_reward_chest", seed, rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	for sim.currentLevel > levelNum {
		nextLevel := sim.currentLevel - 1
		results := descendFromCurrentLevel(t, sim, fmt.Sprintf("descend_to_%d", nextLevel))
		assertAckInResults(t, results, fmt.Sprintf("descend_to_%d", nextLevel))
	}

	var chest *entity
	for _, e := range sim.activeLevel().entities {
		if e.kind == interactableEntity && e.interactableDefID == treasureChestDefID && e.pos == questChest.pos {
			chest = e
			break
		}
	}
	if chest == nil {
		t.Fatalf("missing quest reward chest entity at %+v on level %d", questChest.pos, levelNum)
	}
	view := sim.entityView(chest)
	if !view.QuestReward {
		t.Fatalf("quest reward chest view missing QuestReward: %+v", view)
	}
	sim.activeLevel().entities[sim.playerID].pos = chest.pos
	open := sim.Tick([]Input{{MessageID: "open_quest_reward_chest", CorrelationID: "corr_random_quest_reward", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(chest.id)}}})
	assertAck(t, open, "open_quest_reward_chest")
	if !hasEvent(open, "interactable_activated") || !hasEvent(open, "loot_dropped") {
		t.Fatalf("open quest reward chest events = %+v", open.Events)
	}
	if chest.state != interactableOpen {
		t.Fatalf("quest reward chest state = %s, want open", chest.state)
	}
}

func findRandomQuestRewardFloorForTest(t *testing.T, rules DungeonGenerationRules, maxDepth int) (string, int) {
	t.Helper()
	for seedIndex := 0; seedIndex < 100; seedIndex++ {
		seed := fmt.Sprintf("v155_random_quest_%02d", seedIndex)
		for depth := 1; depth <= maxDepth; depth++ {
			levelNum := -depth
			if dungeonLevelHasRandomQuestReward(seed, levelNum, rules) {
				return seed, levelNum
			}
		}
	}
	t.Fatalf("no random quest reward floor found within test search depth %d", maxDepth)
	return "", 0
}

func findGeneratedQuestRewardChest(level generatedDungeonLevel) *generatedChest {
	for i := range level.chests {
		if level.chests[i].questReward {
			return &level.chests[i]
		}
	}
	return nil
}
