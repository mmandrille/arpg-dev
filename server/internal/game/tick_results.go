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
	ctx := newSimTickCtx(s)

	for _, in := range inputs {
		ps := s.playerForInput(in)
		if ps == nil || !ps.Connected {
			res := ctx.resultFor(s.currentLevel, 0)
			res.reject(in.MessageID, "unknown_actor")
			continue
		}
		s.usePlayer(ps)
		res := ctx.resultFor(ps.CurrentLevel, ps.PlayerID)
		if in.Type == "descend_intent" || in.Type == "ascend_intent" || in.Type == "teleport_intent" {
			if arrival := s.handleLevelTravel(in, res); arrival != nil {
				arrival.ActorPlayerID = ps.PlayerID
				ctx.ordered = append(ctx.ordered, arrival)
				ctx.markTransition()
			}
			s.savePlayer(ps)
			continue
		}
		s.withTickPhase(TickPhaseCombat, func() {
			s.applyInput(in, res)
		})
		s.savePlayer(ps)
	}

	if !ctx.transitionThisTick {
		for _, playerID := range sortedPlayerIDs(s.players) {
			ps := s.players[playerID]
			if ps == nil || !ps.Connected {
				continue
			}
			s.usePlayer(ps)
			res := ctx.resultFor(ps.CurrentLevel, ps.PlayerID)
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
			s.advanceAreaHealZones(ctx.resultFor)
			s.autoPickUpCurrencyLoot(ctx.resultFor)
		})

		for _, levelNum := range s.sortedLevelNums() {
			s.currentLevel = levelNum
			s.syncCompatibilityFields()
			res := ctx.resultFor(levelNum, 0)
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

	return ctx.finalizeResults()
}
