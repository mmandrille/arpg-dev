package game

import (
	"math"
	"sort"
	"strconv"
)

const characterMercenaryMonsterDefID = "character_mercenary"

// MercenaryCharacterSnapshot is durable roster data used to spawn character mercenaries.
type MercenaryCharacterSnapshot struct {
	CharacterID    string
	Name           string
	CharacterClass string
	Level          int
	Dead           bool
	Progression    CharacterProgressionState
	Items          []PersistedItem
}

// MercenaryCandidateView is one hire option on the mercenary board.
type MercenaryCandidateView struct {
	CharacterID    string `json:"character_id"`
	Name           string `json:"name"`
	CharacterClass string `json:"character_class"`
	Level          int    `json:"level"`
	Price          int    `json:"price"`
	Affordable     bool   `json:"affordable"`
}

// LoadMercenaryRoster replaces the session mercenary roster snapshot.
func (s *Sim) LoadMercenaryRoster(roster []MercenaryCharacterSnapshot) {
	s.mercenaryRoster = make(map[string]MercenaryCharacterSnapshot, len(roster))
	for _, snap := range roster {
		if snap.CharacterID == "" || snap.Dead {
			continue
		}
		if snap.Level < 1 {
			snap.Level = 1
		}
		if snap.Progression.Level < 1 {
			snap.Progression.Level = snap.Level
		}
		if snap.CharacterClass != "" {
			snap.Progression.CharacterClass = snap.CharacterClass
		}
		s.mercenaryRoster[snap.CharacterID] = snap
	}
}

func mercenaryEffectiveLevel(sourceLevel, activeLevel int) int {
	if sourceLevel < 1 {
		sourceLevel = 1
	}
	if activeLevel < 1 {
		activeLevel = 1
	}
	if sourceLevel <= activeLevel {
		return sourceLevel
	}

	return activeLevel
}

func scaleMercenaryFloat(value float64, sourceLevel, effectiveLevel int) float64 {
	if sourceLevel < 1 {
		sourceLevel = 1
	}
	if effectiveLevel >= sourceLevel {
		return value
	}

	return value * float64(effectiveLevel) / float64(sourceLevel)
}

func scaleMercenaryInt(value int, sourceLevel, effectiveLevel int) int {
	return int(math.Round(scaleMercenaryFloat(float64(value), sourceLevel, effectiveLevel)))
}

func equippedItemsFromPersisted(items []PersistedItem) map[string]*invItem {
	equipped := make(map[string]*invItem)
	for _, p := range items {
		if !p.Equipped || p.Slot == "" {
			continue
		}
		id, err := strconv.ParseUint(p.InstanceID, 10, 64)
		if err != nil {
			id = 1
		}
		equipped[p.Slot] = &invItem{
			instanceID:  id,
			itemDefID:   p.ItemDefID,
			slot:        p.Slot,
			equipped:    true,
			rollPayload: parseRollPayload(p.RolledStats),
		}
	}

	return equipped
}

func itemBaseAndRollStatsForItem(rules *Rules, item *invItem) (map[string]int, map[string]int) {
	baseStats := map[string]int{}
	rolledStats := map[string]int{}
	if item == nil || item.rollPayload == nil || rules == nil {
		return baseStats, rolledStats
	}
	template, ok := rules.ItemTemplates[item.rollPayload.ItemTemplateID]
	if !ok {
		return baseStats, rolledStats
	}
	for key, value := range template.BaseStats {
		baseStats[key] = value
	}
	for key, total := range item.rollPayload.Stats {
		if base := template.BaseStats[key]; total != base {
			rolledStats[key] = total - base
		}
	}

	return baseStats, rolledStats
}

func equipmentBaseStatBonusesForItems(rules *Rules, equipped map[string]*invItem) BaseStatsView {
	out := BaseStatsView{}
	if rules == nil {
		return out
	}
	for _, slot := range equipmentSlots {
		item := equipped[slot]
		if item == nil {
			continue
		}
		baseStats, rolledStats := itemBaseAndRollStatsForItem(rules, item)
		out.Str += baseStats["str"] + rolledStats["str"]
		out.Dex += baseStats["dex"] + rolledStats["dex"]
		out.Vit += baseStats["vit"] + rolledStats["vit"]
		out.Magic += baseStats["magic"] + rolledStats["magic"]
	}

	return out
}

