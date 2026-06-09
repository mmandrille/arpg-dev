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
	Dead      bool
	CreatedAt time.Time
}

// Session is one authoritative game session. Solo sessions have one host
// member; co-op sessions have one host plus zero or more guests.
type Session struct {
	ID           string
	AccountID    string
	CharacterID  string
	Seed         string // hex-encoded server seed
	WorldID      string // shared/rules/worlds.v0.json preset id
	Mode         string // "solo" | "coop"
	Listed       bool
	JoinCodeHash string
	Status       string // "active" | "ended"
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// SessionSummary is the public active-session browser row. It intentionally
// omits raw join codes and account ids.
type SessionSummary struct {
	SessionID       string
	WorldID         string
	Mode            string
	Listed          bool
	HostCharacterID string
	HostDisplayName string
	MemberCount     int
	ConnectedCount  int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Session status values.
const (
	SessionActive = "active"
	SessionEnded  = "ended"
)

// Session mode values.
const (
	SessionModeSolo = "solo"
	SessionModeCoop = "coop"
)

// Session member roles and statuses.
const (
	SessionMemberHost  = "host"
	SessionMemberGuest = "guest"

	SessionMemberActive = "active"
	SessionMemberLeft   = "left"
)

// defaultWorldID is used when legacy rows omit world_id.
const defaultWorldID = "vertical_slice"

// CharacterItemInstance is a durable character-owned item instance. ID is the
// protocol item_instance_id loaded into fresh Sim snapshots.
type CharacterItemInstance struct {
	ID          string
	AccountID   string
	CharacterID string
	ItemDefID   string
	Location    string
	Slot        string
	Equipped    bool
	RolledStats json.RawMessage
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

const (
	ItemLocationInventory = "inventory"
	ItemLocationEquipped  = "equipped"
	ItemLocationStash     = "stash"
)

// CharacterWaypoint is a durable unlocked waypoint level for a character.
type CharacterWaypoint struct {
	CharacterID  string
	Level        int
	DiscoveredAt time.Time
}

// CharacterBaseStats are the durable base stat allocations for a character.
type CharacterBaseStats struct {
	Str   int
	Dex   int
	Vit   int
	Magic int
}

// CharacterProgressionDefaults is the seed row supplied by game rules when a
// character has no durable progression yet.
type CharacterProgressionDefaults struct {
	Level               int
	Experience          int
	UnspentStatPoints   int
	Stats               CharacterBaseStats
	Gold                int
	DeepestDungeonDepth int
}

// CharacterProgression is durable character-owned XP, level, and base stats.
type CharacterProgression struct {
	AccountID           string
	CharacterID         string
	Level               int
	Experience          int
	UnspentStatPoints   int
	Stats               CharacterBaseStats
	Gold                int
	DeepestDungeonDepth int
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// CharacterHotbarSlot is one durable character-owned hotbar assignment.
type CharacterHotbarSlot struct {
	AccountID      string
	CharacterID    string
	SlotIndex      int
	ItemInstanceID *string
	UpdatedAt      time.Time
}

// SessionStartSnapshot freezes the character progression visible when a
// session was created. Replay uses this instead of mutable live character rows.
type SessionStartSnapshot struct {
	SessionID   string
	AccountID   string
	CharacterID string
	Items       []CharacterItemInstance
	Waypoints   []CharacterWaypoint
	Hotbar      []CharacterHotbarSlot
	Progression *CharacterProgression
}

// SessionMember binds an authenticated account/character to one player entity
// inside a session.
type SessionMember struct {
	SessionID      string
	AccountID      string
	CharacterID    string
	PlayerEntityID string
	Role           string
	Status         string
	Connected      bool
	CurrentLevel   int
	JoinedTick     int64
	LeftTick       *int64
	JoinedAt       time.Time
	UpdatedAt      time.Time
}

// SessionInput is a recorded authoritative input (spec 4.6, ADR D8.2).
type SessionInput struct {
	ID                  string
	SessionID           string
	Tick                int64
	Sequence            int64
	MessageID           string
	CorrelationID       string
	ActorAccountID      string
	ActorCharacterID    string
	ActorPlayerEntityID string
	Payload             json.RawMessage
	CreatedAt           time.Time
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
