package game

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
)

// Simulation constants for the v0 slice.
const (
	baseEntityID                   = 1001 // player=1001, monster=1002, loot=1003, item=1004 ...
	playerStartHP                  = 10
	moveSpeed                      = 1.0
	playerRadius                   = 0.45
	monsterRadius                  = 0.45
	monsterDefID                   = "training_dummy"
	playerEntity                   = "player"
	monsterEntity                  = "monster"
	lootEntity                     = "loot"
	projectileEntity               = "projectile"
	wallEntity                     = "wall"
	interactableEntity             = "interactable"
	monsterBehaviorStatic          = "static"
	monsterBehaviorChase           = "chase"
	monsterAIModeIdle              = "idle"
	monsterAIModeChase             = "chase"
	monsterAIModeReturn            = "return"
	interactableClosed             = "closed"
	interactableOpen               = "open"
	interactableReady              = "ready"
	interactableLocked             = "locked"
	interactableDisabled           = "disabled"
	interactableTransitionAscend   = "ascend"
	interactableTransitionDescend  = "descend"
	interactableTransitionWaypoint = "waypoint"
	stairsDownDefID                = "stairs_down"
	stairsUpDefID                  = "stairs_up"
	teleporterDefID                = "teleporter"
	treasureChestDefID             = "treasure_chest"
	townStashDefID                 = "town_stash"
	accountStashID                 = "account_stash"
	worldModeMultiLevel            = "multi_level"
	attackModeMelee                = "melee"
	attackModeRanged               = "ranged"
	magicBoltSkillID               = "magic_bolt"
	trainingArrowProjectileDefID   = "training_arrow"
	goldItemDefID                  = "gold"
	mainHandSlot                   = "main_hand"
	offHandSlot                    = "off_hand"
	ringLeftSlot                   = "ring_left"
	ringRightSlot                  = "ring_right"
	lootInteractionRadius          = 0.35
	interactableInteractionRadius  = 0.50
	meleeRangeEpsilon              = 0.000001
	directionalMeleeHalfWidth      = 0.35
	projectileRadius               = 0.10
	tickDuration                   = 0.05
	minHotbarCapacity              = 2
	maxHotbarCapacity              = 10
	skillFunctionKeyCount          = 8
	baseInventoryRows              = 3
	inventoryColumns               = 5
	maxInventoryRows               = 20
	defaultStashCapacity           = 50
	minimumChaseLeashTiles         = 25.0
)

var equipmentSlots = []string{
	"head",
	"amulet",
	"chest",
	"gloves",
	"belt",
	"boots",
	ringLeftSlot,
	ringRightSlot,
	mainHandSlot,
	offHandSlot,
}

// DefaultWorldID is the compatibility world used when callers do not choose a
// preset explicitly.
const DefaultWorldID = "vertical_slice"

const (
	townLevel                 = 0
	levelZero                 = 0
	townNavigationMarginCells = 8.0
)

// entity is the internal mutable scene entity.
type entity struct {
	id                   uint64
	kind                 string
	pos                  Vec2
	hp                   int
	maxHP                int
	mana                 int
	maxMana              int
	characterID          string
	displayName          string
	monsterDefID         string
	monsterRarityID      string
	monsterAttackDamage  *DamageRange
	monsterXPReward      int
	isBoss               bool
	bossTemplateID       string
	visualModel          string
	visualTint           string
	visualScale          float64
	bossPatternID        string
	bossPatternDeckIndex int
	bossPhaseIndex       int
	bossPhaseKind        string
	bossPhaseStarted     uint64
	bossPhaseEnds        uint64
	bossCooldownEnds     uint64
	bossActiveHit        map[uint64]bool
	itemDefID            string
	goldAmount           int
	rollPayload          *ItemRollPayload
	interactableDefID    string
	state                string
	lootTable            string
	ownerID              uint64
	targetID             uint64
	projectileDefID      string
	dir                  Vec2
	speed                float64
	traveled             float64
	maxDistance          float64
	damageRange          DamageRange
	sourceMsgID          string
	sourceCorrID         string
	spawnTick            uint64
	spawnPos             Vec2
	aiMode               string
	aiTargetPlayerID     uint64
	lastAttackTick       uint64
	hasAttacked          bool
}

// invItem is an internal inventory item.
type invItem struct {
	instanceID  uint64
	itemDefID   string
	rollPayload *ItemRollPayload
	slot        string
	equipped    bool
}

type stashItem struct {
	stashItemID uint64
	itemDefID   string
	rollPayload *ItemRollPayload
}

type goldRollContext struct {
	levelNum        int
	monsterRarityID string
}

type activeMove struct {
	dir       Vec2
	remaining int
}

type autoNavState struct {
	steps         []Vec2
	pendingAction *ActionIntent
	sourceMsgID   string
	sourceCorrID  string
}

type wallObstacle struct {
	pos         Vec2
	size        Vec2
	source      string
	shapeFamily string
}

type effectiveCombatStats struct {
	DamageMin            float64
	DamageMax            float64
	HitChance            float64
	CritChance           float64
	CritDamage           float64
	Armor                float64
	BlockPercent         float64
	AttackSpeed          float64
	AttackIntervalTicks  int
	MaxHP                float64
	MaxMana              float64
	HealthRegenPerSecond float64
	ManaRegenPerSecond   float64
}

type combatResolution struct {
	Outcome         string
	Damage          int
	RawDamage       int
	MitigatedDamage int
	Blocked         bool
	Critical        bool
	Hit             bool
}

type bossPhaseRuntime struct {
	patternID string
	index     int
	phase     BossPatternPhase
}

type skillCooldownState struct {
	EndsTick   uint64
	TotalTicks int
}

type skillEffectState struct {
	SkillID     string
	Stats       []string
	Percent     int
	VisualScale float64
	EndsTick    uint64
	TotalTicks  int
}

type skillHealApplication struct {
	Target *entity
	Heal   int
}

type playerState struct {
	PlayerID              uint64
	AccountID             string
	CharacterID           string
	DisplayName           string
	Role                  string
	Connected             bool
	CurrentLevel          int
	Move                  *activeMove
	AutoNav               *autoNavState
	Inventory             []*invItem
	Equipped              map[string]uint64
	Hotbar                []uint64
	DiscoveredTeleporters map[int]bool
	Progression           CharacterProgressionState
	SkillCooldowns        map[string]skillCooldownState
	SkillEffects          map[string]skillEffectState
	SkillFunctionKeys     []string
	RightClickSkillID     string
	ShopStock             map[string]*shopStockState
	Gold                  int
	StashItems            []*stashItem
	StashGold             int
	StashCapacity         int
	HPRegenCarry          float64
	ManaRegenCarry        float64
}

// Sim is the deterministic authoritative simulation for one solo session.
// Given the same seed and the same ordered inputs, it produces identical
// outputs (entity ids, events, final state) on every run (ADR-0001 D8.1).
type Sim struct {
	sessionID string
	seed      string
	rng       *RNG
	rules     *Rules

	tick     uint64
	nextID   uint64
	playerID uint64
	players  map[uint64]*playerState
	goldRoll uint64

	levels                map[int]*LevelState
	currentLevel          int
	multiLevel            bool
	entities              map[uint64]*entity
	walls                 []wallObstacle
	move                  *activeMove
	autoNav               *autoNavState
	inventory             []*invItem
	equipped              map[string]uint64 // slot -> instanceID (0 = none)
	hotbar                []uint64          // fixed 10-slot item instance assignments (0 = none)
	discoveredTeleporters map[int]bool
	progression           CharacterProgressionState
	skillCooldowns        map[string]skillCooldownState
	skillEffects          map[string]skillEffectState
	skillFunctionKeys     []string
	rightClickSkillID     string
	shopStock             map[string]*shopStockState
	gold                  int
	stashItems            []*stashItem
	stashGold             int
	stashCapacity         int
	hpRegenCarry          float64
	manaRegenCarry        float64
}

// CharacterProgressionState is the authoritative mutable progression state for
// one character inside a sim session.
type CharacterProgressionState struct {
	Level               int
	Experience          int
	UnspentStatPoints   int
	UnspentSkillPoints  int
	SkillRanks          map[string]int
	BaseStats           BaseStatsView
	Gold                int
	DeepestDungeonDepth int
}

// NewSim builds a fresh session in the default vertical-slice world.
func NewSim(sessionID, seed string, rules *Rules) *Sim {
	s, err := NewSimWithWorld(sessionID, seed, rules, DefaultWorldID)
	if err != nil {
		panic(err)
	}
	return s
}

// NewSimWithWorld builds a fresh session from a deterministic world preset.
func NewSimWithWorld(sessionID, seed string, rules *Rules, worldID string) (*Sim, error) {
	return NewSimWithWorldProgression(sessionID, seed, rules, worldID, rules.DefaultCharacterProgressionState())
}

// NewSimWithWorldProgression builds a session with a caller-supplied durable
// character progression snapshot.
func NewSimWithWorldProgression(sessionID, seed string, rules *Rules, worldID string, progression CharacterProgressionState) (*Sim, error) {
	world, ok := rules.Worlds[worldID]
	if !ok {
		return nil, ErrUnknownWorld{WorldID: worldID}
	}
	progression = rules.normalizeProgressionState(progression)
	s := &Sim{
		sessionID:             sessionID,
		seed:                  seed,
		rng:                   NewRNG(SeedToUint64(seed)),
		rules:                 rules,
		nextID:                baseEntityID,
		players:               make(map[uint64]*playerState),
		levels:                make(map[int]*LevelState),
		currentLevel:          levelZero,
		multiLevel:            world.Mode == worldModeMultiLevel,
		equipped:              newEquippedMap(),
		hotbar:                make([]uint64, 10),
		discoveredTeleporters: make(map[int]bool),
		progression:           progression,
		skillCooldowns:        make(map[string]skillCooldownState),
		skillEffects:          make(map[string]skillEffectState),
		skillFunctionKeys:     make([]string, skillFunctionKeyCount),
		shopStock:             make(map[string]*shopStockState),
		gold:                  progression.Gold,
		stashCapacity:         defaultStashCapacity,
	}

	if s.multiLevel {
		s.currentLevel = townLevel
		townNav := townNavigationForWorld(rules.Navigation, world)
		level := newLevelState(townLevel, &townNav)
		s.levels[townLevel] = level
		if err := s.populatePresetLevel(level, worldID, world); err != nil {
			return nil, err
		}
		s.discoveredTeleporters[townLevel] = true
		if world.StartLevel != nil && *world.StartLevel < townLevel {
			start, err := s.ensureTravelLevel(*world.StartLevel)
			if err != nil {
				return nil, err
			}
			player := level.entities[s.playerID]
			if player == nil {
				return nil, fmt.Errorf("game: missing player for world %s", worldID)
			}
			delete(level.entities, s.playerID)
			player.pos = world.Player.Position
			start.entities[s.playerID] = player
			s.currentLevel = start.levelNum
			if ps := s.players[s.playerID]; ps != nil {
				ps.CurrentLevel = start.levelNum
			}
		}
		s.syncCompatibilityFields()
		return s, nil
	}

	level := newLevelState(levelZero, &rules.Navigation)
	s.levels[levelZero] = level
	if err := s.populatePresetLevel(level, worldID, world); err != nil {
		return nil, err
	}

	s.syncCompatibilityFields()
	return s, nil
}

func townNavigationForWorld(global NavigationRules, world WorldDef) NavigationRules {
	nav := global
	cellSize := nav.CellSize
	if cellSize <= 0 {
		cellSize = 1.0
	}
	minX := world.Player.Position.X
	maxX := world.Player.Position.X
	minY := world.Player.Position.Y
	maxY := world.Player.Position.Y
	for _, entity := range world.Entities {
		entityMinX := entity.Position.X
		entityMaxX := entity.Position.X
		entityMinY := entity.Position.Y
		entityMaxY := entity.Position.Y
		if entity.Type == wallEntity {
			entityMinX -= entity.Size.X / 2
			entityMaxX += entity.Size.X / 2
			entityMinY -= entity.Size.Y / 2
			entityMaxY += entity.Size.Y / 2
		}
		minX = math.Min(minX, entityMinX)
		maxX = math.Max(maxX, entityMaxX)
		minY = math.Min(minY, entityMinY)
		maxY = math.Max(maxY, entityMaxY)
	}
	margin := townNavigationMarginCells * cellSize
	bounds := GridBounds{
		MinX: int(math.Floor((minX - margin) / cellSize)),
		MinY: int(math.Floor((minY - margin) / cellSize)),
		MaxX: int(math.Ceil((maxX + margin) / cellSize)),
		MaxY: int(math.Ceil((maxY + margin) / cellSize)),
	}
	nav.GridBounds = GridBounds{
		MinX: minInt(global.GridBounds.MinX, bounds.MinX),
		MinY: minInt(global.GridBounds.MinY, bounds.MinY),
		MaxX: maxInt(global.GridBounds.MaxX, bounds.MaxX),
		MaxY: maxInt(global.GridBounds.MaxY, bounds.MaxY),
	}
	nav.MaxAutoSteps = maxInt(nav.MaxAutoSteps, (nav.GridBounds.MaxX-nav.GridBounds.MinX)+(nav.GridBounds.MaxY-nav.GridBounds.MinY))
	return nav
}

// DefaultCharacterProgressionState returns the level/stat defaults from shared
// character progression rules.
func (r *Rules) DefaultCharacterProgressionState() CharacterProgressionState {
	return CharacterProgressionState{
		Level:               1,
		Experience:          0,
		UnspentStatPoints:   0,
		UnspentSkillPoints:  0,
		SkillRanks:          r.defaultSkillRanks(),
		BaseStats:           r.CharacterProgression.BaseStats,
		Gold:                0,
		DeepestDungeonDepth: 0,
	}
}

func (r *Rules) normalizeProgressionState(in CharacterProgressionState) CharacterProgressionState {
	if in.Level < 1 {
		in.Level = 1
	}
	if in.Experience < 0 {
		in.Experience = 0
	}
	if in.UnspentStatPoints < 0 {
		in.UnspentStatPoints = 0
	}
	if in.UnspentSkillPoints < 0 {
		in.UnspentSkillPoints = 0
	}
	in.SkillRanks = r.normalizeSkillRanks(in.SkillRanks)
	if in.Gold < 0 {
		in.Gold = 0
	}
	if in.DeepestDungeonDepth < 0 {
		in.DeepestDungeonDepth = 0
	}
	if in.BaseStats.Str <= 0 {
		in.BaseStats.Str = r.CharacterProgression.BaseStats.Str
	}
	if in.BaseStats.Dex <= 0 {
		in.BaseStats.Dex = r.CharacterProgression.BaseStats.Dex
	}
	if in.BaseStats.Vit <= 0 {
		in.BaseStats.Vit = r.CharacterProgression.BaseStats.Vit
	}
	if in.BaseStats.Magic <= 0 {
		in.BaseStats.Magic = r.CharacterProgression.BaseStats.Magic
	}
	if in.Level > r.CharacterProgression.LevelCap {
		in.Level = r.CharacterProgression.LevelCap
	}
	return in
}

func (r *Rules) defaultSkillRanks() map[string]int {
	out := make(map[string]int, len(r.Skills))
	for skillID := range r.Skills {
		out[skillID] = 0
	}
	return out
}

func (r *Rules) normalizeSkillRanks(in map[string]int) map[string]int {
	out := r.defaultSkillRanks()
	for skillID, rank := range in { //nolint:determinism — output is a map, iteration order does not affect the result
		def, ok := r.Skills[skillID]
		if !ok {
			continue
		}
		if rank < 0 {
			rank = 0
		}
		if rank > def.MaxRank {
			rank = def.MaxRank
		}
		out[skillID] = rank
	}
	return out
}

func cloneSkillCooldowns(in map[string]skillCooldownState) map[string]skillCooldownState {
	if len(in) == 0 {
		return make(map[string]skillCooldownState)
	}
	out := make(map[string]skillCooldownState, len(in))
	for skillID, cooldown := range in { //nolint:determinism — pure map clone, output is a map
		out[skillID] = cooldown
	}
	return out
}

func cloneSkillEffects(in map[string]skillEffectState) map[string]skillEffectState {
	if len(in) == 0 {
		return make(map[string]skillEffectState)
	}
	out := make(map[string]skillEffectState, len(in))
	for skillID, effect := range in { //nolint:determinism — pure map clone, output is a map
		effect.Stats = cloneStringSlice(effect.Stats)
		out[skillID] = effect
	}
	return out
}

func (s *Sim) populatePresetLevel(level *LevelState, worldID string, world WorldDef) error {
	maxHP := s.currentMaxHP()
	maxMana := s.currentMaxMana()
	player := &entity{kind: playerEntity, pos: world.Player.Position, hp: maxHP, maxHP: maxHP, mana: maxMana, maxMana: maxMana, displayName: "Hero"}
	player.id = s.alloc()
	s.playerID = player.id
	level.entities[player.id] = player
	s.players[player.id] = &playerState{
		PlayerID:              player.id,
		DisplayName:           "Hero",
		Role:                  "host",
		Connected:             true,
		CurrentLevel:          level.levelNum,
		Equipped:              s.equipped,
		Hotbar:                s.hotbar,
		DiscoveredTeleporters: s.discoveredTeleporters,
		Progression:           s.progression,
		SkillCooldowns:        cloneSkillCooldowns(s.skillCooldowns),
		SkillEffects:          cloneSkillEffects(s.skillEffects),
		SkillFunctionKeys:     cloneStringSlice(s.skillFunctionKeys),
		RightClickSkillID:     s.rightClickSkillID,
		ShopStock:             s.shopStock,
		Gold:                  s.gold,
		StashItems:            s.stashItems,
		StashGold:             s.stashGold,
		StashCapacity:         s.stashCapacity,
	}

	for _, preset := range world.Entities {
		switch preset.Type {
		case monsterEntity:
			def := s.rules.Monsters[preset.MonsterDefID]
			monster := &entity{
				kind:         monsterEntity,
				pos:          preset.Position,
				spawnPos:     preset.Position,
				hp:           def.MaxHP,
				maxHP:        def.MaxHP,
				monsterDefID: preset.MonsterDefID,
				lootTable:    def.LootTable,
				aiMode:       monsterAIModeIdle,
			}
			s.applyPartyHPScale(level, monster)
			monster.id = s.alloc()
			level.entities[monster.id] = monster
		case lootEntity:
			loot := s.newLootEntity(preset.ItemDefID, preset.Position, nil, goldRollContext{levelNum: level.levelNum})
			if preset.ItemTemplateID != "" {
				rolled, ok := s.rollItemTemplate(preset.ItemTemplateID)
				if !ok {
					return ErrUnknownWorldEntity{WorldID: worldID, EntityType: preset.Type}
				}
				loot.itemDefID = rolled.ItemTemplateID
				loot.rollPayload = &rolled
			}
			loot.id = s.alloc()
			level.entities[loot.id] = loot
		case wallEntity:
			level.walls = append(level.walls, wallObstacle{pos: preset.Position, size: preset.Size, source: "preset"})
		case interactableEntity:
			def := s.rules.Interactables[preset.InteractableDefID]
			interactable := &entity{
				kind:              interactableEntity,
				pos:               preset.Position,
				interactableDefID: preset.InteractableDefID,
				state:             def.InitialState,
			}
			interactable.id = s.alloc()
			level.entities[interactable.id] = interactable
		default:
			return ErrUnknownWorldEntity{WorldID: worldID, EntityType: preset.Type}
		}
	}
	return nil
}

// ErrUnknownWorld reports an unknown world preset.
type ErrUnknownWorld struct {
	WorldID string
}

func (e ErrUnknownWorld) Error() string {
	return "game: unknown world " + e.WorldID
}

// ErrUnknownWorldEntity reports invalid world preset data that escaped rules validation.
type ErrUnknownWorldEntity struct {
	WorldID    string
	EntityType string
}

func (e ErrUnknownWorldEntity) Error() string {
	return "game: unknown world entity " + e.WorldID + ": " + e.EntityType
}

// PersistedItem is a durable inventory item reloaded on session resume.
type PersistedItem struct {
	InstanceID  string
	ItemDefID   string
	Slot        string
	Equipped    bool
	RolledStats json.RawMessage
}

// PersistedHotbarSlot is a durable hotbar assignment reloaded on session resume.
type PersistedHotbarSlot struct {
	SlotIndex      int
	ItemInstanceID *string
}

// PersistedSkillBindings is the durable skill control layout reloaded on resume.
type PersistedSkillBindings struct {
	FunctionKeys      []string
	RightClickSkillID string
}

// PersistedStashItem is an account-stash item reloaded at session start.
type PersistedStashItem struct {
	StashItemID string
	ItemDefID   string
	RolledStats json.RawMessage
}

// LoadInventory restores persisted inventory into a fresh sim (used on resume).
// The entity counter is advanced past any reloaded instance id so newly
// allocated ids never collide with reloaded ones.
func (s *Sim) LoadInventory(items []PersistedItem) {
	for _, p := range items {
		id, err := strconv.ParseUint(p.InstanceID, 10, 64)
		if err != nil {
			continue
		}
		it := &invItem{instanceID: id, itemDefID: p.ItemDefID, slot: p.Slot, equipped: p.Equipped, rollPayload: parseRollPayload(p.RolledStats)}
		s.inventory = append(s.inventory, it)
		if p.Equipped && p.Slot != "" {
			s.equipped[p.Slot] = id
		}
		if id >= s.nextID {
			s.nextID = id + 1
		}
	}
	s.savePlayer(s.defaultPlayer())
}

// LoadHotbar restores fixed hotbar assignments into a fresh sim.
func (s *Sim) LoadHotbar(slots []PersistedHotbarSlot) {
	if len(s.hotbar) != 10 {
		s.hotbar = make([]uint64, 10)
	}
	for _, slot := range slots {
		if slot.SlotIndex < 0 || slot.SlotIndex >= len(s.hotbar) {
			continue
		}
		if slot.ItemInstanceID == nil || *slot.ItemInstanceID == "" {
			s.hotbar[slot.SlotIndex] = 0
			continue
		}
		id, err := strconv.ParseUint(*slot.ItemInstanceID, 10, 64)
		if err != nil {
			continue
		}
		s.hotbar[slot.SlotIndex] = id
	}
	s.savePlayer(s.defaultPlayer())
}

func (s *Sim) LoadSkillBindings(bindings PersistedSkillBindings) {
	s.skillFunctionKeys = normalizeSkillFunctionKeys(bindings.FunctionKeys)
	s.rightClickSkillID = bindings.RightClickSkillID
	s.savePlayer(s.defaultPlayer())
}

// LoadShopStock restores durable generated shop stock into a fresh sim. Buyback
// rows are session-local and are intentionally not loaded here.
func (s *Sim) LoadShopStock(items []PersistedShopStockItem) {
	if s.shopStock == nil {
		s.shopStock = make(map[string]*shopStockState)
	}
	for _, p := range items {
		if p.ShopID == "" || p.OfferID == "" || p.ItemTemplateID == "" {
			continue
		}
		payload := parseRollPayload(p.RolledPayload)
		if payload == nil {
			continue
		}
		state := s.shopStock[p.ShopID]
		if state == nil {
			state = &shopStockState{RefreshKey: p.RefreshKey}
			s.shopStock[p.ShopID] = state
		}
		if state.RefreshKey == "" {
			state.RefreshKey = p.RefreshKey
		}
		state.Generated = append(state.Generated, &shopStockItem{
			OfferIndex:     p.OfferIndex,
			OfferID:        p.OfferID,
			SourceDepth:    p.SourceDepth,
			ItemTemplateID: p.ItemTemplateID,
			Payload:        *payload,
			BuyPrice:       p.BuyPrice,
			Available:      p.Available,
		})
	}
	for _, shopID := range sortedStringKeys(s.shopStock) {
		state := s.shopStock[shopID]
		sort.Slice(state.Generated, func(i, j int) bool {
			if state.Generated[i].OfferIndex != state.Generated[j].OfferIndex {
				return state.Generated[i].OfferIndex < state.Generated[j].OfferIndex
			}
			return state.Generated[i].OfferID < state.Generated[j].OfferID
		})
	}
	s.savePlayer(s.defaultPlayer())
}

// LoadAccountStash restores account-owned stash contents into the active
// player's private state.
func (s *Sim) LoadAccountStash(items []PersistedStashItem, gold int, capacity int) {
	if capacity <= 0 {
		capacity = defaultStashCapacity
	}
	s.stashItems = []*stashItem{}
	for _, p := range items {
		id, err := strconv.ParseUint(p.StashItemID, 10, 64)
		if err != nil || p.ItemDefID == "" {
			continue
		}
		s.stashItems = append(s.stashItems, &stashItem{
			stashItemID: id,
			itemDefID:   p.ItemDefID,
			rollPayload: parseRollPayload(p.RolledStats),
		})
		if id >= s.nextID {
			s.nextID = id + 1
		}
	}
	sort.Slice(s.stashItems, func(i, j int) bool {
		return s.stashItems[i].stashItemID < s.stashItems[j].stashItemID
	})
	if gold < 0 {
		gold = 0
	}
	s.stashGold = gold
	s.stashCapacity = capacity
	s.savePlayer(s.defaultPlayer())
}

func parseRollPayload(raw json.RawMessage) *ItemRollPayload {
	if len(raw) == 0 || string(raw) == "{}" {
		return nil
	}
	var payload ItemRollPayload
	if err := json.Unmarshal(raw, &payload); err != nil || payload.ItemTemplateID == "" {
		return nil
	}
	payload.Stats = cloneIntMap(payload.Stats)
	payload.Requirements = cloneIntMap(payload.Requirements)
	payload.EffectIDs = cloneStringSlice(payload.EffectIDs)
	return &payload
}

func cloneRollPayload(in *ItemRollPayload) *ItemRollPayload {
	if in == nil {
		return nil
	}
	return &ItemRollPayload{
		ItemTemplateID: in.ItemTemplateID,
		DisplayName:    in.DisplayName,
		Rarity:         in.Rarity,
		Stats:          cloneIntMap(in.Stats),
		Requirements:   cloneIntMap(in.Requirements),
		EffectIDs:      cloneStringSlice(in.EffectIDs),
	}
}

func cloneIntMap(in map[string]int) map[string]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int, len(in))
	for key, value := range in { //nolint:determinism — pure map clone, output is a map
		out[key] = value
	}
	return out
}

