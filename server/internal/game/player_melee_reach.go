package game

func (s *Sim) playerWeaponSlotReach(slot string) float64 {
	instanceID := s.equipped[slot]
	if instanceID == 0 {
		if slot == mainHandSlot {
			return s.rules.Combat.UnarmedReach
		}

		return 0
	}
	item := s.findItemByID(instanceID)
	if item == nil {
		if slot == mainHandSlot {
			return s.rules.Combat.UnarmedReach
		}

		return 0
	}
	reach, ok := s.itemReach(item)
	if !ok {
		if slot == mainHandSlot {
			return s.rules.Combat.UnarmedReach
		}

		return 0
	}

	return reach
}

func (s *Sim) playerDualWieldMeleeReach() float64 {
	mainReach := s.playerWeaponSlotReach(mainHandSlot)
	item := s.findItemByID(s.equipped[offHandSlot])
	if !s.canOffhandWeapon(item) {
		return mainReach
	}
	offReach, ok := s.itemReach(item)
	if !ok || offReach >= mainReach {
		return mainReach
	}

	return offReach
}

func (s *Sim) inWeaponSlotMeleeRange(target *entity, slot string) bool {
	player := s.activeLevel().entities[s.playerID]
	if player == nil || target == nil {
		return false
	}
	reach := s.playerWeaponSlotReach(slot)
	if slot == offHandSlot && reach <= 0 {
		return false
	}

	return meleeInRange(distance(player.pos, target.pos), reach, s.targetInteractionRadius(target))
}

func (s *Sim) emitPlayerWeaponMiss(target *entity, playerID uint64, corr string, res *TickResult, weaponSlot string) {
	outcome := combatResolution{Outcome: "miss", Hit: false}
	event := combatEvent(s.combatEventType(monsterEntity, outcome), playerID, target.id, corr, outcome)
	event.WeaponSlot = weaponSlot
	res.Events = append(res.Events, event)
	s.aggroMonsterOnHit(target, playerID, corr, res)
}