func mercenaryDerivedStats(rules *Rules, progression CharacterProgressionState, equipped map[string]*invItem) DerivedStatsView {
	stats := progression.BaseStats
	equipment := equipmentBaseStatBonusesForItems(rules, equipped)
	stats.Str += equipment.Str
	stats.Dex += equipment.Dex
	stats.Vit += equipment.Vit
	stats.Magic += equipment.Magic
	eval := func(key string) float64 {
		formula := rules.CharacterProgression.DerivedStats[key]
		return evalProgressionFormula(formula, stats)
	}
	classLight := 0.0
	if classDef, ok := rules.CharacterProgression.Classes[progression.CharacterClass]; ok {
		classLight = classDef.LightRadius
	}

	return DerivedStatsView{
		DamageMin:            eval("damage_min"),
		DamageMax:            eval("damage_max"),
		Armor:                eval("armor"),
		BlockPercent:         0,
		AttackSpeed:          eval("attack_speed"),
		AttackIntervalTicks:  attackIntervalTicksFromSpeed(eval("attack_speed"), rules),
		HitChance:            eval("hit_chance"),
		CritChance:           eval("crit_chance"),
		CritDamage:           eval("crit_damage"),
		EvadeChance:          0,
		MovementSpeed:        eval("movement_speed"),
		MaxHP:                eval("max_hp"),
		MaxMana:              eval("max_mana"),
		HealthRegenPerSecond: eval("health_regen_per_second"),
		ManaRegenPerSecond:   eval("mana_regen_per_second"),
		MagicFindPercent:     0,
		LightRadius:          classLight + eval("light_radius"),
	}
}

func attackIntervalTicksFromSpeed(speed float64, rules *Rules) int {
	if rules == nil || speed <= 0 {
		return 1
	}
	interval := float64(rules.Combat.BaseAttackIntervalTicks) / speed
	if interval < 1 {
		return 1
	}

	return int(math.Round(interval))
}

func mercenaryCombatStats(rules *Rules, progression CharacterProgressionState, equipped map[string]*invItem) (effectiveCombatStats, float64) {
	character := mercenaryDerivedStats(rules, progression, equipped)
	damageMin := float64(rules.Combat.PlayerDamage.Min) + character.DamageMin
	damageMax := float64(rules.Combat.PlayerDamage.Max) + character.DamageMax
	armor := character.Armor
	maxHP := character.MaxHP
	blockPercent := 0.0
	weaponSpeed := 1.0
	hitChancePercent := character.HitChance * 100.0
	critChancePercent := character.CritChance * 100.0
	moveSpeedPercent := 0.0

	if weapon := equipped[mainHandSlot]; weapon != nil {
		baseMin, baseMax, minRoll, maxRoll, ok := weaponDamageContributionsForItem(rules, weapon)
		if ok {
			damageMin = character.DamageMin + baseMin + minRoll
			damageMax = character.DamageMax + baseMax + maxRoll
		}
		if speed, ok := weaponAttackSpeedContributionForItem(rules, weapon); ok {
			weaponSpeed = speed
		}
	}

	for _, slot := range equipmentSlots {
		item := equipped[slot]
		if item == nil {
			continue
		}
		baseStats, rolledStats := itemBaseAndRollStatsForItem(rules, item)
		if value := baseStats["armor"]; value != 0 {
			armor += float64(value)
		}
		if value := rolledStats["armor"]; value != 0 {
			armor += float64(value)
		}
		if value := baseStats["block_percent"]; value != 0 {
			blockPercent += float64(value)
		}
		if value := rolledStats["block_percent"]; value != 0 {
			blockPercent += float64(value)
		}
		if value := baseStats["hit_chance"]; value != 0 {
			hitChancePercent += float64(value)
		}
		if value := rolledStats["hit_chance"]; value != 0 {
			hitChancePercent += float64(value)
		}
		if value := baseStats["crit_chance"]; value != 0 {
			critChancePercent += float64(value)
		}
		if value := rolledStats["crit_chance"]; value != 0 {
			critChancePercent += float64(value)
		}
		if value := baseStats["movement_speed_percent"]; value != 0 {
			moveSpeedPercent += float64(value)
		}
		if value := rolledStats["movement_speed_percent"]; value != 0 {
			moveSpeedPercent += float64(value)
		}
	}

	attackSpeed := clampEffectiveAttackSpeed(character.AttackSpeed*weaponSpeed, rules)
	movementSpeed := character.MovementSpeed * (1.0 + moveSpeedPercent/100.0)
	if movementSpeed < 0 {
		movementSpeed = 0
	}

	return effectiveCombatStats{
		DamageMin:            damageMin,
		DamageMax:            damageMax,
		Armor:                armor,
		BlockPercent:         blockPercent,
		AttackSpeed:          attackSpeed,
		AttackIntervalTicks:  attackIntervalTicksFromSpeed(attackSpeed, rules),
		HitChance:            hitChancePercent / 100.0,
		CritChance:           critChancePercent / 100.0,
		EvadeChance:          0,
		MaxHP:                maxHP,
		MaxMana:              character.MaxMana,
		HealthRegenPerSecond: character.HealthRegenPerSecond,
		ManaRegenPerSecond:   character.ManaRegenPerSecond,
		MagicFindPercent:     0,
		LightRadius:          character.LightRadius,
		MovementSpeedPercent: moveSpeedPercent,
	}, movementSpeed
}

