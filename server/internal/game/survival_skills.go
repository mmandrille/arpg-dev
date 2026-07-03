package game

import "fmt"

func (s *Sim) survivalSkillForClass() (string, SkillDef, bool) {
	classID := s.progression.CharacterClass
	for _, skillID := range sortedStringKeys(s.rules.Skills) {
		def := s.rules.Skills[skillID]
		if def.Kind == "survival_autocast" && def.Class == classID {
			return skillID, def, true
		}
	}

	return "", SkillDef{}, false
}

func survivalProcAllowed(attacker *entity) bool {
	if attacker == nil {
		return true
	}

	return attacker.kind != playerEntity
}

func (s *Sim) adjustPlayerIncomingDamage(player *entity, attacker *entity, outcome *combatResolution, corr string, res *TickResult) {
	if player == nil || player.kind != playerEntity || player.hp <= 0 || outcome == nil || outcome.Damage <= 0 {
		return
	}
	if s.applyPlayerManaShield(player, outcome) {
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	}
	if s.applyPlayerDamageRedirect(player, attacker, outcome, corr, res) {
		return
	}
	if outcome.Damage >= player.hp {
		s.trySurvivalAutocast(player, attacker, outcome, corr, res)
	}
}

func (s *Sim) trySurvivalAutocast(player *entity, attacker *entity, outcome *combatResolution, corr string, res *TickResult) {
	if player == nil || outcome == nil || outcome.Damage < player.hp || player.hp <= 0 {
		return
	}
	if !survivalProcAllowed(attacker) {
		return
	}
	skillID, def, ok := s.survivalSkillForClass()
	if !ok {
		return
	}
	rank := s.effectiveSkillRank(skillID)
	if rank <= 0 {
		return
	}
	if remaining, onCooldown := s.skillCooldownRemaining(skillID); onCooldown {
		_ = remaining
		return
	}
	outcome.Damage = player.hp - 1
	if outcome.Damage < 0 {
		outcome.Damage = 0
	}
	s.activateSurvivalSkill(player, attacker, skillID, def, rank, corr, res)
	cooldownTicks := s.skillCooldownTicks(def)
	s.skillCooldowns[skillID] = skillCooldownState{EndsTick: s.tick + uint64(cooldownTicks), TotalTicks: cooldownTicks}
	s.appendSkillCooldownUpdate(res)
	s.appendSkillCooldownStartedEvent(res, player, skillID, corr, cooldownTicks)
}

func (s *Sim) activateSurvivalSkill(player *entity, attacker *entity, skillID string, def SkillDef, rank int, corr string, res *TickResult) {
	for idx, effect := range def.Effects {
		switch effect.Type {
		case "survival_vit_regen":
			s.applySurvivalVitRegen(player, skillID, effect, rank, corr, res)
		case "survival_mana_shield":
			s.applySurvivalManaShield(player, skillID, effect, rank, corr, res)
		case "survival_immunity_damage":
			s.applySurvivalImmunityDamage(player, skillID, effect, rank, corr, res)
		case "survival_evasive_stance":
			s.applySurvivalEvasiveStance(player, skillID, effect, rank, corr, res)
		case "survival_spectral_path":
			s.applySurvivalSpectralPath(player, skillID, effect, rank, corr, res)
		default:
			continue
		}
		_ = idx
	}
	s.appendCharacterProgressionUpdate(res)
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
}

func (s *Sim) applySurvivalVitRegen(player *entity, skillID string, effect SkillEffectDef, rank int, corr string, res *TickResult) {
	totalTicks := effect.DurationTicks
	if totalTicks <= 0 {
		return
	}
	percent := skillEffectPercent(effect, rank)
	regenMultiplier := float64(effect.HealthRegenMultiplierPercent) / 100.0
	if regenMultiplier <= 0 {
		regenMultiplier = 1
	}
	stateKey := fmt.Sprintf("%s:%d", skillID, player.id)
	s.skillEffects[stateKey] = skillEffectState{
		SkillID:               skillID,
		TargetID:              player.id,
		Stats:                 []string{"vit"},
		Percent:               percent,
		HealthRegenMultiplier: regenMultiplier,
		EffectID:              "second_wind",
		EndsTick:              s.tick + uint64(totalTicks),
		TotalTicks:            totalTicks,
	}
	player.effectIDs = sortedUniqueStrings(append(player.effectIDs, "second_wind"))
	s.emitSurvivalEffectStarted(player, skillID, rank, corr, res, totalTicks, percent)
}