func cloneStringSlice(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func normalizeSkillFunctionKeys(in []string) []string {
	out := make([]string, skillFunctionKeyCount)
	copy(out, in)
	return out
}

// LoadDiscoveredTeleporters restores durable character waypoint unlocks into a
// fresh session. Town remains discovered even if callers omit it.
func (s *Sim) LoadDiscoveredTeleporters(levels []int) {
	if !s.multiLevel {
		return
	}
	s.discoveredTeleporters[townLevel] = true
	for _, level := range levels {
		if s.levelHasTeleporter(level) {
			s.discoveredTeleporters[level] = true
		}
	}
	s.savePlayer(s.defaultPlayer())
}

func (s *Sim) LoadInventoryForPlayer(playerID uint64, items []PersistedItem) {
	ps := s.players[playerID]
	if ps == nil {
		return
	}
	s.usePlayer(ps)
	s.LoadInventory(items)
	s.savePlayer(ps)
	s.usePlayer(s.defaultPlayer())
}

func (s *Sim) LoadHotbarForPlayer(playerID uint64, slots []PersistedHotbarSlot) {
	ps := s.players[playerID]
	if ps == nil {
		return
	}
	s.usePlayer(ps)
	s.LoadHotbar(slots)
	s.savePlayer(ps)
	s.usePlayer(s.defaultPlayer())
}

func (s *Sim) LoadSkillBindingsForPlayer(playerID uint64, bindings PersistedSkillBindings) {
	ps := s.players[playerID]
	if ps == nil {
		return
	}
	s.usePlayer(ps)
	s.LoadSkillBindings(bindings)
	s.savePlayer(ps)
	s.usePlayer(s.defaultPlayer())
}

func (s *Sim) LoadShopStockForPlayer(playerID uint64, items []PersistedShopStockItem) {
	ps := s.players[playerID]
	if ps == nil {
		return
	}
	s.usePlayer(ps)
	s.LoadShopStock(items)
	s.savePlayer(ps)
	s.usePlayer(s.defaultPlayer())
}

func (s *Sim) LoadAccountStashForPlayer(playerID uint64, items []PersistedStashItem, gold int, capacity int) {
	ps := s.players[playerID]
	if ps == nil {
		return
	}
	s.usePlayer(ps)
	s.LoadAccountStash(items, gold, capacity)
	s.savePlayer(ps)
	s.usePlayer(s.defaultPlayer())
}

func (s *Sim) LoadDiscoveredTeleportersForPlayer(playerID uint64, levels []int) {
	ps := s.players[playerID]
	if ps == nil {
		return
	}
	s.usePlayer(ps)
	s.LoadDiscoveredTeleporters(levels)
	s.savePlayer(ps)
	s.usePlayer(s.defaultPlayer())
}

func (s *Sim) alloc() uint64 {
	id := s.nextID
	s.nextID++
	return id
}

// AddGuestPlayer creates another connected player in level 0 town. It is the
// deterministic co-op join path; player-vs-player collision remains disabled.
func (s *Sim) AddGuestPlayer(accountID, characterID, displayName string, progression CharacterProgressionState) (uint64, error) {
	if displayName == "" {
		displayName = "Guest"
	}
	level, err := s.ensureTravelLevel(townLevel)
	if err != nil {
		return 0, err
	}
	spawn := s.findTownSpawnPosition(level)
	progression = s.rules.normalizeProgressionState(progression)
	equipped := newEquippedMap()
	hotbar := make([]uint64, maxHotbarCapacity)
	discovered := map[int]bool{townLevel: true}
	cooldowns := make(map[string]skillCooldownState)
	effects := make(map[string]skillEffectState)
	shopStock := make(map[string]*shopStockState)
	stashItems := []*stashItem{}
	stashCapacity := defaultStashCapacity
	character := progression
	gold := progression.Gold
	s.equipped = equipped
	s.hotbar = hotbar
	s.discoveredTeleporters = discovered
	s.progression = character
	s.skillCooldowns = cooldowns
	s.skillEffects = effects
	s.shopStock = shopStock
	s.gold = gold
	s.stashItems = stashItems
	s.stashGold = 0
	s.stashCapacity = stashCapacity
	maxHP := s.currentMaxHP()
	maxMana := s.currentMaxMana()
	player := &entity{
		kind:        playerEntity,
		pos:         spawn,
		hp:          maxHP,
		maxHP:       maxHP,
		mana:        maxMana,
		maxMana:     maxMana,
		characterID: characterID,
		displayName: displayName,
	}
	player.id = s.alloc()
	level.entities[player.id] = player
	s.players[player.id] = &playerState{
		PlayerID:              player.id,
		AccountID:             accountID,
		CharacterID:           characterID,
		DisplayName:           displayName,
		Role:                  "guest",
		Connected:             true,
		CurrentLevel:          townLevel,
		Equipped:              equipped,
		Hotbar:                hotbar,
		DiscoveredTeleporters: discovered,
		Progression:           character,
		SkillCooldowns:        cooldowns,
		SkillEffects:          effects,
		ShopStock:             shopStock,
		Gold:                  gold,
		StashItems:            stashItems,
		StashGold:             0,
		StashCapacity:         stashCapacity,
	}
	s.usePlayer(s.defaultPlayer())
	return player.id, nil
}

// SetPlayerMetadata fills party/player metadata for an existing player, usually
// the host player created with the Sim.
func (s *Sim) SetPlayerMetadata(playerID uint64, accountID, characterID, displayName, role string) {
	ps := s.players[playerID]
	if ps == nil {
		return
	}
	if displayName == "" {
		displayName = ps.DisplayName
	}
	if role == "" {
		role = ps.Role
	}
	ps.AccountID = accountID
	ps.CharacterID = characterID
	ps.DisplayName = displayName
	ps.Role = role
	if e := s.levels[ps.CurrentLevel].entities[playerID]; e != nil {
		e.characterID = characterID
		e.displayName = displayName
	}
}

func (s *Sim) SetPlayerConnected(playerID uint64, connected bool) {
	if ps := s.players[playerID]; ps != nil {
		ps.Connected = connected
	}
}

func (s *Sim) DefaultPlayerID() uint64 {
	if ps := s.defaultPlayer(); ps != nil {
		return ps.PlayerID
	}
	return 0
}

func (s *Sim) PlayerCurrentLevel(playerID uint64) (int, bool) {
	ps := s.players[playerID]
	if ps == nil {
		return 0, false
	}
	return ps.CurrentLevel, true
}

func (s *Sim) PlayerConnected(playerID uint64) bool {
	ps := s.players[playerID]
	return ps != nil && ps.Connected
}

func (s *Sim) PlayerIDs() []uint64 {
	return sortedPlayerIDs(s.players)
}

func (s *Sim) PlayerIDForCharacter(characterID string) (uint64, bool) {
	if characterID == "" {
		return 0, false
	}
	for _, playerID := range sortedPlayerIDs(s.players) {
		ps := s.players[playerID]
		if ps != nil && ps.CharacterID == characterID {
			return playerID, true
		}
	}
	return 0, false
}

func ParseEntityID(id string) (uint64, bool) {
	n, err := strconv.ParseUint(id, 10, 64)
	return n, err == nil
}

func (s *Sim) RemovePlayerEntity(playerID uint64) {
	ps := s.players[playerID]
	if ps == nil {
		return
	}
	if level := s.levels[ps.CurrentLevel]; level != nil {
		delete(level.entities, playerID)
	}
	ps.Connected = false
	if s.playerID == playerID {
		s.usePlayer(s.defaultPlayer())
	}
}

func (s *Sim) RespawnPlayerInTown(playerID uint64) error {
	ps := s.players[playerID]
	if ps == nil {
		return fmt.Errorf("game: unknown player %d", playerID)
	}
	level, err := s.ensureTravelLevel(townLevel)
	if err != nil {
		return err
	}
	for _, lvl := range s.levels {
		delete(lvl.entities, playerID)
	}
	s.usePlayer(ps)
	maxHP := s.currentMaxHP()
	player := &entity{
		id:          playerID,
		kind:        playerEntity,
		pos:         s.findTownSpawnPosition(level),
		hp:          maxHP,
		maxHP:       maxHP,
		mana:        s.currentMaxMana(),
		maxMana:     s.currentMaxMana(),
		characterID: ps.CharacterID,
		displayName: ps.DisplayName,
	}
	level.entities[playerID] = player
	s.currentLevel = townLevel
	ps.CurrentLevel = townLevel
	ps.Connected = true
	s.savePlayer(ps)
	s.usePlayer(s.defaultPlayer())
	return nil
}

func (s *Sim) findTownSpawnPosition(level *LevelState) Vec2 {
	preferred := s.preferredTownSpawnPosition(level)
	if !s.spawnPositionBlocked(level, preferred) {
		return preferred
	}
	nav := s.navigationForLevel(level)
	step := nav.CellSize
	if step <= 0 {
		step = 1.0
	}
	for ring := 1; ring <= 8; ring++ {
		for dy := -ring; dy <= ring; dy++ {
			for dx := -ring; dx <= ring; dx++ {
				if absInt(dx) != ring && absInt(dy) != ring {
					continue
				}
				candidate := Vec2{
					X: preferred.X + float64(dx)*step,
					Y: preferred.Y + float64(dy)*step,
				}
				if !s.positionInNavigationBounds(nav, candidate) {
					continue
				}
				if !s.spawnPositionBlocked(level, candidate) {
					return candidate
				}
			}
		}
	}
	return preferred
}

func (s *Sim) preferredTownSpawnPosition(level *LevelState) Vec2 {
	if host := s.players[s.playerID]; host != nil {
		if lvl := s.levels[host.CurrentLevel]; lvl != nil && host.CurrentLevel == level.levelNum {
			if e := lvl.entities[host.PlayerID]; e != nil {
				return e.pos
			}
		}
	}
	for _, id := range sortedEntityIDs(level.entities) {
		e := level.entities[id]
		if e != nil && e.kind == playerEntity {
			return e.pos
		}
	}
	return Vec2{X: 4, Y: 10}
}

func (s *Sim) spawnPositionBlocked(level *LevelState, pos Vec2) bool {
	if level == nil {
		return true
	}
	for _, wall := range level.walls {
		if circleIntersectsAABB(pos, playerRadius, wall.pos, wall.size) {
			return true
		}
	}
	for _, id := range sortedEntityIDs(level.entities) {
		e := level.entities[id]
		if e == nil {
			continue
		}
		switch e.kind {
		case playerEntity:
			if e.hp > 0 && circlesOverlap(pos, playerRadius, e.pos, playerRadius) {
				return true
			}
		case monsterEntity:
			if e.hp > 0 && circlesOverlap(pos, playerRadius, e.pos, monsterRadius) {
				return true
			}
		case interactableEntity:
			if e.state == interactableClosed {
				if def, ok := s.rules.Interactables[e.interactableDefID]; ok && def.BarrierWhenClosed != nil {
					if circleIntersectsAABB(pos, playerRadius, e.pos, def.BarrierWhenClosed.Size) {
						return true
					}
				}
			}
		}
	}
	return false
}

func (s *Sim) navigationForLevel(level *LevelState) NavigationRules {
	if level != nil && level.nav != nil {
		return *level.nav
	}
	return s.rules.Navigation
}

func (s *Sim) positionInNavigationBounds(nav NavigationRules, pos Vec2) bool {
	cell := worldToGrid(nav, pos)
	return cellInBounds(nav, cell)
}

func (s *Sim) defaultPlayer() *playerState {
	if ps := s.players[s.playerID]; ps != nil {
		return ps
	}
	for _, id := range sortedPlayerIDs(s.players) {
		return s.players[id]
	}
	return nil
}

func (s *Sim) playerForInput(in Input) *playerState {
	if in.ActorPlayerID != 0 {
		return s.players[in.ActorPlayerID]
	}
	return s.defaultPlayer()
}

func (s *Sim) usePlayer(ps *playerState) {
	if ps == nil {
		return
	}
	s.playerID = ps.PlayerID
	s.currentLevel = ps.CurrentLevel
	s.inventory = ps.Inventory
	s.equipped = ps.Equipped
	s.hotbar = ps.Hotbar
	s.discoveredTeleporters = ps.DiscoveredTeleporters
	s.progression = ps.Progression
	s.skillCooldowns = ps.SkillCooldowns
	if s.skillCooldowns == nil {
		s.skillCooldowns = make(map[string]skillCooldownState)
	}
	s.skillEffects = ps.SkillEffects
	if s.skillEffects == nil {
		s.skillEffects = make(map[string]skillEffectState)
	}
	s.skillFunctionKeys = normalizeSkillFunctionKeys(ps.SkillFunctionKeys)
	s.rightClickSkillID = ps.RightClickSkillID
	s.shopStock = ps.ShopStock
	if s.shopStock == nil {
		s.shopStock = make(map[string]*shopStockState)
	}
	s.gold = ps.Gold
	s.stashItems = ps.StashItems
	s.stashGold = ps.StashGold
	s.stashCapacity = ps.StashCapacity
	if s.stashCapacity <= 0 {
		s.stashCapacity = defaultStashCapacity
	}
	s.hpRegenCarry = ps.HPRegenCarry
	s.manaRegenCarry = ps.ManaRegenCarry
	level := s.activeLevel()
	level.move = ps.Move
	level.autoNav = ps.AutoNav
	s.syncCompatibilityFields()
}

func (s *Sim) savePlayer(ps *playerState) {
	if ps == nil {
		return
	}
	ps.CurrentLevel = s.currentLevel
	ps.Inventory = s.inventory
	ps.Equipped = s.equipped
	ps.Hotbar = s.hotbar
	ps.DiscoveredTeleporters = s.discoveredTeleporters
	ps.Progression = s.progression
	ps.SkillCooldowns = s.skillCooldowns
	ps.SkillEffects = s.skillEffects
	ps.SkillFunctionKeys = normalizeSkillFunctionKeys(s.skillFunctionKeys)
	ps.RightClickSkillID = s.rightClickSkillID
	ps.ShopStock = s.shopStock
	ps.Gold = s.gold
	ps.StashItems = s.stashItems
	ps.StashGold = s.stashGold
	ps.StashCapacity = s.stashCapacity
	ps.HPRegenCarry = s.hpRegenCarry
	ps.ManaRegenCarry = s.manaRegenCarry
	if level := s.levels[ps.CurrentLevel]; level != nil {
		ps.Move = level.move
		ps.AutoNav = level.autoNav
	}
}

func sortedPlayerIDs(players map[uint64]*playerState) []uint64 {
	ids := make([]uint64, 0, len(players))
	for id := range players {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// CurrentTick returns the next tick to be processed.
func (s *Sim) CurrentTick() uint64 { return s.tick }

// Input is a decoded client intent applied to a specific tick.
type Input struct {
	MessageID          string
	CorrelationID      string
	Sequence           int64
	ActorPlayerID      uint64
	Type               string
	Move               *MoveIntent
	MoveTo             *MoveToIntent
	DirectionalAttack  *DirectionalAttackIntent
	Action             *ActionIntent
	Descend            *DescendIntent
	Ascend             *AscendIntent
	Teleport           *TeleportIntent
	Equip              *EquipIntent
	Unequip            *UnequipIntent
	Drop               *DropIntent
	Use                *UseIntent
	AssignHotbar       *AssignHotbarIntent
	UseHotbar          *UseHotbarIntent
	AllocateStat       *AllocateStatIntent
	AllocateSkillPoint *AllocateSkillPointIntent
	CastSkill          *CastSkillIntent
	SetSkillBindings   *SetSkillBindingsIntent
	ShopBuy            *ShopBuyIntent
	ShopSell           *ShopSellIntent
	StashDepositItem   *StashDepositItemIntent
	StashWithdrawItem  *StashWithdrawItemIntent
	StashDepositGold   *StashDepositGoldIntent
	StashWithdrawGold  *StashWithdrawGoldIntent
}

// Intent payloads.
type (
	MoveIntent struct {
		Direction     Vec2
		DurationTicks int
	}
	MoveToIntent struct {
		Position Vec2
	}
	DirectionalAttackIntent struct {
		Direction Vec2
	}
	ActionIntent   struct{ TargetID string }
	DescendIntent  struct{}
	AscendIntent   struct{}
	TeleportIntent struct {
		TargetLevel int
	}
	EquipIntent struct {
		ItemInstanceID string
		Slot           string
	}
	UnequipIntent struct {
		Slot string
	}
	DropIntent struct {
		ItemInstanceID string
	}
	UseIntent struct {
		ItemInstanceID string
	}
	AssignHotbarIntent struct {
		SlotIndex      int
		ItemInstanceID *string
	}
	UseHotbarIntent struct {
		SlotIndex int
	}
	AllocateStatIntent struct {
		Stat   string
		Points int
	}
	AllocateSkillPointIntent struct {
		SkillID string
	}
	CastSkillIntent struct {
		SkillID   string
		TargetID  string
		Direction *Vec2
	}
	SetSkillBindingsIntent struct {
		FunctionKeys      []string
		RightClickSkillID string
	}
	ShopBuyIntent struct {
		ShopEntityID string
		OfferID      string
	}
	ShopSellIntent struct {
		ShopEntityID   string
		ItemInstanceID string
	}
	StashDepositItemIntent struct {
		StashEntityID  string
		ItemInstanceID string
	}
	StashWithdrawItemIntent struct {
		StashEntityID string
		StashItemID   string
	}
	StashDepositGoldIntent struct {
		StashEntityID string
		Amount        int
	}
	StashWithdrawGoldIntent struct {
		StashEntityID string
		Amount        int
	}
)

// Ack/Reject report per-input acceptance.
type (
	Ack    struct{ MessageID string }
	Reject struct {
		MessageID string
		Reason    string
	}
)

// TickResult is everything a single tick produced.
type TickResult struct {
	Tick          uint64
	Level         int
	ActorPlayerID uint64
	Changes       []Change
	Events        []Event
	Acks          []Ack
	Rejects       []Reject
}

func (r *TickResult) ack(id string) { r.Acks = append(r.Acks, Ack{MessageID: id}) }
func (r *TickResult) reject(id, reason string) {
	r.Rejects = append(r.Rejects, Reject{MessageID: id, Reason: reason})
}

// Tick processes one authoritative tick and returns the normal single-level
// result. Runtime protocol code should use TickResults so stair transitions can
// emit scoped from/to level deltas.
func (s *Sim) Tick(inputs []Input) TickResult {
	results := s.TickResults(inputs)
	if len(results) == 0 {
		return TickResult{Tick: s.tick, Level: s.currentLevel, Changes: []Change{}, Events: []Event{}}
	}
	for _, res := range results {
		if len(res.Acks) > 0 || len(res.Rejects) > 0 {
			return res
		}
	}
	return results[len(results)-1]
}

// TickResults processes the inputs stamped for the current tick (already
// ordered by the runner as (sequence, message_id)), applies continuous
// movement, advances the tick counter, and returns one or more scoped results.
func (s *Sim) TickResults(inputs []Input) []TickResult {
	type resultKey struct {
		level int
		actor uint64
	}
	resultByKey := map[resultKey]*TickResult{}
	var ordered []*TickResult
	transitionThisTick := false
	resultFor := func(level int, actor uint64) *TickResult {
		key := resultKey{level: level, actor: actor}
		if res := resultByKey[key]; res != nil {
			return res
		}
		res := &TickResult{Tick: s.tick, Level: level, ActorPlayerID: actor, Changes: []Change{}, Events: []Event{}}
		resultByKey[key] = res
		ordered = append(ordered, res)
		return res
	}

	for _, in := range inputs {
		ps := s.playerForInput(in)
		if ps == nil || !ps.Connected {
			res := resultFor(s.currentLevel, 0)
			res.reject(in.MessageID, "unknown_actor")
			continue
		}
		s.usePlayer(ps)
		res := resultFor(ps.CurrentLevel, ps.PlayerID)
		if in.Type == "descend_intent" || in.Type == "ascend_intent" || in.Type == "teleport_intent" {
			if arrival := s.handleLevelTravel(in, res); arrival != nil {
				arrival.ActorPlayerID = ps.PlayerID
				ordered = append(ordered, arrival)
				transitionThisTick = true
			}
			s.savePlayer(ps)
			continue
		}
		s.applyInput(in, res)
		s.savePlayer(ps)
	}

	if !transitionThisTick {
		for _, playerID := range sortedPlayerIDs(s.players) {
			ps := s.players[playerID]
			if ps == nil || !ps.Connected {
				continue
			}
			s.usePlayer(ps)
			res := resultFor(ps.CurrentLevel, ps.PlayerID)
			s.expireSkillEffects(res)
			s.applyMovement(res)
			s.applyPlayerRegen(res)
			s.savePlayer(ps)
		}

		s.autoPickUpGold(resultFor)

		for _, levelNum := range s.sortedLevelNums() {
			s.currentLevel = levelNum
			s.syncCompatibilityFields()
			res := resultFor(levelNum, 0)
			s.advanceMonsterMovement(res)
			s.advanceBossPhases(res)
			s.advanceMonsterAttack(res)
			s.advanceProjectiles(res)
		}
	}

	s.tick++
	s.usePlayer(s.defaultPlayer())

	results := make([]TickResult, 0, len(ordered))
	for _, res := range ordered {
		if len(res.Changes) == 0 && len(res.Events) == 0 && len(res.Acks) == 0 && len(res.Rejects) == 0 {
			continue
		}
		results = append(results, *res)
	}
	if len(results) == 0 {
		return []TickResult{{Tick: s.tick - 1, Level: s.currentLevel, Changes: []Change{}, Events: []Event{}}}
	}
	return results
}

func (s *Sim) applyInput(in Input, res *TickResult) {
	if in.Type != "client_ready" && s.playerDead() {
		if _, ok := inputHandlers[in.Type]; ok {
			res.reject(in.MessageID, "player_dead")
			return
		}
	}
	if h, ok := inputHandlers[in.Type]; ok {
		h(s, in, res)
		return
	}
	res.reject(in.MessageID, "unknown_type")
}

func (s *Sim) applyPlayerRegen(res *TickResult) {
	player := s.activeLevel().entities[s.playerID]
	if player == nil || player.hp <= 0 {
		return
	}
	stats, _ := s.playerEffectiveCombatStats()
	changed := false
	if delta := regenAmount(stats.HealthRegenPerSecond, tickDuration, &s.hpRegenCarry, player.maxHP-player.hp); delta > 0 {
		player.hp += delta
		changed = true
	}
	if delta := regenAmount(stats.ManaRegenPerSecond, tickDuration, &s.manaRegenCarry, player.maxMana-player.mana); delta > 0 {
		player.mana += delta
		changed = true
	}
	if changed {
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	}
}

func (s *Sim) expireSkillEffects(res *TickResult) {
	if len(s.skillEffects) == 0 {
		return
	}
	player := s.activeLevel().entities[s.playerID]
	changed := false
	for _, skillID := range sortedStringKeys(s.skillEffects) {
		effect := s.skillEffects[skillID]
		if effect.EndsTick > s.tick {
			continue
		}
		delete(s.skillEffects, skillID)
		changed = true
		if player != nil {
			res.Events = append(res.Events, Event{
				EventType: "skill_effect_ended",
				EntityID:  idStr(player.id),
				SkillID:   skillID,
			})
		}
	}
	if !changed || player == nil {
		return
	}
	resourcesChanged := s.syncActivePlayerMaxResources()
	visualChanged := s.syncActivePlayerVisualScale()
	if resourcesChanged || visualChanged || player.hp > 0 {
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	}
}

func (s *Sim) syncActivePlayerMaxResources() bool {
	player := s.activeLevel().entities[s.playerID]
	if player == nil {
		return false
	}
	changed := false
	maxHP := s.currentMaxHP()
	if maxHP != player.maxHP {
		player.maxHP = maxHP
		if player.hp > player.maxHP {
			player.hp = player.maxHP
		}
		changed = true
	}
	maxMana := s.currentMaxMana()
	if maxMana != player.maxMana {
		player.maxMana = maxMana
		if player.mana > player.maxMana {
			player.mana = player.maxMana
		}
		changed = true
	}
	return changed
}

func regenAmount(ratePerSecond float64, secondsPerTick float64, carry *float64, missing int) int {
	if missing <= 0 || ratePerSecond <= 0 || secondsPerTick <= 0 {
		*carry = 0
		return 0
	}
	*carry += ratePerSecond * secondsPerTick
	amount := int(math.Floor(*carry + 0.000000001))
	if amount <= 0 {
		return 0
	}
	if amount > missing {
		amount = missing
	}
	*carry -= float64(amount)
	if amount == missing {
		*carry = 0
	}
	return amount
}

func (s *Sim) activeLevel() *LevelState {
	level := s.levels[s.currentLevel]
	if level == nil {
		panic("game: active level missing")
	}
	return level
}

func (s *Sim) activeNav() NavigationRules {
	level := s.activeLevel()
	if level.nav == nil {
		return s.rules.Navigation
	}
	return *level.nav
}

func (s *Sim) activeWalls() []wallObstacle {
	level := s.activeLevel()
	if s.walls != nil {
		level.walls = s.walls
	}
	return level.walls
}

func (s *Sim) syncCompatibilityFields() {
	level := s.activeLevel()
	s.entities = level.entities
	s.walls = level.walls
	s.move = level.move
	s.autoNav = level.autoNav
}

func (s *Sim) ensureDungeonLevel(levelNum int) (*LevelState, error) {
	if levelNum >= levelZero {
		return nil, fmt.Errorf("game: invalid dungeon level %d", levelNum)
	}
	if level, ok := s.levels[levelNum]; ok {
		return level, nil
	}
	nav := dungeonNavigationForLevel(s.rules.Navigation, s.rules.DungeonGeneration, levelNum)
	level := newLevelState(levelNum, &nav)
	s.levels[levelNum] = level
	if err := s.populateDungeonLevel(level); err != nil {
		delete(s.levels, levelNum)
		return nil, err
	}
	return level, nil
}

func (s *Sim) ensureTravelLevel(levelNum int) (*LevelState, error) {
	if levelNum == townLevel {
		level, ok := s.levels[townLevel]
		if !ok {
			return nil, fmt.Errorf("game: missing town level")
		}
		return level, nil
	}
	return s.ensureDungeonLevel(levelNum)
}

func (s *Sim) populateDungeonLevel(level *LevelState) error {
	gen, err := GenerateDungeonLevel(s.seed, level.levelNum, s.rules.DungeonGeneration)
	if err != nil {
		return err
	}
	level.walls = gen.walls
	for _, stair := range gen.stairs {
		def := s.rules.Interactables[stair.defID]
		state := def.InitialState
		if stair.state != "" {
			state = stair.state
		}
		e := &entity{
			kind:              interactableEntity,
			pos:               stair.pos,
			interactableDefID: stair.defID,
			state:             state,
		}
		e.id = s.alloc()
		level.entities[e.id] = e
	}
	for _, teleporter := range gen.teleporters {
		def := s.rules.Interactables[teleporter.defID]
		state := def.InitialState
		if teleporter.state != "" {
			state = teleporter.state
		}
		e := &entity{
			kind:              interactableEntity,
			pos:               teleporter.pos,
			interactableDefID: teleporter.defID,
			state:             state,
		}
		e.id = s.alloc()
		level.entities[e.id] = e
	}
	for _, chest := range gen.chests {
		def := s.rules.Interactables[chest.defID]
		e := &entity{
			kind:              interactableEntity,
			pos:               chest.pos,
			interactableDefID: chest.defID,
			state:             def.InitialState,
			lootTable:         chest.lootTable,
		}
		e.id = s.alloc()
		level.entities[e.id] = e
	}
	for _, generated := range gen.loot {
		if _, ok := s.rules.Items[generated.itemDefID]; !ok {
			return fmt.Errorf("game: generate dungeon level %d: unknown loot item %s", level.levelNum, generated.itemDefID)
		}
		loot := s.newLootEntity(generated.itemDefID, generated.pos, nil, goldRollContext{levelNum: level.levelNum})
		loot.id = s.alloc()
		level.entities[loot.id] = loot
	}
	for _, generated := range gen.monsters {
		def, ok := s.rules.Monsters[generated.defID]
		if !ok {
			return fmt.Errorf("game: generate dungeon level %d: unknown monster %s", level.levelNum, generated.defID)
		}
		lootTable := generated.lootTable
		if generated.isBoss {
			template, ok := s.rules.BossTemplates[generated.bossTemplate]
			if !ok {
				return fmt.Errorf("game: generate dungeon level %d: unknown boss template %s", level.levelNum, generated.bossTemplate)
			}
			var baseOK bool
			def, baseOK = s.rules.Monsters[template.BaseMonsterDefID]
			if !baseOK {
				return fmt.Errorf("game: generate dungeon level %d: unknown boss base monster %s", level.levelNum, template.BaseMonsterDefID)
			}
			generated.defID = template.BaseMonsterDefID
			lootTable = template.LootTable
			generated.visualModel = template.Visual.Model
			generated.visualTint = template.Visual.Color
			generated.visualScale = template.Visual.Scale
		}
		if _, ok := s.rules.LootTables[lootTable]; !ok {
			return fmt.Errorf("game: generate dungeon level %d: unknown monster loot table %s", level.levelNum, lootTable)
		}
		monster := &entity{
			kind:                 monsterEntity,
			pos:                  generated.pos,
			spawnPos:             generated.pos,
			hp:                   def.MaxHP,
			maxHP:                def.MaxHP,
			monsterDefID:         generated.defID,
			monsterRarityID:      generated.rarityID,
			lootTable:            lootTable,
			aiMode:               monsterAIModeIdle,
			isBoss:               generated.isBoss,
			bossTemplateID:       generated.bossTemplate,
			visualModel:          generated.visualModel,
			visualTint:           generated.visualTint,
			visualScale:          generated.visualScale,
			bossPhaseIndex:       -1,
			bossPatternDeckIndex: -1,
		}
		if generated.isBoss {
			template := s.rules.BossTemplates[generated.bossTemplate]
			if len(template.PatternDeck) > 0 {
				monster.bossPatternDeckIndex = 0
				monster.bossPatternID = template.PatternDeck[0]
			}
			monster.maxHP = roundPositive(float64(def.MaxHP) * template.HPMultiplier)
			monster.hp = monster.maxHP
			if def.AttackDamage != nil {
				scaledAttack := scaleDamageRange(*def.AttackDamage, template.DamageMultiplier)
				monster.monsterAttackDamage = &scaledAttack
			}
			monster.monsterXPReward = roundPositive(float64(def.XPReward) * template.HPMultiplier)
		} else if rarity, ok := s.rules.DungeonGeneration.MonsterRarity(generated.rarityID); ok {
			monster.maxHP = roundPositive(float64(def.MaxHP) * rarity.HPMultiplier)
			monster.hp = monster.maxHP
			monster.visualScale = rarity.VisualScale
			monster.visualTint = rarity.Color
			if def.AttackDamage != nil {
				scaledAttack := scaleDamageRange(*def.AttackDamage, rarity.DamageMultiplier)
				monster.monsterAttackDamage = &scaledAttack
			}
			monster.monsterXPReward = roundPositive(float64(def.XPReward) * rarity.XPMultiplier)
		}
		s.applyPartyHPScale(level, monster)
		monster.id = s.alloc()
		level.entities[monster.id] = monster
	}
	return nil
}

func (s *Sim) dispatchAction(target *entity, in Input, res *TickResult, ack bool) {
	switch target.kind {
	case monsterEntity:
		if s.playerAttackMode() == attackModeRanged {
			s.fireProjectile(target, in, res, ack)
		} else {
			s.attackTarget(target, in, res, ack)
		}
	case lootEntity:
		s.pickUpTarget(target, in, res, ack)
	case interactableEntity:
		s.activateInteractable(target, in, res, ack)
	default:
		res.reject(in.MessageID, "invalid_target")
	}
}

func (s *Sim) attackTarget(target *entity, in Input, res *TickResult, ack bool) {
	if ack {
		res.ack(in.MessageID)
	}
	s.damageMonsterByPlayer(target, s.playerID, in.CorrelationID, res, s.resolvePlayerAttackDamage())
}

func (s *Sim) damageMonsterByPlayer(target *entity, playerID uint64, corr string, res *TickResult, damageRange DamageRange) combatResolution {
	attackerStats, _ := s.playerEffectiveCombatStats()
	defenderStats := s.monsterEffectiveCombatStats(target, DamageRange{})
	outcome := s.resolveCombat(attackerStats, defenderStats, damageRange)
	if !outcome.Hit || outcome.Blocked {
		res.Events = append(res.Events, combatEvent(s.combatEventType(monsterEntity, outcome), playerID, target.id, corr, outcome))
		return outcome
	}

	target.hp -= outcome.Damage
	if target.hp < 0 {
		target.hp = 0
	}
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
	res.Events = append(res.Events, combatEvent(s.combatEventType(monsterEntity, outcome), playerID, target.id, corr, outcome))

	if outcome.Damage > 0 {
		s.aggroMonsterOnHit(target, playerID, corr, res)
	}
	if target.hp == 0 {
		s.finishMonsterKill(target, playerID, corr, res)
	}
	s.retaliate(target, corr, res)
	return outcome
}

func (s *Sim) fireProjectile(target *entity, in Input, res *TickResult, ack bool) {
	player := s.activeLevel().entities[s.playerID]
	if player == nil {
		res.reject(in.MessageID, "player_dead")
		return
	}
	dir := normalize(Vec2{X: target.pos.X - player.pos.X, Y: target.pos.Y - player.pos.Y})
	if dir.X == 0 && dir.Y == 0 {
		dir = Vec2{X: 1}
	}
	s.fireProjectileInDirection(dir, target.id, in, res, ack)
}

func (s *Sim) fireProjectileInDirection(dir Vec2, targetID uint64, in Input, res *TickResult, ack bool) {
	if s.playerProjectileInFlight() {
		res.reject(in.MessageID, "projectile_busy")
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil {
		res.reject(in.MessageID, "player_dead")
		return
	}
	projectileSpeed, ok := s.playerProjectileSpeed()
	if !ok {
		res.reject(in.MessageID, "invalid_target")
		return
	}
	dir = normalize(dir)
	if dir.X == 0 && dir.Y == 0 {
		res.reject(in.MessageID, "invalid_direction")
		return
	}
	maxDistance := s.playerActionReach()
	projectile := &entity{
		kind:            projectileEntity,
		pos:             player.pos,
		ownerID:         player.id,
		targetID:        targetID,
		projectileDefID: trainingArrowProjectileDefID,
		dir:             dir,
		speed:           projectileSpeed,
		maxDistance:     maxDistance,
		damageRange:     s.resolvePlayerAttackDamage(),
		sourceMsgID:     in.MessageID,
		sourceCorrID:    in.CorrelationID,
		spawnTick:       s.tick,
	}
	projectile.id = s.alloc()
	s.activeLevel().entities[projectile.id] = projectile
	res.Changes = append(res.Changes, Change{Op: OpEntitySpawn, Entity: ptrEntityView(s.entityView(projectile))})
	if ack {
		res.ack(in.MessageID)
	}
}

func (s *Sim) directionalMeleeTarget(dir Vec2) *entity {
	player := s.activeLevel().entities[s.playerID]
	if player == nil {
		return nil
	}
	reach := s.playerMeleeReach()
	var best *entity
	bestDist := math.MaxFloat64
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		e := s.activeLevel().entities[id]
		if e == nil || e.kind != monsterEntity || e.hp <= 0 {
			continue
		}
		toTarget := Vec2{X: e.pos.X - player.pos.X, Y: e.pos.Y - player.pos.Y}
		projection := toTarget.X*dir.X + toTarget.Y*dir.Y
		targetRadius := s.targetInteractionRadius(e)
		if projection < 0 || projection > reach+targetRadius+meleeRangeEpsilon {
			continue
		}
		lateral := math.Abs(toTarget.X*dir.Y - toTarget.Y*dir.X)
		if lateral > directionalMeleeHalfWidth+targetRadius+meleeRangeEpsilon {
			continue
		}
		dist := distance(player.pos, e.pos)
		if best == nil || dist < bestDist-1e-9 || (math.Abs(dist-bestDist) <= 1e-9 && e.id < best.id) {
			best = e
			bestDist = dist
		}
	}
	return best
}

func (s *Sim) dropLoot(monster *entity, corr string, res *TickResult) {
	drops := s.rules.LootDrops(monster.lootTable, s.rng)
	s.spawnLootDrops(drops, monster.pos, s.targetInteractionRadius(monster), corr, res, goldRollContext{
		levelNum:        s.activeLevel().levelNum,
		monsterRarityID: monster.monsterRarityID,
	})
}

func (s *Sim) finishMonsterKill(monster *entity, sourceID uint64, corr string, res *TickResult) {
	res.Events = append(res.Events, Event{
		EventType:      "monster_killed",
		EntityID:       idStr(monster.id),
		SourceEntityID: idStr(sourceID),
		TargetEntityID: idStr(monster.id),
		CorrelationID:  corr,
	})
	s.dropLoot(monster, corr, res)
	s.awardMonsterExperience(monster, sourceID, corr, res)
	if monster.isBoss {
		s.unlockBossFloorExits(corr, res)
	}
}

func (s *Sim) unlockBossFloorExits(corr string, res *TickResult) {
	level := s.activeLevel()
	for _, id := range sortedEntityIDs(level.entities) {
		e := level.entities[id]
		if e == nil || e.kind != interactableEntity {
			continue
		}
		if e.interactableDefID != stairsDownDefID && e.interactableDefID != teleporterDefID {
			continue
		}
		if e.state != interactableLocked && e.state != interactableDisabled {
			continue
		}
		e.state = interactableReady
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(e))})
		res.Events = append(res.Events, Event{
			EventType:     "interactable_state_changed",
			EntityID:      idStr(e.id),
			CorrelationID: corr,
			State:         interactableReady,
		})
	}
}

