package game

import (
	"fmt"
	"math"
)

func validateSkillRules(skills map[string]SkillDef, monsters map[string]MonsterDef, baseAttackIntervalTicks int) error {
	if len(skills) == 0 {
		return fmt.Errorf("game: invalid rules skills: at least one skill is required")
	}
	for id, skill := range skills {
		if id == "" {
			return fmt.Errorf("game: invalid rules skills: empty skill id")
		}
		if skill.Name == "" {
			return fmt.Errorf("game: invalid rules skills.%s.name: required", id)
		}
		if skill.Class == "" {
			return fmt.Errorf("game: invalid rules skills.%s.class: required", id)
		}
		if skill.Tree.Tier <= 0 || skill.Tree.Column <= 0 {
			return fmt.Errorf("game: invalid rules skills.%s.tree: tier and column must be positive", id)
		}
		if !isSupportedSkillKind(skill.Kind) {
			return fmt.Errorf("game: invalid rules skills.%s.kind: unsupported %s", id, skill.Kind)
		}
		if skill.MaxRank <= 0 {
			return fmt.Errorf("game: invalid rules skills.%s.max_rank: must be positive", id)
		}
		if err := validateSkillRequirements(id, skill.Requirements, skills); err != nil {
			return err
		}
		if skill.Cost.Mana.Base < 0 || skill.Cost.Mana.PerRank < 0 {
			return fmt.Errorf("game: invalid rules skills.%s.cost.mana: values must be non-negative", id)
		}
		if err := validateSkillKindPayload(id, skill, monsters); err != nil {
			return err
		}
		if skill.Cooldown.Type != "attack_interval_multiplier" && skill.Cooldown.Type != "none" {
			return fmt.Errorf("game: invalid rules skills.%s.cooldown.type: unsupported %s", id, skill.Cooldown.Type)
		}
		if skill.Cooldown.Type == "attack_interval_multiplier" && skill.Cooldown.Multiplier <= 0 {
			return fmt.Errorf("game: invalid rules skills.%s.cooldown.multiplier: must be positive", id)
		}
		if skill.Cooldown.FlatTicks < 0 {
			return fmt.Errorf("game: invalid rules skills.%s.cooldown.flat_ticks: must be non-negative", id)
		}
		if skill.Cooldown.FixedTicks < 0 {
			return fmt.Errorf("game: invalid rules skills.%s.cooldown.fixed_ticks: must be non-negative", id)
		}
		if skill.Cooldown.MagicReductionTicksPerPoint < 0 {
			return fmt.Errorf("game: invalid rules skills.%s.cooldown.magic_reduction_ticks_per_point: must be non-negative", id)
		}
		if err := validateBuffSkillCooldown(id, skill, baseAttackIntervalTicks); err != nil {
			return err
		}
	}
	return nil
}

func validateBuffSkillCooldown(skillID string, skill SkillDef, baseAttackIntervalTicks int) error {
	if skill.Kind != "self_buff" && skill.Kind != "area_stat_buff" {
		return nil
	}
	maxDuration := 0
	for _, effect := range skill.Effects {
		if effect.DurationTicks > maxDuration {
			maxDuration = effect.DurationTicks
		}
	}
	if maxDuration <= 0 {
		return nil
	}
	if baseAttackIntervalTicks < 1 {
		baseAttackIntervalTicks = 1
	}
	cooldownTicks := int(math.Ceil(float64(baseAttackIntervalTicks)*skill.Cooldown.Multiplier)) + skill.Cooldown.FlatTicks
	minCooldownTicks := int(math.Ceil(float64(maxDuration) * 1.5))
	if cooldownTicks < minCooldownTicks {
		return fmt.Errorf("game: invalid rules skills.%s.cooldown: buff cooldown %d must be at least 150%% of duration %d (min %d)",
			skillID, cooldownTicks, maxDuration, minCooldownTicks)
	}
	return nil
}

func isSupportedSkillKind(kind string) bool {
	switch kind {
	case "projectile_attack", "cold_projectile_attack", "chain_projectile_attack", "cone_attack", "self_buff", "area_heal", "area_stat_buff", "summon_companion", "revive_companion", "mobility", "passive_execute", "passive_stat_bonus", "survival_autocast":
		return true
	default:
		return false
	}
}

