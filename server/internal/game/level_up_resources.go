package game

func (s *Sim) restorePlayerResourcesOnLevelUp(corr string, res *TickResult) {
	if !s.rules.CharacterProgression.LevelUpRestoreHPMana {
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil {
		return
	}
	s.syncActivePlayerMaxResources()
	heal := player.maxHP - player.hp
	mana := player.maxMana - player.mana
	if heal <= 0 && mana <= 0 {
		return
	}
	player.hp = player.maxHP
	player.mana = player.maxMana
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	if heal > 0 {
		res.Events = append(res.Events, Event{
			EventType:     "player_healed",
			EntityID:      idStr(player.id),
			CorrelationID: corr,
			Heal:          intPtr(heal),
			Reason:        "level_up",
		})
	}
	if mana > 0 {
		res.Events = append(res.Events, Event{
			EventType:     "player_mana_restored",
			EntityID:      idStr(player.id),
			CorrelationID: corr,
			Mana:          intPtr(mana),
			Reason:        "level_up",
		})
	}
}