func (s *Sim) applySurvivalManaShield(player *entity, skillID string, effect SkillEffectDef, rank int, corr string, res *TickResult) {
	totalTicks := effect.DurationTicks
	manaPerHP := effect.ManaPerHPBase - effect.ManaPerHPPerRank*(rank-1)
	if totalTicks <= 0 || manaPerHP <= 0 {
		return
	}
	stateKey := fmt.Sprintf("%s:%d", skillID, player.id)
	s.skillEffects[stateKey] = skillEffectState{
		SkillID:    skillID,
		TargetID:   player.id,
		ManaPerHP:  manaPerHP,
		EffectID:   "arcane_barrier",
		EndsTick:   s.tick + uint64(totalTicks),
		TotalTicks: totalTicks,
	}
	player.effectIDs = sortedUniqueStrings(append(player.effectIDs, "arcane_barrier"))
	s.emitSurvivalEffectStarted(player, skillID, rank, corr, res, totalTicks, manaPerHP)
}

func (s *Sim) applySurvivalImmunityDamage(player *entity, skillID string, effect SkillEffectDef, rank int, corr string, res *TickResult) {
	totalTicks := effect.DurationTicks
	effectID := effect.EffectID
	if effectID == "" {
		effectID = "divine_protection"
	}
	outgoing := effect.OutgoingDamagePercent
	if outgoing <= 0 {
		outgoing = 500
	}
	if totalTicks <= 0 {
		return
	}
	stateKey := fmt.Sprintf("%s:%d", skillID, player.id)
	s.skillEffects[stateKey] = skillEffectState{
		SkillID:               skillID,
		TargetID:              player.id,
		Immunity:              true,
		OutgoingDamagePercent: outgoing,
		EffectID:              effectID,
		EndsTick:              s.tick + uint64(totalTicks),
		TotalTicks:            totalTicks,
	}
	player.effectIDs = sortedUniqueStrings(append(player.effectIDs, effectID))
	s.emitSurvivalEffectStarted(player, skillID, rank, corr, res, totalTicks, outgoing)
}

func (s *Sim) applySurvivalEvasiveStance(player *entity, skillID string, effect SkillEffectDef, rank int, corr string, res *TickResult) {
	totalTicks := effect.DurationTicks
	if totalTicks <= 0 {
		return
	}
	if effect.CleanseDebuffs {
		s.cleansePlayerDebuffs(player, skillID)
	}
	evadePercent := effect.EvadePercent
	if evadePercent <= 0 {
		evadePercent = 100
	}
	effectID := effect.EffectID
	if effectID == "" {
		effectID = "evasive_stance"
	}
	stateKey := fmt.Sprintf("%s:%d", skillID, player.id)
	s.skillEffects[stateKey] = skillEffectState{
		SkillID:    skillID,
		TargetID:   player.id,
		ForceEvade: true,
		EffectID:   effectID,
		EndsTick:   s.tick + uint64(totalTicks),
		TotalTicks: totalTicks,
	}
	player.effectIDs = sortedUniqueStrings(append(player.effectIDs, effectID))
	s.markNearbyMonsters(player, skillID, effect, rank, corr, res)
	s.emitSurvivalEffectStarted(player, skillID, rank, corr, res, totalTicks, evadePercent)
}

func (s *Sim) applySurvivalSpectralPath(player *entity, skillID string, effect SkillEffectDef, rank int, corr string, res *TickResult) {
	totalTicks := effect.DurationTicks
	if totalTicks <= 0 {
		return
	}
	effectID := effect.EffectID
	if effectID == "" {
		effectID = "spectral_path"
	}
	stateKey := fmt.Sprintf("%s:%d", skillID, player.id)
	s.skillEffects[stateKey] = skillEffectState{
		SkillID:              skillID,
		TargetID:             player.id,
		RedirectDamage:       effect.RedirectDamage,
		PhaseThroughMonsters: effect.PhaseThroughMonsters,
		EffectID:             effectID,
		EndsTick:             s.tick + uint64(totalTicks),
		TotalTicks:           totalTicks,
	}
	player.effectIDs = sortedUniqueStrings(append(player.effectIDs, effectID))
	s.emitSurvivalEffectStarted(player, skillID, rank, corr, res, totalTicks, 0)
}

func (s *Sim) emitSurvivalEffectStarted(player *entity, skillID string, rank int, corr string, res *TickResult, totalTicks int, amount int) {
	res.Events = append(res.Events, Event{
		EventType:      "skill_effect_started",
		EntityID:       idStr(player.id),
		TargetEntityID: idStr(player.id),
		CorrelationID:  corr,
		SkillID:        skillID,
		Rank:           intPtr(rank),
		Amount:         intPtr(amount),
		RemainingTicks: intPtr(totalTicks),
		TotalTicks:     intPtr(totalTicks),
	})
}

