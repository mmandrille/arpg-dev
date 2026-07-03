package game

// skillPointGrantLevel reports whether a character earns a skill point when reaching level.
func (r *Rules) skillPointGrantLevel(level int) bool {
	cadence := r.CharacterProgression.SkillPoints
	if cadence.PointsPerGrant <= 0 || cadence.GrantEveryLevels <= 0 || level < cadence.FirstGrantLevel {
		return false
	}

	if level == cadence.FirstGrantLevel {
		return true
	}

	return level%cadence.GrantEveryLevels == 0
}

func (r *Rules) totalSkillPointGrantsForLevel(level int) int {
	if level < 1 {
		return 0
	}
	grants := 0
	for grantLevel := 1; grantLevel <= level; grantLevel++ {
		if r.skillPointGrantLevel(grantLevel) {
			grants++
		}
	}

	return grants * r.CharacterProgression.SkillPoints.PointsPerGrant
}

func (r *Rules) classLevelStatGrowthTotal(classID string, level int) BaseStatsView {
	out := BaseStatsView{}
	if level <= 1 || r == nil {
		return out
	}
	classDef, ok := r.CharacterProgression.Classes[classID]
	if !ok {
		return out
	}
	growth := classDef.LevelStatGrowth
	levelsGained := level - 1
	out.Str = growth.Str * levelsGained
	out.Dex = growth.Dex * levelsGained
	out.Vit = growth.Vit * levelsGained
	out.Magic = growth.Magic * levelsGained

	return out
}

func (r *Rules) classGrownBaseStats(classID string, level int) BaseStatsView {
	base := r.CharacterProgression.BaseStats
	if classDef, ok := r.CharacterProgression.Classes[classID]; ok {
		base = classDef.BaseStats
	}
	growth := r.classLevelStatGrowthTotal(classID, level)

	return BaseStatsView{
		Str:   base.Str + growth.Str,
		Dex:   base.Dex + growth.Dex,
		Vit:   base.Vit + growth.Vit,
		Magic: base.Magic + growth.Magic,
	}
}

func (r *Rules) applyClassLevelStatGrowthFloor(state *CharacterProgressionState) {
	if state == nil || r == nil {
		return
	}
	floor := r.classGrownBaseStats(state.CharacterClass, state.Level)
	if state.BaseStats.Str < floor.Str {
		state.BaseStats.Str = floor.Str
	}
	if state.BaseStats.Dex < floor.Dex {
		state.BaseStats.Dex = floor.Dex
	}
	if state.BaseStats.Vit < floor.Vit {
		state.BaseStats.Vit = floor.Vit
	}
	if state.BaseStats.Magic < floor.Magic {
		state.BaseStats.Magic = floor.Magic
	}
}

func (r *Rules) singleLevelStatGrowth(classID string) BaseStatsView {
	return r.classLevelStatGrowthTotal(classID, 2)
}

func (s *Sim) applyClassLevelStatGrowthOnLevelUp() {
	growth := s.rules.singleLevelStatGrowth(s.progression.CharacterClass)
	s.progression.BaseStats.Str += growth.Str
	s.progression.BaseStats.Dex += growth.Dex
	s.progression.BaseStats.Vit += growth.Vit
	s.progression.BaseStats.Magic += growth.Magic
}
