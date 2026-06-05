package game

import (
	"math"
	"sort"
	"strconv"
)

// Simulation constants for the v0 slice.
const (
	baseEntityID                  = 1001 // player=1001, monster=1002, loot=1003, item=1004 ...
	playerStartHP                 = 10
	moveSpeed                     = 1.0
	playerRadius                  = 0.45
	monsterRadius                 = 0.45
	monsterDefID                  = "training_dummy"
	playerEntity                  = "player"
	monsterEntity                 = "monster"
	lootEntity                    = "loot"
	projectileEntity              = "projectile"
	wallEntity                    = "wall"
	interactableEntity            = "interactable"
	interactableClosed            = "closed"
	interactableOpen              = "open"
	attackModeMelee               = "melee"
	attackModeRanged              = "ranged"
	trainingArrowProjectileDefID  = "training_arrow"
	weaponSlot                    = "weapon"
	lootInteractionRadius         = 0.35
	interactableInteractionRadius = 0.50
	meleeRangeEpsilon             = 0.000001
	projectileRadius              = 0.10
	tickDuration                  = 0.05
)

// DefaultWorldID is the compatibility world used when callers do not choose a
// preset explicitly.
const DefaultWorldID = "vertical_slice"

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

	entities  map[uint64]*entity
	inventory []*invItem
	equipped  map[string]uint64 // slot -> instanceID (0 = none)
	move      *activeMove
	autoNav   *autoNavState
	walls     []wallObstacle
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
		sessionID: sessionID,
		seed:      seed,
		rng:       NewRNG(SeedToUint64(seed)),
		rules:     rules,
		nextID:    baseEntityID,
		entities:  make(map[uint64]*entity),
		equipped:  map[string]uint64{weaponSlot: 0},
	}

	player := &entity{kind: playerEntity, pos: world.Player.Position, hp: playerStartHP, maxHP: playerStartHP}
	player.id = s.alloc()
	s.playerID = player.id
	s.entities[player.id] = player

	for _, preset := range world.Entities {
		switch preset.Type {
		case monsterEntity:
			def := rules.Monsters[preset.MonsterDefID]
			monster := &entity{
				kind:         monsterEntity,
				pos:          preset.Position,
				hp:           def.MaxHP,
				maxHP:        def.MaxHP,
				monsterDefID: preset.MonsterDefID,
				lootTable:    def.LootTable,
			}
			monster.id = s.alloc()
			s.entities[monster.id] = monster
		case lootEntity:
			loot := &entity{kind: lootEntity, pos: preset.Position, itemDefID: preset.ItemDefID}
			loot.id = s.alloc()
			s.entities[loot.id] = loot
		case wallEntity:
			s.walls = append(s.walls, wallObstacle{pos: preset.Position, size: preset.Size})
		case interactableEntity:
			def := rules.Interactables[preset.InteractableDefID]
			interactable := &entity{
				kind:              interactableEntity,
				pos:               preset.Position,
				interactableDefID: preset.InteractableDefID,
				state:             def.InitialState,
			}
			interactable.id = s.alloc()
			s.entities[interactable.id] = interactable
		default:
			return nil, ErrUnknownWorldEntity{WorldID: worldID, EntityType: preset.Type}
		}
	}

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
	Equip         *EquipIntent
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
	ActionIntent struct{ TargetID string }
	EquipIntent  struct {
		ItemInstanceID string
		Slot           string
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
	Changes []Change
	Events  []Event
	Acks    []Ack
	Rejects []Reject
}

func (r *TickResult) ack(id string) { r.Acks = append(r.Acks, Ack{MessageID: id}) }
func (r *TickResult) reject(id, reason string) {
	r.Rejects = append(r.Rejects, Reject{MessageID: id, Reason: reason})
}

// Tick processes the inputs stamped for the current tick (already ordered by
// the runner as (sequence, message_id)), applies continuous movement, advances
// the tick counter, and returns the resulting changes/events/acks.
func (s *Sim) Tick(inputs []Input) TickResult {
	// Changes/Events are always non-nil so they marshal as [] (not null),
	// satisfying the state_delta schema.
	res := TickResult{Tick: s.tick, Changes: []Change{}, Events: []Event{}}
	for _, in := range inputs {
		s.applyInput(in, &res)
	}
	s.applyMovement(&res)
	s.advanceProjectiles(&res)
	s.tick++
	return res
}

