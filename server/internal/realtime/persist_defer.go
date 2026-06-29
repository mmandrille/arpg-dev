package realtime

import (
	"context"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

type deferredPersistChange struct {
	change game.Change
	member store.SessionMember
}

func persistChangeDeferrable(op string) bool {
	switch op {
	case game.OpCharacterProgressionUpdate, game.OpShopStockReplace, game.OpShopStockAvailability:
		return true
	default:
		return false
	}
}

func (l *sessionLoop) flushDeferredPersist() {
	if len(l.deferredPersistChanges) == 0 {
		return
	}
	ctx := context.Background()
	pending := l.deferredPersistChanges
	l.deferredPersistChanges = nil
	for _, item := range pending {
		l.persistChange(ctx, item.change, item.member)
	}
}

func (l *sessionLoop) persistChange(ctx context.Context, c game.Change, member store.SessionMember) {
	switch c.Op {
	case game.OpCharacterProgressionUpdate:
		if c.Progression == nil {
			return
		}
		if err := l.hub.store.UpsertCharacterProgression(ctx, member.AccountID, storeProgressionFromView(member.AccountID, member.CharacterID, *c.Progression)); err != nil {
			l.hub.metrics.PersistenceErrors.Inc()
			l.log.Error("persist character progression", "error", err)
		}
	case game.OpShopStockReplace:
		if err := l.hub.store.ReplaceCharacterShopStock(ctx, member.AccountID, member.CharacterID, c.ShopID, c.RefreshKey, storeShopStock(member.AccountID, member.CharacterID, c.ShopStock)); err != nil {
			l.hub.metrics.PersistenceErrors.Inc()
			l.log.Error("persist shop stock replace", "shop_id", c.ShopID, "error", err)
		}
	case game.OpShopStockAvailability:
		if err := l.hub.store.SetCharacterShopStockAvailable(ctx, member.AccountID, member.CharacterID, c.ShopID, c.OfferID, c.Available); err != nil {
			l.hub.metrics.PersistenceErrors.Inc()
			l.log.Error("persist shop stock availability", "shop_id", c.ShopID, "offer_id", c.OfferID, "error", err)
		}
	}
}
