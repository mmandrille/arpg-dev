package realtime

import (
	"context"
	"log/slog"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func (h *Hub) loadCharacterCorpses(ctx context.Context, log *slog.Logger, sim *game.Sim, member store.SessionMember) {
	corpses, err := h.store.ListRecoverableCharacterCorpses(ctx, member.AccountID, member.CharacterID)
	if err != nil {
		h.metrics.PersistenceErrors.Inc()
		log.Error("load character corpses", "account_id", member.AccountID, "character_id", member.CharacterID, "error", err)
		return
	}
	sim.LoadCharacterCorpses(persistedCorpsesWithAccount(member.AccountID, corpses))
}