func weaponDamageContributionsForItem(rules *Rules, item *invItem) (baseMin, baseMax, minRoll, maxRoll float64, ok bool) {
	if item == nil || rules == nil {
		return 0, 0, 0, 0, false
	}
	if item.rollPayload != nil {
		template, found := rules.ItemTemplates[item.rollPayload.ItemTemplateID]
		if !found {
			return 0, 0, 0, 0, false
		}
		totalMin, minOK := item.rollPayload.Stats["damage_min"]
		totalMax, maxOK := item.rollPayload.Stats["damage_max"]
		if !minOK || !maxOK || totalMax < totalMin {
			return 0, 0, 0, 0, false
		}
		elementalBonus := elementalBonusDamage(item.rollPayload.Stats)
		baseMinInt := template.BaseStats["damage_min"]
		baseMaxInt := template.BaseStats["damage_max"]

		return float64(baseMinInt), float64(baseMaxInt),
			float64(totalMin-baseMinInt+elementalBonus), float64(totalMax-baseMaxInt+elementalBonus), true
	}
	def, found := rules.Items[item.itemDefID]
	if !found || def.Damage == nil {
		return 0, 0, 0, 0, false
	}

	return float64(def.Damage.Min), float64(def.Damage.Max), 0, 0, true
}

func weaponAttackSpeedContributionForItem(rules *Rules, item *invItem) (float64, bool) {
	if item == nil || rules == nil {
		return 0, false
	}
	if item.rollPayload != nil {
		template, ok := rules.ItemTemplates[item.rollPayload.ItemTemplateID]
		if !ok || template.AttackSpeed <= 0 {
			return 0, false
		}

		return template.AttackSpeed, true
	}
	def, found := rules.Items[item.itemDefID]
	if !found || def.AttackSpeed <= 0 {
		return 0, false
	}

	return def.AttackSpeed, true
}

func clampEffectiveAttackSpeed(speed float64, rules *Rules) float64 {
	if rules == nil {
		return speed
	}
	if speed < rules.Combat.MinEffectiveAttackSpeed {
		return rules.Combat.MinEffectiveAttackSpeed
	}
	if speed > rules.Combat.MaxEffectiveAttackSpeed {
		return rules.Combat.MaxEffectiveAttackSpeed
	}

	return speed
}

func mercenaryWeaponAttackMode(rules *Rules, equipped map[string]*invItem) string {
	weapon := equipped[mainHandSlot]
	if weapon == nil {
		return attackModeMelee
	}
	if weapon.rollPayload != nil {
		template, ok := rules.ItemTemplates[weapon.rollPayload.ItemTemplateID]
		if ok && template.AttackMode != "" {
			return template.AttackMode
		}

		return attackModeMelee
	}
	def, ok := rules.Items[weapon.itemDefID]
	if !ok || def.AttackMode == "" {
		return attackModeMelee
	}

	return def.AttackMode
}

func mercenaryWeaponReach(rules *Rules, equipped map[string]*invItem) float64 {
	weapon := equipped[mainHandSlot]
	if weapon == nil {
		return rules.Combat.UnarmedReach
	}
	if weapon.rollPayload != nil {
		template, ok := rules.ItemTemplates[weapon.rollPayload.ItemTemplateID]
		if ok && template.Reach > 0 {
			return template.Reach
		}
	} else if def, ok := rules.Items[weapon.itemDefID]; ok && def.Reach != nil {
		return *def.Reach
	}

	return rules.Combat.UnarmedReach
}

