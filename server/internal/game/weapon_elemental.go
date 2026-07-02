package game

import "fmt"

func isElementalWeaponAffix(stat string) bool {
	switch stat {
	case "bonus_cold_damage", "bonus_fire_damage", "bonus_lightning_damage", "bonus_poison_damage":
		return true
	default:
		return false
	}
}

func filterOutElementalWeaponAffixes(stats []RollableStatDef) []RollableStatDef {
	if len(stats) == 0 {
		return stats
	}
	out := make([]RollableStatDef, 0, len(stats))
	for _, stat := range stats {
		if isElementalWeaponAffix(stat.Stat) {
			continue
		}
		out = append(out, stat)
	}

	return out
}

func rollAffixStatsOntoMap(stats map[string]int, rollableStats []RollableStatDef, rng *RNG, rollCount int) {
	pool := rollableStats
	for i := 0; i < rollCount; i++ {
		stat, ok := weightedRollableStat(pool, rng)
		if !ok {
			continue
		}
		stats[stat.Stat] += stat.Min + rng.IntN(stat.Max-stat.Min+1)
		if isElementalWeaponAffix(stat.Stat) {
			pool = filterOutElementalWeaponAffixes(pool)
		}
	}
}

func elementalAffixStatForDamageType(damageType string) string {
	switch canonicalDamageType(damageType) {
	case damageTypeCold:
		return "bonus_cold_damage"
	case damageTypeFire:
		return "bonus_fire_damage"
	case damageTypeLightning:
		return "bonus_lightning_damage"
	case damageTypePoison:
		return "bonus_poison_damage"
	default:
		return ""
	}
}

func elementalAffixDisplayName(stat string) string {
	switch stat {
	case "bonus_cold_damage":
		return "Cold"
	case "bonus_fire_damage":
		return "Fire"
	case "bonus_lightning_damage":
		return "Lightning"
	case "bonus_poison_damage":
		return "Poison"
	default:
		return ""
	}
}

func weaponElementalDamageFromItem(item *invItem) (int, string) {
	if item == nil || item.rollPayload == nil {
		return 0, damageTypeForce
	}
	damageType := dominantElementalDamageType(item.rollPayload.Stats)
	statKey := elementalAffixStatForDamageType(damageType)
	if statKey == "" {
		return 0, damageTypeForce
	}
	amount := item.rollPayload.Stats[statKey]
	if amount <= 0 {
		return 0, damageTypeForce
	}

	return amount, damageType
}

func elementalWeaponAffixOrder() []string {
	return []string{"bonus_cold_damage", "bonus_fire_damage", "bonus_lightning_damage", "bonus_poison_damage"}
}

func elementalStatSummaryLines(stats map[string]int) []string {
	if len(stats) == 0 {
		return nil
	}
	lines := []string{}
	for _, stat := range elementalWeaponAffixOrder() {
		if value := stats[stat]; value > 0 {
			if label := elementalAffixDisplayName(stat); label != "" {
				lines = append(lines, fmt.Sprintf("+%d %s Damage", value, label))
			}
		}
	}

	return lines
}

func physicalWeaponRollBonus(stats map[string]int, baseMin, baseMax int) (rollMin, rollMax int) {
	if stats == nil {
		return 0, 0
	}
	totalMin, minOK := stats["damage_min"]
	totalMax, maxOK := stats["damage_max"]
	if !minOK || !maxOK || totalMax < totalMin {
		return 0, 0
	}
	elementalBonus := elementalBonusDamage(stats)
	rollMin = totalMin - baseMin - elementalBonus
	rollMax = totalMax - baseMax - elementalBonus
	if rollMin < 0 {
		rollMin = 0
	}
	if rollMax < rollMin {
		rollMax = rollMin
	}

	return rollMin, rollMax
}

func (s *Sim) weaponItemForSlot(slot string) *invItem {
	if slot == "" {
		slot = mainHandSlot
	}

	return s.findItemByID(s.equipped[slot])
}

func (s *Sim) applyWeaponElementalDamageFromSlot(target *entity, playerID uint64, corr string, weaponSlot string, physicalHitDamage int, res *TickResult) {
	if target == nil || target.hp <= 0 {
		return
	}
	item := s.weaponItemForSlot(weaponSlot)
	s.applyWeaponElementalDamageWithItem(target, playerID, corr, item, physicalHitDamage, res, weaponSlot)
}