func validateSkillKindPayload(skillID string, skill SkillDef, monsters map[string]MonsterDef) error {
	if err := validateDamageType("skills."+skillID+".damage_type", skill.DamageType); err != nil {
		return err
	}
	switch skill.Kind {
	case "projectile_attack":
		if skill.Targeting != "direction_or_target" {
			return fmt.Errorf("game: invalid rules skills.%s.targeting: unsupported %s for projectile_attack", skillID, skill.Targeting)
		}
		return validateProjectileSkillPayload(skillID, skill)
	case "cold_projectile_attack":
		if skill.Targeting != "direction_or_target" {
			return fmt.Errorf("game: invalid rules skills.%s.targeting: unsupported %s for cold_projectile_attack", skillID, skill.Targeting)
		}
		if err := validateProjectileSkillPayload(skillID, skill); err != nil {
			return err
		}
		return validateColdSkillPayload(skillID, skill)
	case "chain_projectile_attack":
		if skill.Targeting != "direction_or_target" {
			return fmt.Errorf("game: invalid rules skills.%s.targeting: unsupported %s for chain_projectile_attack", skillID, skill.Targeting)
		}
		if err := validateProjectileSkillPayload(skillID, skill); err != nil {
			return err
		}
		return validateChainSkillPayload(skillID, skill)
	case "cone_attack":
		if skill.Targeting != "direction_or_target" {
			return fmt.Errorf("game: invalid rules skills.%s.targeting: unsupported %s for cone_attack", skillID, skill.Targeting)
		}
		return validateConeSkillPayload(skillID, skill)
	case "self_buff":
		if skill.Targeting != "self" {
			return fmt.Errorf("game: invalid rules skills.%s.targeting: unsupported %s for self_buff", skillID, skill.Targeting)
		}
		return validateSkillEffects(skillID, skill.Effects, "stat_percent_buff", "reflect_on_block_buff")
	case "area_heal":
		if skill.Targeting != "direction_or_target_area" {
			return fmt.Errorf("game: invalid rules skills.%s.targeting: unsupported %s for area_heal", skillID, skill.Targeting)
		}
		return validateSkillEffects(skillID, skill.Effects, "area_percent_heal")
	case "area_stat_buff":
		if skill.Targeting != "self_or_ally_area" {
			return fmt.Errorf("game: invalid rules skills.%s.targeting: unsupported %s for area_stat_buff", skillID, skill.Targeting)
		}
		return validateSkillEffects(skillID, skill.Effects, "area_stat_percent_buff", "area_immunity_buff")
	case "summon_companion":
		if skill.Targeting != "self" {
			return fmt.Errorf("game: invalid rules skills.%s.targeting: unsupported %s for summon_companion", skillID, skill.Targeting)
		}
		return validateSummonCompanionSkillPayload(skillID, skill, monsters)
	case "revive_companion":
		return validateReviveCompanionSkillPayload(skillID, skill)
	case "mobility":
		if skill.Targeting != "direction_or_target" {
			return fmt.Errorf("game: invalid rules skills.%s.targeting: unsupported %s for mobility", skillID, skill.Targeting)
		}
		return validateRogueConeSkillPayload(skillID, skill)
	case "passive_execute":
		if skill.Targeting != "self" {
			return fmt.Errorf("game: invalid rules skills.%s.targeting: unsupported %s for passive_execute", skillID, skill.Targeting)
		}
		return validatePassiveExecuteSkillPayload(skillID, skill)
	case "passive_stat_bonus":
		if skill.Targeting != "self" {
			return fmt.Errorf("game: invalid rules skills.%s.targeting: unsupported %s for passive_stat_bonus", skillID, skill.Targeting)
		}
		return validatePassiveStatSkillPayload(skillID, skill)
	case "survival_autocast":
		return validateSurvivalAutocastSkill(skillID, skill)
	default:
		return fmt.Errorf("game: invalid rules skills.%s.kind: unsupported %s", skillID, skill.Kind)
	}
}

func validatePassiveExecuteSkillPayload(skillID string, skill SkillDef) error {
	if skill.Execute.ThresholdPercentBase <= 0 || skill.Execute.ThresholdPercentBase > 100 ||
		skill.Execute.ThresholdPercentPerRank < 0 || skill.Execute.ChancePercent <= 0 || skill.Execute.ChancePercent > 100 {
		return fmt.Errorf("game: invalid rules skills.%s.execute: threshold and chance must be valid", skillID)
	}
	maxThreshold := skill.Execute.ThresholdPercentBase + skill.Execute.ThresholdPercentPerRank*(skill.MaxRank-1)
	if maxThreshold > 100 {
		return fmt.Errorf("game: invalid rules skills.%s.execute: max threshold must be <= 100", skillID)
	}
	if len(skill.Effects) > 0 || len(skill.PassiveStats.Stats) > 0 || skill.Projectile.Range > 0 || skill.Cone.Range > 0 || skill.Dash.RangeBase > 0 || skill.Mobility.RangeBase > 0 {
		return fmt.Errorf("game: invalid rules skills.%s: passive_execute does not support active payloads", skillID)
	}
	return nil
}

