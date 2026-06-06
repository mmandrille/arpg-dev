package game

import (
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
	interactableTransitionAscend   = "ascend"
	interactableTransitionDescend  = "descend"
	interactableTransitionWaypoint = "waypoint"
	stairsDownDefID                = "stairs_down"
	stairsUpDefID                  = "stairs_up"
	teleporterDefID                = "teleporter"
	worldModeMultiLevel            = "multi_level"
	attackModeMelee                = "melee"
	attackModeRanged               = "ranged"
	trainingArrowProjectileDefID   = "training_arrow"
	weaponSlot                     = "weapon"
	lootInteractionRadius          = 0.35
	interactableInteractionRadius  = 0.50
	meleeRangeEpsilon              = 0.000001
	projectileRadius               = 0.10
	tickDuration                   = 0.05
)

// DefaultWorldID is the compatibility world used when callers do not choose a
// preset explicitly.
const DefaultWorldID = "vertical_slice"

const (
	entryLevel = -1
	levelZero  = 0
)

// entity is the internal mutable scene entity.
type entity struct {
	id                uint64
	kind              string
	pos               Vec2
	hp                int
	maxHP             int
	monsterDefID      string
	itemDefID         string
	interactableDefID string
	state             string
	lootTable         string
	ownerID           uint64
	targetID          uint64
	projectileDefID   string
	dir               Vec2
	speed             float64
	traveled          float64
	maxDistance       float64
	damageRange       DamageRange
	sourceMsgID       string
	sourceCorrID      string
	spawnTick         uint64
	spawnPos          Vec2
	aiMode            string
}

// invItem is an internal inventory item.
type invItem struct {
	instanceID uint64
	itemDefID  string
	slot       string
	equipped   bool
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
	pos  Vec2
	size Vec2
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

	levels                map[int]*LevelState
	currentLevel          int
	multiLevel            bool
	entities              map[uint64]*entity
	walls                 []wallObstacle
	move                  *activeMove
	autoNav               *autoNavState
	inventory             []*invItem
	equipped              map[string]uint64 // slot -> instanceID (0 = none)
	discoveredTeleporters map[int]bool
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
	world, ok := rules.Worlds[worldID]
	if !ok {
		return nil, ErrUnknownWorld{WorldID: worldID}
	}
	s := &Sim{
		sessionID:             sessionID,
		seed:                  seed,
		rng:                   NewRNG(SeedToUint64(seed)),
		rules:                 rules,
		nextID:                baseEntityID,
		levels:                make(map[int]*LevelState),
		currentLevel:          levelZero,
		multiLevel:            world.Mode == worldModeMultiLevel,
		equipped:              map[string]uint64{weaponSlot: 0},
		discoveredTeleporters: make(map[int]bool),
	}

	if s.multiLevel {
		s.currentLevel = entryLevel
		nav := dungeonNavigation(rules.Navigation, rules.DungeonGeneration)
		level := newLevelState(entryLevel, &nav)
		s.levels[entryLevel] = level
		player := &entity{kind: playerEntity, pos: rules.DungeonGeneration.PlayerSpawn, hp: playerStartHP, maxHP: playerStartHP}
		player.id = s.alloc()
		s.playerID = player.id
		level.entities[player.id] = player
		if err := s.populateDungeonLevel(level); err != nil {
			return nil, err
		}
		s.syncCompatibilityFields()
		return s, nil
	}

	level := newLevelState(levelZero, &rules.Navigation)
	s.levels[levelZero] = level

	player := &entity{kind: playerEntity, pos: world.Player.Position, hp: playerStartHP, maxHP: playerStartHP}
	player.id = s.alloc()
	s.playerID = player.id
	level.entities[player.id] = player

	for _, preset := range world.Entities {
		switch preset.Type {
		case monsterEntity:
			def := rules.Monsters[preset.MonsterDefID]
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
			monster.id = s.alloc()
			level.entities[monster.id] = monster
		case lootEntity:
			loot := &entity{kind: lootEntity, pos: preset.Position, itemDefID: preset.ItemDefID}
			loot.id = s.alloc()
			level.entities[loot.id] = loot
		case wallEntity:
			level.walls = append(level.walls, wallObstacle{pos: preset.Position, size: preset.Size})
		case interactableEntity:
			def := rules.Interactables[preset.InteractableDefID]
			interactable := &entity{
				kind:              interactableEntity,
				pos:               preset.Position,
				interactableDefID: preset.InteractableDefID,
				state:             def.InitialState,
			}
			interactable.id = s.alloc()
			level.entities[interactable.id] = interactable
		default:
			return nil, ErrUnknownWorldEntity{WorldID: worldID, EntityType: preset.Type}
		}
	}

	s.syncCompatibilityFields()
	return s, nil
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
	InstanceID string
	ItemDefID  string
	Slot       string
	Equipped   bool
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
		it := &invItem{instanceID: id, itemDefID: p.ItemDefID, slot: p.Slot, equipped: p.Equipped}
		s.inventory = append(s.inventory, it)
		if p.Equipped && p.Slot != "" {
			s.equipped[p.Slot] = id
		}
		if id >= s.nextID {
			s.nextID = id + 1
		}
	}
}

