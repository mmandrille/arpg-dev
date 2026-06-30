package game

type bishopDebugAction struct {
	BishopEntityID string
	EventType      string
	Apply          func() (int, bool)
}

func (s *Sim) handleBishopDebugLevel(in Input, res *TickResult) {
	if in.BishopDebugLevel == nil {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	s.handleBishopDebugAction(in, res, bishopDebugAction{
		BishopEntityID: in.BishopDebugLevel.BishopEntityID,
		EventType:      "bishop_debug_level_gained",
		Apply: func() (int, bool) {
			return s.debugGrantSingleLevel()
		},
	})
}

func (s *Sim) handleBishopDebugSkillPoint(in Input, res *TickResult) {
	if in.BishopDebugSkill == nil {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	s.handleBishopDebugAction(in, res, bishopDebugAction{
		BishopEntityID: in.BishopDebugSkill.BishopEntityID,
		EventType:      "bishop_debug_skill_point_gained",
		Apply: func() (int, bool) {
			s.progression.UnspentSkillPoints++
			return 1, true
		},
	})
}

func (s *Sim) handleBishopDebugStatPoint(in Input, res *TickResult) {
	if in.BishopDebugStat == nil {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	s.handleBishopDebugAction(in, res, bishopDebugAction{
		BishopEntityID: in.BishopDebugStat.BishopEntityID,
		EventType:      "bishop_debug_stat_point_gained",
		Apply: func() (int, bool) {
			s.progression.UnspentStatPoints++
			return 1, true
		},
	})
}

func (s *Sim) handleBishopDebugDropUpgradeShard(in Input, res *TickResult) {
	if !s.gameplayDebug {
		res.reject(in.MessageID, "debug_disabled")
		return
	}
	if in.BishopDebugDropUpgradeShard == nil || in.BishopDebugDropUpgradeShard.BishopEntityID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	bishopEntity, ok, reason := s.resolveBishopIntentTarget(in.BishopDebugDropUpgradeShard.BishopEntityID)
	if !ok {
		res.reject(in.MessageID, reason)
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil || player.hp <= 0 {
		res.reject(in.MessageID, "player_dead")
		return
	}

	depth := maxInt(1, s.progression.DeepestDungeonDepth)
	lootID, level, ok := s.spawnUpgradeShardLoot(player.pos, s.targetInteractionRadius(player), depth, in.CorrelationID, res)
	if !ok {
		res.reject(in.MessageID, "drop_failed")
		return
	}

	healed, restored := s.restorePlayerResources(player, res)
	res.Events = append(res.Events, Event{
		EventType:      "bishop_debug_upgrade_shard_dropped",
		EntityID:       idStr(bishopEntity.id),
		TargetEntityID: idStr(lootID),
		CorrelationID:  in.CorrelationID,
		Service:        "bishop",
		Amount:         intPtr(level),
		Heal:           intPtr(healed),
		Mana:           intPtr(restored),
	})
	res.ack(in.MessageID)
	s.savePlayer(s.defaultPlayer())
}

func (s *Sim) handleBishopDebugAction(in Input, res *TickResult, action bishopDebugAction) {
	if !s.gameplayDebug {
		res.reject(in.MessageID, "debug_disabled")
		return
	}
	if action.BishopEntityID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	bishopEntity, ok, reason := s.resolveBishopIntentTarget(action.BishopEntityID)
	if !ok {
		res.reject(in.MessageID, reason)
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil || player.hp <= 0 {
		res.reject(in.MessageID, "player_dead")
		return
	}

	fromLevel := s.progression.Level
	amount, ok := action.Apply()
	if !ok {
		res.reject(in.MessageID, "level_cap")
		return
	}
	player.maxHP = s.currentMaxHP()
	player.maxMana = s.currentMaxMana()
	healed, restored := s.restorePlayerResources(player, res)
	s.appendProgressionAndSkillUpdates(res)
	res.Events = append(res.Events, Event{
		EventType:          action.EventType,
		EntityID:           idStr(bishopEntity.id),
		CorrelationID:      in.CorrelationID,
		Service:            "bishop",
		Heal:               intPtr(healed),
		Mana:               intPtr(restored),
		FromLevel:          intPtr(fromLevel),
		ToLevel:            intPtr(s.progression.Level),
		Amount:             intPtr(amount),
		TotalExperience:    intPtr(s.progression.Experience),
		UnspentStatPoints:  intPtr(s.progression.UnspentStatPoints),
		UnspentSkillPoints: intPtr(s.progression.UnspentSkillPoints),
	})
	res.ack(in.MessageID)
	s.savePlayer(s.defaultPlayer())
}

func (s *Sim) debugGrantSingleLevel() (int, bool) {
	if s.progression.Level >= s.rules.CharacterProgression.LevelCap {
		return 0, false
	}
	nextXP, ok := s.rules.nextLevelTotalXP(s.progression.Level)
	if !ok {
		return 0, false
	}
	xpGranted := nextXP - s.progression.Experience
	if xpGranted < 0 {
		xpGranted = 0
	}
	s.progression.Experience += xpGranted
	s.progression.Level++
	s.progression.UnspentStatPoints += s.rules.CharacterProgression.PointsPerLevel
	if gained := s.rules.skillPointsGrantedAtLevel(s.progression.Level); gained > 0 {
		s.progression.UnspentSkillPoints += gained
	}
	return xpGranted, true
}
