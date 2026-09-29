package inputdecode

import (
	"encoding/json"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
)

// TypeSystemLoadShed is the stored-only, server-authored load-shedding input.
// It is deliberately absent from IsClientIntent and Decode: clients can never
// submit it, only the persisted input stream can carry it.
const TypeSystemLoadShed = game.SystemLoadShedInputType

type loadShedPayloadWire struct {
	OverloadDegrade        bool `json:"overload_degrade"`
	CombatMovementThrottle bool `json:"combat_movement_throttle"`
}

// EncodeStoredLoadShed builds the persisted envelope for a load-shed directive.
func EncodeStoredLoadShed(messageID string, d game.LoadShedDirective) (json.RawMessage, error) {
	payload, err := json.Marshal(loadShedPayloadWire{
		OverloadDegrade:        d.OverloadDegrade,
		CombatMovementThrottle: d.CombatMovementThrottle,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(envelope{Type: TypeSystemLoadShed, MessageID: messageID, Payload: payload})
}

func decodeStoredSystem(env envelope) (game.Input, bool) {
	if env.Type != TypeSystemLoadShed {
		return game.Input{}, false
	}
	var p loadShedPayloadWire
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		return game.Input{}, false
	}
	return game.Input{
		MessageID: env.MessageID,
		Type:      env.Type,
		LoadShed: &game.LoadShedDirective{
			OverloadDegrade:        p.OverloadDegrade,
			CombatMovementThrottle: p.CombatMovementThrottle,
		},
	}, true
}