func (s *Sim) applyInput(in Input, res *TickResult) {
	if in.Type != "client_ready" && s.playerDead() {
		switch in.Type {
		case "move_intent", "move_to_intent", "action_intent", "equip_intent":
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
	case "equip_intent":
		s.handleEquip(in, res)
	default:
		res.reject(in.MessageID, "unknown_type")
	}
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
		s.move = &activeMove{dir: dir, remaining: dur}
	}
	res.ack(in.MessageID)
}

func (s *Sim) handleMoveTo(in Input, res *TickResult) {
	if in.MoveTo == nil || !finiteVec2(in.MoveTo.Position) {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	player := s.entities[s.playerID]
	if player == nil {
		res.reject(in.MessageID, "player_dead")
		return
	}
	if distance(player.pos, in.MoveTo.Position) <= s.rules.Navigation.StopDistance {
		s.clearAutoNav()
		res.ack(in.MessageID)
		return
	}
	steps, ok := PlanPath(s.rules.Navigation, player.pos, in.MoveTo.Position, s.buildBlockedFn())
	if !ok {
		res.reject(in.MessageID, "no_path")
		return
	}
	if len(steps) > s.rules.Navigation.MaxAutoSteps {
		res.reject(in.MessageID, "path_too_long")
		return
	}
	s.move = nil
	s.autoNav = &autoNavState{steps: steps, sourceMsgID: in.MessageID, sourceCorrID: in.CorrelationID}
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
	if len(steps) > s.rules.Navigation.MaxAutoSteps {
		res.reject(in.MessageID, "path_too_long")
		return
	}
	s.move = nil
	s.autoNav = &autoNavState{
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
	player := s.entities[s.playerID]
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
	s.entities[projectile.id] = projectile
	res.Changes = append(res.Changes, Change{Op: OpEntitySpawn, Entity: ptrEntityView(projectile.view())})
	if ack {
		res.ack(in.MessageID)
	}
}

func (s *Sim) dropLoot(monster *entity, corr string, res *TickResult) {
	itemDefID, ok := s.rules.RollLoot(monster.lootTable, s.rng)
	if !ok {
		return
	}
	loot := &entity{kind: lootEntity, pos: monster.pos, itemDefID: itemDefID}
	loot.id = s.alloc()
	s.entities[loot.id] = loot
	res.Changes = append(res.Changes, Change{Op: OpEntitySpawn, Entity: ptrEntityView(loot.view())})
	res.Events = append(res.Events, Event{EventType: "loot_dropped", EntityID: idStr(loot.id), CorrelationID: corr})
}

func (s *Sim) retaliate(monster *entity, corr string, res *TickResult) {
	def := s.rules.Monsters[monster.monsterDefID]
	if def.RetaliationDamage == nil {
		return
	}
	player := s.entities[s.playerID]
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
	delete(s.entities, e.id)
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
	e.state = interactableOpen
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(e.view())})
	res.Events = append(res.Events, Event{EventType: "interactable_activated", EntityID: idStr(e.id), CorrelationID: in.CorrelationID})
	if ack {
		res.ack(in.MessageID)
	}
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

func (s *Sim) applyMovement(res *TickResult) {
	if s.autoNav != nil && s.move == nil {
		s.applyAutoNav(res)
		return
	}
	if s.move == nil || s.move.remaining <= 0 {
		return
	}
	if s.playerDead() {
		s.move = nil
		return
	}
	player := s.entities[s.playerID]
	before := player.pos
	player.pos = s.resolveMovement(player.pos, Vec2{
		X: s.move.dir.X * moveSpeed,
		Y: s.move.dir.Y * moveSpeed,
	})
	s.move.remaining--
	if s.move.remaining == 0 {
		s.move = nil
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
	if len(s.autoNav.steps) == 0 {
		s.finishAutoNav(res)
		return
	}
	player := s.entities[s.playerID]
	before := player.pos
	step := normalize(s.autoNav.steps[0])
	s.autoNav.steps = s.autoNav.steps[1:]
	player.pos = s.resolveMovement(player.pos, Vec2{X: step.X * moveSpeed, Y: step.Y * moveSpeed})
	if player.pos != before {
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(player.view())})
	}
	if len(s.autoNav.steps) == 0 {
		s.finishAutoNav(res)
	}
}

