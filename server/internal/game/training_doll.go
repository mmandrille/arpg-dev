package game

const townTrainingDollDefID = "town_training_doll"

func (s *Sim) syncTownTrainingDollsFromHost() {
	if !s.multiLevel {
		return
	}

	level := s.levels[townLevel]
	if level == nil {
		return
	}

	player := level.entities[s.playerID]
	if player == nil {
		return
	}

	stats, _ := s.playerEffectiveCombatStats()

	for _, id := range sortedEntityIDs(level.entities) {
		monster := level.entities[id]
		if monster == nil || !monster.isTrainingDoll {
			continue
		}

		monster.maxHP = player.maxHP
		monster.hp = player.maxHP
		monster.monsterArmor = stats.Armor
		monster.monsterBlockPercent = stats.BlockPercent
	}
}

func (s *Sim) isTrainingDoll(monster *entity) bool {
	return monster != nil && monster.isTrainingDoll
}

func (s *Sim) trainingDollReviveDelayTicks(monster *entity) int {
	if monster == nil || monster.monsterDefID == "" {
		return 30
	}

	def, ok := s.rules.Monsters[monster.monsterDefID]
	if !ok || def.ReviveDelayTicks <= 0 {
		return 30
	}

	return def.ReviveDelayTicks
}

func (s *Sim) finishTrainingDollDown(monster *entity, sourceID uint64, corr string, res *TickResult) {
	res.Events = append(res.Events, Event{
		EventType:      "monster_killed",
		EntityID:       idStr(monster.id),
		SourceEntityID: idStr(sourceID),
		TargetEntityID: idStr(monster.id),
		MonsterDefID:   monster.monsterDefID,
		CorrelationID:  corr,
	})
	monster.trainingDollReviveAt = s.tick + uint64(s.trainingDollReviveDelayTicks(monster))
}

func (s *Sim) tickTrainingDollRevives(res *TickResult) {
	if !s.multiLevel {
		return
	}

	level := s.levels[townLevel]
	if level == nil {
		return
	}

	for _, id := range sortedEntityIDs(level.entities) {
		monster := level.entities[id]
		if monster == nil || !monster.isTrainingDoll || monster.trainingDollReviveAt == 0 {
			continue
		}

		if s.tick < monster.trainingDollReviveAt {
			continue
		}

		monster.hp = monster.maxHP
		monster.trainingDollReviveAt = 0
		monster.pos = monster.spawnPos
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(monster))})
		hp := monster.hp
		res.Events = append(res.Events, Event{
			EventType:    "training_doll_revived",
			EntityID:     idStr(monster.id),
			MonsterDefID: monster.monsterDefID,
			Damage:       &hp,
		})
	}
}

func (s *Sim) appendMonsterCombatEvent(
	res *TickResult,
	eventType string,
	sourceID, targetID uint64,
	corr string,
	outcome combatResolution,
	breakdown []CombatBreakdownLineView,
	weaponSlot string,
	skillID string,
) {
	event := combatEvent(eventType, sourceID, targetID, corr, outcome)
	if len(breakdown) > 0 {
		event.DamageBreakdown = breakdown
	}

	if weaponSlot != "" {
		event.WeaponSlot = weaponSlot
	}

	if skillID != "" {
		event.SkillID = skillID
	}

	if target := s.activeLevel().entities[targetID]; target != nil && target.monsterDefID != "" {
		event.MonsterDefID = target.monsterDefID
	}

	res.Events = append(res.Events, event)
}

func (s *Sim) trainingDollBreakdownForBasicAttack(
	attacker effectiveCombatStats,
	defender effectiveCombatStats,
	damageRange DamageRange,
	outcome combatResolution,
	target *entity,
	damageType string,
	weaponSlot string,
) []CombatBreakdownLineView {
	if !s.isTrainingDoll(target) {
		return nil
	}

	resistance := s.monsterResistance(target, damageType)
	label := "Basic attack"
	if weaponSlot != "" {
		label = "Basic attack (" + weaponSlot + ")"
	}

	return buildCombatDamageBreakdown(attacker, defender, damageRange, outcome, resistance, damageType, label)
}

func (s *Sim) trainingDollBreakdownForSkill(
	defender effectiveCombatStats,
	damageRange DamageRange,
	outcome combatResolution,
	target *entity,
	damageType string,
	skillID string,
) []CombatBreakdownLineView {
	if !s.isTrainingDoll(target) {
		return nil
	}

	resistance := s.monsterResistance(target, damageType)
	label := skillID
	if def, ok := s.rules.Skills[skillID]; ok && def.Name != "" {
		label = def.Name
	}

	return buildSkillDamageBreakdown(defender, damageRange, outcome, resistance, damageType, label)
}
