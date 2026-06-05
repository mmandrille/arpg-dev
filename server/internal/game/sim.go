package game

import (
	"math"
	"sort"
	"strconv"
)

// Simulation constants for the v0 slice.
const (
	baseEntityID  = 1001 // player=1001, monster=1002, loot=1003, item=1004 ...
	playerStartHP = 10
	moveSpeed     = 1.0
	monsterDefID  = "training_dummy"
	playerEntity  = "player"
	monsterEntity = "monster"
	lootEntity    = "loot"
	weaponSlot    = "weapon"
)

var (
	playerStartPos  = Vec2{X: 10, Y: 5}
	monsterStartPos = Vec2{X: 12, Y: 5}
)

// entity is the internal mutable scene entity.
type entity struct {
	id           uint64
	kind         string
	pos          Vec2
	hp           int
	maxHP        int
	monsterDefID string
	itemDefID    string
	lootTable    string
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
}

// NewSim builds a fresh session with the player and the v0 monster spawned.
func NewSim(sessionID, seed string, rules *Rules) *Sim {
	s := &Sim{
		sessionID: sessionID,
		seed:      seed,
		rng:       NewRNG(SeedToUint64(seed)),
		rules:     rules,
		nextID:    baseEntityID,
		entities:  make(map[uint64]*entity),
		equipped:  map[string]uint64{weaponSlot: 0},
	}

	player := &entity{kind: playerEntity, pos: playerStartPos, hp: playerStartHP, maxHP: playerStartHP}
	player.id = s.alloc()
	s.playerID = player.id
	s.entities[player.id] = player

	def := rules.Monsters[monsterDefID]
	monster := &entity{
		kind:         monsterEntity,
		pos:          monsterStartPos,
		hp:           def.MaxHP,
		maxHP:        def.MaxHP,
		monsterDefID: monsterDefID,
		lootTable:    def.LootTable,
	}
	monster.id = s.alloc()
	s.entities[monster.id] = monster

	return s
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
	Attack        *AttackIntent
	PickUp        *PickUpIntent
	Equip         *EquipIntent
}

// Intent payloads.
type (
	MoveIntent struct {
		Direction     Vec2
		DurationTicks int
	}
	AttackIntent struct{ TargetID string }
	PickUpIntent struct{ EntityID string }
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
		case "move_intent", "attack_intent", "pick_up_intent", "equip_intent":
			res.reject(in.MessageID, "player_dead")
			return
		}
	}
	switch in.Type {
	case "client_ready":
		res.ack(in.MessageID)
	case "move_intent":
		s.handleMove(in, res)
	case "attack_intent":
		s.handleAttack(in, res)
	case "pick_up_intent":
		s.handlePickUp(in, res)
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

func (s *Sim) handleAttack(in Input, res *TickResult) {
	if in.Attack == nil {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	target := s.findEntity(in.Attack.TargetID)
	if target == nil || target.kind != monsterEntity || target.hp <= 0 {
		res.reject(in.MessageID, "invalid_target")
		return
	}

	// Always consume two draws (hit, damage) so the RNG stream is independent
	// of the hit/miss branch (base_hit_chance is 1.0 in v0).
	hitDraw := s.rng.Next()
	dmg := s.rollDamage()
	hit := s.rules.Combat.BaseHitChance >= 1.0 ||
		float64(hitDraw%10000)/10000.0 < s.rules.Combat.BaseHitChance

	res.ack(in.MessageID)
	if !hit {
		res.Events = append(res.Events, Event{EventType: "attack_missed", EntityID: in.Attack.TargetID, CorrelationID: in.CorrelationID})
		return
	}

	target.hp -= dmg
	if target.hp < 0 {
		target.hp = 0
	}
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(target.view())})
	res.Events = append(res.Events, Event{EventType: "monster_damaged", EntityID: in.Attack.TargetID, CorrelationID: in.CorrelationID})

	if target.hp == 0 {
		res.Events = append(res.Events, Event{EventType: "monster_killed", EntityID: in.Attack.TargetID, CorrelationID: in.CorrelationID})
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
	res.Events = append(res.Events, Event{EventType: eventType, EntityID: idStr(player.id), CorrelationID: corr})
}

func (s *Sim) handlePickUp(in Input, res *TickResult) {
	if in.PickUp == nil {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	e := s.findEntity(in.PickUp.EntityID)
	if e == nil || e.kind != lootEntity {
		// Covers duplicate pickup: the loot entity is already gone.
		res.reject(in.MessageID, "invalid_target")
		return
	}

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
	player.pos.X += s.move.dir.X * moveSpeed
	player.pos.Y += s.move.dir.Y * moveSpeed
	s.move.remaining--
	if s.move.remaining == 0 {
		s.move = nil
	}
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(player.view())})
}

func (s *Sim) rollDamage() int {
	return s.rollRange(s.rules.Combat.PlayerDamage)
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

// Snapshot returns the full authoritative state, with entities ordered by id.
func (s *Sim) Snapshot() Snapshot {
	ids := make([]uint64, 0, len(s.entities))
	for id := range s.entities {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

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

func normalize(v Vec2) Vec2 {
	length := math.Hypot(v.X, v.Y)
	if length == 0 {
		return Vec2{}
	}
	return Vec2{X: v.X / length, Y: v.Y / length}
}