func (s *Sim) alloc() uint64 {
	id := s.nextID
	s.nextID++
	return id
}

// CurrentTick returns the next tick to be processed.
func (s *Sim) CurrentTick() uint64 { return s.tick }

// Input is a decoded client intent applied to a specific tick.
type Input struct {
	MessageID     string
	CorrelationID string
	Sequence      int64
	Type          string
	Move          *MoveIntent
	MoveTo        *MoveToIntent
	Action        *ActionIntent
	Descend       *DescendIntent
	Ascend        *AscendIntent
	Teleport      *TeleportIntent
	Equip         *EquipIntent
	Unequip       *UnequipIntent
	Drop          *DropIntent
	Use           *UseIntent
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
	Tick    uint64
	Level   int
	Changes []Change
	Events  []Event
	Acks    []Ack
	Rejects []Reject
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
	return results[len(results)-1]
}

// TickResults processes the inputs stamped for the current tick (already
// ordered by the runner as (sequence, message_id)), applies continuous
// movement, advances the tick counter, and returns one or more scoped results.
func (s *Sim) TickResults(inputs []Input) []TickResult {
	// Changes/Events are always non-nil so they marshal as [] (not null),
	// satisfying the state_delta schema.
	res := TickResult{Tick: s.tick, Level: s.currentLevel, Changes: []Change{}, Events: []Event{}}
	var transitionArrival *TickResult
	for _, in := range inputs {
		if in.Type == "descend_intent" || in.Type == "ascend_intent" || in.Type == "teleport_intent" {
			if transitionArrival == nil {
				transitionArrival = s.handleLevelTravel(in, &res)
			} else {
				res.reject(in.MessageID, "invalid_level")
			}
			continue
		}
		s.applyInput(in, &res)
	}
	if transitionArrival != nil {
		s.tick++
		s.syncCompatibilityFields()
		return []TickResult{res, *transitionArrival}
	}
	s.applyMovement(&res)
	s.advanceMonsterMovement(&res)
	s.advanceProjectiles(&res)
	s.tick++
	s.syncCompatibilityFields()
	return []TickResult{res}
}

func (s *Sim) applyInput(in Input, res *TickResult) {
	if in.Type != "client_ready" && s.playerDead() {
		switch in.Type {
		case "move_intent", "move_to_intent", "action_intent", "descend_intent", "ascend_intent", "teleport_intent", "equip_intent", "unequip_intent", "drop_intent", "use_intent":
			res.reject(in.MessageID, "player_dead")
			return
		}
	}
	switch in.Type {
	case "client_ready":
		res.ack(in.MessageID)
	case "move_intent":
		s.handleMove(in, res)
	case "move_to_intent":
		s.handleMoveTo(in, res)
	case "action_intent":
		s.handleAction(in, res)
	case "descend_intent", "ascend_intent", "teleport_intent":
		if arrival := s.handleLevelTravel(in, res); arrival != nil {
			res.Changes = append(res.Changes, arrival.Changes...)
			res.Events = append(res.Events, arrival.Events...)
		}
	case "equip_intent":
		s.handleEquip(in, res)
	case "unequip_intent":
		s.handleUnequip(in, res)
	case "drop_intent":
		s.handleDrop(in, res)
	case "use_intent":
		s.handleUse(in, res)
	default:
		res.reject(in.MessageID, "unknown_type")
	}
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
	if level, ok := s.levels[levelNum]; ok {
		return level, nil
	}
	nav := dungeonNavigation(s.rules.Navigation, s.rules.DungeonGeneration)
	level := newLevelState(levelNum, &nav)
	s.levels[levelNum] = level
	if err := s.populateDungeonLevel(level); err != nil {
		delete(s.levels, levelNum)
		return nil, err
	}
	return level, nil
}

func (s *Sim) populateDungeonLevel(level *LevelState) error {
	gen, err := GenerateDungeonLevel(s.seed, level.levelNum, s.rules.DungeonGeneration)
	if err != nil {
		return err
	}
	level.walls = gen.walls
	for _, stair := range gen.stairs {
		def := s.rules.Interactables[stair.defID]
		e := &entity{
			kind:              interactableEntity,
			pos:               stair.pos,
			interactableDefID: stair.defID,
			state:             def.InitialState,
		}
		e.id = s.alloc()
		level.entities[e.id] = e
	}
	for _, teleporter := range gen.teleporters {
		def := s.rules.Interactables[teleporter.defID]
		e := &entity{
			kind:              interactableEntity,
			pos:               teleporter.pos,
			interactableDefID: teleporter.defID,
			state:             def.InitialState,
		}
		e.id = s.alloc()
		level.entities[e.id] = e
	}
	for _, generated := range gen.loot {
		if _, ok := s.rules.Items[generated.itemDefID]; !ok {
			return fmt.Errorf("game: generate dungeon level %d: unknown loot item %s", level.levelNum, generated.itemDefID)
		}
		loot := &entity{kind: lootEntity, pos: generated.pos, itemDefID: generated.itemDefID}
		loot.id = s.alloc()
		level.entities[loot.id] = loot
	}
	return nil
}

