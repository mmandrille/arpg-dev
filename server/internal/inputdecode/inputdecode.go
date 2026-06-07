// Package inputdecode converts protocol intent envelopes into game inputs
// without depending on the realtime WebSocket runner.
package inputdecode

import (
	"encoding/json"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
)

const (
	TypeClientReady  = "client_ready"
	TypeMoveIntent   = "move_intent"
	TypeMoveTo       = "move_to_intent"
	TypeAction       = "action_intent"
	TypeDescend      = "descend_intent"
	TypeAscend       = "ascend_intent"
	TypeTeleport     = "teleport_intent"
	TypeEquip        = "equip_intent"
	TypeUnequip      = "unequip_intent"
	TypeDrop         = "drop_intent"
	TypeUse          = "use_intent"
	TypeAssignHotbar = "assign_hotbar_intent"
	TypeUseHotbar    = "use_hotbar_intent"
	TypeAllocateStat = "allocate_stat_intent"
)

type envelope struct {
	Type          string          `json:"type"`
	MessageID     string          `json:"message_id"`
	CorrelationID string          `json:"correlation_id"`
	Payload       json.RawMessage `json:"payload"`
}

type (
	movePayloadWire struct {
		Direction     game.Vec2 `json:"direction"`
		DurationTicks int       `json:"duration_ticks"`
	}
	moveToPayloadWire struct {
		Position game.Vec2 `json:"position"`
	}
	actionPayloadWire struct {
		TargetID string `json:"target_id"`
	}
	descendPayloadWire  struct{}
	ascendPayloadWire   struct{}
	teleportPayloadWire struct {
		TargetLevel *int `json:"target_level"`
	}
	equipPayloadWire struct {
		ItemInstanceID string `json:"item_instance_id"`
		Slot           string `json:"slot"`
	}
	unequipPayloadWire struct {
		Slot string `json:"slot"`
	}
	dropPayloadWire struct {
		ItemInstanceID string `json:"item_instance_id"`
	}
	usePayloadWire struct {
		ItemInstanceID string `json:"item_instance_id"`
	}
	assignHotbarPayloadWire struct {
		SlotIndex      int     `json:"slot_index"`
		ItemInstanceID *string `json:"item_instance_id"`
	}
	useHotbarPayloadWire struct {
		SlotIndex int `json:"slot_index"`
	}
	allocateStatPayloadWire struct {
		Stat   string `json:"stat"`
		Points int    `json:"points"`
	}
)

// IsClientIntent reports whether the type is a buffered authoritative intent.
func IsClientIntent(t string) bool {
	switch t {
	case TypeMoveIntent, TypeMoveTo, TypeAction, TypeDescend, TypeAscend, TypeTeleport, TypeEquip, TypeUnequip, TypeDrop, TypeUse, TypeAssignHotbar, TypeUseHotbar, TypeAllocateStat:
		return true
	}
	return false
}

// Decode converts an intent type and payload into a sim Input. It returns
// ok=false if the payload does not match the message type.
func Decode(typ, messageID, correlationID string, payload json.RawMessage) (game.Input, bool) {
	in := game.Input{
		MessageID:     messageID,
		CorrelationID: correlationID,
		Type:          typ,
	}
	switch typ {
	case TypeMoveIntent:
		var p movePayloadWire
		if err := json.Unmarshal(payload, &p); err != nil {
			return in, false
		}
		in.Move = &game.MoveIntent{Direction: p.Direction, DurationTicks: p.DurationTicks}
	case TypeMoveTo:
		var p moveToPayloadWire
		if err := json.Unmarshal(payload, &p); err != nil {
			return in, false
		}
		in.MoveTo = &game.MoveToIntent{Position: p.Position}
	case TypeAction:
		var p actionPayloadWire
		if err := json.Unmarshal(payload, &p); err != nil || p.TargetID == "" {
			return in, false
		}
		in.Action = &game.ActionIntent{TargetID: p.TargetID}
	case TypeDescend:
		var p descendPayloadWire
		if err := json.Unmarshal(payload, &p); err != nil {
			return in, false
		}
		in.Descend = &game.DescendIntent{}
	case TypeAscend:
		var p ascendPayloadWire
		if err := json.Unmarshal(payload, &p); err != nil {
			return in, false
		}
		in.Ascend = &game.AscendIntent{}
	case TypeTeleport:
		var p teleportPayloadWire
		if err := json.Unmarshal(payload, &p); err != nil || p.TargetLevel == nil || *p.TargetLevel > 0 {
			return in, false
		}
		in.Teleport = &game.TeleportIntent{TargetLevel: *p.TargetLevel}
	case TypeEquip:
		var p equipPayloadWire
		if err := json.Unmarshal(payload, &p); err != nil || p.ItemInstanceID == "" || p.Slot == "" {
			return in, false
		}
		in.Equip = &game.EquipIntent{ItemInstanceID: p.ItemInstanceID, Slot: p.Slot}
	case TypeUnequip:
		var p unequipPayloadWire
		if err := json.Unmarshal(payload, &p); err != nil || p.Slot == "" {
			return in, false
		}
		in.Unequip = &game.UnequipIntent{Slot: p.Slot}
	case TypeDrop:
		var p dropPayloadWire
		if err := json.Unmarshal(payload, &p); err != nil || p.ItemInstanceID == "" {
			return in, false
		}
		in.Drop = &game.DropIntent{ItemInstanceID: p.ItemInstanceID}
	case TypeUse:
		var p usePayloadWire
		if err := json.Unmarshal(payload, &p); err != nil || p.ItemInstanceID == "" {
			return in, false
		}
		in.Use = &game.UseIntent{ItemInstanceID: p.ItemInstanceID}
	case TypeAssignHotbar:
		var p assignHotbarPayloadWire
		if err := json.Unmarshal(payload, &p); err != nil || p.SlotIndex < 0 || p.SlotIndex > 9 {
			return in, false
		}
		in.AssignHotbar = &game.AssignHotbarIntent{SlotIndex: p.SlotIndex, ItemInstanceID: p.ItemInstanceID}
	case TypeUseHotbar:
		var p useHotbarPayloadWire
		if err := json.Unmarshal(payload, &p); err != nil || p.SlotIndex < 0 || p.SlotIndex > 9 {
			return in, false
		}
		in.UseHotbar = &game.UseHotbarIntent{SlotIndex: p.SlotIndex}
	case TypeAllocateStat:
		var p allocateStatPayloadWire
		if err := json.Unmarshal(payload, &p); err != nil || !validStat(p.Stat) || p.Points <= 0 {
			return in, false
		}
		in.AllocateStat = &game.AllocateStatIntent{Stat: p.Stat, Points: p.Points}
	default:
		return in, false
	}
	return in, true
}

func validStat(stat string) bool {
	switch stat {
	case "str", "dex", "vit", "magic":
		return true
	}
	return false
}

// DecodeStored recovers a sim Input from a persisted input envelope. The caller
// should overwrite tick/sequence/message metadata from the durable row.
func DecodeStored(raw []byte) (game.Input, bool) {
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return game.Input{}, false
	}
	if !IsClientIntent(env.Type) {
		return game.Input{}, false
	}
	return Decode(env.Type, env.MessageID, env.CorrelationID, env.Payload)
}
