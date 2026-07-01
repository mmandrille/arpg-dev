package game

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	baseEntityID                   = 1001 // player=1001, monster=1002, loot=1003, item=1004 ...
	playerStartHP                  = 10
	defaultMoveSpeed               = 1.0
	simulationTickHz               = 10.0
	playerRadius                   = 0.45
	monsterRadius                  = 0.45
	monsterDefID                   = "training_dummy"
	playerEntity                   = "player"
	monsterEntity                  = "monster"
	companionEntity                = "companion"
	lootEntity                     = "loot"
	projectileEntity               = "projectile"
	wallEntity                     = "wall"
	interactableEntity             = "interactable"
	monsterBehaviorStatic          = "static"
	monsterBehaviorChase           = "chase"
	monsterAIModeIdle              = "idle"
	monsterAIModeChase             = "chase"
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
	uniqueTestChestService         = "unique_test_chest"
	townStashDefID                 = "town_stash"
	accountStashID                 = "account_stash"
	heroCorpseDefID                = "hero_corpse"
	worldModeMultiLevel            = "multi_level"
	attackModeMelee                = "melee"
	attackModeRanged               = "ranged"
	magicBoltSkillID               = "magic_bolt"
	trainingArrowProjectileDefID   = "training_arrow"
	staffOrbProjectileDefID        = "staff_orb"
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
	healRainPulseIntervalTicks     = 10
	minHotbarCapacity              = 2
	maxHotbarCapacity              = 10
	weaponSetCount                 = 2
	defaultWeaponSet               = 0
	skillFunctionKeyCount          = 16
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

const (
	DefaultWorldID            = "vertical_slice"
	townLevel                 = 0
	levelZero                 = 0
	townNavigationMarginCells = 8.0
)

type entity struct {
	id                    uint64
	kind                  string
	pos                   Vec2
	hp                    int
	maxHP                 int
	mana                  int
	maxMana               int
	characterID           string
	displayName           string
	monsterDefID          string
	monsterRarityID       string
	monsterPackID         string
	monsterPackLeader     bool
	monsterAttackDamage   *DamageRange
	monsterAttackCooldown int
	monsterArmor          float64
	monsterHitChance      float64
	monsterCritChance     float64
	monsterBlockPercent   float64
	monsterXPReward       int
	isBoss                bool
	bossTemplateID        string
	visualModel           string
	visualTint            string
	visualScale           float64
	bossPatternID         string
	bossPatternDeckIndex  int
	bossPhaseIndex        int
	bossPhaseKind         string
	bossPhaseStarted      uint64
	bossPhaseEnds         uint64
	bossCooldownEnds      uint64
	bossActiveHit         map[uint64]bool
	bossPhaseExecuted     bool
	bossPhaseAim          Vec2
	bossPhaseHasAim       bool
	bossEnraged           bool
	bossEnrageThreshold   float64
	itemDefID             string
	goldAmount            int
	rollPayload           *ItemRollPayload
	interactableDefID     string
	state                 string
	lootTable             string
	corpseCharacterID     string
	corpseName            string
	corpseLevel           int
	corpseItemCount       int
	ownerID               uint64
	targetID              uint64
	companionStance       string
	projectileDefID       string
	sourceSkillID         string
	expiresTick           uint64
	totalDurationTicks    int
	sourceDamageType      string
	shardProjectile       bool
	effectIDs             []string
	dir                   Vec2
	speed                 float64
	traveled              float64
	maxDistance           float64
	damageRange           DamageRange
	sourceMsgID           string
	sourceCorrID          string
	spawnTick             uint64
	spawnPos              Vec2
	aiMode                string
	aiTargetPlayerID      uint64
	navPath               []Vec2
	navGoal               Vec2
	navPathValid          bool
	navPathCell           gridCell
	navTargetPlayerID     uint64
	navPathTick           uint64
	navNextRepathTick     uint64
	lastAttackTick        uint64
	hasAttacked           bool
	attackWindupRemaining int
	attackWindupTargetID  uint64
	attackWindupDamage    DamageRange
	rangedMeleeEngagedTick uint64
}

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

type uniqueChestState struct {
	items []*stashItem
}

type goldRollContext struct {
	levelNum        int
	monsterRarityID string
	magicFind       bool
}

type activeMove struct {
	dir       Vec2
	remaining int
}

type autoNavState struct {
	steps              []Vec2
	goal               Vec2
	hasGoal            bool
	lastReplanPos      Vec2 // position where the last re-plan fired; zero before first re-plan
	pathStepsExhausted bool // set when the queued path was fully consumed without a blocked step
	replanAttempts     int
	pendingAction      *ActionIntent
	pendingSkill       *CastSkillIntent
	sourceMsgID        string
	sourceCorrID       string
}

type activeSkillChannel struct {
	skillID            string
	rank               int
	dir                Vec2
	correlationID      string
	manaPer10Seconds   int
	manaAccumulator    int
	impactedMonsterIDs map[uint64]bool
}

type wallObstacle struct {
	pos         Vec2
	size        Vec2
	source      string
	shapeFamily string
	kind        string
	blocksLOS   *bool
}

type effectiveCombatStats struct {
	DamageMin            float64
	DamageMax            float64
	HitChance            float64
	CritChance           float64
	CritDamage           float64
	EvadeChance          float64
	Armor                float64
	BlockPercent         float64
	AttackSpeed          float64
	AttackIntervalTicks  int
	MaxHP                float64
	MaxMana              float64
	HealthRegenPerSecond float64
	ManaRegenPerSecond   float64
	MagicFindPercent     float64
	LightRadius          float64
	MovementSpeedPercent float64
}

type combatResolution struct {
	Outcome         string
	Damage          int
	DamageType      string
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
	TargetID    uint64
	Stats       []string
	Percent     int
	VisualScale float64
	EffectID    string
	EndsTick    uint64
	TotalTicks  int
}

type areaHealZoneState struct {
	ID            uint64
	Level         int
	Center        Vec2
	CasterID      uint64
	SkillID       string
	Rank          int
	Percent       int
	Radius        float64
	IncludeCaster bool
	CorrelationID string
	NextPulseTick uint64
	EndsTick      uint64
}

type skillHealApplication struct {
	Target *entity
	Heal   int
}

type skillBuffApplication struct {
	Target      *entity
	Effect      SkillEffectDef
	Percent     int
	VisualScale float64
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
	ActiveChannel         *activeSkillChannel
	Inventory             []*invItem
	Equipped              map[string]uint64
	WeaponSets            []map[string]uint64
	ActiveWeaponSet       int
	Hotbar                []uint64
	DiscoveredTeleporters map[int]bool
	Progression           CharacterProgressionState
	SkillCooldowns        map[string]skillCooldownState
	SkillEffects          map[string]skillEffectState
	PoisonDots            map[uint64]poisonDotState
	BleedDots             map[uint64]bleedDotState
	RogueMarks            map[uint64]rogueMarkState
	UniqueBurnDots        map[string]uniqueBurnDotState
	UniqueExecutionMarks  map[uint64]uniqueExecutionMarkState
	UniqueHungerStacks    map[uint64]uniqueHungerStackState
	UniqueAshenReprisals  map[uint64]uniqueAshenReprisalState
	UniquePilgrimMomentum map[uint64]uniquePilgrimMomentumState
	UniqueChestItems      map[uint64][]*stashItem
	SkillFunctionKeys     []string
	RightClickSkillID     string
	ShopStock             map[string]*shopStockState
	Gold                  int
	StashItems            []*stashItem
	StashGold             int
	StashCapacity         int
	ResourceWallet        map[string]int
	HPRegenCarry          float64
	ManaRegenCarry        float64
	FogVisibleLevel       int
	VisibleMonsterIDs     map[uint64]bool
	NextBasicAttackTick   uint64
	NextOffHandAttackTick uint64
}

// Sim is the deterministic authoritative simulation for one solo session.
// Given the same seed and the same ordered inputs, it produces identical
// outputs (entity ids, events, final state) on every run (ADR-0001 D8.1).
type Sim struct {
	sessionID                   string
	seed                        string
	rng                         *RNG
	rules                       *Rules
	gameplayDebug               bool
	tickPerf                    PerfCounters
	tickProfiler                TickProfiler
	monsterPathRequestsThisTick int
	monsterPathNodesThisTick    int
	playerPathNodesThisTick     int
	overloadDegradeUntilTick    uint64
	combatMovementThrottled     bool
	tickCollisionCache          tickCollisionCache
	tick                        uint64
	nextID                      uint64
	playerID                    uint64
	players                     map[uint64]*playerState
	goldRoll                    uint64
	nextAreaHealZoneID          uint64

	levels                map[int]*LevelState
	currentLevel          int
	multiLevel            bool
	fogOfWarEnabled       bool
	entities              map[uint64]*entity
	walls                 []wallObstacle
	move                  *activeMove
	autoNav               *autoNavState
	inventory             []*invItem
	equipped              map[string]uint64 // slot -> instanceID (0 = none)
	weaponSets            []map[string]uint64
	activeWeaponSet       int
	hotbar                []uint64 // fixed 10-slot item instance assignments (0 = none)
	discoveredTeleporters map[int]bool
	progression           CharacterProgressionState
	skillCooldowns        map[string]skillCooldownState
	skillEffects          map[string]skillEffectState
	poisonDots            map[uint64]poisonDotState
	bleedDots             map[uint64]bleedDotState
	rogueMarks            map[uint64]rogueMarkState
	uniqueBurnDots        map[string]uniqueBurnDotState
	uniqueExecutionMarks  map[uint64]uniqueExecutionMarkState
	uniqueHungerStacks    map[uint64]uniqueHungerStackState
	uniqueAshenReprisals  map[uint64]uniqueAshenReprisalState
	uniquePilgrimMomentum map[uint64]uniquePilgrimMomentumState
	uniqueChests          map[uint64]*uniqueChestState
	areaHealZones         map[uint64]areaHealZoneState
	skillFunctionKeys     []string
	rightClickSkillID     string
	shopStock             map[string]*shopStockState
	gold                  int
	stashItems            []*stashItem
	stashGold             int
	stashCapacity         int
	resourceWallet        map[string]int
	corpses               map[string]*corpseState
	hpRegenCarry          float64
	manaRegenCarry        float64
	nextBasicAttackTick   uint64
	nextOffHandAttackTick uint64
}

