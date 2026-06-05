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
	wallEntity                    = "wall"
	interactableEntity            = "interactable"
	interactableClosed            = "closed"
	interactableOpen              = "open"
	weaponSlot                    = "weapon"
	lootInteractionRadius         = 0.35
	interactableInteractionRadius = 0.50
	meleeRangeEpsilon             = 0.000001
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
	Action        *ActionIntent
	Equip         *EquipIntent
}

// Intent payloads.
type (
	MoveIntent struct {
		Direction     Vec2
		DurationTicks int
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
	s.tick++
	return res
}

func (s *Sim) applyInput(in Input, res *TickResult) {
	if in.Type != "client_ready" && s.playerDead() {
		switch in.Type {
		case "move_intent", "action_intent", "equip_intent":
			res.reject(in.MessageID, "player_dead")
			return
		}
	}
	switch in.Type {
	case "client_ready":
		res.ack(in.MessageID)
	case "move_intent":
		s.handleMove(in, res)
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
		s.move = &activeMove{dir: dir, remaining: dur}
	}
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
	if !s.inMeleeRange(target) {
		res.reject(in.MessageID, "out_of_range")
		return
	}

	switch target.kind {
	case monsterEntity:
		s.attackTarget(target, in, res)
	case lootEntity:
		s.pickUpTarget(target, in, res)
	case interactableEntity:
		s.activateInteractable(target, in, res)
	default:
		res.reject(in.MessageID, "invalid_target")
	}
}

func (s *Sim) attackTarget(target *entity, in Input, res *TickResult) {
	// Always consume two draws (hit, damage) so the RNG stream is independent
	// of the hit/miss branch (base_hit_chance is 1.0 in v0).
	hitDraw := s.rng.Next()
	dmg := s.rollDamage()
	hit := s.rules.Combat.BaseHitChance >= 1.0 ||
		float64(hitDraw%10000)/10000.0 < s.rules.Combat.BaseHitChance

	res.ack(in.MessageID)
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

func (s *Sim) pickUpTarget(e *entity, in Input, res *TickResult) {
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
	res.ack(in.MessageID)
}

func (s *Sim) activateInteractable(e *entity, in Input, res *TickResult) {
	e.state = interactableOpen
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(e.view())})
	res.Events = append(res.Events, Event{EventType: "interactable_activated", EntityID: idStr(e.id), CorrelationID: in.CorrelationID})
	res.ack(in.MessageID)
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

func circlesOverlap(a Vec2, ar float64, b Vec2, br float64) bool {
	dx := a.X - b.X
	dy := a.Y - b.Y
	r := ar + br
	return dx*dx+dy*dy < r*r-1e-9
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
