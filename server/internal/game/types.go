package game

import (
	"encoding/json"
	"fmt"
	"strconv"
)
// Vec2 is a 2D position in scene units.
type Vec2 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// EntityView is the protocol view of a scene entity.
// HP/MaxHP preserve dead monsters while loot entities omit them.
type EntityView struct {
	ID                string                  `json:"id"`
	Type              string                  `json:"type"`
	Position          Vec2                    `json:"position"`
	HP                *int                    `json:"hp,omitempty"`
	MaxHP             *int                    `json:"max_hp,omitempty"`
	Mana              *int                    `json:"mana,omitempty"`
	MaxMana           *int                    `json:"max_mana,omitempty"`
	CharacterID       string                  `json:"character_id,omitempty"`
	MonsterDefID      string                  `json:"monster_def_id,omitempty"`
	MonsterPackID     string                  `json:"monster_pack_id,omitempty"`
	MonsterPackLeader bool                    `json:"monster_pack_leader,omitempty"`
	IsBoss            bool                    `json:"is_boss,omitempty"`
	BossTemplateID    string                  `json:"boss_template_id,omitempty"`
	VisualModel       string                  `json:"visual_model,omitempty"`
	VisualScale       float64                 `json:"visual_scale,omitempty"`
	VisualTint        string                  `json:"visual_tint,omitempty"`
	BossPhase         *BossPhaseView          `json:"boss_phase,omitempty"`
	ItemDefID         string                  `json:"item_def_id,omitempty"`
	Amount            *int                    `json:"amount,omitempty"`
	ItemTemplateID    string                  `json:"item_template_id,omitempty"`
	DisplayName       string                  `json:"display_name,omitempty"`
	Rarity            string                  `json:"rarity,omitempty"`
	RolledStats       map[string]int          `json:"rolled_stats,omitempty"`
	Requirements      map[string]int          `json:"requirements,omitempty"`
	RequirementStatus []RequirementStatusView `json:"requirement_status,omitempty"`
	RequirementsMet   *bool                   `json:"requirements_met,omitempty"`
	EquipPreview      *EquipPreviewView       `json:"equip_preview,omitempty"`
	EffectIDs         []string                `json:"effect_ids,omitempty"`
	InteractableDefID string                  `json:"interactable_def_id,omitempty"`
	EliteObjective    bool                    `json:"elite_objective,omitempty"`
	QuestReward       bool                    `json:"quest_reward,omitempty"`
	CorpseCharacterID string                  `json:"corpse_character_id,omitempty"`
	CorpseName        string                  `json:"corpse_name,omitempty"`
	CorpseLevel       int                     `json:"corpse_level,omitempty"`
	CorpseItemCount   *int                    `json:"corpse_item_count,omitempty"`
	OwnerID           string                  `json:"owner_id,omitempty"`
	TargetID          string                  `json:"target_id,omitempty"`
	ProjectileDefID   string                  `json:"projectile_def_id,omitempty"`
	State             string                  `json:"state,omitempty"`
}

// ItemView is the protocol view of an inventory item.
type ItemView struct {
	ItemInstanceID    string                  `json:"item_instance_id"`
	ItemDefID         string                  `json:"item_def_id"`
	ItemTemplateID    string                  `json:"item_template_id,omitempty"`
	DisplayName       string                  `json:"display_name,omitempty"`
	Rarity            string                  `json:"rarity,omitempty"`
	RolledStats       map[string]int          `json:"rolled_stats,omitempty"`
	Requirements      map[string]int          `json:"requirements,omitempty"`
	RequirementStatus []RequirementStatusView `json:"requirement_status,omitempty"`
	RequirementsMet   *bool                   `json:"requirements_met,omitempty"`
	EquipPreview      *EquipPreviewView       `json:"equip_preview,omitempty"`
	EffectIDs         []string                `json:"effect_ids,omitempty"`
	SummaryLines      []string                `json:"summary_lines,omitempty"`
	Slot              string                  `json:"slot"`
	Equipped          bool                    `json:"equipped"`
}