func (s *Sim) handleMove(in Input, res *TickResult) {
	if in.Move == nil {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	dir := normalize(in.Move.Direction)
	dur := in.Move.DurationTicks
	if dur < 1 {
		dur = 1
	}
	if dir.X != 0 || dir.Y != 0 {
		s.clearAutoNav()
		s.activeLevel().move = &activeMove{dir: dir, remaining: dur}
	}
	res.ack(in.MessageID)
}

func (s *Sim) handleMoveTo(in Input, res *TickResult) {
	if in.MoveTo == nil || !finiteVec2(in.MoveTo.Position) {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil {
		res.reject(in.MessageID, "player_dead")
		return
	}
	if distance(player.pos, in.MoveTo.Position) <= s.activeNav().StopDistance {
		s.clearAutoNav()
		res.ack(in.MessageID)
		return
	}
	steps, ok := PlanPath(s.activeNav(), player.pos, in.MoveTo.Position, s.buildBlockedFn())
	if !ok {
		res.reject(in.MessageID, "no_path")
		return
	}
	if len(steps) > s.activeNav().MaxAutoSteps {
		res.reject(in.MessageID, "path_too_long")
		return
	}
	s.activeLevel().move = nil
	s.activeLevel().autoNav = &autoNavState{steps: steps, sourceMsgID: in.MessageID, sourceCorrID: in.CorrelationID}
	res.ack(in.MessageID)
}

func (s *Sim) handleAction(in Input, res *TickResult) {
	if in.Action == nil {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	target := s.findEntity(in.Action.TargetID)
	if target == nil || !s.actionable(target) {
		res.reject(in.MessageID, "invalid_target")
		return
	}
	if s.inDispatchRange(target) {
		s.dispatchAction(target, in, res, true)
		return
	}

	_, steps, ok := s.findApproachGoal(target)
	if !ok {
		res.reject(in.MessageID, "no_path")
		return
	}
	if len(steps) > s.activeNav().MaxAutoSteps {
		res.reject(in.MessageID, "path_too_long")
		return
	}
	s.activeLevel().move = nil
	s.activeLevel().autoNav = &autoNavState{
		steps:         steps,
		pendingAction: &ActionIntent{TargetID: in.Action.TargetID},
		sourceMsgID:   in.MessageID,
		sourceCorrID:  in.CorrelationID,
	}
	res.ack(in.MessageID)
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
	// Always consume two draws (hit, damage) so the RNG stream is independent
	// of the hit/miss branch (base_hit_chance is 1.0 in v0).
	hitDraw := s.rng.Next()
	dmg := s.rollDamage()
	hit := s.rules.Combat.BaseHitChance >= 1.0 ||
		float64(hitDraw%10000)/10000.0 < s.rules.Combat.BaseHitChance

	if ack {
		res.ack(in.MessageID)
	}
	if !hit {
		res.Events = append(res.Events, Event{EventType: "attack_missed", EntityID: idStr(target.id), CorrelationID: in.CorrelationID})
		return
	}

	target.hp -= dmg
	if target.hp < 0 {
		target.hp = 0
	}
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(target.view())})
	res.Events = append(res.Events, Event{EventType: "monster_damaged", EntityID: idStr(target.id), CorrelationID: in.CorrelationID, Damage: intPtr(dmg)})

	if target.hp == 0 {
		res.Events = append(res.Events, Event{EventType: "monster_killed", EntityID: idStr(target.id), CorrelationID: in.CorrelationID})
		s.dropLoot(target, in.CorrelationID, res)
	}
	s.retaliate(target, in.CorrelationID, res)
}

func (s *Sim) fireProjectile(target *entity, in Input, res *TickResult, ack bool) {
	if s.playerProjectileInFlight() {
		res.reject(in.MessageID, "projectile_busy")
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil {
		res.reject(in.MessageID, "player_dead")
		return
	}
	weapon, ok := s.equippedWeaponDef()
	if !ok || weapon.AttackMode != attackModeRanged || weapon.ProjectileSpeed == nil {
		res.reject(in.MessageID, "invalid_target")
		return
	}
	dir := normalize(Vec2{X: target.pos.X - player.pos.X, Y: target.pos.Y - player.pos.Y})
	if dir.X == 0 && dir.Y == 0 {
		dir = Vec2{X: 1}
	}
	maxDistance := s.playerActionReach()
	projectile := &entity{
		kind:            projectileEntity,
		pos:             player.pos,
		ownerID:         player.id,
		targetID:        target.id,
		projectileDefID: trainingArrowProjectileDefID,
		dir:             dir,
		speed:           *weapon.ProjectileSpeed,
		maxDistance:     maxDistance,
		damageRange:     *weapon.Damage,
		sourceMsgID:     in.MessageID,
		sourceCorrID:    in.CorrelationID,
		spawnTick:       s.tick,
	}
	projectile.id = s.alloc()
	s.activeLevel().entities[projectile.id] = projectile
	res.Changes = append(res.Changes, Change{Op: OpEntitySpawn, Entity: ptrEntityView(projectile.view())})
	if ack {
		res.ack(in.MessageID)
	}
}

func (s *Sim) dropLoot(monster *entity, corr string, res *TickResult) {
	drops := s.rules.LootDrops(monster.lootTable, s.rng)
	var clusterAnchor Vec2
	clusterReady := false

	for i, itemDefID := range drops {
		var dropPos Vec2
		var ok bool

		if i == 0 {
			dropPos, ok = s.findEntityLootDropPosition(monster.pos, s.targetInteractionRadius(monster))
			if !ok {
				dropPos = monster.pos
			}
			clusterAnchor = dropPos
			clusterReady = true
		} else if clusterReady {
			dropPos, ok = s.findClusterLootDropPosition(clusterAnchor, i)
			if !ok {
				dropPos, ok = s.findEntityLootDropPosition(monster.pos, s.targetInteractionRadius(monster))
				if !ok {
					dropPos = monster.pos
				}
			}
		} else {
			dropPos, ok = s.findEntityLootDropPosition(monster.pos, s.targetInteractionRadius(monster))
			if !ok {
				dropPos = monster.pos
			}
		}

		loot := &entity{kind: lootEntity, pos: dropPos, itemDefID: itemDefID}
		loot.id = s.alloc()
		s.activeLevel().entities[loot.id] = loot
		res.Changes = append(res.Changes, Change{Op: OpEntitySpawn, Entity: ptrEntityView(loot.view())})
		res.Events = append(res.Events, Event{EventType: "loot_dropped", EntityID: idStr(loot.id), CorrelationID: corr})
	}
}

func (s *Sim) retaliate(monster *entity, corr string, res *TickResult) {
	def := s.rules.Monsters[monster.monsterDefID]
	if def.RetaliationDamage == nil {
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil || player.hp <= 0 {
		return
	}
	dmg := s.rollRange(*def.RetaliationDamage)
	player.hp -= dmg
	if player.hp < 0 {
		player.hp = 0
	}
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(player.view())})
	eventType := "player_damaged"
	if player.hp == 0 {
		eventType = "player_killed"
	}
	res.Events = append(res.Events, Event{EventType: eventType, EntityID: idStr(player.id), CorrelationID: corr, Damage: intPtr(dmg)})
}

func (s *Sim) pickUpTarget(e *entity, in Input, res *TickResult, ack bool) {
	delete(s.activeLevel().entities, e.id)
	res.Changes = append(res.Changes, Change{Op: OpEntityRemove, EntityID: idStr(e.id)})

	item := &invItem{
		instanceID: s.alloc(),
		itemDefID:  e.itemDefID,
		slot:       s.rules.Items[e.itemDefID].Slot,
		equipped:   false,
	}
	s.inventory = append(s.inventory, item)
	res.Changes = append(res.Changes, Change{Op: OpInventoryAdd, Item: ptrItemView(item.view())})
	res.Events = append(res.Events, Event{EventType: "item_picked_up", EntityID: idStr(item.instanceID), CorrelationID: in.CorrelationID})
	if ack {
		res.ack(in.MessageID)
	}
}

func (s *Sim) activateInteractable(e *entity, in Input, res *TickResult, ack bool) {
	if e.interactableDefID == teleporterDefID {
		s.activateTeleporter(e, in, res, ack)
		return
	}
	e.state = interactableOpen
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(e.view())})
	res.Events = append(res.Events, Event{EventType: "interactable_activated", EntityID: idStr(e.id), CorrelationID: in.CorrelationID})
	if ack {
		res.ack(in.MessageID)
	}
}