func (s *Sim) cleansePlayerDebuffs(player *entity, keepSkillID string) {
	if player == nil {
		return
	}
	for _, stateKey := range sortedStringKeys(s.skillEffects) {
		effect := s.skillEffects[stateKey]
		if effect.TargetID != player.id || effect.SkillID == keepSkillID {
			continue
		}
		delete(s.skillEffects, stateKey)
	}
	delete(s.poisonDots, player.id)
	player.effectIDs = []string{}
}

func (s *Sim) markNearbyMonsters(player *entity, skillID string, effect SkillEffectDef, rank int, corr string, res *TickResult) {
	if player == nil || effect.MarkRadius <= 0 {
		return
	}
	mark := SkillMarkDef{
		DamageBonusPercent:        effect.MarkDamageBonusPercent,
		DamageBonusPercentPerRank: effect.MarkDamageBonusPercentPerRank,
		DurationTicks:             effect.MarkDurationTicks,
		EffectID:                  effect.MarkEffectID,
	}
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		target := s.activeLevel().entities[id]
		if target == nil || target.kind != monsterEntity || target.hp <= 0 {
			continue
		}
		if distance(player.pos, target.pos) > effect.MarkRadius {
			continue
		}
		s.startSkillMark(player, target, skillID, mark, rank, corr, res)
	}
}

func (s *Sim) applyPlayerManaShield(player *entity, outcome *combatResolution) bool {
	if player == nil || outcome == nil || outcome.Damage <= 0 {
		return false
	}
	manaPerHP, skillID, ok := s.activePlayerManaPerHP(player)
	if !ok || manaPerHP <= 0 {
		return false
	}
	manaCost := outcome.Damage * manaPerHP
	if player.mana >= manaCost {
		player.mana -= manaCost
		outcome.MitigatedDamage += outcome.Damage
		outcome.Damage = 0
		return true
	}
	if player.mana <= 0 {
		return false
	}
	absorbedHP := player.mana / manaPerHP
	if absorbedHP <= 0 {
		return false
	}
	if absorbedHP > outcome.Damage {
		absorbedHP = outcome.Damage
	}
	player.mana -= absorbedHP * manaPerHP
	outcome.Damage -= absorbedHP
	outcome.MitigatedDamage += absorbedHP
	_ = skillID
	return true
}

func (s *Sim) activePlayerManaPerHP(player *entity) (int, string, bool) {
	for _, stateKey := range sortedStringKeys(s.skillEffects) {
		effect := s.skillEffects[stateKey]
		if effect.TargetID != player.id || effect.ManaPerHP <= 0 || effect.EndsTick <= s.tick {
			continue
		}
		return effect.ManaPerHP, effect.SkillID, true
	}

	return 0, "", false
}

func (s *Sim) applyPlayerDamageRedirect(player *entity, attacker *entity, outcome *combatResolution, corr string, res *TickResult) bool {
	if player == nil || attacker == nil || attacker.kind != monsterEntity || attacker.hp <= 0 || outcome == nil || outcome.Damage <= 0 {
		return false
	}
	for _, stateKey := range sortedStringKeys(s.skillEffects) {
		effect := s.skillEffects[stateKey]
		if effect.TargetID != player.id || !effect.RedirectDamage || effect.EndsTick <= s.tick {
			continue
		}
		redirect := outcome.Damage
		s.applyUniqueDirectDamage(player.id, attacker, effect.SkillID, redirect, damageTypeForce, corr, res)
		outcome.MitigatedDamage += outcome.Damage
		outcome.Damage = 0
		return true
	}

	return false
}

func (s *Sim) playerPhasingThroughMonsters() bool {
	player := s.activeLevel().entities[s.playerID]
	if player == nil {
		return false
	}
	for _, stateKey := range sortedStringKeys(s.skillEffects) {
		effect := s.skillEffects[stateKey]
		if effect.TargetID == player.id && effect.PhaseThroughMonsters && effect.EndsTick > s.tick {
			return true
		}
	}

	return false
}

func (s *Sim) playerForceEvadeActive(player *entity) bool {
	if player == nil {
		return false
	}
	for _, stateKey := range sortedStringKeys(s.skillEffects) {
		effect := s.skillEffects[stateKey]
		if effect.TargetID == player.id && effect.ForceEvade && effect.EndsTick > s.tick {
			return true
		}
	}

	return false
}

