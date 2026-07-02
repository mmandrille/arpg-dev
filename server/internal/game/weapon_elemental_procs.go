package game

import (
	"fmt"
	"math"
)

const (
	weaponElementalFreezeSkillID = "weapon_elemental_freeze"
	weaponElementalBurnSkillID   = "weapon_elemental_burn"
	weaponElementalPoisonSkillID = "weapon_elemental_poison"
	weaponElementalStunSkillID   = "weapon_elemental_stun"
)

type WeaponElementalProcsConfig struct {
	Cold      WeaponElementalColdProcConfig      `json:"cold"`
	Fire      WeaponElementalFireProcConfig      `json:"fire"`
	Lightning WeaponElementalLightningProcConfig `json:"lightning"`
	Poison    WeaponElementalPoisonProcConfig    `json:"poison"`
}

type WeaponElementalColdProcConfig struct {
	ProcChancePercent int    `json:"proc_chance_percent"`
	SlowPercent       int    `json:"slow_percent"`
	DurationSeconds   int    `json:"duration_seconds"`
	EffectID          string `json:"effect_id"`
}

type WeaponElementalFireProcConfig struct {
	ProcChancePercent      int    `json:"proc_chance_percent"`
	BurnPercentOfTotalHit  int    `json:"burn_percent_of_total_hit"`
	DurationSeconds        int    `json:"duration_seconds"`
	TickIntervalSeconds    int    `json:"tick_interval_seconds"`
	EffectID               string `json:"effect_id"`
}

type WeaponElementalLightningProcConfig struct {
	ProcChancePercent   int    `json:"proc_chance_percent"`
	StunDurationSeconds int    `json:"stun_duration_seconds"`
	EffectID            string `json:"effect_id"`
}

type WeaponElementalPoisonProcConfig struct {
	ProcChancePercent            int    `json:"proc_chance_percent"`
	DotPercentOfElementalDamage  int    `json:"dot_percent_of_elemental_damage"`
	DurationSeconds              int    `json:"duration_seconds"`
	TickIntervalSeconds          int    `json:"tick_interval_seconds"`
	EffectID                     string `json:"effect_id"`
}

func validateWeaponElementalProcsConfig(cfg WeaponElementalProcsConfig) error {
	if err := validateProcChance("cold", cfg.Cold.ProcChancePercent); err != nil {
		return err
	}
	if cfg.Cold.SlowPercent <= 0 || cfg.Cold.SlowPercent > 95 {
		return fmt.Errorf("game: invalid rules main_config.weapon_elemental_procs.cold.slow_percent: must be within [1,95]")
	}
	if cfg.Cold.DurationSeconds <= 0 {
		return fmt.Errorf("game: invalid rules main_config.weapon_elemental_procs.cold.duration_seconds: must be positive")
	}
	if cfg.Cold.EffectID == "" {
		return fmt.Errorf("game: invalid rules main_config.weapon_elemental_procs.cold.effect_id: required")
	}
	if err := validateProcChance("fire", cfg.Fire.ProcChancePercent); err != nil {
		return err
	}
	if cfg.Fire.BurnPercentOfTotalHit <= 0 || cfg.Fire.BurnPercentOfTotalHit > 100 {
		return fmt.Errorf("game: invalid rules main_config.weapon_elemental_procs.fire.burn_percent_of_total_hit: must be within [1,100]")
	}
	if cfg.Fire.DurationSeconds <= 0 || cfg.Fire.TickIntervalSeconds <= 0 {
		return fmt.Errorf("game: invalid rules main_config.weapon_elemental_procs.fire: duration and tick interval must be positive")
	}
	if cfg.Fire.EffectID == "" {
		return fmt.Errorf("game: invalid rules main_config.weapon_elemental_procs.fire.effect_id: required")
	}
	if err := validateProcChance("lightning", cfg.Lightning.ProcChancePercent); err != nil {
		return err
	}
	if cfg.Lightning.StunDurationSeconds <= 0 {
		return fmt.Errorf("game: invalid rules main_config.weapon_elemental_procs.lightning.stun_duration_seconds: must be positive")
	}
	if cfg.Lightning.EffectID == "" {
		return fmt.Errorf("game: invalid rules main_config.weapon_elemental_procs.lightning.effect_id: required")
	}
	if err := validateProcChance("poison", cfg.Poison.ProcChancePercent); err != nil {
		return err
	}
	if cfg.Poison.DotPercentOfElementalDamage <= 0 || cfg.Poison.DotPercentOfElementalDamage > 100 {
		return fmt.Errorf("game: invalid rules main_config.weapon_elemental_procs.poison.dot_percent_of_elemental_damage: must be within [1,100]")
	}
	if cfg.Poison.DurationSeconds <= 0 || cfg.Poison.TickIntervalSeconds <= 0 {
		return fmt.Errorf("game: invalid rules main_config.weapon_elemental_procs.poison: duration and tick interval must be positive")
	}
	if cfg.Poison.EffectID == "" {
		return fmt.Errorf("game: invalid rules main_config.weapon_elemental_procs.poison.effect_id: required")
	}

	return nil
}