func (s *Sim) spawnLootDrops(drops []LootDrop, sourcePos Vec2, sourceRadius float64, corr string, res *TickResult, goldCtx goldRollContext) {
	var clusterAnchor Vec2
	clusterReady := false

	for i, drop := range drops {
		var dropPos Vec2
		var ok bool

		if i == 0 {
			dropPos, ok = s.findEntityLootDropPosition(sourcePos, sourceRadius)
			if !ok {
				dropPos = sourcePos
			}
			clusterAnchor = dropPos
			clusterReady = true
		} else if clusterReady {
			dropPos, ok = s.findClusterLootDropPosition(clusterAnchor, i)
			if !ok {
				dropPos, ok = s.findEntityLootDropPosition(sourcePos, sourceRadius)
				if !ok {
					dropPos = sourcePos
				}
			}
		} else {
			dropPos, ok = s.findEntityLootDropPosition(sourcePos, sourceRadius)
			if !ok {
				dropPos = sourcePos
			}
		}

		itemDefID := drop.ItemDefID
		var payload *ItemRollPayload
		if drop.ItemTemplateID != "" {
			rolled, ok := s.rollItemTemplate(drop.ItemTemplateID)
			if !ok {
				continue
			}
			payload = &rolled
			itemDefID = rolled.ItemTemplateID
		}
		loot := s.newLootEntity(itemDefID, dropPos, payload, goldCtx)
		loot.id = s.alloc()
		s.activeLevel().entities[loot.id] = loot
		res.Changes = append(res.Changes, Change{Op: OpEntitySpawn, Entity: ptrEntityView(s.entityView(loot))})
		res.Events = append(res.Events, Event{EventType: "loot_dropped", EntityID: idStr(loot.id), CorrelationID: corr})
	}
}

func (s *Sim) retaliate(monster *entity, corr string, res *TickResult) {
	if monster.isBoss {
		return
	}
	def := s.rules.Monsters[monster.monsterDefID]
	if def.RetaliationDamage == nil {
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil || player.hp <= 0 {
		return
	}
	retaliationDamage := s.scaleMonsterDamageForParty(s.currentLevel, *def.RetaliationDamage)
	attackerStats := s.monsterEffectiveCombatStats(monster, retaliationDamage)
	defenderStats, _ := s.playerEffectiveCombatStats()
	outcome := s.resolveCombat(attackerStats, defenderStats, retaliationDamage)
	if !outcome.Hit || outcome.Blocked {
		res.Events = append(res.Events, combatEvent(s.combatEventType(playerEntity, outcome), monster.id, player.id, corr, outcome))
		return
	}
	player.hp -= outcome.Damage
	if player.hp < 0 {
		player.hp = 0
	}
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	eventType := "player_damaged"
	if player.hp == 0 {
		eventType = "player_killed"
	}
	res.Events = append(res.Events, combatEvent(eventType, monster.id, player.id, corr, outcome))
}

func (s *Sim) pickUpTarget(e *entity, in Input, res *TickResult, ack bool) {
	if e.itemDefID == goldItemDefID && e.goldAmount > 0 {
		ackMessageID := ""
		if ack {
			ackMessageID = in.MessageID
		}
		s.pickUpGoldForPlayer(e, s.playerID, in.CorrelationID, ackMessageID, res)
		return
	}
	if s.bagOccupancyCount()+1 > s.inventoryCapacity() {
		res.reject(in.MessageID, "inventory_full")
		return
	}

	item := &invItem{
		instanceID:  s.alloc(),
		itemDefID:   e.itemDefID,
		rollPayload: cloneRollPayload(e.rollPayload),
		slot:        s.itemSlot(e.itemDefID, e.rollPayload),
		equipped:    false,
	}

	delete(s.activeLevel().entities, e.id)
	res.Changes = append(res.Changes, Change{Op: OpEntityRemove, EntityID: idStr(e.id)})

	s.inventory = append(s.inventory, item)
	res.Changes = append(res.Changes, Change{Op: OpInventoryAdd, Item: ptrItemView(s.itemView(item))})
	res.Events = append(res.Events, Event{
		EventType:      "item_picked_up",
		EntityID:       idStr(s.playerID),
		CorrelationID:  in.CorrelationID,
		ItemInstanceID: idStr(item.instanceID),
	})
	if ack {
		res.ack(in.MessageID)
	}
}

func (s *Sim) autoPickUpGold(resultFor func(level int, actor uint64) *TickResult) {
	for _, levelNum := range s.sortedLevelNums() {
		level := s.levels[levelNum]
		if level == nil {
			continue
		}
		for _, entityID := range sortedEntityIDs(level.entities) {
			gold := level.entities[entityID]
			if !isAutoPickableGold(gold) {
				continue
			}
			winnerID := s.goldAutoPickupWinner(levelNum, gold)
			if winnerID == 0 {
				continue
			}
			res := resultFor(levelNum, 0)
			s.pickUpGoldForPlayer(gold, winnerID, "", "", res)
		}
	}
}

func (s *Sim) goldAutoPickupWinner(levelNum int, gold *entity) uint64 {
	level := s.levels[levelNum]
	if level == nil || !isAutoPickableGold(gold) {
		return 0
	}
	for _, playerID := range sortedPlayerIDs(s.players) {
		ps := s.players[playerID]
		if ps == nil || !ps.Connected || ps.CurrentLevel != levelNum {
			continue
		}
		player := level.entities[playerID]
		if player == nil || player.hp <= 0 {
			continue
		}
		s.usePlayer(ps)
		if s.inLootPickupRangeFrom(player.pos, gold) {
			return playerID
		}
	}
	return 0
}

func (s *Sim) inLootPickupRangeFrom(pos Vec2, target *entity) bool {
	return meleeInRange(distance(pos, target.pos), s.playerMeleeReach(), s.targetInteractionRadius(target))
}

func isAutoPickableGold(e *entity) bool {
	return e != nil && e.kind == lootEntity && e.itemDefID == goldItemDefID && e.goldAmount > 0
}

func (s *Sim) pickUpGoldForPlayer(e *entity, playerID uint64, correlationID, ackMessageID string, res *TickResult) bool {
	if !isAutoPickableGold(e) {
		return false
	}
	ps := s.players[playerID]
	if ps == nil {
		return false
	}
	s.usePlayer(ps)
	level := s.activeLevel()
	if level == nil || level.entities[e.id] != e {
		return false
	}
	delete(level.entities, e.id)
	res.Changes = append(res.Changes, Change{Op: OpEntityRemove, EntityID: idStr(e.id)})
	amount := e.goldAmount
	s.gold += amount
	s.progression.Gold = s.gold
	res.Changes = append(res.Changes, Change{Op: OpGoldUpdate, OwnerPlayerID: playerID, Gold: intPtr(s.gold)})
	view := s.CharacterProgressionView()
	res.Changes = append(res.Changes, Change{Op: OpCharacterProgressionUpdate, OwnerPlayerID: playerID, Progression: &view})
	res.Events = append(res.Events, Event{
		EventType:     "gold_picked_up",
		EntityID:      idStr(playerID),
		CorrelationID: correlationID,
		Amount:        intPtr(amount),
		TotalGold:     intPtr(s.gold),
	})
	if ackMessageID != "" {
		res.ack(ackMessageID)
	}
	s.savePlayer(ps)
	return true
}

func (s *Sim) activateInteractable(e *entity, in Input, res *TickResult, ack bool) {
	if e.interactableDefID == teleporterDefID {
		s.activateTeleporter(e, in, res, ack)
		return
	}
	if shopID := s.shopIDForInteractable(e); shopID != "" {
		s.openShop(e, shopID, in, res, ack)
		return
	}
	if stashID := s.stashIDForInteractable(e); stashID != "" {
		s.openStash(e, stashID, in, res, ack)
		return
	}
	if e.state != interactableClosed {
		res.reject(in.MessageID, "already_open")
		return
	}
	e.state = interactableOpen
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(e))})
	res.Events = append(res.Events, Event{EventType: "interactable_activated", EntityID: idStr(e.id), CorrelationID: in.CorrelationID})
	if e.interactableDefID == treasureChestDefID && e.lootTable != "" {
		s.spawnLootDrops(s.rules.LootDrops(e.lootTable, s.rng), e.pos, s.targetInteractionRadius(e), in.CorrelationID, res, goldRollContext{levelNum: s.activeLevel().levelNum})
	}
	if ack {
		res.ack(in.MessageID)
	}
}

func (s *Sim) openShop(e *entity, shopID string, in Input, res *TickResult, ack bool) {
	if e.state != interactableReady {
		res.reject(in.MessageID, "not_actionable")
		return
	}
	offers, ok := s.shopCatalogWithChanges(shopID, res)
	if !ok {
		res.reject(in.MessageID, "invalid_target")
		return
	}
	res.Events = append(res.Events, Event{
		EventType:      "shop_opened",
		EntityID:       idStr(e.id),
		CorrelationID:  in.CorrelationID,
		ShopID:         shopID,
		Offers:         offers,
		SellAppraisals: s.shopSellAppraisals(shopID),
	})
	if ack {
		res.ack(in.MessageID)
	}
}

func (s *Sim) openStash(e *entity, stashID string, in Input, res *TickResult, ack bool) {
	if e.state != interactableReady {
		res.reject(in.MessageID, "not_actionable")
		return
	}
	res.Events = append(res.Events, Event{
		EventType:     "stash_opened",
		EntityID:      idStr(e.id),
		CorrelationID: in.CorrelationID,
		StashID:       stashID,
		StashItems:    s.stashItemViews(),
		StashGold:     intPtr(s.stashGold),
		StashCapacity: intPtr(s.stashCapacity),
	})
	if ack {
		res.ack(in.MessageID)
	}
}

func (s *Sim) appendStashGoldChanges(res *TickResult, transferID string) {
	res.Changes = append(res.Changes, Change{Op: OpStashGoldUpdate, StashGold: intPtr(s.stashGold), StashTransferID: transferID})
	res.Changes = append(res.Changes, Change{Op: OpGoldUpdate, Gold: intPtr(s.gold), StashTransferID: transferID})
	view := s.CharacterProgressionView()
	res.Changes = append(res.Changes, Change{Op: OpCharacterProgressionUpdate, Progression: &view, StashTransferID: transferID})
}

func (s *Sim) resolveShopIntentTarget(shopEntityID string) (*entity, string, bool, string) {
	shopEntity, levelNum, ok := s.findEntityAnyLevel(shopEntityID)
	if !ok || shopEntity.kind != interactableEntity {
		return nil, "", false, "invalid_target"
	}
	shopID := s.shopIDForInteractable(shopEntity)
	if shopID == "" {
		return nil, "", false, "invalid_target"
	}
	if levelNum != s.currentLevel || s.currentLevel != townLevel {
		return nil, "", false, "out_of_range"
	}
	if !s.inDispatchRange(shopEntity) {
		return nil, "", false, "out_of_range"
	}
	if shopEntity.state != interactableReady {
		return nil, "", false, "not_actionable"
	}
	return shopEntity, shopID, true, ""
}

func (s *Sim) resolveStashIntentTarget(stashEntityID string) (*entity, string, bool, string) {
	stashEntity, levelNum, ok := s.findEntityAnyLevel(stashEntityID)
	if !ok || stashEntity.kind != interactableEntity {
		return nil, "", false, "invalid_target"
	}
	stashID := s.stashIDForInteractable(stashEntity)
	if stashID == "" {
		return nil, "", false, "invalid_target"
	}
	if levelNum != s.currentLevel || s.currentLevel != townLevel {
		return nil, "", false, "out_of_range"
	}
	if !s.inDispatchRange(stashEntity) {
		return nil, "", false, "out_of_range"
	}
	if stashEntity.state != interactableReady {
		return nil, "", false, "not_actionable"
	}
	return stashEntity, stashID, true, ""
}

func (s *Sim) shopIDForInteractable(e *entity) string {
	if e == nil || e.kind != interactableEntity {
		return ""
	}
	return s.rules.Interactables[e.interactableDefID].ShopID
}

func (s *Sim) stashIDForInteractable(e *entity) string {
	if e == nil || e.kind != interactableEntity {
		return ""
	}
	return s.rules.Interactables[e.interactableDefID].StashID
}

func (s *Sim) activateTeleporter(e *entity, in Input, res *TickResult, ack bool) {
	if !s.multiLevel {
		res.reject(in.MessageID, "not_dungeon_world")
		return
	}
	if e.state == interactableDisabled || e.state == interactableLocked {
		res.reject(in.MessageID, s.rules.DungeonGeneration.BossFloor.LockedExitReason)
		res.Events = append(res.Events, Event{EventType: "teleport_blocked", EntityID: idStr(e.id), CorrelationID: in.CorrelationID, Reason: s.rules.DungeonGeneration.BossFloor.LockedExitReason})
		return
	}
	if ack {
		res.ack(in.MessageID)
	}
	if s.discoveredTeleporters[s.currentLevel] {
		return
	}
	s.discoveredTeleporters[s.currentLevel] = true
	res.Changes = append(res.Changes, Change{Op: OpTeleporterDiscoveryUpdate, Level: s.currentLevel, Discovered: true})
	res.Events = append(res.Events, Event{
		EventType:     "teleporter_discovered",
		EntityID:      idStr(e.id),
		CorrelationID: in.CorrelationID,
		Level:         intPtr(s.currentLevel),
	})
	if s.currentLevel < townLevel {
		s.refreshExistingGeneratedShopStock(res)
	}
}

func (s *Sim) movePlayerToLevel(in Input, res *TickResult, current, dest *LevelState, arrivalPos Vec2) *TickResult {
	player := current.entities[s.playerID]
	if player == nil {
		res.reject(in.MessageID, "player_dead")
		return nil
	}
	fromLevel := s.currentLevel
	destLevel := dest.levelNum
	if fromLevel == townLevel && destLevel != townLevel {
		s.clearShopBuyback()
	}
	delete(current.entities, player.id)
	player.pos = arrivalPos
	dest.entities[player.id] = player
	s.currentLevel = destLevel
	current.move = nil
	current.autoNav = nil
	dest.move = nil
	dest.autoNav = nil

	res.ack(in.MessageID)
	res.Changes = append(res.Changes, Change{Op: OpEntityRemove, EntityID: idStr(player.id)})
	res.Events = append(res.Events, Event{
		EventType:     "level_changed",
		CorrelationID: in.CorrelationID,
		FromLevel:     intPtr(fromLevel),
		ToLevel:       intPtr(destLevel),
	})

	arrivalRes := TickResult{Tick: res.Tick, Level: destLevel, Changes: []Change{}, Events: []Event{}}
	arrivalRes.Changes = append(arrivalRes.Changes, Change{Op: OpWallLayoutUpdate, Walls: wallViewsForLevel(dest)})
	if s.multiLevel && s.levelHasTeleporter(destLevel) && !s.discoveredTeleporters[destLevel] {
		arrivalRes.Changes = append(arrivalRes.Changes, Change{
			Op:         OpTeleporterDiscoveryUpdate,
			Level:      destLevel,
			Discovered: false,
		})
	}
	s.appendDeepestDungeonDepthChange(destLevel, &arrivalRes)
	for _, id := range sortedEntityIDs(dest.entities) {
		arrivalRes.Changes = append(arrivalRes.Changes, Change{Op: OpEntitySpawn, Entity: ptrEntityView(s.entityView(dest.entities[id]))})
	}
	return &arrivalRes
}

func (s *Sim) appendDeepestDungeonDepthChange(destLevel int, res *TickResult) {
	if destLevel >= townLevel {
		return
	}
	depth := absInt(destLevel)
	if depth <= s.progression.DeepestDungeonDepth {
		return
	}
	s.progression.DeepestDungeonDepth = depth
	view := s.CharacterProgressionView()
	res.Changes = append(res.Changes, Change{Op: OpCharacterProgressionUpdate, Progression: &view})
}

func (s *Sim) travelArrivalPosition(level *LevelState, marker Vec2, movingPlayerID uint64) Vec2 {
	nav := s.navigationForLevel(level)
	step := nav.CellSize
	if step <= 0 {
		step = 1.0
	}
	unitOffsets := []Vec2{
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: -1, Y: 0},
		{X: 0, Y: -1},
		{X: 1, Y: 1},
		{X: -1, Y: 1},
		{X: -1, Y: -1},
		{X: 1, Y: -1},
	}
	for ring := 1; ring <= 8; ring++ {
		scale := float64(ring) * step
		for _, offset := range unitOffsets {
			candidate := Vec2{X: marker.X + offset.X*scale, Y: marker.Y + offset.Y*scale}
			if !s.positionInNavigationBounds(nav, candidate) {
				continue
			}
			if s.travelArrivalBlocked(level, candidate, movingPlayerID) {
				continue
			}
			return candidate
		}
	}
	return marker
}