func (s *Sim) finishAutoNav(res *TickResult) {
	nav := s.autoNav
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
	s.autoNav = nil
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
	for _, wall := range s.walls {
		if circleIntersectsAABB(pos, playerRadius, wall.pos, wall.size) {
			return true
		}
	}
	for _, id := range sortedEntityIDs(s.entities) {
		e := s.entities[id]
		if e.kind == monsterEntity && e.hp > 0 {
			if circlesOverlap(pos, playerRadius, e.pos, monsterRadius) {
				return true
			}
			continue
		}
		if e.kind == interactableEntity && e.state == interactableClosed {
			if def, ok := s.rules.Interactables[e.interactableDefID]; ok {
				if circleIntersectsAABB(pos, playerRadius, e.pos, def.BarrierWhenClosed.Size) {
					return true
				}
			}
		}
	}
	return false
}

func (s *Sim) buildBlockedFn() func(gx, gy int) bool {
	return func(gx, gy int) bool {
		center := gridToWorld(s.rules.Navigation, gridCell{x: gx, y: gy})
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
	player := s.entities[s.playerID]
	if player == nil {
		return Vec2{}, nil, false
	}
	nav := s.rules.Navigation
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
		return meleeInRange(distance(pos, target.pos), s.playerReach(), s.targetInteractionRadius(target))
	})
}

func (s *Sim) findApproachGoalMatching(target *entity, inRange func(Vec2, *entity) bool) (Vec2, []Vec2, bool) {
	player := s.entities[s.playerID]
	if player == nil {
		return Vec2{}, nil, false
	}
	nav := s.rules.Navigation
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

func (s *Sim) advanceProjectiles(res *TickResult) {
	ids := sortedEntityIDs(s.entities)
	for _, id := range ids {
		p := s.entities[id]
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
		delete(s.entities, p.id)
		res.Changes = append(res.Changes, Change{Op: OpEntityRemove, EntityID: idStr(p.id)})
		return
	}
	if p.traveled+segmentLength >= p.maxDistance-meleeRangeEpsilon {
		res.Events = append(res.Events, Event{EventType: "projectile_expired", CorrelationID: p.sourceCorrID})
		delete(s.entities, p.id)
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
	for _, wall := range s.walls {
		if t, ok := segmentIntersectsInflatedAABB(p.pos, candidate, wall.pos, wall.size, projectileRadius); ok {
			consider(projectileHit{t: t, category: projectileHitWall})
		}
	}
	for _, id := range sortedEntityIDs(s.entities) {
		e := s.entities[id]
		if e == nil || e.id == p.id {
			continue
		}
		switch e.kind {
		case interactableEntity:
			if e.state != interactableClosed {
				continue
			}
			def, ok := s.rules.Interactables[e.interactableDefID]
			if !ok {
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
	target := s.entities[hit.entityID]
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
	player := s.entities[s.playerID]
	if player == nil {
		return false
	}
	return meleeInRange(distance(player.pos, target.pos), s.playerReach(), s.targetInteractionRadius(target))
}

func (s *Sim) inActionRange(target *entity) bool {
	player := s.entities[s.playerID]
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
		player := s.entities[s.playerID]
		return player != nil && s.inActionRange(target) && s.hasClearRangedShot(player.pos, target)
	}
	return s.inMeleeRange(target)
}

func (s *Sim) hasClearRangedShot(from Vec2, target *entity) bool {
	if target == nil || target.kind != monsterEntity || target.hp <= 0 {
		return false
	}
	for _, wall := range s.walls {
		if _, ok := segmentIntersectsInflatedAABB(from, target.pos, wall.pos, wall.size, projectileRadius); ok {
			return false
		}
	}
	for _, id := range sortedEntityIDs(s.entities) {
		e := s.entities[id]
		if e == nil || e.id == target.id {
			continue
		}
		switch e.kind {
		case interactableEntity:
			if e.state != interactableClosed {
				continue
			}
			def, ok := s.rules.Interactables[e.interactableDefID]
			if !ok {
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
		return e.state == interactableClosed
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
	for _, e := range s.entities {
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
	player := s.entities[s.playerID]
	return player == nil || player.hp <= 0
}

func (s *Sim) findEntity(id string) *entity {
	n, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil
	}
	return s.entities[n]
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
	ids := sortedEntityIDs(s.entities)

	entities := make([]EntityView, 0, len(ids))
	for _, id := range ids {
		entities = append(entities, s.entities[id].view())
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
		ServerTick:   s.tick,
		SessionID:    s.sessionID,
		Seed:         s.seed,
		Entities:     entities,
		Inventory:    inventory,
		Equipped:     equipped,
		RecentEvents: []Event{},
	}
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
