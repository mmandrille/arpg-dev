package game

import "fmt"

func validateMainGameplayEconomyConfig(gameplay MainGameplayConfig) error {
	if gameplay.ItemUpgradeResourceCost < 0 {
		return fmt.Errorf("game: invalid rules main_config.gameplay.item_upgrade_resource_count: must be non-negative")
	}
	if gameplay.MercenaryHireCostGold < 0 {
		return fmt.Errorf("game: invalid rules main_config.gameplay.mercenary_hire_cost_gold: must be non-negative")
	}
	return nil
}
