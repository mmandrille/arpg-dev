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
	ListCharacters(ctx context.Context, accountID string) ([]CharacterSummary, error)
	CreateCharacter(ctx context.Context, charID, accountID, name, characterClass string) (Character, error)
	RenameCharacter(ctx context.Context, accountID, characterID, name string) (Character, error)
	MarkCharacterDead(ctx context.Context, accountID, characterID string, deathLevel int) error
	ReviveDeadCharacters(ctx context.Context, accountID string) (int, error)
	DeleteCharacter(ctx context.Context, accountID, characterID string) error
}

// SessionRepo manages session lifecycle records.
type SessionRepo interface {
	CreateSession(ctx context.Context, s Session) error
	GetSession(ctx context.Context, id string) (Session, error)
	ListActiveListedSessions(ctx context.Context) ([]SessionSummary, error)
	TouchSession(ctx context.Context, id string) error
	SetSessionStatus(ctx context.Context, id, status string) error
	EndListedSessionIfNoConnected(ctx context.Context, id string) (bool, error)
	CreateSessionHostMember(ctx context.Context, m SessionMember) error
	CreateSessionGuestMember(ctx context.Context, m SessionMember) error
	ListSessionMembers(ctx context.Context, sessionID string) ([]SessionMember, error)
	GetSessionMemberByAccount(ctx context.Context, sessionID, accountID string) (SessionMember, error)
	GetSessionMember(ctx context.Context, sessionID, accountID, characterID string) (SessionMember, error)
	ClaimSessionMemberConnection(ctx context.Context, sessionID, accountID, characterID string) (bool, error)
	SetSessionMemberConnected(ctx context.Context, sessionID, accountID, characterID, playerEntityID string, currentLevel int, tick int64) error
	SetSessionMemberDisconnected(ctx context.Context, sessionID, accountID, characterID string, currentLevel int, tick int64) error
	SetSessionMemberPlayer(ctx context.Context, sessionID, accountID, characterID, playerEntityID string, currentLevel int) error
}