// StashItemView is the protocol view of an account-stash item.
type StashItemView struct {
	StashItemID       string                  `json:"stash_item_id"`
	ItemDefID         string                  `json:"item_def_id"`
	ItemTemplateID    string                  `json:"item_template_id,omitempty"`
	DisplayName       string                  `json:"display_name,omitempty"`
	Rarity            string                  `json:"rarity,omitempty"`
	RolledStats       map[string]int          `json:"rolled_stats,omitempty"`
	Requirements      map[string]int          `json:"requirements,omitempty"`
	RequirementStatus []RequirementStatusView `json:"requirement_status,omitempty"`
	RequirementsMet   *bool                   `json:"requirements_met,omitempty"`
	EquipPreview      *EquipPreviewView       `json:"equip_preview,omitempty"`
	EffectIDs         []string                `json:"effect_ids,omitempty"`
	SummaryLines      []string                `json:"summary_lines,omitempty"`
}

// ItemRollPayload is the durable JSON payload stored in rolled_stats columns.
type ItemRollPayload struct {
	ItemTemplateID string         `json:"item_template_id"`
	DisplayName    string         `json:"display_name"`
	Rarity         string         `json:"rarity"`
	Stats          map[string]int `json:"stats"`
	Requirements   map[string]int `json:"requirements"`
	EffectIDs      []string       `json:"effect_ids"`
}

func (p ItemRollPayload) itemViewFields(v *ItemView) {
	if p.ItemTemplateID == "" {
		return
	}
	v.ItemTemplateID = p.ItemTemplateID
	v.DisplayName = p.DisplayName
	v.Rarity = p.Rarity
	v.RolledStats = cloneIntMap(p.Stats)
	v.Requirements = cloneIntMap(p.Requirements)
	v.EffectIDs = cloneStringSlice(p.EffectIDs)
}

func (p ItemRollPayload) stashItemViewFields(v *StashItemView) {
	if p.ItemTemplateID == "" {
		return
	}
	v.ItemTemplateID = p.ItemTemplateID
	v.DisplayName = p.DisplayName
	v.Rarity = p.Rarity
	v.RolledStats = cloneIntMap(p.Stats)
	v.Requirements = cloneIntMap(p.Requirements)
	v.EffectIDs = cloneStringSlice(p.EffectIDs)
}

// RollPayload returns the durable payload represented by optional rolled item
// fields in this protocol view.
func (v ItemView) RollPayload() *ItemRollPayload {
	if v.ItemTemplateID == "" {
		return nil
	}
	return &ItemRollPayload{
		ItemTemplateID: v.ItemTemplateID,
		DisplayName:    v.DisplayName,
		Rarity:         v.Rarity,
		Stats:          cloneIntMap(v.RolledStats),
		Requirements:   cloneIntMap(v.Requirements),
		EffectIDs:      cloneStringSlice(v.EffectIDs),
	}
}

// RollPayload returns the durable payload represented by optional rolled stash
// item fields.
func (v StashItemView) RollPayload() *ItemRollPayload {
	if v.ItemTemplateID == "" {
		return nil
	}
	return &ItemRollPayload{
		ItemTemplateID: v.ItemTemplateID,
		DisplayName:    v.DisplayName,
		Rarity:         v.Rarity,
		Stats:          cloneIntMap(v.RolledStats),
		Requirements:   cloneIntMap(v.Requirements),
		EffectIDs:      cloneStringSlice(v.EffectIDs),
	}
}

// BaseStatsView is the protocol view of allocated base stats.
type BaseStatsView struct {
	Str   int `json:"str"`
	Dex   int `json:"dex"`
	Vit   int `json:"vit"`
	Magic int `json:"magic"`
}

// DerivedStatsView is the protocol view of stat-derived combat/display values.
type DerivedStatsView struct {
	DamageMin            float64 `json:"damage_min"`
	DamageMax            float64 `json:"damage_max"`
	Armor                float64 `json:"armor"`
	AttackSpeed          float64 `json:"attack_speed"`
	AttackIntervalTicks  int     `json:"attack_interval_ticks"`
	HitChance            float64 `json:"hit_chance"`
	CritChance           float64 `json:"crit_chance"`
	CritDamage           float64 `json:"crit_damage"`
	MovementSpeed        float64 `json:"movement_speed"`
	MaxHP                float64 `json:"max_hp"`
	MaxMana              float64 `json:"max_mana"`
	HealthRegenPerSecond float64 `json:"health_regen_per_second"`
	ManaRegenPerSecond   float64 `json:"mana_regen_per_second"`
}