func (s *Sim) playerOutgoingDamageMultiplier() float64 {
	player := s.activeLevel().entities[s.playerID]
	if player == nil {
		return 1
	}
	multiplier := 1.0
	for _, stateKey := range sortedStringKeys(s.skillEffects) {
		effect := s.skillEffects[stateKey]
		if effect.TargetID != player.id || effect.OutgoingDamagePercent <= 0 || effect.EndsTick <= s.tick {
			continue
		}
		multiplier = float64(effect.OutgoingDamagePercent) / 100.0
	}

	return multiplier
}

func (s *Sim) applyPlayerOutgoingDamageMultiplier(outcome *combatResolution) {
	if outcome == nil || outcome.Damage <= 0 {
		return
	}
	multiplier := s.playerOutgoingDamageMultiplier()
	if multiplier <= 1 {
		return
	}
	outcome.Damage = int(float64(outcome.Damage) * multiplier)
	if outcome.RawDamage > 0 {
		outcome.RawDamage = int(float64(outcome.RawDamage) * multiplier)
	}
}

func (s *Sim) playerHealthRegenMultiplier() float64 {
	player := s.activeLevel().entities[s.playerID]
	if player == nil {
		return 1
	}
	multiplier := 1.0
	for _, stateKey := range sortedStringKeys(s.skillEffects) {
		effect := s.skillEffects[stateKey]
		if effect.TargetID != player.id || effect.HealthRegenMultiplier <= 0 || effect.EndsTick <= s.tick {
			continue
		}
		if effect.HealthRegenMultiplier > multiplier {
			multiplier = effect.HealthRegenMultiplier
		}
	}

	return multiplier
}

func validateSurvivalAutocastSkill(skillID string, skill SkillDef) error {
	if skill.Targeting != "self" {
		return fmt.Errorf("game: invalid rules skills.%s.targeting: survival_autocast must target self", skillID)
	}
	if skill.Cost.Mana.Base != 0 || skill.Cost.Mana.PerRank != 0 {
		return fmt.Errorf("game: invalid rules skills.%s.cost.mana: survival_autocast must be mana-free", skillID)
	}
	if skill.Requirements.Level < 10 {
		return fmt.Errorf("game: invalid rules skills.%s.requirements.level: survival_autocast requires level 10", skillID)
	}
	if skill.Tree.Branch != "survival" {
		return fmt.Errorf("game: invalid rules skills.%s.tree.branch: survival_autocast requires branch survival", skillID)
	}
	if len(skill.Effects) == 0 {
		return fmt.Errorf("game: invalid rules skills.%s.effects: survival_autocast requires at least one effect", skillID)
	}
	for idx, effect := range skill.Effects {
		if err := validateSurvivalEffect(skillID, idx, effect); err != nil {
			return err
		}
	}
	cooldownTicks := skill.Cooldown.FixedTicks
	if cooldownTicks <= 0 {
		return fmt.Errorf("game: invalid rules skills.%s.cooldown.fixed_ticks: survival_autocast requires fixed_ticks cooldown", skillID)
	}

	return nil
}

func validateSurvivalEffect(skillID string, idx int, effect SkillEffectDef) error {
	switch effect.Type {
	case "survival_vit_regen":
		if effect.DurationTicks <= 0 || effect.PercentBase <= 0 || effect.HealthRegenMultiplierPercent <= 0 {
			return fmt.Errorf("game: invalid rules skills.%s.effects[%d]: survival_vit_regen requires duration, vit percent, and regen multiplier", skillID, idx)
		}
	case "survival_mana_shield":
		if effect.DurationTicks <= 0 || effect.ManaPerHPBase <= 0 {
			return fmt.Errorf("game: invalid rules skills.%s.effects[%d]: survival_mana_shield requires duration and mana_per_hp_base", skillID, idx)
		}
	case "survival_immunity_damage":
		if effect.DurationTicks <= 0 || effect.OutgoingDamagePercent <= 0 {
			return fmt.Errorf("game: invalid rules skills.%s.effects[%d]: survival_immunity_damage requires duration and outgoing_damage_percent", skillID, idx)
		}
	case "survival_evasive_stance":
		if effect.DurationTicks <= 0 || effect.EvadePercent <= 0 {
			return fmt.Errorf("game: invalid rules skills.%s.effects[%d]: survival_evasive_stance requires duration and evade_percent", skillID, idx)
		}
	case "survival_spectral_path":
		if effect.DurationTicks <= 0 || !effect.RedirectDamage || !effect.PhaseThroughMonsters {
			return fmt.Errorf("game: invalid rules skills.%s.effects[%d]: survival_spectral_path requires duration, redirect, and phasing", skillID, idx)
		}
	default:
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d].type: unsupported survival effect %q", skillID, idx, effect.Type)
	}

	return nil
}
