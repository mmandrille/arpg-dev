package game

import (
	"fmt"
	"math"
)

// SkillSynergyDef is a declarative bonus from an allocated prerequisite skill rank.
type SkillSynergyDef struct {
	SourceSkillID         string  `json:"source_skill_id"`
	Modifier              string  `json:"modifier"`
	PercentPerSourceRank  float64 `json:"percent_per_source_rank"`
}

func (s *Sim) allocatedSkillRank(skillID string) int {
	if skillID == "" {
		return 0
	}

	return s.progression.SkillRanks[skillID]
}

func (s *Sim) synergyBonusPercent(skillID, modifier string) int {
	if s == nil || s.rules == nil || skillID == "" || modifier == "" {
		return 0
	}
	def, ok := s.rules.Skills[skillID]
	if !ok || len(def.Synergies) == 0 {
		return 0
	}

	total := 0.0
	for _, synergy := range def.Synergies {
		if synergy.Modifier != modifier {
			continue
		}
		sourceRank := s.allocatedSkillRank(synergy.SourceSkillID)
		if sourceRank <= 0 {
			continue
		}
		total += float64(sourceRank) * synergy.PercentPerSourceRank
	}

	return int(math.Round(total))
}

func (s *Sim) synergyScaledInt(skillID, modifier string, base int) int {
	if base <= 0 {
		return base
	}
	bonus := s.synergyBonusPercent(skillID, modifier)
	if bonus <= 0 {
		return base
	}

	return scaleStatPercent(base, bonus)
}

func (s *Sim) synergyScaledFloat(skillID, modifier string, base float64) float64 {
	if base <= 0 {
		return base
	}
	bonus := s.synergyBonusPercent(skillID, modifier)
	if bonus <= 0 {
		return base
	}

	return base * (1.0 + float64(bonus)/100.0)
}

func (s *Sim) skillDamageRangeForSkill(skillID string, def SkillDef, rank int) DamageRange {
	damageRange := s.skillDamageRange(def, rank)
	bonus := s.synergyBonusPercent(skillID, "damage_percent")
	if bonus <= 0 {
		return damageRange
	}

	return DamageRange{
		Min: scaleStatPercent(damageRange.Min, bonus),
		Max: scaleStatPercent(damageRange.Max, bonus),
	}
}

func (s *Sim) effectiveConeForSkill(skillID string, cone SkillConeDef) SkillConeDef {
	if cone.Range <= 0 && cone.AngleDegrees <= 0 {
		return cone
	}
	out := cone
	out.Range = s.synergyScaledFloat(skillID, "cone_size_percent", cone.Range)
	if cone.AngleDegrees >= 360 {
		out.AngleDegrees = 360
	} else {
		out.AngleDegrees = s.synergyScaledFloat(skillID, "cone_size_percent", cone.AngleDegrees)
	}

	return out
}

func (s *Sim) effectiveVolleySpreadForSkill(skillID string, spreadDegrees float64) float64 {
	return s.synergyScaledFloat(skillID, "volley_spread_percent", spreadDegrees)
}

func (s *Sim) effectiveProjectileRangeForSkill(skillID string, castRange float64) float64 {
	return s.synergyScaledFloat(skillID, "projectile_range_percent", castRange)
}

func (s *Sim) skillSynergyStatus(skillID string) []SkillSynergyStatusView {
	if s == nil || s.rules == nil {
		return nil
	}
	def, ok := s.rules.Skills[skillID]
	if !ok || len(def.Synergies) == 0 {
		return nil
	}

	out := make([]SkillSynergyStatusView, 0, len(def.Synergies))
	for _, synergy := range def.Synergies {
		sourceRank := s.allocatedSkillRank(synergy.SourceSkillID)
		bonus := int(math.Round(float64(sourceRank) * synergy.PercentPerSourceRank))
		sourceName := synergy.SourceSkillID
		if sourceDef, sourceOK := s.rules.Skills[synergy.SourceSkillID]; sourceOK && sourceDef.Name != "" {
			sourceName = sourceDef.Name
		}
		out = append(out, SkillSynergyStatusView{
			SourceSkillID: synergy.SourceSkillID,
			SourceName:    sourceName,
			SourceRank:    sourceRank,
			Modifier:      synergy.Modifier,
			BonusPercent:  bonus,
			Display:       fmt.Sprintf("+%d%% from %s (rank %d)", bonus, sourceName, sourceRank),
		})
	}

	return out
}

func passiveSkillRankedStatWithSynergy(s *Sim, skillID string, def SkillDef, stat string, rank int) int {
	base := passiveSkillRankedStat(s.rules, def, stat, rank)
	if s == nil {
		return base
	}

	return s.synergyScaledInt(skillID, "passive_stat_percent", base)
}

func skillBleedValuesWithSynergy(s *Sim, skillID string, bleed SkillBleedDef, rank int) (damagePercentMaxHP, durationTicks, intervalTicks int) {
	damagePercentMaxHP, durationTicks, intervalTicks = skillBleedValues(s.rules, bleed, rank)
	if s == nil {
		return damagePercentMaxHP, durationTicks, intervalTicks
	}
	durationTicks = s.synergyScaledInt(skillID, "bleed_duration_percent", durationTicks)

	return damagePercentMaxHP, durationTicks, intervalTicks
}

func skillMarkValuesWithSynergy(s *Sim, skillID string, mark SkillMarkDef, rank int) (damageBonusPercent, durationTicks int) {
	damageBonusPercent, durationTicks = skillMarkValues(s.rules, mark, rank)
	if s == nil {
		return damageBonusPercent, durationTicks
	}
	durationTicks = s.synergyScaledInt(skillID, "mark_duration_percent", durationTicks)

	return damageBonusPercent, durationTicks
}

func revivePowerPercentWithSynergy(s *Sim, skillID string, def SkillDef, rank int) int {
	power := revivePowerPercent(s.rules, def, rank)
	if s == nil {
		return power
	}

	return s.synergyScaledInt(skillID, "revive_power_percent", power)
}

func skillEffectPercentWithSynergy(s *Sim, skillID string, effect SkillEffectDef, rank int) int {
	percent := skillEffectPercent(s.rules, effect, rank)
	if s == nil {
		return percent
	}

	return s.synergyScaledInt(skillID, "buff_power_percent", percent)
}

func skillEffectDurationWithSynergy(s *Sim, skillID string, durationTicks int) int {
	if s == nil {
		return durationTicks
	}

	return s.synergyScaledInt(skillID, "buff_duration_percent", durationTicks)
}
