package inputdecode

import (
	"encoding/json"
	"strconv"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
)

// TypeSystemAccountStashSync is the stored-only, server-authored account stash replacement (v488).
// Like TypeSystemLoadShed it is absent from IsClientIntent and Decode.
const TypeSystemAccountStashSync = game.SystemAccountStashSyncInputType

type stashSyncItemWire struct {
	StashItemID string          `json:"stash_item_id"`
	ItemDefID   string          `json:"item_def_id"`
	RolledStats json.RawMessage `json:"rolled_stats,omitempty"`
}

type stashSyncPayloadWire struct {
	PlayerEntityID string              `json:"player_entity_id"`
	Gold           int                 `json:"gold"`
	Items          []stashSyncItemWire `json:"items"`
}

// EncodeStoredAccountStashSync builds the persisted envelope for an account stash sync.
func EncodeStoredAccountStashSync(messageID string, sync game.AccountStashSync) (json.RawMessage, error) {
	items := make([]stashSyncItemWire, 0, len(sync.Items))
	for _, item := range sync.Items {
		items = append(items, stashSyncItemWire{StashItemID: item.StashItemID, ItemDefID: item.ItemDefID, RolledStats: item.RolledStats})
	}
	payload, err := json.Marshal(stashSyncPayloadWire{
		PlayerEntityID: strconv.FormatUint(sync.PlayerID, 10),
		Gold:           sync.Gold,
		Items:          items,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(envelope{Type: TypeSystemAccountStashSync, MessageID: messageID, Payload: payload})
}

func decodeStoredAccountStashSync(env envelope) (game.Input, bool) {
	var p stashSyncPayloadWire
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		return game.Input{}, false
	}
	playerID, err := strconv.ParseUint(p.PlayerEntityID, 10, 64)
	if err != nil || playerID == 0 {
		return game.Input{}, false
	}
	items := make([]game.PersistedStashItem, 0, len(p.Items))
	for _, item := range p.Items {
		items = append(items, game.PersistedStashItem{StashItemID: item.StashItemID, ItemDefID: item.ItemDefID, RolledStats: item.RolledStats})
	}
	return game.Input{
		MessageID: env.MessageID,
		Type:      env.Type,
		StashSync: &game.AccountStashSync{PlayerID: playerID, Items: items, Gold: p.Gold},
	}, true
}