func validateProjectileSkillPayload(skillID string, skill SkillDef) error {
	switch skill.Damage.Type {
	case "", "rank_linear_range":
		if skill.Damage.MinBase < 0 || skill.Damage.MaxBase < skill.Damage.MinBase {
			return fmt.Errorf("game: invalid rules skills.%s.damage: base damage must be valid", skillID)
		}
	case "weapon_multiplier_range":
		if skill.Damage.MinBase < 1 || skill.Damage.MaxBase < 1 {
			return fmt.Errorf("game: invalid rules skills.%s.damage: weapon multiplier percents must be positive", skillID)
		}
	default:
		return fmt.Errorf("game: invalid rules skills.%s.damage.type: unsupported %s", skillID, skill.Damage.Type)
	}
	if skill.Damage.MinPerRank < 0 || skill.Damage.MaxPerRank < 0 {
		return fmt.Errorf("game: invalid rules skills.%s.damage: per-rank damage must be non-negative", skillID)
	}
	if err := validateSkillMagicScaling(fmt.Sprintf("skills.%s.damage.magic_scaling", skillID), skill.Damage.MagicScaling); err != nil {
		return err
	}
	if skill.Projectile.Range <= 0 {
		return fmt.Errorf("game: invalid rules skills.%s.projectile.range: must be positive", skillID)
	}
	if skill.Projectile.Speed <= 0 {
		return fmt.Errorf("game: invalid rules skills.%s.projectile.speed: must be positive", skillID)
	}
	if skill.Projectile.Visual == "" {
		return fmt.Errorf("game: invalid rules skills.%s.projectile.visual: required", skillID)
	}
	if len(skill.Effects) > 0 {
		return fmt.Errorf("game: invalid rules skills.%s.effects: projectile_attack does not support effects", skillID)
	}
	return validateRangerProjectileSkillPayload(skillID, skill)
}

func validateRangerProjectileSkillPayload(skillID string, skill SkillDef) error {
	if skill.Pierce.MaxHits > 0 {
		if skill.Pierce.MaxHits < 2 {
			return fmt.Errorf("game: invalid rules skills.%s.pierce.max_hits: must be at least 2", skillID)
		}
		if skill.Pierce.DamagePercentPerExtraHit <= 0 || skill.Pierce.DamagePercentPerExtraHit > 100 {
			return fmt.Errorf("game: invalid rules skills.%s.pierce.damage_percent_per_extra_hit: must be 1..100", skillID)
		}
	}
	if skill.Root.DurationTicks > 0 || skill.Root.EffectID != "" {
		if skill.Root.EffectID == "" {
			return fmt.Errorf("game: invalid rules skills.%s.root.effect_id: required", skillID)
		}
		if skill.Root.DurationTicks <= 0 {
			return fmt.Errorf("game: invalid rules skills.%s.root.duration_ticks: must be positive", skillID)
		}
	}
	if skill.Volley.ArrowCount > 0 || skill.Volley.SpreadDegrees > 0 {
		if skill.Volley.ArrowCount < 3 || skill.Volley.ArrowCount > 9 {
			return fmt.Errorf("game: invalid rules skills.%s.volley.arrow_count: must be 3..9", skillID)
		}
		if skill.Volley.SpreadDegrees <= 0 || skill.Volley.SpreadDegrees > 120 {
			return fmt.Errorf("game: invalid rules skills.%s.volley.spread_degrees: must be > 0 and <= 120", skillID)
		}
	}
	mechanicCount := 0
	if skill.Pierce.MaxHits > 0 {
		mechanicCount++
	}
	if skill.Root.DurationTicks > 0 {
		mechanicCount++
	}
	if skill.Volley.ArrowCount > 0 {
		mechanicCount++
	}
	if mechanicCount > 1 {
		return fmt.Errorf("game: invalid rules skills.%s: ranger projectile mechanics cannot be combined", skillID)
	}
	return nil
}

