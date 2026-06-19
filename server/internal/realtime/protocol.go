// Package realtime implements the authenticated WebSocket game protocol: the
// per-connection session runner that drives the 10 Hz authoritative tick loop,
// validates envelopes, applies buffered inputs in deterministic order, and
// persists inputs/events/inventory (ADR-0001 D2/D3/D8.1/D8.2).
package realtime

import (
	"encoding/json"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/inputdecode"
)

// tickHz is the live authoritative session rate. The deterministic sim still
// advances one fixed tick per loop; lowering the loop rate slows live gameplay
// without changing replay tick semantics.
const tickHz = 10

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
		ServerTick  uint64                    `json:"server_tick"`
		Level       int                       `json:"level"`
		Changes     []game.Change             `json:"changes"`
		Events      []game.Event              `json:"events"`
		Performance *performanceStatusPayload `json:"performance,omitempty"`
	}
	performanceStatusPayload struct {
		Tick               uint64  `json:"tick"`
		TotalMS            float64 `json:"total_ms"`
		SimMS              float64 `json:"sim_ms"`
		AIMS               float64 `json:"ai_ms"`
		PathfindMS         float64 `json:"pathfind_ms"`
		CombatMS           float64 `json:"combat_ms"`
		BroadcastMS        float64 `json:"broadcast_ms"`
		PersistMS          float64 `json:"persist_ms"`
		PathRequests       int     `json:"path_requests"`
		PathCacheHits      int     `json:"path_cache_hits"`
		PathNodesVisited   int     `json:"path_nodes_visited"`
		MonstersMoved      int     `json:"monsters_moved"`
		TickBudgetMS       float64 `json:"tick_budget_ms"`
		TickOverBudget     bool    `json:"tick_over_budget"`
		TickOverrunMS      float64 `json:"tick_overrun_ms"`
		DegradationApplied bool    `json:"degradation_applied"`
		Inputs             int     `json:"inputs"`
		Results            int     `json:"results"`
		Changes            int     `json:"changes"`
		Events             int     `json:"events"`
		Acks               int     `json:"acks"`
		Rejects            int     `json:"rejects"`
		Clients            int     `json:"clients"`
		GameLevel          int     `json:"game_level"`
		Entities           int     `json:"entities"`
		Players            int     `json:"players"`
		Monsters           int     `json:"monsters"`
		LiveMonsters       int     `json:"live_monsters"`
		Companions         int     `json:"companions"`
		Projectiles        int     `json:"projectiles"`
		Loot               int     `json:"loot"`
		Interactables      int     `json:"interactables"`
		Walls              int     `json:"walls"`
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
