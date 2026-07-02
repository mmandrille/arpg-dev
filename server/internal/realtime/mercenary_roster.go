package realtime

import (
	"context"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/mercenaryroster"
)

func (h *Hub) loadMercenaryRosterIntoSim(ctx context.Context, sim *game.Sim, accountID, activeCharacterID string) error {
	return mercenaryroster.LoadIntoSim(ctx, h.store, h.rules, sim, accountID, activeCharacterID)
}