func mercenaryWeaponProjectile(rules *Rules, equipped map[string]*invItem) (projectileDefID string, speed float64, attackRange float64, ok bool) {
	weapon := equipped[mainHandSlot]
	if weapon == nil {
		return "", 0, 0, false
	}
	if weapon.rollPayload != nil {
		template, found := rules.ItemTemplates[weapon.rollPayload.ItemTemplateID]
		if !found || template.AttackMode != attackModeRanged || template.ProjectileSpeed <= 0 {
			return "", 0, 0, false
		}

		return "player_arrow", template.ProjectileSpeed, template.Reach, true
	}
	def, found := rules.Items[weapon.itemDefID]
	if !found || def.AttackMode != attackModeRanged || def.ProjectileSpeed == nil || *def.ProjectileSpeed <= 0 {
		return "", 0, 0, false
	}
	reach := rules.Combat.UnarmedReach
	if def.Reach != nil {
		reach = *def.Reach
	}

	return "player_arrow", *def.ProjectileSpeed, reach, true
}

func (s *Sim) mercenaryCandidatesForPlayer(player *entity) []MercenaryCandidateView {
	if player == nil {
		return nil
	}
	activeCharacterID := ""
	if ps := s.players[player.id]; ps != nil {
		activeCharacterID = ps.CharacterID
	}
	price := s.mercenaryHireCostGold()
	candidates := make([]MercenaryCandidateView, 0, len(s.mercenaryRoster))
	for _, characterID := range sortedStringKeys(s.mercenaryRoster) {
		snap := s.mercenaryRoster[characterID]
		if snap.Dead || snap.CharacterID == activeCharacterID {
			continue
		}
		candidates = append(candidates, MercenaryCandidateView{
			CharacterID:    snap.CharacterID,
			Name:           snap.Name,
			CharacterClass: snap.CharacterClass,
			Level:          snap.Level,
			Price:          price,
			Affordable:     s.gold >= price,
		})
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Name != candidates[j].Name {
			return candidates[i].Name < candidates[j].Name
		}

		return candidates[i].CharacterID < candidates[j].CharacterID
	})

	return candidates
}