func (s *Sim) travelArrivalBlocked(level *LevelState, pos Vec2, movingPlayerID uint64) bool {
	if level == nil {
		return true
	}
	for _, wall := range level.walls {
		if circleIntersectsAABB(pos, playerRadius, wall.pos, wall.size) {
			return true
		}
	}
	for _, id := range sortedEntityIDs(level.entities) {
		if id == movingPlayerID {
			continue
		}
		e := level.entities[id]
		if e == nil {
			continue
		}
		switch e.kind {
		case playerEntity:
			if e.hp > 0 && circlesOverlap(pos, playerRadius, e.pos, playerRadius) {
				return true
			}
		case monsterEntity:
			if e.hp > 0 && circlesOverlap(pos, playerRadius, e.pos, monsterRadius) {
				return true
			}
		case interactableEntity:
			if circlesOverlap(pos, playerRadius, e.pos, interactableInteractionRadius) {
				return true
			}
			if e.state == interactableClosed {
				if def, ok := s.rules.Interactables[e.interactableDefID]; ok && def.BarrierWhenClosed != nil {
					if circleIntersectsAABB(pos, playerRadius, e.pos, def.BarrierWhenClosed.Size) {
						return true
					}
				}
			}
		}
	}
	return false
}

func (s *Sim) appendEquipmentProgressionChanges(res *TickResult) {
	player := s.activeLevel().entities[s.playerID]
	if player != nil {
		maxHP := s.currentMaxHP()
		if maxHP != player.maxHP {
			delta := maxHP - player.maxHP
			player.maxHP = maxHP
			if delta > 0 {
				player.hp += delta
			}
			if player.hp > player.maxHP {
				player.hp = player.maxHP
			}
			res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
		}
	}
	view := s.CharacterProgressionView()
	res.Changes = append(res.Changes, Change{Op: OpCharacterProgressionUpdate, Progression: &view})
	s.appendInventoryPresentationUpdates(res)
}

func (s *Sim) appendInventoryPresentationUpdates(res *TickResult) {
	s.appendInventoryPresentationUpdatesForOwner(res, 0)
}

func (s *Sim) appendInventoryPresentationUpdatesForOwner(res *TickResult, ownerPlayerID uint64) {
	for _, item := range s.inventory {
		if item == nil {
			continue
		}
		view := s.itemView(item)
		if len(view.RequirementStatus) == 0 && view.EquipPreview == nil {
			continue
		}
		res.Changes = append(res.Changes, Change{Op: OpInventoryUpdate, OwnerPlayerID: ownerPlayerID, Item: ptrItemView(view)})
	}
}

func (s *Sim) consumeItem(item *invItem, correlationID string, res *TickResult) (bool, string) {
	player := s.activeLevel().entities[s.playerID]
	if player == nil || player.hp <= 0 {
		return false, "player_dead"
	}
	def := s.rules.Items[item.itemDefID]
	if def.Heal == nil && def.ManaRestore == nil {
		return false, "not_usable"
	}
	heal := 0
	mana := 0
	if def.Heal != nil {
		if player.hp >= player.maxHP {
			return false, "already_full_hp"
		}
		heal = s.rollRange(*def.Heal)
		if player.hp+heal > player.maxHP {
			heal = player.maxHP - player.hp
		}
		if heal <= 0 {
			return false, "already_full_hp"
		}
	}
	if def.ManaRestore != nil {
		if player.mana >= player.maxMana {
			return false, "already_full_mana"
		}
		mana = s.rollRange(*def.ManaRestore)
		if player.mana+mana > player.maxMana {
			mana = player.maxMana - player.mana
		}
		if mana <= 0 {
			return false, "already_full_mana"
		}
	}

	removedID := idStr(item.instanceID)
	s.removeItemByID(item.instanceID)
	res.Changes = append(res.Changes, Change{Op: OpInventoryRemove, ItemInstanceID: &removedID})
	s.clearHotbarReferences(item.instanceID, res)

	player.hp += heal
	player.mana += mana
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	used := Event{
		EventType:      "item_used",
		EntityID:       idStr(player.id),
		CorrelationID:  correlationID,
		ItemInstanceID: removedID,
	}
	if heal > 0 {
		used.Heal = intPtr(heal)
	}
	if mana > 0 {
		used.Mana = intPtr(mana)
	}
	res.Events = append(res.Events, used)
	if heal > 0 {
		res.Events = append(res.Events, Event{
			EventType:      "player_healed",
			EntityID:       idStr(player.id),
			CorrelationID:  correlationID,
			Heal:           intPtr(heal),
			ItemInstanceID: removedID,
		})
	}
	if mana > 0 {
		res.Events = append(res.Events, Event{
			EventType:      "player_mana_restored",
			EntityID:       idStr(player.id),
			CorrelationID:  correlationID,
			Mana:           intPtr(mana),
			ItemInstanceID: removedID,
		})
	}
	return true, ""
}

func isBaseStat(stat string) bool {
	switch stat {
	case "str", "dex", "vit", "magic":
		return true
	}
	return false
}

func (s *Sim) appendProgressionAndSkillUpdates(res *TickResult) {
	view := s.CharacterProgressionView()
	res.Changes = append(res.Changes, Change{Op: OpCharacterProgressionUpdate, Progression: &view})
	skillView := s.SkillProgressionView()
	res.Changes = append(res.Changes, Change{Op: OpSkillProgressionUpdate, SkillProgression: &skillView})
	s.appendInventoryPresentationUpdates(res)
}

func (s *Sim) appendSkillCooldownUpdate(res *TickResult) {
	res.Changes = append(res.Changes, Change{Op: OpSkillCooldownUpdate, SkillCooldowns: s.SkillCooldownViews()})
}

func (s *Sim) skillCooldownRemaining(skillID string) (int, bool) {
	cooldown, ok := s.skillCooldowns[skillID]
	if !ok || cooldown.TotalTicks <= 0 {
		return 0, false
	}
	if cooldown.EndsTick <= s.tick {
		delete(s.skillCooldowns, skillID)
		return 0, false
	}
	return int(cooldown.EndsTick - s.tick), true
}

func (s *Sim) skillCooldownView(skillID string) (SkillCooldownView, bool) {
	cooldown, ok := s.skillCooldowns[skillID]
	if !ok || cooldown.TotalTicks <= 0 {
		return SkillCooldownView{}, false
	}
	remaining, active := s.skillCooldownRemaining(skillID)
	if !active {
		return SkillCooldownView{}, false
	}
	return SkillCooldownView{SkillID: skillID, RemainingTicks: remaining, TotalTicks: cooldown.TotalTicks}, true
}

func (s *Sim) skillCooldownTicks(def SkillDef) int {
	interval := s.DerivedStatsView().AttackIntervalTicks
	if interval < 1 {
		interval = s.rules.Combat.BaseAttackIntervalTicks
	}
	cooldown := int(math.Ceil(float64(interval) * def.Cooldown.Multiplier))
	if cooldown < 1 {
		return 1
	}
	return cooldown
}

func skillManaCost(def SkillDef, rank int) int {
	if rank < 1 {
		rank = 1
	}
	cost := def.Cost.Mana.Base + def.Cost.Mana.PerRank*(rank-1)
	if cost < 0 {
		return 0
	}
	return cost
}

func skillDamageRange(def SkillDef, rank int) DamageRange {
	if rank < 1 {
		rank = 1
	}
	minDamage := def.Damage.MinBase + def.Damage.MinPerRank*(rank-1)
	maxDamage := def.Damage.MaxBase + def.Damage.MaxPerRank*(rank-1)
	if minDamage < 0 {
		minDamage = 0
	}
	if maxDamage < minDamage {
		maxDamage = minDamage
	}
	return DamageRange{Min: minDamage, Max: maxDamage}
}

func (s *Sim) skillCastDirection(def SkillDef, cast *CastSkillIntent, player *entity) (Vec2, uint64, string) {
	if cast == nil || player == nil {
		return Vec2{}, 0, "invalid_payload"
	}
	if cast.TargetID != "" {
		target := s.findEntity(cast.TargetID)
		if target == nil || target.kind != monsterEntity || target.hp <= 0 {
			return Vec2{}, 0, "invalid_target"
		}
		if distance(player.pos, target.pos) > def.Projectile.Range+meleeRangeEpsilon {
			return Vec2{}, 0, "target_out_of_range"
		}
		dir := normalize(Vec2{X: target.pos.X - player.pos.X, Y: target.pos.Y - player.pos.Y})
		if dir.X == 0 && dir.Y == 0 {
			if cast.Direction == nil {
				return Vec2{}, 0, "invalid_direction"
			}
			dir = normalize(*cast.Direction)
		}
		if dir.X == 0 && dir.Y == 0 {
			return Vec2{}, 0, "invalid_direction"
		}
		return dir, target.id, ""
	}
	if cast.Direction == nil || !finiteVec2(*cast.Direction) {
		return Vec2{}, 0, "invalid_payload"
	}
	dir := normalize(*cast.Direction)
	if dir.X == 0 && dir.Y == 0 {
		return Vec2{}, 0, "invalid_direction"
	}
	return dir, 0, ""
}

func (s *Sim) spawnSkillProjectile(player *entity, skillID string, def SkillDef, rank int, dir Vec2, targetID uint64, in Input) *entity {
	projectile := &entity{
		kind:            projectileEntity,
		pos:             player.pos,
		ownerID:         player.id,
		targetID:        targetID,
		projectileDefID: skillID,
		dir:             normalize(dir),
		speed:           def.Projectile.Speed,
		maxDistance:     def.Projectile.Range,
		damageRange:     skillDamageRange(def, rank),
		sourceMsgID:     in.MessageID,
		sourceCorrID:    in.CorrelationID,
		spawnTick:       s.tick,
	}
	projectile.id = s.alloc()
	s.activeLevel().entities[projectile.id] = projectile
	return projectile
}

func skillEffectPercent(effect SkillEffectDef, rank int) int {
	if rank < 1 {
		rank = 1
	}
	percent := effect.PercentBase + effect.PercentPerRank*(rank-1)
	if percent < 0 {
		return 0
	}
	return percent
}

func (s *Sim) applySkillBuff(player *entity, skillID string, def SkillDef, rank int, correlationID string, res *TickResult) {
	if player == nil {
		return
	}
	for _, effect := range def.Effects {
		if effect.Type != "stat_percent_buff" {
			continue
		}
		percent := skillEffectPercent(effect, rank)
		scale := 1.0
		if effect.VisualScale {
			scale += float64(percent) / 100.0
		}
		totalTicks := effect.DurationTicks
		s.skillEffects[skillID] = skillEffectState{
			SkillID:     skillID,
			Stats:       cloneStringSlice(effect.Stats),
			Percent:     percent,
			VisualScale: scale,
			EndsTick:    s.tick + uint64(totalTicks),
			TotalTicks:  totalTicks,
		}
		s.syncActivePlayerVisualScale()
		res.Events = append(res.Events, Event{
			EventType:      "skill_effect_started",
			EntityID:       idStr(player.id),
			CorrelationID:  correlationID,
			SkillID:        skillID,
			Rank:           intPtr(rank),
			Amount:         intPtr(percent),
			RemainingTicks: intPtr(totalTicks),
			TotalTicks:     intPtr(totalTicks),
		})
	}
}

func (s *Sim) skillAreaCenter(effect SkillEffectDef, cast *CastSkillIntent, player *entity) (Vec2, string) {
	if cast == nil || player == nil {
		return Vec2{}, "invalid_payload"
	}
	if cast.TargetID != "" {
		target := s.findEntity(cast.TargetID)
		if target == nil || (target.kind != monsterEntity && target.kind != playerEntity) || target.hp <= 0 {
			return Vec2{}, "invalid_target"
		}
		if distance(player.pos, target.pos) > effect.Range+meleeRangeEpsilon {
			return Vec2{}, "target_out_of_range"
		}
		return target.pos, ""
	}
	if cast.Direction == nil {
		return player.pos, ""
	}
	if !finiteVec2(*cast.Direction) {
		return Vec2{}, "invalid_payload"
	}
	dir := normalize(*cast.Direction)
	if dir.X == 0 && dir.Y == 0 {
		return Vec2{}, "invalid_direction"
	}
	return Vec2{X: player.pos.X + dir.X*effect.Range, Y: player.pos.Y + dir.Y*effect.Range}, ""
}

func (s *Sim) healSkillTargets(center Vec2, effect SkillEffectDef, casterID uint64) []*entity {
	targets := []*entity{}
	level := s.activeLevel()
	for _, id := range sortedEntityIDs(level.entities) {
		entity := level.entities[id]
		if entity == nil || entity.kind != playerEntity || entity.hp <= 0 {
			continue
		}
		if entity.id == casterID && !effect.IncludeCaster {
			continue
		}
		if distance(center, entity.pos) > effect.Radius+meleeRangeEpsilon {
			continue
		}
		targets = append(targets, entity)
	}
	return targets
}

func (s *Sim) areaHealApplications(player *entity, def SkillDef, rank int, cast *CastSkillIntent) ([]skillHealApplication, string) {
	if player == nil {
		return nil, "player_dead"
	}
	applications := []skillHealApplication{}
	for _, effect := range def.Effects {
		if effect.Type != "area_percent_heal" {
			continue
		}
		center, rejectReason := s.skillAreaCenter(effect, cast, player)
		if rejectReason != "" {
			return nil, rejectReason
		}
		percent := skillEffectPercent(effect, rank)
		targets := s.healSkillTargets(center, effect, player.id)
		for _, target := range targets {
			if target.hp >= target.maxHP {
				continue
			}
			heal := int(math.Floor(float64(target.maxHP)*float64(percent)/100.0 + 0.000000001))
			if heal < 1 {
				heal = 1
			}
			if target.hp+heal > target.maxHP {
				heal = target.maxHP - target.hp
			}
			if heal <= 0 {
				continue
			}
			applications = append(applications, skillHealApplication{Target: target, Heal: heal})
		}
	}
	if len(applications) == 0 {
		return nil, "already_full_hp"
	}
	return applications, ""
}

func (s *Sim) syncActivePlayerVisualScale() bool {
	player := s.activeLevel().entities[s.playerID]
	if player == nil {
		return false
	}
	scale := 1.0
	for _, skillID := range sortedStringKeys(s.skillEffects) {
		effect := s.skillEffects[skillID]
		if effect.EndsTick <= s.tick || effect.VisualScale <= scale {
			continue
		}
		scale = effect.VisualScale
	}
	stored := 0.0
	if scale > 1.0 {
		stored = scale
	}
	if math.Abs(player.visualScale-stored) <= 0.000001 {
		return false
	}
	player.visualScale = stored
	return true
}

func (s *Sim) applyAreaHeal(player *entity, skillID string, rank int, applications []skillHealApplication, correlationID string, res *TickResult) {
	for _, app := range applications {
		target := app.Target
		if target == nil || app.Heal <= 0 {
			continue
		}
		target.hp += app.Heal
		if target.hp > target.maxHP {
			target.hp = target.maxHP
		}
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
		res.Events = append(res.Events, Event{
			EventType:      "player_healed",
			EntityID:       idStr(target.id),
			SourceEntityID: idStr(player.id),
			TargetEntityID: idStr(target.id),
			CorrelationID:  correlationID,
			SkillID:        skillID,
			Rank:           intPtr(rank),
			Heal:           intPtr(app.Heal),
		})
	}
}

func (s *Sim) applyMovement(res *TickResult) {
	if s.activeLevel().autoNav != nil && s.activeLevel().move == nil {
		s.applyAutoNav(res)
		return
	}
	if s.activeLevel().move == nil || s.activeLevel().move.remaining <= 0 {
		return
	}
	if s.playerDead() {
		s.activeLevel().move = nil
		return
	}
	player := s.activeLevel().entities[s.playerID]
	before := player.pos
	player.pos = s.resolveMovement(player.pos, Vec2{
		X: s.activeLevel().move.dir.X * moveSpeed,
		Y: s.activeLevel().move.dir.Y * moveSpeed,
	})
	s.activeLevel().move.remaining--
	if s.activeLevel().move.remaining == 0 {
		s.activeLevel().move = nil
	}
	if player.pos == before {
		return
	}
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
}

func (s *Sim) applyAutoNav(res *TickResult) {
	if s.playerDead() {
		s.clearAutoNav()
		return
	}
	if len(s.activeLevel().autoNav.steps) == 0 {
		s.finishAutoNav(res)
		return
	}
	player := s.activeLevel().entities[s.playerID]
	before := player.pos
	step := s.activeLevel().autoNav.steps[0]
	s.activeLevel().autoNav.steps = s.activeLevel().autoNav.steps[1:]
	player.pos = s.resolveMovement(player.pos, Vec2{X: step.X * moveSpeed, Y: step.Y * moveSpeed})
	if player.pos != before {
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	}
	if len(s.activeLevel().autoNav.steps) == 0 {
		s.finishAutoNav(res)
	}
}

func (s *Sim) finishAutoNav(res *TickResult) {
	nav := s.activeLevel().autoNav
	s.clearAutoNav()
	if nav == nil || nav.pendingAction == nil {
		return
	}
	in := Input{
		MessageID:     nav.sourceMsgID,
		CorrelationID: nav.sourceCorrID,
		Type:          "action_intent",
		Action:        nav.pendingAction,
	}
	target := s.findEntity(nav.pendingAction.TargetID)
	if target == nil || !s.actionable(target) || !s.inDispatchRange(target) {
		return
	}
	s.dispatchAction(target, in, res, false)
}

func (s *Sim) clearAutoNav() {
	s.activeLevel().autoNav = nil
}

func (s *Sim) resolveMovement(pos, delta Vec2) Vec2 {
	candidate := Vec2{X: pos.X + delta.X, Y: pos.Y + delta.Y}
	if !s.playerPositionBlocked(candidate) {
		return candidate
	}
	xOnly := Vec2{X: pos.X + delta.X, Y: pos.Y}
	if delta.X != 0 && !s.playerPositionBlocked(xOnly) {
		return xOnly
	}
	yOnly := Vec2{X: pos.X, Y: pos.Y + delta.Y}
	if delta.Y != 0 && !s.playerPositionBlocked(yOnly) {
		return yOnly
	}
	return pos
}

func (s *Sim) playerPositionBlocked(pos Vec2) bool {
	for _, wall := range s.activeWalls() {
		if circleIntersectsAABB(pos, playerRadius, wall.pos, wall.size) {
			return true
		}
	}
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		e := s.activeLevel().entities[id]
		if e.kind == monsterEntity && e.hp > 0 {
			if circlesOverlap(pos, playerRadius, e.pos, monsterRadius) {
				return true
			}
			continue
		}
		if e.kind == interactableEntity && e.state == interactableClosed {
			if def, ok := s.rules.Interactables[e.interactableDefID]; ok && def.BarrierWhenClosed != nil {
				if circleIntersectsAABB(pos, playerRadius, e.pos, def.BarrierWhenClosed.Size) {
					return true
				}
			}
		}
	}
	return false
}

func (s *Sim) findDropPosition() (Vec2, bool) {
	player := s.activeLevel().entities[s.playerID]
	if player == nil {
		return Vec2{}, false
	}
	return s.findAdjacentLootDropPosition(player.pos, playerRadius)
}

func (s *Sim) findEntityLootDropPosition(source Vec2, sourceRadius float64) (Vec2, bool) {
	return s.findAdjacentLootDropPosition(source, sourceRadius)
}

func (s *Sim) findAdjacentLootDropPosition(source Vec2, sourceRadius float64) (Vec2, bool) {
	step := s.activeNav().CellSize
	if step <= 0 {
		step = 1.0
	}
	unitOffsets := []Vec2{
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: -1, Y: 0},
		{X: 0, Y: -1},
		{X: 1, Y: 1},
		{X: -1, Y: 1},
		{X: -1, Y: -1},
		{X: 1, Y: -1},
	}
	start := s.rng.IntN(len(unitOffsets))
	for ring := 1; ring <= 6; ring++ {
		scale := float64(ring) * step
		for i := 0; i < len(unitOffsets); i++ {
			offset := unitOffsets[(start+i)%len(unitOffsets)]
			pos := Vec2{X: source.X + offset.X*scale, Y: source.Y + offset.Y*scale}
			if distance(pos, source) < sourceRadius+lootInteractionRadius {
				continue
			}
			if s.lootDropBlocked(pos) {
				continue
			}
			if s.lootPositionBlocked(pos) {
				continue
			}
			return pos, true
		}
	}
	return Vec2{}, false
}

func (s *Sim) lootDropBlocked(pos Vec2) bool {
	for _, wall := range s.activeWalls() {
		if circleIntersectsAABB(pos, lootInteractionRadius, wall.pos, wall.size) {
			return true
		}
	}
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		e := s.activeLevel().entities[id]
		if e.kind == interactableEntity && e.state == interactableClosed {
			if def, ok := s.rules.Interactables[e.interactableDefID]; ok && def.BarrierWhenClosed != nil {
				if circleIntersectsAABB(pos, lootInteractionRadius, e.pos, def.BarrierWhenClosed.Size) {
					return true
				}
			}
		}
	}
	return false
}

func (s *Sim) findClusterLootDropPosition(anchor Vec2, index int) (Vec2, bool) {
	spacing := lootInteractionRadius * 2.1
	offsets := []Vec2{
		{X: spacing, Y: 0},
		{X: -spacing, Y: 0},
		{X: 0, Y: spacing},
		{X: 0, Y: -spacing},
		{X: spacing, Y: spacing},
		{X: -spacing, Y: spacing},
		{X: -spacing, Y: -spacing},
		{X: spacing, Y: -spacing},
	}

	for try := 0; try < len(offsets); try++ {
		offset := offsets[(index-1+try)%len(offsets)]
		pos := Vec2{X: anchor.X + offset.X, Y: anchor.Y + offset.Y}
		if s.lootDropBlocked(pos) {
			continue
		}
		if s.lootPositionBlocked(pos) {
			continue
		}

		return pos, true
	}

	return Vec2{}, false
}

func (s *Sim) lootPositionBlocked(pos Vec2) bool {
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		e := s.activeLevel().entities[id]
		if e.kind != lootEntity {
			continue
		}
		if circlesOverlap(pos, lootInteractionRadius, e.pos, lootInteractionRadius) {
			return true
		}
	}
	return false
}

func (s *Sim) removeItemByID(id uint64) {
	for i, it := range s.inventory {
		if it.instanceID == id {
			s.inventory = append(s.inventory[:i], s.inventory[i+1:]...)
			return
		}
	}
}

func (s *Sim) removeStashItemByID(id uint64) {
	for i, it := range s.stashItems {
		if it == nil {
			continue
		}
		if it.stashItemID == id {
			s.stashItems = append(s.stashItems[:i], s.stashItems[i+1:]...)
			return
		}
	}
}

func (s *Sim) buildBlockedFn() func(gx, gy int) bool {
	return func(gx, gy int) bool {
		center := gridToWorld(s.activeNav(), gridCell{x: gx, y: gy})
		return s.playerPositionBlocked(center)
	}
}

func (s *Sim) findApproachGoal(target *entity) (Vec2, []Vec2, bool) {
	if target.kind == monsterEntity && s.playerAttackMode() == attackModeRanged {
		return s.findRangedApproachGoal(target)
	}
	return s.findMeleeApproachGoal(target)
}

func (s *Sim) findRangedApproachGoal(target *entity) (Vec2, []Vec2, bool) {
	player := s.activeLevel().entities[s.playerID]
	if player == nil {
		return Vec2{}, nil, false
	}
	nav := s.activeNav()
	playerCell := worldToGrid(nav, player.pos)
	blocked := s.buildBlockedFn()
	maxRadius := maxInt(nav.GridBounds.MaxX-nav.GridBounds.MinX, nav.GridBounds.MaxY-nav.GridBounds.MinY) + 1
	for radius := 0; radius <= maxRadius; radius++ {
		candidates := ringCells(playerCell, radius)
		for _, cell := range candidates {
			if !cellInBounds(nav, cell) || blocked(cell.x, cell.y) {
				continue
			}
			goal := gridToWorld(nav, cell)
			if !s.inActionRangeFrom(goal, target) || !s.hasClearRangedShot(goal, target) {
				continue
			}
			steps, ok := PlanPath(nav, player.pos, goal, blocked)
			if ok {
				return goal, steps, true
			}
		}
	}
	return Vec2{}, nil, false
}

func (s *Sim) findMeleeApproachGoal(target *entity) (Vec2, []Vec2, bool) {
	return s.findApproachGoalMatching(target, func(pos Vec2, target *entity) bool {
		return meleeInRange(distance(pos, target.pos), s.playerMeleeReach(), s.targetInteractionRadius(target))
	})
}

func (s *Sim) findApproachGoalMatching(target *entity, inRange func(Vec2, *entity) bool) (Vec2, []Vec2, bool) {
	player := s.activeLevel().entities[s.playerID]
	if player == nil {
		return Vec2{}, nil, false
	}
	nav := s.activeNav()
	targetCell := worldToGrid(nav, target.pos)
	blocked := s.buildBlockedFn()
	maxRadius := maxInt(nav.GridBounds.MaxX-nav.GridBounds.MinX, nav.GridBounds.MaxY-nav.GridBounds.MinY) + 1
	for radius := 0; radius <= maxRadius; radius++ {
		candidates := ringCells(targetCell, radius)
		for _, cell := range candidates {
			if !cellInBounds(nav, cell) || blocked(cell.x, cell.y) {
				continue
			}
			goal := gridToWorld(nav, cell)
			if !inRange(goal, target) {
				continue
			}
			steps, ok := PlanPath(nav, player.pos, goal, blocked)
			if ok {
				return goal, steps, true
			}
		}
	}
	return Vec2{}, nil, false
}

func ringCells(center gridCell, radius int) []gridCell {
	if radius == 0 {
		return []gridCell{center}
	}
	out := []gridCell{}
	for y := center.y - radius; y <= center.y+radius; y++ {
		for x := center.x - radius; x <= center.x+radius; x++ {
			if absInt(x-center.x) != radius && absInt(y-center.y) != radius {
				continue
			}
			out = append(out, gridCell{x: x, y: y})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].y != out[j].y {
			return out[i].y < out[j].y
		}
		return out[i].x < out[j].x
	})
	return out
}

func circlesOverlap(a Vec2, ar float64, b Vec2, br float64) bool {
	dx := a.X - b.X
	dy := a.Y - b.Y
	r := ar + br
	return dx*dx+dy*dy < r*r-1e-9
}

func finiteVec2(v Vec2) bool {
	return !math.IsNaN(v.X) && !math.IsInf(v.X, 0) && !math.IsNaN(v.Y) && !math.IsInf(v.Y, 0)
}

func circleIntersectsAABB(center Vec2, radius float64, rectCenter Vec2, rectSize Vec2) bool {
	halfX := rectSize.X / 2
	halfY := rectSize.Y / 2
	closestX := math.Max(rectCenter.X-halfX, math.Min(center.X, rectCenter.X+halfX))
	closestY := math.Max(rectCenter.Y-halfY, math.Min(center.Y, rectCenter.Y+halfY))
	dx := center.X - closestX
	dy := center.Y - closestY
	return dx*dx+dy*dy < radius*radius-1e-9
}

