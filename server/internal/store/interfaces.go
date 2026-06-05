package store

import "context"

// The repository interfaces below isolate persistence behind narrow contracts
// (ADR-0001 persistence default). *Store implements all of them; consumers
// depend only on the slice they need.

// AccountRepo manages accounts and their default character.
type AccountRepo interface {
	UpsertAccountByEmail(ctx context.Context, id, email string) (Account, error)
	GetAccount(ctx context.Context, id string) (Account, error)
}

// CharacterRepo manages characters.
type CharacterRepo interface {
	GetOrCreateDefaultCharacter(ctx context.Context, charID, accountID, name string) (Character, error)
	GetCharacter(ctx context.Context, id string) (Character, error)
}

// SessionRepo manages session lifecycle records.
type SessionRepo interface {
	CreateSession(ctx context.Context, s Session) error
	GetSession(ctx context.Context, id string) (Session, error)
	TouchSession(ctx context.Context, id string) error
	SetSessionStatus(ctx context.Context, id, status string) error
}

// InventoryRepo manages session-scoped inventory.
type InventoryRepo interface {
	ListInventory(ctx context.Context, sessionID string) ([]InventoryItem, error)
	AddInventoryItem(ctx context.Context, item InventoryItem) error
	SetEquipped(ctx context.Context, sessionID, itemInstanceID, slot string, equipped bool) error
}

// InputRepo records and reads authoritative inputs for replay.
type InputRepo interface {
	AppendInput(ctx context.Context, in SessionInput) error
	ListInputs(ctx context.Context, sessionID string) ([]SessionInput, error)
}

// EventRepo records and reads authoritative events.
type EventRepo interface {
	AppendEvent(ctx context.Context, ev SessionEvent) error
	ListEvents(ctx context.Context, sessionID string) ([]SessionEvent, error)
}

// Repository aggregates every repo; *Store satisfies it.
type Repository interface {
	AccountRepo
	CharacterRepo
	SessionRepo
	InventoryRepo
	InputRepo
	EventRepo
	Ping(ctx context.Context) error
}

// compile-time assertion that *Store implements Repository.
var _ Repository = (*Store)(nil)