func (s *Sim) applyWeaponElementalDamageWithItem(target *entity, playerID uint64, corr string, item *invItem, physicalHitDamage int, res *TickResult, weaponSlot ...string) {
	if target == nil || target.hp <= 0 {
		return
	}
	amount, damageType := weaponElementalDamageFromItem(item)
	if amount <= 0 || damageType == damageTypeForce {
		return
	}
	outcome := combatResolution{
		Hit:             true,
		Outcome:         "hit",
		Damage:          amount,
		MitigatedDamage: amount,
		RawDamage:       amount,
	}
	s.applyMonsterResistanceToOutcome(target, damageType, &outcome)
	if outcome.Damage <= 0 {
		return
	}
	target.hp -= outcome.Damage
	if target.hp < 0 {
		target.hp = 0
	}
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
	event := combatEvent(s.combatEventType(monsterEntity, outcome), playerID, target.id, corr, outcome)
	if len(weaponSlot) > 0 && weaponSlot[0] != "" {
		event.WeaponSlot = weaponSlot[0]
	}
	res.Events = append(res.Events, event)
	if outcome.Damage > 0 {
		s.tryPassiveExecute(target, playerID, corr, res)
	}
	totalHit := physicalHitDamage + outcome.Damage
	s.tryWeaponElementalProcs(target, playerID, corr, damageType, outcome.Damage, totalHit, res)
	if target.hp == 0 {
		s.finishMonsterKill(target, playerID, corr, res)
	}
}

func (s *Sim) lootPayloadFromWorldPreset(preset *WorldLootPreset) (ItemRollPayload, bool) {
	if preset == nil || preset.ItemTemplateID == "" || len(preset.Stats) == 0 {
		return ItemRollPayload{}, false
	}
	template, ok := s.rules.ItemTemplates[preset.ItemTemplateID]
	if !ok {
		return ItemRollPayload{}, false
	}
	rarity := preset.Rarity
	if rarity == "" {
		rarity = "magic"
	}
	itemLevel := preset.ItemLevel
	if itemLevel < 1 {
		itemLevel = 1
	}
	stats := cloneIntMap(preset.Stats)
	displayName := preset.DisplayName
	if displayName == "" {
		displayName = s.rules.affixDisplayName(template, rarity, stats)
	}

	return ItemRollPayload{
		ItemTemplateID: preset.ItemTemplateID,
		DisplayName:    displayName,
		Rarity:         rarity,
		ItemLevel:      itemLevel,
		Stats:          stats,
		Requirements:   cloneIntMap(template.Requirements),
	}, true
}

func validateWorldLootEntity(r *Rules, label string, entity WorldEntity) error {
	sourceCount := 0
	if entity.ItemDefID != "" {
		sourceCount++
	}
	if entity.ItemTemplateID != "" {
		sourceCount++
	}
	if entity.LootPreset != nil {
		sourceCount++
	}
	if sourceCount != 1 {
		return fmt.Errorf("game: invalid rules %s: declare exactly one of item_def_id, item_template_id, or loot_preset", label)
	}
	if entity.ItemDefID != "" {
		if _, ok := r.Items[entity.ItemDefID]; !ok {
			return fmt.Errorf("game: invalid rules %s: unknown item %s", label, entity.ItemDefID)
		}
	}
	if entity.ItemTemplateID != "" {
		if _, ok := r.ItemTemplates[entity.ItemTemplateID]; !ok {
			return fmt.Errorf("game: invalid rules %s: unknown item template %s", label, entity.ItemTemplateID)
		}
	}
	if entity.LootPreset != nil {
		if _, ok := r.ItemTemplates[entity.LootPreset.ItemTemplateID]; !ok {
			return fmt.Errorf("game: invalid rules %s: unknown loot preset template %s", label, entity.LootPreset.ItemTemplateID)
		}
		if len(entity.LootPreset.Stats) == 0 {
			return fmt.Errorf("game: invalid rules %s: loot_preset.stats required", label)
		}
	}

	return nil
}
