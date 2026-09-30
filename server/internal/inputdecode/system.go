package inputdecode

import (
	"encoding/json"
	"strconv"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
)

// TypeSystemLoadShed is the stored-only, server-authored load-shedding input.
// It is deliberately absent from IsClientIntent and Decode: clients can never
// submit it, only the persisted input stream can carry it.
const TypeSystemLoadShed = game.SystemLoadShedInputType

// Stored-only, server-authored co-op membership inputs (v479). Like
// TypeSystemLoadShed they are absent from IsClientIntent and Decode.
const (
	TypeSystemMemberJoin   = game.SystemMemberJoinInputType
	TypeSystemMemberLeave  = game.SystemMemberLeaveInputType
	TypeSystemMemberRejoin = game.SystemMemberRejoinInputType
)

// TypeSystemTickCheckpoint is the stored-only, payload-free marker that live
// finished a tick (v481).
const TypeSystemTickCheckpoint = game.SystemTickCheckpointInputType

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

// memberLifecyclePayloadWire carries the player entity ID as a string, like
// every other entity ID on the wire.
type memberLifecyclePayloadWire struct {
	PlayerEntityID string `json:"player_entity_id"`
	AccountID      string `json:"account_id,omitempty"`
	CharacterID    string `json:"character_id,omitempty"`
	Respawn        bool   `json:"respawn,omitempty"`
}

// EncodeStoredMemberLifecycle builds the persisted envelope for a membership
// change. typ must be one of the TypeSystemMember* constants.
func EncodeStoredMemberLifecycle(typ, messageID string, m game.MemberLifecycle) (json.RawMessage, error) {
	payload, err := json.Marshal(memberLifecyclePayloadWire{
		PlayerEntityID: strconv.FormatUint(m.PlayerID, 10),
		AccountID:      m.AccountID,
		CharacterID:    m.CharacterID,
		Respawn:        m.Respawn,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(envelope{Type: typ, MessageID: messageID, Payload: payload})
}

// EncodeStoredTickCheckpoint builds the persisted envelope for a tick
// checkpoint. It has no payload fields.
func EncodeStoredTickCheckpoint(messageID string) (json.RawMessage, error) {
	return json.Marshal(envelope{Type: TypeSystemTickCheckpoint, MessageID: messageID, Payload: json.RawMessage(`{}`)})
}

func decodeStoredSystem(env envelope) (game.Input, bool) {
	if env.Type == TypeSystemTickCheckpoint {
		return game.Input{MessageID: env.MessageID, Type: env.Type}, true
	}
	if game.IsMemberLifecycleInput(game.Input{Type: env.Type}) {
		return decodeStoredMemberLifecycle(env)
	}
	if env.Type == TypeSystemAccountStashSync {
		return decodeStoredAccountStashSync(env)
	}
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

func decodeStoredMemberLifecycle(env envelope) (game.Input, bool) {
	var p memberLifecyclePayloadWire
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		return game.Input{}, false
	}
	playerID, err := strconv.ParseUint(p.PlayerEntityID, 10, 64)
	if err != nil || playerID == 0 {
		return game.Input{}, false
	}
	return game.Input{
		MessageID: env.MessageID,
		Type:      env.Type,
		Member: &game.MemberLifecycle{
			PlayerID:    playerID,
			AccountID:   p.AccountID,
			CharacterID: p.CharacterID,
			Respawn:     p.Respawn,
		},
	}, true
}
