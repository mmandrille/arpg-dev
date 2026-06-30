package game

import "fmt"

func validateMainGameplayEconomyConfig(gameplay MainGameplayConfig) error {
	if gameplay.ItemUpgradeResourceCost < 0 {
		return fmt.Errorf("game: invalid rules main_config.gameplay.item_upgrade_resource_count: must be non-negative")
	}
	if err := validateResourceLootDropsConfig(gameplay.ResourceLootDrops); err != nil {
		return err
	}
	if gameplay.BishopRespecResourceCost < 0 {
		return fmt.Errorf("game: invalid rules main_config.gameplay.bishop_respec_resource_count: must be non-negative")
	}
	if gameplay.BishopRespecResourceCost > 0 && gameplay.BishopRespecResourceID == "" {
		return fmt.Errorf("game: invalid rules main_config.gameplay.bishop_respec_resource_item_def_id: required when count is positive")
	}
	if gameplay.BishopReviveResourceCost < 0 {
		return fmt.Errorf("game: invalid rules main_config.gameplay.bishop_revive_resource_count: must be non-negative")
	}
	if gameplay.BishopReviveResourceCost > 0 && gameplay.BishopReviveResourceID == "" {
		return fmt.Errorf("game: invalid rules main_config.gameplay.bishop_revive_resource_item_def_id: required when count is positive")
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
	if len(gameplay.BadgeRewardRules) == 0 {
		return fmt.Errorf("game: invalid rules main_config.gameplay.badge_reward_rules: must not be empty")
	}
	seen := map[string]bool{}
	for idx, rule := range gameplay.BadgeRewardRules {
		label := fmt.Sprintf("main_config.gameplay.badge_reward_rules[%d]", idx)
		if rule.ResourceItemDefID == "" {
			return fmt.Errorf("game: invalid rules %s.resource_item_def_id: must be non-empty", label)
		}
		if seen[rule.ResourceItemDefID] {
			return fmt.Errorf("game: invalid rules %s.resource_item_def_id: duplicate resource %q", label, rule.ResourceItemDefID)
		}
		seen[rule.ResourceItemDefID] = true
		if rule.UnlockDepth <= 0 {
			return fmt.Errorf("game: invalid rules %s.unlock_depth: must be positive", label)
		}
		if rule.BaseChancePercent < 0 || rule.BaseChancePercent > 100 {
			return fmt.Errorf("game: invalid rules %s.base_chance_percent: must be within [0,100]", label)
		}
		if rule.ChancePerDepthPercent < 0 {
			return fmt.Errorf("game: invalid rules %s.chance_per_depth_percent: must be non-negative", label)
		}
	}
	return nil
}

func validateResourceLootDropsConfig(cfg ResourceLootDropsConfig) error {
	chanceFields := []struct {
		label string
		value int
	}{
		{"monster_common_rare_chance_percent", cfg.MonsterCommonRareChancePercent},
		{"monster_champion_chance_percent", cfg.MonsterChampionChancePercent},
		{"monster_unique_chance_percent", cfg.MonsterUniqueChancePercent},
		{"boss_kill_chance_percent", cfg.BossKillChancePercent},
		{"chest_regular_chance_percent", cfg.ChestRegularChancePercent},
		{"chest_boss_chance_percent", cfg.ChestBossChancePercent},
	}
	for _, field := range chanceFields {
		if field.value < 0 || field.value > 100 {
			return fmt.Errorf("game: invalid rules main_config.gameplay.resource_loot_drops.%s: must be within [0,100]", field.label)
		}
	}
	if len(cfg.Pool) == 0 {
		return fmt.Errorf("game: invalid rules main_config.gameplay.resource_loot_drops.pool: must not be empty")
	}
	totalWeight := 0
	for idx, entry := range cfg.Pool {
		if entry.ItemDefID == "" {
			return fmt.Errorf("game: invalid rules main_config.gameplay.resource_loot_drops.pool[%d].item_def_id: must be non-empty", idx)
		}
		if entry.Weight <= 0 {
			return fmt.Errorf("game: invalid rules main_config.gameplay.resource_loot_drops.pool[%d].weight: must be positive", idx)
		}
		totalWeight += entry.Weight
	}
	if totalWeight <= 0 {
		return fmt.Errorf("game: invalid rules main_config.gameplay.resource_loot_drops.pool: total weight must be positive")
	}

	return nil
}

func validateMainGameplayResourceItems(gameplay MainGameplayConfig, items map[string]ItemDef) error {
	if gameplay.ItemUpgradeResourceCost > 0 {
		if _, ok := items[gameplay.ItemUpgradeResourceID]; !ok {
			return fmt.Errorf("game: invalid rules main_config.gameplay.item_upgrade_resource_item_def_id: unknown item %q", gameplay.ItemUpgradeResourceID)
		}
	}
	if err := validateCurrencyResourceItem(items, gameplay.BishopRespecResourceID, gameplay.BishopRespecResourceCost, "bishop_respec_resource_item_def_id"); err != nil {
		return err
	}
	if err := validateCurrencyResourceItem(items, gameplay.BishopReviveResourceID, gameplay.BishopReviveResourceCost, "bishop_revive_resource_item_def_id"); err != nil {
		return err
	}
	turnInItem, ok := items[gameplay.QuestTurnInItemDefID]
	if !ok {
		return fmt.Errorf("game: invalid rules main_config.gameplay.quest_turn_in_item_def_id: unknown item %q", gameplay.QuestTurnInItemDefID)
	}
	if turnInItem.Category != "quest" {
		return fmt.Errorf("game: invalid rules main_config.gameplay.quest_turn_in_item_def_id: item %q must be category quest", gameplay.QuestTurnInItemDefID)
	}
	for idx, entry := range gameplay.ResourceLootDrops.Pool {
		if _, ok := items[entry.ItemDefID]; !ok {
			return fmt.Errorf("game: invalid rules main_config.gameplay.resource_loot_drops.pool[%d].item_def_id: unknown item %q", idx, entry.ItemDefID)
		}
	}
	for idx, rule := range gameplay.BadgeRewardRules {
		item, ok := items[rule.ResourceItemDefID]
		if !ok {
			return fmt.Errorf("game: invalid rules main_config.gameplay.badge_reward_rules[%d].resource_item_def_id: unknown item %q", idx, rule.ResourceItemDefID)
		}
		if rule.ResourceItemDefID == gameplay.ItemUpgradeResourceID {
			continue
		}
		if item.Category != "currency" || item.Equippable {
			return fmt.Errorf("game: invalid rules main_config.gameplay.badge_reward_rules[%d].resource_item_def_id: item %q must be a non-equippable currency", idx, rule.ResourceItemDefID)
		}
	}
	return nil
}

func validateCurrencyResourceItem(items map[string]ItemDef, itemDefID string, count int, field string) error {
	if count <= 0 {
		return nil
	}
	item, ok := items[itemDefID]
	if !ok {
		return fmt.Errorf("game: invalid rules main_config.gameplay.%s: unknown item %q", field, itemDefID)
	}
	if item.Category != "currency" || item.Equippable {
		return fmt.Errorf("game: invalid rules main_config.gameplay.%s: item %q must be a non-equippable currency", field, itemDefID)
	}
	return nil
}