func (s *Sim) advanceMonsterMovement(res *TickResult) {
	nav := s.activeNav()
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		monster := s.activeLevel().entities[id]
		if monster == nil || monster.kind != monsterEntity || monster.hp <= 0 {
			continue
		}
		def, ok := s.rules.Monsters[monster.monsterDefID]
		if !ok || def.effectiveBehavior() != monsterBehaviorChase {
			continue
		}
		if monster.isBoss && monster.bossPhaseKind == "active" {
			continue
		}
		targetPlayer := s.nearestLivingPlayerForMonster(s.activeLevel(), monster)
		if targetPlayer == nil {
			continue
		}
		player := s.activeLevel().entities[targetPlayer.PlayerID]
		if player == nil {
			continue
		}
		s.usePlayer(targetPlayer)
		prevMode := monster.aiMode
		if monster.isBoss {
			monster.aiMode = monsterAIModeChase
		} else {
			s.updateMonsterAIMode(monster, player, def, prevMode, res)
		}
		if monster.aiMode == monsterAIModeIdle {
			continue
		}
		goal, hasGoal := s.monsterMovementGoal(monster, player, def)
		if !hasGoal {
			continue
		}
		if distance(monster.pos, goal) <= nav.StopDistance && s.monsterInAttackRange(monster, player, def) {
			continue
		}
		blocked := s.buildMonsterBlockedFn(monster.id)
		steps, ok := PlanPath(nav, monster.pos, goal, blocked)
		if !ok || len(steps) == 0 {
			if distance(monster.pos, goal) > nav.CellSize+nav.StopDistance {
				continue
			}
		}
		moveSpeed := def.effectiveMoveSpeed(nav)
		before := monster.pos
		monster.pos = s.resolveMonsterMovement(monster, s.monsterMoveDelta(monster.pos, goal, steps, moveSpeed))
		if monster.pos != before {
			res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(monster))})
		}
	}
}

func (s *Sim) advanceMonsterAttack(res *TickResult) {
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		monster := s.activeLevel().entities[id]
		if monster == nil || monster.kind != monsterEntity || monster.hp <= 0 {
			continue
		}
		if monster.isBoss {
			continue
		}
		def, ok := s.rules.Monsters[monster.monsterDefID]
		if !ok || def.AttackDamage == nil || def.AttackCooldown <= 0 {
			continue
		}
		if monster.aiMode == monsterAIModeReturn {
			continue
		}
		targetPlayer := s.nearestLivingPlayerForMonster(s.activeLevel(), monster)
		if targetPlayer == nil {
			continue
		}
		player := s.activeLevel().entities[targetPlayer.PlayerID]
		if player == nil || player.hp <= 0 {
			continue
		}
		s.usePlayer(targetPlayer)
		if !s.monsterInAttackRange(monster, player, def) {
			continue
		}
		if monster.hasAttacked && s.tick-monster.lastAttackTick < uint64(def.AttackCooldown) {
			continue
		}
		attackDamage := def.AttackDamage
		if monster.monsterAttackDamage != nil {
			attackDamage = monster.monsterAttackDamage
		}
		scaledAttackDamage := s.scaleMonsterDamageForParty(s.currentLevel, *attackDamage)
		attackDamage = &scaledAttackDamage
		monster.lastAttackTick = s.tick
		monster.hasAttacked = true
		if def.effectiveAttackMode() == attackModeRanged {
			s.fireMonsterProjectile(monster, player, def, *attackDamage, res)
			continue
		}
		s.damagePlayerByMonster(monster, player, *attackDamage, "", res)
	}
}

func (s *Sim) fireMonsterProjectile(monster *entity, player *entity, def MonsterDef, damageRange DamageRange, res *TickResult) {
	dir := normalize(Vec2{X: player.pos.X - monster.pos.X, Y: player.pos.Y - monster.pos.Y})
	if dir.X == 0 && dir.Y == 0 {
		dir = Vec2{X: 1}
	}
	projectile := &entity{
		kind:            projectileEntity,
		pos:             monster.pos,
		ownerID:         monster.id,
		targetID:        player.id,
		projectileDefID: def.ProjectileDefID,
		dir:             dir,
		speed:           def.ProjectileSpeed,
		maxDistance:     def.AttackRange,
		damageRange:     damageRange,
		spawnTick:       s.tick,
	}
	projectile.id = s.alloc()
	s.activeLevel().entities[projectile.id] = projectile
	res.Changes = append(res.Changes, Change{Op: OpEntitySpawn, Entity: ptrEntityView(s.entityView(projectile))})
}

func (s *Sim) damagePlayerByMonster(monster *entity, player *entity, damageRange DamageRange, corr string, res *TickResult) combatResolution {
	attackerStats := s.monsterEffectiveCombatStats(monster, damageRange)
	defenderStats, _ := s.playerEffectiveCombatStats()
	outcome := s.resolveCombat(attackerStats, defenderStats, damageRange)
	if !outcome.Hit || outcome.Blocked {
		res.Events = append(res.Events, combatEvent(s.combatEventType(playerEntity, outcome), monster.id, player.id, corr, outcome))
		return outcome
	}
	player.hp -= outcome.Damage
	if player.hp < 0 {
		player.hp = 0
	}
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	eventType := "player_damaged"
	if player.hp == 0 {
		eventType = "player_killed"
	}
	res.Events = append(res.Events, combatEvent(eventType, monster.id, player.id, corr, outcome))
	return outcome
}

func (s *Sim) nearestLivingPlayerForMonster(level *LevelState, monster *entity) *playerState {
	if monster != nil && monster.aiTargetPlayerID != 0 {
		ps := s.players[monster.aiTargetPlayerID]
		e := level.entities[monster.aiTargetPlayerID]
		if ps != nil && ps.Connected && ps.CurrentLevel == level.levelNum && e != nil && e.kind == playerEntity && e.hp > 0 {
			return ps
		}
		monster.aiTargetPlayerID = 0
	}
	var best *playerState
	bestDist := math.MaxFloat64
	for _, playerID := range sortedPlayerIDs(s.players) {
		ps := s.players[playerID]
		if ps == nil || !ps.Connected || ps.CurrentLevel != level.levelNum {
			continue
		}
		e := level.entities[playerID]
		if e == nil || e.kind != playerEntity || e.hp <= 0 {
			continue
		}
		dist := distance(monster.pos, e.pos)
		if best == nil || dist < bestDist-1e-9 || (math.Abs(dist-bestDist) <= 1e-9 && playerID < best.PlayerID) {
			best = ps
			bestDist = dist
		}
	}
	return best
}

func (s *Sim) updateMonsterAIMode(monster *entity, player *entity, def MonsterDef, prevMode string, res *TickResult) {
	nav := s.activeNav()
	distPlayer := distance(monster.pos, player.pos)
	distPlayerFromSpawn := distance(player.pos, monster.spawnPos)
	leashRadius := effectiveMonsterLeashRadius(def, nav)

	if leashRadius > 0 && distPlayerFromSpawn > leashRadius {
		if prevMode != monsterAIModeReturn {
			res.Events = append(res.Events, Event{EventType: "monster_leashed", EntityID: idStr(monster.id)})
		}
		monster.aiMode = monsterAIModeReturn

		return
	}

	if distPlayer <= def.AggroRadius {
		if prevMode != monsterAIModeChase {
			res.Events = append(res.Events, Event{EventType: "monster_aggro", EntityID: idStr(monster.id)})
		}
		if monster.aiTargetPlayerID == 0 {
			monster.aiTargetPlayerID = player.id
		}
		monster.aiMode = monsterAIModeChase

		return
	}

	if monster.aiTargetPlayerID != 0 && prevMode == monsterAIModeChase {
		monster.aiMode = monsterAIModeChase
		return
	}

	if distance(monster.pos, monster.spawnPos) <= nav.StopDistance {
		monster.aiMode = monsterAIModeIdle

		return
	}

	if prevMode == monsterAIModeReturn {
		monster.aiMode = monsterAIModeReturn

		return
	}

	monster.aiMode = monsterAIModeIdle
}

func effectiveMonsterLeashRadius(def MonsterDef, nav NavigationRules) float64 {
	minimumRadius := minimumChaseLeashTiles * nav.CellSize
	if def.LeashRadius <= 0 {
		return minimumRadius
	}

	return maxFloat(def.LeashRadius, minimumRadius)
}

func (s *Sim) monsterMovementGoal(monster *entity, player *entity, def MonsterDef) (Vec2, bool) {
	nav := s.activeNav()
	switch monster.aiMode {
	case monsterAIModeChase:
		if s.monsterInAttackRange(monster, player, def) {
			return Vec2{}, false
		}

		return s.findMonsterChaseGoal(monster, player, def)
	case monsterAIModeReturn:
		if distance(monster.pos, monster.spawnPos) <= nav.StopDistance {
			return Vec2{}, false
		}

		return monster.spawnPos, true
	default:
		return Vec2{}, false
	}
}

func (s *Sim) findMonsterChaseGoal(monster *entity, player *entity, def MonsterDef) (Vec2, bool) {
	nav := s.activeNav()
	candidates := s.monsterAttackSlotCandidates(monster, player, def)
	var (
		bestGoal       Vec2
		bestPathLen    int
		bestMonsterDst = math.MaxFloat64
		found          bool
	)
	for _, goal := range candidates {
		if !s.positionInNavigationBounds(nav, goal) || s.monsterPositionBlocked(goal, monster.id) {
			continue
		}
		if def.effectiveAttackMode() == attackModeRanged && !s.hasClearMonsterRangedShot(goal, player) {
			continue
		}
		blocked := s.buildMonsterBlockedFn(monster.id)
		steps, ok := PlanPath(nav, monster.pos, goal, blocked)
		if !ok {
			continue
		}
		if len(steps) == 0 && distance(monster.pos, goal) > nav.CellSize+nav.StopDistance {
			continue
		}
		monsterDst := distance(monster.pos, goal)
		if !found || len(steps) < bestPathLen ||
			(len(steps) == bestPathLen && monsterDst < bestMonsterDst-1e-9) ||
			(len(steps) == bestPathLen && math.Abs(monsterDst-bestMonsterDst) <= 1e-9 && vecLess(goal, bestGoal)) {
			bestGoal = goal
			bestPathLen = len(steps)
			bestMonsterDst = monsterDst
			found = true
		}
	}
	if !found {
		return Vec2{}, false
	}

	return bestGoal, true
}

func (s *Sim) monsterAttackSlotCandidates(monster *entity, player *entity, def MonsterDef) []Vec2 {
	attackDistance := s.monsterAttackReach(def) + playerRadius - 0.05
	minSeparation := playerRadius + monsterRadius + 0.05
	if attackDistance < minSeparation {
		attackDistance = minSeparation
	}
	directions := []Vec2{}
	addDirection := func(dir Vec2) {
		if dir.X == 0 && dir.Y == 0 {
			return
		}
		normalized := normalize(dir)
		for _, existing := range directions {
			if math.Abs(existing.X-normalized.X) <= 1e-6 && math.Abs(existing.Y-normalized.Y) <= 1e-6 {
				return
			}
		}
		directions = append(directions, normalized)
	}
	addDirection(Vec2{X: monster.pos.X - player.pos.X, Y: monster.pos.Y - player.pos.Y})
	for i := 0; i < 16; i++ {
		angle := (2 * math.Pi * float64(i)) / 16
		addDirection(Vec2{X: math.Cos(angle), Y: math.Sin(angle)})
	}
	candidates := make([]Vec2, 0, len(directions))
	for _, dir := range directions {
		candidates = append(candidates, Vec2{
			X: player.pos.X + dir.X*attackDistance,
			Y: player.pos.Y + dir.Y*attackDistance,
		})
	}
	sort.Slice(candidates, func(i, j int) bool {
		di := distance(monster.pos, candidates[i])
		dj := distance(monster.pos, candidates[j])
		if math.Abs(di-dj) > 1e-9 {
			return di < dj
		}
		return vecLess(candidates[i], candidates[j])
	})
	return candidates
}

func (s *Sim) monsterMoveDelta(pos Vec2, goal Vec2, steps []Vec2, speed float64) Vec2 {
	toGoal := Vec2{X: goal.X - pos.X, Y: goal.Y - pos.Y}
	dist := distance(pos, goal)
	if dist <= 1e-9 {
		return Vec2{}
	}
	if len(steps) == 0 || dist <= speed+1e-9 || dist <= s.activeNav().CellSize+s.activeNav().StopDistance {
		if dist <= speed+1e-9 {
			return toGoal
		}
		dir := normalize(toGoal)
		return Vec2{X: dir.X * speed, Y: dir.Y * speed}
	}
	return Vec2{
		X: steps[0].X * speed,
		Y: steps[0].Y * speed,
	}
}

func cellLess(a, b gridCell) bool {
	if a.y != b.y {
		return a.y < b.y
	}

	return a.x < b.x
}

func vecLess(a, b Vec2) bool {
	if math.Abs(a.Y-b.Y) > 1e-9 {
		return a.Y < b.Y
	}

	return a.X < b.X-1e-9
}

func (s *Sim) buildMonsterBlockedFn(excludeMonsterID uint64) func(gx, gy int) bool {
	return func(gx, gy int) bool {
		center := gridToWorld(s.activeNav(), gridCell{x: gx, y: gy})
		return s.monsterPositionBlocked(center, excludeMonsterID)
	}
}

func (s *Sim) monsterPositionBlocked(pos Vec2, excludeMonsterID uint64) bool {
	for _, wall := range s.activeWalls() {
		if circleIntersectsAABB(pos, monsterRadius, wall.pos, wall.size) {
			return true
		}
	}
	for _, playerID := range sortedPlayerIDs(s.players) {
		ps := s.players[playerID]
		if ps == nil || !ps.Connected || ps.CurrentLevel != s.currentLevel {
			continue
		}
		player := s.activeLevel().entities[playerID]
		if player != nil && player.hp > 0 {
			if circlesOverlap(pos, monsterRadius, player.pos, playerRadius) {
				return true
			}
		}
	}
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		if id == excludeMonsterID {
			continue
		}
		e := s.activeLevel().entities[id]
		if e == nil {
			continue
		}
		if e.kind == monsterEntity && e.hp > 0 {
			if circlesOverlap(pos, monsterRadius, e.pos, monsterRadius) {
				return true
			}
			continue
		}
		if e.kind == interactableEntity && e.state == interactableClosed {
			if def, ok := s.rules.Interactables[e.interactableDefID]; ok && def.BarrierWhenClosed != nil {
				if circleIntersectsAABB(pos, monsterRadius, e.pos, def.BarrierWhenClosed.Size) {
					return true
				}
			}
		}
	}

	return false
}

func (s *Sim) resolveMonsterMovement(monster *entity, delta Vec2) Vec2 {
	candidate := Vec2{X: monster.pos.X + delta.X, Y: monster.pos.Y + delta.Y}
	if !s.monsterPositionBlocked(candidate, monster.id) {
		return candidate
	}
	xOnly := Vec2{X: monster.pos.X + delta.X, Y: monster.pos.Y}
	if delta.X != 0 && !s.monsterPositionBlocked(xOnly, monster.id) {
		return xOnly
	}
	yOnly := Vec2{X: monster.pos.X, Y: monster.pos.Y + delta.Y}
	if delta.Y != 0 && !s.monsterPositionBlocked(yOnly, monster.id) {
		return yOnly
	}

	return monster.pos
}

func (s *Sim) advanceBossPhases(res *TickResult) {
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		boss := s.activeLevel().entities[id]
		if boss == nil || boss.kind != monsterEntity || !boss.isBoss || boss.hp <= 0 {
			continue
		}
		runtime, ok := s.ensureBossPhase(boss, res)
		if !ok {
			continue
		}
		if boss.bossPhaseKind == "active" {
			s.applyBossActivePhase(boss, runtime.phase, res)
		}
		if s.tick+1 >= boss.bossPhaseEnds {
			s.endBossPhase(boss, runtime, res)
		}
	}
}

func (s *Sim) ensureBossPhase(boss *entity, res *TickResult) (bossPhaseRuntime, bool) {
	if boss.bossPhaseKind != "" && s.tick < boss.bossPhaseEnds {
		return s.currentBossPhase(boss)
	}
	if boss.bossCooldownEnds > s.tick {
		return bossPhaseRuntime{}, false
	}
	next, ok := s.nextBossPhase(boss)
	if !ok {
		return bossPhaseRuntime{}, false
	}
	boss.bossPatternID = next.patternID
	boss.bossPhaseIndex = next.index
	boss.bossPhaseKind = next.phase.Kind
	boss.bossPhaseStarted = s.tick
	boss.bossPhaseEnds = s.tick + uint64(next.phase.DurationTicks)
	boss.bossActiveHit = map[uint64]bool{}
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(boss))})
	res.Events = append(res.Events, bossPhaseEvent("boss_phase_started", boss, next))
	return next, true
}

func (s *Sim) currentBossPhase(boss *entity) (bossPhaseRuntime, bool) {
	pattern, ok := s.rules.BossPatterns[boss.bossPatternID]
	if !ok || boss.bossPhaseIndex < 0 || boss.bossPhaseIndex >= len(pattern.Phases) {
		return bossPhaseRuntime{}, false
	}
	return bossPhaseRuntime{
		patternID: boss.bossPatternID,
		index:     boss.bossPhaseIndex,
		phase:     pattern.Phases[boss.bossPhaseIndex],
	}, true
}

func (s *Sim) nextBossPhase(boss *entity) (bossPhaseRuntime, bool) {
	template, ok := s.rules.BossTemplates[boss.bossTemplateID]
	if !ok || len(template.PatternDeck) == 0 {
		return bossPhaseRuntime{}, false
	}
	patternID := boss.bossPatternID
	if patternID == "" {
		boss.bossPatternDeckIndex = 0
		patternID = template.PatternDeck[0]
		boss.bossPatternID = patternID
	}
	pattern, ok := s.rules.BossPatterns[patternID]
	if !ok || len(pattern.Phases) == 0 {
		return bossPhaseRuntime{}, false
	}
	nextIndex := boss.bossPhaseIndex + 1
	if boss.bossPhaseKind == "" {
		nextIndex = 0
	}
	if nextIndex >= len(pattern.Phases) {
		nextIndex = 0
	}
	return bossPhaseRuntime{patternID: patternID, index: nextIndex, phase: pattern.Phases[nextIndex]}, true
}

func (s *Sim) endBossPhase(boss *entity, runtime bossPhaseRuntime, res *TickResult) {
	res.Events = append(res.Events, bossPhaseEvent("boss_phase_ended", boss, runtime))
	pattern := s.rules.BossPatterns[runtime.patternID]
	if runtime.index >= len(pattern.Phases)-1 {
		boss.bossCooldownEnds = s.tick + 1 + uint64(pattern.CooldownTicks)
		boss.bossPhaseKind = ""
		boss.bossPhaseIndex = -1
		boss.bossPhaseStarted = 0
		boss.bossPhaseEnds = 0
		boss.bossActiveHit = nil
		s.advanceBossPatternDeck(boss)
	} else {
		next := bossPhaseRuntime{patternID: runtime.patternID, index: runtime.index + 1, phase: pattern.Phases[runtime.index+1]}
		boss.bossPhaseIndex = next.index
		boss.bossPhaseKind = next.phase.Kind
		boss.bossPhaseStarted = s.tick + 1
		boss.bossPhaseEnds = s.tick + 1 + uint64(next.phase.DurationTicks)
		boss.bossActiveHit = map[uint64]bool{}
		res.Events = append(res.Events, bossPhaseEvent("boss_phase_started", boss, next))
	}
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(boss))})
}

func (s *Sim) advanceBossPatternDeck(boss *entity) {
	template, ok := s.rules.BossTemplates[boss.bossTemplateID]
	if !ok || len(template.PatternDeck) == 0 {
		return
	}
	nextIndex := boss.bossPatternDeckIndex + 1
	if nextIndex < 0 || nextIndex >= len(template.PatternDeck) {
		nextIndex = 0
	}
	boss.bossPatternDeckIndex = nextIndex
	boss.bossPatternID = template.PatternDeck[nextIndex]
}

func (s *Sim) applyBossActivePhase(boss *entity, phase BossPatternPhase, res *TickResult) {
	if phase.Damage == nil {
		return
	}
	for _, playerID := range sortedPlayerIDs(s.players) {
		ps := s.players[playerID]
		if ps == nil || !ps.Connected || ps.CurrentLevel != s.currentLevel || boss.bossActiveHit[playerID] {
			continue
		}
		player := s.activeLevel().entities[playerID]
		if player == nil || player.hp <= 0 || !bossPhaseHitsPlayer(boss, player, phase) {
			continue
		}
		s.usePlayer(ps)
		scaledDamage := s.scaleMonsterDamageForParty(s.currentLevel, *phase.Damage)
		attackerStats := s.monsterEffectiveCombatStats(boss, scaledDamage)
		defenderStats, _ := s.playerEffectiveCombatStats()
		outcome := s.resolveCombat(attackerStats, defenderStats, scaledDamage)
		boss.bossActiveHit[playerID] = true
		if !outcome.Hit || outcome.Blocked {
			res.Events = append(res.Events, combatEvent(s.combatEventType(playerEntity, outcome), boss.id, player.id, "", outcome))
			continue
		}
		player.hp -= outcome.Damage
		if player.hp < 0 {
			player.hp = 0
		}
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
		eventType := "player_damaged"
		if player.hp == 0 {
			eventType = "player_killed"
		}
		res.Events = append(res.Events, combatEvent(eventType, boss.id, player.id, "", outcome))
	}
}

func bossPhaseHitsPlayer(boss, player *entity, phase BossPatternPhase) bool {
	switch phase.Shape {
	case "melee_contact":
		radius := phase.Radius
		if radius <= 0 {
			radius = monsterRadius + playerRadius
		}
		return distance(boss.pos, player.pos) <= radius
	case "circle":
		if phase.Radius <= 0 {
			return false
		}
		return distance(boss.pos, player.pos) <= phase.Radius
	default:
		return false
	}
}

func bossPhaseEvent(eventType string, boss *entity, runtime bossPhaseRuntime) Event {
	return Event{
		EventType:     eventType,
		EntityID:      idStr(boss.id),
		PatternID:     runtime.patternID,
		PhaseIndex:    intPtr(runtime.index),
		PhaseKind:     runtime.phase.Kind,
		DurationTicks: intPtr(runtime.phase.DurationTicks),
		Telegraph:     bossTelegraphView(runtime.phase),
		HitShape:      bossHitShapeView(runtime.phase),
	}
}

func bossTelegraphView(phase BossPatternPhase) *BossTelegraphView {
	if phase.TelegraphType == "" {
		return nil
	}
	return &BossTelegraphView{
		Type:      phase.TelegraphType,
		FromColor: phase.FromColor,
		ToColor:   phase.ToColor,
		HitShape:  phase.HitShape,
		Radius:    phase.Radius,
	}
}

func bossHitShapeView(phase BossPatternPhase) *BossHitShapeView {
	shape := phase.Shape
	if shape == "" {
		shape = phase.HitShape
	}
	if shape == "" {
		return nil
	}
	return &BossHitShapeView{Shape: shape, Radius: phase.Radius}
}

func (s *Sim) advanceProjectiles(res *TickResult) {
	ids := sortedEntityIDs(s.activeLevel().entities)
	for _, id := range ids {
		p := s.activeLevel().entities[id]
		if p == nil || p.kind != projectileEntity {
			continue
		}
		s.advanceProjectile(p, res)
	}
}

func (s *Sim) advanceProjectile(p *entity, res *TickResult) {
	if p.spawnTick == s.tick {
		return
	}
	delta := p.speed * tickDuration
	if delta <= 0 {
		return
	}
	candidate := Vec2{X: p.pos.X + p.dir.X*delta, Y: p.pos.Y + p.dir.Y*delta}
	segmentLength := distance(p.pos, candidate)
	hit, ok := s.firstProjectileHit(p, candidate)
	if ok {
		p.pos = hit.pos
		s.resolveProjectileHit(p, hit, res)
		delete(s.activeLevel().entities, p.id)
		res.Changes = append(res.Changes, Change{Op: OpEntityRemove, EntityID: idStr(p.id)})
		return
	}
	if p.traveled+segmentLength >= p.maxDistance-meleeRangeEpsilon {
		res.Events = append(res.Events, Event{EventType: "projectile_expired", CorrelationID: p.sourceCorrID})
		delete(s.activeLevel().entities, p.id)
		res.Changes = append(res.Changes, Change{Op: OpEntityRemove, EntityID: idStr(p.id)})
		return
	}
	p.pos = candidate
	p.traveled += segmentLength
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(p))})
}

type projectileHit struct {
	t        float64
	category int
	entityID uint64
	pos      Vec2
}

const (
	projectileHitWall = iota
	projectileHitInteractable
	projectileHitMonster
	projectileHitPlayer
)

func (s *Sim) firstProjectileHit(p *entity, candidate Vec2) (projectileHit, bool) {
	best := projectileHit{t: math.Inf(1)}
	found := false
	ownerKind := ""
	if owner := s.activeLevel().entities[p.ownerID]; owner != nil {
		ownerKind = owner.kind
	}
	consider := func(hit projectileHit) {
		hit.pos = Vec2{
			X: p.pos.X + (candidate.X-p.pos.X)*hit.t,
			Y: p.pos.Y + (candidate.Y-p.pos.Y)*hit.t,
		}
		if !found || projectileHitLess(hit, best) {
			best = hit
			found = true
		}
	}
	for _, wall := range s.activeWalls() {
		if t, ok := segmentIntersectsInflatedAABB(p.pos, candidate, wall.pos, wall.size, projectileRadius); ok {
			consider(projectileHit{t: t, category: projectileHitWall})
		}
	}
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		e := s.activeLevel().entities[id]
		if e == nil || e.id == p.id {
			continue
		}
		switch e.kind {
		case interactableEntity:
			if e.state != interactableClosed {
				continue
			}
			def, ok := s.rules.Interactables[e.interactableDefID]
			if !ok || def.BarrierWhenClosed == nil {
				continue
			}
			if t, ok := segmentIntersectsInflatedAABB(p.pos, candidate, e.pos, def.BarrierWhenClosed.Size, projectileRadius); ok {
				consider(projectileHit{t: t, category: projectileHitInteractable, entityID: e.id})
			}
		case monsterEntity:
			if ownerKind != playerEntity || e.hp <= 0 {
				continue
			}
			if t, ok := segmentIntersectsCircle(p.pos, candidate, e.pos, monsterRadius+projectileRadius); ok {
				consider(projectileHit{t: t, category: projectileHitMonster, entityID: e.id})
			}
		case playerEntity:
			if ownerKind != monsterEntity || e.hp <= 0 {
				continue
			}
			if p.targetID != 0 && e.id != p.targetID {
				continue
			}
			if t, ok := segmentIntersectsCircle(p.pos, candidate, e.pos, playerRadius+projectileRadius); ok {
				consider(projectileHit{t: t, category: projectileHitPlayer, entityID: e.id})
			}
		}
	}
	return best, found
}

