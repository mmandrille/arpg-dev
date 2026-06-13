package game

func (s *Sim) applyEliteAuraToMonsterDamage(monster *entity, damage DamageRange) DamageRange {
	if monster == nil || monster.kind != monsterEntity || monster.hp <= 0 {
		return damage
	}
	if monster.monsterPackID == "" || monster.monsterPackLeader {
		return damage
	}
	aura := s.rules.DungeonGeneration.MonsterPlacement.EliteAura
	if aura == nil || aura.DamageBonusPercent <= 0 {
		return damage
	}
	if !s.monsterHasLivingPackLeaderInAura(monster, aura.Radius) {
		return damage
	}
	return DamageRange{
		Min: applyPercentBonus(damage.Min, aura.DamageBonusPercent),
		Max: applyPercentBonus(damage.Max, aura.DamageBonusPercent),
	}
}

func (s *Sim) monsterHasLivingPackLeaderInAura(monster *entity, radius float64) bool {
	level := s.activeLevel()
	for _, id := range sortedEntityIDs(level.entities) {
		candidate := level.entities[id]
		if candidate == nil || candidate.kind != monsterEntity || candidate.hp <= 0 {
			continue
		}
		if !candidate.monsterPackLeader || candidate.monsterPackID != monster.monsterPackID {
			continue
		}
		if distance(candidate.pos, monster.pos) <= radius {
			return true
		}
	}
	return false
}