func validateProcChance(label string, chance int) error {
	if chance < 0 || chance > 100 {
		return fmt.Errorf("game: invalid rules main_config.weapon_elemental_procs.%s.proc_chance_percent: must be within [0,100]", label)
	}

	return nil
}

func (s *Sim) rollWeaponElementalProc(chancePercent int) bool {
	if chancePercent <= 0 {
		return false
	}
	if chancePercent >= 100 {
		return true
	}

	return s.rng.IntN(100) < chancePercent
}

func (s *Sim) tryWeaponElementalProcs(target *entity, playerID uint64, corr string, damageType string, elementalDamage int, totalHitDamage int, res *TickResult) {
	if target == nil || target.hp <= 0 || elementalDamage <= 0 || s == nil {
		return
	}
	cfg := s.rules.MainConfig.WeaponElementalProcs
	switch canonicalDamageType(damageType) {
	case damageTypeCold:
		if !s.rollWeaponElementalProc(cfg.Cold.ProcChancePercent) {
			return
		}
		s.applyWeaponElementalFreeze(target, playerID, corr, cfg.Cold, res)
	case damageTypeFire:
		if !s.rollWeaponElementalProc(cfg.Fire.ProcChancePercent) {
			return
		}
		s.startWeaponElementalBurnDot(playerID, target, cfg.Fire, totalHitDamage, corr, res)
	case damageTypeLightning:
		if !s.rollWeaponElementalProc(cfg.Lightning.ProcChancePercent) {
			return
		}
		s.applyWeaponElementalStun(target, playerID, corr, cfg.Lightning, res)
	case damageTypePoison:
		if !s.rollWeaponElementalProc(cfg.Poison.ProcChancePercent) {
			return
		}
		s.startWeaponElementalPoisonDot(playerID, target, cfg.Poison, elementalDamage, corr, res)
	}
}

func (s *Sim) applyWeaponElementalFreeze(target *entity, sourceID uint64, corr string, cfg WeaponElementalColdProcConfig, res *TickResult) {
	if target == nil || target.kind != monsterEntity || target.hp <= 0 {
		return
	}
	durationTicks := cfg.DurationSeconds * 10
	if durationTicks <= 0 {
		return
	}
	effectID := cfg.EffectID
	if effectID == "" {
		effectID = "weapon_freeze"
	}
	stateKey := fmt.Sprintf("%s:%d", weaponElementalFreezeSkillID, target.id)
	s.skillEffects[stateKey] = skillEffectState{
		SkillID:    weaponElementalFreezeSkillID,
		TargetID:   target.id,
		Stats:      []string{"movement_speed", "attack_speed"},
		Percent:    cfg.SlowPercent,
		EffectID:   effectID,
		EndsTick:   s.tick + uint64(durationTicks),
		TotalTicks: durationTicks,
	}
	target.effectIDs = sortedUniqueStrings(append(target.effectIDs, effectID))
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
	res.Events = append(res.Events, Event{
		EventType:      "skill_effect_started",
		EntityID:       idStr(target.id),
		SourceEntityID: idStr(sourceID),
		TargetEntityID: idStr(target.id),
		CorrelationID:  corr,
		SkillID:        weaponElementalFreezeSkillID,
		Amount:         intPtr(cfg.SlowPercent),
		RemainingTicks: intPtr(durationTicks),
		TotalTicks:     intPtr(durationTicks),
	})
}

func (s *Sim) applyWeaponElementalStun(target *entity, sourceID uint64, corr string, cfg WeaponElementalLightningProcConfig, res *TickResult) {
	durationTicks := cfg.StunDurationSeconds * 10
	if durationTicks <= 0 {
		return
	}
	effectID := cfg.EffectID
	if effectID == "" {
		effectID = "stun"
	}
	s.applyMonsterRoot(target, sourceID, weaponElementalStunSkillID, SkillRootDef{
		EffectID:      effectID,
		DurationTicks: durationTicks,
	}, corr, res)
}

