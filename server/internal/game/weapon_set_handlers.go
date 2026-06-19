package game

func (s *Sim) handleSwapWeaponSet(in Input, res *TickResult) {
	if in.SwapWeaponSet == nil {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	s.ensureWeaponSets()
	s.activeWeaponSet = (s.activeWeaponSet + 1) % weaponSetCount
	s.syncActiveWeaponSetToEquipped()
	res.Changes = append(res.Changes,
		Change{Op: OpEquippedUpdate, Slot: mainHandSlot, ItemInstanceID: stringPtrFromUint(s.equipped[mainHandSlot]), WeaponSet: intPtr(s.activeWeaponSet)},
		Change{Op: OpEquippedUpdate, Slot: offHandSlot, ItemInstanceID: stringPtrFromUint(s.equipped[offHandSlot]), WeaponSet: intPtr(s.activeWeaponSet)},
		Change{Op: OpWeaponSetUpdate, ActiveWeaponSet: intPtr(s.activeWeaponSet), WeaponSets: s.weaponSetViews()},
	)
	res.Events = append(res.Events, Event{EventType: "weapon_set_swapped", CorrelationID: in.CorrelationID, WeaponSet: intPtr(s.activeWeaponSet)})
	res.ack(in.MessageID)
}
