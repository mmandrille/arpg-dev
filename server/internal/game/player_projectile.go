package game

import "strings"

func (s *Sim) playerEquippedWeaponItemType() string {
	item := s.equippedWeaponItem()
	if item == nil {
		return ""
	}
	if item.rollPayload != nil {
		if template, ok := s.rules.ItemTemplates[item.rollPayload.ItemTemplateID]; ok {
			return template.ItemType
		}

		return ""
	}
	if strings.Contains(item.itemDefID, "staff") {
		return "staff"
	}
	if strings.Contains(item.itemDefID, "bow") {
		return "bow"
	}

	return ""
}

func (s *Sim) playerProjectileDefID() string {
	if s.playerEquippedWeaponItemType() == "staff" {
		return staffOrbProjectileDefID
	}

	return trainingArrowProjectileDefID
}