// CharacterProgressionRepo manages durable character items, waypoints, and the
// immutable session-start progression boundary used by replay.
type CharacterProgressionRepo interface {
	ListCharacterItems(ctx context.Context, accountID, characterID string) ([]CharacterItemInstance, error)
	AddCharacterItem(ctx context.Context, item CharacterItemInstance) error
	SetCharacterItemLocation(ctx context.Context, accountID, characterID, itemInstanceID, location string) error
	SetCharacterItemEquipped(ctx context.Context, accountID, characterID, itemInstanceID, slot string, equipped bool, weaponSet int) error
	RemoveCharacterItem(ctx context.Context, accountID, characterID, itemInstanceID string) error
	ListAccountWaypoints(ctx context.Context, accountID, characterID string) ([]CharacterWaypoint, error)
	AddAccountWaypoint(ctx context.Context, accountID string, level int) (bool, error)
	GetOrCreateCharacterProgression(ctx context.Context, accountID, characterID string, defaults CharacterProgressionDefaults) (CharacterProgression, error)
	GetCharacterProgression(ctx context.Context, accountID, characterID string) (CharacterProgression, error)
	UpsertCharacterProgression(ctx context.Context, accountID string, progression CharacterProgression) error
	SetCharacterGold(ctx context.Context, accountID, characterID string, gold int) error
	ListCharacterHotbar(ctx context.Context, accountID, characterID string) ([]CharacterHotbarSlot, error)
	SetCharacterHotbarSlot(ctx context.Context, accountID, characterID string, slotIndex int, itemInstanceID *string) error
	GetOrCreateCharacterSkillBindings(ctx context.Context, accountID, characterID string) (CharacterSkillBindings, error)
	SetCharacterSkillBindings(ctx context.Context, bindings CharacterSkillBindings) error
	ListCharacterShopStock(ctx context.Context, accountID, characterID string) ([]CharacterShopStockItem, error)
	ReplaceCharacterShopStock(ctx context.Context, accountID, characterID, shopID, refreshKey string, stock []CharacterShopStockItem) error
	SetCharacterShopStockAvailable(ctx context.Context, accountID, characterID, shopID, offerID string, available bool) error
	ListAccountStashItems(ctx context.Context, accountID string) ([]AccountStashItem, error)
	GetOrCreateAccountStashGold(ctx context.Context, accountID string) (AccountStashGold, error)
	TransferCharacterItemToAccountStash(ctx context.Context, accountID, characterID, itemInstanceID, stashItemID string) (AccountStashItem, error)
	TransferAccountStashItemToCharacter(ctx context.Context, accountID, characterID, stashItemID, itemInstanceID string) (CharacterItemInstance, error)
	TransferAccountStashItemToCharacterWithPlacement(ctx context.Context, accountID, characterID, stashItemID, itemInstanceID, location, slot string, equipped bool) (CharacterItemInstance, error)
	TransferCharacterGoldToAccountStash(ctx context.Context, accountID, characterID string, amount int) (characterGold int, stashGold int, err error)
	TransferAccountStashGoldToCharacter(ctx context.Context, accountID, characterID string, amount int) (characterGold int, stashGold int, err error)
	ListRecoverableCharacterCorpses(ctx context.Context, accountID, excludeCharacterID string) ([]CharacterCorpse, error)
	TransferCorpseItemToCharacter(ctx context.Context, accountID, corpseCharacterID, targetCharacterID, corpseItemID, newItemID string) (CharacterItemInstance, error)
	UpgradeAccountStashItem(ctx context.Context, accountID, stashItemID string, baseCostGold, costGrowthPerLevel, maxLevel, successChancePercent, successRoll, pityFailureThreshold int, eligibleItemDefs map[string]struct{}) (AccountStashItem, int, int, bool, error)
	UpgradeAccountStashItemWithWallet(ctx context.Context, accountID, characterID, stashItemID string, baseCostGold, costGrowthPerLevel, maxLevel, successChancePercent, successRoll, pityFailureThreshold int, eligibleItemDefs map[string]struct{}) (AccountStashItem, int, int, int, bool, error)
	ListActiveMarketListings(ctx context.Context) ([]MarketListing, error)
	CreateMarketListingFromStash(ctx context.Context, accountID, stashItemID, listingID string, priceGold int) (MarketListing, error)
	CancelMarketListing(ctx context.Context, accountID, listingID string) (MarketListing, error)
	PurchaseMarketListing(ctx context.Context, buyerAccountID, listingID string) (MarketListing, error)
	CreateMarketOffer(ctx context.Context, bidderAccountID, listingID, offerID string, stashItemIDs []string) (MarketOffer, error)
	CancelMarketOffer(ctx context.Context, bidderAccountID, listingID, offerID string) (MarketOffer, error)
	ListMarketOffersForSeller(ctx context.Context, sellerAccountID, listingID string) ([]MarketOffer, error)
	AcceptMarketOffer(ctx context.Context, sellerAccountID, listingID, offerID string) (MarketOffer, error)
	ExpireMarketListings(ctx context.Context) (int, error)
	GetMarketSummary(ctx context.Context, accountID string) (MarketSummary, error)
	CreateSessionStartSnapshot(ctx context.Context, sessionID, accountID, characterID string, items []CharacterItemInstance, waypoints []CharacterWaypoint, hotbar []CharacterHotbarSlot, skillBinds CharacterSkillBindings, shopStock []CharacterShopStockItem, stashItems []AccountStashItem, stashGold AccountStashGold, progression CharacterProgression) error
	LoadSessionStartSnapshot(ctx context.Context, sessionID string) (SessionStartSnapshot, error)
	LoadSessionStartSnapshotForMember(ctx context.Context, sessionID, accountID, characterID string) (SessionStartSnapshot, error)
	LoadSessionStartSnapshots(ctx context.Context, sessionID string) ([]SessionStartSnapshot, error)
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
	CharacterProgressionRepo
	InputRepo
	EventRepo
	Ping(ctx context.Context) error
}

// compile-time assertion that *Store implements Repository.
var _ Repository = (*Store)(nil)
