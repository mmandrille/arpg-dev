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
	ID             string
	AccountID      string
	Name           string
	CharacterClass string
	Dead           bool
	DeathLevel     *int
	CreatedAt      time.Time
}

// CharacterSummary is the account-scoped character-list row exposed before a
// session starts. Progression fields are display-only summaries from durable
// character_progression, with safe defaults when no row exists yet.
type CharacterSummary struct {
	ID                  string
	AccountID           string
	Name                string
	CharacterClass      string
	Dead                bool
	DeathLevel          *int
	Level               int
	Gold                int
	DeepestDungeonDepth int
	CreatedAt           time.Time
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
	WeaponSet   int
	RolledStats json.RawMessage
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CharacterCorpse is a dead character with recoverable inventory/equipment.
type CharacterCorpse struct {
	CharacterID string
	Name        string
	Level       int
	DeathLevel  int
	Items       []CharacterItemInstance
}

const (
	ItemLocationInventory = "inventory"
	ItemLocationEquipped  = "equipped"
	ItemLocationStash     = "stash"
)

// AccountStashItem is an account-owned item row that can be shared across
// characters on the same account.
type AccountStashItem struct {
	AccountID         string
	StashItemID       string
	SourceCharacterID string
	ItemDefID         string
	RolledStats       json.RawMessage
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// AccountStashGold is the account-owned stash wallet.
type AccountStashGold struct {
	AccountID string
	Gold      int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AccountResourceAmount is an account-owned material/resource balance.
type AccountResourceAmount struct {
	AccountID  string
	ResourceID string
	Amount     int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

const (
	MarketListingActive   = "active"
	MarketListingCanceled = "canceled"
	MarketListingAccepted = "accepted"
	MarketListingExpired  = "expired"
)

// MarketListing is a durable seller listing backed by one former stash item.
type MarketListing struct {
	ID                string
	SellerAccountID   string
	StashItemID       string
	SourceCharacterID string
	ItemDefID         string
	RolledStats       json.RawMessage
	PriceGold         int
	Status            string
	ExpiresAt         time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	CanceledAt        *time.Time
	AcceptedAt        *time.Time
	ExpiredAt         *time.Time
}

const (
	MarketOfferActive   = "active"
	MarketOfferAccepted = "accepted"
	MarketOfferRejected = "rejected"
	MarketOfferCanceled = "canceled"
)

// MarketOffer is a durable item-for-item bid on one market listing.
type MarketOffer struct {
	ID              string
	ListingID       string
	BidderAccountID string
	Status          string
	Items           []MarketOfferItem
	Listing         *MarketListing
	CreatedAt       time.Time
	UpdatedAt       time.Time
	AcceptedAt      *time.Time
	RejectedAt      *time.Time
	CanceledAt      *time.Time
}

// MarketOfferItem is one item snapshot reserved inside a market offer.
type MarketOfferItem struct {
	OfferID           string
	BidderAccountID   string
	StashItemID       string
	SourceCharacterID string
	ItemDefID         string
	RolledStats       json.RawMessage
	CreatedAt         time.Time
}

// MarketSummary is the count-only notification surface for the town board.
type MarketSummary struct {
	PublishedListings int
	IncomingBids      int
}

// MarketAuditRecord is an immutable ownership-transition trace for market actions.
type MarketAuditRecord struct {
	ID              int64
	Action          string
	ListingID       string
	OfferID         string
	ActorAccountID  string
	SellerAccountID string
	BidderAccountID string
	ItemDefID       string
	StashItemID     string
	Details         json.RawMessage
	CreatedAt       time.Time
}

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
	UnspentSkillPoints  int
	Stats               CharacterBaseStats
	Gold                int
	DeepestDungeonDepth int
	SkillRanks          map[string]int
}

// CharacterProgression is durable character-owned XP, level, base stats, and
// skill progression.
type CharacterProgression struct {
	AccountID           string
	CharacterID         string
	CharacterClass      string
	Level               int
	Experience          int
	UnspentStatPoints   int
	UnspentSkillPoints  int
	Stats               CharacterBaseStats
	Gold                int
	DeepestDungeonDepth int
	SkillRanks          map[string]int
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

// CharacterSkillBindings is the durable character-owned skill control layout.
type CharacterSkillBindings struct {
	AccountID         string
	CharacterID       string
	FunctionKeys      []string
	RightClickSkillID string
}

// CharacterShopStockItem is one durable generated shop-stock row. Buyback rows
// are intentionally excluded from persistence and session-start snapshots.
type CharacterShopStockItem struct {
	AccountID      string
	CharacterID    string
	ShopID         string
	RefreshKey     string
	OfferIndex     int
	OfferID        string
	SourceDepth    int
	ItemTemplateID string
	RolledPayload  json.RawMessage
	BuyPrice       int
	Available      bool
	CreatedAt      time.Time
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
	SkillBinds  CharacterSkillBindings
	ShopStock   []CharacterShopStockItem
	StashItems  []AccountStashItem
	StashGold   AccountStashGold
	Resources   []AccountResourceAmount
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
