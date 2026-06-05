// Package realtime implements the authenticated WebSocket game protocol: the
// per-connection session runner that drives the 20 Hz authoritative tick loop,
// validates envelopes, applies buffered inputs in deterministic order, and
// persists inputs/events/inventory (ADR-0001 D2/D3/D8.1/D8.2).
package realtime

import (
	"encoding/json"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
)

// tickHz is the authoritative simulation rate (20 Hz, 50 ms/tick).
const tickHz = 20

// Client-to-server message types.
const (
	typeClientReady = "client_ready"
	typeMoveIntent  = "move_intent"
	typeAttack      = "attack_intent"
	typePickUp      = "pick_up_intent"
	typeEquip       = "equip_intent"
)

// Server-to-client message types.
const (
	typeSnapshot       = "session_snapshot"
	typeStateDelta     = "state_delta"
	typeIntentAccepted = "intent_accepted"
	typeIntentRejected = "intent_rejected"
	typeError          = "error"
)

// inEnvelope is an inbound message; the payload is decoded per type.
type inEnvelope struct {
	Type          string          `json:"type"`
	MessageID     string          `json:"message_id"`
	SessionID     string          `json:"session_id"`
	Tick          uint64          `json:"tick"`
	CorrelationID string          `json:"correlation_id"`
	Payload       json.RawMessage `json:"payload"`
}

// outEnvelope is an outbound message wrapping an arbitrary payload.
type outEnvelope struct {
	Type          string `json:"type"`
	MessageID     string `json:"message_id"`
	SessionID     string `json:"session_id"`
	Tick          uint64 `json:"tick"`
	CorrelationID string `json:"correlation_id,omitempty"`
	Payload       any    `json:"payload"`
}

// Server-to-client payloads.
type (
	intentAcceptedPayload struct {
		AcceptedMessageID string `json:"accepted_message_id"`
		ServerTick        uint64 `json:"server_tick"`
	}
	intentRejectedPayload struct {
		RejectedMessageID string `json:"rejected_message_id"`
		Reason            string `json:"reason"`
	}
	errorPayload struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	stateDeltaPayload struct {
		ServerTick uint64        `json:"server_tick"`
		Changes    []game.Change `json:"changes"`
		Events     []game.Event  `json:"events"`
	}
)

// Inbound intent payloads (decoded from inEnvelope.Payload).
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

// decodeInput converts a validated inbound envelope into a sim Input. It
// returns ok=false if the payload does not match the message type.
func decodeInput(env inEnvelope) (game.Input, bool) {
	in := game.Input{
		MessageID:     env.MessageID,
		CorrelationID: env.CorrelationID,
		Type:          env.Type,
	}
	switch env.Type {
	case typeMoveIntent:
		var p movePayloadWire
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return in, false
		}
		in.Move = &game.MoveIntent{Direction: p.Direction, DurationTicks: p.DurationTicks}
	case typeAttack:
		var p attackPayloadWire
		if err := json.Unmarshal(env.Payload, &p); err != nil || p.TargetID == "" {
			return in, false
		}
		in.Attack = &game.AttackIntent{TargetID: p.TargetID}
	case typePickUp:
		var p pickUpPayloadWire
		if err := json.Unmarshal(env.Payload, &p); err != nil || p.EntityID == "" {
			return in, false
		}
		in.PickUp = &game.PickUpIntent{EntityID: p.EntityID}
	case typeEquip:
		var p equipPayloadWire
		if err := json.Unmarshal(env.Payload, &p); err != nil || p.ItemInstanceID == "" || p.Slot == "" {
			return in, false
		}
		in.Equip = &game.EquipIntent{ItemInstanceID: p.ItemInstanceID, Slot: p.Slot}
	default:
		return in, false
	}
	return in, true
}

// DecodeStored recovers a sim Input from a persisted input envelope (the full
// raw message stored in session_inputs.payload). It is used by the replay
// engine. The authoritative tick/sequence come from the stored row, not the
// envelope, so only the message type and intent payload are taken here.
func DecodeStored(raw []byte) (game.Input, bool) {
	var env inEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return game.Input{}, false
	}
	if !isClientIntent(env.Type) {
		return game.Input{}, false
	}
	return decodeInput(env)
}

// isClientIntent reports whether the type is a buffered authoritative intent
// (everything except client_ready, which the runner handles inline).
func isClientIntent(t string) bool {
	switch t {
	case typeMoveIntent, typeAttack, typePickUp, typeEquip:
		return true
	}
	return false
}
