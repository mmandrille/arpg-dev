package realtime

import (
	"context"
	"encoding/json"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

type deferredPersistChange struct {
	change game.Change
	member store.SessionMember
}

func persistChangeDeferrable(op string) bool {
	switch op {
	case game.OpCharacterProgressionUpdate, game.OpShopStockReplace, game.OpShopStockAvailability,
		game.OpResourceWalletUpdate, game.OpResourceBagItemAdd,
		game.OpInventoryAdd, game.OpInventoryUpdate, game.OpEquippedUpdate, game.OpGoldUpdate:
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
	case game.OpResourceWalletUpdate:
		if c.ResourceID == "" {
			return
		}
		if _, err := l.hub.store.AddAccountResource(ctx, member.AccountID, c.ResourceID, 1); err != nil {
			l.hub.metrics.PersistenceErrors.Inc()
			l.log.Error("persist account resource", "resource_id", c.ResourceID, "error", err)
		}
	case game.OpResourceBagItemAdd:
		if c.StashItem == nil {
			return
		}
		rolledStats := json.RawMessage(`{}`)
		if payload := c.StashItem.RollPayload(); payload != nil {
			if raw, err := json.Marshal(payload); err == nil {
				rolledStats = raw
			} else {
				l.hub.metrics.PersistenceErrors.Inc()
				l.log.Error("marshal resource bag item payload", "error", err)
			}
		}
		if _, err := l.hub.store.InsertAccountResourceBagItem(ctx, member.AccountID, member.CharacterID, c.StashItem.StashItemID, c.StashItem.ItemDefID, rolledStats); err != nil {
			l.hub.metrics.PersistenceErrors.Inc()
			l.log.Error("persist resource bag item add", "bag_item_id", c.StashItem.StashItemID, "error", err)
		}
	case game.OpInventoryAdd:
		l.persistInventoryAdd(ctx, c, member)
	case game.OpInventoryUpdate:
		l.persistInventoryUpdate(ctx, c, member)
	case game.OpEquippedUpdate:
		l.persistEquippedUpdate(ctx, c, member)
	case game.OpGoldUpdate:
		l.persistGoldUpdate(ctx, c, member)
	}
}

func (l *sessionLoop) persistInventoryAdd(ctx context.Context, c game.Change, member store.SessionMember) {
	if c.Item == nil {
		return
	}
	location := store.ItemLocationInventory
	if c.Item.Equipped {
		location = store.ItemLocationEquipped
	}
	rolledStats := json.RawMessage(`{}`)
	if payload := c.Item.RollPayload(); payload != nil {
		if raw, err := json.Marshal(payload); err == nil {
			rolledStats = raw
		} else {
			l.hub.metrics.PersistenceErrors.Inc()
			l.log.Error("marshal rolled item payload", "error", err)
		}
	}
	if err := l.hub.store.AddCharacterItem(ctx, store.CharacterItemInstance{
		ID:          c.Item.ItemInstanceID,
		AccountID:   member.AccountID,
		CharacterID: member.CharacterID,
		ItemDefID:   c.Item.ItemDefID,
		Location:    location,
		Slot:        c.Item.Slot,
		Equipped:    c.Item.Equipped,
		WeaponSet:   changeWeaponSet(c),
		RolledStats: rolledStats,
	}); err != nil {
		l.hub.metrics.PersistenceErrors.Inc()
		l.log.Error("persist inventory add", "error", err)
	}
}

func (l *sessionLoop) persistInventoryUpdate(ctx context.Context, c game.Change, member store.SessionMember) {
	if c.Item == nil {
		return
	}
	if changeRequiresExplicitWeaponSet(c) && !changeHasExplicitWeaponSet(c) {
		return
	}
	if err := l.hub.store.SetCharacterItemEquipped(ctx, member.AccountID, member.CharacterID, c.Item.ItemInstanceID, c.Item.Slot, c.Item.Equipped, changeWeaponSet(c)); err != nil {
		l.hub.metrics.PersistenceErrors.Inc()
		l.log.Error("persist inventory update", "error", err)
	}
}

func (l *sessionLoop) persistEquippedUpdate(ctx context.Context, c game.Change, member store.SessionMember) {
	if c.ItemInstanceID == nil || c.Slot == "" {
		return
	}
	if err := l.hub.store.SetCharacterItemEquipped(ctx, member.AccountID, member.CharacterID, *c.ItemInstanceID, c.Slot, true, changeWeaponSet(c)); err != nil {
		l.hub.metrics.PersistenceErrors.Inc()
		l.log.Error("persist equipped update", "error", err)
	}
}

func (l *sessionLoop) persistGoldUpdate(ctx context.Context, c game.Change, member store.SessionMember) {
	if c.Gold == nil {
		return
	}
	if err := l.hub.store.SetCharacterGold(ctx, member.AccountID, member.CharacterID, *c.Gold); err != nil {
		l.hub.metrics.PersistenceErrors.Inc()
		l.log.Error("persist character gold", "error", err)
	}
}
