package game

import "encoding/json"

// Shop/vendor protocol view types. Extracted from types.go to keep that protocol
// view-type file off its maintainability ceiling; these are pure wire-serialization
// structs with no behavior, so the file they live in does not affect their identity.

type ShopOfferView struct {
	OfferID           string                  `json:"offer_id"`
	Kind              string                  `json:"kind"`
	Concealed         bool                    `json:"concealed,omitempty"`
	MysteryLabel      string                  `json:"mystery_label,omitempty"`
	ItemDefID         string                  `json:"item_def_id,omitempty"`
	ItemTemplateID    string                  `json:"item_template_id,omitempty"`
	DisplayName       string                  `json:"display_name,omitempty"`
	Rarity            string                  `json:"rarity,omitempty"`
	ItemLevel         int                     `json:"item_level,omitempty"`
	Slot              string                  `json:"slot,omitempty"`
	Category          string                  `json:"category,omitempty"`
	RolledStats       map[string]int          `json:"rolled_stats,omitempty"`
	Requirements      map[string]int          `json:"requirements,omitempty"`
	RequirementStatus []RequirementStatusView `json:"requirement_status,omitempty"`
	RequirementsMet   *bool                   `json:"requirements_met,omitempty"`
	EquipPreview      *EquipPreviewView       `json:"equip_preview,omitempty"`
	EffectIDs         []string                `json:"effect_ids,omitempty"`
	BuyPrice          int                     `json:"buy_price"`
	SummaryLines      []string                `json:"summary_lines,omitempty"`
	Comparison        *ShopComparisonView     `json:"comparison,omitempty"`
	Source            string                  `json:"source,omitempty"`
	Depth             int                     `json:"depth,omitempty"`
	SourceDepth       int                     `json:"source_depth,omitempty"`
	SourceDepthMin    int                     `json:"source_depth_min,omitempty"`
	SourceDepthMax    int                     `json:"source_depth_max,omitempty"`
}

// PersistedShopStockItem is a generated shop-stock row carried between the
// store/replay boundary and the sim. Runtime buyback rows are intentionally not
// represented by this type.
type PersistedShopStockItem struct {
	ShopID         string          `json:"shop_id"`
	RefreshKey     string          `json:"refresh_key"`
	OfferIndex     int             `json:"offer_index"`
	OfferID        string          `json:"offer_id"`
	SourceDepth    int             `json:"source_depth"`
	ItemTemplateID string          `json:"item_template_id"`
	RolledPayload  json.RawMessage `json:"rolled_payload"`
	BuyPrice       int             `json:"buy_price"`
	Available      bool            `json:"available"`
}

// ShopComparisonDeltaView describes one direct stat comparison between a
// vendor item and the actor's currently equipped item in the same slot.
type ShopComparisonDeltaView struct {
	Stat     string `json:"stat"`
	Offered  int    `json:"offered"`
	Equipped int    `json:"equipped"`
	Delta    int    `json:"delta"`
}

// ShopComparisonView is server-authored comparison data rendered by the shop UI.
type ShopComparisonView struct {
	Slot                   string                    `json:"slot"`
	EquippedItemInstanceID string                    `json:"equipped_item_instance_id,omitempty"`
	Deltas                 []ShopComparisonDeltaView `json:"deltas"`
}

// ShopSellAppraisalView is one server-authored sell quote for an unequipped
// inventory item at the currently opened vendor.
type ShopSellAppraisalView struct {
	ItemInstanceID    string                  `json:"item_instance_id"`
	ItemDefID         string                  `json:"item_def_id"`
	ItemTemplateID    string                  `json:"item_template_id,omitempty"`
	DisplayName       string                  `json:"display_name"`
	Rarity            string                  `json:"rarity,omitempty"`
	ItemLevel         int                     `json:"item_level,omitempty"`
	Slot              string                  `json:"slot,omitempty"`
	Category          string                  `json:"category,omitempty"`
	RolledStats       map[string]int          `json:"rolled_stats,omitempty"`
	Requirements      map[string]int          `json:"requirements,omitempty"`
	RequirementStatus []RequirementStatusView `json:"requirement_status,omitempty"`
	RequirementsMet   *bool                   `json:"requirements_met,omitempty"`
	EquipPreview      *EquipPreviewView       `json:"equip_preview,omitempty"`
	EffectIDs         []string                `json:"effect_ids,omitempty"`
	SellPrice         int                     `json:"sell_price"`
	SummaryLines      []string                `json:"summary_lines,omitempty"`
	Comparison        *ShopComparisonView     `json:"comparison,omitempty"`
}