func (s *Sim) activateTeleporter(e *entity, in Input, res *TickResult, ack bool) {
	if !s.multiLevel {
		res.reject(in.MessageID, "not_dungeon_world")
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
}

func (s *Sim) handleLevelTravel(in Input, res *TickResult) *TickResult {
	if in.Type == "teleport_intent" {
		return s.handleTeleport(in, res)
	}
	return s.handleTransition(in, res)
}

func (s *Sim) handleTransition(in Input, res *TickResult) *TickResult {
	if !s.multiLevel {
		res.reject(in.MessageID, "not_dungeon_world")
		return nil
	}
	if s.playerDead() {
		res.reject(in.MessageID, "player_dead")
		return nil
	}

	var (
		stairDefID string
		destLevel  int
		arrivalDef string
	)
	switch in.Type {
	case "descend_intent":
		if in.Descend == nil {
			res.reject(in.MessageID, "invalid_payload")
			return nil
		}
		stairDefID = stairsDownDefID
		destLevel = s.currentLevel - 1
		arrivalDef = stairsUpDefID
	case "ascend_intent":
		if in.Ascend == nil {
			res.reject(in.MessageID, "invalid_payload")
			return nil
		}
		if s.currentLevel >= entryLevel {
			res.reject(in.MessageID, "already_at_entry")
			return nil
		}
		stairDefID = stairsUpDefID
		destLevel = s.currentLevel + 1
		arrivalDef = stairsDownDefID
	default:
		res.reject(in.MessageID, "invalid_payload")
		return nil
	}
	if destLevel >= levelZero {
		res.reject(in.MessageID, "invalid_level")
		return nil
	}

	current := s.activeLevel()
	player := current.entities[s.playerID]
	if player == nil {
		res.reject(in.MessageID, "player_dead")
		return nil
	}
	stair := s.findReachableStair(current, stairDefID, player.pos)
	if stair == nil {
		res.reject(in.MessageID, "no_stair_in_range")
		return nil
	}

	dest, err := s.ensureDungeonLevel(destLevel)
	if err != nil {
		res.reject(in.MessageID, "invalid_level")
		return nil
	}
	arrival := s.findStair(dest, arrivalDef)
	if arrival == nil {
		res.reject(in.MessageID, "invalid_level")
		return nil
	}
	return s.movePlayerToLevel(in, res, current, dest, arrival.pos)
}

func (s *Sim) handleTeleport(in Input, res *TickResult) *TickResult {
	if in.Teleport == nil {
		res.reject(in.MessageID, "invalid_payload")
		return nil
	}
	if !s.multiLevel {
		res.reject(in.MessageID, "not_dungeon_world")
		return nil
	}
	if s.playerDead() {
		res.reject(in.MessageID, "player_dead")
		return nil
	}
	targetLevel := in.Teleport.TargetLevel
	if targetLevel >= levelZero {
		res.reject(in.MessageID, "invalid_level")
		return nil
	}
	if !s.discoveredTeleporters[s.currentLevel] {
		res.reject(in.MessageID, "teleporter_not_discovered")
		return nil
	}
	if !s.discoveredTeleporters[targetLevel] {
		res.reject(in.MessageID, "target_level_not_discovered")
		return nil
	}
	current := s.activeLevel()
	player := current.entities[s.playerID]
	if player == nil {
		res.reject(in.MessageID, "player_dead")
		return nil
	}
	if s.findReachableTeleporter(current, player.pos) == nil {
		res.reject(in.MessageID, "no_teleporter_in_range")
		return nil
	}
	dest, err := s.ensureDungeonLevel(targetLevel)
	if err != nil {
		res.reject(in.MessageID, "invalid_level")
		return nil
	}
	arrival := s.findTeleporter(dest)
	if arrival == nil {
		res.reject(in.MessageID, "invalid_level")
		return nil
	}
	return s.movePlayerToLevel(in, res, current, dest, arrival.pos)
}

func (s *Sim) movePlayerToLevel(in Input, res *TickResult, current, dest *LevelState, arrivalPos Vec2) *TickResult {
	player := current.entities[s.playerID]
	if player == nil {
		res.reject(in.MessageID, "player_dead")
		return nil
	}
	fromLevel := s.currentLevel
	destLevel := dest.levelNum
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
	if s.multiLevel && destLevel < levelZero && !s.discoveredTeleporters[destLevel] {
		arrivalRes.Changes = append(arrivalRes.Changes, Change{
			Op:         OpTeleporterDiscoveryUpdate,
			Level:      destLevel,
			Discovered: false,
		})
	}
	for _, id := range sortedEntityIDs(dest.entities) {
		arrivalRes.Changes = append(arrivalRes.Changes, Change{Op: OpEntitySpawn, Entity: ptrEntityView(dest.entities[id].view())})
	}
	return &arrivalRes
}

func (s *Sim) handleEquip(in Input, res *TickResult) {
	if in.Equip == nil {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	item := s.findItem(in.Equip.ItemInstanceID)
	if item == nil {
		res.reject(in.MessageID, "not_in_inventory")
		return
	}
	def, ok := s.rules.Items[item.itemDefID]
	if !ok || !def.Equippable {
		res.reject(in.MessageID, "not_equippable")
		return
	}
	if in.Equip.Slot != def.Slot {
		res.reject(in.MessageID, "wrong_slot")
		return
	}

	// Unequip whatever currently occupies the slot.
	if prevID := s.equipped[in.Equip.Slot]; prevID != 0 && prevID != item.instanceID {
		if prev := s.findItemByID(prevID); prev != nil {
			prev.equipped = false
			res.Changes = append(res.Changes, Change{Op: OpInventoryUpdate, Item: ptrItemView(prev.view())})
		}
	}

	item.equipped = true
	s.equipped[in.Equip.Slot] = item.instanceID

	res.Changes = append(res.Changes, Change{Op: OpInventoryUpdate, Item: ptrItemView(item.view())})
	idCopy := idStr(item.instanceID)
	res.Changes = append(res.Changes, Change{Op: OpEquippedUpdate, Slot: in.Equip.Slot, ItemInstanceID: &idCopy})
	res.Events = append(res.Events, Event{EventType: "item_equipped", EntityID: idCopy, CorrelationID: in.CorrelationID})
	res.ack(in.MessageID)
}

func (s *Sim) handleUnequip(in Input, res *TickResult) {
	if in.Unequip == nil || in.Unequip.Slot != weaponSlot {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	instanceID := s.equipped[in.Unequip.Slot]
	if instanceID == 0 {
		res.reject(in.MessageID, "slot_empty")
		return
	}
	item := s.findItemByID(instanceID)
	if item == nil {
		res.reject(in.MessageID, "slot_empty")
		return
	}
	item.equipped = false
	s.equipped[in.Unequip.Slot] = 0
	res.Changes = append(res.Changes, Change{Op: OpInventoryUpdate, Item: ptrItemView(item.view())})
	res.Changes = append(res.Changes, Change{Op: OpEquippedUpdate, Slot: in.Unequip.Slot, ItemInstanceID: nil})
	idCopy := idStr(item.instanceID)
	res.Events = append(res.Events, Event{EventType: "item_unequipped", EntityID: idCopy, CorrelationID: in.CorrelationID})
	res.ack(in.MessageID)
}

func (s *Sim) handleDrop(in Input, res *TickResult) {
	if in.Drop == nil || in.Drop.ItemInstanceID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	item := s.findItem(in.Drop.ItemInstanceID)
	if item == nil {
		res.reject(in.MessageID, "not_in_inventory")
		return
	}
	dropPos, ok := s.findDropPosition()
	if !ok {
		res.reject(in.MessageID, "no_drop_space")
		return
	}

	if item.equipped {
		for slot, instanceID := range s.equipped {
			if instanceID == item.instanceID {
				s.equipped[slot] = 0
				res.Changes = append(res.Changes, Change{Op: OpEquippedUpdate, Slot: slot, ItemInstanceID: nil})
			}
		}
	}

	removedID := idStr(item.instanceID)
	itemDefID := item.itemDefID
	s.removeItemByID(item.instanceID)
	res.Changes = append(res.Changes, Change{Op: OpInventoryRemove, ItemInstanceID: &removedID})

	loot := &entity{kind: lootEntity, pos: dropPos, itemDefID: itemDefID}
	loot.id = s.alloc()
	s.activeLevel().entities[loot.id] = loot
	res.Changes = append(res.Changes, Change{Op: OpEntitySpawn, Entity: ptrEntityView(loot.view())})
	res.Events = append(res.Events, Event{
		EventType:      "item_dropped",
		EntityID:       idStr(loot.id),
		CorrelationID:  in.CorrelationID,
		ItemInstanceID: removedID,
	})
	res.ack(in.MessageID)
}

func (s *Sim) handleUse(in Input, res *TickResult) {
	if in.Use == nil || in.Use.ItemInstanceID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil || player.hp <= 0 {
		res.reject(in.MessageID, "player_dead")
		return
	}
	item := s.findItem(in.Use.ItemInstanceID)
	if item == nil {
		res.reject(in.MessageID, "not_in_inventory")
		return
	}
	def, ok := s.rules.Items[item.itemDefID]
	if !ok || def.Category != "consumable" {
		res.reject(in.MessageID, "not_consumable")
		return
	}
	if def.Heal == nil {
		res.reject(in.MessageID, "not_usable")
		return
	}
	if player.hp >= player.maxHP {
		res.reject(in.MessageID, "already_full_hp")
		return
	}

	rolled := s.rollRange(*def.Heal)
	heal := rolled
	if player.hp+heal > player.maxHP {
		heal = player.maxHP - player.hp
	}
	if heal <= 0 {
		res.reject(in.MessageID, "already_full_hp")
		return
	}

	removedID := idStr(item.instanceID)
	s.removeItemByID(item.instanceID)
	res.Changes = append(res.Changes, Change{Op: OpInventoryRemove, ItemInstanceID: &removedID})

	player.hp += heal
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(player.view())})
	res.Events = append(res.Events, Event{
		EventType:      "item_used",
		EntityID:       idStr(player.id),
		CorrelationID:  in.CorrelationID,
		Heal:           intPtr(heal),
		ItemInstanceID: removedID,
	})
	res.Events = append(res.Events, Event{
		EventType:      "player_healed",
		EntityID:       idStr(player.id),
		CorrelationID:  in.CorrelationID,
		Heal:           intPtr(heal),
		ItemInstanceID: removedID,
	})
	res.ack(in.MessageID)
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
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(player.view())})
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
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(player.view())})
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
	if s.playerDead() {
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil {
		return
	}
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
		prevMode := monster.aiMode
		s.updateMonsterAIMode(monster, player, def, prevMode, res)
		if monster.aiMode == monsterAIModeIdle {
			continue
		}
		goal, hasGoal := s.monsterMovementGoal(monster, player)
		if !hasGoal {
			continue
		}
		if distance(monster.pos, goal) <= nav.StopDistance {
			continue
		}
		blocked := s.buildMonsterBlockedFn(monster.id)
		steps, ok := PlanPath(nav, monster.pos, goal, blocked)
		if !ok || len(steps) == 0 {
			continue
		}
		moveSpeed := def.effectiveMoveSpeed(nav)
		before := monster.pos
		monster.pos = s.resolveMonsterMovement(monster, Vec2{
			X: steps[0].X * moveSpeed,
			Y: steps[0].Y * moveSpeed,
		})
		if monster.pos != before {
			res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(monster.view())})
		}
	}
}

