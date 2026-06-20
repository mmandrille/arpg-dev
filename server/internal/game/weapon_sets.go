package game

func newWeaponSetMaps() []map[string]uint64 {
	sets := make([]map[string]uint64, weaponSetCount)
	for i := 0; i < weaponSetCount; i++ {
		sets[i] = map[string]uint64{
			mainHandSlot: 0,
			offHandSlot:  0,
		}
	}
	return sets
}

func cloneWeaponSetMaps(in []map[string]uint64) []map[string]uint64 {
	out := newWeaponSetMaps()
	for i := 0; i < len(in) && i < weaponSetCount; i++ {
		for _, slot := range []string{mainHandSlot, offHandSlot} {
			out[i][slot] = in[i][slot]
		}
	}
	return out
}

func normalizeWeaponSetIndex(index int) int {
	if index < 0 || index >= weaponSetCount {
		return defaultWeaponSet
	}
	return index
}

func (s *Sim) ensureWeaponSets() {
	if len(s.weaponSets) != weaponSetCount {
		s.weaponSets = newWeaponSetMaps()
	}
	for i := 0; i < weaponSetCount; i++ {
		if s.weaponSets[i] == nil {
			s.weaponSets[i] = map[string]uint64{}
		}
		for _, slot := range []string{mainHandSlot, offHandSlot} {
			if _, ok := s.weaponSets[i][slot]; !ok {
				s.weaponSets[i][slot] = 0
			}
		}
	}
	s.activeWeaponSet = normalizeWeaponSetIndex(s.activeWeaponSet)
	if s.equipped == nil {
		s.equipped = newEquippedMap()
	}
}

func (s *Sim) syncActiveWeaponSetToEquipped() {
	s.ensureWeaponSets()
	s.equipped[mainHandSlot] = s.weaponSets[s.activeWeaponSet][mainHandSlot]
	s.equipped[offHandSlot] = s.weaponSets[s.activeWeaponSet][offHandSlot]
}

func (s *Sim) syncEquippedHandsToActiveWeaponSet() {
	s.ensureWeaponSets()
	s.weaponSets[s.activeWeaponSet][mainHandSlot] = s.equipped[mainHandSlot]
	s.weaponSets[s.activeWeaponSet][offHandSlot] = s.equipped[offHandSlot]
}

func (s *Sim) setEquippedSlot(slot string, instanceID uint64, weaponSet int) {
	if isHandSlot(slot) {
		s.ensureWeaponSets()
		s.weaponSets[normalizeWeaponSetIndex(weaponSet)][slot] = instanceID
		s.syncActiveWeaponSetToEquipped()
		return
	}
	s.equipped[slot] = instanceID
}

func (s *Sim) equippedSlot(slot string, weaponSet int) uint64 {
	if isHandSlot(slot) {
		s.ensureWeaponSets()
		return s.weaponSets[normalizeWeaponSetIndex(weaponSet)][slot]
	}
	if s.equipped == nil {
		return 0
	}
	return s.equipped[slot]
}

func (s *Sim) clearEquippedItem(instanceID uint64) []string {
	cleared := map[string]bool{}
	for _, slot := range sortedStringKeys(s.equipped) {
		if !isHandSlot(slot) && s.equipped[slot] == instanceID {
			s.equipped[slot] = 0
			cleared[slot] = true
		}
	}
	s.ensureWeaponSets()
	for _, slot := range []string{mainHandSlot, offHandSlot} {
		if s.equipped[slot] == instanceID {
			s.weaponSets[s.activeWeaponSet][slot] = 0
			cleared[slot] = true
		}
	}
	for setIndex := range s.weaponSets {
		for _, slot := range []string{mainHandSlot, offHandSlot} {
			if s.weaponSets[setIndex][slot] == instanceID {
				s.weaponSets[setIndex][slot] = 0
				if setIndex == s.activeWeaponSet {
					cleared[slot] = true
				}
			}
		}
	}
	s.syncActiveWeaponSetToEquipped()
	return sortedStringKeys(cleared)
}

func (s *Sim) itemEquippedInAnySlot(instanceID uint64) bool {
	if instanceID == 0 {
		return false
	}
	s.ensureWeaponSets()
	for _, slot := range sortedStringKeys(s.equipped) {
		if s.equipped[slot] == instanceID {
			return true
		}
	}
	for setIndex := range s.weaponSets {
		for _, slot := range []string{mainHandSlot, offHandSlot} {
			if s.weaponSets[setIndex][slot] == instanceID {
				return true
			}
		}
	}
	return false
}

func (s *Sim) weaponSetForEquipIntent(intent *EquipIntent) int {
	if intent == nil || intent.WeaponSet == nil {
		return s.activeWeaponSet
	}
	return normalizeWeaponSetIndex(*intent.WeaponSet)
}

func (s *Sim) weaponSetForUnequipIntent(intent *UnequipIntent) int {
	if intent == nil || intent.WeaponSet == nil {
		return s.activeWeaponSet
	}
	return normalizeWeaponSetIndex(*intent.WeaponSet)
}

func (s *Sim) weaponSetViews() []WeaponSetView {
	s.ensureWeaponSets()
	return weaponSetViewsFromMaps(s.weaponSets)
}

func weaponSetViewsFromMaps(sets []map[string]uint64) []WeaponSetView {
	out := make([]WeaponSetView, 0, weaponSetCount)
	for i := 0; i < weaponSetCount; i++ {
		out = append(out, WeaponSetView{
			Index:    i,
			MainHand: stringPtrFromUint(sets[i][mainHandSlot]),
			OffHand:  stringPtrFromUint(sets[i][offHandSlot]),
		})
	}
	return out
}

func stringPtrFromUint(id uint64) *string {
	if id == 0 {
		return nil
	}
	out := idStr(id)
	return &out
}