func validateConeSkillPayload(skillID string, skill SkillDef) error {
	if skill.Cone.Range <= 0 {
		return fmt.Errorf("game: invalid rules skills.%s.cone.range: must be positive", skillID)
	}
	if skill.Cone.AngleDegrees <= 0 || skill.Cone.AngleDegrees > 360 {
		return fmt.Errorf("game: invalid rules skills.%s.cone.angle_degrees: must be > 0 and <= 360", skillID)
	}
	if skill.Cone.PushMin < 0 || skill.Cone.PushMax < skill.Cone.PushMin {
		return fmt.Errorf("game: invalid rules skills.%s.cone.push: min/max must be valid", skillID)
	}
	if skill.Cone.DamageSource != "weapon" {
		return fmt.Errorf("game: invalid rules skills.%s.cone.damage_source: unsupported %s", skillID, skill.Cone.DamageSource)
	}
	if len(skill.Effects) > 0 || skill.Projectile.Range > 0 {
		return fmt.Errorf("game: invalid rules skills.%s: cone_attack does not support effects or projectile", skillID)
	}
	return validateRogueConeSkillPayload(skillID, skill)
}

func validateColdSkillPayload(skillID string, skill SkillDef) error {
	if skill.Slow.EffectID == "" {
		return fmt.Errorf("game: invalid rules skills.%s.slow.effect_id: required", skillID)
	}
	if skill.Slow.Percent <= 0 || skill.Slow.Percent > 100 {
		return fmt.Errorf("game: invalid rules skills.%s.slow.percent: must be between 1 and 100", skillID)
	}
	if skill.Slow.MaxPercent < skill.Slow.Percent || skill.Slow.MaxPercent > 100 {
		return fmt.Errorf("game: invalid rules skills.%s.slow.max_percent: must be between percent and 100", skillID)
	}
	if skill.Slow.DurationTicks <= 0 {
		return fmt.Errorf("game: invalid rules skills.%s.slow.duration_ticks: must be positive", skillID)
	}
	if skill.Shatter.MinShards <= 0 || skill.Shatter.MaxShards < skill.Shatter.MinShards {
		return fmt.Errorf("game: invalid rules skills.%s.shatter.shards: min/max must be valid", skillID)
	}
	if skill.Shatter.Range <= 0 || skill.Shatter.Speed <= 0 || skill.Shatter.Visual == "" {
		return fmt.Errorf("game: invalid rules skills.%s.shatter: range, speed, and visual are required", skillID)
	}
	return nil
}

func validateChainSkillPayload(skillID string, skill SkillDef) error {
	if skill.Chain.RangeMultiplier <= 0 || skill.Chain.RangeMultiplier >= 1 {
		return fmt.Errorf("game: invalid rules skills.%s.chain.range_multiplier: must be > 0 and < 1", skillID)
	}
	if skill.Chain.MaxJumps <= 0 {
		return fmt.Errorf("game: invalid rules skills.%s.chain.max_jumps: must be positive", skillID)
	}
	if skill.Chain.Visual == "" {
		return fmt.Errorf("game: invalid rules skills.%s.chain.visual: required", skillID)
	}
	return nil
}

func validateSkillEffects(skillID string, effects []SkillEffectDef, expectedTypes ...string) error {
	if len(effects) == 0 {
		return fmt.Errorf("game: invalid rules skills.%s.effects: at least one supported effect is required", skillID)
	}
	for idx, effect := range effects {
		if !stringInSlice(effect.Type, expectedTypes) {
			return fmt.Errorf("game: invalid rules skills.%s.effects[%d].type: unsupported %s for skill kind", skillID, idx, effect.Type)
		}
		switch effect.Type {
		case "stat_percent_buff":
			if err := validateStatPercentBuffEffect(skillID, idx, effect); err != nil {
				return err
			}
		case "area_percent_heal":
			if err := validateAreaPercentHealEffect(skillID, idx, effect); err != nil {
				return err
			}
		case "area_stat_percent_buff":
			if err := validateAreaStatPercentBuffEffect(skillID, idx, effect); err != nil {
				return err
			}
		case "area_immunity_buff":
			if err := validateAreaImmunityBuffEffect(skillID, idx, effect); err != nil {
				return err
			}
		case "reflect_on_block_buff":
			if err := validateReflectOnBlockBuffEffect(skillID, idx, effect); err != nil {
				return err
			}
		default:
			return fmt.Errorf("game: invalid rules skills.%s.effects[%d].type: unsupported %s", skillID, idx, effect.Type)
		}
	}
	return nil
}