func (s *Sim) updateMonsterAIMode(monster *entity, player *entity, def MonsterDef, prevMode string, res *TickResult) {
	nav := s.activeNav()
	distPlayer := distance(monster.pos, player.pos)
	distPlayerFromSpawn := distance(player.pos, monster.spawnPos)

	if def.LeashRadius > 0 && distPlayerFromSpawn > def.LeashRadius {
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

func (s *Sim) monsterMovementGoal(monster *entity, player *entity) (Vec2, bool) {
	nav := s.activeNav()
	switch monster.aiMode {
	case monsterAIModeChase:
		stopDist := playerRadius + monsterRadius
		if distance(monster.pos, player.pos) <= stopDist+1e-9 {
			return Vec2{}, false
		}

		return s.findMonsterChaseGoal(monster, player)
	case monsterAIModeReturn:
		if distance(monster.pos, monster.spawnPos) <= nav.StopDistance {
			return Vec2{}, false
		}

		return monster.spawnPos, true
	default:
		return Vec2{}, false
	}
}

func (s *Sim) findMonsterChaseGoal(monster *entity, player *entity) (Vec2, bool) {
	nav := s.activeNav()
	playerCell := worldToGrid(nav, player.pos)
	blocked := s.buildMonsterBlockedFn(monster.id)
	maxReach := playerRadius + monsterRadius + nav.CellSize
	maxRadius := maxInt(nav.GridBounds.MaxX-nav.GridBounds.MinX, nav.GridBounds.MaxY-nav.GridBounds.MinY) + 1
	var (
		bestGoal       Vec2
		bestPlayerDist = math.MaxFloat64
		bestCell       gridCell
		found          bool
	)
	for radius := 0; radius <= maxRadius; radius++ {
		for _, cell := range ringCells(playerCell, radius) {
			if !cellInBounds(nav, cell) || blocked(cell.x, cell.y) {
				continue
			}
			goal := gridToWorld(nav, cell)
			goalDist := distance(goal, player.pos)
			if goalDist > maxReach {
				continue
			}
			steps, ok := PlanPath(nav, monster.pos, goal, blocked)
			if !ok {
				continue
			}
			if len(steps) == 0 && distance(monster.pos, goal) > nav.StopDistance+1e-9 {
				continue
			}
			if !found || goalDist < bestPlayerDist-1e-9 ||
				(math.Abs(goalDist-bestPlayerDist) <= 1e-9 && cellLess(cell, bestCell)) {
				bestGoal = goal
				bestPlayerDist = goalDist
				bestCell = cell
				found = true
			}
		}
	}
	if !found {
		return Vec2{}, false
	}

	return bestGoal, true
}

func cellLess(a, b gridCell) bool {
	if a.y != b.y {
		return a.y < b.y
	}

	return a.x < b.x
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
	player := s.activeLevel().entities[s.playerID]
	if player != nil && player.hp > 0 {
		if circlesOverlap(pos, monsterRadius, player.pos, playerRadius) {
			return true
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
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(p.view())})
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
)

func (s *Sim) firstProjectileHit(p *entity, candidate Vec2) (projectileHit, bool) {
	best := projectileHit{t: math.Inf(1)}
	found := false
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
			if e.hp <= 0 {
				continue
			}
			if t, ok := segmentIntersectsCircle(p.pos, candidate, e.pos, monsterRadius+projectileRadius); ok {
				consider(projectileHit{t: t, category: projectileHitMonster, entityID: e.id})
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
	if hit.category != projectileHitMonster {
		res.Events = append(res.Events, Event{EventType: "projectile_blocked", CorrelationID: p.sourceCorrID})
		return
	}
	target := s.activeLevel().entities[hit.entityID]
	if target == nil || target.kind != monsterEntity || target.hp <= 0 {
		res.Events = append(res.Events, Event{EventType: "projectile_expired", CorrelationID: p.sourceCorrID})
		return
	}
	hitDraw := s.rng.Next()
	hitOK := s.rules.Combat.BaseHitChance >= 1.0 ||
		float64(hitDraw%10000)/10000.0 < s.rules.Combat.BaseHitChance
	if !hitOK {
		res.Events = append(res.Events, Event{EventType: "attack_missed", EntityID: idStr(target.id), CorrelationID: p.sourceCorrID})
		return
	}
	dmg := s.rollRange(p.damageRange)
	target.hp -= dmg
	if target.hp < 0 {
		target.hp = 0
	}
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(target.view())})
	res.Events = append(res.Events, Event{EventType: "monster_damaged", EntityID: idStr(target.id), CorrelationID: p.sourceCorrID, Damage: intPtr(dmg)})

	if target.hp == 0 {
		res.Events = append(res.Events, Event{EventType: "monster_killed", EntityID: idStr(target.id), CorrelationID: p.sourceCorrID})
		s.dropLoot(target, p.sourceCorrID, res)
	}
	s.retaliate(target, p.sourceCorrID, res)
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

func (s *Sim) playerReach() float64 {
	base := s.rules.Combat.UnarmedReach
	instanceID := s.equipped[weaponSlot]
	if instanceID == 0 {
		return base
	}
	item := s.findItemByID(instanceID)
	if item == nil {
		return base
	}
	def, ok := s.rules.Items[item.itemDefID]
	if !ok || def.Reach == nil {
		return base
	}
	return *def.Reach
}

func (s *Sim) playerMeleeReach() float64 {
	def, ok := s.equippedWeaponDef()
	if !ok || def.AttackMode == attackModeRanged {
		return s.rules.Combat.UnarmedReach
	}
	if def.Reach == nil {
		return s.rules.Combat.UnarmedReach
	}
	return *def.Reach
}

func (s *Sim) playerActionReach() float64 {
	return s.playerReach()
}

func (s *Sim) playerAttackMode() string {
	def, ok := s.equippedWeaponDef()
	if !ok || def.AttackMode == "" {
		return attackModeMelee
	}
	return def.AttackMode
}

func (s *Sim) equippedWeaponDef() (ItemDef, bool) {
	instanceID := s.equipped[weaponSlot]
	if instanceID == 0 {
		return ItemDef{}, false
	}
	item := s.findItemByID(instanceID)
	if item == nil {
		return ItemDef{}, false
	}
	def, ok := s.rules.Items[item.itemDefID]
	return def, ok
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
		return e.state == interactableClosed || (e.state == interactableReady && e.interactableDefID == teleporterDefID)
	default:
		return false
	}
}

func (s *Sim) resolvePlayerAttackDamage() DamageRange {
	base := s.rules.Combat.PlayerDamage
	instanceID := s.equipped[weaponSlot]
	if instanceID == 0 {
		return base
	}
	item := s.findItemByID(instanceID)
	if item == nil {
		return base
	}
	def, ok := s.rules.Items[item.itemDefID]
	if !ok || def.Damage == nil {
		return base
	}
	return *def.Damage
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
	if span <= 0 {
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

func sortedEntityIDs(entities map[uint64]*entity) []uint64 {
	ids := make([]uint64, 0, len(entities))
	for id := range entities {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// Snapshot returns the full authoritative state, with entities ordered by id.
func (s *Sim) Snapshot() Snapshot {
	ids := sortedEntityIDs(s.activeLevel().entities)

	entities := make([]EntityView, 0, len(ids))
	for _, id := range ids {
		entities = append(entities, s.activeLevel().entities[id].view())
	}

	inventory := make([]ItemView, 0, len(s.inventory))
	for _, it := range s.inventory {
		inventory = append(inventory, it.view())
	}

	equipped := make(map[string]*string, len(s.equipped))
	for slot, instanceID := range s.equipped {
		if instanceID == 0 {
			equipped[slot] = nil
			continue
		}
		v := idStr(instanceID)
		equipped[slot] = &v
	}

	return Snapshot{
		ServerTick:            s.tick,
		SessionID:             s.sessionID,
		Seed:                  s.seed,
		CurrentLevel:          s.currentLevel,
		Entities:              entities,
		Inventory:             inventory,
		Equipped:              equipped,
		DiscoveredTeleporters: s.teleporterDiscoveryView(),
		RecentEvents:          []Event{},
	}
}

func (s *Sim) teleporterDiscoveryView() []TeleporterDiscoveryView {
	if !s.multiLevel {
		return []TeleporterDiscoveryView{}
	}
	levels := make([]int, 0, len(s.levels))
	for levelNum := range s.levels {
		if levelNum < levelZero {
			levels = append(levels, levelNum)
		}
	}
	sort.Ints(levels)
	out := make([]TeleporterDiscoveryView, 0, len(levels))
	for _, levelNum := range levels {
		out = append(out, TeleporterDiscoveryView{Level: levelNum, Discovered: s.discoveredTeleporters[levelNum]})
	}
	return out
}

func (e *entity) view() EntityView {
	ev := EntityView{ID: idStr(e.id), Type: e.kind, Position: e.pos}
	switch e.kind {
	case playerEntity, monsterEntity:
		hp, maxHP := e.hp, e.maxHP
		ev.HP = &hp
		ev.MaxHP = &maxHP
		if e.kind == monsterEntity {
			ev.MonsterDefID = e.monsterDefID
		}
	case lootEntity:
		ev.ItemDefID = e.itemDefID
	case interactableEntity:
		ev.InteractableDefID = e.interactableDefID
		ev.State = e.state
	case projectileEntity:
		ev.OwnerID = idStr(e.ownerID)
		ev.TargetID = idStr(e.targetID)
		ev.ProjectileDefID = e.projectileDefID
	}
	return ev
}

func (it *invItem) view() ItemView {
	return ItemView{
		ItemInstanceID: idStr(it.instanceID),
		ItemDefID:      it.itemDefID,
		Slot:           it.slot,
		Equipped:       it.equipped,
	}
}

func ptrEntityView(v EntityView) *EntityView { return &v }
func ptrItemView(v ItemView) *ItemView       { return &v }
func intPtr(v int) *int                      { return &v }

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
