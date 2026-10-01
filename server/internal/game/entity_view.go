package game

// Entity wire views: the authoritative entity state projected onto EntityView
// for snapshots and deltas. View-only; nothing here mutates sim state.

func (s *Sim) entityView(e *entity) EntityView {
	if e == nil {
		return EntityView{}
	}
	view := e.view()
	if e.kind == playerEntity {
		view.CharacterClass = s.playerEntityClass(e)
	}
	if e.kind == monsterEntity || e.kind == companionEntity {
		view.EffectIDs = sortedUniqueStrings(append(cloneStringSlice(e.effectIDs), s.eliteAuraEffectIDs(e)...))
	}
	if e.kind == companionEntity {
		view.CombatStats = s.companionCombatStatsView(e)
		if e.expiresTick > 0 && e.totalDurationTicks > 0 {
			view.RemainingTicks, view.TotalTicks = companionDurationTicks(e, s.tick)
		}
	}
	if e.kind == interactableEntity {
		if level := s.activeLevel(); level != nil {
			view.EliteObjective = level.eliteObjectiveChestIDs[e.id]
			view.QuestReward = level.questRewardChestIDs[e.id]
		}
	}
	if e.kind != lootEntity {
		return view
	}
	s.annotateItemRequirementStatus(view.Requirements, s.lootItemForRequirements(e), func(status []RequirementStatusView, met *bool) {
		view.RequirementStatus = status
		view.RequirementsMet = met
	})
	if e.rollPayload != nil {
		s.annotateClassAffinityStatus(e.rollPayload, func(status []ClassAffinityStatusView) {
			view.ClassAffinityStatus = status
		})
		s.annotateSkillBonusStatus(e.rollPayload, func(status []SkillBonusStatusView) {
			view.SkillBonusStatus = status
		})
	}
	if preview := s.equipPreviewForLoot(e); preview != nil {
		view.EquipPreview = preview
	}
	return view
}

func (e *entity) view() EntityView {
	ev := EntityView{ID: idStr(e.id), Type: e.kind, Position: e.pos}
	switch e.kind {
	case playerEntity, monsterEntity, companionEntity:
		hp, maxHP := e.hp, e.maxHP
		ev.HP = &hp
		ev.MaxHP = &maxHP
		if e.kind == playerEntity {
			mana, maxMana := e.mana, e.maxMana
			ev.Mana = &mana
			ev.MaxMana = &maxMana
			ev.CharacterID = e.characterID
			ev.DisplayName = e.displayName
			ev.EffectIDs = cloneStringSlice(e.effectIDs)
			if e.visualScale > 0 {
				ev.VisualScale = e.visualScale
			}
		}
		if e.kind == monsterEntity || e.kind == companionEntity {
			e.applyMonsterLikeViewFields(&ev)
		}
	case lootEntity:
		ev.ItemDefID = e.itemDefID
		if e.goldAmount > 0 {
			amount := e.goldAmount
			ev.Amount = &amount
		}
		if e.rollPayload != nil {
			ev.ItemDefID = e.rollPayload.ItemTemplateID
			ev.ItemTemplateID = e.rollPayload.ItemTemplateID
			ev.DisplayName = e.rollPayload.DisplayName
			ev.Rarity = e.rollPayload.Rarity
			ev.ItemLevel = e.rollPayload.ItemLevel
			ev.RolledStats = cloneIntMap(e.rollPayload.Stats)
			ev.Requirements = cloneIntMap(e.rollPayload.Requirements)
			ev.EffectIDs = cloneStringSlice(e.rollPayload.EffectIDs)
		}
	case interactableEntity:
		ev.InteractableDefID = e.interactableDefID
		ev.State = e.state
		if e.interactableDefID == heroCorpseDefID {
			ev.CorpseCharacterID = e.corpseCharacterID
			ev.CorpseName = e.corpseName
			if e.corpseLevel > 0 {
				ev.CorpseLevel = e.corpseLevel
			}
			count := e.corpseItemCount
			ev.CorpseItemCount = &count
		}
	case projectileEntity:
		ev.OwnerID = idStr(e.ownerID)
		if e.targetID != 0 {
			ev.TargetID = idStr(e.targetID)
		}
		ev.ProjectileDefID = e.projectileDefID
	}
	return ev
}

func (e *entity) bossPhaseView() *BossPhaseView {
	view := &BossPhaseView{
		PatternID:     e.bossPatternID,
		PhaseIndex:    e.bossPhaseIndex,
		PhaseKind:     e.bossPhaseKind,
		StartedTick:   e.bossPhaseStarted,
		DurationTicks: int(e.bossPhaseEnds - e.bossPhaseStarted),
		Lane:          e.bossLane.withPhase(e.bossPhaseIndex, BossPatternPhase{Kind: e.bossPhaseKind}),
	}
	if e.bossLane != nil && e.bossPhaseKind == "telegraph" {
		view.Telegraph = &BossTelegraphView{Type: "lanes", ToColor: e.bossLane.DangerColor, HitShape: "lanes"}
	}
	return view
}

// playerEntityClass reads the owning member's class from live progression so
// co-op partners render as their own hero. Player entity ids are PlayerIDs. The
// active member's progression lives in s.progression (playerState.Progression
// is only refreshed by savePlayer); every other member's lives on its
// playerState.
func (s *Sim) playerEntityClass(e *entity) string {
	if e.id == s.playerID {
		return s.progression.CharacterClass
	}
	if ps := s.players[e.id]; ps != nil {
		return ps.Progression.CharacterClass
	}
	return ""
}