func validateStatPercentBuffEffect(skillID string, idx int, effect SkillEffectDef) error {
	if len(effect.Stats) == 0 {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d].stats: at least one stat is required", skillID, idx)
	}
	seen := map[string]bool{}
	for _, stat := range effect.Stats {
		if !isSupportedRequirementStat(stat) {
			return fmt.Errorf("game: invalid rules skills.%s.effects[%d].stats.%s: unsupported stat", skillID, idx, stat)
		}
		if seen[stat] {
			return fmt.Errorf("game: invalid rules skills.%s.effects[%d].stats.%s: duplicate stat", skillID, idx, stat)
		}
		seen[stat] = true
	}
	if effect.PercentBase < 0 || effect.PercentPerRank < 0 {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d]: percent values must be non-negative", skillID, idx)
	}
	if effect.PercentBase == 0 && effect.PercentPerRank == 0 {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d]: percent values cannot both be zero", skillID, idx)
	}
	if effect.DurationTicks <= 0 {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d].duration_ticks: must be positive", skillID, idx)
	}
	return nil
}

func validateAreaPercentHealEffect(skillID string, idx int, effect SkillEffectDef) error {
	if effect.Target != "allies" {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d].target: unsupported %s", skillID, idx, effect.Target)
	}
	if effect.Range <= 0 {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d].range: must be positive", skillID, idx)
	}
	if effect.Radius <= 0 {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d].radius: must be positive", skillID, idx)
	}
	if effect.PercentBase < 0 || effect.PercentPerRank < 0 {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d]: percent values must be non-negative", skillID, idx)
	}
	if effect.PercentBase == 0 && effect.PercentPerRank == 0 {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d]: percent values cannot both be zero", skillID, idx)
	}
	if effect.DurationTicks <= 0 {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d].duration_ticks: must be positive", skillID, idx)
	}
	if err := validateSkillMagicScaling(fmt.Sprintf("skills.%s.effects[%d].magic_scaling", skillID, idx), effect.MagicScaling); err != nil {
		return err
	}
	return nil
}

func validateReflectOnBlockBuffEffect(skillID string, idx int, effect SkillEffectDef) error {
	if effect.PercentBase <= 0 {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d].percent_base: must be positive", skillID, idx)
	}
	if effect.PercentPerRank < 0 {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d].percent_per_rank: must be non-negative", skillID, idx)
	}
	if effect.DurationTicks <= 0 {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d].duration_ticks: must be positive", skillID, idx)
	}
	if effect.EffectID == "" {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d].effect_id: required", skillID, idx)
	}

	return nil
}

func validateSkillMagicScaling(path string, scaling SkillScalingDef) error {
	if scaling.Stat == "" {
		return nil
	}
	if scaling.Stat != "magic" {
		return fmt.Errorf("game: invalid rules %s.stat: unsupported %s", path, scaling.Stat)
	}
	if scaling.PercentPerPoint <= 0 {
		return fmt.Errorf("game: invalid rules %s.percent_per_point: must be positive", path)
	}
	if scaling.MaxBonusPercent <= 0 || scaling.MaxBonusPercent > 100 {
		return fmt.Errorf("game: invalid rules %s.max_bonus_percent: must be within (0,100]", path)
	}
	return nil
}

func validateSkillRequirements(skillID string, req SkillRequirementDef, skills map[string]SkillDef) error {
	if req.Level <= 0 {
		return fmt.Errorf("game: invalid rules skills.%s.requirements.level: must be positive", skillID)
	}
	if req.LevelPerRank < 0 {
		return fmt.Errorf("game: invalid rules skills.%s.requirements.level_per_rank: must be non-negative", skillID)
	}
	for stat, value := range req.Stats {
		if !isSupportedRequirementStat(stat) {
			return fmt.Errorf("game: invalid rules skills.%s.requirements.stats.%s: unsupported requirement", skillID, stat)
		}
		if value <= 0 {
			return fmt.Errorf("game: invalid rules skills.%s.requirements.stats.%s: must be positive", skillID, stat)
		}
	}
	for stat, value := range req.StatsPerRank {
		if !isSupportedRequirementStat(stat) {
			return fmt.Errorf("game: invalid rules skills.%s.requirements.stats_per_rank.%s: unsupported requirement", skillID, stat)
		}
		if value < 0 {
			return fmt.Errorf("game: invalid rules skills.%s.requirements.stats_per_rank.%s: must be non-negative", skillID, stat)
		}
	}
	for _, prereq := range req.Skills {
		if prereq.SkillID == "" {
			return fmt.Errorf("game: invalid rules skills.%s.requirements.skills: skill_id is required", skillID)
		}
		required, ok := skills[prereq.SkillID]
		if !ok {
			return fmt.Errorf("game: invalid rules skills.%s.requirements.skills.%s: unknown skill", skillID, prereq.SkillID)
		}
		if prereq.Rank <= 0 || prereq.Rank > required.MaxRank {
			return fmt.Errorf("game: invalid rules skills.%s.requirements.skills.%s.rank: must be within max rank", skillID, prereq.SkillID)
		}
	}
	return nil
}
