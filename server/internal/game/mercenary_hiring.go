package game

const (
	mercenaryService           = "mercenary"
	mercenaryGuardOfferID      = "fixed:mercenary_guard"
	mercenaryGuardMonsterDefID = "mercenary_guard"
	mercenaryScoutOfferID      = "fixed:mercenary_scout"
	mercenaryScoutMonsterDefID = "mercenary_scout"
	mercenaryHireSourceID      = "mercenary_hire"
)

func (s *Sim) mercenaryHireCostGold() int {
	return s.rules.MainConfig.Gameplay.MercenaryHireCostGold
}

func (s *Sim) selectedMercenaryOffer(board *entity) (MercenaryOfferDef, bool) {
	if board == nil {
		return MercenaryOfferDef{}, false
	}
	return s.rules.Mercenaries.SelectOffer(s.seed, idStr(board.id))
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
	offer, ok := s.selectedMercenaryOffer(board)
	if !ok {
		res.reject(in.MessageID, "invalid_offer")
		return
	}

	cost := s.mercenaryHireCostGold()
	affordable := s.gold >= cost
	res.Events = append(res.Events, Event{
		EventType:     "mercenary_board_opened",
		EntityID:      idStr(board.id),
		CorrelationID: in.CorrelationID,
		Service:       mercenaryService,
		OfferID:       offer.OfferID,
		MonsterDefID:  offer.MonsterDefID,
		Price:         intPtr(cost),
		Affordable:    boolPtr(affordable),
		TotalGold:     intPtr(s.gold),
	})
	if !affordable {
		res.reject(in.MessageID, "not_enough_gold")
		return
	}

	s.gold -= cost
	s.progression.Gold = s.gold
	res.Changes = append(res.Changes, Change{Op: OpGoldUpdate, Gold: intPtr(s.gold)})
	s.appendCharacterProgressionUpdate(res)

	companion := s.spawnHiredMercenary(player, offer, res)
	if companion == nil {
		res.reject(in.MessageID, "invalid_offer")
		return
	}
	res.Events = append(res.Events, Event{
		EventType:      "mercenary_hired",
		EntityID:       idStr(board.id),
		TargetEntityID: idStr(companion.id),
		CorrelationID:  in.CorrelationID,
		Service:        mercenaryService,
		OfferID:        offer.OfferID,
		MonsterDefID:   offer.MonsterDefID,
		Price:          intPtr(cost),
		TotalGold:      intPtr(s.gold),
	})
	if ack {
		res.ack(in.MessageID)
	}
	s.savePlayer(s.defaultPlayer())
}

func (s *Sim) spawnHiredMercenary(owner *entity, offer MercenaryOfferDef, res *TickResult) *entity {
	if owner == nil {
		return nil
	}
	level := s.activeLevel()
	if level == nil {
		return nil
	}
	def, ok := s.rules.Monsters[offer.MonsterDefID]
	if !ok {
		return nil
	}
	s.pruneCompanionsForNewSpawn(owner.id, mercenaryHireSourceID, 1, res)
	companion := &entity{
		kind:                  companionEntity,
		pos:                   s.companionSpawnPosition(owner),
		spawnPos:              owner.pos,
		hp:                    def.MaxHP,
		maxHP:                 def.MaxHP,
		ownerID:               owner.id,
		monsterDefID:          offer.MonsterDefID,
		lootTable:             def.LootTable,
		speed:                 def.MoveSpeed,
		monsterAttackDamage:   def.AttackDamage,
		monsterAttackCooldown: def.AttackCooldown,
		monsterHitChance:      def.effectiveHitChance(s.rules.Combat),
		monsterCritChance:     def.effectiveCritChance(s.rules.Combat),
		monsterBlockPercent:   float64(def.BlockPercent),
		monsterArmor:          float64(def.Armor),
		aiMode:                monsterAIModeIdle,
		sourceSkillID:         mercenaryHireSourceID,
	}
	companion.id = s.alloc()
	level.entities[companion.id] = companion
	res.Changes = append(res.Changes, Change{Op: OpEntitySpawn, Entity: ptrEntityView(s.entityView(companion))})
	return companion
}
