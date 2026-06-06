package game

import (
	"encoding/json"
	"strconv"
)

// Vec2 is a 2D position in scene units.
type Vec2 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// EntityView is the protocol view of a scene entity.
// HP/MaxHP are pointers so a value of 0 (a dead monster) is preserved while a
// loot entity simply omits them.
type EntityView struct {
	ID                string `json:"id"`
	Type              string `json:"type"`
	Position          Vec2   `json:"position"`
	HP                *int   `json:"hp,omitempty"`
	MaxHP             *int   `json:"max_hp,omitempty"`
	MonsterDefID      string `json:"monster_def_id,omitempty"`
	ItemDefID         string `json:"item_def_id,omitempty"`
	InteractableDefID string `json:"interactable_def_id,omitempty"`
	OwnerID           string `json:"owner_id,omitempty"`
	TargetID          string `json:"target_id,omitempty"`
	ProjectileDefID   string `json:"projectile_def_id,omitempty"`
	State             string `json:"state,omitempty"`
}

// ItemView is the protocol view of an inventory item.
type ItemView struct {
	ItemInstanceID string `json:"item_instance_id"`
	ItemDefID      string `json:"item_def_id"`
	Slot           string `json:"slot"`
	Equipped       bool   `json:"equipped"`
}

// Event is an authoritative event emitted by the sim.
type Event struct {
	EventType      string `json:"event_type"`
	EntityID       string `json:"entity_id,omitempty"`
	CorrelationID  string `json:"correlation_id,omitempty"`
	Damage         *int   `json:"damage,omitempty"`
	Heal           *int   `json:"heal,omitempty"`
	ItemInstanceID string `json:"item_instance_id,omitempty"`
}

// Snapshot is the full authoritative state for rendering (session_snapshot).
type Snapshot struct {
	ServerTick   uint64             `json:"server_tick"`
	SessionID    string             `json:"session_id"`
	Seed         string             `json:"seed"`
	Entities     []EntityView       `json:"entities"`
	Inventory    []ItemView         `json:"inventory"`
	Equipped     map[string]*string `json:"equipped"`
	RecentEvents []Event            `json:"recent_events"`
}

// Change operation names (the state_delta ops).
const (
	OpEntitySpawn     = "entity_spawn"
	OpEntityUpdate    = "entity_update"
	OpEntityRemove    = "entity_remove"
	OpInventoryAdd    = "inventory_add"
	OpInventoryUpdate = "inventory_update"
	OpInventoryRemove = "inventory_remove"
	OpEquippedUpdate  = "equipped_update"
)

// Change is one ordered authoritative change within a tick. It marshals to
// exactly the fields required for its op (matching the state_delta schema's
// oneOf), which is why it has a custom MarshalJSON.
type Change struct {
	Op             string
	Entity         *EntityView
	EntityID       string
	Item           *ItemView
	Slot           string
	ItemInstanceID *string // for equipped_update; nil marshals as null
}

// MarshalJSON renders the change as the precise object for its op.
func (c Change) MarshalJSON() ([]byte, error) {
	switch c.Op {
	case OpEntitySpawn, OpEntityUpdate:
		return json.Marshal(struct {
			Op     string      `json:"op"`
			Entity *EntityView `json:"entity"`
		}{c.Op, c.Entity})
	case OpEntityRemove:
		return json.Marshal(struct {
			Op       string `json:"op"`
			EntityID string `json:"entity_id"`
		}{c.Op, c.EntityID})
	case OpInventoryAdd, OpInventoryUpdate:
		return json.Marshal(struct {
			Op   string    `json:"op"`
			Item *ItemView `json:"item"`
		}{c.Op, c.Item})
	case OpInventoryRemove:
		id := ""
		if c.ItemInstanceID != nil {
			id = *c.ItemInstanceID
		}
		return json.Marshal(struct {
			Op             string `json:"op"`
			ItemInstanceID string `json:"item_instance_id"`
		}{c.Op, id})
	case OpEquippedUpdate:
		return json.Marshal(struct {
			Op             string  `json:"op"`
			Slot           string  `json:"slot"`
			ItemInstanceID *string `json:"item_instance_id"`
		}{c.Op, c.Slot, c.ItemInstanceID})
	default:
		return nil, &json.UnsupportedValueError{Str: "unknown change op: " + c.Op}
	}
}

// idStr renders an unsigned 64-bit entity id as a decimal string.
func idStr(id uint64) string { return strconv.FormatUint(id, 10) }