func projectileHitLess(a, b projectileHit) bool {
	if math.Abs(a.t-b.t) > 1e-9 {
		return a.t < b.t
	}
	if a.category != b.category {
		return a.category < b.category
	}
	return a.entityID < b.entityID
}

func (s *Sim) resolveProjectileHit(p *entity, hit projectileHit, res *TickResult) {
	owner := s.players[p.ownerID]
	if owner != nil {
		s.usePlayer(owner)
		defer s.savePlayer(owner)
	}
	if hit.category == projectileHitPlayer {
		s.resolveMonsterProjectileHit(p, hit, res)
		return
	}
	if hit.category != projectileHitMonster {
		res.Events = append(res.Events, Event{EventType: "projectile_blocked", CorrelationID: p.sourceCorrID})
		return
	}
	target := s.activeLevel().entities[hit.entityID]
	if target == nil || target.kind != monsterEntity || target.hp <= 0 {
		res.Events = append(res.Events, Event{EventType: "projectile_expired", CorrelationID: p.sourceCorrID})
		return
	}
	s.damageMonsterByPlayer(target, p.ownerID, p.sourceCorrID, res, p.damageRange)
}

func (s *Sim) resolveMonsterProjectileHit(p *entity, hit projectileHit, res *TickResult) {
	owner := s.activeLevel().entities[p.ownerID]
	target := s.activeLevel().entities[hit.entityID]
	if owner == nil || owner.kind != monsterEntity || owner.hp <= 0 || target == nil || target.kind != playerEntity || target.hp <= 0 {
		res.Events = append(res.Events, Event{EventType: "projectile_expired", CorrelationID: p.sourceCorrID})
		return
	}
	ps := s.players[target.id]
	if ps == nil || !ps.Connected || ps.CurrentLevel != s.currentLevel {
		res.Events = append(res.Events, Event{EventType: "projectile_expired", CorrelationID: p.sourceCorrID})
		return
	}
	s.usePlayer(ps)
	s.damagePlayerByMonster(owner, target, p.damageRange, p.sourceCorrID, res)
}

func (s *Sim) aggroMonsterOnHit(monster *entity, playerID uint64, corr string, res *TickResult) {
	if monster == nil || monster.kind != monsterEntity || playerID == 0 {
		return
	}
	level := s.activeLevel()
	player := level.entities[playerID]
	queue := []*entity{monster}
	queued := map[uint64]bool{monster.id: true}
	if player != nil && player.kind == playerEntity && player.hp > 0 {
		for _, candidateID := range sortedEntityIDs(level.entities) {
			candidate := level.entities[candidateID]
			if candidate == nil || queued[candidate.id] || !s.canAggroAttackingPlayer(candidate, player) {
				continue
			}
			queued[candidate.id] = true
			queue = append(queue, candidate)
		}
	}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current == nil || current.kind != monsterEntity {
			continue
		}
		if current.hp > 0 {
			s.aggroSingleMonster(current, playerID, corr, res)
		}
		for _, candidateID := range sortedEntityIDs(level.entities) {
			candidate := level.entities[candidateID]
			if candidate == nil || queued[candidate.id] || !s.canJoinGroupAggro(current, candidate) {
				continue
			}
			queued[candidate.id] = true
			queue = append(queue, candidate)
		}
	}
}

func (s *Sim) aggroSingleMonster(monster *entity, playerID uint64, corr string, res *TickResult) {
	wasChasingTarget := monster.aiMode == monsterAIModeChase && monster.aiTargetPlayerID == playerID
	monster.aiTargetPlayerID = playerID
	monster.aiMode = monsterAIModeChase
	if wasChasingTarget {
		return
	}
	res.Events = append(res.Events, Event{
		EventType:      "monster_aggro",
		EntityID:       idStr(monster.id),
		SourceEntityID: idStr(playerID),
		TargetEntityID: idStr(monster.id),
		CorrelationID:  corr,
	})
}

func (s *Sim) canAggroAttackingPlayer(candidate, player *entity) bool {
	if candidate == nil || player == nil || candidate.kind != monsterEntity || candidate.hp <= 0 || player.kind != playerEntity || player.hp <= 0 {
		return false
	}
	def, ok := s.rules.Monsters[candidate.monsterDefID]
	if !ok || def.effectiveBehavior() != monsterBehaviorChase || def.AggroRadius <= 0 {
		return false
	}
	return distance(candidate.pos, player.pos) <= def.AggroRadius
}

func (s *Sim) canJoinGroupAggro(source, candidate *entity) bool {
	if source == nil || candidate == nil || source.id == candidate.id || candidate.kind != monsterEntity || candidate.hp <= 0 {
		return false
	}
	def, ok := s.rules.Monsters[candidate.monsterDefID]
	if !ok || def.effectiveBehavior() != monsterBehaviorChase {
		return false
	}
	radius := def.AggroRadius
	if radius <= 0 {
		return false
	}
	return distance(source.pos, candidate.pos) <= radius
}

func (s *Sim) awardMonsterExperience(monster *entity, sourceID uint64, corr string, res *TickResult) {
	def, ok := s.rules.Monsters[monster.monsterDefID]
	if !ok || def.XPReward <= 0 {
		return
	}
	xpReward := def.XPReward
	if monster.monsterXPReward > 0 {
		xpReward = monster.monsterXPReward
	}
	if !s.rules.Combat.Coop.XPShare.Enabled {
		s.awardExperienceToPlayer(sourceID, xpReward, corr, res)
		return
	}
	eligible := s.coopXPEligiblePlayers(monster)
	if len(eligible) == 0 && sourceID != 0 {
		eligible = []uint64{sourceID}
	}
	for _, playerID := range eligible {
		s.awardExperienceToPlayer(playerID, xpReward, corr, res)
	}
}

func (s *Sim) awardExperience(amount int, corr string, res *TickResult) {
	s.awardExperienceForCurrentPlayer(amount, corr, res, 0)
}

func (s *Sim) awardExperienceToPlayer(playerID uint64, amount int, corr string, res *TickResult) {
	ps := s.players[playerID]
	if ps == nil {
		return
	}
	restore := s.players[s.playerID]
	s.usePlayer(ps)
	s.awardExperienceForCurrentPlayer(amount, corr, res, playerID)
	s.savePlayer(ps)
	if restore != nil {
		s.usePlayer(restore)
	}
}

func (s *Sim) awardExperienceForCurrentPlayer(amount int, corr string, res *TickResult, ownerPlayerID uint64) {
	if amount <= 0 {
		return
	}
	s.progression.Experience += amount
	res.Events = append(res.Events, Event{
		EventType:       "experience_gained",
		EntityID:        idStr(s.playerID),
		CorrelationID:   corr,
		Amount:          intPtr(amount),
		TotalExperience: intPtr(s.progression.Experience),
	})

	for s.progression.Level < s.rules.CharacterProgression.LevelCap {
		nextXP, ok := s.rules.nextLevelTotalXP(s.progression.Level)
		if !ok || s.progression.Experience < nextXP {
			break
		}
		from := s.progression.Level
		s.progression.Level++
		s.progression.UnspentStatPoints += s.rules.CharacterProgression.PointsPerLevel
		to := s.progression.Level
		res.Events = append(res.Events, Event{
			EventType:         "character_leveled",
			EntityID:          idStr(s.playerID),
			CorrelationID:     corr,
			FromLevel:         intPtr(from),
			ToLevel:           intPtr(to),
			UnspentStatPoints: intPtr(s.progression.UnspentStatPoints),
		})
		if gained := s.rules.skillPointsGrantedAtLevel(to); gained > 0 {
			s.progression.UnspentSkillPoints += gained
			res.Events = append(res.Events, Event{
				EventType:          "skill_point_gained",
				EntityID:           idStr(s.playerID),
				CorrelationID:      corr,
				Amount:             intPtr(gained),
				UnspentSkillPoints: intPtr(s.progression.UnspentSkillPoints),
			})
		}
	}

	view := s.CharacterProgressionView()
	res.Changes = append(res.Changes, Change{Op: OpCharacterProgressionUpdate, OwnerPlayerID: ownerPlayerID, Progression: &view})
	skillView := s.SkillProgressionView()
	res.Changes = append(res.Changes, Change{Op: OpSkillProgressionUpdate, OwnerPlayerID: ownerPlayerID, SkillProgression: &skillView})
	s.appendInventoryPresentationUpdatesForOwner(res, ownerPlayerID)
}

func (s *Sim) coopXPEligiblePlayers(monster *entity) []uint64 {
	if monster == nil {
		return nil
	}
	radius := s.rules.Combat.Coop.XPShare.Radius
	level, ok := s.levelContainingEntity(monster)
	if !ok {
		return nil
	}
	levelNum := level.levelNum
	eligible := []uint64{}
	for _, playerID := range sortedPlayerIDs(s.players) {
		ps := s.players[playerID]
		if ps == nil || !ps.Connected || ps.CurrentLevel != levelNum {
			continue
		}
		player := level.entities[playerID]
		if player == nil || player.hp <= 0 {
			continue
		}
		if distance(player.pos, monster.pos) > radius {
			continue
		}
		eligible = append(eligible, playerID)
	}
	return eligible
}

func (s *Sim) levelContainingEntity(target *entity) (*LevelState, bool) {
	if target == nil {
		return nil, false
	}
	if level := s.levels[s.currentLevel]; level != nil && level.entities[target.id] == target {
		return level, true
	}
	for _, levelNum := range s.sortedLevelNums() {
		level := s.levels[levelNum]
		if level != nil && level.entities[target.id] == target {
			return level, true
		}
	}
	return nil, false
}

func (s *Sim) applyPartyHPScale(level *LevelState, monster *entity) {
	if level == nil || monster == nil || !s.rules.Combat.Coop.PartyChallenge.HPScalesAtSpawn {
		return
	}
	multiplier := s.partyChallengeMultiplierForLevel(level.levelNum)
	if multiplier <= 1 {
		return
	}
	monster.maxHP = roundPositive(float64(monster.maxHP) * multiplier)
	monster.hp = monster.maxHP
}

func (s *Sim) scaleMonsterDamageForParty(levelNum int, base DamageRange) DamageRange {
	if !s.rules.Combat.Coop.PartyChallenge.DamageScalesAtAttack {
		return base
	}
	multiplier := s.partyChallengeMultiplierForLevel(levelNum)
	if multiplier <= 1 {
		return base
	}
	return scaleDamageRange(base, multiplier)
}

func (s *Sim) partyChallengeMultiplierForLevel(levelNum int) float64 {
	count := s.aliveConnectedPlayerCountOnLevel(levelNum)
	return s.rules.Combat.Coop.PartyChallenge.Multiplier(count)
}

func (s *Sim) aliveConnectedPlayerCountOnLevel(levelNum int) int {
	level := s.levels[levelNum]
	if level == nil {
		return 0
	}
	count := 0
	for _, playerID := range sortedPlayerIDs(s.players) {
		ps := s.players[playerID]
		if ps == nil || !ps.Connected || ps.CurrentLevel != levelNum {
			continue
		}
		player := level.entities[playerID]
		if player == nil || player.hp <= 0 {
			continue
		}
		count++
	}
	return count
}

func segmentIntersectsInflatedAABB(start, end, rectCenter, rectSize Vec2, inflate float64) (float64, bool) {
	halfX := rectSize.X/2 + inflate
	halfY := rectSize.Y/2 + inflate
	minX, maxX := rectCenter.X-halfX, rectCenter.X+halfX
	minY, maxY := rectCenter.Y-halfY, rectCenter.Y+halfY
	dx := end.X - start.X
	dy := end.Y - start.Y
	tmin, tmax := 0.0, 1.0
	if !clipSegmentAxis(start.X, dx, minX, maxX, &tmin, &tmax) {
		return 0, false
	}
	if !clipSegmentAxis(start.Y, dy, minY, maxY, &tmin, &tmax) {
		return 0, false
	}
	return tmin, true
}

func clipSegmentAxis(start, delta, minV, maxV float64, tmin, tmax *float64) bool {
	if math.Abs(delta) < 1e-12 {
		return start >= minV && start <= maxV
	}
	inv := 1 / delta
	t1 := (minV - start) * inv
	t2 := (maxV - start) * inv
	if t1 > t2 {
		t1, t2 = t2, t1
	}
	if t1 > *tmin {
		*tmin = t1
	}
	if t2 < *tmax {
		*tmax = t2
	}
	return *tmin <= *tmax && *tmax >= 0 && *tmin <= 1
}

func segmentIntersectsCircle(start, end, center Vec2, radius float64) (float64, bool) {
	d := Vec2{X: end.X - start.X, Y: end.Y - start.Y}
	f := Vec2{X: start.X - center.X, Y: start.Y - center.Y}
	a := d.X*d.X + d.Y*d.Y
	if a == 0 {
		if distance(start, center) <= radius {
			return 0, true
		}
		return 0, false
	}
	b := 2 * (f.X*d.X + f.Y*d.Y)
	c := f.X*f.X + f.Y*f.Y - radius*radius
	discriminant := b*b - 4*a*c
	if discriminant < 0 {
		return 0, false
	}
	root := math.Sqrt(discriminant)
	t1 := (-b - root) / (2 * a)
	t2 := (-b + root) / (2 * a)
	if t1 >= 0 && t1 <= 1 {
		return t1, true
	}
	if t2 >= 0 && t2 <= 1 {
		return t2, true
	}
	return 0, false
}

func (s *Sim) rollDamage() int {
	return s.rollRange(s.resolvePlayerAttackDamage())
}

func (s *Sim) resolveCombat(attacker, defender effectiveCombatStats, damageRange DamageRange) combatResolution {
	hit := s.rollChance(attacker.HitChance)
	if !hit {
		return combatResolution{Outcome: "miss", Hit: false}
	}
	blocked := s.rollChance(defender.BlockPercent / 100.0)
	if blocked {
		return combatResolution{Outcome: "block", Hit: true, Blocked: true}
	}
	if damageRange.Max < damageRange.Min {
		damageRange.Max = damageRange.Min
	}
	raw := s.rollRange(damageRange)
	critical := s.rollChance(attacker.CritChance)
	rawOrCrit := raw
	outcome := "hit"
	if critical {
		rawOrCrit = roundPositive(float64(raw) * attacker.CritDamage)
		outcome = "crit"
	}
	mitigated := rawOrCrit - int(math.Round(defender.Armor))
	finalDamage := mitigated
	if finalDamage < s.rules.Combat.MinimumDamage {
		finalDamage = s.rules.Combat.MinimumDamage
	}
	return combatResolution{
		Outcome:         outcome,
		Damage:          finalDamage,
		RawDamage:       rawOrCrit,
		MitigatedDamage: mitigated,
		Blocked:         false,
		Critical:        critical,
		Hit:             true,
	}
}

func (s *Sim) rollChance(chance float64) bool {
	draw := s.rng.Next()
	if chance <= 0 {
		return false
	}
	if chance >= 1 {
		return true
	}
	return float64(draw%10000)/10000.0 < chance
}

func (s *Sim) combatEventType(defenderKind string, outcome combatResolution) string {
	if outcome.Outcome == "miss" {
		return "attack_missed"
	}
	if defenderKind == playerEntity {
		return "player_damaged"
	}
	return "monster_damaged"
}

func combatEvent(eventType string, sourceID, targetID uint64, corr string, outcome combatResolution) Event {
	return Event{
		EventType:       eventType,
		EntityID:        idStr(targetID),
		SourceEntityID:  idStr(sourceID),
		TargetEntityID:  idStr(targetID),
		CorrelationID:   corr,
		Damage:          intPtr(outcome.Damage),
		Outcome:         outcome.Outcome,
		RawDamage:       intPtr(outcome.RawDamage),
		MitigatedDamage: intPtr(outcome.MitigatedDamage),
		Blocked:         boolPtr(outcome.Blocked),
		Critical:        boolPtr(outcome.Critical),
	}
}

func (s *Sim) rollItemTemplate(templateID string) (ItemRollPayload, bool) {
	return s.rules.rollItemTemplateWithRNG(templateID, s.rng)
}

func weightedRollableStat(stats []RollableStatDef, rng *RNG) (RollableStatDef, bool) {
	total := 0
	for _, stat := range stats {
		total += stat.Weight
	}
	if total <= 0 {
		return RollableStatDef{}, false
	}
	roll := rng.IntN(total)
	for _, stat := range stats {
		roll -= stat.Weight
		if roll < 0 {
			return stat, true
		}
	}
	return stats[len(stats)-1], true
}

func (s *Sim) itemSlot(itemDefID string, payload *ItemRollPayload) string {
	if payload != nil {
		if template, ok := s.rules.ItemTemplates[payload.ItemTemplateID]; ok {
			return template.Slot
		}
	}
	if def, ok := s.rules.Items[itemDefID]; ok {
		return def.Slot
	}
	return ""
}

func newEquippedMap() map[string]uint64 {
	out := make(map[string]uint64, len(equipmentSlots))
	for _, slot := range equipmentSlots {
		out[slot] = 0
	}
	return out
}

func slotAcceptsItemSlot(slot, itemSlot string) bool {
	if itemSlot == "ring" {
		return slot == ringLeftSlot || slot == ringRightSlot
	}
	return slot == itemSlot
}

func (s *Sim) itemOccupiesHands(item *invItem) []string {
	if item == nil {
		return nil
	}
	if item.rollPayload != nil {
		if template, ok := s.rules.ItemTemplates[item.rollPayload.ItemTemplateID]; ok {
			return cloneStringSlice(template.OccupiesHands)
		}
		return nil
	}
	if def, ok := s.rules.Items[item.itemDefID]; ok {
		return cloneStringSlice(def.OccupiesHands)
	}
	return nil
}

func validHotbarSlot(slotIndex int) bool {
	return slotIndex >= 0 && slotIndex < maxHotbarCapacity
}

func (s *Sim) hotbarCapacity() int {
	capacity := minHotbarCapacity
	belt := s.findItemByID(s.equipped["belt"])
	if belt != nil {
		if belt.rollPayload != nil {
			if rolled := belt.rollPayload.Stats["hotbar_slots"]; rolled > 0 {
				capacity = rolled
			} else if template, ok := s.rules.ItemTemplates[belt.rollPayload.ItemTemplateID]; ok {
				capacity = template.BaseStats["hotbar_slots"]
			}
		} else if template, ok := s.rules.ItemTemplates[belt.itemDefID]; ok {
			capacity = template.BaseStats["hotbar_slots"]
		}
	}
	if capacity < minHotbarCapacity {
		return minHotbarCapacity
	}
	if capacity > maxHotbarCapacity {
		return maxHotbarCapacity
	}
	return capacity
}

func (s *Sim) inventoryRows() int {
	rows := baseInventoryRows
	for _, slot := range equipmentSlots {
		item := s.findItemByID(s.equipped[slot])
		rows += s.itemInventoryRows(item)
	}
	if rows < 0 {
		return 0
	}
	if rows > maxInventoryRows {
		return maxInventoryRows
	}
	return rows
}

func inventoryCapacityForRows(rows int) int {
	if rows < 0 {
		rows = 0
	}
	if rows > maxInventoryRows {
		rows = maxInventoryRows
	}
	return rows * inventoryColumns
}

func (s *Sim) inventoryCapacity() int {
	return inventoryCapacityForRows(s.inventoryRows())
}

func (s *Sim) inventoryCapacityWithItemUnequipped(item *invItem) int {
	rows := s.inventoryRows()
	if item != nil && item.equipped {
		rows -= s.itemInventoryRows(item)
	}
	return inventoryCapacityForRows(rows)
}

func (s *Sim) inventoryRowsAfterEquip(slot string, item *invItem, clearedSlots []string) int {
	rows := s.inventoryRows()
	cleared := map[string]bool{}
	for _, clearedSlot := range clearedSlots {
		if cleared[clearedSlot] {
			continue
		}
		cleared[clearedSlot] = true
		prevID := s.equipped[clearedSlot]
		if prevID == 0 {
			continue
		}
		prev := s.findItemByID(prevID)
		if prev != nil && prev != item {
			rows -= s.itemInventoryRows(prev)
		}
	}
	if item != nil {
		if item.equipped {
			currentSlot := ""
			for _, eqSlot := range sortedStringKeys(s.equipped) {
				if s.equipped[eqSlot] == item.instanceID {
					currentSlot = eqSlot
					break
				}
			}
			if currentSlot == "" || currentSlot != slot {
				rows -= s.itemInventoryRows(item)
				rows += s.itemInventoryRows(item)
			}
		} else {
			rows += s.itemInventoryRows(item)
		}
	}
	return rows
}

func (s *Sim) itemInventoryRows(item *invItem) int {
	if item == nil {
		return 0
	}
	if item.rollPayload != nil {
		if rolled := item.rollPayload.Stats["inventory_rows"]; rolled > 0 {
			return rolled
		}
		if template, ok := s.rules.ItemTemplates[item.rollPayload.ItemTemplateID]; ok {
			return template.BaseStats["inventory_rows"]
		}
		return 0
	}
	if template, ok := s.rules.ItemTemplates[item.itemDefID]; ok {
		return template.BaseStats["inventory_rows"]
	}
	return 0
}

func (s *Sim) bagOccupancyCount() int {
	count := 0
	for _, item := range s.inventory {
		if item == nil || item.equipped || s.hotbarHasItem(item.instanceID) {
			continue
		}
		count++
	}
	return count
}

func (s *Sim) hotbarHasItem(instanceID uint64) bool {
	if instanceID == 0 {
		return false
	}
	for _, assigned := range s.hotbar {
		if assigned == instanceID {
			return true
		}
	}
	return false
}

func (s *Sim) hotbarHasItemExcept(instanceID uint64, exceptSlot int) bool {
	if instanceID == 0 {
		return false
	}
	for slot, assigned := range s.hotbar {
		if slot == exceptSlot {
			continue
		}
		if assigned == instanceID {
			return true
		}
	}
	return false
}

func (s *Sim) hotbarView() []HotbarSlotView {
	if len(s.hotbar) != maxHotbarCapacity {
		s.hotbar = make([]uint64, maxHotbarCapacity)
	}
	out := make([]HotbarSlotView, 0, maxHotbarCapacity)
	for i, itemID := range s.hotbar {
		slot := HotbarSlotView{SlotIndex: i}
		if itemID != 0 {
			id := idStr(itemID)
			slot.ItemInstanceID = &id
			if item := s.findItemByID(itemID); item != nil {
				view := s.itemView(item)
				slot.Item = &view
			}
		}
		out = append(out, slot)
	}
	return out
}

func (s *Sim) itemIsConsumable(item *invItem) bool {
	if item == nil || item.rollPayload != nil {
		return false
	}
	def, ok := s.rules.Items[item.itemDefID]
	return ok && def.Category == "consumable"
}

func (s *Sim) newLootEntity(itemDefID string, pos Vec2, payload *ItemRollPayload, goldCtx goldRollContext) *entity {
	loot := &entity{kind: lootEntity, pos: pos, itemDefID: itemDefID, rollPayload: cloneRollPayload(payload)}
	if payload == nil && itemDefID == goldItemDefID {
		if amount, ok := s.rollGoldAmount(itemDefID, goldCtx); ok {
			loot.goldAmount = amount
		}
	}
	return loot
}

func (s *Sim) rollGoldAmount(itemDefID string, goldCtx goldRollContext) (int, bool) {
	def, ok := s.rules.Items[itemDefID]
	if !ok || def.Category != "currency" || def.Gold == nil {
		return 0, false
	}
	r := s.goldRangeForContext(*def.Gold, goldCtx)
	span := r.Max - r.Min + 1
	if span <= 1 {
		return r.Min, r.Min > 0
	}
	rollSeed := fmt.Sprintf("%s|gold|%d|%d|%s", s.seed, s.goldRoll, goldCtx.levelNum, goldCtx.monsterRarityID)
	s.goldRoll++
	amount := r.Min + NewRNG(SeedToUint64(rollSeed)).IntN(span)
	if amount <= 0 {
		return 0, false
	}
	return amount, true
}

func (s *Sim) goldRangeForContext(base DamageRange, goldCtx goldRollContext) DamageRange {
	scale := 1.0
	depth := 0
	if goldCtx.levelNum < 0 {
		depth = absInt(goldCtx.levelNum)
	}
	if goldCtx.monsterRarityID != "" {
		if rarity, ok := s.rules.DungeonGeneration.MonsterRarity(goldCtx.monsterRarityID); ok {
			scale *= rarity.XPMultiplier
			depth += rarity.LootDepthOffset
		}
	}
	if depth > 1 {
		scale *= 1.0 + float64(depth-1)*0.25
	}
	minAmount := roundPositive(float64(base.Min) * scale)
	maxAmount := roundPositive(float64(base.Max) * scale)
	if maxAmount < minAmount {
		maxAmount = minAmount
	}
	return DamageRange{Min: minAmount, Max: maxAmount}
}

func (s *Sim) clearHotbarReferences(instanceID uint64, res *TickResult) {
	for i, assigned := range s.hotbar {
		if assigned != instanceID {
			continue
		}
		s.hotbar[i] = 0
		res.Changes = append(res.Changes, Change{
			Op:             OpHotbarUpdate,
			SlotIndex:      i,
			ItemInstanceID: nil,
			InventoryRows:  intPtr(s.inventoryRows()),
			InventoryCap:   intPtr(s.inventoryCapacity()),
		})
	}
}

func (s *Sim) slotBlockedByHands(slot string, item *invItem) bool {
	if slot != offHandSlot {
		return false
	}
	mainHand := s.findItemByID(s.equipped[mainHandSlot])
	if mainHand == nil || mainHand.instanceID == item.instanceID {
		return false
	}
	for _, occupied := range s.itemOccupiesHands(mainHand) {
		if occupied == offHandSlot {
			return true
		}
	}
	return false
}

func (s *Sim) slotsClearedByEquip(slot string, item *invItem) []string {
	seen := map[string]bool{}
	out := make([]string, 0, 2)
	add := func(candidate string) {
		if !isEquipmentSlot(candidate) || seen[candidate] {
			return
		}
		seen[candidate] = true
		out = append(out, candidate)
	}
	add(slot)
	for _, occupied := range s.itemOccupiesHands(item) {
		add(occupied)
	}
	sort.Strings(out)
	return out
}

func (s *Sim) playerReach() float64 {
	base := s.rules.Combat.UnarmedReach
	instanceID := s.equipped[mainHandSlot]
	if instanceID == 0 {
		return base
	}
	item := s.findItemByID(instanceID)
	if item == nil {
		return base
	}
	reach, ok := s.itemReach(item)
	if !ok {
		return base
	}
	return reach
}

func (s *Sim) playerMeleeReach() float64 {
	item := s.equippedWeaponItem()
	if item == nil || s.playerAttackMode() == attackModeRanged {
		return s.rules.Combat.UnarmedReach
	}
	reach, ok := s.itemReach(item)
	if !ok {
		return s.rules.Combat.UnarmedReach
	}
	return reach
}