// RequirementStatusView is the server-authored usability state for one item
// requirement against the current character.
type RequirementStatusView struct {
	Stat     string `json:"stat"`
	Required int    `json:"required"`
	Current  int    `json:"current"`
	Met      bool   `json:"met"`
}

// EquipPreviewDeltaView describes one derived-stat change if an item were
// equipped through the same server stat path used by combat.
type EquipPreviewDeltaView struct {
	Stat    string  `json:"stat"`
	Current float64 `json:"current"`
	Preview float64 `json:"preview"`
	Delta   float64 `json:"delta"`
}

// EquipPreviewView is a server-authored equipment preview rendered by the
// inventory and shop UI.
type EquipPreviewView struct {
	Slot            string                  `json:"slot"`
	RequirementsMet bool                    `json:"requirements_met"`
	Deltas          []EquipPreviewDeltaView `json:"deltas"`
}

// StatBreakdownSourceView is one source row for an effective combat stat.
type StatBreakdownSourceView struct {
	Label          string  `json:"label"`
	Value          float64 `json:"value"`
	Kind           string  `json:"kind"`
	ItemInstanceID string  `json:"item_instance_id,omitempty"`
}

// StatBreakdownView explains how one effective stat was assembled.
type StatBreakdownView struct {
	Key           string                    `json:"key"`
	Value         float64                   `json:"value"`
	UncappedValue float64                   `json:"uncapped_value,omitempty"`
	Cap           *float64                  `json:"cap"`
	Sources       []StatBreakdownSourceView `json:"sources"`
}

// CharacterProgressionView is the protocol view of durable character XP/stat
// progression plus derived display stats.
type CharacterProgressionView struct {
	CharacterClass        string              `json:"character_class"`
	Level                 int                 `json:"level"`
	Experience            int                 `json:"experience"`
	ExperienceToNextLevel *int                `json:"experience_to_next_level"`
	LevelCap              int                 `json:"level_cap"`
	UnspentStatPoints     int                 `json:"unspent_stat_points"`
	UnspentSkillPoints    int                 `json:"-"`
	Gold                  int                 `json:"gold"`
	DeepestDungeonDepth   int                 `json:"deepest_dungeon_depth"`
	BaseStats             BaseStatsView       `json:"base_stats"`
	DerivedStats          DerivedStatsView    `json:"derived_stats"`
	StatBreakdowns        []StatBreakdownView `json:"stat_breakdowns,omitempty"`
	SkillRanks            map[string]int      `json:"-"`
}

// SkillProgressionSkillView is one skill row in the server-owned skill
// progression state.
type SkillProgressionSkillView struct {
	SkillID  string `json:"skill_id"`
	Rank     int    `json:"rank"`
	MaxRank  int    `json:"max_rank"`
	CanSpend bool   `json:"can_spend"`
}

// SkillProgressionView is the protocol view of spendable skill points and
// server-known ranks.
type SkillProgressionView struct {
	UnspentSkillPoints int                         `json:"unspent_skill_points"`
	Skills             []SkillProgressionSkillView `json:"skills"`
}

// SkillCooldownView is one active skill cooldown row.
type SkillCooldownView struct {
	SkillID        string `json:"skill_id"`
	RemainingTicks int    `json:"remaining_ticks"`
	TotalTicks     int    `json:"total_ticks"`
}