// CharacterProgressionState is the authoritative mutable progression state for
// one character inside a sim session.
type CharacterProgressionState struct {
	CharacterClass      string
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
func NewSim(sessionID, seed string, rules *Rules) (*Sim, error) {
	return NewSimWithWorld(sessionID, seed, rules, DefaultWorldID)
}

// MustNewSim builds a fresh default-world session or panics.
// It is intended for tests with known-valid fixtures; runtime callers should use NewSim.
func MustNewSim(sessionID, seed string, rules *Rules) *Sim {
	s, err := NewSim(sessionID, seed, rules)
	if err != nil {
		panic(err)
	}
	return s
}

func fogOfWarEnabledForSeed(seed string) bool {
	return strings.Contains(seed, "fog_of_war")
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
		gameplayDebug:         gameplayDebugEnabledFromEnv(),
		nextID:                baseEntityID,
		nextAreaHealZoneID:    1,
		players:               make(map[uint64]*playerState),
		levels:                make(map[int]*LevelState),
		currentLevel:          levelZero,
		multiLevel:            world.Mode == worldModeMultiLevel,
		fogOfWarEnabled:       fogOfWarEnabledForSeed(seed),
		equipped:              newEquippedMap(),
		weaponSets:            newWeaponSetMaps(),
		activeWeaponSet:       defaultWeaponSet,
		hotbar:                make([]uint64, 10),
		discoveredTeleporters: make(map[int]bool),
		progression:           progression,
		skillCooldowns:        make(map[string]skillCooldownState),
		skillEffects:          make(map[string]skillEffectState),
		poisonDots:            make(map[uint64]poisonDotState),
		bleedDots:             make(map[uint64]bleedDotState),
		rogueMarks:            make(map[uint64]rogueMarkState),
		uniqueBurnDots:        make(map[string]uniqueBurnDotState),
		uniqueExecutionMarks:  make(map[uint64]uniqueExecutionMarkState),
		uniqueHungerStacks:    make(map[uint64]uniqueHungerStackState),
		uniqueAshenReprisals:  make(map[uint64]uniqueAshenReprisalState),
		uniqueChests:          make(map[uint64]*uniqueChestState),
		areaHealZones:         make(map[uint64]areaHealZoneState),
		skillFunctionKeys:     make([]string, skillFunctionKeyCount),
		shopStock:             make(map[string]*shopStockState),
		resourceWallet:        make(map[string]int),
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
		s.seedUniqueTestChests()

		return s, nil
	}

	level := newLevelState(levelZero, &rules.Navigation)
	s.levels[levelZero] = level
	if err := s.populatePresetLevel(level, worldID, world); err != nil {
		return nil, err
	}

	s.syncCompatibilityFields()
	s.seedUniqueTestChests()

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
		CharacterClass:      "barbarian",
		Experience:          0,
		UnspentStatPoints:   0,
		UnspentSkillPoints:  r.skillPointsGrantedAtLevel(1),
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
	if in.CharacterClass == "" {
		in.CharacterClass = "barbarian"
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
		WeaponSets:            cloneWeaponSetMaps(s.weaponSets),
		ActiveWeaponSet:       s.activeWeaponSet,
		Hotbar:                s.hotbar,
		DiscoveredTeleporters: s.discoveredTeleporters,
		Progression:           s.progression,
		SkillCooldowns:        cloneSkillCooldowns(s.skillCooldowns),
		SkillEffects:          cloneSkillEffects(s.skillEffects),
		PoisonDots:            clonePoisonDots(s.poisonDots),
		BleedDots:             cloneBleedDots(s.bleedDots),
		RogueMarks:            cloneRogueMarks(s.rogueMarks),
		UniqueBurnDots:        cloneUniqueBurnDots(s.uniqueBurnDots),
		UniqueExecutionMarks:  cloneUniqueExecutionMarks(s.uniqueExecutionMarks),
		UniqueHungerStacks:    cloneUniqueHungerStacks(s.uniqueHungerStacks),
		UniqueAshenReprisals:  cloneUniqueAshenReprisals(s.uniqueAshenReprisals),
		UniquePilgrimMomentum: cloneUniquePilgrimMomentum(s.uniquePilgrimMomentum),
		UniqueChestItems:      cloneUniqueChestItems(s.uniqueChests),
		SkillFunctionKeys:     cloneStringSlice(s.skillFunctionKeys),
		RightClickSkillID:     s.rightClickSkillID,
		ShopStock:             s.shopStock,
		Gold:                  s.gold,
		StashItems:            s.stashItems,
		StashGold:             s.stashGold,
		StashCapacity:         s.stashCapacity,
		ResourceWallet:        cloneIntMap(s.resourceWallet),
	}

	for _, preset := range world.Entities {
		switch preset.Type {
		case monsterEntity, companionEntity:
			monster := s.newPresetMonsterOrCompanion(level, preset, player.id)
			monster.id = s.alloc()
			level.entities[monster.id] = monster
		case lootEntity:
			loot := s.newLootEntity(preset.ItemDefID, preset.Position, nil, goldRollContext{levelNum: level.levelNum})
			if preset.ItemDefID == UpgradeShardItemDefID {
				loot.rollPayload = NewUpgradeShardRollPayload(1)
			}
			if preset.ItemDefID == RenewStoneItemDefID {
				loot.rollPayload = NewRenewStoneRollPayload(1)
			}
			if preset.ItemTemplateID != "" {
				rolled, ok := s.rollItemTemplate(preset.ItemTemplateID, absInt(level.levelNum))
				if !ok {
					return ErrUnknownWorldEntity{WorldID: worldID, EntityType: preset.Type}
				}
				loot.itemDefID = rolled.ItemTemplateID
				loot.rollPayload = &rolled
			}
			loot.id = s.alloc()
			level.entities[loot.id] = loot
		case wallEntity:
			source := "preset"
			if preset.Kind == obstacleKindWood {
				source = "town_perimeter"
			}
			level.walls = append(level.walls, wallObstacle{pos: preset.Position, size: preset.Size, kind: preset.Kind, source: source, blocksLOS: preset.BlocksLineOfSight})
		case interactableEntity:
			def := s.rules.Interactables[preset.InteractableDefID]
			if def.Service == uniqueTestChestService && !s.gameplayDebug {
				continue
			}
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

func (s *Sim) alloc() uint64 {
	id := s.nextID
	s.nextID++
	return id
}

// CurrentTick returns the next tick to be processed.
func (s *Sim) CurrentTick() uint64 { return s.tick }

// SetGameplayDebug enables local development-only gameplay conveniences.
func (s *Sim) SetGameplayDebug(enabled bool) {
	s.gameplayDebug = enabled
	if enabled {
		s.seedUniqueTestChests()
	}
}

func gameplayDebugEnabledFromEnv() bool {
	switch os.Getenv("ARPG_GAMEPLAY_DEBUG") {
	case "1", "true", "TRUE", "yes", "YES", "on", "ON":
		return true
	default:
		return false
	}
}

// Input is a decoded client intent applied to a specific tick.
type Input struct {
	MessageID           string
	CorrelationID       string
	Sequence            int64
	ActorPlayerID       uint64
	Type                string
	Move                *MoveIntent
	MoveTo              *MoveToIntent
	DirectionalAttack   *DirectionalAttackIntent
	Action              *ActionIntent
	Descend             *DescendIntent
	Ascend              *AscendIntent
	Teleport            *TeleportIntent
	Equip               *EquipIntent
	Unequip             *UnequipIntent
	SwapWeaponSet       *SwapWeaponSetIntent
	Drop                *DropIntent
	Use                 *UseIntent
	AssignHotbar        *AssignHotbarIntent
	UseHotbar           *UseHotbarIntent
	AllocateStat        *AllocateStatIntent
	AllocateSkillPoint  *AllocateSkillPointIntent
	CastSkill           *CastSkillIntent
	ChannelSkill        *ChannelSkillIntent
	SetSkillBindings    *SetSkillBindingsIntent
	CompanionCommand    *CompanionCommandIntent
	ShopBuy             *ShopBuyIntent
	ShopSell            *ShopSellIntent
	ShopReroll          *ShopRerollIntent
	BishopRespec        *BishopRespecIntent
	BishopReviveAll     *BishopReviveAllIntent
	BishopDebugLevel    *BishopDebugLevelIntent
	BishopDebugSkill    *BishopDebugSkillPointIntent
	BishopDebugStat              *BishopDebugStatPointIntent
	BishopDebugDropUpgradeShard  *BishopDebugDropUpgradeShardIntent
	BishopDebugDropRenewStone    *BishopDebugDropRenewStoneIntent
	BishopDebugDropRespecBadge   *BishopDebugDropWalletBadgeIntent
	BishopDebugDropResurrectionBadge *BishopDebugDropWalletBadgeIntent
	StashDepositItem    *StashDepositItemIntent
	StashWithdrawItem   *StashWithdrawItemIntent
	StashDepositGold    *StashDepositGoldIntent
	StashWithdrawGold   *StashWithdrawGoldIntent
	CorpseWithdrawItem  *CorpseWithdrawItemIntent
	UniqueChestTakeItem *UniqueChestTakeItemIntent
}

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
		WeaponSet      *int
	}
	UnequipIntent struct {
		Slot      string
		WeaponSet *int
	}
	SwapWeaponSetIntent struct{}
	DropIntent          struct {
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
	ChannelSkillIntent struct {
		SkillID   string
		Phase     string
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
	ShopRerollIntent struct {
		ShopEntityID string
	}
	BishopRespecIntent     struct{ BishopEntityID string }
	BishopReviveAllIntent  struct{ BishopEntityID string }
	BishopDebugLevelIntent struct {
		BishopEntityID string
	}
	BishopDebugSkillPointIntent struct {
		BishopEntityID string
	}
	BishopDebugStatPointIntent struct {
		BishopEntityID string
	}
	BishopDebugDropUpgradeShardIntent struct {
		BishopEntityID string
	}
	BishopDebugDropRenewStoneIntent struct {
		BishopEntityID string
	}
	BishopDebugDropWalletBadgeIntent struct {
		BishopEntityID string
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
	CorpseWithdrawItemIntent struct {
		CorpseEntityID string
		ItemInstanceID string
	}
	UniqueChestTakeItemIntent struct {
		ChestEntityID string
		ChestItemID   string
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

	aggroedMonsterIDs map[uint64]map[uint64]bool
}

func (r *TickResult) ack(id string) { r.Acks = append(r.Acks, Ack{MessageID: id}) }
func (r *TickResult) reject(id, reason string) {
	r.Rejects = append(r.Rejects, Reject{MessageID: id, Reason: reason})
}
func (r *TickResult) aggroAlreadyProcessed(playerID, monsterID uint64) bool {
	if r == nil || r.aggroedMonsterIDs == nil {
		return false
	}
	return r.aggroedMonsterIDs[playerID][monsterID]
}
func (r *TickResult) markAggroProcessed(playerID, monsterID uint64) {
	if r == nil {
		return
	}
	if r.aggroedMonsterIDs == nil {
		r.aggroedMonsterIDs = map[uint64]map[uint64]bool{}
	}
	if r.aggroedMonsterIDs[playerID] == nil {
		r.aggroedMonsterIDs[playerID] = map[uint64]bool{}
	}
	r.aggroedMonsterIDs[playerID][monsterID] = true
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
		res.Events = append(res.Events, Event{
			EventType: "player_mana_regenerated",
			EntityID:  idStr(player.id),
			Mana:      intPtr(delta),
		})
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
	activePlayerProgressionChanged := false
	updatedTargets := map[uint64]bool{}
	for _, stateKey := range sortedStringKeys(s.skillEffects) {
		effect := s.skillEffects[stateKey]
		if effect.EndsTick > s.tick {
			continue
		}
		delete(s.skillEffects, stateKey)
		changed = true
		targetID := effect.TargetID
		if targetID == 0 {
			targetID = s.playerID
		}
		target := s.activeLevel().entities[targetID]
		if target != nil {
			if effect.EffectID != "" {
				target.effectIDs = removeStringValue(target.effectIDs, effect.EffectID)
			}
			if target.id == s.playerID {
				activePlayerProgressionChanged = true
			}
			updatedTargets[target.id] = true
			res.Events = append(res.Events, Event{
				EventType: "skill_effect_ended",
				EntityID:  idStr(target.id),
				SkillID:   effect.SkillID,
			})
		}
	}
	if !changed || player == nil {
		return
	}
	resourcesChanged := s.syncActivePlayerMaxResources()
	visualChanged := s.syncActivePlayerVisualScale()
	if resourcesChanged || visualChanged {
		updatedTargets[player.id] = true
	}
	if activePlayerProgressionChanged {
		s.appendCharacterProgressionUpdate(res)
	}
	for _, id := range sortedUint64Keys(updatedTargets) {
		target := s.activeLevel().entities[id]
		if target != nil && target.hp > 0 {
			res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
		}
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

func (s *Sim) appendCharacterProgressionUpdate(res *TickResult) {
	view := s.CharacterProgressionView()
	res.Changes = append(res.Changes, Change{Op: OpCharacterProgressionUpdate, Progression: &view})
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

// ActiveNavigationRules exposes navigation tuning for realtime guardrails.
func (s *Sim) ActiveNavigationRules() NavigationRules {
	return s.activeNav()
}

func (s *Sim) activeWalls() []wallObstacle {
	level := s.activeLevel()
	if s.walls != nil {
		level.walls = s.walls
	}
	return level.walls
}

func (s *Sim) syncCompatibilityFields() {
	s.syncActiveWeaponSetToEquipped()
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
		if target.interactableDefID == heroCorpseDefID {
			s.openCorpse(target, in, res, ack)
			return
		}
		s.activateInteractable(target, in, res, ack)
	default:
		res.reject(in.MessageID, "invalid_target")
	}
}

func (s *Sim) attackTarget(target *entity, in Input, res *TickResult, ack bool) {
	weaponSlot, ok := s.consumeBasicAttack(in, res)
	if !ok {
		return
	}
	if ack {
		res.ack(in.MessageID)
	}
	s.damageMonsterByPlayerWithSlot(target, s.playerID, in.CorrelationID, res, s.resolvePlayerAttackDamageForSlot(weaponSlot), s.playerWeaponDamageTypeForSlot(weaponSlot), weaponSlot)
}

func (s *Sim) damageMonsterByPlayer(target *entity, playerID uint64, corr string, res *TickResult, damageRange DamageRange) combatResolution {
	return s.damageMonsterByPlayerWithSlot(target, playerID, corr, res, damageRange, damageTypeForce, "")
}

func (s *Sim) damageMonsterByPlayerWithSlot(target *entity, playerID uint64, corr string, res *TickResult, damageRange DamageRange, damageType string, weaponSlot string) combatResolution {
	damageRange = s.applyUniqueDamageBeforeHeroHit(target, playerID, damageRange)
	damageRange = s.applyRogueMarkDamageBonus(target, damageRange)
	attackerStats, _ := s.playerEffectiveCombatStats()
	defenderStats := s.monsterEffectiveCombatStats(target, DamageRange{})
	outcome := s.resolveCombat(attackerStats, defenderStats, damageRange)
	s.applyMonsterResistanceToOutcome(target, damageType, &outcome)
	if !outcome.Hit || outcome.Blocked {
		res.Events = append(res.Events, combatEvent(s.combatEventType(monsterEntity, outcome), playerID, target.id, corr, outcome))
		if weaponSlot != "" {
			res.Events[len(res.Events)-1].WeaponSlot = weaponSlot
		}
		s.aggroMonsterOnHit(target, playerID, corr, res)
		return outcome
	}

	target.hp -= outcome.Damage
	if target.hp < 0 {
		target.hp = 0
	}
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
	res.Events = append(res.Events, combatEvent(s.combatEventType(monsterEntity, outcome), playerID, target.id, corr, outcome))
	if weaponSlot != "" {
		res.Events[len(res.Events)-1].WeaponSlot = weaponSlot
	}

	if outcome.Damage > 0 {
		s.aggroMonsterOnHit(target, playerID, corr, res)
	}
	s.triggerUniqueEffectsAfterHeroDamage(target, playerID, corr, res, outcome, uniqueHeroDamageSource{BasicAttack: true})
	if outcome.Damage > 0 {
		s.tryPassiveExecute(target, playerID, corr, res)
	}
	if target.hp == 0 {
		s.finishMonsterKill(target, playerID, corr, res)
	}
	s.retaliate(target, corr, res)
	return outcome
}

func (s *Sim) damageMonsterByPlayerSkill(target *entity, playerID uint64, corr string, res *TickResult, damageRange DamageRange) combatResolution {
	return s.damageMonsterByPlayerSkillTyped(target, playerID, corr, res, damageRange, damageTypeForce)
}

func (s *Sim) damageMonsterByPlayerSkillTyped(target *entity, playerID uint64, corr string, res *TickResult, damageRange DamageRange, damageType string) combatResolution {
	return s.damageMonsterByPlayerSkillTypedWithID(target, playerID, "", corr, res, damageRange, damageType)
}

func (s *Sim) damageMonsterByPlayerSkillTypedWithID(target *entity, playerID uint64, skillID string, corr string, res *TickResult, damageRange DamageRange, damageType string) combatResolution {
	damageRange = s.applyUniqueDamageBeforeHeroHit(target, playerID, damageRange)
	damageRange = s.applyRogueMarkDamageBonus(target, s.applySkillDamageModifiers(playerID, skillID, damageRange))
	defenderStats := s.monsterEffectiveCombatStats(target, DamageRange{})
	outcome := s.resolveSkillDamage(defenderStats, damageRange)
	s.applyMonsterResistanceToOutcome(target, damageType, &outcome)
	target.hp -= outcome.Damage
	if target.hp < 0 {
		target.hp = 0
	}
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
	res.Events = append(res.Events, combatEvent(s.combatEventType(monsterEntity, outcome), playerID, target.id, corr, outcome))
	if outcome.Damage > 0 {
		s.aggroMonsterOnHit(target, playerID, corr, res)
	}
	s.triggerUniqueEffectsAfterHeroDamage(target, playerID, corr, res, outcome, uniqueHeroDamageSource{BasicAttack: false})
	if outcome.Damage > 0 {
		s.tryPassiveExecute(target, playerID, corr, res)
	}
	if target.hp == 0 {
		s.finishMonsterKill(target, playerID, corr, res)
	}
	s.retaliate(target, corr, res)
	return outcome
}

func (s *Sim) resolveSkillDamage(defender effectiveCombatStats, damageRange DamageRange) combatResolution {
	if damageRange.Max < damageRange.Min {
		damageRange.Max = damageRange.Min
	}
	raw := s.rollRange(damageRange)
	mitigated := raw - int(math.Round(defender.Armor))
	finalDamage := mitigated
	if finalDamage < s.rules.Combat.MinimumDamage {
		finalDamage = s.rules.Combat.MinimumDamage
	}
	return combatResolution{
		Outcome:         "hit",
		Damage:          finalDamage,
		RawDamage:       raw,
		MitigatedDamage: mitigated,
		Blocked:         false,
		Critical:        false,
		Hit:             true,
	}
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
	if _, ok := s.consumeBasicAttack(in, res); !ok {
		return
	}
	maxDistance := s.playerActionReach()
	projectile := &entity{
		kind:             projectileEntity,
		pos:              player.pos,
		ownerID:          player.id,
		targetID:         targetID,
		projectileDefID:  s.playerProjectileDefID(),
		dir:              dir,
		speed:            projectileSpeed,
		maxDistance:      maxDistance,
		damageRange:      s.resolvePlayerAttackDamage(),
		sourceDamageType: s.playerWeaponDamageTypeForSlot(mainHandSlot),
		sourceMsgID:      in.MessageID,
		sourceCorrID:     in.CorrelationID,
		spawnTick:        s.tick,
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
func (s *Sim) dropLoot(monster *entity, sourceID uint64, corr string, res *TickResult) {
	drops := s.rules.LootDrops(monster.lootTable, s.rng)
	s.spawnLootDrops(drops, monster.pos, s.targetInteractionRadius(monster), corr, res, goldRollContext{
		levelNum:        s.activeLevel().levelNum,
		monsterRarityID: monster.monsterRarityID,
		magicFind:       true,
	})
	depth := absInt(s.activeLevel().levelNum)
	hook := monsterResourceLootHook(monster.monsterRarityID, monster.isBoss)
	s.tryDropResourceLoot(monster.pos, s.targetInteractionRadius(monster), depth, hook, corr, res)
	s.grantBossBadgeRewards(monster, sourceID, corr, res)
}

func (s *Sim) finishMonsterKill(monster *entity, sourceID uint64, corr string, res *TickResult) {
	s.triggerUniqueEffectsOnMonsterKilled(monster, sourceID, corr, res)
	res.Events = append(res.Events, Event{
		EventType:      "monster_killed",
		EntityID:       idStr(monster.id),
		SourceEntityID: idStr(sourceID),
		TargetEntityID: idStr(monster.id),
		CorrelationID:  corr,
	})
	if monster.isBoss {
		res.Events = append(res.Events, Event{
			EventType:      "boss_killed",
			EntityID:       idStr(monster.id),
			SourceEntityID: idStr(sourceID),
			TargetEntityID: idStr(monster.id),
			BossTemplateID: monster.bossTemplateID,
			CorrelationID:  corr,
		})
	}
	s.dropLoot(monster, sourceID, corr, res)
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
		if drop.UniqueItemID != "" {
			rolled, ok := s.rules.namedUniquePayload(drop.UniqueItemID)
			if !ok {
				continue
			}
			payload = &rolled
			itemDefID = rolled.ItemTemplateID
		} else if drop.SetItemID != "" {
			rolled, ok := s.rules.setItemPayload(drop.SetItemID)
			if !ok {
				continue
			}
			payload = &rolled
			itemDefID = rolled.ItemTemplateID
		} else if drop.ItemTemplateID != "" {
			rolled, ok := s.rollItemTemplateForLoot(drop.ItemTemplateID, s.itemRollSourceDepth(goldCtx), goldCtx)
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
	if outcome, immune := s.playerDamageImmunityOutcome(player); immune {
		res.Events = append(res.Events, combatEvent(s.combatEventType(playerEntity, outcome), monster.id, player.id, corr, outcome))
		return
	}
	attackerStats := s.monsterEffectiveCombatStats(monster, retaliationDamage)
	defenderStats, _ := s.playerEffectiveCombatStats()
	outcome := s.resolveCombat(attackerStats, defenderStats, retaliationDamage)
	if !outcome.Hit || outcome.Blocked {
		s.triggerUniqueEffectsAfterPlayerAvoidedHit(player, monster, corr, res)
		res.Events = append(res.Events, combatEvent(s.combatEventType(playerEntity, outcome), monster.id, player.id, corr, outcome))
		return
	}
	outcome = s.applyUniqueEffectsBeforePlayerDamage(player, monster, corr, res, outcome, uniqueIncomingDamageSource{})
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
	s.triggerUniqueEffectsAfterPlayerDamage(player, monster, corr, res, outcome)
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
	if s.isWalletResourceItem(e.itemDefID) {
		s.pickUpWalletResource(e, in, res, ack)
		return
	}
	itemDefID := e.itemDefID
	itemSlot := s.itemSlot(itemDefID, e.rollPayload)
	hotbarSlot := 0
	assignToHotbar := false
	if e.rollPayload == nil {
		if def, ok := s.rules.Items[itemDefID]; ok && def.Category == "consumable" {
			hotbarSlot, assignToHotbar = s.firstEmptyActiveHotbarSlot()
		}
	}
	if !assignToHotbar && s.bagOccupancyCount()+1 > s.inventoryCapacity() {
		res.reject(in.MessageID, "inventory_full")
		return
	}

	item := &invItem{
		instanceID:  s.alloc(),
		itemDefID:   itemDefID,
		rollPayload: cloneRollPayload(e.rollPayload),
		slot:        itemSlot,
		equipped:    false,
	}

	delete(s.activeLevel().entities, e.id)
	res.Changes = append(res.Changes, Change{Op: OpEntityRemove, EntityID: idStr(e.id)})

	s.inventory = append(s.inventory, item)
	res.Changes = append(res.Changes, Change{Op: OpInventoryAdd, Item: ptrItemView(s.itemView(item))})
	if assignToHotbar {
		itemID := idStr(item.instanceID)
		itemView := s.itemView(item)
		s.hotbar[hotbarSlot] = item.instanceID
		res.Changes = append(res.Changes,
			Change{Op: OpInventoryRemove, ItemInstanceID: &itemID},
			Change{
				Op:             OpHotbarUpdate,
				SlotIndex:      hotbarSlot,
				ItemInstanceID: &itemID,
				Item:           &itemView,
				InventoryRows:  intPtr(s.inventoryRows()),
				InventoryCap:   intPtr(s.inventoryCapacity()),
			},
		)
	}
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

func (s *Sim) openMarketService(e *entity, in Input, res *TickResult, ack bool) {
	if e.state != interactableReady {
		res.reject(in.MessageID, "not_actionable")
		return
	}
	res.Events = append(res.Events, Event{
		EventType:     "market_service_opened",
		EntityID:      idStr(e.id),
		CorrelationID: in.CorrelationID,
		Service:       "market",
	})
	if ack {
		res.ack(in.MessageID)
	}
}

func (s *Sim) openBlacksmithService(e *entity, in Input, res *TickResult, ack bool) {
	if e.state != interactableReady {
		res.reject(in.MessageID, "not_actionable")
		return
	}
	res.Events = append(res.Events, Event{
		EventType:     "blacksmith_service_opened",
		EntityID:      idStr(e.id),
		CorrelationID: in.CorrelationID,
		Service:       "blacksmith",
		StashItems:    s.stashItemViews(),
		StashGold:     intPtr(s.stashGold),
		StashCapacity: intPtr(s.stashCapacity),
	})
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

func (s *Sim) resolveBishopIntentTarget(bishopEntityID string) (*entity, bool, string) {
	bishopEntity, levelNum, ok := s.findEntityAnyLevel(bishopEntityID)
	if !ok || bishopEntity.kind != interactableEntity {
		return nil, false, "invalid_target"
	}
	if s.serviceForInteractable(bishopEntity) != "bishop" {
		return nil, false, "invalid_target"
	}
	if levelNum != s.currentLevel || s.currentLevel != townLevel {
		return nil, false, "out_of_range"
	}
	if !s.inDispatchRange(bishopEntity) {
		return nil, false, "out_of_range"
	}
	if bishopEntity.state != interactableReady {
		return nil, false, "not_actionable"
	}
	return bishopEntity, true, ""
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

func (s *Sim) serviceForInteractable(e *entity) string {
	if e == nil || e.kind != interactableEntity {
		return ""
	}
	return s.rules.Interactables[e.interactableDefID].Service
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
	s.transferOwnedCompanionsToLevel(player.id, current, dest, arrivalPos, res)
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
		if obstacleBlocksMovement(wall) && circleIntersectsAABB(pos, playerRadius, wall.pos, wall.size) {
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
	s.syncActivePlayerResourceCaps(res)
	view := s.CharacterProgressionView()
	res.Changes = append(res.Changes, Change{Op: OpCharacterProgressionUpdate, Progression: &view})
	s.appendInventoryPresentationUpdates(res)
}

func (s *Sim) syncActivePlayerResourceCaps(res *TickResult) {
	player := s.activeLevel().entities[s.playerID]
	if player == nil {
		return
	}
	changed := false
	if maxHP := s.currentMaxHP(); maxHP != player.maxHP {
		delta := maxHP - player.maxHP
		player.maxHP = maxHP
		if delta > 0 {
			player.hp += delta
		}
		if player.hp > player.maxHP {
			player.hp = player.maxHP
		}
		changed = true
	}
	if maxMana := s.currentMaxMana(); maxMana != player.maxMana {
		delta := maxMana - player.maxMana
		player.maxMana = maxMana
		if delta > 0 {
			player.mana += delta
		}
		if player.mana > player.maxMana {
			player.mana = player.maxMana
		}
		changed = true
	}
	if changed && res != nil {
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	}
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
	return s.effectiveSkillCooldownTicks(def)
}

func (s *Sim) baseSkillCooldownTicks(def SkillDef) int {
	if def.Cooldown.Type == "none" {
		return 0
	}
	if def.Cooldown.FixedTicks > 0 {
		return def.Cooldown.FixedTicks
	}
	interval := s.DerivedStatsView().AttackIntervalTicks
	if interval < 1 {
		interval = s.rules.Combat.BaseAttackIntervalTicks
	}
	cooldown := int(math.Ceil(float64(interval)*def.Cooldown.Multiplier)) + def.Cooldown.FlatTicks
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

func (s *Sim) skillCastDirection(def SkillDef, cast *CastSkillIntent, player *entity) (Vec2, uint64, string) {
	return s.skillCastDirectionWithRange(def, cast, player, def.Projectile.Range)
}

func (s *Sim) skillCastDirectionWithRange(def SkillDef, cast *CastSkillIntent, player *entity, castRange float64) (Vec2, uint64, string) {
	if cast == nil || player == nil {
		return Vec2{}, 0, "invalid_payload"
	}
	if cast.TargetID != "" {
		target := s.findEntity(cast.TargetID)
		if target == nil || target.kind != monsterEntity || target.hp <= 0 {
			return Vec2{}, 0, "invalid_target"
		}
		if distance(player.pos, target.pos) > castRange+meleeRangeEpsilon {
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
	damageRange := s.scaleSkillDamageForMagic(def, rank, s.skillDamageRange(def, rank))
	projectile := &entity{
		kind:             projectileEntity,
		pos:              player.pos,
		ownerID:          player.id,
		targetID:         targetID,
		projectileDefID:  skillID,
		dir:              normalize(dir),
		speed:            def.Projectile.Speed,
		maxDistance:      def.Projectile.Range,
		damageRange:      damageRange,
		sourceSkillID:    skillID,
		sourceDamageType: s.skillDamageType(def),
		sourceMsgID:      in.MessageID,
		sourceCorrID:     in.CorrelationID,
		spawnTick:        s.tick,
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

func skillCastRange(def SkillDef) float64 {
	if def.Projectile.Range > 0 {
		return def.Projectile.Range
	}
	castRange := 0.0
	for _, effect := range def.Effects {
		if effect.Range > castRange {
			castRange = effect.Range
		}
	}
	return castRange
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
		if target.kind == monsterEntity {
			return target.pos, ""
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

func (s *Sim) healSkillTargets(center Vec2, effect SkillEffectDef, casterID uint64, radius float64) []*entity {
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
		if distance(center, entity.pos) > radius+meleeRangeEpsilon {
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
		percent := s.scaleSkillPercentForMagic(def, rank, effect, skillEffectPercent(effect, rank))
		targets := s.healSkillTargets(center, effect, player.id, s.scaleSkillRadiusForMagic(def, rank, effect))
		for _, target := range targets {
			if target.hp >= target.maxHP {
				continue
			}
			heal := healAmountForTarget(target, percent)
			if heal <= 0 {
				continue
			}
			applications = append(applications, skillHealApplication{Target: target, Heal: heal})
		}
		if effect.IncludeCaster && player.hp < player.maxHP && !skillHealApplicationsContainTarget(applications, player.id) {
			heal := healAmountForTarget(player, percent)
			if heal > 0 {
				applications = append(applications, skillHealApplication{Target: player, Heal: heal})
			}
		}
	}
	return applications, ""
}

func skillHealApplicationsContainTarget(applications []skillHealApplication, targetID uint64) bool {
	for _, app := range applications {
		if app.Target != nil && app.Target.id == targetID {
			return true
		}
	}
	return false
}

func (s *Sim) startAreaHealZones(player *entity, skillID string, def SkillDef, rank int, cast *CastSkillIntent, correlationID string) {
	if player == nil {
		return
	}
	for _, effect := range def.Effects {
		if effect.Type != "area_percent_heal" || effect.DurationTicks <= 0 {
			continue
		}
		center, rejectReason := s.skillAreaCenter(effect, cast, player)
		if rejectReason != "" {
			continue
		}
		id := s.nextAreaHealZoneID
		s.nextAreaHealZoneID++
		s.areaHealZones[id] = areaHealZoneState{
			ID:            id,
			Level:         s.currentLevel,
			Center:        center,
			CasterID:      player.id,
			SkillID:       skillID,
			Rank:          rank,
			Percent:       s.scaleSkillPercentForMagic(def, rank, effect, skillEffectPercent(effect, rank)),
			Radius:        s.scaleSkillRadiusForMagic(def, rank, effect),
			IncludeCaster: effect.IncludeCaster,
			CorrelationID: correlationID,
			NextPulseTick: s.tick + uint64(healRainPulseIntervalTicks),
			EndsTick:      s.tick + uint64(effect.DurationTicks),
		}
	}
}

func (s *Sim) advanceAreaHealZones(resultFor func(level int, actor uint64) *TickResult) {
	if len(s.areaHealZones) == 0 {
		return
	}
	for _, zoneID := range sortedUint64Keys(s.areaHealZones) {
		zone := s.areaHealZones[zoneID]
		if s.tick >= zone.EndsTick {
			delete(s.areaHealZones, zoneID)
			continue
		}
		if s.tick < zone.NextPulseTick {
			continue
		}
		res := resultFor(zone.Level, 0)
		s.applyAreaHealZonePulse(zone, res)
		for zone.NextPulseTick <= s.tick {
			zone.NextPulseTick += uint64(healRainPulseIntervalTicks)
		}
		s.areaHealZones[zoneID] = zone
	}
}

func (s *Sim) applyAreaHealZonePulse(zone areaHealZoneState, res *TickResult) {
	level := s.levels[zone.Level]
	if level == nil {
		return
	}
	for _, id := range sortedEntityIDs(level.entities) {
		target := level.entities[id]
		if target == nil || target.kind != playerEntity || target.hp <= 0 {
			continue
		}
		if target.id == zone.CasterID && !zone.IncludeCaster {
			continue
		}
		if distance(zone.Center, target.pos) > zone.Radius+meleeRangeEpsilon {
			continue
		}
		heal := healAmountForTarget(target, zone.Percent)
		if heal <= 0 {
			continue
		}
		target.hp += heal
		if target.hp > target.maxHP {
			target.hp = target.maxHP
		}
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
		res.Events = append(res.Events, Event{
			EventType:      "player_healed",
			EntityID:       idStr(target.id),
			SourceEntityID: idStr(zone.CasterID),
			TargetEntityID: idStr(target.id),
			CorrelationID:  zone.CorrelationID,
			SkillID:        zone.SkillID,
			Rank:           intPtr(zone.Rank),
			Heal:           intPtr(heal),
		})
	}
}

func healAmountForTarget(target *entity, percent int) int {
	if target == nil || target.hp >= target.maxHP {
		return 0
	}
	heal := int(math.Floor(float64(target.maxHP)*float64(percent)/100.0 + 0.000000001))
	if heal < 1 {
		heal = 1
	}
	if target.hp+heal > target.maxHP {
		heal = target.maxHP - target.hp
	}
	return heal
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

func (s *Sim) coneSkillTargets(player *entity, dir Vec2, cone SkillConeDef) []*entity {
	targets := []*entity{}
	if player == nil {
		return targets
	}
	dir = normalize(dir)
	if dir.X == 0 && dir.Y == 0 {
		return targets
	}
	cosLimit := math.Cos(cone.AngleDegrees * math.Pi / 360.0)
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		target := s.activeLevel().entities[id]
		if target == nil || target.kind != monsterEntity || target.hp <= 0 {
			continue
		}
		toTarget := Vec2{X: target.pos.X - player.pos.X, Y: target.pos.Y - player.pos.Y}
		dist := distance(player.pos, target.pos)
		if dist > cone.Range+monsterRadius+meleeRangeEpsilon || dist <= meleeRangeEpsilon {
			continue
		}
		targetDir := normalize(toTarget)
		if targetDir.X*dir.X+targetDir.Y*dir.Y+0.000000001 < cosLimit {
			continue
		}
		targets = append(targets, target)
	}
	return targets
}

func (s *Sim) applyConeSkill(player *entity, skillID string, def SkillDef, targets []*entity, correlationID string, res *TickResult) {
	for _, target := range targets {
		if target == nil || target.hp <= 0 {
			continue
		}
		outcome := s.damageMonsterByPlayerSkillTypedWithID(target, player.id, skillID, correlationID, res, s.resolvePlayerAttackDamage(), s.skillDamageType(def))
		if def.Poison.DurationTicks > 0 && outcome.Hit && !outcome.Blocked && target.hp > 0 {
			s.startPoisonDot(player, target, skillID, def, outcome.Damage, correlationID, res)
		}
		if target.hp <= 0 || outcome.Damage <= 0 {
			continue
		}
		push := s.rollFloatRange(def.Cone.PushMin, def.Cone.PushMax)
		if push <= 0 {
			continue
		}
		away := normalize(Vec2{X: target.pos.X - player.pos.X, Y: target.pos.Y - player.pos.Y})
		if away.X == 0 && away.Y == 0 {
			away = Vec2{X: 1}
		}
		before := target.pos
		target.pos = s.resolveMonsterMovement(target, Vec2{X: away.X * push, Y: away.Y * push})
		if target.pos != before {
			res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
			res.Events = append(res.Events, Event{
				EventType:      "monster_pushed",
				EntityID:       idStr(target.id),
				SourceEntityID: idStr(player.id),
				TargetEntityID: idStr(target.id),
				CorrelationID:  correlationID,
				SkillID:        skillID,
				Amount:         intPtr(int(math.Round(push))),
			})
		}
	}
}

func (s *Sim) rollFloatRange(minValue, maxValue float64) float64 {
	if maxValue < minValue {
		maxValue = minValue
	}
	if math.Abs(maxValue-minValue) <= 0.000001 {
		return minValue
	}
	return minValue + s.rollUnitFloat()*(maxValue-minValue)
}

func (s *Sim) applyMovement(res *TickResult) {
	if s.activeLevel().autoNav != nil && s.activeLevel().move == nil {
		s.applyAutoNav(res)
		return
	}
	level := s.activeLevel()
	if level.move == nil || level.move.remaining <= 0 {
		return
	}
	if s.playerDead() {
		level.move = nil
		level.moveMomentumTicks = 0
		return
	}
	player := level.entities[s.playerID]
	before := player.pos
	level.moveMomentumTicks++
	speed := s.playerMoveSpeed() * s.playerMoveMomentumMultiplier(level.moveMomentumTicks)
	player.pos = s.resolveMovement(player.pos, Vec2{
		X: level.move.dir.X * speed,
		Y: level.move.dir.Y * speed,
	})
	level.move.remaining--
	if level.move.remaining == 0 {
		level.move = nil
		level.moveMomentumTicks = 0
	}
	s.updatePilgrimMomentumMovement(player, player.pos != before, res)
	if player.pos == before {
		return
	}
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
}

func dot2(a, b Vec2) float64 {
	return a.X*b.X + a.Y*b.Y
}

func effectiveMonsterAggroRadius(def MonsterDef, rules *Rules) float64 {
	radius := def.AggroRadius
	if rules == nil {
		return radius
	}
	minimum := rules.MainConfig.Gameplay.MinimumMonsterAggroRadius
	if minimum <= 0 || radius < 1.0 || radius >= minimum {
		return radius
	}

	return minimum
}

func effectiveMonsterAssistRadius(def MonsterDef, rules *Rules) float64 {
	radius := def.effectiveAssistRadius()
	aggro := effectiveMonsterAggroRadius(def, rules)
	if radius < aggro {
		return aggro
	}

	return radius
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
	nav := s.activeLevel().autoNav
	player := s.activeLevel().entities[s.playerID]
	before := player.pos
	step := nav.steps[0]
	nav.steps = nav.steps[1:]
	navDist := s.activeNav().CellSize
	// Grid path steps represent one-cell transitions. Use cell_size as the step
	// magnitude so each tick advances one grid cell even when player move speed
	// is below cell_size (e.g. 0.75 vs 1.0). Normalize diagonals to cell_size
	// total displacement so corner-cutting does not drift into wall pockets.
	mag := math.Sqrt(step.X*step.X + step.Y*step.Y)
	if mag < 1e-9 {
		mag = 1
	}
	delta := Vec2{X: step.X * navDist / mag, Y: step.Y * navDist / mag}
	player.pos = s.resolveMovement(player.pos, delta)
	if player.pos != before {
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	} else {
		// Step produced zero movement (wall blocked the planned direction).
		// Discard remaining steps so finishAutoNav re-plans from the exact
		// current position rather than executing steps designed for a position
		// the player never reached.  Clears lastReplanPos so re-plan is allowed.
		nav.steps = nav.steps[:0]
		nav.pathStepsExhausted = false
		nav.lastReplanPos = Vec2{}
	}
	if len(nav.steps) == 0 {
		nav.pathStepsExhausted = true
		s.finishAutoNav(res)
	}
}

func (s *Sim) finishAutoNav(res *TickResult) {
	nav := s.activeLevel().autoNav
	if nav != nil && nav.hasGoal {
		if player := s.activeLevel().entities[s.playerID]; player != nil {
			if s.continueAutoNavToGoal(player, res) {
				return
			}
			// Re-plan when the path ends but the goal is not yet reached.
			// Allow another attempt when the queued path was fully walked, even if
			// the player returned to the same position, but cap attempts to avoid
			// infinite loops on unreachable goals.
			canReplan := nav.replanAttempts < maxAutoNavReplans &&
				(player.pos != nav.lastReplanPos || nav.pathStepsExhausted)
			if distance(player.pos, nav.goal) > s.activeNav().StopDistance && canReplan {
				if steps, ok := s.planPlayerPath(s.activeNav(), player.pos, nav.goal, s.buildBlockedFn()); ok && len(steps) > 0 {
					next := s.newAutoNavState(
						steps, nav.goal,
						nav.pendingAction, nav.pendingSkill,
						nav.sourceMsgID, nav.sourceCorrID,
					)
					next.lastReplanPos = player.pos
					next.replanAttempts = nav.replanAttempts + 1
					s.activeLevel().autoNav = next
					return
				}
			}
		}
	}
	s.clearAutoNav()
	if nav == nil {
		return
	}
	if nav.pendingSkill != nil {
		in := Input{
			MessageID:     nav.sourceMsgID,
			CorrelationID: nav.sourceCorrID,
			Type:          "cast_skill_intent",
			CastSkill:     nav.pendingSkill,
		}
		s.handleCastSkill(in, res)
		return
	}
	if nav.pendingAction == nil {
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
		if obstacleBlocksMovement(wall) && circleIntersectsAABB(pos, playerRadius, wall.pos, wall.size) {
			return true
		}
	}
	return s.playerDynamicBlocked(pos)
}

// playerDynamicBlocked reports whether pos is blocked by a dynamic entity
// (live monster or closed-door barrier). Walls are not checked here so that
// buildBlockedFn can use separate probe positions for walls vs. entities.
func (s *Sim) playerDynamicBlocked(pos Vec2) bool {
	level := s.activeLevel()
	if level == nil {
		return false
	}

	for _, id := range s.cachedSortedEntityIDs() {
		e := level.entities[id]
		if e == nil {
			continue
		}
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
		if obstacleBlocksMovement(wall) && circleIntersectsAABB(pos, lootInteractionRadius, wall.pos, wall.size) {
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
	nav := s.activeNav()
	return func(gx, gy int) bool {
		origin := gridToWorld(nav, gridCell{x: gx, y: gy})
		// Probe static walls at cell center: a cell whose origin lies just
		// outside a wall AABB may still have its center inside it. A* routing
		// through such cells causes continuous movement to stall when the
		// player's floating-point position drifts to the boundary.
		center := Vec2{X: origin.X + nav.CellSize/2, Y: origin.Y + nav.CellSize/2}
		for _, wall := range s.activeWalls() {
			if obstacleBlocksMovement(wall) && circleIntersectsAABB(center, playerRadius, wall.pos, wall.size) {
				return true
			}
		}
		// Probe dynamic entities at cell origin (door barriers use origin so
		// approach cells on the player's side of the door remain accessible).
		return s.playerDynamicBlocked(origin)
	}
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

func (s *Sim) monsterMoveSpeed(monster *entity, def MonsterDef, nav NavigationRules) float64 {
	speed := def.effectiveMoveSpeed(nav)
	if monster == nil || speed <= 0 {
		return speed
	}
	if s.monsterRooted(monster) {
		return 0
	}
	slowPercent := 0
	for _, stateKey := range sortedStringKeys(s.skillEffects) {
		effect := s.skillEffects[stateKey]
		if effect.TargetID != monster.id || effect.EndsTick <= s.tick {
			continue
		}
		if !containsStringValue(effect.Stats, "movement_speed") || effect.Percent <= slowPercent {
			continue
		}
		slowPercent = effect.Percent
	}
	if slowPercent <= 0 {
		return speed
	}
	if slowPercent > 95 {
		slowPercent = 95
	}
	return speed * (1.0 - float64(slowPercent)/100.0)
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
		targetPlayer := s.eliteMinionAttackTarget(s.activeLevel(), monster)
		if targetPlayer == nil {
			continue
		}
		player := s.activeLevel().entities[targetPlayer.PlayerID]
		if player == nil || player.hp <= 0 {
			continue
		}
		s.usePlayer(targetPlayer)
		target := s.monsterAttackTarget(monster, player, def)
		if target == nil {
			continue
		}
		attackCooldown := def.AttackCooldown
		if monster.monsterAttackCooldown > 0 {
			attackCooldown = monster.monsterAttackCooldown
		}
		if monster.hasAttacked && s.tick-monster.lastAttackTick < uint64(attackCooldown) {
			continue
		}
		if monster.attackWindupRemaining > 0 {
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
			s.fireMonsterProjectile(monster, target, def, *attackDamage, res)
			continue
		}
		if s.tryStartMonsterMeleeWindup(monster, target, def, *attackDamage, res) {
			continue
		}
		if target.kind == companionEntity {
			s.damageCompanionByMonster(monster, target, *attackDamage, "", res)
			continue
		}
		s.damagePlayerByMonster(monster, target, *attackDamage, "", res)
	}
}

func (s *Sim) fireMonsterProjectile(monster *entity, target *entity, def MonsterDef, damageRange DamageRange, res *TickResult) {
	dir := normalize(Vec2{X: target.pos.X - monster.pos.X, Y: target.pos.Y - monster.pos.Y})
	if dir.X == 0 && dir.Y == 0 {
		dir = Vec2{X: 1}
	}
	projectile := &entity{
		kind:            projectileEntity,
		pos:             monster.pos,
		ownerID:         monster.id,
		targetID:        target.id,
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
		if prevMode == monsterAIModeChase {
			res.Events = append(res.Events, Event{EventType: "monster_leashed", EntityID: idStr(monster.id)})
		}
		monster.aiMode = monsterAIModeIdle
		monster.aiTargetPlayerID = 0

		return
	}

	if distPlayer <= effectiveMonsterAggroRadius(def, s.rules) {
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
	switch monster.aiMode {
	case monsterAIModeChase:
		if goal, ok := s.monsterRangedRetreatGoal(monster, player, def); ok {
			return goal, true
		}
		if s.monsterInAttackRange(monster, player, def) {
			return Vec2{}, false
		}
		return s.findMonsterChaseGoal(monster, player, def)
	default:
		return Vec2{}, false
	}
}
func (s *Sim) findMonsterChaseGoal(monster *entity, player *entity, def MonsterDef) (Vec2, bool) {
	nav := s.activeNav()
	if goal, ok := s.cachedMonsterNavigationGoal(monster, player); ok {
		return goal, true
	}
	if !s.monsterCanRepath(monster) {
		return Vec2{}, false
	}
	candidates := s.monsterAttackSlotCandidates(monster, player, def)
	var (
		bestGoal       Vec2
		bestSteps      []Vec2
		bestPathLen    int
		bestMonsterDst = math.MaxFloat64
		found          bool
		attempted      bool
	)
	blocked := s.buildMonsterBlockedFn(monster.id)
	for _, goal := range candidates {
		if !s.monsterPathBudgetAvailable() {
			break
		}
		if !s.positionInNavigationBounds(nav, goal) || s.monsterPositionBlocked(goal, monster.id) {
			continue
		}
		if def.effectiveAttackMode() == attackModeRanged && !s.hasClearMonsterRangedShot(goal, player) {
			continue
		}
		attempted = true
		steps, ok := s.planMonsterPath(monster, nav, monster.pos, goal, blocked)
		if !ok {
			continue
		}
		if len(steps) == 0 && distance(monster.pos, goal) > nav.CellSize+nav.StopDistance {
			continue
		}
		if len(steps) > 0 && s.resolveMonsterMovement(monster, s.monsterMoveDelta(monster.pos, goal, steps, s.monsterMoveSpeed(monster, def, nav))) == monster.pos {
			continue
		}
		monsterDst := distance(monster.pos, goal)
		if !found || len(steps) < bestPathLen ||
			(len(steps) == bestPathLen && monsterDst < bestMonsterDst-1e-9) ||
			(len(steps) == bestPathLen && math.Abs(monsterDst-bestMonsterDst) <= 1e-9 && vecLess(goal, bestGoal)) {
			bestGoal = goal
			bestSteps = steps
			bestPathLen = len(steps)
			bestMonsterDst = monsterDst
			found = true
		}
	}
	if !found {
		return s.monsterFallbackChaseGoal(monster, player, def, blocked, attempted)
	}
	s.cacheMonsterNavigationPath(monster, player.id, bestGoal, bestSteps)
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
	nav := s.activeNav()
	walls := s.activeWalls()
	playerIDs := s.cachedSortedPlayerIDs()
	entityIDs := s.cachedSortedEntityIDs()
	return func(gx, gy int) bool {
		center := gridToWorld(nav, gridCell{x: gx, y: gy})
		return s.monsterPositionBlockedWithIDs(center, excludeMonsterID, walls, playerIDs, entityIDs)
	}
}

func (s *Sim) monsterPositionBlocked(pos Vec2, excludeMonsterID uint64) bool {
	return s.monsterPositionBlockedWithIDs(pos, excludeMonsterID, s.activeWalls(), s.cachedSortedPlayerIDs(), s.cachedSortedEntityIDs())
}

func (s *Sim) monsterPositionBlockedWithIDs(pos Vec2, excludeMonsterID uint64, walls []wallObstacle, playerIDs []uint64, entityIDs []uint64) bool {
	monsterDef := s.monsterNavigationDef(excludeMonsterID)
	for _, wall := range walls {
		if monsterObstacleBlocksMovement(wall, monsterDef) && circleIntersectsAABB(pos, monsterRadius, wall.pos, wall.size) {
			return true
		}
	}
	for _, playerID := range playerIDs {
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
	level := s.activeLevel()
	for _, id := range entityIDs {
		if id == excludeMonsterID {
			continue
		}
		e := level.entities[id]
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
	projectileHitCompanion
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
		if !obstacleBlocksProjectiles(wall) {
			continue
		}
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
		case companionEntity:
			if ownerKind != monsterEntity || e.hp <= 0 || e.id != p.targetID {
				continue
			}
			if t, ok := segmentIntersectsCircle(p.pos, candidate, e.pos, monsterRadius+projectileRadius); ok {
				consider(projectileHit{t: t, category: projectileHitCompanion, entityID: e.id})
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
	if hit.category == projectileHitCompanion {
		s.resolveMonsterProjectileCompanionHit(p, hit, res)
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
	if p.sourceSkillID != "" {
		s.resolveSkillProjectileMonsterHit(p, target, res)
		return
	}
	s.damageMonsterByPlayerWithSlot(target, p.ownerID, p.sourceCorrID, res, p.damageRange, p.sourceDamageType, "")
}

func (s *Sim) resolveSkillProjectileMonsterHit(p *entity, target *entity, res *TickResult) {
	skillID := p.sourceSkillID
	def, ok := s.rules.Skills[skillID]
	if !ok {
		s.damageMonsterByPlayerSkillTyped(target, p.ownerID, p.sourceCorrID, res, p.damageRange, p.sourceDamageType)
		return
	}
	damageType := s.skillDamageType(def)
	outcome := s.damageMonsterByPlayerSkillTypedWithID(target, p.ownerID, skillID, p.sourceCorrID, res, p.damageRange, damageType)
	if def.Kind == "chain_projectile_attack" && outcome.Damage > 0 {
		s.applySkillChain(target, p.ownerID, skillID, def, p.damageRange, p.sourceCorrID, res)
		return
	}
	if def.Kind != "cold_projectile_attack" || outcome.Damage <= 0 {
		return
	}
	s.applyMonsterSlow(target, p.ownerID, skillID, def.Slow, p.sourceCorrID, res)
	if p.shardProjectile || target.hp <= 0 {
		return
	}
	s.spawnIceShardProjectiles(target, p.ownerID, skillID, def, outcome.Damage, p.sourceCorrID, res)
}

func (s *Sim) applySkillChain(origin *entity, ownerID uint64, skillID string, def SkillDef, damageRange DamageRange, correlationID string, res *TickResult) {
	if origin == nil || origin.kind != monsterEntity || def.Chain.RangeMultiplier <= 0 || def.Chain.RangeMultiplier >= 1 || def.Chain.MaxJumps <= 0 {
		return
	}
	visited := map[uint64]bool{origin.id: true}
	current := origin
	currentRange := def.Projectile.Range * def.Chain.RangeMultiplier
	for jump := 1; jump <= def.Chain.MaxJumps; jump++ {
		next := s.nearestChainMonster(current.pos, currentRange, visited)
		if next == nil {
			return
		}
		visited[next.id] = true
		dir := normalize(Vec2{X: next.pos.X - current.pos.X, Y: next.pos.Y - current.pos.Y})
		res.Events = append(res.Events, Event{
			EventType:       "skill_chain_hit",
			EntityID:        idStr(ownerID),
			SourceEntityID:  idStr(current.id),
			TargetEntityID:  idStr(next.id),
			CorrelationID:   correlationID,
			SkillID:         skillID,
			Rank:            intPtr(jump),
			ProjectileDefID: def.Chain.Visual,
			Position:        cloneVec2Ptr(&current.pos),
			Direction:       cloneVec2Ptr(&dir),
			Range:           floatPtr(distance(current.pos, next.pos)),
		})
		s.damageMonsterByPlayerSkillTypedWithID(next, ownerID, skillID, correlationID, res, damageRange, s.skillDamageType(def))
		current = next
		currentRange *= def.Chain.RangeMultiplier
	}
}

func (s *Sim) nearestChainMonster(origin Vec2, maxRange float64, visited map[uint64]bool) *entity {
	if maxRange <= 0 {
		return nil
	}
	var best *entity
	bestDist := math.Inf(1)
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		target := s.activeLevel().entities[id]
		if target == nil || target.kind != monsterEntity || target.hp <= 0 || visited[target.id] {
			continue
		}
		dist := distance(origin, target.pos)
		if dist > maxRange+meleeRangeEpsilon {
			continue
		}
		if dist < bestDist {
			best = target
			bestDist = dist
		}
	}
	return best
}

func (s *Sim) spawnIceShardProjectiles(origin *entity, ownerID uint64, skillID string, def SkillDef, damage int, correlationID string, res *TickResult) {
	if origin == nil || def.Shatter.MaxShards <= 0 {
		return
	}
	count := def.Shatter.MinShards
	if def.Shatter.MaxShards > def.Shatter.MinShards {
		count += s.rng.IntN(def.Shatter.MaxShards - def.Shatter.MinShards + 1)
	}
	if count <= 0 {
		return
	}
	shardDamage := damage / count
	if shardDamage < 1 {
		shardDamage = 1
	}
	for i := 0; i < count; i++ {
		angle := s.rollUnitFloat() * 2 * math.Pi
		dir := Vec2{X: math.Cos(angle), Y: math.Sin(angle)}
		start := Vec2{
			X: origin.pos.X + dir.X*(monsterRadius+projectileRadius+0.05),
			Y: origin.pos.Y + dir.Y*(monsterRadius+projectileRadius+0.05),
		}
		projectile := &entity{
			kind:             projectileEntity,
			pos:              start,
			ownerID:          ownerID,
			projectileDefID:  def.Shatter.Visual,
			sourceSkillID:    skillID,
			sourceDamageType: s.skillDamageType(def),
			shardProjectile:  true,
			dir:              dir,
			speed:            def.Shatter.Speed,
			maxDistance:      def.Shatter.Range,
			damageRange:      DamageRange{Min: shardDamage, Max: shardDamage},
			sourceCorrID:     correlationID,
			spawnTick:        s.tick,
		}
		projectile.id = s.alloc()
		s.activeLevel().entities[projectile.id] = projectile
		res.Changes = append(res.Changes, Change{Op: OpEntitySpawn, Entity: ptrEntityView(s.entityView(projectile))})
	}
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
	s.damagePlayerByMonsterWithSource(owner, target, p.damageRange, p.sourceCorrID, res, uniqueIncomingDamageSource{Projectile: true})
}

func (s *Sim) aggroMonsterOnHit(monster *entity, playerID uint64, corr string, res *TickResult) {
	if monster == nil || monster.kind != monsterEntity || playerID == 0 {
		return
	}
	if res.aggroAlreadyProcessed(playerID, monster.id) {
		return
	}
	level := s.activeLevel()
	player := level.entities[playerID]
	entityIDs := sortedEntityIDs(level.entities)
	queue := []*entity{monster}
	res.markAggroProcessed(playerID, monster.id)
	if player != nil && player.kind == playerEntity && player.hp > 0 {
		for _, candidateID := range entityIDs {
			candidate := level.entities[candidateID]
			if candidate == nil || res.aggroAlreadyProcessed(playerID, candidate.id) || !s.canAggroAttackingPlayer(candidate, player) {
				continue
			}
			res.markAggroProcessed(playerID, candidate.id)
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
		for _, candidateID := range entityIDs {
			candidate := level.entities[candidateID]
			if candidate == nil || res.aggroAlreadyProcessed(playerID, candidate.id) || !s.canJoinGroupAggro(current, candidate) {
				continue
			}
			res.markAggroProcessed(playerID, candidate.id)
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
	if !ok || def.effectiveBehavior() != monsterBehaviorChase || def.effectiveAssistRadius() <= 0 {
		return false
	}
	return distance(candidate.pos, player.pos) <= effectiveMonsterAssistRadius(def, s.rules)
}

func (s *Sim) canJoinGroupAggro(source, candidate *entity) bool {
	if source == nil || candidate == nil || source.id == candidate.id || candidate.kind != monsterEntity || candidate.hp <= 0 {
		return false
	}
	def, ok := s.rules.Monsters[candidate.monsterDefID]
	if !ok || def.effectiveBehavior() != monsterBehaviorChase {
		return false
	}
	radius := effectiveMonsterAssistRadius(def, s.rules)
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
		s.restorePlayerResourcesOnLevelUp(corr, res)
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
	if defender.EvadeChance > 0 && s.rollChance(defender.EvadeChance) {
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

func (s *Sim) rollUnitFloat() float64 {
	return float64(s.rng.Next()%1000000) / 1000000.0
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

func (s *Sim) playerDamageImmunityOutcome(player *entity) (combatResolution, bool) {
	if player == nil || player.kind != playerEntity || player.hp <= 0 {
		return combatResolution{}, false
	}
	for _, stateKey := range sortedStringKeys(s.skillEffects) {
		effect := s.skillEffects[stateKey]
		if effect.TargetID != player.id || effect.EffectID != "sanctuary" || effect.EndsTick <= s.tick {
			continue
		}
		return combatResolution{
			Outcome:         "immune",
			Damage:          0,
			RawDamage:       0,
			MitigatedDamage: 0,
			Hit:             true,
		}, true
	}
	return combatResolution{}, false
}

func combatEvent(eventType string, sourceID, targetID uint64, corr string, outcome combatResolution) Event {
	return Event{
		EventType:       eventType,
		EntityID:        idStr(targetID),
		SourceEntityID:  idStr(sourceID),
		TargetEntityID:  idStr(targetID),
		CorrelationID:   corr,
		Damage:          intPtr(outcome.Damage),
		DamageType:      outcome.DamageType,
		Outcome:         outcome.Outcome,
		RawDamage:       intPtr(outcome.RawDamage),
		MitigatedDamage: intPtr(outcome.MitigatedDamage),
		Blocked:         boolPtr(outcome.Blocked),
		Critical:        boolPtr(outcome.Critical),
	}
}

func (s *Sim) rollItemTemplate(templateID string, sourceDepth int) (ItemRollPayload, bool) {
	return s.rules.rollItemTemplateWithRNG(templateID, s.rng, sourceDepth)
}

func (s *Sim) rollItemTemplateForLoot(templateID string, sourceDepth int, ctx goldRollContext) (ItemRollPayload, bool) {
	if !ctx.magicFind {
		return s.rollItemTemplate(templateID, sourceDepth)
	}
	return s.rules.rollItemTemplateWithMagicFind(templateID, s.rng, sourceDepth, int(s.playerMagicFindPercent()))
}

func (s *Sim) itemRollSourceDepth(ctx goldRollContext) int {
	depth := absInt(ctx.levelNum)
	if depth < 1 {
		depth = 1
	}
	if ctx.monsterRarityID != "" {
		if rarity, ok := s.rules.DungeonGeneration.MonsterRarity(ctx.monsterRarityID); ok {
			depth += rarity.LootDepthOffset
		}
	}
	if depth < 1 {
		return 1
	}
	return depth
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

func (s *Sim) slotAcceptsItem(slot string, item *invItem, itemSlot string) bool {
	if slotAcceptsItemSlot(slot, itemSlot) {
		return true
	}
	return slot == offHandSlot && itemSlot == mainHandSlot && s.canOffhandWeapon(item)
}

func (s *Sim) canOffhandWeapon(item *invItem) bool {
	if s.progression.CharacterClass != "rogue" || item == nil {
		return false
	}
	return s.itemHandedness(item) == "one_handed" && s.itemAttackMode(item) == attackModeMelee
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

func (s *Sim) itemOccupiesHandsForSlot(slot string, item *invItem) []string {
	if slot == offHandSlot && s.canOffhandWeapon(item) {
		return []string{offHandSlot}
	}
	return s.itemOccupiesHands(item)
}

func (s *Sim) itemSlotAfterEquip(slot string, item *invItem) string {
	if slot == offHandSlot && s.canOffhandWeapon(item) {
		return mainHandSlot
	}
	return slot
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

func (s *Sim) firstEmptyActiveHotbarSlot() (int, bool) {
	capacity := s.hotbarCapacity()
	if len(s.hotbar) != maxHotbarCapacity {
		s.hotbar = make([]uint64, maxHotbarCapacity)
	}
	for i := 0; i < capacity; i++ {
		if s.hotbar[i] == 0 {
			return i, true
		}
	}
	return 0, false
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
	return s.slotBlockedByHandsForSet(slot, item, s.activeWeaponSet)
}

func (s *Sim) slotBlockedByHandsForSet(slot string, item *invItem, weaponSet int) bool {
	if slot != offHandSlot {
		return false
	}
	mainHand := s.findItemByID(s.equippedSlot(mainHandSlot, weaponSet))
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
	for _, occupied := range s.itemOccupiesHandsForSlot(slot, item) {
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

func (s *Sim) itemHandedness(item *invItem) string {
	if item == nil {
		return ""
	}
	if item.rollPayload != nil {
		if template, ok := s.rules.ItemTemplates[item.rollPayload.ItemTemplateID]; ok {
			return template.Handedness
		}
		return ""
	}
	if def, ok := s.rules.Items[item.itemDefID]; ok {
		return def.Handedness
	}
	return ""
}

func (s *Sim) itemAttackMode(item *invItem) string {
	if item == nil {
		return ""
	}
	if item.rollPayload != nil {
		if template, ok := s.rules.ItemTemplates[item.rollPayload.ItemTemplateID]; ok && template.AttackMode != "" {
			return template.AttackMode
		}
		return attackModeMelee
	}
	if def, ok := s.rules.Items[item.itemDefID]; ok && def.AttackMode != "" {
		return def.AttackMode
	}
	return attackModeMelee
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
	var reach float64
	if item.rollPayload != nil {
		template, ok := s.rules.ItemTemplates[item.rollPayload.ItemTemplateID]
		if !ok || template.Reach <= 0 {
			return 0, false
		}
		reach = template.Reach
	} else {
		def, ok := s.rules.Items[item.itemDefID]
		if !ok || def.Reach == nil {
			return 0, false
		}
		reach = *def.Reach
	}
	if pct := s.itemClassAffinityTotal(item, "reach_percent"); pct != 0 {
		reach = applyPercentDelta(reach, pct)
	}
	return reach, true
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
	if def.effectiveAttackStyle() == monsterAttackStylePounce && def.AttackRange > 0 {
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
		if !obstacleBlocksProjectiles(wall) {
			continue
		}
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
		if !obstacleBlocksProjectiles(wall) {
			continue
		}
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
		if e.interactableDefID == heroCorpseDefID {
			return e.state == interactableReady
		}
		if e.interactableDefID == teleporterDefID && (e.state == interactableReady || e.state == interactableLocked || e.state == interactableDisabled) {
			return true
		}
		if s.shopIDForInteractable(e) != "" && e.state == interactableReady {
			return true
		}
		if s.stashIDForInteractable(e) != "" && e.state == interactableReady {
			return true
		}
		if s.serviceForInteractable(e) != "" && e.state == interactableReady {
			return true
		}
		if s.serviceForInteractable(e) == uniqueTestChestService && e.state == interactableOpen {
			return true
		}
		if e.state == interactableOpen && s.hasClosedBarrier(e) {
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

func (s *Sim) respecCostGold() int {
	if s.rules == nil {
		return 0
	}
	return s.rules.MainConfig.Gameplay.RespecCostGold
}

func (s *Sim) resetCharacterBuildForRespec() {
	classDef, ok := s.rules.CharacterProgression.Classes[s.progression.CharacterClass]
	if ok {
		s.progression.BaseStats = classDef.BaseStats
	} else {
		s.progression.BaseStats = s.rules.CharacterProgression.BaseStats
	}
	s.progression.UnspentStatPoints = s.totalEarnedStatPoints()
	s.progression.UnspentSkillPoints = s.totalEarnedSkillPoints()
	s.progression.SkillRanks = make(map[string]int)
}

func (s *Sim) totalEarnedStatPoints() int {
	level := maxInt(1, s.progression.Level)
	return (level - 1) * s.rules.CharacterProgression.PointsPerLevel
}

func (s *Sim) totalEarnedSkillPoints() int {
	rules := s.rules.CharacterProgression.SkillPoints
	level := maxInt(1, s.progression.Level)
	if level < rules.FirstGrantLevel {
		return 0
	}
	grants := 0
	for grantLevel := rules.FirstGrantLevel; grantLevel <= level; grantLevel += rules.GrantEveryLevels {
		grants++
	}
	return grants * rules.PointsPerGrant
}

func (s *Sim) restorePlayerResources(player *entity, res *TickResult) (int, int) {
	if player == nil {
		return 0, 0
	}
	healed := maxInt(0, player.maxHP-player.hp)
	restored := maxInt(0, player.maxMana-player.mana)
	if healed == 0 && restored == 0 {
		return 0, 0
	}
	player.hp = player.maxHP
	player.mana = player.maxMana
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	return healed, restored
}

// CharacterProgressionView returns the authoritative protocol view of the
// current progression state.
func (s *Sim) CharacterProgressionView() CharacterProgressionView {
	remaining := s.experienceToNextLevel()
	return CharacterProgressionView{
		CharacterClass:        s.progression.CharacterClass,
		Level:                 s.progression.Level,
		Experience:            s.progression.Experience,
		ExperienceToNextLevel: remaining,
		LevelCap:              s.rules.CharacterProgression.LevelCap,
		UnspentStatPoints:     s.progression.UnspentStatPoints,
		UnspentSkillPoints:    s.progression.UnspentSkillPoints,
		Gold:                  s.gold,
		DeepestDungeonDepth:   s.progression.DeepestDungeonDepth,
		BaseStats:             s.progression.BaseStats,
		EffectiveBaseStats:    s.effectiveBaseStatsView(),
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
		baseRank := s.progression.SkillRanks[skillID]
		rank := s.effectiveSkillRank(skillID)
		skills = append(skills, SkillProgressionSkillView{
			SkillID:  skillID,
			Rank:     rank,
			MaxRank:  def.MaxRank,
			CanSpend: s.progression.UnspentSkillPoints > 0 && baseRank < def.MaxRank && s.skillClassAllowed(def) && s.skillRequirementsMet(def, baseRank+1),
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

func (s *Sim) effectiveBaseStatsView() BaseStatsView {
	stats := s.progression.BaseStats
	equipment := s.equipmentBaseStatBonuses()
	stats.Str += equipment.Str
	stats.Dex += equipment.Dex
	stats.Vit += equipment.Vit
	stats.Magic += equipment.Magic
	stats.Str += s.passiveSkillStatTotal("str")
	stats.Dex += s.passiveSkillStatTotal("dex")
	stats.Vit += s.passiveSkillStatTotal("vit")
	stats.Magic += s.passiveSkillStatTotal("magic")
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

func (s *Sim) equipmentBaseStatBonuses() BaseStatsView {
	out := BaseStatsView{}
	for _, slot := range equipmentSlots {
		item := s.findItemByID(s.equipped[slot])
		if item == nil {
			continue
		}
		stats := s.statsForInventoryItem(item)
		out.Str += stats["str"]
		out.Dex += stats["dex"]
		out.Vit += stats["vit"]
		out.Magic += stats["magic"]
	}
	setStats := s.equippedSetBonusStats()
	out.Str += setStats["str"]
	out.Dex += setStats["dex"]
	out.Vit += setStats["vit"]
	out.Magic += setStats["magic"]
	return out
}

func (s *Sim) effectiveSkillRank(skillID string) int {
	baseRank := s.progression.SkillRanks[skillID]
	if baseRank <= 0 {
		return 0
	}
	rank := baseRank + s.allSkillsBonus()
	def, ok := s.rules.Skills[skillID]
	if !ok {
		return rank
	}
	if def.MaxRank > 0 && rank > def.MaxRank {
		rank = def.MaxRank
	}
	for rank > baseRank && !s.skillRequirementsMet(def, rank) {
		rank--
	}
	return rank
}

func (s *Sim) allSkillsBonus() int {
	bonus := 0
	for _, slot := range equipmentSlots {
		item := s.findItemByID(s.equipped[slot])
		if item == nil {
			continue
		}
		stats := s.statsForInventoryItem(item)
		bonus += stats["all_skills"]
	}
	bonus += s.equippedSetBonusStats()["all_skills"]
	return bonus
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
	breakdowns = append(s.baseStatBreakdownViews(), breakdowns...)
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

func (s *Sim) skillClassAllowed(def SkillDef) bool {
	return def.Class == "" || def.Class == s.progression.CharacterClass
}

func (s *Sim) skillEffectLabel(effect skillEffectState) string {
	if def, ok := s.rules.Skills[effect.SkillID]; ok && def.Name != "" {
		return def.Name
	}
	if effect.SkillID != "" {
		return effect.SkillID
	}
	return "Skill effect"
}

func (s *Sim) itemClassAllowed(item *invItem) bool {
	if item == nil || item.rollPayload != nil {
		return true
	}
	def, ok := s.rules.Items[item.itemDefID]
	if !ok {
		return true
	}
	return def.ClassRequired == "" || def.ClassRequired == s.progression.CharacterClass
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
	view.Equipped = s.itemEquippedInAnySlot(item.instanceID)
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
	if item.rollPayload != nil {
		s.annotateClassAffinityStatus(item.rollPayload, func(status []ClassAffinityStatusView) {
			view.ClassAffinityStatus = status
		})
	}
	if previewItem := item.previewItem(); previewItem != nil {
		view.SummaryLines = s.itemSummaryLines("", viewSlotForSummary(previewItem, view.ItemTemplateID, s.rules), s.itemHandedness(previewItem), s.statsForInventoryItem(previewItem), view.Requirements, itemDefPtr(s.rules.Items[item.itemDefID]))
		view.SummaryLines = append(view.SummaryLines, s.setItemSummaryLines(previewItem)...)
		if preview := s.equipPreviewForItem(previewItem, ""); preview != nil {
			view.EquipPreview = preview
		}
	}
	return view
}

func viewSlotForSummary(item *invItem, itemTemplateID string, rules *Rules) string {
	if item != nil && item.slot != "" {
		return item.slot
	}
	if itemTemplateID != "" {
		if template, ok := rules.ItemTemplates[itemTemplateID]; ok {
			return template.Slot
		}
	}
	return ""
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
	view.SummaryLines = s.itemSummaryLines("", view.Slot, s.itemHandedness(item), s.statsForInventoryItem(item), view.Requirements, itemDefPtr(s.rules.Items[item.itemDefID]))
	view.SummaryLines = append(view.SummaryLines, s.setItemSummaryLines(item)...)
	s.annotateRequirementStatus(view.Requirements, func(status []RequirementStatusView, met *bool) {
		view.RequirementStatus = status
		view.RequirementsMet = met
	})
	if item.rollPayload != nil {
		s.annotateClassAffinityStatus(item.rollPayload, func(status []ClassAffinityStatusView) {
			view.ClassAffinityStatus = status
		})
	}
	if preview := s.equipPreviewForItem(item, view.Slot); preview != nil {
		view.EquipPreview = preview
	}
}

func (s *Sim) entityView(e *entity) EntityView {
	if e == nil {
		return EntityView{}
	}
	view := e.view()
	if e.kind == monsterEntity || e.kind == companionEntity {
		view.EffectIDs = sortedUniqueStrings(append(cloneStringSlice(e.effectIDs), s.eliteAuraEffectIDs(e)...))
	}
	if e.kind == companionEntity {
		view.CombatStats = s.companionCombatStatsView(e)
		if e.expiresTick > 0 && e.totalDurationTicks > 0 {
			view.RemainingTicks, view.TotalTicks = companionDurationTicks(e, s.tick)
		}
	}
	if e.kind == interactableEntity {
		if level := s.activeLevel(); level != nil {
			view.EliteObjective = level.eliteObjectiveChestIDs[e.id]
			view.QuestReward = level.questRewardChestIDs[e.id]
		}
	}
	if e.kind != lootEntity {
		return view
	}
	s.annotateRequirementStatus(view.Requirements, func(status []RequirementStatusView, met *bool) {
		view.RequirementStatus = status
		view.RequirementsMet = met
	})
	if e.rollPayload != nil {
		s.annotateClassAffinityStatus(e.rollPayload, func(status []ClassAffinityStatusView) {
			view.ClassAffinityStatus = status
		})
	}
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
		{"light_radius", current.LightRadius, preview.LightRadius},
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
	maxMana := character.MaxMana
	healthRegen := character.HealthRegenPerSecond
	manaRegen := character.ManaRegenPerSecond
	magicFindPercent := 0.0
	lightRadius := character.LightRadius
	blockPercent := 0.0
	weaponSpeed := 1.0
	itemSpeedPercent := 0.0
	hitChancePercent := character.HitChance * 100.0
	critChancePercent := character.CritChance * 100.0
	evadeChancePercent := 0.0
	moveSpeedPercent := 0.0

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
	maxManaSources := []StatBreakdownSourceView{{Label: "Magic", Value: character.MaxMana, Kind: "character_formula"}}
	healthRegenSources := []StatBreakdownSourceView{{Label: "Vitality", Value: character.HealthRegenPerSecond, Kind: "character_formula"}}
	manaRegenSources := []StatBreakdownSourceView{{Label: "Magic", Value: character.ManaRegenPerSecond, Kind: "character_formula"}}
	magicFindSources := []StatBreakdownSourceView{}
	lightRadiusSources := []StatBreakdownSourceView{{Label: "Class light radius", Value: character.LightRadius, Kind: "character_formula"}}
	blockSources := []StatBreakdownSourceView{}
	hitChanceSources := []StatBreakdownSourceView{{Label: "Dexterity", Value: character.HitChance, Kind: "character_formula"}}
	critChanceSources := []StatBreakdownSourceView{{Label: "Dexterity", Value: character.CritChance, Kind: "character_formula"}}
	evadeChanceSources := []StatBreakdownSourceView{}
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
		if value := baseStats["max_mana"]; value != 0 {
			maxMana += float64(value)
			maxManaSources = append(maxManaSources, StatBreakdownSourceView{Label: label, Value: float64(value), Kind: "equipment_base", ItemInstanceID: itemID})
		}
		if value := rolledStats["max_mana"]; value != 0 {
			maxMana += float64(value)
			maxManaSources = append(maxManaSources, StatBreakdownSourceView{Label: "Rolled max mana", Value: float64(value), Kind: "equipment_roll", ItemInstanceID: itemID})
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
		if value := rolledStats["movement_speed_percent"]; value != 0 {
			moveSpeedPercent += float64(value)
		}
		if value := rolledStats["hit_chance"]; value != 0 {
			hitChancePercent += float64(value)
			hitChanceSources = append(hitChanceSources, StatBreakdownSourceView{Label: "Rolled hit chance", Value: float64(value) / 100.0, Kind: "equipment_roll", ItemInstanceID: itemID})
		}
		if value := rolledStats["crit_chance"]; value != 0 {
			critChancePercent += float64(value)
			critChanceSources = append(critChanceSources, StatBreakdownSourceView{Label: "Rolled crit chance", Value: float64(value) / 100.0, Kind: "equipment_roll", ItemInstanceID: itemID})
		}
		if value := rolledStats["evade_chance"]; value != 0 {
			evadeChancePercent += float64(value)
			evadeChanceSources = append(evadeChanceSources, StatBreakdownSourceView{Label: "Rolled evade chance", Value: float64(value) / 100.0, Kind: "equipment_roll", ItemInstanceID: itemID})
		}
		if value := rolledStats["magic_find_percent"]; value != 0 {
			magicFindPercent += float64(value)
			magicFindSources = append(magicFindSources, StatBreakdownSourceView{Label: "Rolled Magic Find", Value: float64(value), Kind: "equipment_roll", ItemInstanceID: itemID})
		}
		if value := baseStats["light_radius"]; value != 0 {
			lightRadius += float64(value)
			lightRadiusSources = append(lightRadiusSources, StatBreakdownSourceView{Label: label, Value: float64(value), Kind: "equipment_base", ItemInstanceID: itemID})
		}
		if value := rolledStats["light_radius"]; value != 0 {
			lightRadius += float64(value)
			lightRadiusSources = append(lightRadiusSources, StatBreakdownSourceView{Label: "Rolled light radius", Value: float64(value), Kind: "equipment_roll", ItemInstanceID: itemID})
		}
	}
	applySetCombatStats(s.equippedSetBonusStats(), &damageMin, &damageMax, &armor, &maxHP, &maxMana, &healthRegen, &manaRegen, &blockPercent, &itemSpeedPercent, &hitChancePercent, &critChancePercent, &evadeChancePercent, &magicFindPercent, &damageMinSources, &damageMaxSources, &armorSources, &maxHPSources, &maxManaSources, &healthRegenSources, &manaRegenSources, &blockSources, &attackSpeedSources, &hitChanceSources, &critChanceSources, &evadeChanceSources, &magicFindSources)

	s.applyClassAffinityCombatStats(&damageMin, &damageMax, &itemSpeedPercent, &maxMana, &damageMinSources, &damageMaxSources, &attackSpeedSources, &maxManaSources)

	s.applyPassiveCombatStats(&damageMin, &damageMax, &armor, &maxHP, &maxMana, &healthRegen, &manaRegen, &blockPercent, &itemSpeedPercent, &hitChancePercent, &critChancePercent, &evadeChancePercent, &magicFindPercent, &lightRadius, &damageMinSources, &damageMaxSources, &armorSources, &maxHPSources, &maxManaSources, &healthRegenSources, &manaRegenSources, &blockSources, &attackSpeedSources, &hitChanceSources, &critChanceSources, &evadeChanceSources, &magicFindSources, &lightRadiusSources)

	armorEffectPercentRows := []StatBreakdownSourceView{}
	for _, stateKey := range sortedStringKeys(s.skillEffects) {
		effect := s.skillEffects[stateKey]
		if effect.TargetID != 0 && effect.TargetID != s.playerID {
			continue
		}
		if effect.EndsTick <= s.tick || effect.Percent <= 0 {
			continue
		}
		for _, stat := range effect.Stats {
			switch stat {
			case "armor":
				armorEffectPercentRows = append(armorEffectPercentRows, StatBreakdownSourceView{Label: s.skillEffectLabel(effect), Value: float64(effect.Percent), Kind: "skill_effect"})
			case "block_percent":
				blockPercent += float64(effect.Percent)
				blockSources = append(blockSources, StatBreakdownSourceView{Label: s.skillEffectLabel(effect), Value: float64(effect.Percent), Kind: "skill_effect"})
			}
		}
	}
	applyPercentSourceRows(&armor, &armorSources, armorEffectPercentRows)

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
		HitChance:            clampFloat(hitChancePercent/100.0, 0, 1),
		CritChance:           clampFloat(critChancePercent/100.0, 0, 1),
		CritDamage:           maxFloat(1, character.CritDamage),
		EvadeChance:          clampFloat(evadeChancePercent/100.0, 0, 1),
		Armor:                maxFloat(0, armor),
		BlockPercent:         maxFloat(0, blockPercent),
		AttackSpeed:          attackSpeed,
		AttackIntervalTicks:  attackInterval,
		MaxHP:                maxFloat(1, maxHP),
		MaxMana:              maxFloat(0, maxMana),
		HealthRegenPerSecond: maxFloat(0, healthRegen),
		ManaRegenPerSecond:   maxFloat(0, manaRegen),
		MagicFindPercent:     maxFloat(0, magicFindPercent),
		LightRadius:          maxFloat(0, lightRadius),
		MovementSpeedPercent: moveSpeedPercent,
	}
	if effective.DamageMax < effective.DamageMin {
		effective.DamageMax = effective.DamageMin
	}

	breakdowns := []StatBreakdownView{
		{Key: "damage_min", Value: effective.DamageMin, UncappedValue: effective.DamageMin, Cap: nil, Sources: damageMinSources},
		{Key: "damage_max", Value: effective.DamageMax, UncappedValue: effective.DamageMax, Cap: nil, Sources: damageMaxSources},
		{Key: "armor", Value: effective.Armor, UncappedValue: effective.Armor, Cap: nil, Sources: armorSources},
		{Key: "hit_chance", Value: effective.HitChance, UncappedValue: hitChancePercent / 100.0, Cap: floatPtr(1), Sources: hitChanceSources},
		{Key: "crit_chance", Value: effective.CritChance, UncappedValue: critChancePercent / 100.0, Cap: floatPtr(1), Sources: critChanceSources},
		{Key: "evade_chance", Value: effective.EvadeChance, UncappedValue: evadeChancePercent / 100.0, Cap: floatPtr(1), Sources: evadeChanceSources},
		{Key: "attack_speed", Value: effective.AttackSpeed, UncappedValue: uncappedAttackSpeed, Cap: floatPtr(s.rules.Combat.MaxEffectiveAttackSpeed), Sources: attackSpeedSources},
		{Key: "attack_interval_ticks", Value: float64(effective.AttackIntervalTicks), UncappedValue: float64(effective.AttackIntervalTicks), Cap: nil, Sources: attackIntervalSources},
		{Key: "max_hp", Value: effective.MaxHP, UncappedValue: effective.MaxHP, Cap: nil, Sources: maxHPSources},
		{Key: "max_mana", Value: effective.MaxMana, UncappedValue: effective.MaxMana, Cap: nil, Sources: maxManaSources},
		{Key: "health_regen_per_second", Value: effective.HealthRegenPerSecond, UncappedValue: effective.HealthRegenPerSecond, Cap: nil, Sources: healthRegenSources},
		{Key: "mana_regen_per_second", Value: effective.ManaRegenPerSecond, UncappedValue: effective.ManaRegenPerSecond, Cap: nil, Sources: manaRegenSources},
		{Key: "magic_find_percent", Value: effective.MagicFindPercent, UncappedValue: effective.MagicFindPercent, Cap: nil, Sources: magicFindSources},
		{Key: "light_radius", Value: effective.LightRadius, UncappedValue: effective.LightRadius, Cap: nil, Sources: lightRadiusSources},
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
		elementalBonus := elementalBonusDamage(item.rollPayload.Stats)
		baseMinInt := template.BaseStats["damage_min"]
		baseMaxInt := template.BaseStats["damage_max"]
		return float64(baseMinInt), float64(baseMaxInt), float64(totalMin-baseMinInt+elementalBonus), float64(totalMax-baseMaxInt+elementalBonus), label, itemID, true
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
	hitChance := def.effectiveHitChance(s.rules.Combat)
	if monster.monsterHitChance > 0 {
		hitChance = monster.monsterHitChance
	}
	critChance := def.effectiveCritChance(s.rules.Combat)
	if monster.monsterCritChance > 0 {
		critChance = monster.monsterCritChance
	}
	armor := float64(def.Armor)
	if monster.monsterArmor > 0 {
		armor = monster.monsterArmor
	}
	blockPercent := clampFloat(float64(def.BlockPercent), 0, float64(s.rules.Combat.BlockCap))
	if monster.monsterBlockPercent > 0 {
		blockPercent = clampFloat(monster.monsterBlockPercent, 0, float64(s.rules.Combat.BlockCap))
	}
	return effectiveCombatStats{
		DamageMin:    float64(damage.Min),
		DamageMax:    float64(damage.Max),
		HitChance:    hitChance,
		CritChance:   critChance,
		CritDamage:   def.effectiveCritDamage(s.rules.Combat),
		Armor:        armor,
		BlockPercent: blockPercent,
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
		if it != nil && it.instanceID == id {
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

func sortedUint64Keys[V any](m map[uint64]V) []uint64 {
	ids := make([]uint64, 0, len(m))
	for id := range m {
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
		ActiveWeaponSet:   defaultWeaponSet,
		WeaponSets:        weaponSetViewsFromMaps(newWeaponSetMaps()),
		Hotbar:            []HotbarSlotView{},
		InventoryRows:     baseInventoryRows,
		InventoryCapacity: inventoryCapacityForRows(baseInventoryRows),
		Gold:              0,
		StashItems:        []StashItemView{},
		StashGold:         0,
		StashCapacity:     defaultStashCapacity,
		ResourceWallet:    []ResourceAmountView{},
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
	entities := s.visibleEntityViewsForPlayer(ps)

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
	weaponSets := s.weaponSetViews()
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
		ActiveWeaponSet:       s.activeWeaponSet,
		WeaponSets:            weaponSets,
		HotbarCapacity:        s.hotbarCapacity(),
		Hotbar:                s.hotbarView(),
		InventoryRows:         s.inventoryRows(),
		InventoryCapacity:     s.inventoryCapacity(),
		Gold:                  s.gold,
		StashItems:            stashItems,
		StashGold:             s.stashGold,
		StashCapacity:         s.stashCapacity,
		ResourceWallet:        s.ResourceWalletView(),
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
		view := WallView{
			ID:       wallID(level.levelNum, i),
			Position: wall.pos,
			Size:     wall.size,
			Source:   source,
		}
		if kind := wall.obstacleKind(); kind != obstacleKindWall {
			view.Kind = kind
		}
		if wall.blocksLOS != nil {
			view.BlocksLineOfSight = boolPtr(*wall.blocksLOS)
		}
		out = append(out, view)
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
	case playerEntity, monsterEntity, companionEntity:
		hp, maxHP := e.hp, e.maxHP
		ev.HP = &hp
		ev.MaxHP = &maxHP
		if e.kind == playerEntity {
			mana, maxMana := e.mana, e.maxMana
			ev.Mana = &mana
			ev.MaxMana = &maxMana
			ev.CharacterID = e.characterID
			ev.DisplayName = e.displayName
			ev.EffectIDs = cloneStringSlice(e.effectIDs)
			if e.visualScale > 0 {
				ev.VisualScale = e.visualScale
			}
		}
		if e.kind == monsterEntity || e.kind == companionEntity {
			e.applyMonsterLikeViewFields(&ev)
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
			ev.ItemLevel = e.rollPayload.ItemLevel
			ev.RolledStats = cloneIntMap(e.rollPayload.Stats)
			ev.Requirements = cloneIntMap(e.rollPayload.Requirements)
			ev.EffectIDs = cloneStringSlice(e.rollPayload.EffectIDs)
		}
	case interactableEntity:
		ev.InteractableDefID = e.interactableDefID
		ev.State = e.state
		if e.interactableDefID == heroCorpseDefID {
			ev.CorpseCharacterID = e.corpseCharacterID
			ev.CorpseName = e.corpseName
			if e.corpseLevel > 0 {
				ev.CorpseLevel = e.corpseLevel
			}
			count := e.corpseItemCount
			ev.CorpseItemCount = &count
		}
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
