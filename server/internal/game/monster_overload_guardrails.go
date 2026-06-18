package game

// ApplyOverloadDegradation starts or extends a transient server-owned
// degradation window for the current room simulation.
func (s *Sim) ApplyOverloadDegradation() bool {
	ticks := s.activeNav().MonsterOverloadDegradeTicks
	if ticks <= 0 {
		return false
	}
	until := s.tick + uint64(ticks)
	if until > s.overloadDegradeUntilTick {
		s.overloadDegradeUntilTick = until
	}
	return true
}

func (s *Sim) overloadDegraded() bool {
	return s.tick < s.overloadDegradeUntilTick
}