func (s *Sim) playerActionReach() float64 {
	return s.playerReach()
}

func (s *Sim) playerAttackMode() string {
	item := s.equippedWeaponItem()
	if item == nil {
		return attackModeMelee
	}
	if item.rollPayload != nil {
		if template, ok := s.rules.ItemTemplates[item.rollPayload.ItemTemplateID]; ok && template.AttackMode != "" {
			return template.AttackMode
		}
	}
	def, ok := s.rules.Items[item.itemDefID]
	if !ok || def.AttackMode == "" {
		return attackModeMelee
	}
	return def.AttackMode
}

func (s *Sim) equippedWeaponDef() (ItemDef, bool) {
	item := s.equippedWeaponItem()
	if item == nil || item.rollPayload != nil {
		return ItemDef{}, false
	}
	def, ok := s.rules.Items[item.itemDefID]
	return def, ok
}

func (s *Sim) equippedWeaponItem() *invItem {
	instanceID := s.equipped[mainHandSlot]
	if instanceID == 0 {
		return nil
	}
	item := s.findItemByID(instanceID)
	if item == nil {
		return nil
	}
	return item
}

func (s *Sim) playerProjectileSpeed() (float64, bool) {
	item := s.equippedWeaponItem()
	if item == nil {
		return 0, false
	}
	if item.rollPayload != nil {
		template, ok := s.rules.ItemTemplates[item.rollPayload.ItemTemplateID]
		if !ok || template.AttackMode != attackModeRanged || template.ProjectileSpeed <= 0 {
			return 0, false
		}
		return template.ProjectileSpeed, true
	}
	def, ok := s.rules.Items[item.itemDefID]
	if !ok || def.AttackMode != attackModeRanged || def.ProjectileSpeed == nil || *def.ProjectileSpeed <= 0 {
		return 0, false
	}
	return *def.ProjectileSpeed, true
}

func (s *Sim) itemReach(item *invItem) (float64, bool) {
	if item.rollPayload != nil {
		template, ok := s.rules.ItemTemplates[item.rollPayload.ItemTemplateID]
		if !ok || template.Reach <= 0 {
			return 0, false
		}
		return template.Reach, true
	}
	def, ok := s.rules.Items[item.itemDefID]
	if !ok || def.Reach == nil {
		return 0, false
	}
	return *def.Reach, true
}

func (s *Sim) itemEquipSlot(item *invItem) (string, bool) {
	if item.rollPayload != nil {
		template, ok := s.rules.ItemTemplates[item.rollPayload.ItemTemplateID]
		if !ok || !template.Equippable || template.Slot == "" {
			return "", false
		}
		return template.Slot, true
	}
	def, ok := s.rules.Items[item.itemDefID]
	if !ok || !def.Equippable {
		return "", false
	}
	return def.Slot, true
}

func (s *Sim) targetInteractionRadius(e *entity) float64 {
	switch e.kind {
	case monsterEntity:
		return monsterRadius
	case lootEntity:
		return lootInteractionRadius
	case interactableEntity:
		return interactableInteractionRadius
	default:
		return 0
	}
}

func (s *Sim) monsterAttackReach(def MonsterDef) float64 {
	if def.effectiveAttackMode() == attackModeRanged && def.AttackRange > 0 {
		return def.AttackRange
	}
	return s.rules.Combat.UnarmedReach
}

func (s *Sim) monsterInAttackRange(monster *entity, player *entity, def MonsterDef) bool {
	if !meleeInRange(distance(player.pos, monster.pos), s.monsterAttackReach(def), playerRadius) {
		return false
	}
	if def.effectiveAttackMode() == attackModeRanged {
		return s.hasClearMonsterRangedShot(monster.pos, player)
	}
	return true
}

func (s *Sim) inMeleeRange(target *entity) bool {
	player := s.activeLevel().entities[s.playerID]
	if player == nil {
		return false
	}
	return meleeInRange(distance(player.pos, target.pos), s.playerMeleeReach(), s.targetInteractionRadius(target))
}

func (s *Sim) inActionRange(target *entity) bool {
	player := s.activeLevel().entities[s.playerID]
	if player == nil {
		return false
	}
	return s.inActionRangeFrom(player.pos, target)
}

func (s *Sim) inActionRangeFrom(pos Vec2, target *entity) bool {
	return meleeInRange(distance(pos, target.pos), s.playerActionReach(), s.targetInteractionRadius(target))
}

func (s *Sim) inDispatchRange(target *entity) bool {
	if target.kind == monsterEntity && s.playerAttackMode() == attackModeRanged {
		player := s.activeLevel().entities[s.playerID]
		return player != nil && s.inActionRange(target) && s.hasClearRangedShot(player.pos, target)
	}
	return s.inMeleeRange(target)
}

func (s *Sim) hasClearRangedShot(from Vec2, target *entity) bool {
	if target == nil || target.kind != monsterEntity || target.hp <= 0 {
		return false
	}
	for _, wall := range s.activeWalls() {
		if _, ok := segmentIntersectsInflatedAABB(from, target.pos, wall.pos, wall.size, projectileRadius); ok {
			return false
		}
	}
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		e := s.activeLevel().entities[id]
		if e == nil || e.id == target.id {
			continue
		}
		switch e.kind {
		case interactableEntity:
			if e.state != interactableClosed {
				continue
			}
			def, ok := s.rules.Interactables[e.interactableDefID]
			if !ok || def.BarrierWhenClosed == nil {
				continue
			}
			if _, ok := segmentIntersectsInflatedAABB(from, target.pos, e.pos, def.BarrierWhenClosed.Size, projectileRadius); ok {
				return false
			}
		case monsterEntity:
			if e.hp <= 0 {
				continue
			}
			if _, ok := segmentIntersectsCircle(from, target.pos, e.pos, monsterRadius+projectileRadius); ok {
				return false
			}
		}
	}
	return true
}

func (s *Sim) hasClearMonsterRangedShot(from Vec2, target *entity) bool {
	if target == nil || target.kind != playerEntity || target.hp <= 0 {
		return false
	}
	for _, wall := range s.activeWalls() {
		if _, ok := segmentIntersectsInflatedAABB(from, target.pos, wall.pos, wall.size, projectileRadius); ok {
			return false
		}
	}
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		e := s.activeLevel().entities[id]
		if e == nil || e.id == target.id {
			continue
		}
		if e.kind != interactableEntity || e.state != interactableClosed {
			continue
		}
		def, ok := s.rules.Interactables[e.interactableDefID]
		if !ok || def.BarrierWhenClosed == nil {
			continue
		}
		if _, ok := segmentIntersectsInflatedAABB(from, target.pos, e.pos, def.BarrierWhenClosed.Size, projectileRadius); ok {
			return false
		}
	}
	return true
}

func meleeInRange(dist, reach, targetRadius float64) bool {
	return dist <= reach+targetRadius+meleeRangeEpsilon
}

func distance(a, b Vec2) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}

func (s *Sim) actionable(e *entity) bool {
	switch e.kind {
	case monsterEntity:
		return e.hp > 0
	case lootEntity:
		return true
	case interactableEntity:
		if e.interactableDefID == teleporterDefID && (e.state == interactableReady || e.state == interactableLocked || e.state == interactableDisabled) {
			return true
		}
		if s.shopIDForInteractable(e) != "" && e.state == interactableReady {
			return true
		}
		if s.stashIDForInteractable(e) != "" && e.state == interactableReady {
			return true
		}
		return e.state == interactableClosed || (e.state == interactableReady && e.interactableDefID == teleporterDefID)
	default:
		return false
	}
}

func (s *Sim) resolvePlayerAttackDamage() DamageRange {
	stats, _ := s.playerEffectiveCombatStats()
	minDamage := int(math.Floor(stats.DamageMin))
	maxDamage := int(math.Floor(stats.DamageMax))
	if minDamage < 0 {
		minDamage = 0
	}
	if maxDamage < minDamage {
		maxDamage = minDamage
	}
	return DamageRange{Min: minDamage, Max: maxDamage}
}

func (s *Sim) applyStrengthDamageBonus(base DamageRange) DamageRange {
	derived := s.characterDerivedStatsView()
	out := DamageRange{
		Min: base.Min + int(math.Floor(derived.DamageMin)),
		Max: base.Max + int(math.Floor(derived.DamageMax)),
	}
	if out.Min < 0 {
		out.Min = 0
	}
	if out.Max < out.Min {
		out.Max = out.Min
	}
	return out
}

func (s *Sim) currentMaxHP() int {
	stats, _ := s.playerEffectiveCombatStats()
	maxHP := int(math.Round(stats.MaxHP))
	if maxHP < 1 {
		return 1
	}
	return maxHP
}

func (s *Sim) currentMaxMana() int {
	mana := int(math.Round(s.characterDerivedStatsView().MaxMana))
	if mana < 0 {
		return 0
	}
	return mana
}

// CharacterProgressionView returns the authoritative protocol view of the
// current progression state.
func (s *Sim) CharacterProgressionView() CharacterProgressionView {
	remaining := s.experienceToNextLevel()
	return CharacterProgressionView{
		Level:                 s.progression.Level,
		Experience:            s.progression.Experience,
		ExperienceToNextLevel: remaining,
		LevelCap:              s.rules.CharacterProgression.LevelCap,
		UnspentStatPoints:     s.progression.UnspentStatPoints,
		UnspentSkillPoints:    s.progression.UnspentSkillPoints,
		Gold:                  s.gold,
		DeepestDungeonDepth:   s.progression.DeepestDungeonDepth,
		BaseStats:             s.progression.BaseStats,
		DerivedStats:          s.DerivedStatsView(),
		StatBreakdowns:        s.StatBreakdownViews(),
		SkillRanks:            cloneIntMap(s.progression.SkillRanks),
	}
}

// SkillProgressionView returns the authoritative protocol view of skill points
// and known skill ranks.
func (s *Sim) SkillProgressionView() SkillProgressionView {
	skills := make([]SkillProgressionSkillView, 0, len(s.rules.Skills))
	for _, skillID := range sortedStringKeys(s.rules.Skills) {
		def := s.rules.Skills[skillID]
		rank := s.progression.SkillRanks[skillID]
		skills = append(skills, SkillProgressionSkillView{
			SkillID:  skillID,
			Rank:     rank,
			MaxRank:  def.MaxRank,
			CanSpend: s.progression.UnspentSkillPoints > 0 && rank < def.MaxRank && s.skillRequirementsMet(def, rank+1),
		})
	}
	return SkillProgressionView{
		UnspentSkillPoints: s.progression.UnspentSkillPoints,
		Skills:             skills,
	}
}

// SkillCooldownViews returns active server-owned skill cooldowns.
func (s *Sim) SkillCooldownViews() []SkillCooldownView {
	if len(s.skillCooldowns) == 0 {
		return []SkillCooldownView{}
	}
	out := []SkillCooldownView{}
	for _, skillID := range sortedStringKeys(s.skillCooldowns) {
		if view, ok := s.skillCooldownView(skillID); ok {
			out = append(out, view)
		}
	}
	return out
}

func (s *Sim) SkillBindingsView() SkillBindingsView {
	return SkillBindingsView{
		FunctionKeys:      normalizeSkillFunctionKeys(s.skillFunctionKeys),
		RightClickSkillID: s.rightClickSkillID,
	}
}

// ProgressionState returns a copy of the mutable progression state.
func (s *Sim) ProgressionState() CharacterProgressionState {
	s.progression.Gold = s.gold
	out := s.progression
	out.SkillRanks = cloneIntMap(s.progression.SkillRanks)
	return out
}

func (s *Sim) DerivedStatsView() DerivedStatsView {
	effective, _ := s.playerEffectiveCombatStats()
	character := s.characterDerivedStatsView()
	return DerivedStatsView{
		DamageMin:            effective.DamageMin,
		DamageMax:            effective.DamageMax,
		Armor:                effective.Armor,
		AttackSpeed:          effective.AttackSpeed,
		AttackIntervalTicks:  effective.AttackIntervalTicks,
		HitChance:            effective.HitChance,
		CritChance:           effective.CritChance,
		CritDamage:           effective.CritDamage,
		MovementSpeed:        character.MovementSpeed,
		MaxHP:                effective.MaxHP,
		MaxMana:              effective.MaxMana,
		HealthRegenPerSecond: effective.HealthRegenPerSecond,
		ManaRegenPerSecond:   effective.ManaRegenPerSecond,
	}
}

func (s *Sim) characterDerivedStatsView() DerivedStatsView {
	stats := s.effectiveBaseStatsView()
	eval := func(key string) float64 {
		formula := s.rules.CharacterProgression.DerivedStats[key]
		return evalProgressionFormula(formula, stats)
	}
	return DerivedStatsView{
		DamageMin:            eval("damage_min"),
		DamageMax:            eval("damage_max"),
		Armor:                eval("armor"),
		AttackSpeed:          eval("attack_speed"),
		AttackIntervalTicks:  s.attackIntervalTicksFromSpeed(eval("attack_speed")),
		HitChance:            eval("hit_chance"),
		CritChance:           eval("crit_chance"),
		CritDamage:           eval("crit_damage"),
		MovementSpeed:        eval("movement_speed"),
		MaxHP:                eval("max_hp"),
		MaxMana:              eval("max_mana"),
		HealthRegenPerSecond: eval("health_regen_per_second"),
		ManaRegenPerSecond:   eval("mana_regen_per_second"),
	}
}

func (s *Sim) effectiveBaseStatsView() BaseStatsView {
	stats := s.progression.BaseStats
	if len(s.skillEffects) == 0 {
		return stats
	}
	for _, skillID := range sortedStringKeys(s.skillEffects) {
		effect := s.skillEffects[skillID]
		if effect.EndsTick <= s.tick || effect.Percent <= 0 {
			continue
		}
		for _, stat := range effect.Stats {
			switch stat {
			case "str":
				stats.Str = scaleStatPercent(stats.Str, effect.Percent)
			case "dex":
				stats.Dex = scaleStatPercent(stats.Dex, effect.Percent)
			case "vit":
				stats.Vit = scaleStatPercent(stats.Vit, effect.Percent)
			case "magic":
				stats.Magic = scaleStatPercent(stats.Magic, effect.Percent)
			}
		}
	}
	return stats
}

func scaleStatPercent(value int, percent int) int {
	if value <= 0 || percent <= 0 {
		return value
	}
	return int(math.Round(float64(value) * (1.0 + float64(percent)/100.0)))
}

func evalProgressionFormula(formula LinearStatFormula, stats BaseStatsView) float64 {
	value := formula.Base
	switch formula.Type {
	case "logarithmic":
		statValue := progressionStatValue(stats, formula.Stat)
		gainInput := statValue - formula.Offset
		if gainInput < 0 {
			gainInput = 0
		}
		denominator := formula.Denominator
		if denominator <= 0 {
			denominator = 1
		}
		value += formula.Scale * (math.Log1p(gainInput) / math.Log1p(denominator))
	default:
		value += formula.PerStr*float64(stats.Str) +
			formula.PerDex*float64(stats.Dex) +
			formula.PerVit*float64(stats.Vit) +
			formula.PerMagic*float64(stats.Magic)
	}
	if formula.Min != nil && value < *formula.Min {
		value = *formula.Min
	}
	if formula.Max != nil && value > *formula.Max {
		value = *formula.Max
	}
	return value
}

func progressionStatValue(stats BaseStatsView, stat string) float64 {
	switch stat {
	case "str":
		return float64(stats.Str)
	case "dex":
		return float64(stats.Dex)
	case "vit":
		return float64(stats.Vit)
	case "magic":
		return float64(stats.Magic)
	default:
		return 0
	}
}

func (s *Sim) StatBreakdownViews() []StatBreakdownView {
	_, breakdowns := s.playerEffectiveCombatStats()
	return breakdowns
}

func (s *Sim) requirementStatus(requirements map[string]int) ([]RequirementStatusView, bool) {
	if len(requirements) == 0 {
		return nil, true
	}
	status := []RequirementStatusView{}
	allMet := true
	for _, stat := range requirementStatOrder() {
		required := requirements[stat]
		if required <= 0 {
			continue
		}
		current := s.requirementCurrentValue(stat)
		met := current >= required
		if !met {
			allMet = false
		}
		status = append(status, RequirementStatusView{
			Stat:     stat,
			Required: required,
			Current:  current,
			Met:      met,
		})
	}
	return status, allMet
}

func requirementStatOrder() []string {
	return []string{"level", "str", "dex", "vit", "magic"}
}

func (s *Sim) requirementCurrentValue(stat string) int {
	switch stat {
	case "level":
		return s.progression.Level
	case "str":
		return s.progression.BaseStats.Str
	case "dex":
		return s.progression.BaseStats.Dex
	case "vit":
		return s.progression.BaseStats.Vit
	case "magic":
		return s.progression.BaseStats.Magic
	default:
		return 0
	}
}

func (s *Sim) requirementsMet(requirements map[string]int) bool {
	_, met := s.requirementStatus(requirements)
	return met
}

func (s *Sim) skillRequirementsMet(def SkillDef, rank int) bool {
	if !s.requirementsMet(skillRequirementsForRank(def.Requirements, rank)) {
		return false
	}
	for _, prereq := range def.Requirements.Skills {
		if s.progression.SkillRanks[prereq.SkillID] < prereq.Rank {
			return false
		}
	}
	return true
}

func skillRequirementsForRank(req SkillRequirementDef, rank int) map[string]int {
	if rank < 1 {
		rank = 1
	}
	rankOffset := rank - 1
	out := map[string]int{}
	level := req.Level + req.LevelPerRank*rankOffset
	if level > 0 {
		out["level"] = level
	}
	for _, stat := range []string{"str", "dex", "vit", "magic"} {
		if required := req.Stats[stat] + req.StatsPerRank[stat]*rankOffset; required > 0 {
			out[stat] = required
		}
	}
	return out
}

func (s *Sim) annotateRequirementStatus(requirements map[string]int, set func([]RequirementStatusView, *bool)) {
	status, met := s.requirementStatus(requirements)
	if len(status) == 0 {
		return
	}
	metCopy := met
	set(status, &metCopy)
}

func (s *Sim) itemView(item *invItem) ItemView {
	view := item.view()
	s.annotateItemView(&view, item)
	return view
}

func (s *Sim) stashItemView(item *stashItem) StashItemView {
	if item == nil {
		return StashItemView{}
	}
	view := item.view()
	s.annotateRequirementStatus(view.Requirements, func(status []RequirementStatusView, met *bool) {
		view.RequirementStatus = status
		view.RequirementsMet = met
	})
	if previewItem := item.previewItem(); previewItem != nil {
		if preview := s.equipPreviewForItem(previewItem, ""); preview != nil {
			view.EquipPreview = preview
		}
	}
	return view
}

func (s *Sim) stashItemViews() []StashItemView {
	out := make([]StashItemView, 0, len(s.stashItems))
	for _, item := range s.stashItems {
		if item == nil {
			continue
		}
		out = append(out, s.stashItemView(item))
	}
	return out
}

func (s *Sim) annotateItemView(view *ItemView, item *invItem) {
	if item == nil {
		return
	}
	s.annotateRequirementStatus(view.Requirements, func(status []RequirementStatusView, met *bool) {
		view.RequirementStatus = status
		view.RequirementsMet = met
	})
	if preview := s.equipPreviewForItem(item, view.Slot); preview != nil {
		view.EquipPreview = preview
	}
}

func (s *Sim) entityView(e *entity) EntityView {
	if e == nil {
		return EntityView{}
	}
	view := e.view()
	if e.kind != lootEntity {
		return view
	}
	s.annotateRequirementStatus(view.Requirements, func(status []RequirementStatusView, met *bool) {
		view.RequirementStatus = status
		view.RequirementsMet = met
	})
	if preview := s.equipPreviewForLoot(e); preview != nil {
		view.EquipPreview = preview
	}
	return view
}

func (s *Sim) equipPreviewForLoot(e *entity) *EquipPreviewView {
	if e == nil || e.kind != lootEntity {
		return nil
	}
	if e.rollPayload != nil {
		template, ok := s.rules.ItemTemplates[e.rollPayload.ItemTemplateID]
		if !ok {
			return nil
		}
		item := &invItem{
			instanceID:  previewItemInstanceID(),
			itemDefID:   e.rollPayload.ItemTemplateID,
			rollPayload: cloneRollPayload(e.rollPayload),
		}
		return s.equipPreviewForItemWithSlot(item, template.Slot)
	}
	def, ok := s.rules.Items[e.itemDefID]
	if !ok || !def.Equippable {
		return nil
	}
	item := &invItem{instanceID: previewItemInstanceID(), itemDefID: e.itemDefID}
	return s.equipPreviewForItemWithSlot(item, def.Slot)
}

func (s *Sim) equipPreviewForItem(item *invItem, currentSlot string) *EquipPreviewView {
	if item == nil {
		return nil
	}
	slot, ok := s.itemEquipSlot(item)
	if !ok {
		return nil
	}
	if currentSlot != "" && currentSlot != slot && !slotAcceptsItemSlot(currentSlot, slot) {
		return nil
	}
	return s.equipPreviewForItemWithSlot(item, s.comparisonSlot(slot))
}

func (s *Sim) equipPreviewForItemWithSlot(item *invItem, slot string) *EquipPreviewView {
	if item == nil || slot == "" {
		return nil
	}
	requirements := map[string]int{}
	if item.rollPayload != nil {
		requirements = item.rollPayload.Requirements
	}
	requirementsMet := s.requirementsMet(requirements)
	current, _ := s.playerEffectiveCombatStats()
	preview := s.previewEffectiveCombatStats(item, slot)
	deltas := equipPreviewDeltas(current, preview)
	if len(deltas) == 0 && len(requirements) == 0 {
		return nil
	}
	return &EquipPreviewView{Slot: slot, RequirementsMet: requirementsMet, Deltas: deltas}
}

func (s *Sim) previewEffectiveCombatStats(item *invItem, slot string) effectiveCombatStats {
	loadout := s.currentEquippedItems()
	for _, clearedSlot := range s.slotsClearedByEquip(slot, item) {
		delete(loadout, clearedSlot)
	}
	loadout[slot] = item
	stats, _ := s.playerEffectiveCombatStatsFor(loadout)
	return stats
}

func previewItemInstanceID() uint64 {
	return ^uint64(0)
}

func equipPreviewDeltas(current, preview effectiveCombatStats) []EquipPreviewDeltaView {
	values := []struct {
		stat    string
		current float64
		preview float64
	}{
		{"damage_min", current.DamageMin, preview.DamageMin},
		{"damage_max", current.DamageMax, preview.DamageMax},
		{"armor", current.Armor, preview.Armor},
		{"block_percent", current.BlockPercent, preview.BlockPercent},
		{"attack_speed", current.AttackSpeed, preview.AttackSpeed},
		{"attack_interval_ticks", float64(current.AttackIntervalTicks), float64(preview.AttackIntervalTicks)},
		{"max_hp", current.MaxHP, preview.MaxHP},
		{"max_mana", current.MaxMana, preview.MaxMana},
		{"health_regen_per_second", current.HealthRegenPerSecond, preview.HealthRegenPerSecond},
		{"mana_regen_per_second", current.ManaRegenPerSecond, preview.ManaRegenPerSecond},
	}
	deltas := []EquipPreviewDeltaView{}
	for _, value := range values {
		delta := value.preview - value.current
		if math.Abs(delta) < 0.000001 {
			continue
		}
		deltas = append(deltas, EquipPreviewDeltaView{
			Stat:    value.stat,
			Current: value.current,
			Preview: value.preview,
			Delta:   delta,
		})
	}
	return deltas
}

func (s *Sim) currentEquippedItems() map[string]*invItem {
	out := make(map[string]*invItem, len(equipmentSlots))
	for _, slot := range equipmentSlots {
		if item := s.findItemByID(s.equipped[slot]); item != nil {
			out[slot] = item
		}
	}
	return out
}

func (s *Sim) playerEffectiveCombatStats() (effectiveCombatStats, []StatBreakdownView) {
	return s.playerEffectiveCombatStatsFor(s.currentEquippedItems())
}

