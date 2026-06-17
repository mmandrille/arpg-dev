package realtime

import "github.com/mmandrille_meli/arpg-dev/server/internal/game"

func changeWeaponSet(c game.Change) int {
	if c.WeaponSet == nil {
		return 0
	}
	return *c.WeaponSet
}
