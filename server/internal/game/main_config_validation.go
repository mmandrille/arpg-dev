package game

import "fmt"

func validateMainGameplayEconomyConfig(gameplay MainGameplayConfig) error {
	if gameplay.ItemUpgradeResourceCost < 0 {
		return fmt.Errorf("game: invalid rules main_config.gameplay.item_upgrade_resource_count: must be non-negative")
	}
	if gameplay.MercenaryHireCostGold < 0 {
		return fmt.Errorf("game: invalid rules main_config.gameplay.mercenary_hire_cost_gold: must be non-negative")
	}
	if gameplay.QuestTurnInItemDefID == "" {
		return fmt.Errorf("game: invalid rules main_config.gameplay.quest_turn_in_item_def_id: must be non-empty")
	}
	if gameplay.QuestTurnInRewardGold < 0 {
		return fmt.Errorf("game: invalid rules main_config.gameplay.quest_turn_in_reward_gold: must be non-negative")
	}
	return nil
}
