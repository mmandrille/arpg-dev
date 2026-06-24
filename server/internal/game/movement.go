package game

func (s *Sim) playerMoveSpeed() float64 {
	return s.playerEffectiveMovementSpeed() * s.playerSlowMultiplier()
}

func (s *Sim) classBaseMovementSpeed() float64 {
	if s.rules != nil {
		if classDef, ok := s.rules.CharacterProgression.Classes[s.progression.CharacterClass]; ok {
			if classDef.BaseMovementSpeed > 0 {
				return classDef.BaseMovementSpeed
			}
		}
		if s.rules.MainConfig.Gameplay.BaseMovementSpeed > 0 {
			return s.rules.MainConfig.Gameplay.BaseMovementSpeed
		}
	}
	return defaultMoveSpeed
}

func (s *Sim) playerEffectiveMovementSpeed() float64 {
	classBase := s.classBaseMovementSpeed()
	character := s.characterDerivedStatsView()
	effective, _ := s.playerEffectiveCombatStats()
	return classBase * character.MovementSpeed * (1 + effective.MovementSpeedPercent/100)
}

func (s *Sim) playerSlowMultiplier() float64 {
	slowPercent := 0
	for _, stateKey := range sortedStringKeys(s.skillEffects) {
		effect := s.skillEffects[stateKey]
		if effect.EndsTick <= s.tick {
			continue
		}
		if effect.TargetID != 0 && effect.TargetID != s.playerID {
			continue
		}
		if !containsStringValue(effect.Stats, "movement_speed") || effect.Percent <= slowPercent {
			continue
		}
		slowPercent = effect.Percent
	}
	if slowPercent <= 0 {
		return 1.0
	}
	if slowPercent > 95 {
		slowPercent = 95
	}
	return 1.0 - float64(slowPercent)/100.0
}

func (s *Sim) playerMoveMomentumMultiplier(heldTicks int) float64 {
	if s.rules == nil || heldTicks <= 0 {
		return 1
	}
	accelSeconds := s.rules.MainConfig.Gameplay.MovementAccelerationSeconds
	if accelSeconds <= 0 {
		return 1
	}
	wantTicks := int(accelSeconds * simulationTickHz)
	if wantTicks < 1 {
		wantTicks = 1
	}
	mult := float64(heldTicks) / float64(wantTicks)
	if mult > 1 {
		mult = 1
	}
	minFactor := s.rules.MainConfig.Gameplay.MovementMinSpeedFactor
	if minFactor <= 0 || minFactor > 1 {
		minFactor = 0.2
	}
	if mult < minFactor {
		mult = minFactor
	}
	return mult
}