func (s *Sim) spawnCharacterMercenary(owner *entity, snap MercenaryCharacterSnapshot, res *TickResult) *entity {
	if owner == nil {
		return nil
	}
	level := s.activeLevel()
	if level == nil {
		return nil
	}
	activeLevel := s.progression.Level
	if activeLevel < 1 {
		activeLevel = 1
	}
	effectiveLevel := mercenaryEffectiveLevel(snap.Level, activeLevel)
	equipped := equippedItemsFromPersisted(snap.Items)
	stats, movementSpeed := mercenaryCombatStats(s.rules, snap.Progression, equipped)
	stats.DamageMin = scaleMercenaryFloat(stats.DamageMin, snap.Level, effectiveLevel)
	stats.DamageMax = scaleMercenaryFloat(stats.DamageMax, snap.Level, effectiveLevel)
	stats.Armor = scaleMercenaryFloat(stats.Armor, snap.Level, effectiveLevel)
	stats.MaxHP = scaleMercenaryFloat(stats.MaxHP, snap.Level, effectiveLevel)
	stats.BlockPercent = scaleMercenaryFloat(stats.BlockPercent, snap.Level, effectiveLevel)
	stats.HitChance = scaleMercenaryFloat(stats.HitChance, snap.Level, effectiveLevel)
	stats.CritChance = scaleMercenaryFloat(stats.CritChance, snap.Level, effectiveLevel)

	maxHP := scaleMercenaryInt(int(math.Round(stats.MaxHP)), snap.Level, effectiveLevel)
	if maxHP < 1 {
		maxHP = 1
	}
	damage := DamageRange{
		Min: scaleMercenaryInt(int(math.Round(stats.DamageMin)), snap.Level, effectiveLevel),
		Max: scaleMercenaryInt(int(math.Round(stats.DamageMax)), snap.Level, effectiveLevel),
	}
	if damage.Max < damage.Min {
		damage.Max = damage.Min
	}
	if damage.Min < s.rules.Combat.MinimumDamage {
		damage.Min = s.rules.Combat.MinimumDamage
	}
	if damage.Max < s.rules.Combat.MinimumDamage {
		damage.Max = s.rules.Combat.MinimumDamage
	}

	attackMode := mercenaryWeaponAttackMode(s.rules, equipped)
	attackReach := mercenaryWeaponReach(s.rules, equipped)
	projectileDefID := ""
	projectileSpeed := 0.0
	attackRange := attackReach
	if attackMode == attackModeRanged {
		if defID, speed, reach, ok := mercenaryWeaponProjectile(s.rules, equipped); ok {
			projectileDefID = defID
			projectileSpeed = speed
			attackRange = reach
		} else {
			attackMode = attackModeMelee
		}
	}

	s.pruneCompanionsForNewSpawn(owner.id, mercenaryHireSourceID, 1, res)
	companion := &entity{
		kind:                     companionEntity,
		pos:                      s.companionSpawnPosition(owner),
		spawnPos:                 owner.pos,
		hp:                       maxHP,
		maxHP:                    maxHP,
		ownerID:                  owner.id,
		monsterDefID:             characterMercenaryMonsterDefID,
		characterID:              snap.CharacterID,
		displayName:              snap.Name,
		characterClass:           snap.CharacterClass,
		sourceCharacterID:        snap.CharacterID,
		speed:                    movementSpeed,
		monsterAttackDamage:      &damage,
		monsterAttackCooldown:    stats.AttackIntervalTicks,
		monsterHitChance:         stats.HitChance,
		monsterCritChance:        stats.CritChance,
		monsterBlockPercent:      stats.BlockPercent,
		monsterArmor:             stats.Armor,
		companionAttackMode:      attackMode,
		companionAttackReach:     attackReach,
		companionAttackRange:     attackRange,
		companionProjectileDefID: projectileDefID,
		companionProjectileSpeed: projectileSpeed,
		aiMode:                   monsterAIModeIdle,
		sourceSkillID:            mercenaryHireSourceID,
	}
	companion.id = s.alloc()
	level.entities[companion.id] = companion
	res.Changes = append(res.Changes, Change{Op: OpEntitySpawn, Entity: ptrEntityView(s.entityView(companion))})

	return companion
}

func (s *Sim) clearHiredMercenaryForOwner(ownerID uint64, res *TickResult) {
	if ownerID == 0 {
		return
	}
	if ps := s.players[ownerID]; ps != nil {
		s.usePlayer(ps)
	}
	if s.progression.HiredMercenaryCharacterID == "" {
		return
	}
	s.progression.HiredMercenaryCharacterID = ""
	if ps := s.players[ownerID]; ps != nil {
		s.savePlayer(ps)
	}
	if res != nil {
		s.appendCharacterProgressionUpdate(res)
	}
}

// RestoreHiredMercenaryCompanion respawns the durable hired mercenary for one
// player when a session starts or replay reconstructs without charging gold.
func (s *Sim) RestoreHiredMercenaryCompanion(ownerID uint64) {
	if s == nil || ownerID == 0 {
		return
	}
	owner := s.activeLevel().entities[ownerID]
	if owner == nil || owner.hp <= 0 {
		return
	}
	characterID := s.progression.HiredMercenaryCharacterID
	if ps := s.players[ownerID]; ps != nil && ps.Progression.HiredMercenaryCharacterID != "" {
		characterID = ps.Progression.HiredMercenaryCharacterID
	}
	if characterID == "" {
		return
	}
	if hiredMercenaryForOwner(s, ownerID) != nil {
		return
	}
	snap, ok := s.mercenaryRoster[characterID]
	if !ok || snap.Dead || snap.CharacterID == "" {
		s.clearHiredMercenaryForOwner(ownerID, nil)
		return
	}
	var res TickResult
	s.spawnCharacterMercenary(owner, snap, &res)
}

func hiredMercenaryForOwner(sim *Sim, ownerID uint64) *entity {
	for _, id := range sortedEntityIDs(sim.activeLevel().entities) {
		e := sim.activeLevel().entities[id]
		if e != nil && e.kind == companionEntity && e.ownerID == ownerID && e.sourceSkillID == mercenaryHireSourceID && e.hp > 0 {
			return e
		}
	}

	return nil
}
