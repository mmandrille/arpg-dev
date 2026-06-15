package game

func (s *Sim) playerMagicFindPercent() float64 {
	stats, _ := s.playerEffectiveCombatStats()
	return stats.MagicFindPercent
}