func (s *Sim) startWeaponElementalBurnDot(playerID uint64, target *entity, cfg WeaponElementalFireProcConfig, totalHitDamage int, corr string, res *TickResult) {
	if target == nil || totalHitDamage < 0 {
		return
	}
	damage := int(math.Round(float64(totalHitDamage) * float64(cfg.BurnPercentOfTotalHit) / 100.0))
	if totalHitDamage > 0 && damage < 1 {
		damage = 1
	}
	intervalTicks := cfg.TickIntervalSeconds * 10
	totalTicks := cfg.DurationSeconds * 10
	if intervalTicks <= 0 || totalTicks <= 0 {
		return
	}
	effectID := cfg.EffectID
	if effectID == "" {
		effectID = "weapon_burn"
	}
	s.uniqueBurnDots[uniqueBurnDotKey(effectID, target.id)] = uniqueBurnDotState{
		SourcePlayerID: playerID,
		TargetID:       target.id,
		EffectID:       effectID,
		DamageType:     damageTypeFire,
		DamagePerTick:  damage,
		NextTick:       s.tick + uint64(intervalTicks),
		IntervalTicks:  intervalTicks,
		RemainingTicks: totalTicks,
		TotalTicks:     totalTicks,
		CorrelationID:  corr,
	}
	target.effectIDs = sortedUniqueStrings(append(target.effectIDs, effectID))
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
	res.Events = append(res.Events, Event{
		EventType:      "skill_effect_started",
		EntityID:       idStr(target.id),
		SourceEntityID: idStr(playerID),
		TargetEntityID: idStr(target.id),
		CorrelationID:  corr,
		SkillID:        weaponElementalBurnSkillID,
		Amount:         intPtr(damage),
		RemainingTicks: intPtr(totalTicks),
		TotalTicks:     intPtr(totalTicks),
		DamageType:     damageTypeFire,
	})
}

func (s *Sim) startWeaponElementalPoisonDot(playerID uint64, target *entity, cfg WeaponElementalPoisonProcConfig, elementalDamage int, corr string, res *TickResult) {
	if target == nil || elementalDamage < 0 {
		return
	}
	damage := int(math.Round(float64(elementalDamage) * float64(cfg.DotPercentOfElementalDamage) / 100.0))
	if elementalDamage > 0 && damage < 1 {
		damage = 1
	}
	intervalTicks := cfg.TickIntervalSeconds * 10
	durationTicks := cfg.DurationSeconds * 10
	if intervalTicks <= 0 || durationTicks <= 0 {
		return
	}
	effectID := cfg.EffectID
	if effectID == "" {
		effectID = "weapon_poison"
	}
	s.poisonDots[target.id] = poisonDotState{
		SourcePlayerID: playerID,
		TargetID:       target.id,
		SkillID:        weaponElementalPoisonSkillID,
		Rank:           1,
		DamagePerTick:  damage,
		NextTick:       s.tick + uint64(intervalTicks),
		RemainingTicks: durationTicks,
		CorrelationID:  corr,
	}
	target.effectIDs = sortedUniqueStrings(append(target.effectIDs, effectID))
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
	res.Events = append(res.Events, Event{
		EventType:      "skill_effect_started",
		EntityID:       idStr(target.id),
		SourceEntityID: idStr(playerID),
		TargetEntityID: idStr(target.id),
		CorrelationID:  corr,
		SkillID:        weaponElementalPoisonSkillID,
		Rank:           intPtr(1),
		Amount:         intPtr(damage),
		RemainingTicks: intPtr(durationTicks),
		TotalTicks:     intPtr(durationTicks),
	})
}

func (s *Sim) monsterAttackCooldownTicks(monster *entity, baseCooldown int) int {
	if baseCooldown <= 0 || monster == nil {
		return baseCooldown
	}
	slowPercent := 0
	for _, stateKey := range sortedStringKeys(s.skillEffects) {
		effect := s.skillEffects[stateKey]
		if effect.TargetID != monster.id || effect.EndsTick <= s.tick {
			continue
		}
		if !containsStringValue(effect.Stats, "attack_speed") || effect.Percent <= slowPercent {
			continue
		}
		slowPercent = effect.Percent
	}
	if slowPercent <= 0 {
		return baseCooldown
	}
	if slowPercent > 95 {
		slowPercent = 95
	}
	scaled := float64(baseCooldown) * (1.0 + float64(slowPercent)/100.0)

	return int(math.Ceil(scaled))
}

func (s *Sim) mercenaryMainHandItem(companion *entity) *invItem {
	if companion == nil || companion.sourceCharacterID == "" || s.mercenaryRoster == nil {
		return nil
	}
	snap, ok := s.mercenaryRoster[companion.sourceCharacterID]
	if !ok {
		return nil
	}
	equipped := equippedItemsFromPersisted(snap.Items)

	return equipped[mainHandSlot]
}
