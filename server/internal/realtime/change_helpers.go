package realtime

import "github.com/mmandrille_meli/arpg-dev/server/internal/game"

func changeWeaponSet(c game.Change) int {
	if c.WeaponSet == nil {
		return 0
	}
	return *c.WeaponSet
}

func changeRequiresExplicitWeaponSet(c game.Change) bool {
	slot := c.Slot
	if c.Item != nil {
		slot = c.Item.Slot
	}
	return slot == "main_hand" || slot == "off_hand"
}

func changeHasExplicitWeaponSet(c game.Change) bool {
	return c.WeaponSet != nil
}