// ShopOfferView is one server-authoritative offer rendered inside shop events.
type ShopOfferView struct {
	OfferID           string                  `json:"offer_id"`
	Kind              string                  `json:"kind"`
	Concealed         bool                    `json:"concealed,omitempty"`
	MysteryLabel      string                  `json:"mystery_label,omitempty"`
	ItemDefID         string                  `json:"item_def_id,omitempty"`
	ItemTemplateID    string                  `json:"item_template_id,omitempty"`
	DisplayName       string                  `json:"display_name,omitempty"`
	Rarity            string                  `json:"rarity,omitempty"`
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

// HotbarSlotView is one fixed hotbar assignment in the protocol snapshot.
type HotbarSlotView struct {
	SlotIndex      int       `json:"slot_index"`
	ItemInstanceID *string   `json:"item_instance_id"`
	Item           *ItemView `json:"item,omitempty"`
}

type WallView struct {
	ID       string `json:"id"`
	Position Vec2   `json:"position"`
	Size     Vec2   `json:"size"`
	Source   string `json:"source,omitempty"`
}

// BossPhaseView is the protocol view of an authoritative boss pattern phase.
type BossPhaseView struct {
	PatternID     string             `json:"pattern_id"`
	PhaseIndex    int                `json:"phase_index"`
	PhaseKind     string             `json:"phase_kind"`
	StartedTick   uint64             `json:"started_tick"`
	DurationTicks int                `json:"duration_ticks"`
	Telegraph     *BossTelegraphView `json:"telegraph,omitempty"`
	HitShape      *BossHitShapeView  `json:"hit_shape,omitempty"`
}

// BossTelegraphView describes the warning data clients render before damage.
type BossTelegraphView struct {
	Type      string  `json:"type"`
	FromColor string  `json:"from_color,omitempty"`
	ToColor   string  `json:"to_color,omitempty"`
	HitShape  string  `json:"hit_shape,omitempty"`
	Radius    float64 `json:"radius,omitempty"`
	Width     float64 `json:"width,omitempty"`
}

// BossHitShapeView describes the authoritative active hit predicate.
type BossHitShapeView struct {
	Shape  string  `json:"shape"`
	Radius float64 `json:"radius,omitempty"`
	Width  float64 `json:"width,omitempty"`
}

// Event is an authoritative event emitted by the sim.
type Event struct {
	EventType          string                  `json:"event_type"`
	EntityID           string                  `json:"entity_id,omitempty"`
	SourceEntityID     string                  `json:"source_entity_id,omitempty"`
	TargetEntityID     string                  `json:"target_entity_id,omitempty"`
	MonsterDefID       string                  `json:"monster_def_id,omitempty"`
	BossTemplateID     string                  `json:"boss_template_id,omitempty"`
	CorrelationID      string                  `json:"correlation_id,omitempty"`
	Damage             *int                    `json:"damage,omitempty"`
	DamageType         string                  `json:"damage_type,omitempty"`
	Outcome            string                  `json:"outcome,omitempty"`
	RawDamage          *int                    `json:"raw_damage,omitempty"`
	MitigatedDamage    *int                    `json:"mitigated_damage,omitempty"`
	Blocked            *bool                   `json:"blocked,omitempty"`
	Critical           *bool                   `json:"critical,omitempty"`
	Heal               *int                    `json:"heal,omitempty"`
	Mana               *int                    `json:"mana,omitempty"`
	ItemInstanceID     string                  `json:"item_instance_id,omitempty"`
	Level              *int                    `json:"level,omitempty"`
	FromLevel          *int                    `json:"from_level,omitempty"`
	ToLevel            *int                    `json:"to_level,omitempty"`
	Amount             *int                    `json:"amount,omitempty"`
	TotalExperience    *int                    `json:"total_experience,omitempty"`
	TotalGold          *int                    `json:"total_gold,omitempty"`
	Stat               string                  `json:"stat,omitempty"`
	UnspentStatPoints  *int                    `json:"unspent_stat_points,omitempty"`
	UnspentSkillPoints *int                    `json:"unspent_skill_points,omitempty"`
	SkillID            string                  `json:"skill_id,omitempty"`
	Rank               *int                    `json:"rank,omitempty"`
	MaxRank            *int                    `json:"max_rank,omitempty"`
	RemainingTicks     *int                    `json:"remaining_ticks,omitempty"`
	TotalTicks         *int                    `json:"total_ticks,omitempty"`
	ProjectileDefID    string                  `json:"projectile_def_id,omitempty"`
	Position           *Vec2                   `json:"position,omitempty"`
	Direction          *Vec2                   `json:"direction,omitempty"`
	Range              *float64                `json:"range,omitempty"`
	AngleDegrees       *float64                `json:"angle_degrees,omitempty"`
	WeaponSlot         string                  `json:"weapon_slot,omitempty"`
	WeaponSet          *int                    `json:"weapon_set,omitempty"`
	Reason             string                  `json:"reason,omitempty"`
	ShopID             string                  `json:"shop_id,omitempty"`
	Service            string                  `json:"service,omitempty"`
	Offers             []ShopOfferView         `json:"offers,omitempty"`
	SellAppraisals     []ShopSellAppraisalView `json:"sell_appraisals,omitempty"`
	OfferID            string                  `json:"offer_id,omitempty"`
	Price              *int                    `json:"price,omitempty"`
	Affordable         *bool                   `json:"affordable,omitempty"`
	RefreshKey         string                  `json:"refresh_key,omitempty"`
	Item               *ItemView               `json:"item,omitempty"`
	StashID            string                  `json:"stash_id,omitempty"`
	StashItemID        string                  `json:"stash_item_id,omitempty"`
	StashItems         []StashItemView         `json:"stash_items,omitempty"`
	StashGold          *int                    `json:"stash_gold,omitempty"`
	StashCapacity      *int                    `json:"stash_capacity,omitempty"`
	CorpseCharacterID  string                  `json:"corpse_character_id,omitempty"`
	CorpseName         string                  `json:"corpse_name,omitempty"`
	CorpseItems        []ItemView              `json:"corpse_items,omitempty"`
	Inventory          []ItemView              `json:"inventory,omitempty"`
	Equipped           map[string]*string      `json:"equipped,omitempty"`
	Hotbar             []HotbarSlotView        `json:"hotbar,omitempty"`
	Gold               *int                    `json:"gold,omitempty"`
	InventoryRows      *int                    `json:"inventory_rows,omitempty"`
	InventoryCapacity  *int                    `json:"inventory_capacity,omitempty"`
	HotbarCapacity     *int                    `json:"hotbar_capacity,omitempty"`
	CharacterClass     string                  `json:"character_class,omitempty"`
	CharacterLevel     *int                    `json:"character_level,omitempty"`
	CharacterXP        *int                    `json:"character_xp,omitempty"`
	PatternID          string                  `json:"pattern_id,omitempty"`
	PhaseIndex         *int                    `json:"phase_index,omitempty"`
	PhaseKind          string                  `json:"phase_kind,omitempty"`
	DurationTicks      *int                    `json:"duration_ticks,omitempty"`
	Telegraph          *BossTelegraphView      `json:"telegraph,omitempty"`
	HitShape           *BossHitShapeView       `json:"hit_shape,omitempty"`
	State              string                  `json:"state,omitempty"`
}

// TeleporterDiscoveryView is the protocol view of a generated dungeon level's
// waypoint discovery state.
type TeleporterDiscoveryView struct {
	Level      int  `json:"level"`
	Discovered bool `json:"discovered"`
}

// PartyMemberView describes one co-op member in recipient-scoped snapshots.
type PartyMemberView struct {
	PlayerID     string `json:"player_id"`
	CharacterID  string `json:"character_id"`
	DisplayName  string `json:"display_name"`
	Role         string `json:"role"`
	Connected    bool   `json:"connected"`
	CurrentLevel int    `json:"current_level"`
}

// SkillBindingsView is the client control layout for active skill shortcuts.
type SkillBindingsView struct {
	FunctionKeys      []string `json:"function_keys"`
	RightClickSkillID string   `json:"right_click_skill_id"`
}

// Snapshot is the full authoritative state for rendering (session_snapshot).
type Snapshot struct {
	ServerTick            uint64                    `json:"server_tick"`
	SessionID             string                    `json:"session_id"`
	Seed                  string                    `json:"seed"`
	CurrentLevel          int                       `json:"current_level"`
	LocalPlayerID         string                    `json:"local_player_id,omitempty"`
	Party                 []PartyMemberView         `json:"party,omitempty"`
	Walls                 []WallView                `json:"walls"`
	Entities              []EntityView              `json:"entities"`
	Inventory             []ItemView                `json:"inventory"`
	Equipped              map[string]*string        `json:"equipped"`
	ActiveWeaponSet       int                       `json:"active_weapon_set"`
	WeaponSets            []WeaponSetView           `json:"weapon_sets"`
	HotbarCapacity        int                       `json:"hotbar_capacity"`
	Hotbar                []HotbarSlotView          `json:"hotbar"`
	InventoryRows         int                       `json:"inventory_rows"`
	InventoryCapacity     int                       `json:"inventory_capacity"`
	Gold                  int                       `json:"gold"`
	StashItems            []StashItemView           `json:"stash_items"`
	StashGold             int                       `json:"stash_gold"`
	StashCapacity         int                       `json:"stash_capacity"`
	DiscoveredTeleporters []TeleporterDiscoveryView `json:"discovered_teleporters"`
	CharacterProgression  CharacterProgressionView  `json:"character_progression"`
	SkillProgression      SkillProgressionView      `json:"skill_progression"`
	SkillCooldowns        []SkillCooldownView       `json:"skill_cooldowns"`
	SkillBindings         SkillBindingsView         `json:"skill_bindings"`
	RecentEvents          []Event                   `json:"recent_events"`
}

type WeaponSetView struct {
	Index    int     `json:"index"`
	MainHand *string `json:"main_hand"`
	OffHand  *string `json:"off_hand"`
}

// Change operation names (the state_delta ops).
const (
	OpEntitySpawn                = "entity_spawn"
	OpEntityUpdate               = "entity_update"
	OpEntityRemove               = "entity_remove"
	OpInventoryAdd               = "inventory_add"
	OpInventoryUpdate            = "inventory_update"
	OpInventoryRemove            = "inventory_remove"
	OpEquippedUpdate             = "equipped_update"
	OpWeaponSetUpdate            = "weapon_set_update"
	OpHotbarUpdate               = "hotbar_update"
	OpGoldUpdate                 = "gold_update"
	OpStashItemAdd               = "stash_item_add"
	OpStashItemRemove            = "stash_item_remove"
	OpStashGoldUpdate            = "stash_gold_update"
	OpWallLayoutUpdate           = "wall_layout_update"
	OpTeleporterDiscoveryUpdate  = "teleporter_discovery_update"
	OpCharacterProgressionUpdate = "character_progression_update"
	OpSkillProgressionUpdate     = "skill_progression_update"
	OpSkillCooldownUpdate        = "skill_cooldown_update"
	OpSkillBindingsUpdate        = "skill_bindings_update"
	OpShopStockReplace           = "shop_stock_replace"
	OpShopStockAvailability      = "shop_stock_availability_update"
)

// Change is one ordered authoritative change within a tick. It marshals to
// exactly the fields required for its op (matching the state_delta schema's
// oneOf), which is why it has a custom MarshalJSON.
type Change struct {
	Op               string
	OwnerPlayerID    uint64
	Entity           *EntityView
	EntityID         string
	Item             *ItemView
	StashItem        *StashItemView
	StashItemID      string
	Slot             string
	ItemInstanceID   *string // for equipped_update; nil marshals as null
	ActiveWeaponSet  *int
	WeaponSets       []WeaponSetView
	SlotIndex        int
	HotbarCapacity   *int
	InventoryRows    *int
	InventoryCap     *int
	Gold             *int
	StashGold        *int
	Walls            []WallView
	Level            int
	Discovered       bool
	Progression      *CharacterProgressionView
	SkillProgression *SkillProgressionView
	SkillCooldowns   []SkillCooldownView
	SkillBindings    *SkillBindingsView
	ShopID           string
	RefreshKey       string
	ShopStock        []PersistedShopStockItem
	OfferID          string
	Available        bool
	StashTransferID  string
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
			HotbarCapacity *int    `json:"hotbar_capacity,omitempty"`
			InventoryRows  *int    `json:"inventory_rows,omitempty"`
			InventoryCap   *int    `json:"inventory_capacity,omitempty"`
		}{c.Op, c.Slot, c.ItemInstanceID, c.HotbarCapacity, c.InventoryRows, c.InventoryCap})
	case OpWeaponSetUpdate:
		return json.Marshal(struct {
			Op              string          `json:"op"`
			ActiveWeaponSet *int            `json:"active_weapon_set"`
			WeaponSets      []WeaponSetView `json:"weapon_sets"`
		}{c.Op, c.ActiveWeaponSet, c.WeaponSets})
	case OpHotbarUpdate:
		return json.Marshal(struct {
			Op             string    `json:"op"`
			SlotIndex      int       `json:"slot_index"`
			ItemInstanceID *string   `json:"item_instance_id"`
			Item           *ItemView `json:"item,omitempty"`
			InventoryRows  *int      `json:"inventory_rows,omitempty"`
			InventoryCap   *int      `json:"inventory_capacity,omitempty"`
		}{c.Op, c.SlotIndex, c.ItemInstanceID, c.Item, c.InventoryRows, c.InventoryCap})
	case OpGoldUpdate:
		gold := 0
		if c.Gold != nil {
			gold = *c.Gold
		}
		return json.Marshal(struct {
			Op   string `json:"op"`
			Gold int    `json:"gold"`
		}{c.Op, gold})
	case OpStashItemAdd:
		return json.Marshal(struct {
			Op   string         `json:"op"`
			Item *StashItemView `json:"item"`
		}{c.Op, c.StashItem})
	case OpStashItemRemove:
		return json.Marshal(struct {
			Op          string `json:"op"`
			StashItemID string `json:"stash_item_id"`
		}{c.Op, c.StashItemID})
	case OpStashGoldUpdate:
		stashGold := 0
		if c.StashGold != nil {
			stashGold = *c.StashGold
		}
		return json.Marshal(struct {
			Op        string `json:"op"`
			StashGold int    `json:"stash_gold"`
		}{c.Op, stashGold})
	case OpWallLayoutUpdate:
		return json.Marshal(struct {
			Op    string     `json:"op"`
			Walls []WallView `json:"walls"`
		}{c.Op, c.Walls})
	case OpTeleporterDiscoveryUpdate:
		return json.Marshal(struct {
			Op         string `json:"op"`
			Level      int    `json:"level"`
			Discovered bool   `json:"discovered"`
		}{c.Op, c.Level, c.Discovered})
	case OpCharacterProgressionUpdate:
		return json.Marshal(struct {
			Op          string                    `json:"op"`
			Progression *CharacterProgressionView `json:"character_progression"`
		}{c.Op, c.Progression})
	case OpSkillProgressionUpdate:
		return json.Marshal(struct {
			Op               string                `json:"op"`
			SkillProgression *SkillProgressionView `json:"skill_progression"`
		}{c.Op, c.SkillProgression})
	case OpSkillCooldownUpdate:
		return json.Marshal(struct {
			Op             string              `json:"op"`
			SkillCooldowns []SkillCooldownView `json:"skill_cooldowns"`
		}{c.Op, c.SkillCooldowns})
	case OpSkillBindingsUpdate:
		return json.Marshal(struct {
			Op            string             `json:"op"`
			SkillBindings *SkillBindingsView `json:"skill_bindings"`
		}{c.Op, c.SkillBindings})
	case OpShopStockReplace:
		return json.Marshal(struct {
			Op         string                   `json:"op"`
			ShopID     string                   `json:"shop_id"`
			RefreshKey string                   `json:"refresh_key"`
			Stock      []PersistedShopStockItem `json:"stock"`
		}{c.Op, c.ShopID, c.RefreshKey, c.ShopStock})
	case OpShopStockAvailability:
		return json.Marshal(struct {
			Op        string `json:"op"`
			ShopID    string `json:"shop_id"`
			OfferID   string `json:"offer_id"`
			Available bool   `json:"available"`
		}{c.Op, c.ShopID, c.OfferID, c.Available})
	default:
		return nil, &json.UnsupportedValueError{Str: "unknown change op: " + c.Op}
	}
}

// idStr renders an unsigned 64-bit entity id as a decimal string.
func idStr(id uint64) string { return strconv.FormatUint(id, 10) }

func wallID(levelNum int, index int) string {
	return fmt.Sprintf("wall_%d_%04d", levelNum, index)
}
