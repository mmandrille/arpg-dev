// Package realtime implements the authenticated WebSocket game protocol: the
// per-connection session runner that drives the 20 Hz authoritative tick loop,
// validates envelopes, applies buffered inputs in deterministic order, and
// persists inputs/events/inventory (ADR-0001 D2/D3/D8.1/D8.2).
package realtime

import (
	"encoding/json"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/inputdecode"
)

// tickHz is the authoritative simulation rate (20 Hz, 50 ms/tick).
const tickHz = 20

// Client-to-server message types.
const (
	typeClientReady = inputdecode.TypeClientReady
	typeMoveIntent  = inputdecode.TypeMoveIntent
	typeAction      = inputdecode.TypeAction
	typeEquip       = inputdecode.TypeEquip
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

// decodeInput converts a validated inbound envelope into a sim Input. It
// returns ok=false if the payload does not match the message type.
func decodeInput(env inEnvelope) (game.Input, bool) {
	return inputdecode.Decode(env.Type, env.MessageID, env.CorrelationID, env.Payload)
}

// isClientIntent reports whether the type is a buffered authoritative intent
// (everything except client_ready, which the runner handles inline).
func isClientIntent(t string) bool {
	return inputdecode.IsClientIntent(t)
}
