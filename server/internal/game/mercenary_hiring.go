package game

const (
	mercenaryService      = "mercenary"
	mercenaryHireSourceID = "mercenary_hire"
)

func (s *Sim) mercenaryHireCostGold() int {
	perLevel := s.rules.MainConfig.Gameplay.MercenaryHireCostGoldPerLevel
	if perLevel <= 0 {
		perLevel = 10
	}
	level := s.progression.Level
	if level < 1 {
		level = 1
	}

	return perLevel * level
}

func (s *Sim) hireMercenaryFromBoard(board *entity, in Input, res *TickResult, ack bool) {
	if board == nil || board.kind != interactableEntity || s.serviceForInteractable(board) != mercenaryService {
		res.reject(in.MessageID, "invalid_target")
		return
	}
	if board.state != interactableReady {
		res.reject(in.MessageID, "not_actionable")
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil || player.hp <= 0 {
		res.reject(in.MessageID, "player_dead")
		return
	}

	cost := s.mercenaryHireCostGold()
	candidates := s.mercenaryCandidatesForPlayer(player)
	affordable := s.gold >= cost
	res.Events = append(res.Events, Event{
		EventType:             "mercenary_board_opened",
		EntityID:              idStr(board.id),
		CorrelationID:         in.CorrelationID,
		Service:               mercenaryService,
		Price:                 intPtr(cost),
		Affordable:            boolPtr(affordable),
		TotalGold:             intPtr(s.gold),
		MercenaryCandidates:   candidates,
	})

	characterID := ""
	if in.Action != nil {
		characterID = in.Action.MercenaryCharacterID
	}
	if characterID == "" {
		if ack {
			res.ack(in.MessageID)
		}

		return
	}
	if !affordable {
		res.reject(in.MessageID, "not_enough_gold")
		return
	}

	snap, ok := s.mercenaryRoster[characterID]
	if !ok || snap.Dead {
		res.reject(in.MessageID, "invalid_character")
		return
	}
	activeCharacterID := ""
	if ps := s.players[s.playerID]; ps != nil {
		activeCharacterID = ps.CharacterID
	}
	if snap.CharacterID == activeCharacterID {
		res.reject(in.MessageID, "invalid_character")
		return
	}

	s.gold -= cost
	s.progression.Gold = s.gold
	s.progression.HiredMercenaryCharacterID = snap.CharacterID
	res.Changes = append(res.Changes, Change{Op: OpGoldUpdate, Gold: intPtr(s.gold)})
	s.appendCharacterProgressionUpdate(res)

	companion := s.spawnCharacterMercenary(player, snap, res)
	if companion == nil {
		res.reject(in.MessageID, "invalid_character")
		return
	}
	res.Events = append(res.Events, Event{
		EventType:           "mercenary_hired",
		EntityID:            idStr(board.id),
		TargetEntityID:      idStr(companion.id),
		CorrelationID:       in.CorrelationID,
		Service:             mercenaryService,
		SourceCharacterID:   snap.CharacterID,
		CharacterClass:      snap.CharacterClass,
		Price:               intPtr(cost),
		TotalGold:           intPtr(s.gold),
	})
	if ack {
		res.ack(in.MessageID)
	}
	s.savePlayer(s.defaultPlayer())
}