func (s *Sim) playerEffectiveCombatStatsFor(equippedItems map[string]*invItem) (effectiveCombatStats, []StatBreakdownView) {
	character := s.characterDerivedStatsView()
	damageMin := float64(s.rules.Combat.PlayerDamage.Min) + character.DamageMin
	damageMax := float64(s.rules.Combat.PlayerDamage.Max) + character.DamageMax
	armor := character.Armor
	maxHP := character.MaxHP
	healthRegen := character.HealthRegenPerSecond
	manaRegen := character.ManaRegenPerSecond
	blockPercent := 0.0
	weaponSpeed := 1.0
	itemSpeedPercent := 0.0

	damageMinSources := []StatBreakdownSourceView{
		{Label: "Base damage", Value: float64(s.rules.Combat.PlayerDamage.Min), Kind: "character_formula"},
		{Label: "Strength", Value: character.DamageMin, Kind: "character_formula"},
	}
	damageMaxSources := []StatBreakdownSourceView{
		{Label: "Base damage", Value: float64(s.rules.Combat.PlayerDamage.Max), Kind: "character_formula"},
		{Label: "Strength", Value: character.DamageMax, Kind: "character_formula"},
	}
	armorSources := []StatBreakdownSourceView{{Label: "Dexterity", Value: character.Armor, Kind: "character_formula"}}
	maxHPSources := []StatBreakdownSourceView{{Label: "Vitality", Value: character.MaxHP, Kind: "character_formula"}}
	healthRegenSources := []StatBreakdownSourceView{{Label: "Vitality", Value: character.HealthRegenPerSecond, Kind: "character_formula"}}
	manaRegenSources := []StatBreakdownSourceView{{Label: "Magic", Value: character.ManaRegenPerSecond, Kind: "character_formula"}}
	blockSources := []StatBreakdownSourceView{}
	attackSpeedSources := []StatBreakdownSourceView{{Label: "Dexterity", Value: character.AttackSpeed, Kind: "character_formula"}}

	if weapon := equippedItems[mainHandSlot]; weapon != nil {
		baseMin, baseMax, minRoll, maxRoll, label, itemID, ok := s.weaponDamageContributions(weapon)
		if ok {
			damageMin = character.DamageMin + baseMin + minRoll
			damageMax = character.DamageMax + baseMax + maxRoll
			damageMinSources = []StatBreakdownSourceView{
				{Label: label, Value: baseMin, Kind: "equipment_base", ItemInstanceID: itemID},
				{Label: "Rolled damage", Value: minRoll, Kind: "equipment_roll", ItemInstanceID: itemID},
				{Label: "Strength", Value: character.DamageMin, Kind: "character_formula"},
			}
			damageMaxSources = []StatBreakdownSourceView{
				{Label: label, Value: baseMax, Kind: "equipment_base", ItemInstanceID: itemID},
				{Label: "Rolled damage", Value: maxRoll, Kind: "equipment_roll", ItemInstanceID: itemID},
				{Label: "Strength", Value: character.DamageMax, Kind: "character_formula"},
			}
		}
		if speed, label, itemID, ok := s.weaponAttackSpeedContribution(weapon); ok {
			weaponSpeed = speed
			attackSpeedSources = append(attackSpeedSources, StatBreakdownSourceView{Label: label, Value: speed, Kind: "equipment_base", ItemInstanceID: itemID})
		}
	}

	for _, slot := range equipmentSlots {
		item := equippedItems[slot]
		if item == nil {
			continue
		}
		label := s.itemDisplayName(item)
		itemID := idStr(item.instanceID)
		baseStats, rolledStats := s.itemBaseAndRollStats(item)
		if value := baseStats["armor"]; value != 0 {
			armor += float64(value)
			armorSources = append(armorSources, StatBreakdownSourceView{Label: label, Value: float64(value), Kind: "equipment_base", ItemInstanceID: itemID})
		}
		if value := rolledStats["armor"]; value != 0 {
			armor += float64(value)
			armorSources = append(armorSources, StatBreakdownSourceView{Label: "Rolled armor", Value: float64(value), Kind: "equipment_roll", ItemInstanceID: itemID})
		}
		if value := baseStats["max_hp"]; value != 0 {
			maxHP += float64(value)
			maxHPSources = append(maxHPSources, StatBreakdownSourceView{Label: label, Value: float64(value), Kind: "equipment_base", ItemInstanceID: itemID})
		}
		if value := rolledStats["max_hp"]; value != 0 {
			maxHP += float64(value)
			maxHPSources = append(maxHPSources, StatBreakdownSourceView{Label: "Rolled max HP", Value: float64(value), Kind: "equipment_roll", ItemInstanceID: itemID})
		}
		if value := baseStats["health_regen_per_10_seconds"]; value != 0 {
			perSecond := float64(value) / 10.0
			healthRegen += perSecond
			healthRegenSources = append(healthRegenSources, StatBreakdownSourceView{Label: label, Value: perSecond, Kind: "equipment_base", ItemInstanceID: itemID})
		}
		if value := rolledStats["health_regen_per_10_seconds"]; value != 0 {
			perSecond := float64(value) / 10.0
			healthRegen += perSecond
			healthRegenSources = append(healthRegenSources, StatBreakdownSourceView{Label: "Rolled HP regen", Value: perSecond, Kind: "equipment_roll", ItemInstanceID: itemID})
		}
		if value := baseStats["mana_regen_per_10_seconds"]; value != 0 {
			perSecond := float64(value) / 10.0
			manaRegen += perSecond
			manaRegenSources = append(manaRegenSources, StatBreakdownSourceView{Label: label, Value: perSecond, Kind: "equipment_base", ItemInstanceID: itemID})
		}
		if value := rolledStats["mana_regen_per_10_seconds"]; value != 0 {
			perSecond := float64(value) / 10.0
			manaRegen += perSecond
			manaRegenSources = append(manaRegenSources, StatBreakdownSourceView{Label: "Rolled mana regen", Value: perSecond, Kind: "equipment_roll", ItemInstanceID: itemID})
		}
		if value := baseStats["block_percent"]; value != 0 {
			blockPercent += float64(value)
			blockSources = append(blockSources, StatBreakdownSourceView{Label: label, Value: float64(value), Kind: "equipment_base", ItemInstanceID: itemID})
		}
		if value := rolledStats["block_percent"]; value != 0 {
			blockPercent += float64(value)
			blockSources = append(blockSources, StatBreakdownSourceView{Label: "Rolled block", Value: float64(value), Kind: "equipment_roll", ItemInstanceID: itemID})
		}
		if value := baseStats["attack_speed_percent"]; value != 0 {
			itemSpeedPercent += float64(value)
			attackSpeedSources = append(attackSpeedSources, StatBreakdownSourceView{Label: label, Value: float64(value), Kind: "equipment_base", ItemInstanceID: itemID})
		}
		if value := rolledStats["attack_speed_percent"]; value != 0 {
			itemSpeedPercent += float64(value)
			attackSpeedSources = append(attackSpeedSources, StatBreakdownSourceView{Label: "Rolled attack speed", Value: float64(value), Kind: "equipment_roll", ItemInstanceID: itemID})
		}
	}

	uncappedBlock := blockPercent
	blockCap := float64(s.rules.Combat.BlockCap)
	if blockPercent > blockCap {
		blockPercent = blockCap
		blockSources = append(blockSources, StatBreakdownSourceView{Label: "Block cap", Value: blockPercent - uncappedBlock, Kind: "cap"})
	}
	uncappedAttackSpeed := character.AttackSpeed * weaponSpeed * (1 + itemSpeedPercent/100.0)
	attackSpeed := s.clampEffectiveAttackSpeed(uncappedAttackSpeed)
	if math.Abs(attackSpeed-uncappedAttackSpeed) > 0.000001 {
		attackSpeedSources = append(attackSpeedSources, StatBreakdownSourceView{Label: "Attack speed clamp", Value: attackSpeed - uncappedAttackSpeed, Kind: "cap"})
	}
	attackInterval := s.attackIntervalTicksFromSpeed(attackSpeed)
	attackIntervalSources := []StatBreakdownSourceView{
		{Label: "Base attack interval", Value: float64(s.rules.Combat.BaseAttackIntervalTicks), Kind: "combat_rule"},
		{Label: "Effective attack speed", Value: attackSpeed, Kind: "derived"},
	}

	effective := effectiveCombatStats{
		DamageMin:            maxFloat(0, damageMin),
		DamageMax:            maxFloat(0, damageMax),
		HitChance:            clampFloat(character.HitChance, 0, 1),
		CritChance:           clampFloat(character.CritChance, 0, 1),
		CritDamage:           maxFloat(1, character.CritDamage),
		Armor:                maxFloat(0, armor),
		BlockPercent:         maxFloat(0, blockPercent),
		AttackSpeed:          attackSpeed,
		AttackIntervalTicks:  attackInterval,
		MaxHP:                maxFloat(1, maxHP),
		MaxMana:              maxFloat(0, character.MaxMana),
		HealthRegenPerSecond: maxFloat(0, healthRegen),
		ManaRegenPerSecond:   maxFloat(0, manaRegen),
	}
	if effective.DamageMax < effective.DamageMin {
		effective.DamageMax = effective.DamageMin
	}

	breakdowns := []StatBreakdownView{
		{Key: "damage_min", Value: effective.DamageMin, UncappedValue: effective.DamageMin, Cap: nil, Sources: damageMinSources},
		{Key: "damage_max", Value: effective.DamageMax, UncappedValue: effective.DamageMax, Cap: nil, Sources: damageMaxSources},
		{Key: "armor", Value: effective.Armor, UncappedValue: effective.Armor, Cap: nil, Sources: armorSources},
		{Key: "attack_speed", Value: effective.AttackSpeed, UncappedValue: uncappedAttackSpeed, Cap: floatPtr(s.rules.Combat.MaxEffectiveAttackSpeed), Sources: attackSpeedSources},
		{Key: "attack_interval_ticks", Value: float64(effective.AttackIntervalTicks), UncappedValue: float64(effective.AttackIntervalTicks), Cap: nil, Sources: attackIntervalSources},
		{Key: "max_hp", Value: effective.MaxHP, UncappedValue: effective.MaxHP, Cap: nil, Sources: maxHPSources},
		{Key: "health_regen_per_second", Value: effective.HealthRegenPerSecond, UncappedValue: effective.HealthRegenPerSecond, Cap: nil, Sources: healthRegenSources},
		{Key: "mana_regen_per_second", Value: effective.ManaRegenPerSecond, UncappedValue: effective.ManaRegenPerSecond, Cap: nil, Sources: manaRegenSources},
		{Key: "block_percent", Value: effective.BlockPercent, UncappedValue: uncappedBlock, Cap: floatPtr(blockCap), Sources: blockSources},
	}
	return effective, breakdowns
}

func (s *Sim) weaponDamageContributions(item *invItem) (baseMin, baseMax, rollMin, rollMax float64, label, itemID string, ok bool) {
	itemID = idStr(item.instanceID)
	label = s.itemDisplayName(item)
	if item.rollPayload != nil {
		template, found := s.rules.ItemTemplates[item.rollPayload.ItemTemplateID]
		if !found {
			return 0, 0, 0, 0, "", "", false
		}
		totalMin, minOK := item.rollPayload.Stats["damage_min"]
		totalMax, maxOK := item.rollPayload.Stats["damage_max"]
		if !minOK || !maxOK || totalMax < totalMin {
			return 0, 0, 0, 0, "", "", false
		}
		baseMinInt := template.BaseStats["damage_min"]
		baseMaxInt := template.BaseStats["damage_max"]
		return float64(baseMinInt), float64(baseMaxInt), float64(totalMin - baseMinInt), float64(totalMax - baseMaxInt), label, itemID, true
	}
	def, found := s.rules.Items[item.itemDefID]
	if !found || def.Damage == nil {
		return 0, 0, 0, 0, "", "", false
	}
	return float64(def.Damage.Min), float64(def.Damage.Max), 0, 0, label, itemID, true
}

func (s *Sim) weaponAttackSpeedContribution(item *invItem) (speed float64, label, itemID string, ok bool) {
	if item == nil {
		return 0, "", "", false
	}
	itemID = idStr(item.instanceID)
	label = s.itemDisplayName(item)
	if item.rollPayload != nil {
		template, found := s.rules.ItemTemplates[item.rollPayload.ItemTemplateID]
		if !found || template.AttackSpeed <= 0 {
			return 0, "", "", false
		}
		return template.AttackSpeed, label, itemID, true
	}
	def, found := s.rules.Items[item.itemDefID]
	if !found || def.AttackSpeed <= 0 {
		return 0, "", "", false
	}
	return def.AttackSpeed, label, itemID, true
}

func (s *Sim) clampEffectiveAttackSpeed(speed float64) float64 {
	minSpeed := s.rules.Combat.MinEffectiveAttackSpeed
	maxSpeed := s.rules.Combat.MaxEffectiveAttackSpeed
	if minSpeed <= 0 {
		minSpeed = 0.25
	}
	if maxSpeed < minSpeed {
		maxSpeed = minSpeed
	}
	return clampFloat(speed, minSpeed, maxSpeed)
}

func (s *Sim) attackIntervalTicksFromSpeed(speed float64) int {
	speed = s.clampEffectiveAttackSpeed(speed)
	if speed <= 0 {
		speed = 1
	}
	interval := int(math.Ceil(float64(s.rules.Combat.BaseAttackIntervalTicks) / speed))
	if interval < 1 {
		return 1
	}
	return interval
}

func (s *Sim) itemBaseAndRollStats(item *invItem) (map[string]int, map[string]int) {
	baseStats := map[string]int{}
	rolledStats := map[string]int{}
	if item == nil || item.rollPayload == nil {
		return baseStats, rolledStats
	}
	template, ok := s.rules.ItemTemplates[item.rollPayload.ItemTemplateID]
	if !ok {
		return baseStats, rolledStats
	}
	for key, value := range template.BaseStats { //nolint:determinism — output is a map, order irrelevant
		baseStats[key] = value
	}
	for key, total := range item.rollPayload.Stats { //nolint:determinism — output is a map, order irrelevant
		if base := template.BaseStats[key]; total != base {
			rolledStats[key] = total - base
		}
	}
	return baseStats, rolledStats
}

func (s *Sim) itemDisplayName(item *invItem) string {
	if item == nil {
		return "Item"
	}
	if item.rollPayload != nil {
		if item.rollPayload.DisplayName != "" {
			return item.rollPayload.DisplayName
		}
		if template, ok := s.rules.ItemTemplates[item.rollPayload.ItemTemplateID]; ok {
			return template.Name
		}
	}
	if def, ok := s.rules.Items[item.itemDefID]; ok {
		return def.Name
	}
	return item.itemDefID
}

func (s *Sim) monsterEffectiveCombatStats(monster *entity, damage DamageRange) effectiveCombatStats {
	if monster == nil {
		return effectiveCombatStats{
			HitChance:  s.rules.Combat.BaseHitChance,
			CritDamage: s.rules.Combat.BaseCritDamage,
			MaxHP:      1,
		}
	}
	def := s.rules.Monsters[monster.monsterDefID]
	return effectiveCombatStats{
		DamageMin:    float64(damage.Min),
		DamageMax:    float64(damage.Max),
		HitChance:    def.effectiveHitChance(s.rules.Combat),
		CritChance:   def.effectiveCritChance(s.rules.Combat),
		CritDamage:   def.effectiveCritDamage(s.rules.Combat),
		Armor:        float64(def.Armor),
		BlockPercent: clampFloat(float64(def.BlockPercent), 0, float64(s.rules.Combat.BlockCap)),
		MaxHP:        float64(monster.maxHP),
	}
}

func (s *Sim) experienceToNextLevel() *int {
	nextXP, ok := s.rules.nextLevelTotalXP(s.progression.Level)
	if !ok {
		return nil
	}
	remaining := nextXP - s.progression.Experience
	if remaining < 0 {
		remaining = 0
	}
	return &remaining
}

func (r *Rules) nextLevelTotalXP(level int) (int, bool) {
	nextXP, ok := r.CharacterProgression.XPThresholds[level]
	return nextXP, ok
}

func (r *Rules) skillPointsGrantedAtLevel(level int) int {
	cadence := r.CharacterProgression.SkillPoints
	if cadence.PointsPerGrant <= 0 || cadence.GrantEveryLevels <= 0 || level < cadence.FirstGrantLevel {
		return 0
	}
	if (level-cadence.FirstGrantLevel)%cadence.GrantEveryLevels != 0 {
		return 0
	}
	return cadence.PointsPerGrant
}

func (s *Sim) playerProjectileInFlight() bool {
	for _, e := range s.activeLevel().entities {
		if e.kind == projectileEntity && e.ownerID == s.playerID {
			return true
		}
	}
	return false
}

func (s *Sim) rollRange(d DamageRange) int {
	span := d.Max - d.Min + 1
	if span <= 1 {
		return d.Min
	}
	return d.Min + s.rng.IntN(span)
}

func (s *Sim) playerDead() bool {
	player := s.activeLevel().entities[s.playerID]
	return player == nil || player.hp <= 0
}

func (s *Sim) findEntity(id string) *entity {
	n, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil
	}
	return s.activeLevel().entities[n]
}

func (s *Sim) findEntityAnyLevel(id string) (*entity, int, bool) {
	n, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, 0, false
	}
	for _, levelNum := range s.sortedLevelNums() {
		level := s.levels[levelNum]
		if level == nil {
			continue
		}
		if e := level.entities[n]; e != nil {
			return e, levelNum, true
		}
	}
	return nil, 0, false
}

func (s *Sim) findReachableStair(level *LevelState, defID string, playerPos Vec2) *entity {
	stair := s.findStair(level, defID)
	if stair == nil {
		return nil
	}
	if meleeInRange(distance(playerPos, stair.pos), s.rules.Combat.UnarmedReach, interactableInteractionRadius) {
		return stair
	}
	return nil
}

func (s *Sim) findReachableTeleporter(level *LevelState, playerPos Vec2) *entity {
	teleporter := s.findTeleporter(level)
	if teleporter == nil {
		return nil
	}
	if meleeInRange(distance(playerPos, teleporter.pos), s.rules.Combat.UnarmedReach, interactableInteractionRadius) {
		return teleporter
	}
	return nil
}

func (s *Sim) findStair(level *LevelState, defID string) *entity {
	for _, id := range sortedEntityIDs(level.entities) {
		e := level.entities[id]
		if e != nil && e.kind == interactableEntity && e.interactableDefID == defID {
			return e
		}
	}
	return nil
}

func (s *Sim) findTeleporter(level *LevelState) *entity {
	for _, id := range sortedEntityIDs(level.entities) {
		e := level.entities[id]
		if e != nil && e.kind == interactableEntity && e.interactableDefID == teleporterDefID {
			return e
		}
	}
	return nil
}

func (s *Sim) findItem(instanceID string) *invItem {
	n, err := strconv.ParseUint(instanceID, 10, 64)
	if err != nil {
		return nil
	}
	return s.findItemByID(n)
}

func (s *Sim) findItemByID(id uint64) *invItem {
	for _, it := range s.inventory {
		if it.instanceID == id {
			return it
		}
	}
	return nil
}

func (s *Sim) findStashItem(stashItemID string) *stashItem {
	n, err := strconv.ParseUint(stashItemID, 10, 64)
	if err != nil {
		return nil
	}
	for _, it := range s.stashItems {
		if it == nil {
			continue
		}
		if it.stashItemID == n {
			return it
		}
	}
	return nil
}

func sortedEntityIDs(entities map[uint64]*entity) []uint64 {
	ids := make([]uint64, 0, len(entities))
	for id := range entities {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func (s *Sim) sortedLevelNums() []int {
	levels := make([]int, 0, len(s.levels))
	for levelNum := range s.levels {
		levels = append(levels, levelNum)
	}
	sort.Ints(levels)
	return levels
}

// Snapshot returns the default player's authoritative state, with entities
// ordered by id. Solo callers keep using this compatibility path.
func (s *Sim) Snapshot() Snapshot {
	if ps := s.defaultPlayer(); ps != nil {
		return s.SnapshotForPlayer(ps.PlayerID)
	}
	return Snapshot{
		ServerTick:        s.tick,
		SessionID:         s.sessionID,
		Seed:              s.seed,
		CurrentLevel:      s.currentLevel,
		Walls:             []WallView{},
		Entities:          []EntityView{},
		Inventory:         []ItemView{},
		Equipped:          newSnapshotEquippedMap(newEquippedMap()),
		Hotbar:            []HotbarSlotView{},
		InventoryRows:     baseInventoryRows,
		InventoryCapacity: inventoryCapacityForRows(baseInventoryRows),
		Gold:              0,
		StashItems:        []StashItemView{},
		StashGold:         0,
		StashCapacity:     defaultStashCapacity,
		SkillProgression:  SkillProgressionView{Skills: []SkillProgressionSkillView{}},
		SkillCooldowns:    []SkillCooldownView{},
		SkillBindings:     SkillBindingsView{FunctionKeys: make([]string, skillFunctionKeyCount)},
		RecentEvents:      []Event{},
	}
}

// SnapshotForPlayer returns a recipient-scoped snapshot. Entities are limited
// to the player's current level; inventory/progression fields belong only to
// the receiving player.
func (s *Sim) SnapshotForPlayer(playerID uint64) Snapshot {
	ps := s.players[playerID]
	if ps == nil {
		return s.Snapshot()
	}
	s.usePlayer(ps)
	ids := sortedEntityIDs(s.activeLevel().entities)

	entities := make([]EntityView, 0, len(ids))
	for _, id := range ids {
		entities = append(entities, s.entityView(s.activeLevel().entities[id]))
	}

	inventory := make([]ItemView, 0, len(s.inventory))
	for _, it := range s.inventory {
		if it == nil || s.hotbarHasItem(it.instanceID) {
			continue
		}
		inventory = append(inventory, s.itemView(it))
	}
	stashItems := make([]StashItemView, 0, len(s.stashItems))
	for _, it := range s.stashItems {
		stashItems = append(stashItems, s.stashItemView(it))
	}

	equipped := newSnapshotEquippedMap(s.equipped)
	party := s.partyView()

	snap := Snapshot{
		ServerTick:            s.tick,
		SessionID:             s.sessionID,
		Seed:                  s.seed,
		CurrentLevel:          s.currentLevel,
		LocalPlayerID:         idStr(ps.PlayerID),
		Party:                 party,
		Walls:                 wallViewsForLevel(s.activeLevel()),
		Entities:              entities,
		Inventory:             inventory,
		Equipped:              equipped,
		HotbarCapacity:        s.hotbarCapacity(),
		Hotbar:                s.hotbarView(),
		InventoryRows:         s.inventoryRows(),
		InventoryCapacity:     s.inventoryCapacity(),
		Gold:                  s.gold,
		StashItems:            stashItems,
		StashGold:             s.stashGold,
		StashCapacity:         s.stashCapacity,
		DiscoveredTeleporters: s.teleporterDiscoveryView(),
		CharacterProgression:  s.CharacterProgressionView(),
		SkillProgression:      s.SkillProgressionView(),
		SkillCooldowns:        s.SkillCooldownViews(),
		SkillBindings:         s.SkillBindingsView(),
		RecentEvents:          []Event{},
	}
	s.savePlayer(ps)
	return snap
}

func wallViewsForLevel(level *LevelState) []WallView {
	if level == nil {
		return []WallView{}
	}
	out := make([]WallView, 0, len(level.walls))
	for i, wall := range level.walls {
		source := wall.source
		if source == "" {
			source = "preset"
		}
		out = append(out, WallView{
			ID:       wallID(level.levelNum, i),
			Position: wall.pos,
			Size:     wall.size,
			Source:   source,
		})
	}
	return out
}

func newSnapshotEquippedMap(equipped map[string]uint64) map[string]*string {
	out := make(map[string]*string, len(equipmentSlots))
	for _, slot := range equipmentSlots {
		instanceID := equipped[slot]
		if instanceID == 0 {
			out[slot] = nil
			continue
		}
		v := idStr(instanceID)
		out[slot] = &v
	}
	return out
}

func (s *Sim) partyView() []PartyMemberView {
	out := make([]PartyMemberView, 0, len(s.players))
	for _, playerID := range sortedPlayerIDs(s.players) {
		ps := s.players[playerID]
		if ps == nil {
			continue
		}
		out = append(out, PartyMemberView{
			PlayerID:     idStr(ps.PlayerID),
			CharacterID:  ps.CharacterID,
			DisplayName:  ps.DisplayName,
			Role:         ps.Role,
			Connected:    ps.Connected,
			CurrentLevel: ps.CurrentLevel,
		})
	}
	return out
}

func (s *Sim) teleporterDiscoveryView() []TeleporterDiscoveryView {
	if !s.multiLevel {
		return []TeleporterDiscoveryView{}
	}
	levelSet := make(map[int]bool, len(s.levels)+len(s.discoveredTeleporters))
	for levelNum := range s.levels {
		if s.levelHasTeleporter(levelNum) {
			levelSet[levelNum] = true
		}
	}
	for levelNum := range s.discoveredTeleporters {
		if s.levelHasTeleporter(levelNum) {
			levelSet[levelNum] = true
		}
	}
	levels := make([]int, 0, len(levelSet))
	for levelNum := range levelSet {
		levels = append(levels, levelNum)
	}
	sort.Ints(levels)
	out := make([]TeleporterDiscoveryView, 0, len(levels))
	for _, levelNum := range levels {
		out = append(out, TeleporterDiscoveryView{Level: levelNum, Discovered: s.discoveredTeleporters[levelNum]})
	}
	return out
}

func (s *Sim) levelHasTeleporter(levelNum int) bool {
	if !s.multiLevel {
		return false
	}
	if levelNum == townLevel {
		return true
	}
	return dungeonLevelHasTeleporter(levelNum)
}

func (e *entity) view() EntityView {
	ev := EntityView{ID: idStr(e.id), Type: e.kind, Position: e.pos}
	switch e.kind {
	case playerEntity, monsterEntity:
		hp, maxHP := e.hp, e.maxHP
		ev.HP = &hp
		ev.MaxHP = &maxHP
		if e.kind == playerEntity {
			mana, maxMana := e.mana, e.maxMana
			ev.Mana = &mana
			ev.MaxMana = &maxMana
			ev.CharacterID = e.characterID
			ev.DisplayName = e.displayName
			if e.visualScale > 0 {
				ev.VisualScale = e.visualScale
			}
		}
		if e.kind == monsterEntity {
			ev.MonsterDefID = e.monsterDefID
			if e.monsterRarityID != "" {
				ev.Rarity = e.monsterRarityID
			}
			ev.IsBoss = e.isBoss
			ev.BossTemplateID = e.bossTemplateID
			ev.VisualModel = e.visualModel
			ev.VisualScale = e.visualScale
			ev.VisualTint = e.visualTint
			if e.isBoss && e.bossPhaseKind != "" {
				ev.BossPhase = e.bossPhaseView()
			}
		}
	case lootEntity:
		ev.ItemDefID = e.itemDefID
		if e.goldAmount > 0 {
			amount := e.goldAmount
			ev.Amount = &amount
		}
		if e.rollPayload != nil {
			ev.ItemDefID = e.rollPayload.ItemTemplateID
			ev.ItemTemplateID = e.rollPayload.ItemTemplateID
			ev.DisplayName = e.rollPayload.DisplayName
			ev.Rarity = e.rollPayload.Rarity
			ev.RolledStats = cloneIntMap(e.rollPayload.Stats)
			ev.Requirements = cloneIntMap(e.rollPayload.Requirements)
			ev.EffectIDs = cloneStringSlice(e.rollPayload.EffectIDs)
		}
	case interactableEntity:
		ev.InteractableDefID = e.interactableDefID
		ev.State = e.state
	case projectileEntity:
		ev.OwnerID = idStr(e.ownerID)
		if e.targetID != 0 {
			ev.TargetID = idStr(e.targetID)
		}
		ev.ProjectileDefID = e.projectileDefID
	}
	return ev
}

func (e *entity) bossPhaseView() *BossPhaseView {
	return &BossPhaseView{
		PatternID:     e.bossPatternID,
		PhaseIndex:    e.bossPhaseIndex,
		PhaseKind:     e.bossPhaseKind,
		StartedTick:   e.bossPhaseStarted,
		DurationTicks: int(e.bossPhaseEnds - e.bossPhaseStarted),
	}
}

func (it *invItem) view() ItemView {
	v := ItemView{
		ItemInstanceID: idStr(it.instanceID),
		ItemDefID:      it.itemDefID,
		Slot:           it.slot,
		Equipped:       it.equipped,
	}
	if it.rollPayload != nil {
		it.rollPayload.itemViewFields(&v)
	}
	return v
}

func (it *stashItem) view() StashItemView {
	v := StashItemView{
		StashItemID: idStr(it.stashItemID),
		ItemDefID:   it.itemDefID,
	}
	if it.rollPayload != nil {
		it.rollPayload.stashItemViewFields(&v)
	}
	return v
}

func (it *stashItem) previewItem() *invItem {
	if it == nil {
		return nil
	}
	return &invItem{
		instanceID:  it.stashItemID,
		itemDefID:   it.itemDefID,
		rollPayload: it.rollPayload,
	}
}

func ptrEntityView(v EntityView) *EntityView { return &v }
func ptrItemView(v ItemView) *ItemView       { return &v }
func ptrStashItemView(v StashItemView) *StashItemView {
	return &v
}
func intPtr(v int) *int           { return &v }
func floatPtr(v float64) *float64 { return &v }
func boolPtr(v bool) *bool        { return &v }

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func clampFloat(v, minValue, maxValue float64) float64 {
	if v < minValue {
		return minValue
	}
	if v > maxValue {
		return maxValue
	}
	return v
}

func normalize(v Vec2) Vec2 {
	length := math.Hypot(v.X, v.Y)
	if length == 0 {
		return Vec2{}
	}
	return Vec2{X: v.X / length, Y: v.Y / length}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
