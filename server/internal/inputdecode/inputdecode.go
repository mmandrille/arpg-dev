// Package inputdecode converts protocol intent envelopes into game inputs
// without depending on the realtime WebSocket runner.
package inputdecode

import (
	"encoding/json"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
)

const (
	TypeClientReady = "client_ready"
	TypeMoveIntent  = "move_intent"
	TypeAttack      = "attack_intent"
	TypePickUp      = "pick_up_intent"
	TypeEquip       = "equip_intent"
)

type envelope struct {
	Type          string          `json:"type"`
	MessageID     string          `json:"message_id"`
	CorrelationID string          `json:"correlation_id"`
	Payload       json.RawMessage `json:"payload"`
}

type (
	movePayloadWire struct {
		Direction     game.Vec2 `json:"direction"`
		DurationTicks int       `json:"duration_ticks"`
	}
	attackPayloadWire struct {
		TargetID string `json:"target_id"`
	}
	pickUpPayloadWire struct {
		EntityID string `json:"entity_id"`
	}
	equipPayloadWire struct {
		ItemInstanceID string `json:"item_instance_id"`
		Slot           string `json:"slot"`
	}
)

// IsClientIntent reports whether the type is a buffered authoritative intent.
func IsClientIntent(t string) bool {
	switch t {
	case TypeMoveIntent, TypeAttack, TypePickUp, TypeEquip:
		return true
	}
	return false
}

// Decode converts an intent type and payload into a sim Input. It returns
// ok=false if the payload does not match the message type.
func Decode(typ, messageID, correlationID string, payload json.RawMessage) (game.Input, bool) {
	in := game.Input{
		MessageID:     messageID,
		CorrelationID: correlationID,
		Type:          typ,
	}
	switch typ {
	case TypeMoveIntent:
		var p movePayloadWire
		if err := json.Unmarshal(payload, &p); err != nil {
			return in, false
		}
		in.Move = &game.MoveIntent{Direction: p.Direction, DurationTicks: p.DurationTicks}
	case TypeAttack:
		var p attackPayloadWire
		if err := json.Unmarshal(payload, &p); err != nil || p.TargetID == "" {
			return in, false
		}
		in.Attack = &game.AttackIntent{TargetID: p.TargetID}
	case TypePickUp:
		var p pickUpPayloadWire
		if err := json.Unmarshal(payload, &p); err != nil || p.EntityID == "" {
			return in, false
		}
		in.PickUp = &game.PickUpIntent{EntityID: p.EntityID}
	case TypeEquip:
		var p equipPayloadWire
		if err := json.Unmarshal(payload, &p); err != nil || p.ItemInstanceID == "" || p.Slot == "" {
			return in, false
		}
		in.Equip = &game.EquipIntent{ItemInstanceID: p.ItemInstanceID, Slot: p.Slot}
	default:
		return in, false
	}
	return in, true
}

// DecodeStored recovers a sim Input from a persisted input envelope. The caller
// should overwrite tick/sequence/message metadata from the durable row.
func DecodeStored(raw []byte) (game.Input, bool) {
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return game.Input{}, false
	}
	if !IsClientIntent(env.Type) {
		return game.Input{}, false
	}
	return Decode(env.Type, env.MessageID, env.CorrelationID, env.Payload)
}
