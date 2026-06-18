package game

// TickResults processes current-tick inputs, applies continuous effects, and returns scoped results.
func (s *Sim) TickResults(inputs []Input) []TickResult {
	return s.TickResultsProfiled(inputs, nil)
}

// TickResultsProfiled processes one tick while letting runtime callers measure
// selected phases. The profiler must not mutate sim state.
func (s *Sim) TickResultsProfiled(inputs []Input, profiler TickProfiler) []TickResult {
	s.resetTickPerf()
	previousProfiler := s.tickProfiler
	s.tickProfiler = profiler
	defer func() {
		s.tickProfiler = previousProfiler
	}()
	type resultKey struct {
		level int
		actor uint64
	}
	resultByKey := map[resultKey]*TickResult{}
	var ordered []*TickResult
	transitionThisTick := false
	resultFor := func(level int, actor uint64) *TickResult {
		key := resultKey{level: level, actor: actor}
		if res := resultByKey[key]; res != nil {
			return res
		}
		res := &TickResult{Tick: s.tick, Level: level, ActorPlayerID: actor, Changes: []Change{}, Events: []Event{}}
		resultByKey[key] = res
		ordered = append(ordered, res)
		return res
	}

	for _, in := range inputs {
		ps := s.playerForInput(in)
		if ps == nil || !ps.Connected {
			res := resultFor(s.currentLevel, 0)
			res.reject(in.MessageID, "unknown_actor")
			continue
		}
		s.usePlayer(ps)
		res := resultFor(ps.CurrentLevel, ps.PlayerID)
		if in.Type == "descend_intent" || in.Type == "ascend_intent" || in.Type == "teleport_intent" {
			if arrival := s.handleLevelTravel(in, res); arrival != nil {
				arrival.ActorPlayerID = ps.PlayerID
				ordered = append(ordered, arrival)
				transitionThisTick = true
			}
			s.savePlayer(ps)
			continue
		}
		s.withTickPhase(TickPhaseCombat, func() {
			s.applyInput(in, res)
		})
		s.savePlayer(ps)
	}

	if !transitionThisTick {
		for _, playerID := range sortedPlayerIDs(s.players) {
			ps := s.players[playerID]
			if ps == nil || !ps.Connected {
				continue
			}
			s.usePlayer(ps)
			res := resultFor(ps.CurrentLevel, ps.PlayerID)
			channelActive := false
			s.withTickPhase(TickPhaseCombat, func() {
				s.expireSkillEffects(res)
				s.advanceRogueMarks(res)
				s.advancePoisonDots(res)
				s.advanceUniqueBurnDots(res)
				s.advanceOffensiveUniqueEffectStates(res)
				channelActive = s.applyActiveSkillChannel(res)
			})
			if !channelActive {
				s.applyMovement(res)
			}
			s.applyPlayerRegen(res)
			s.savePlayer(ps)
		}

		s.withTickPhase(TickPhaseCombat, func() {
			s.advanceAreaHealZones(resultFor)
			s.autoPickUpCurrencyLoot(resultFor)
		})

		for _, levelNum := range s.sortedLevelNums() {
			s.currentLevel = levelNum
			s.syncCompatibilityFields()
			res := resultFor(levelNum, 0)
			s.withTickPhase(TickPhaseAI, func() {
				s.advanceMonsterMovement(res)
				s.advanceCompanions(res)
			})
			s.withTickPhase(TickPhaseCombat, func() {
				s.advanceBossPhases(res)
				s.advanceMonsterAttack(res)
				s.advanceProjectiles(res)
			})
		}
	}

	s.tick++
	s.usePlayer(s.defaultPlayer())

	results := make([]TickResult, 0, len(ordered))
	for _, res := range ordered {
		if len(res.Changes) == 0 && len(res.Events) == 0 && len(res.Acks) == 0 && len(res.Rejects) == 0 {
			continue
		}
		results = append(results, *res)
	}
	if len(results) == 0 {
		return []TickResult{{Tick: s.tick - 1, Level: s.currentLevel, Changes: []Change{}, Events: []Event{}}}
	}
	return results
}
