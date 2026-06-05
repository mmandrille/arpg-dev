package store

import (
	"encoding/json"
	"time"
)

// Account is a platform identity (spec 4.6).
type Account struct {
	ID        string
	Email     string
	CreatedAt time.Time
}

// Character belongs to an account.
type Character struct {
	ID        string
	AccountID string
	Name      string
	CreatedAt time.Time
}

// Session is one solo authoritative game session.
type Session struct {
	ID          string
	AccountID   string
	CharacterID string
	Seed        string // hex-encoded server seed
	Status      string // "active" | "ended"
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Session status values.
const (
	SessionActive = "active"
	SessionEnded  = "ended"
)

// InventoryItem is a session-scoped inventory entry (see migration note). ID is
// the protocol item_instance_id (decimal string), unique within its session.
type InventoryItem struct {
	ID          string
	SessionID   string
	AccountID   string
	CharacterID string
	ItemDefID   string
	Slot        string // "" when not slotted
	Equipped    bool
	CreatedAt   time.Time
}

// SessionInput is a recorded authoritative input (spec 4.6, ADR D8.2).
type SessionInput struct {
	ID            string
	SessionID     string
	Tick          int64
	Sequence      int64
	MessageID     string
	CorrelationID string
	Payload       json.RawMessage
	CreatedAt     time.Time
}

// SessionEvent is a recorded authoritative output event (spec 4.6, ADR D8.2).
type SessionEvent struct {
	ID            string
	SessionID     string
	Tick          int64
	Sequence      int64
	EventType     string
	CorrelationID string
	Payload       json.RawMessage
	CreatedAt     time.Time
}
