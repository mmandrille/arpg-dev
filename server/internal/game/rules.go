package game

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Rules is the in-memory form of the shared rules-as-data (shared/rules). The
// Go server and the Godot client read the same files (ADR-0001 D6); this is the
// server's loader and typed view.
type Rules struct {
	MainConfig           MainConfig
	Combat               Combat
	Navigation           NavigationRules
	Items                map[string]ItemDef
	ItemTemplates        map[string]ItemTemplateDef
	SetCatalogs          map[string]SetItemCatalogDef
	UniqueItems          map[string]UniqueItemDef
	SetItems             map[string]SetItemDef
	UniqueEffects        map[string]UniqueEffectDef
	StatusEffects        map[string]StatusEffectDef
	Skills               map[string]SkillDef
	Rarities             map[string]RarityDef
	RarityOrder          []string
	TreasureClasses      map[string]TreasureClassDef
	CharacterProgression CharacterProgressionRules
	Monsters             map[string]MonsterDef
	Mercenaries          MercenaryRules
	LootTables           map[string]LootTable
	Shops                map[string]ShopDef
	Interactables        map[string]InteractableDef
	Worlds               map[string]WorldDef
	DungeonGeneration    DungeonGenerationRules
	QuestSteward         QuestStewardRules
	BossTemplates        map[string]BossTemplateDef
	BossPatterns         map[string]BossPatternDef
}

// MainConfig holds high-level gameplay tuning values that are being promoted
// into one designer-facing file before older rules consume them directly.
type MainConfig struct {
	Gameplay             MainGameplayConfig         `json:"gameplay"`
	WeaponElementalProcs WeaponElementalProcsConfig `json:"weapon_elemental_procs"`
}

type MainGameplayConfig struct {
	BaseAttackIntervalTicks                    int                     `json:"base_attack_interval_ticks"`
	BaseMovementSpeed                          float64                 `json:"base_movement_speed"`
	MovementAccelerationSeconds                float64                 `json:"movement_acceleration_seconds"`
	MovementMinSpeedFactor                     float64                 `json:"movement_min_speed_factor"`
	MinimumMonsterAggroRadius                  float64                 `json:"minimum_monster_aggro_radius"`
	BaseDropRatePercent                        int                     `json:"base_drop_rate_percent"`
	BossBonusDropRatePercent                   int                     `json:"boss_bonus_drop_rate_percent"`
	BossExtraDropRatePercent                   int                     `json:"boss_extra_drop_rate_percent"`
	BossLootMagicFindBonusPercent              int                     `json:"boss_loot_magic_find_bonus_percent"`
	RespecCostGold                             int                     `json:"respec_cost_gold"`
	BishopRespecResourceID                     string                  `json:"bishop_respec_resource_item_def_id"`
	BishopRespecResourceCost                   int                     `json:"bishop_respec_resource_count"`
	BishopReviveResourceID                     string                  `json:"bishop_revive_resource_item_def_id"`
	BishopReviveResourceCost                   int                     `json:"bishop_revive_resource_count"`
	ItemUpgradeCostGold                        int                     `json:"item_upgrade_cost_gold"`
	ItemUpgradeCostGrowth                      int                     `json:"item_upgrade_cost_growth_per_level"`
	ItemUpgradeMaxLevel                        int                     `json:"item_upgrade_max_level"`
	ItemUpgradeSuccessPct                      int                     `json:"item_upgrade_success_chance_percent"`
	ItemUpgradeFailureCurve                    ItemUpgradeFailureCurve `json:"item_upgrade_failure_curve"`
	ItemUpgradeShardSuccessBonusPercentPerTier int                     `json:"item_upgrade_shard_success_bonus_percent_per_tier"`
	ItemUpgradePityFailures                    int                     `json:"item_upgrade_pity_failure_threshold"`
	ItemUpgradeResourceID                      string                  `json:"item_upgrade_resource_item_def_id"`
	ItemUpgradeResourceCost                    int                     `json:"item_upgrade_resource_count"`
	ResourceLootDrops                          ResourceLootDropsConfig `json:"resource_loot_drops"`
	BadgeRewardRules                           []BadgeRewardRule       `json:"badge_reward_rules"`
	MercenaryHireCostGold                      int                     `json:"mercenary_hire_cost_gold"`
	MercenaryHireCostGoldPerLevel              int                     `json:"mercenary_hire_cost_gold_per_level"`
	QuestTurnInItemDefID                       string                  `json:"quest_turn_in_item_def_id"`
	QuestTurnInRewardGold                      int                     `json:"quest_turn_in_reward_gold"`
	CompanionAssistRadius                      float64                 `json:"companion_assist_radius"`
	CompanionFollowDistance                    float64                 `json:"companion_follow_distance"`
	CompanionFollowStop                        float64                 `json:"companion_follow_stop_radius"`
	RangedRetreatMinMeleeEngagementSeconds     float64                 `json:"ranged_retreat_min_melee_engagement_seconds"`
}

// DamageRange is an inclusive [Min, Max] integer range.
type DamageRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// WeaponDamageScalingDef maps an item type to derived-stat keys for weapon damage bonuses.
type WeaponDamageScalingDef struct {
	DamageMin   string `json:"damage_min"`
	DamageMax   string `json:"damage_max"`
	SourceLabel string `json:"source_label"`
}

// Combat holds combat parameters.
type Combat struct {
	BaseHitChance                     float64                           `json:"base_hit_chance"`
	BaseCritChance                    float64                           `json:"base_crit_chance"`
	BaseCritDamage                    float64                           `json:"base_crit_damage"`
	MinimumDamage                     int                               `json:"minimum_damage"`
	BlockCap                          int                               `json:"block_cap_percent"`
	BaseAttackIntervalTicks           int                               `json:"base_attack_interval_ticks"`
	MinEffectiveAttackSpeed           float64                           `json:"min_effective_attack_speed"`
	MaxEffectiveAttackSpeed           float64                           `json:"max_effective_attack_speed"`
	PlayerDamage                      DamageRange                       `json:"player_damage"`
	TwoHandedStrengthDamageMultiplier float64                           `json:"two_handed_strength_damage_multiplier"`
	WeaponDamageScaling               map[string]WeaponDamageScalingDef `json:"weapon_damage_scaling"`
	UnarmedReach                      float64                           `json:"unarmed_reach"`
	Coop                              CoopCombat                        `json:"coop"`
}

type CoopCombat struct {
	XPShare        CoopXPShareRules        `json:"xp_share"`
	PartyChallenge CoopPartyChallengeRules `json:"party_challenge"`
}

type CoopXPShareRules struct {
	Enabled                 bool    `json:"enabled"`
	Radius                  float64 `json:"radius"`
	FullXPPerEligiblePlayer bool    `json:"full_xp_per_eligible_player"`
	IncludeDeadPlayers      bool    `json:"include_dead_players"`
	IncludeDisconnected     bool    `json:"include_disconnected_players"`
}

type CoopPartyChallengeRules struct {
	Enabled              bool    `json:"enabled"`
	PerDoubleBonus       float64 `json:"per_double_bonus"`
	MaxBonus             float64 `json:"max_bonus"`
	HPScalesAtSpawn      bool    `json:"hp_scales_at_spawn"`
	DamageScalesAtAttack bool    `json:"damage_scales_at_attack"`
}

func (r CoopPartyChallengeRules) Multiplier(partyCount int) float64 {
	if !r.Enabled || partyCount <= 1 {
		return 1
	}
	bonus := r.PerDoubleBonus * math.Log2(float64(partyCount))
	if bonus > r.MaxBonus {
		bonus = r.MaxBonus
	}
	if bonus < 0 {
		bonus = 0
	}
	return 1 + bonus
}

// NavigationRules bounds server-owned auto-navigation.
type NavigationRules struct {
	CellSize                              float64    `json:"cell_size"`
	MaxAutoSteps                          int        `json:"max_auto_steps"`
	PathPlanningHorizonSeconds            float64    `json:"path_planning_horizon_seconds"`
	PlayerMaxAutoSteps                    int        `json:"player_max_auto_steps"`
	PlayerPathNodesPerSearch              int        `json:"player_path_nodes_per_search"`
	PlayerPathNodesPerTick                int        `json:"player_path_nodes_per_tick"`
	GridBounds                            GridBounds `json:"grid_bounds"`
	StopDistance                          float64    `json:"stop_distance"`
	MonsterPathRequestsPerTick            int        `json:"monster_path_requests_per_tick"`
	MonsterPathNodesPerSearch             int        `json:"monster_path_nodes_per_search"`
	MonsterPathNodesPerTick               int        `json:"monster_path_nodes_per_tick"`
	MonsterPathCacheTicks                 int        `json:"monster_path_cache_ticks"`
	MonsterRepathThrottleTicks            int        `json:"monster_repath_throttle_ticks"`
	MonsterRepathStaggerTicks             int        `json:"monster_repath_stagger_ticks"`
	MonsterPathCacheGoalTolerance         float64    `json:"monster_path_cache_goal_tolerance"`
	MonsterMovementLODMinLiveMonsters     int        `json:"monster_movement_lod_min_live_monsters"`
	MonsterMovementLODNearDistance        float64    `json:"monster_movement_lod_near_distance"`
	MonsterMovementLODUpdateIntervalTicks int        `json:"monster_movement_lod_update_interval_ticks"`
	MonsterOverloadDegradeTicks           int        `json:"monster_overload_degrade_ticks"`
	MonsterOverloadLiveMonsterThreshold   int        `json:"monster_overload_live_monster_threshold"`
}

// CharacterProgressionRules controls XP thresholds, level-up points, base
// stats, and derived-stat formulas.
type CharacterProgressionRules struct {
	BaseStats            BaseStatsView
	Classes              map[string]CharacterClassDef
	PointsPerLevel       int
	SkillPoints          SkillPointRules
	SkillRankScaling     RankScalingCurve
	SkillManaScaling     RankScalingCurve
	LevelCap             int
	XPThresholds         map[int]int
	DerivedStats         map[string]LinearStatFormula
	LevelUpRestoreHPMana bool
}

type CharacterClassDef struct {
	Name              string        `json:"name"`
	LightRadius       float64       `json:"light_radius"`
	BaseStats         BaseStatsView `json:"base_stats"`
	BaseMovementSpeed float64       `json:"base_movement_speed"`
	LevelStatGrowth   BaseStatsView `json:"level_stat_growth"`
}

// SkillPointRules controls deterministic skill-point grants on level-up.
type SkillPointRules struct {
	PointsPerGrant     int `json:"points_per_grant"`
	GrantEveryLevels   int `json:"grant_every_levels"`
	FirstGrantLevel    int `json:"first_grant_level"`
	SecondGrantLevel   int `json:"second_grant_level"`
	GrantEveryMinLevel int `json:"grant_every_min_level"`
}

type LinearStatFormula struct {
	Type        string   `json:"type"`
	Base        float64  `json:"base"`
	PerStr      float64  `json:"per_str"`
	PerDex      float64  `json:"per_dex"`
	PerVit      float64  `json:"per_vit"`
	PerMagic    float64  `json:"per_magic"`
	Stat        string   `json:"stat"`
	Scale       float64  `json:"scale"`
	Offset      float64  `json:"offset"`
	Denominator float64  `json:"denominator"`
	Min         *float64 `json:"min"`
	Max         *float64 `json:"max"`
}

type BossTemplateDef struct {
	Name             string         `json:"name"`
	BaseMonsterDefID string         `json:"base_monster_def_id"`
	PatternDeck      []string       `json:"pattern_deck"`
	HPMultiplier     float64        `json:"hp_multiplier"`
	DamageMultiplier float64        `json:"damage_multiplier"`
	Enrage           *BossEnrageDef `json:"enrage,omitempty"`
	LootTable        string         `json:"loot_table"`
	Visual           BossVisualDef  `json:"visual"`
}

type BossVisualDef struct {
	Model     string   `json:"model"`
	Color     string   `json:"color"`
	Scale     float64  `json:"scale"`
	ModelPool []string `json:"model_pool"`
}

type BossPatternDef struct {
	Phases        []BossPatternPhase `json:"phases"`
	CooldownTicks int                `json:"cooldown_ticks"`
}

type BossPatternPhase struct {
	Kind               string       `json:"kind"`
	DurationTicks      int          `json:"duration_ticks"`
	TelegraphType      string       `json:"telegraph_type,omitempty"`
	FromColor          string       `json:"from_color,omitempty"`
	ToColor            string       `json:"to_color,omitempty"`
	HitShape           string       `json:"hit_shape,omitempty"`
	Shape              string       `json:"shape,omitempty"`
	Radius             float64      `json:"radius,omitempty"`
	Width              float64      `json:"width,omitempty"`
	SummonMonsterDefID string       `json:"summon_monster_def_id,omitempty"`
	SummonCount        int          `json:"summon_count,omitempty"`
	SummonRadius       float64      `json:"summon_radius,omitempty"`
	Damage             *DamageRange `json:"damage,omitempty"`
}

func roundPositive(value float64) int {
	return max(1, int(math.Floor(value+0.5)))
}

func scaleDamageRange(base DamageRange, multiplier float64) DamageRange {
	return DamageRange{
		Min: roundPositive(float64(base.Min) * multiplier),
		Max: roundPositive(float64(base.Max) * multiplier),
	}
}

// GridBounds is the inclusive grid rectangle searched by A*.
type GridBounds struct {
	MinX int `json:"min_x"`
	MinY int `json:"min_y"`
	MaxX int `json:"max_x"`
	MaxY int `json:"max_y"`
}

// ItemDef is a single item definition.
type ItemDef struct {
	Name            string       `json:"name"`
	Category        string       `json:"category"`
	ItemType        string       `json:"item_type,omitempty"`
	Slot            string       `json:"slot"`
	Equippable      bool         `json:"equippable"`
	ClassRequired   string       `json:"class_required,omitempty"`
	Handedness      string       `json:"handedness,omitempty"`
	OccupiesHands   []string     `json:"occupies_hands,omitempty"`
	AttackMode      string       `json:"attack_mode,omitempty"`
	DamageType      string       `json:"damage_type,omitempty"`
	AttackSpeed     float64      `json:"attack_speed,omitempty"`
	Damage          *DamageRange `json:"damage,omitempty"`
	Reach           *float64     `json:"reach,omitempty"`
	ProjectileSpeed *float64     `json:"projectile_speed,omitempty"`
	Heal            *DamageRange `json:"heal,omitempty"`
	ManaRestore     *DamageRange `json:"mana_restore,omitempty"`
	Gold            *DamageRange `json:"gold,omitempty"`
}

// ItemTemplateDef is a server-authoritative rolled item template.
type ItemTemplateDef struct {
	Name              string             `json:"name"`
	Category          string             `json:"category"`
	EquipmentCategory string             `json:"equipment_category"`
	ItemType          string             `json:"item_type"`
	Slot              string             `json:"slot"`
	Equippable        bool               `json:"equippable"`
	Handedness        string             `json:"handedness,omitempty"`
	OccupiesHands     []string           `json:"occupies_hands,omitempty"`
	AttackMode        string             `json:"attack_mode,omitempty"`
	AttackSpeed       float64            `json:"attack_speed,omitempty"`
	Reach             float64            `json:"reach"`
	ProjectileSpeed   float64            `json:"projectile_speed,omitempty"`
	Requirements      map[string]int     `json:"requirements"`
	BaseStats         map[string]int     `json:"base_stats"`
	RollableStats     []RollableStatDef  `json:"rollable_stats"`
	ClassAffinities   []ClassAffinityDef `json:"class_affinities,omitempty"`
	ClassRequired     string             `json:"class_required,omitempty"`
	EffectPool        []string           `json:"effect_pool"`
}

// SkillDef is a server-authoritative active skill definition.
type SkillDef struct {
	Name         string               `json:"name"`
	Class        string               `json:"class"`
	DamageType   string               `json:"damage_type,omitempty"`
	Tree         SkillTreeDef         `json:"tree"`
	Kind         string               `json:"kind"`
	MaxRank      int                  `json:"max_rank"`
	Targeting    string               `json:"targeting"`
	Requirements SkillRequirementDef  `json:"requirements"`
	Cost         SkillCostDef         `json:"cost"`
	Damage       SkillDamageDef       `json:"damage"`
	Projectile   SkillProjectileDef   `json:"projectile"`
	Pierce       SkillPierceDef       `json:"pierce"`
	Root         SkillRootDef         `json:"root"`
	Volley       SkillVolleyDef       `json:"volley"`
	Cone         SkillConeDef         `json:"cone"`
	Poison       SkillPoisonDef       `json:"poison"`
	Bleed        SkillBleedDef        `json:"bleed"`
	Mark         SkillMarkDef         `json:"mark"`
	Dash         SkillDashDef         `json:"dash"`
	Mobility     SkillMobilityDef     `json:"mobility"`
	Execute      SkillExecuteDef      `json:"execute"`
	PassiveStats SkillPassiveStatsDef `json:"passive_stats"`
	Slow         SkillSlowDef         `json:"slow"`
	Shatter      SkillShatterDef      `json:"shatter"`
	Chain        SkillChainDef        `json:"chain"`
	Resolution   string               `json:"resolution,omitempty"`
	Companion    SkillCompanionDef    `json:"companion"`
	Revive       SkillReviveDef       `json:"revive"`
	Synergies    []SkillSynergyDef    `json:"synergies,omitempty"`
	Effects      []SkillEffectDef     `json:"effects"`
	Cooldown     SkillCooldownDef     `json:"cooldown"`
}

type SkillTreeDef struct {
	Tier   int    `json:"tier"`
	Column int    `json:"column"`
	Branch string `json:"branch,omitempty"`
}

type SkillRequirementDef struct {
	Level        int                    `json:"level"`
	LevelPerRank int                    `json:"level_per_rank"`
	Stats        map[string]int         `json:"stats"`
	StatsPerRank map[string]int         `json:"stats_per_rank"`
	Skills       []SkillPrerequisiteDef `json:"skills"`
}

type SkillPrerequisiteDef struct {
	SkillID string `json:"skill_id"`
	Rank    int    `json:"rank"`
}

type SkillCostDef struct {
	Mana SkillRankValueDef `json:"mana"`
}

type SkillRankValueDef struct {
	Base    int `json:"base"`
	PerRank int `json:"per_rank"`
}

// SkillDamageDef controls the deterministic rank-scaled damage range.
type SkillDamageDef struct {
	Type         string          `json:"type"`
	MinBase      int             `json:"min_base"`
	MaxBase      int             `json:"max_base"`
	MinPerRank   int             `json:"min_per_rank"`
	MaxPerRank   int             `json:"max_per_rank"`
	MagicScaling SkillScalingDef `json:"magic_scaling,omitempty"`
}

// SkillScalingDef lets a skill opt into gentle caster-stat scaling.
type SkillScalingDef struct {
	Stat                   string  `json:"stat"`
	PercentPerPoint        float64 `json:"percent_per_point"`
	MaxBonusPercent        float64 `json:"max_bonus_percent"`
	UseRequirementBaseline bool    `json:"use_requirement_baseline"`
}

// SkillProjectileDef defines server-owned projectile behavior for a skill.
type SkillProjectileDef struct {
	Range  float64 `json:"range"`
	Speed  float64 `json:"speed"`
	Visual string  `json:"visual"`
}

// SkillPierceDef defines a deterministic projectile that can hit multiple monsters.
type SkillPierceDef struct {
	MaxHits                  int `json:"max_hits"`
	DamagePercentPerExtraHit int `json:"damage_percent_per_extra_hit"`
}

// SkillRootDef defines a movement root applied by a projectile skill.
type SkillRootDef struct {
	EffectID      string `json:"effect_id"`
	DurationTicks int    `json:"duration_ticks"`
}

// SkillVolleyDef defines a deterministic fan of projectile rays.
type SkillVolleyDef struct {
	ArrowCount    int     `json:"arrow_count"`
	SpreadDegrees float64 `json:"spread_degrees"`
}

// SkillConeDef defines a server-owned cone attack shape and push behavior.
type SkillConeDef struct {
	Range        float64 `json:"range"`
	AngleDegrees float64 `json:"angle_degrees"`
	PushMin      float64 `json:"push_min"`
	PushMax      float64 `json:"push_max"`
	DamageSource string  `json:"damage_source"`
}

// SkillSlowDef defines a stackable movement slow applied by cold skills.
type SkillSlowDef struct {
	EffectID      string `json:"effect_id"`
	Percent       int    `json:"percent"`
	DurationTicks int    `json:"duration_ticks"`
	MaxPercent    int    `json:"max_percent"`
}

// SkillShatterDef defines secondary projectile fan-out after a cold hit.
type SkillShatterDef struct {
	MinShards int     `json:"min_shards"`
	MaxShards int     `json:"max_shards"`
	Range     float64 `json:"range"`
	Speed     float64 `json:"speed"`
	Visual    string  `json:"visual"`
}

type SkillChainDef struct {
	RangeMultiplier float64 `json:"range_multiplier"`
	MaxJumps        int     `json:"max_jumps"`
	Visual          string  `json:"visual"`
}

// SkillExecuteDef defines a passive low-health execute roll.
type SkillExecuteDef struct {
	ThresholdPercentBase    int `json:"threshold_percent_base"`
	ThresholdPercentPerRank int `json:"threshold_percent_per_rank"`
	ChancePercent           int `json:"chance_percent"`
}

// SkillPassiveStatsDef defines always-on bonuses applied from learned ranks.
type SkillPassiveStatsDef struct {
	Stats           map[string]SkillRankValueDef   `json:"stats"`
	AffinityScaling SkillPassiveAffinityScalingDef `json:"affinity_scaling,omitempty"`
}

type SkillPassiveAffinityScalingDef struct {
	MaxActiveAffinities int                          `json:"max_active_affinities"`
	Stats               map[string]SkillRankValueDef `json:"stats"`
}

// SkillEffectDef is a closed data contract for supported active-skill effects.
type SkillEffectDef struct {
	Type                          string          `json:"type"`
	Stats                         []string        `json:"stats"`
	PercentBase                   int             `json:"percent_base"`
	PercentPerRank                int             `json:"percent_per_rank"`
	DurationTicks                 int             `json:"duration_ticks"`
	VisualScale                   bool            `json:"visual_scale"`
	Target                        string          `json:"target"`
	IncludeCaster                 bool            `json:"include_caster"`
	Range                         float64         `json:"range"`
	Radius                        float64         `json:"radius"`
	EffectID                      string          `json:"effect_id"`
	MagicScaling                  SkillScalingDef `json:"magic_scaling,omitempty"`
	ManaPerHPBase                 int             `json:"mana_per_hp_base,omitempty"`
	ManaPerHPPerRank              int             `json:"mana_per_hp_per_rank,omitempty"`
	OutgoingDamagePercent         int             `json:"outgoing_damage_percent,omitempty"`
	HealthRegenMultiplierPercent  int             `json:"health_regen_multiplier_percent,omitempty"`
	CleanseDebuffs                bool            `json:"cleanse_debuffs,omitempty"`
	MarkRadius                    float64         `json:"mark_radius,omitempty"`
	MarkDamageBonusPercent        int             `json:"mark_damage_bonus_percent,omitempty"`
	MarkDamageBonusPercentPerRank int             `json:"mark_damage_bonus_percent_per_rank,omitempty"`
	MarkDurationTicks             int             `json:"mark_duration_ticks,omitempty"`
	MarkEffectID                  string          `json:"mark_effect_id,omitempty"`
	RedirectDamage                bool            `json:"redirect_damage,omitempty"`
	PhaseThroughMonsters          bool            `json:"phase_through_monsters,omitempty"`
	DamageImmunity                bool            `json:"damage_immunity,omitempty"`
	EvadePercent                  int             `json:"evade_percent,omitempty"`
}

// SkillCooldownDef defines how a skill cooldown is derived.
type SkillCooldownDef struct {
	Type                        string  `json:"type"`
	Multiplier                  float64 `json:"multiplier"`
	FlatTicks                   int     `json:"flat_ticks,omitempty"`
	FixedTicks                  int     `json:"fixed_ticks,omitempty"`
	MagicReductionTicksPerPoint int     `json:"magic_reduction_ticks_per_point,omitempty"`
}

// InteractableDef is a single activatable world object definition.
type InteractableDef struct {
	Name              string               `json:"name"`
	InitialState      string               `json:"initial_state"`
	Transition        string               `json:"transition,omitempty"`
	ShopID            string               `json:"shop_id,omitempty"`
	StashID           string               `json:"stash_id,omitempty"`
	Service           string               `json:"service,omitempty"`
	BarrierWhenClosed *InteractableBarrier `json:"barrier_when_closed,omitempty"`
	LockedExitReason  string               `json:"locked_exit_reason,omitempty"`
}

// InteractableBarrier is the closed-state movement blocker for an interactable.
type InteractableBarrier struct {
	Size Vec2 `json:"size"`
}

// MonsterDef is a single monster definition.
type MonsterDef struct {
	Name                 string             `json:"name"`
	MaxHP                int                `json:"max_hp"`
	LootTable            string             `json:"loot_table"`
	HitChance            *float64           `json:"hit_chance,omitempty"`
	CritChance           *float64           `json:"crit_chance,omitempty"`
	CritDamage           *float64           `json:"crit_damage,omitempty"`
	Armor                int                `json:"armor,omitempty"`
	BlockPercent         int                `json:"block_percent,omitempty"`
	Resistances          map[string]float64 `json:"resistances,omitempty"`
	RetaliationDamage    *DamageRange       `json:"retaliation_damage,omitempty"`
	AttackDamage         *DamageRange       `json:"attack_damage,omitempty"`
	AttackCooldown       int                `json:"attack_cooldown_ticks,omitempty"`
	AttackWindupTicks    int                `json:"attack_windup_ticks,omitempty"`
	AttackMode           string             `json:"attack_mode,omitempty"`
	AttackStyle          string             `json:"attack_style,omitempty"`
	AttackRange          float64            `json:"attack_range,omitempty"`
	PreferredMinRange    float64            `json:"preferred_min_range,omitempty"`
	ProjectileSpeed      float64            `json:"projectile_speed,omitempty"`
	ProjectileDefID      string             `json:"projectile_def_id,omitempty"`
	Behavior             string             `json:"behavior,omitempty"`
	NavigationTrait      string             `json:"navigation_trait,omitempty"`
	PackRole             string             `json:"pack_role,omitempty"`
	AggroRadius          float64            `json:"aggro_radius,omitempty"`
	AssistRadius         float64            `json:"assist_radius,omitempty"`
	LeashRadius          float64            `json:"leash_radius,omitempty"`
	MoveSpeed            float64            `json:"move_speed,omitempty"`
	XPReward             int                `json:"xp_reward,omitempty"`
	TrainingTarget       bool               `json:"training_target,omitempty"`
	TrainingHPMultiplier int                `json:"training_hp_multiplier,omitempty"`
	ReviveDelayTicks     int                `json:"revive_delay_ticks,omitempty"`
}

func (d MonsterDef) effectiveTrainingHPMultiplier() int {
	if d.TrainingHPMultiplier < 1 {
		return 1
	}

	return d.TrainingHPMultiplier
}

func (d MonsterDef) effectiveAssistRadius() float64 {
	if d.AssistRadius > 0 {
		return d.AssistRadius
	}
	return d.AggroRadius
}

func (d MonsterDef) effectiveBehavior() string {
	if d.Behavior == "" {
		return monsterBehaviorStatic
	}

	return d.Behavior
}

func (d MonsterDef) effectiveAttackMode() string {
	if d.AttackMode == "" {
		return attackModeMelee
	}

	return d.AttackMode
}

func (d MonsterDef) effectiveMoveSpeed(nav NavigationRules) float64 {
	if d.MoveSpeed > 0 {
		return d.MoveSpeed
	}

	return nav.CellSize
}

func (d MonsterDef) effectiveHitChance(combat Combat) float64 {
	if d.HitChance != nil {
		return clampFloat(*d.HitChance, 0, 1)
	}
	return combat.BaseHitChance
}

func (d MonsterDef) effectiveCritChance(combat Combat) float64 {
	if d.CritChance != nil {
		return clampFloat(*d.CritChance, 0, 1)
	}
	return combat.BaseCritChance
}

func (d MonsterDef) effectiveCritDamage(combat Combat) float64 {
	if d.CritDamage != nil {
		return maxFloat(1, *d.CritDamage)
	}
	return combat.BaseCritDamage
}

// LootEntry is one weighted entry in a loot table.
type LootEntry struct {
	ItemDefID      string `json:"item_def_id"`
	ItemTemplateID string `json:"item_template_id"`
	UniqueItemID   string `json:"unique_item_id"`
	SetItemID      string `json:"set_item_id"`
	Weight         int    `json:"weight"`
}

type UniqueEffectDef struct {
	ID                  string                 `json:"id"`
	Enabled             bool                   `json:"enabled"`
	DisplayName         string                 `json:"display_name"`
	Summary             string                 `json:"summary"`
	Hook                string                 `json:"hook"`
	CompatibleItemTypes []string               `json:"compatible_item_types"`
	Params              map[string]interface{} `json:"params"`
	Status              string                 `json:"status"`
}

type UniqueItemDef struct {
	ID             string         `json:"id"`
	Enabled        bool           `json:"enabled"`
	BaseTemplateID string         `json:"base_template_id"`
	DisplayName    string         `json:"display_name"`
	MinimumLevel   int            `json:"minimum_level"`
	FixedStats     map[string]int `json:"fixed_stats"`
	FixedEffectIDs []string       `json:"fixed_effect_ids"`
	BehaviorHook   string         `json:"behavior_hook"`
	Status         string         `json:"status"`
}

type TreasureClassEntry struct {
	ItemDefID      string `json:"item_def_id"`
	ItemTemplateID string `json:"item_template_id"`
	UniqueItemID   string `json:"unique_item_id"`
	SetItemID      string `json:"set_item_id"`
	Weight         int    `json:"weight"`
}

type TreasureAttemptDef struct {
	AttemptID     string               `json:"attempt_id"`
	SuccessWeight int                  `json:"success_weight"`
	NoDropWeight  int                  `json:"no_drop_weight"`
	Entries       []TreasureClassEntry `json:"entries"`
}

type TreasureClassDef struct {
	Attempts []TreasureAttemptDef `json:"attempts"`
}

// LootTable is a weighted set of loot entries.
type LootTable struct {
	Drops           []string    `json:"drops,omitempty"`
	Entries         []LootEntry `json:"entries"`
	TreasureClassID string      `json:"treasure_class_id,omitempty"`
}

type ShopFixedOffer struct {
	OfferID   string `json:"offer_id"`
	ItemDefID string `json:"item_def_id"`
	BuyPrice  int    `json:"buy_price"`
}

type ShopGeneratedOffers struct {
	OfferCount        int    `json:"offer_count"`
	Source            string `json:"source"`
	MinDepth          int    `json:"min_depth"`
	SourceDepthPolicy string `json:"source_depth_policy"`
	MaxRarity         string `json:"max_rarity"`
	RefreshOn         string `json:"refresh_on"`
	MaxRollAttempts   int    `json:"max_roll_attempts"`
}

type ShopMysteryOffers struct {
	Enabled           bool     `json:"enabled"`
	EligibleSlots     []string `json:"eligible_slots"`
	Source            string   `json:"source"`
	SourceDepthWindow int      `json:"source_depth_window"`
	MinRarity         string   `json:"min_rarity"`
	MaxRarity         string   `json:"max_rarity"`
	RefreshOn         string   `json:"refresh_on"`
	PriceMultiplier   float64  `json:"price_multiplier"`
	RerollCost        int      `json:"reroll_cost"`
	MaxRollAttempts   int      `json:"max_roll_attempts"`
}

type ShopBuyback struct {
	Enabled            bool    `json:"enabled"`
	Scope              string  `json:"scope"`
	BuyPriceMultiplier float64 `json:"buy_price_multiplier"`
	ClearOnLeaveTown   bool    `json:"clear_on_leave_town"`
}

type ShopPricing struct {
	SellMultiplier    float64            `json:"sell_multiplier"`
	RoundBuyTo        int                `json:"round_buy_to"`
	RarityMultipliers map[string]float64 `json:"rarity_multipliers"`
	SlotBase          map[string]int     `json:"slot_base"`
	StatWeights       map[string]int     `json:"stat_weights"`
}

type ShopDef struct {
	Name            string              `json:"name"`
	FixedOffers     []ShopFixedOffer    `json:"fixed_offers"`
	GeneratedOffers ShopGeneratedOffers `json:"generated_offers"`
	MysteryOffers   ShopMysteryOffers   `json:"mystery_offers"`
	Buyback         ShopBuyback         `json:"buyback"`
	Pricing         ShopPricing         `json:"pricing"`
}

func shopRarityAllowedByCap(rarity, maxRarity string) bool {
	r, ok := shopRarityRank(rarity)
	if !ok {
		return false
	}
	maxRank, ok := shopRarityRank(maxRarity)
	if !ok {
		return false
	}
	return r <= maxRank
}

func shopRarityRank(rarity string) (int, bool) {
	rank := map[string]int{"common": 0, "magic": 1, "rare": 2, "unique": 3, "set": 3}
	out, ok := rank[rarity]
	return out, ok
}

// WorldDef is a deterministic initial session layout.
type WorldDef struct {
	Mode       string        `json:"mode,omitempty"`
	StartLevel *int          `json:"start_level,omitempty"`
	Player     WorldPlayer   `json:"player"`
	Entities   []WorldEntity `json:"entities"`
}

// WorldPlayer is the initial player placement for a world.
type WorldPlayer struct {
	Position Vec2 `json:"position"`
}

// WorldEntity is an initial non-player entity in a world.
type WorldEntity struct {
	Type              string           `json:"type"`
	MonsterDefID      string           `json:"monster_def_id,omitempty"`
	ItemDefID         string           `json:"item_def_id,omitempty"`
	ItemTemplateID    string           `json:"item_template_id,omitempty"`
	LootPreset        *WorldLootPreset `json:"loot_preset,omitempty"`
	InteractableDefID string           `json:"interactable_def_id,omitempty"`
	Kind              string           `json:"kind,omitempty"`
	BlocksLineOfSight *bool            `json:"blocks_line_of_sight,omitempty"`
	Position          Vec2             `json:"position"`
	Size              Vec2             `json:"size,omitempty"`
}

// WorldLootPreset pins rolled stats for lab loot entities.
type WorldLootPreset struct {
	ItemTemplateID    string                `json:"item_template_id"`
	Rarity            string                `json:"rarity,omitempty"`
	ItemLevel         int                   `json:"item_level,omitempty"`
	DisplayName       string                `json:"display_name,omitempty"`
	Stats             map[string]int        `json:"stats"`
	SkillLevelBonuses []SkillLevelBonusRoll `json:"skill_level_bonuses,omitempty"`
}

// LoadRules reads and parses the v0 rules files from a directory.
func LoadRules(dir string) (*Rules, error) {
	r := &Rules{}

	var mainConfig MainConfig
	if err := readJSON(filepath.Join(dir, "main_config.v0.json"), &mainConfig); err != nil {
		return nil, err
	}
	if mainConfig.Gameplay.BaseAttackIntervalTicks <= 0 {
		return nil, fmt.Errorf("game: invalid rules main_config.gameplay.base_attack_interval_ticks: must be positive")
	}
	if mainConfig.Gameplay.BaseMovementSpeed <= 0 {
		return nil, fmt.Errorf("game: invalid rules main_config.gameplay.base_movement_speed: must be positive")
	}
	if mainConfig.Gameplay.MovementAccelerationSeconds <= 0 {
		mainConfig.Gameplay.MovementAccelerationSeconds = 2.0
	}
	if mainConfig.Gameplay.MovementMinSpeedFactor <= 0 || mainConfig.Gameplay.MovementMinSpeedFactor > 1 {
		mainConfig.Gameplay.MovementMinSpeedFactor = 0.2
	}
	if mainConfig.Gameplay.MinimumMonsterAggroRadius < 0 {
		return nil, fmt.Errorf("game: invalid rules main_config.gameplay.minimum_monster_aggro_radius: must be non-negative")
	}
	if mainConfig.Gameplay.RangedRetreatMinMeleeEngagementSeconds < 0 {
		return nil, fmt.Errorf("game: invalid rules main_config.gameplay.ranged_retreat_min_melee_engagement_seconds: must be non-negative")
	}
	if mainConfig.Gameplay.BaseDropRatePercent < 0 || mainConfig.Gameplay.BaseDropRatePercent > 100 {
		return nil, fmt.Errorf("game: invalid rules main_config.gameplay.base_drop_rate_percent: must be within [0,100]")
	}
	if mainConfig.Gameplay.BossBonusDropRatePercent < 0 || mainConfig.Gameplay.BossBonusDropRatePercent > 100 {
		return nil, fmt.Errorf("game: invalid rules main_config.gameplay.boss_bonus_drop_rate_percent: must be within [0,100]")
	}
	if mainConfig.Gameplay.BossExtraDropRatePercent < 0 || mainConfig.Gameplay.BossExtraDropRatePercent > 100 {
		return nil, fmt.Errorf("game: invalid rules main_config.gameplay.boss_extra_drop_rate_percent: must be within [0,100]")
	}
	if mainConfig.Gameplay.BossLootMagicFindBonusPercent < 0 || mainConfig.Gameplay.BossLootMagicFindBonusPercent > 500 {
		return nil, fmt.Errorf("game: invalid rules main_config.gameplay.boss_loot_magic_find_bonus_percent: must be within [0,500]")
	}
	if mainConfig.Gameplay.RespecCostGold < 0 {
		return nil, fmt.Errorf("game: invalid rules main_config.gameplay.respec_cost_gold: must be non-negative")
	}
	if mainConfig.Gameplay.ItemUpgradeCostGold < 0 {
		return nil, fmt.Errorf("game: invalid rules main_config.gameplay.item_upgrade_cost_gold: must be non-negative")
	}
	if mainConfig.Gameplay.ItemUpgradeCostGrowth < 0 {
		return nil, fmt.Errorf("game: invalid rules main_config.gameplay.item_upgrade_cost_growth_per_level: must be non-negative")
	}
	if err := validateItemUpgradeChanceConfig(mainConfig.Gameplay.ItemUpgradeMaxLevel, mainConfig.Gameplay.ItemUpgradeFailureCurve, mainConfig.Gameplay.ItemUpgradeShardSuccessBonusPercentPerTier); err != nil {
		return nil, err
	}
	if mainConfig.Gameplay.ItemUpgradeSuccessPct < 0 || mainConfig.Gameplay.ItemUpgradeSuccessPct > 100 {
		return nil, fmt.Errorf("game: invalid rules main_config.gameplay.item_upgrade_success_chance_percent: must be 0-100")
	}
	if mainConfig.Gameplay.ItemUpgradePityFailures < 0 {
		return nil, fmt.Errorf("game: invalid rules main_config.gameplay.item_upgrade_pity_failure_threshold: must be non-negative")
	}
	if mainConfig.Gameplay.CompanionAssistRadius <= 0 {
		return nil, fmt.Errorf("game: invalid rules main_config.gameplay.companion_assist_radius: must be positive")
	}
	if mainConfig.Gameplay.CompanionFollowDistance <= 0 {
		return nil, fmt.Errorf("game: invalid rules main_config.gameplay.companion_follow_distance: must be positive")
	}
	if mainConfig.Gameplay.CompanionFollowStop <= 0 {
		return nil, fmt.Errorf("game: invalid rules main_config.gameplay.companion_follow_stop_radius: must be positive")
	}
	if err := validateMainGameplayEconomyConfig(mainConfig.Gameplay); err != nil {
		return nil, err
	}
	if err := validateWeaponElementalProcsConfig(mainConfig.WeaponElementalProcs); err != nil {
		return nil, err
	}
	if mainConfig.Gameplay.ItemUpgradeResourceCost > 0 {
		if mainConfig.Gameplay.ItemUpgradeResourceID == "" {
			return nil, fmt.Errorf("game: invalid rules main_config.gameplay.item_upgrade_resource_item_def_id: required when count is positive")
		}
	}
	r.MainConfig = mainConfig

	var combat struct {
		Version                           int                               `json:"version"`
		BaseHitChance                     float64                           `json:"base_hit_chance"`
		BaseCritChance                    float64                           `json:"base_crit_chance"`
		BaseCritDamage                    float64                           `json:"base_crit_damage"`
		MinimumDamage                     int                               `json:"minimum_damage"`
		BlockCap                          int                               `json:"block_cap_percent"`
		BaseAttackIntervalTicks           int                               `json:"base_attack_interval_ticks"`
		MinEffectiveAttackSpeed           float64                           `json:"min_effective_attack_speed"`
		MaxEffectiveAttackSpeed           float64                           `json:"max_effective_attack_speed"`
		PlayerDamage                      DamageRange                       `json:"player_damage"`
		TwoHandedStrengthDamageMultiplier float64                           `json:"two_handed_strength_damage_multiplier"`
		WeaponDamageScaling               map[string]WeaponDamageScalingDef `json:"weapon_damage_scaling"`
		UnarmedReach                      float64                           `json:"unarmed_reach"`
		Coop                              CoopCombat                        `json:"coop"`
	}
	if err := readJSON(filepath.Join(dir, "combat.v0.json"), &combat); err != nil {
		return nil, err
	}
	if err := validateDamageRange("combat.player_damage", combat.PlayerDamage); err != nil {
		return nil, err
	}
	if combat.BaseHitChance < 0 || combat.BaseHitChance > 1 {
		return nil, fmt.Errorf("game: invalid rules combat.base_hit_chance: must be within [0,1]")
	}
	if combat.BaseCritChance < 0 || combat.BaseCritChance > 1 {
		return nil, fmt.Errorf("game: invalid rules combat.base_crit_chance: must be within [0,1]")
	}
	if combat.BaseCritDamage < 1 {
		return nil, fmt.Errorf("game: invalid rules combat.base_crit_damage: must be >= 1")
	}
	if combat.MinimumDamage < 1 {
		return nil, fmt.Errorf("game: invalid rules combat.minimum_damage: must be >= 1")
	}
	if combat.BlockCap < 0 || combat.BlockCap > 75 {
		return nil, fmt.Errorf("game: invalid rules combat.block_cap_percent: must be within [0,75]")
	}
	if combat.BaseAttackIntervalTicks <= 0 {
		return nil, fmt.Errorf("game: invalid rules combat.base_attack_interval_ticks: must be positive")
	}
	if combat.MinEffectiveAttackSpeed <= 0 {
		return nil, fmt.Errorf("game: invalid rules combat.min_effective_attack_speed: must be positive")
	}
	if combat.MaxEffectiveAttackSpeed < combat.MinEffectiveAttackSpeed {
		return nil, fmt.Errorf("game: invalid rules combat.max_effective_attack_speed: must be >= min_effective_attack_speed")
	}
	if combat.UnarmedReach <= 0 {
		return nil, fmt.Errorf("game: invalid rules combat.unarmed_reach: must be positive")
	}
	if combat.TwoHandedStrengthDamageMultiplier < 1 {
		return nil, fmt.Errorf("game: invalid rules combat.two_handed_strength_damage_multiplier: must be >= 1")
	}
	if err := validateWeaponDamageScaling(combat.WeaponDamageScaling); err != nil {
		return nil, err
	}
	if err := validateCoopCombatRules(combat.Coop); err != nil {
		return nil, err
	}
	r.Combat = Combat{
		BaseHitChance:                     combat.BaseHitChance,
		BaseCritChance:                    combat.BaseCritChance,
		BaseCritDamage:                    combat.BaseCritDamage,
		MinimumDamage:                     combat.MinimumDamage,
		BlockCap:                          combat.BlockCap,
		BaseAttackIntervalTicks:           mainConfig.Gameplay.BaseAttackIntervalTicks,
		MinEffectiveAttackSpeed:           combat.MinEffectiveAttackSpeed,
		MaxEffectiveAttackSpeed:           combat.MaxEffectiveAttackSpeed,
		PlayerDamage:                      combat.PlayerDamage,
		TwoHandedStrengthDamageMultiplier: combat.TwoHandedStrengthDamageMultiplier,
		WeaponDamageScaling:               combat.WeaponDamageScaling,
		UnarmedReach:                      combat.UnarmedReach,
		Coop:                              combat.Coop,
	}

	navigation, err := loadNavigationRules(dir)
	if err != nil {
		return nil, err
	}
	r.Navigation = navigation

	var progression struct {
		Version          int                          `json:"version"`
		BaseStats        BaseStatsView                `json:"base_stats"`
		Classes          map[string]CharacterClassDef `json:"classes"`
		PointsPerLevel   int                          `json:"points_per_level"`
		SkillPoints      SkillPointRules              `json:"skill_points"`
		SkillRankScaling RankScalingCurve             `json:"skill_rank_scaling"`
		SkillManaScaling RankScalingCurve             `json:"skill_mana_scaling"`
		LevelCap         int                          `json:"level_cap"`
		ExperienceCurve  struct {
			Type   string `json:"type"`
			Levels []struct {
				Level            int `json:"level"`
				NextLevelTotalXP int `json:"next_level_total_xp"`
			} `json:"levels"`
		} `json:"experience_curve"`
		DerivedStats map[string]LinearStatFormula `json:"derived_stats"`
		LevelUp      struct {
			RestoreHPAndMana bool `json:"restore_hp_and_mana"`
		} `json:"level_up"`
	}
	if err := readJSON(filepath.Join(dir, "character_progression.v0.json"), &progression); err != nil {
		return nil, err
	}
	if progression.BaseStats.Str <= 0 || progression.BaseStats.Dex <= 0 || progression.BaseStats.Vit <= 0 || progression.BaseStats.Magic <= 0 {
		return nil, fmt.Errorf("game: invalid rules character_progression.base_stats: all stats must be positive")
	}
	if len(progression.Classes) == 0 {
		return nil, fmt.Errorf("game: invalid rules character_progression.classes: at least one class is required")
	}
	for id, classDef := range progression.Classes {
		if id == "" {
			return nil, fmt.Errorf("game: invalid rules character_progression.classes: empty class id")
		}
		if classDef.Name == "" {
			return nil, fmt.Errorf("game: invalid rules character_progression.classes.%s.name: required", id)
		}
		if classDef.LightRadius <= 0 {
			return nil, fmt.Errorf("game: invalid rules character_progression.classes.%s.light_radius: must be positive", id)
		}
		if classDef.BaseStats.Str <= 0 || classDef.BaseStats.Dex <= 0 || classDef.BaseStats.Vit <= 0 || classDef.BaseStats.Magic <= 0 {
			return nil, fmt.Errorf("game: invalid rules character_progression.classes.%s.base_stats: all stats must be positive", id)
		}
	}
	if progression.PointsPerLevel <= 0 {
		return nil, fmt.Errorf("game: invalid rules character_progression.points_per_level: must be positive")
	}
	if progression.SkillPoints.PointsPerGrant <= 0 {
		return nil, fmt.Errorf("game: invalid rules character_progression.skill_points.points_per_grant: must be positive")
	}
	if progression.SkillPoints.GrantEveryLevels <= 0 {
		return nil, fmt.Errorf("game: invalid rules character_progression.skill_points.grant_every_levels: must be positive")
	}
	if progression.SkillPoints.FirstGrantLevel < 1 {
		return nil, fmt.Errorf("game: invalid rules character_progression.skill_points.first_grant_level: must be >= 1")
	}
	if progression.LevelCap < 2 {
		return nil, fmt.Errorf("game: invalid rules character_progression.level_cap: must be >= 2")
	}
	if progression.ExperienceCurve.Type != "table" {
		return nil, fmt.Errorf("game: invalid rules character_progression.experience_curve.type: %s", progression.ExperienceCurve.Type)
	}
	thresholds := make(map[int]int, len(progression.ExperienceCurve.Levels))
	prevXP := 0
	for _, row := range progression.ExperienceCurve.Levels {
		if row.Level < 1 || row.Level >= progression.LevelCap {
			return nil, fmt.Errorf("game: invalid rules character_progression.experience_curve.level: %d", row.Level)
		}
		if row.NextLevelTotalXP <= prevXP {
			return nil, fmt.Errorf("game: invalid rules character_progression.experience_curve.%d: xp must increase", row.Level)
		}
		thresholds[row.Level] = row.NextLevelTotalXP
		prevXP = row.NextLevelTotalXP
	}
	for level := 1; level < progression.LevelCap; level++ {
		if _, ok := thresholds[level]; !ok {
			return nil, fmt.Errorf("game: invalid rules character_progression.experience_curve: missing level %d", level)
		}
	}
	requiredDerived := []string{"damage_min", "damage_max", "armor", "attack_speed", "hit_chance", "crit_chance", "crit_damage", "movement_speed", "max_hp", "max_mana", "health_regen_per_second", "mana_regen_per_second", "light_radius"}
	for _, key := range requiredDerived {
		formula, ok := progression.DerivedStats[key]
		if !ok {
			return nil, fmt.Errorf("game: invalid rules character_progression.derived_stats: missing %s", key)
		}
		if formula.Type != "linear" && formula.Type != "logarithmic" {
			return nil, fmt.Errorf("game: invalid rules character_progression.derived_stats.%s.type: %s", key, formula.Type)
		}
		if formula.Type == "logarithmic" {
			switch formula.Stat {
			case "str", "dex", "vit", "magic":
			default:
				return nil, fmt.Errorf("game: invalid rules character_progression.derived_stats.%s.stat: %s", key, formula.Stat)
			}
			if formula.Denominator <= 0 {
				return nil, fmt.Errorf("game: invalid rules character_progression.derived_stats.%s.denominator: must be positive", key)
			}
		}
		if formula.Min != nil && formula.Max != nil && *formula.Max < *formula.Min {
			return nil, fmt.Errorf("game: invalid rules character_progression.derived_stats.%s: max must be >= min", key)
		}
	}
	r.CharacterProgression = CharacterProgressionRules{
		BaseStats:            progression.BaseStats,
		Classes:              progression.Classes,
		PointsPerLevel:       progression.PointsPerLevel,
		SkillPoints:          progression.SkillPoints,
		SkillRankScaling:     progression.SkillRankScaling,
		SkillManaScaling:     progression.SkillManaScaling,
		LevelCap:             progression.LevelCap,
		XPThresholds:         thresholds,
		DerivedStats:         progression.DerivedStats,
		LevelUpRestoreHPMana: progression.LevelUp.RestoreHPAndMana,
	}

	var items struct {
		Items map[string]ItemDef `json:"items"`
	}
	if err := readJSON(filepath.Join(dir, "items.v0.json"), &items); err != nil {
		return nil, err
	}
	for id, def := range items.Items {
		if def.ClassRequired != "" {
			if !def.Equippable {
				return nil, fmt.Errorf("game: invalid rules items.%s.class_required: only valid on equippable items", id)
			}
			if _, ok := r.CharacterProgression.Classes[def.ClassRequired]; !ok {
				return nil, fmt.Errorf("game: invalid rules items.%s.class_required: unknown class %s", id, def.ClassRequired)
			}
		}
		if def.Equippable && def.Slot == "" {
			return nil, fmt.Errorf("game: invalid rules items.%s: equippable item must declare slot", id)
		}
		if !def.Equippable && def.Slot != "" {
			return nil, fmt.Errorf("game: invalid rules items.%s: non-equippable item must not declare slot", id)
		}
		if def.Damage != nil {
			if !def.Equippable || !isHandSlot(def.Slot) {
				return nil, fmt.Errorf("game: invalid rules items.%s.damage: damage is only valid on equippable hand items", id)
			}
			if err := validateDamageRange("items."+id+".damage", *def.Damage); err != nil {
				return nil, err
			}
			if err := validateDamageType("items."+id+".damage_type", def.DamageType); err != nil {
				return nil, err
			}
			if def.AttackSpeed <= 0 {
				return nil, fmt.Errorf("game: invalid rules items.%s.attack_speed: weapon attack speed must be positive", id)
			}
		} else if def.AttackSpeed != 0 {
			return nil, fmt.Errorf("game: invalid rules items.%s.attack_speed: only valid on weapons", id)
		}
		if def.Reach != nil {
			if !def.Equippable || !isHandSlot(def.Slot) {
				return nil, fmt.Errorf("game: invalid rules items.%s.reach: reach is only valid on equippable hand items", id)
			}
			if *def.Reach <= 0 {
				return nil, fmt.Errorf("game: invalid rules items.%s.reach: must be positive", id)
			}
		}
		if def.Heal != nil {
			if def.Category != "consumable" || def.Equippable {
				return nil, fmt.Errorf("game: invalid rules items.%s.heal: heal is only valid on non-equippable consumables", id)
			}
			if err := validateDamageRange("items."+id+".heal", *def.Heal); err != nil {
				return nil, err
			}
		}
		if def.ManaRestore != nil {
			if def.Category != "consumable" || def.Equippable {
				return nil, fmt.Errorf("game: invalid rules items.%s.mana_restore: mana_restore is only valid on non-equippable consumables", id)
			}
			if err := validateDamageRange("items."+id+".mana_restore", *def.ManaRestore); err != nil {
				return nil, err
			}
		}
		if def.Gold != nil {
			if def.Category != "currency" || def.Equippable {
				return nil, fmt.Errorf("game: invalid rules items.%s.gold: gold is only valid on non-equippable currency", id)
			}
			if err := validateDamageRange("items."+id+".gold", *def.Gold); err != nil {
				return nil, err
			}
		}
		mode := def.AttackMode
		if mode == "" {
			mode = attackModeMelee
		}
		switch mode {
		case attackModeMelee:
			if def.ProjectileSpeed != nil {
				return nil, fmt.Errorf("game: invalid rules items.%s.projectile_speed: only valid on ranged weapons", id)
			}
		case attackModeRanged:
			if !def.Equippable || !isHandSlot(def.Slot) || def.Damage == nil || def.Reach == nil || def.ProjectileSpeed == nil {
				return nil, fmt.Errorf("game: invalid rules items.%s: ranged weapons require slot, damage, reach, and projectile_speed", id)
			}
			if *def.ProjectileSpeed <= 0 {
				return nil, fmt.Errorf("game: invalid rules items.%s.projectile_speed: must be positive", id)
			}
		default:
			return nil, fmt.Errorf("game: invalid rules items.%s.attack_mode: %s", id, def.AttackMode)
		}
	}
	if err := validateMainGameplayResourceItems(mainConfig.Gameplay, items.Items); err != nil {
		return nil, err
	}
	r.Items = items.Items

	var itemTemplates struct {
		Rarities  map[string]RarityDef       `json:"rarities"`
		Templates map[string]ItemTemplateDef `json:"templates"`
	}
	if err := readJSON(filepath.Join(dir, "item_templates.v0.json"), &itemTemplates); err != nil {
		return nil, err
	}
	rarityOrder := sortedStringKeys(itemTemplates.Rarities)
	for id, rarity := range itemTemplates.Rarities {
		if rarity.Weight <= 0 {
			return nil, fmt.Errorf("game: invalid rules item_templates.rarities.%s.weight: must be positive", id)
		}
		if rarity.StatRollsMin == 0 && rarity.StatRollsMax == 0 && rarity.StatRolls > 0 {
			rarity.StatRollsMin = rarity.StatRolls
			rarity.StatRollsMax = rarity.StatRolls
		}
		if rarity.StatRollsMin <= 0 {
			return nil, fmt.Errorf("game: invalid rules item_templates.rarities.%s.stat_rolls_min: must be positive", id)
		}
		if rarity.StatRollsMax < rarity.StatRollsMin {
			return nil, fmt.Errorf("game: invalid rules item_templates.rarities.%s.stat_rolls_max: must be >= stat_rolls_min", id)
		}
		if rarity.NamePrefix == "" {
			return nil, fmt.Errorf("game: invalid rules item_templates.rarities.%s.name_prefix: required", id)
		}
		itemTemplates.Rarities[id] = rarity
	}
	for id, def := range itemTemplates.Templates {
		if !def.Equippable || def.Category != "equipment" {
			return nil, fmt.Errorf("game: invalid rules item_templates.%s: templates must be equippable equipment", id)
		}
		switch def.EquipmentCategory {
		case "weapon_1h", "weapon_2h", "off_hand", "gear", "jewelry", "class_specialist":
		default:
			return nil, fmt.Errorf("game: invalid rules item_templates.%s.equipment_category: required", id)
		}
		if def.EquipmentCategory == "class_specialist" {
			if def.ClassRequired == "" {
				return nil, fmt.Errorf("game: invalid rules item_templates.%s.class_required: required for class_specialist", id)
			}
			if _, ok := r.CharacterProgression.Classes[def.ClassRequired]; !ok {
				return nil, fmt.Errorf("game: invalid rules item_templates.%s.class_required: unknown class %s", id, def.ClassRequired)
			}
			if def.BaseStats["armor"] != 0 {
				return nil, fmt.Errorf("game: invalid rules item_templates.%s.base_stats.armor: class_specialist must not grant armor", id)
			}
			for _, roll := range def.RollableStats {
				if roll.Stat == "armor" {
					return nil, fmt.Errorf("game: invalid rules item_templates.%s.rollable_stats.armor: class_specialist must not roll armor", id)
				}
			}
		} else if def.ClassRequired != "" {
			return nil, fmt.Errorf("game: invalid rules item_templates.%s.class_required: only valid on class_specialist templates", id)
		}
		if !isEquipmentSlot(def.Slot) && def.Slot != "ring" {
			return nil, fmt.Errorf("game: invalid rules item_templates.%s.slot: unsupported slot %s", id, def.Slot)
		}
		isWeaponTemplate := isHandSlot(def.Slot) && def.ItemType != "shield" && def.ItemType != "book"
		if isWeaponTemplate {
			if def.AttackMode == "" {
				def.AttackMode = attackModeMelee
			}
			if def.AttackMode != attackModeMelee && def.AttackMode != attackModeRanged {
				return nil, fmt.Errorf("game: invalid rules item_templates.%s.attack_mode: %s", id, def.AttackMode)
			}
			if def.Reach <= 0 {
				return nil, fmt.Errorf("game: invalid rules item_templates.%s.reach: must be positive", id)
			}
			if def.AttackSpeed <= 0 {
				return nil, fmt.Errorf("game: invalid rules item_templates.%s.attack_speed: must be positive", id)
			}
			if def.AttackMode == attackModeRanged && def.ProjectileSpeed <= 0 {
				return nil, fmt.Errorf("game: invalid rules item_templates.%s.projectile_speed: must be positive", id)
			}
			if def.Handedness != "one_handed" && def.Handedness != "two_handed" {
				return nil, fmt.Errorf("game: invalid rules item_templates.%s.handedness: required for hand item", id)
			}
			if def.Handedness == "two_handed" && !occupiesExactly(def.OccupiesHands, "main_hand", "off_hand") {
				return nil, fmt.Errorf("game: invalid rules item_templates.%s.occupies_hands: two-handed items must occupy both hands", id)
			}
			if def.Handedness == "one_handed" && len(def.OccupiesHands) == 0 {
				return nil, fmt.Errorf("game: invalid rules item_templates.%s.occupies_hands: one-handed items must occupy a hand", id)
			}
		} else if def.AttackMode != "" || def.Reach != 0 || def.ProjectileSpeed != 0 || def.AttackSpeed != 0 {
			return nil, fmt.Errorf("game: invalid rules item_templates.%s: non-weapon equipment must not declare attack fields", id)
		}
		for stat, required := range def.Requirements {
			if !isSupportedRequirementStat(stat) {
				return nil, fmt.Errorf("game: invalid rules item_templates.%s.requirements.%s: unsupported requirement", id, stat)
			}
			if required < 1 {
				return nil, fmt.Errorf("game: invalid rules item_templates.%s.requirements.%s: must be >= 1", id, stat)
			}
		}
		min, max := def.BaseStats["damage_min"], def.BaseStats["damage_max"]
		if _, ok := def.BaseStats["damage_min"]; ok && max < min {
			return nil, fmt.Errorf("game: invalid rules item_templates.%s.base_stats: damage_max must be >= damage_min", id)
		}
		seen := map[string]bool{}
		for stat := range def.BaseStats {
			if !isSupportedItemStat(stat) {
				return nil, fmt.Errorf("game: invalid rules item_templates.%s.base_stats: unsupported stat %s", id, stat)
			}
		}
		for _, roll := range def.RollableStats {
			if !isSupportedItemStat(roll.Stat) {
				return nil, fmt.Errorf("game: invalid rules item_templates.%s.rollable_stats: unsupported stat %s", id, roll.Stat)
			}
			if roll.MinRarity != "" {
				if _, ok := itemTemplates.Rarities[roll.MinRarity]; !ok {
					return nil, fmt.Errorf("game: invalid rules item_templates.%s.rollable_stats.%s.min_rarity: unknown rarity %s", id, roll.Stat, roll.MinRarity)
				}
			}
			if roll.Max < roll.Min {
				return nil, fmt.Errorf("game: invalid rules item_templates.%s.rollable_stats.%s: max must be >= min", id, roll.Stat)
			}
			if roll.Weight <= 0 {
				return nil, fmt.Errorf("game: invalid rules item_templates.%s.rollable_stats.%s: weight must be positive", id, roll.Stat)
			}
			if err := validateSkillLevelRollStat(id, roll, def); err != nil {
				return nil, err
			}
			seen[roll.Stat] = true
		}
		if !seen["damage_min"] || !seen["damage_max"] {
			if isWeaponTemplate {
				return nil, fmt.Errorf("game: invalid rules item_templates.%s.rollable_stats: damage_min and damage_max are required", id)
			}
		}
		if err := validateClassAffinityDefs(id, def.ClassAffinities, r.CharacterProgression.Classes); err != nil {
			return nil, err
		}
		itemTemplates.Templates[id] = def
	}
	r.ItemTemplates = itemTemplates.Templates
	r.Rarities = itemTemplates.Rarities
	r.RarityOrder = rarityOrder

	var uniqueEffects struct {
		Effects map[string]UniqueEffectDef `json:"effects"`
	}
	if err := readJSON(filepath.Join(dir, "unique_effects.v0.json"), &uniqueEffects); err != nil {
		return nil, err
	}
	for effectID, effect := range uniqueEffects.Effects {
		if effect.ID != effectID {
			return nil, fmt.Errorf("game: invalid rules unique_effects.%s.id: must match key", effectID)
		}
		if !effect.Enabled || effect.Status != "ready" {
			return nil, fmt.Errorf("game: invalid rules unique_effects.%s: must be enabled and ready", effectID)
		}
		if len(effect.CompatibleItemTypes) == 0 {
			return nil, fmt.Errorf("game: invalid rules unique_effects.%s.compatible_item_types: required", effectID)
		}
	}
	r.UniqueEffects = uniqueEffects.Effects

	var statusEffects struct {
		Statuses map[string]StatusEffectDef `json:"statuses"`
	}
	if err := readJSON(filepath.Join(dir, "status_effects.v0.json"), &statusEffects); err != nil {
		return nil, err
	}
	if err := validateStatusEffects(statusEffects.Statuses); err != nil {
		return nil, err
	}
	r.StatusEffects = statusEffects.Statuses

	var uniqueItems struct {
		Uniques map[string]UniqueItemDef `json:"uniques"`
	}
	if err := readJSON(filepath.Join(dir, "unique_items.v0.json"), &uniqueItems); err != nil {
		return nil, err
	}
	if err := r.validateUniqueItemRules(uniqueItems.Uniques); err != nil {
		return nil, err
	}
	r.UniqueItems = uniqueItems.Uniques

	var setItems struct {
		Sets map[string]SetItemCatalogDef `json:"sets"`
	}
	if err := readJSON(filepath.Join(dir, "set_items.v0.json"), &setItems); err != nil {
		return nil, err
	}
	if err := r.validateSetItemRules(setItems.Sets); err != nil {
		return nil, err
	}
	r.SetCatalogs = setItems.Sets

	skills, err := loadSkillRulesFromManifest(dir)
	if err != nil {
		return nil, err
	}
	if err := validateSkillRules(skills, nil, r.Combat.BaseAttackIntervalTicks); err != nil {
		return nil, err
	}
	r.Skills = skills

	var treasureClasses struct {
		Classes map[string]TreasureClassDef `json:"classes"`
	}
	if err := readJSON(filepath.Join(dir, "treasure_classes.v0.json"), &treasureClasses); err != nil {
		return nil, err
	}
	for classID, classDef := range treasureClasses.Classes {
		if len(classDef.Attempts) == 0 {
			return nil, fmt.Errorf("game: invalid rules treasure_classes.%s.attempts: required", classID)
		}
		seenAttempts := map[string]bool{}
		for _, attempt := range classDef.Attempts {
			if attempt.AttemptID == "" {
				return nil, fmt.Errorf("game: invalid rules treasure_classes.%s.attempts: attempt_id required", classID)
			}
			if seenAttempts[attempt.AttemptID] {
				return nil, fmt.Errorf("game: invalid rules treasure_classes.%s.attempts.%s: duplicate attempt_id", classID, attempt.AttemptID)
			}
			seenAttempts[attempt.AttemptID] = true
			if attempt.SuccessWeight < 0 || attempt.NoDropWeight < 0 || attempt.SuccessWeight+attempt.NoDropWeight <= 0 {
				return nil, fmt.Errorf("game: invalid rules treasure_classes.%s.attempts.%s: success/no_drop total must be positive", classID, attempt.AttemptID)
			}
			if attempt.SuccessWeight > 0 && len(attempt.Entries) == 0 {
				return nil, fmt.Errorf("game: invalid rules treasure_classes.%s.attempts.%s: success requires entries", classID, attempt.AttemptID)
			}
			for _, entry := range attempt.Entries {
				if countDropEntryRefs(entry.ItemDefID, entry.ItemTemplateID, entry.UniqueItemID, entry.SetItemID) != 1 {
					return nil, fmt.Errorf("game: invalid rules treasure_classes.%s.attempts.%s: entry must declare exactly one drop reference", classID, attempt.AttemptID)
				}
				if entry.Weight <= 0 {
					return nil, fmt.Errorf("game: invalid rules treasure_classes.%s.attempts.%s: entry weight must be positive", classID, attempt.AttemptID)
				}
				if entry.ItemDefID != "" {
					if _, ok := r.Items[entry.ItemDefID]; !ok {
						return nil, fmt.Errorf("game: invalid rules treasure_classes.%s.attempts.%s: unknown item %s", classID, attempt.AttemptID, entry.ItemDefID)
					}
				}
				if entry.ItemTemplateID != "" {
					if _, ok := r.ItemTemplates[entry.ItemTemplateID]; !ok {
						return nil, fmt.Errorf("game: invalid rules treasure_classes.%s.attempts.%s: unknown item template %s", classID, attempt.AttemptID, entry.ItemTemplateID)
					}
				}
				if entry.UniqueItemID != "" {
					unique, ok := r.UniqueItems[entry.UniqueItemID]
					if !ok || !unique.Enabled || unique.Status != "ready" {
						return nil, fmt.Errorf("game: invalid rules treasure_classes.%s.attempts.%s: unknown or inactive unique item %s", classID, attempt.AttemptID, entry.UniqueItemID)
					}
				}
				if entry.SetItemID != "" {
					if _, ok := r.SetItems[entry.SetItemID]; !ok {
						return nil, fmt.Errorf("game: invalid rules treasure_classes.%s.attempts.%s: unknown or inactive set item %s", classID, attempt.AttemptID, entry.SetItemID)
					}
				}
			}
		}
	}
	r.TreasureClasses = treasureClasses.Classes

	var monsters struct {
		Monsters map[string]MonsterDef `json:"monsters"`
	}
	if err := readJSON(filepath.Join(dir, "monsters.v0.json"), &monsters); err != nil {
		return nil, err
	}
	for id, def := range monsters.Monsters {
		if def.HitChance != nil && (*def.HitChance < 0 || *def.HitChance > 1) {
			return nil, fmt.Errorf("game: invalid rules monsters.%s.hit_chance: must be within [0,1]", id)
		}
		if def.CritChance != nil && (*def.CritChance < 0 || *def.CritChance > 1) {
			return nil, fmt.Errorf("game: invalid rules monsters.%s.crit_chance: must be within [0,1]", id)
		}
		if def.CritDamage != nil && *def.CritDamage < 1 {
			return nil, fmt.Errorf("game: invalid rules monsters.%s.crit_damage: must be >= 1", id)
		}
		if def.Armor < 0 {
			return nil, fmt.Errorf("game: invalid rules monsters.%s.armor: must be non-negative", id)
		}
		if def.BlockPercent < 0 || def.BlockPercent > 100 {
			return nil, fmt.Errorf("game: invalid rules monsters.%s.block_percent: must be within [0,100]", id)
		}
		for damageType, resistance := range def.Resistances {
			if err := validateDamageType("monsters."+id+".resistances", damageType); err != nil {
				return nil, err
			}
			if resistance < -1 || resistance > 1 {
				return nil, fmt.Errorf("game: invalid rules monsters.%s.resistances.%s: must be within [-1,1]", id, damageType)
			}
		}
		if def.XPReward < 0 {
			return nil, fmt.Errorf("game: invalid rules monsters.%s.xp_reward: must be non-negative", id)
		}
		if def.RetaliationDamage != nil {
			if err := validateDamageRange("monsters."+id+".retaliation_damage", *def.RetaliationDamage); err != nil {
				return nil, err
			}
		}
		if def.AttackDamage != nil {
			if err := validateDamageRange("monsters."+id+".attack_damage", *def.AttackDamage); err != nil {
				return nil, err
			}
			if def.AttackCooldown <= 0 {
				return nil, fmt.Errorf("game: invalid rules monsters.%s.attack_cooldown_ticks: required when attack_damage is set", id)
			}
		} else if def.AttackCooldown > 0 {
			return nil, fmt.Errorf("game: invalid rules monsters.%s.attack_cooldown_ticks: requires attack_damage", id)
		}
		attackMode := def.effectiveAttackMode()
		behavior := def.effectiveBehavior()
		attackStyle := def.effectiveAttackStyle()
		if !validMonsterNavigationTrait(def.NavigationTrait) {
			return nil, fmt.Errorf("game: invalid rules monsters.%s.navigation_trait: %s", id, def.NavigationTrait)
		}
		if def.PreferredMinRange < 0 {
			return nil, fmt.Errorf("game: invalid rules monsters.%s.preferred_min_range: must be non-negative", id)
		}
		switch attackMode {
		case attackModeMelee:
			if def.AttackRange > 0 && attackStyle != monsterAttackStylePounce {
				return nil, fmt.Errorf("game: invalid rules monsters.%s.attack_range: only valid for ranged or pounce attacks", id)
			}
			if def.ProjectileSpeed > 0 {
				return nil, fmt.Errorf("game: invalid rules monsters.%s.projectile_speed: only valid for ranged attacks", id)
			}
			if def.ProjectileDefID != "" {
				return nil, fmt.Errorf("game: invalid rules monsters.%s.projectile_def_id: only valid for ranged attacks", id)
			}
		case attackModeRanged:
			if def.AttackDamage == nil || def.AttackCooldown <= 0 {
				return nil, fmt.Errorf("game: invalid rules monsters.%s: ranged attacks require attack_damage and attack_cooldown_ticks", id)
			}
			if def.AttackRange <= r.Combat.UnarmedReach {
				return nil, fmt.Errorf("game: invalid rules monsters.%s.attack_range: must exceed melee reach", id)
			}
			if def.ProjectileSpeed <= 0 {
				return nil, fmt.Errorf("game: invalid rules monsters.%s.projectile_speed: must be positive for ranged attacks", id)
			}
			if def.ProjectileDefID == "" {
				return nil, fmt.Errorf("game: invalid rules monsters.%s.projectile_def_id: required for ranged attacks", id)
			}
			if def.PreferredMinRange > 0 && def.PreferredMinRange >= def.AttackRange {
				return nil, fmt.Errorf("game: invalid rules monsters.%s.preferred_min_range: must be below attack_range", id)
			}
		default:
			return nil, fmt.Errorf("game: invalid rules monsters.%s.attack_mode: %s", id, def.AttackMode)
		}
		if def.PreferredMinRange > 0 && attackMode != attackModeRanged {
			return nil, fmt.Errorf("game: invalid rules monsters.%s.preferred_min_range: only valid for ranged attacks", id)
		}
		if err := validateMonsterAttackStyle(id, def, attackMode, behavior, r.Combat.UnarmedReach); err != nil {
			return nil, err
		}
		if def.AttackWindupTicks < 0 {
			return nil, fmt.Errorf("game: invalid rules monsters.%s.attack_windup_ticks: must be non-negative", id)
		}
		if def.AttackWindupTicks > 0 && attackStyle != monsterAttackStyleMelee && attackStyle != monsterAttackStylePounce {
			return nil, fmt.Errorf("game: invalid rules monsters.%s.attack_windup_ticks: only valid for melee or pounce attack_style", id)
		}
		if def.AttackWindupTicks > 0 && attackMode != attackModeMelee {
			return nil, fmt.Errorf("game: invalid rules monsters.%s.attack_windup_ticks: only valid for melee attack_mode", id)
		}
		switch behavior {
		case monsterBehaviorStatic:
			if def.AggroRadius > 0 || def.AssistRadius > 0 || def.LeashRadius > 0 || def.MoveSpeed > 0 {
				return nil, fmt.Errorf("game: invalid rules monsters.%s: aggro/assist/leash/move_speed only valid for chase behavior", id)
			}
			if def.AttackDamage != nil {
				return nil, fmt.Errorf("game: invalid rules monsters.%s: attack_damage only valid for chase behavior", id)
			}
			if attackMode == attackModeRanged {
				return nil, fmt.Errorf("game: invalid rules monsters.%s: ranged attacks only valid for chase behavior", id)
			}
		case monsterBehaviorChase:
			if def.AggroRadius <= 0 {
				return nil, fmt.Errorf("game: invalid rules monsters.%s: chase requires positive aggro_radius", id)
			}
			if def.AssistRadius > 0 && def.AssistRadius < def.AggroRadius {
				return nil, fmt.Errorf("game: invalid rules monsters.%s: assist_radius must be >= aggro_radius", id)
			}
			if def.LeashRadius > 0 && def.LeashRadius < def.AggroRadius {
				return nil, fmt.Errorf("game: invalid rules monsters.%s: leash_radius must be >= aggro_radius", id)
			}
			if def.MoveSpeed > 0 && def.MoveSpeed > r.Navigation.CellSize {
				return nil, fmt.Errorf("game: invalid rules monsters.%s: move_speed must be <= navigation.cell_size %.1f", id, r.Navigation.CellSize)
			}
		default:
			return nil, fmt.Errorf("game: invalid rules monsters.%s.behavior: %s", id, def.Behavior)
		}
	}
	r.Monsters = monsters.Monsters
	if err := validateSkillCompanionMonsterRefs(r.Skills, r.Monsters); err != nil {
		return nil, err
	}
	if err := r.loadMercenaryRules(dir); err != nil {
		return nil, err
	}

	var loot struct {
		LootTables map[string]LootTable `json:"loot_tables"`
	}
	if err := readJSON(filepath.Join(dir, "loot_tables.v0.json"), &loot); err != nil {
		return nil, err
	}
	r.LootTables = loot.LootTables
	for tableID, table := range r.LootTables {
		if table.TreasureClassID != "" {
			if len(table.Entries) > 0 || len(table.Drops) > 0 {
				return nil, fmt.Errorf("game: invalid rules loot_tables.%s: treasure_class_id cannot mix with drops or entries", tableID)
			}
			if _, ok := r.TreasureClasses[table.TreasureClassID]; !ok {
				return nil, fmt.Errorf("game: invalid rules loot_tables.%s: unknown treasure class %s", tableID, table.TreasureClassID)
			}
			continue
		}
		for _, entry := range table.Entries {
			if countDropEntryRefs(entry.ItemDefID, entry.ItemTemplateID, entry.UniqueItemID, entry.SetItemID) != 1 {
				return nil, fmt.Errorf("game: invalid rules loot_tables.%s: entry must declare exactly one drop reference", tableID)
			}
			if entry.ItemDefID != "" {
				if _, ok := r.Items[entry.ItemDefID]; !ok {
					return nil, fmt.Errorf("game: invalid rules loot_tables.%s: unknown item %s", tableID, entry.ItemDefID)
				}
			}
			if entry.ItemTemplateID != "" {
				if _, ok := r.ItemTemplates[entry.ItemTemplateID]; !ok {
					return nil, fmt.Errorf("game: invalid rules loot_tables.%s: unknown item template %s", tableID, entry.ItemTemplateID)
				}
			}
			if entry.UniqueItemID != "" {
				unique, ok := r.UniqueItems[entry.UniqueItemID]
				if !ok || !unique.Enabled || unique.Status != "ready" {
					return nil, fmt.Errorf("game: invalid rules loot_tables.%s: unknown or inactive unique item %s", tableID, entry.UniqueItemID)
				}
			}
			if entry.SetItemID != "" {
				if _, ok := r.SetItems[entry.SetItemID]; !ok {
					return nil, fmt.Errorf("game: invalid rules loot_tables.%s: unknown or inactive set item %s", tableID, entry.SetItemID)
				}
			}
			if entry.Weight <= 0 {
				return nil, fmt.Errorf("game: invalid rules loot_tables.%s: entry weight must be positive", tableID)
			}
		}
		for _, itemDefID := range table.Drops {
			if _, ok := r.Items[itemDefID]; !ok {
				return nil, fmt.Errorf("game: invalid rules loot_tables.%s: unknown drop item %s", tableID, itemDefID)
			}
		}
	}
	for id, def := range r.Monsters {
		if _, ok := r.LootTables[def.LootTable]; !ok {
			return nil, fmt.Errorf("game: invalid rules monsters.%s: unknown loot table %s", id, def.LootTable)
		}
	}

	var shops struct {
		Shops map[string]ShopDef `json:"shops"`
	}
	if err := readJSON(filepath.Join(dir, "shops.v0.json"), &shops); err != nil {
		return nil, err
	}
	r.Shops = shops.Shops
	if err := validateShopRules(r); err != nil {
		return nil, err
	}

	var interactables struct {
		Interactables map[string]InteractableDef `json:"interactables"`
	}
	if err := readJSON(filepath.Join(dir, "interactables.v0.json"), &interactables); err != nil {
		return nil, err
	}
	for id, def := range interactables.Interactables {
		switch def.InitialState {
		case interactableClosed:
			if def.Transition != "" {
				return nil, fmt.Errorf("game: invalid rules interactables.%s.transition: closed interactable must not declare transition", id)
			}
			if def.ShopID != "" {
				return nil, fmt.Errorf("game: invalid rules interactables.%s.shop_id: closed interactable must not declare shop_id", id)
			}
			if def.StashID != "" {
				return nil, fmt.Errorf("game: invalid rules interactables.%s.stash_id: closed interactable must not declare stash_id", id)
			}
			if def.Service != "" {
				return nil, fmt.Errorf("game: invalid rules interactables.%s.service: closed interactable must not declare service", id)
			}
			if def.BarrierWhenClosed != nil && (def.BarrierWhenClosed.Size.X <= 0 || def.BarrierWhenClosed.Size.Y <= 0) {
				return nil, fmt.Errorf("game: invalid rules interactables.%s.barrier_when_closed.size: must be positive", id)
			}
			if def.LockedExitReason != "" && def.BarrierWhenClosed == nil {
				return nil, fmt.Errorf("game: invalid rules interactables.%s.locked_exit_reason: requires barrier_when_closed", id)
			}
		case interactableReady, interactableLocked, interactableDisabled:
			if def.BarrierWhenClosed != nil {
				return nil, fmt.Errorf("game: invalid rules interactables.%s.barrier_when_closed: transition interactable must not declare barrier", id)
			}
			actionCount := 0
			if def.Transition != "" {
				actionCount++
			}
			if def.ShopID != "" {
				actionCount++
			}
			if def.StashID != "" {
				actionCount++
			}
			if def.Service != "" {
				actionCount++
			}
			if actionCount != 1 {
				return nil, fmt.Errorf("game: invalid rules interactables.%s: must declare exactly one of transition, shop_id, stash_id, or service", id)
			}
			if def.Transition != "" {
				switch def.Transition {
				case interactableTransitionAscend, interactableTransitionDescend, interactableTransitionWaypoint:
				default:
					return nil, fmt.Errorf("game: invalid rules interactables.%s.transition: must be ascend, descend, or waypoint", id)
				}
			}
			if def.ShopID != "" {
				if _, ok := r.Shops[def.ShopID]; !ok {
					return nil, fmt.Errorf("game: invalid rules interactables.%s.shop_id: unknown shop %s", id, def.ShopID)
				}
			}
			if def.StashID != "" && def.StashID != "account_stash" {
				return nil, fmt.Errorf("game: invalid rules interactables.%s.stash_id: unknown stash %s", id, def.StashID)
			}
			if def.Service != "" && def.Service != "bishop" && def.Service != "market" && def.Service != "mercenary" && def.Service != "blacksmith" && def.Service != uniqueTestChestService && def.Service != questTurnInService {
				return nil, fmt.Errorf("game: invalid rules interactables.%s.service: unsupported service %s", id, def.Service)
			}
		default:
			return nil, fmt.Errorf("game: invalid rules interactables.%s.initial_state: unsupported state %s", id, def.InitialState)
		}
	}
	r.Interactables = interactables.Interactables

	var dungeonGeneration struct {
		Version                  int                      `json:"version"`
		FloorSize                DungeonFloorSize         `json:"floor_size"`
		FloorProfiles            []DungeonFloorProfile    `json:"floor_profiles"`
		WallThickness            float64                  `json:"wall_thickness"`
		PlayerSpawn              Vec2                     `json:"player_spawn"`
		StairPlacement           StairPlacementRules      `json:"stair_placement"`
		TeleporterPlacement      TeleporterPlacementRules `json:"teleporter_placement"`
		MonsterPlacement         MonsterPlacementRules    `json:"monster_placement"`
		ChestPlacement           ChestPlacementRules      `json:"chest_placement"`
		EliteObjective           EliteObjectiveRules      `json:"elite_objective"`
		RoomLayout               RoomLayoutRules          `json:"room_layout"`
		RoomCorridorPCG          RoomCorridorPCGRules     `json:"room_corridor_pcg"`
		ObstacleGeneration       ObstacleGenerationRules  `json:"obstacle_generation"`
		BossFloor                BossFloorRules           `json:"boss_floor"`
		MonsterRarityNote        string                   `json:"monster_rarity_note"`
		MonsterDepthScaling      MonsterDepthScalingRules `json:"monster_depth_scaling"`
		ItemLevelTiers           ItemLevelTierRules       `json:"item_level_tiers"`
		MonsterRarities          []MonsterRarityDef       `json:"monster_rarities"`
		LootBandNote             string                   `json:"loot_band_note"`
		LootBands                []DungeonLootBand        `json:"loot_bands"`
		LevelNames               map[string]string        `json:"level_names"`
		DefaultLevelNameTemplate string                   `json:"default_level_name_template"`
	}
	if err := readJSON(filepath.Join(dir, "dungeon_generation.v0.json"), &dungeonGeneration); err != nil {
		return nil, err
	}
	if dungeonGeneration.FloorSize.Width < 16 || dungeonGeneration.FloorSize.Height < 10 {
		return nil, fmt.Errorf("game: invalid rules dungeon_generation.floor_size: must be at least 16x10")
	}
	if dungeonGeneration.WallThickness <= 0 {
		return nil, fmt.Errorf("game: invalid rules dungeon_generation.wall_thickness: must be positive")
	}
	if dungeonGeneration.StairPlacement.MinSeparation <= 0 {
		return nil, fmt.Errorf("game: invalid rules dungeon_generation.stair_placement.min_separation: must be positive")
	}
	if dungeonGeneration.StairPlacement.MarginFromWall < 0 {
		return nil, fmt.Errorf("game: invalid rules dungeon_generation.stair_placement.margin_from_wall: must be non-negative")
	}
	if dungeonGeneration.StairPlacement.MaxAttempts <= 0 {
		return nil, fmt.Errorf("game: invalid rules dungeon_generation.stair_placement.max_attempts: must be positive")
	}
	if dungeonGeneration.TeleporterPlacement.MarginFromWall < 0 {
		return nil, fmt.Errorf("game: invalid rules dungeon_generation.teleporter_placement.margin_from_wall: must be non-negative")
	}
	if dungeonGeneration.TeleporterPlacement.MinStairDistance <= 0 {
		return nil, fmt.Errorf("game: invalid rules dungeon_generation.teleporter_placement.min_stair_distance: must be positive")
	}
	if dungeonGeneration.TeleporterPlacement.MaxAttempts <= 0 {
		return nil, fmt.Errorf("game: invalid rules dungeon_generation.teleporter_placement.max_attempts: must be positive")
	}
	if err := validateAreaCountFormula("dungeon_generation.monster_placement.population_formula", dungeonGeneration.MonsterPlacement.PopulationFormula); err != nil {
		return nil, err
	}
	if err := validateAreaRangeFormula("dungeon_generation.monster_placement.pack_count_formula", dungeonGeneration.MonsterPlacement.PackCountFormula); err != nil {
		return nil, err
	}
	monsterID := dungeonGeneration.MonsterPlacement.MonsterDefID
	monsterDef, ok := r.Monsters[monsterID]
	if !ok {
		return nil, fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.monster_def_id: unknown monster %s", monsterID)
	}
	if monsterDef.effectiveBehavior() != monsterBehaviorChase {
		return nil, fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.monster_def_id: %s must use chase behavior", monsterID)
	}
	baseDungeonGeneration := DungeonGenerationRules{
		FloorSize:          dungeonGeneration.FloorSize,
		MonsterPlacement:   dungeonGeneration.MonsterPlacement,
		ObstacleGeneration: dungeonGeneration.ObstacleGeneration,
	}.withDensityForSize(dungeonGeneration.FloorSize)
	if err := validateMonsterPlacementPool(baseDungeonGeneration.MonsterPlacement, r); err != nil {
		return nil, err
	}
	if aura := dungeonGeneration.MonsterPlacement.EliteAura; aura != nil {
		if aura.ID == "" {
			return nil, fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.elite_aura.id: must be non-empty")
		}
		if aura.Radius <= 0 {
			return nil, fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.elite_aura.radius: must be positive")
		}
		if aura.DamageBonusPercent < 0 || aura.DamageBonusPercent > 200 {
			return nil, fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.elite_aura.damage_bonus_percent: must be between 0 and 200")
		}
	}
	if dungeonGeneration.MonsterPlacement.MarginFromWall < 0 {
		return nil, fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.margin_from_wall: must be non-negative")
	}
	if dungeonGeneration.MonsterPlacement.MinSpawnDistance <= 0 {
		return nil, fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.min_spawn_distance: must be positive")
	}
	if dungeonGeneration.MonsterPlacement.MaxAttempts <= 0 {
		return nil, fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.max_attempts: must be positive")
	}
	chestPlacement := dungeonGeneration.ChestPlacement
	if chestPlacement.Enabled {
		if chestPlacement.ChanceWeight+chestPlacement.NoChestWeight <= 0 {
			return nil, fmt.Errorf("game: invalid rules dungeon_generation.chest_placement: chance/no_chest total must be positive")
		}
		if chestPlacement.InteractableDefID != treasureChestDefID {
			return nil, fmt.Errorf("game: invalid rules dungeon_generation.chest_placement.interactable_def_id: must be %s", treasureChestDefID)
		}
		if _, ok := r.Interactables[chestPlacement.InteractableDefID]; !ok {
			return nil, fmt.Errorf("game: invalid rules dungeon_generation.chest_placement.interactable_def_id: unknown interactable %s", chestPlacement.InteractableDefID)
		}
		table, ok := r.LootTables[chestPlacement.LootTable]
		if !ok {
			return nil, fmt.Errorf("game: invalid rules dungeon_generation.chest_placement.loot_table: unknown table %s", chestPlacement.LootTable)
		}
		if table.TreasureClassID == "" {
			return nil, fmt.Errorf("game: invalid rules dungeon_generation.chest_placement.loot_table: must resolve to a treasure class")
		}
		if chestPlacement.MonsterCountBonus < 0 {
			return nil, fmt.Errorf("game: invalid rules dungeon_generation.chest_placement.monster_count_bonus: must be non-negative")
		}
		if chestPlacement.MinStairDistance <= 0 {
			return nil, fmt.Errorf("game: invalid rules dungeon_generation.chest_placement.min_stair_distance: must be positive")
		}
		if chestPlacement.MaxAttempts <= 0 {
			return nil, fmt.Errorf("game: invalid rules dungeon_generation.chest_placement.max_attempts: must be positive")
		}
	}
	if err := validateEliteObjectiveRules(dungeonGeneration.EliteObjective, r); err != nil {
		return nil, err
	}
	if err := validateRoomLayoutRules(dungeonGeneration.RoomLayout); err != nil {
		return nil, err
	}
	if err := validateRoomCorridorPCGRules(dungeonGeneration.RoomCorridorPCG); err != nil {
		return nil, err
	}
	if err := validateAreaRangeFormula("dungeon_generation.obstacle_generation.target_group_count_formula", dungeonGeneration.ObstacleGeneration.TargetGroupCountFormula); err != nil {
		return nil, err
	}
	if err := validateObstacleGenerationRules(dungeonGeneration.ObstacleGeneration, dungeonGeneration.FloorSize); err != nil {
		return nil, err
	}
	if err := validateDoorGenerationRules(dungeonGeneration.ObstacleGeneration.Doors, r); err != nil {
		return nil, err
	}
	if err := validateWaterGenerationRules(dungeonGeneration.ObstacleGeneration.Water, dungeonGeneration.FloorSize); err != nil {
		return nil, err
	}
	if err := validateHoleGenerationRules(dungeonGeneration.ObstacleGeneration.Holes, dungeonGeneration.FloorSize); err != nil {
		return nil, err
	}
	if err := validateDungeonFloorProfiles(dungeonGeneration.FloorProfiles); err != nil {
		return nil, err
	}
	if dungeonGeneration.LootBandNote == "" {
		return nil, fmt.Errorf("game: invalid rules dungeon_generation.loot_band_note: required")
	}
	if dungeonGeneration.MonsterRarityNote == "" {
		return nil, fmt.Errorf("game: invalid rules dungeon_generation.monster_rarity_note: required")
	}
	if err := validateMonsterDepthScaling(dungeonGeneration.MonsterDepthScaling, r.Combat); err != nil {
		return nil, err
	}
	if dungeonGeneration.ItemLevelTiers.LevelsPerTier < 1 {
		return nil, fmt.Errorf("game: invalid rules dungeon_generation.item_level_tiers.levels_per_tier: must be positive")
	}
	if err := validateMonsterRarities(dungeonGeneration.MonsterRarities); err != nil {
		return nil, err
	}
	if err := validateBossFloorRules(dungeonGeneration.BossFloor, r); err != nil {
		return nil, err
	}
	if err := validateDungeonLootBands(dungeonGeneration.LootBands, r); err != nil {
		return nil, err
	}
	for key := range dungeonGeneration.LevelNames {
		level, err := strconv.Atoi(key)
		if err != nil || level >= 0 {
			return nil, fmt.Errorf("game: invalid rules dungeon_generation.level_names.%s: key must be a negative integer string", key)
		}
	}
	monsterPackRoles := map[string]string{}
	for monsterID, def := range r.Monsters {
		if def.PackRole != "" {
			monsterPackRoles[monsterID] = def.PackRole
		}
	}
	r.DungeonGeneration = DungeonGenerationRules{
		FloorSize:                dungeonGeneration.FloorSize,
		FloorProfiles:            dungeonGeneration.FloorProfiles,
		WallThickness:            dungeonGeneration.WallThickness,
		PlayerSpawn:              dungeonGeneration.PlayerSpawn,
		StairPlacement:           dungeonGeneration.StairPlacement,
		TeleporterPlacement:      dungeonGeneration.TeleporterPlacement,
		MonsterPlacement:         dungeonGeneration.MonsterPlacement,
		ChestPlacement:           dungeonGeneration.ChestPlacement,
		EliteObjective:           dungeonGeneration.EliteObjective,
		RoomLayout:               dungeonGeneration.RoomLayout,
		RoomCorridorPCG:          dungeonGeneration.RoomCorridorPCG,
		ObstacleGeneration:       dungeonGeneration.ObstacleGeneration,
		BossFloor:                dungeonGeneration.BossFloor,
		MonsterRarityNote:        dungeonGeneration.MonsterRarityNote,
		MonsterDepthScaling:      dungeonGeneration.MonsterDepthScaling,
		ItemLevelTiers:           dungeonGeneration.ItemLevelTiers,
		MonsterRarities:          dungeonGeneration.MonsterRarities,
		LootBandNote:             dungeonGeneration.LootBandNote,
		LootBands:                dungeonGeneration.LootBands,
		LevelNames:               dungeonGeneration.LevelNames,
		DefaultLevelNameTemplate: dungeonGeneration.DefaultLevelNameTemplate,
		monsterPackRoles:         monsterPackRoles,
	}.withDensityForSize(dungeonGeneration.FloorSize)
	if err := r.applyMainConfigDungeonMonsterDropRate(); err != nil {
		return nil, err
	}
	if err := r.applyMainConfigBossDropRate(); err != nil {
		return nil, err
	}

	var bossPatterns struct {
		MinimumTelegraphTicks int                       `json:"minimum_telegraph_ticks"`
		Patterns              map[string]BossPatternDef `json:"patterns"`
	}
	if err := readJSON(filepath.Join(dir, "boss_patterns.v0.json"), &bossPatterns); err != nil {
		return nil, err
	}
	if bossPatterns.MinimumTelegraphTicks <= 0 {
		return nil, fmt.Errorf("game: invalid rules boss_patterns.minimum_telegraph_ticks: must be positive")
	}
	if err := validateBossPatterns(bossPatterns.Patterns, bossPatterns.MinimumTelegraphTicks, r.Monsters); err != nil {
		return nil, err
	}
	r.BossPatterns = bossPatterns.Patterns

	var bossTemplates struct {
		Bosses map[string]BossTemplateDef `json:"bosses"`
	}
	if err := readJSON(filepath.Join(dir, "boss_templates.v0.json"), &bossTemplates); err != nil {
		return nil, err
	}
	if err := validateBossTemplates(bossTemplates.Bosses, r); err != nil {
		return nil, err
	}
	r.BossTemplates = bossTemplates.Bosses

	var worlds struct {
		Worlds map[string]WorldDef `json:"worlds"`
	}
	if err := readJSON(filepath.Join(dir, "worlds.v0.json"), &worlds); err != nil {
		return nil, err
	}
	for worldID, world := range worlds.Worlds {
		switch world.Mode {
		case "", worldModeMultiLevel:
		default:
			return nil, fmt.Errorf("game: invalid rules worlds.%s.mode: unsupported mode %s", worldID, world.Mode)
		}
		if world.StartLevel != nil {
			if world.Mode != worldModeMultiLevel {
				return nil, fmt.Errorf("game: invalid rules worlds.%s.start_level: only multi_level worlds can set start_level", worldID)
			}
			if *world.StartLevel > townLevel {
				return nil, fmt.Errorf("game: invalid rules worlds.%s.start_level: must be <= 0", worldID)
			}
		}
		for i, entity := range world.Entities {
			label := fmt.Sprintf("worlds.%s.entities[%d]", worldID, i)
			if entity.Type != wallEntity && entity.Kind != "" {
				return nil, fmt.Errorf("game: invalid rules %s.kind: only valid for wall entities", label)
			}
			switch entity.Type {
			case monsterEntity, companionEntity:
				if entity.MonsterDefID == "" {
					return nil, fmt.Errorf("game: invalid rules %s: missing monster_def_id", label)
				}
				if _, ok := r.Monsters[entity.MonsterDefID]; !ok {
					return nil, fmt.Errorf("game: invalid rules %s: unknown monster %s", label, entity.MonsterDefID)
				}
			case lootEntity:
				if err := validateWorldLootEntity(r, label, entity); err != nil {
					return nil, err
				}
			case wallEntity:
				if entity.Size.X <= 0 || entity.Size.Y <= 0 {
					return nil, fmt.Errorf("game: invalid rules %s: wall size must be positive", label)
				}
				if !validObstacleKind(entity.Kind) {
					return nil, fmt.Errorf("game: invalid rules %s.kind: unsupported wall kind %s", label, entity.Kind)
				}
			case interactableEntity:
				if entity.InteractableDefID == "" {
					return nil, fmt.Errorf("game: invalid rules %s: missing interactable_def_id", label)
				}
				if _, ok := r.Interactables[entity.InteractableDefID]; !ok {
					return nil, fmt.Errorf("game: invalid rules %s: unknown interactable %s", label, entity.InteractableDefID)
				}
			default:
				return nil, fmt.Errorf("game: invalid rules %s: unknown type %s", label, entity.Type)
			}
		}
	}
	r.Worlds = worlds.Worlds

	if err := loadQuestStewardRules(dir, r); err != nil {
		return nil, err
	}

	return r, nil
}

// LootDrop is one resolved loot table result.
type LootDrop struct {
	ItemDefID      string `json:"item_def_id,omitempty"`
	ItemTemplateID string `json:"item_template_id,omitempty"`
	UniqueItemID   string `json:"unique_item_id,omitempty"`
	SetItemID      string `json:"set_item_id,omitempty"`
}

// RollLoot selects an item or item template from a loot table using the RNG. A
// single-entry table is deterministic regardless of the draw.
func (r *Rules) RollLoot(tableID string, rng *RNG) (LootDrop, bool) {
	table, ok := r.LootTables[tableID]
	if !ok || len(table.Entries) == 0 {
		return LootDrop{}, false
	}
	total := 0
	for _, e := range table.Entries {
		total += e.Weight
	}
	if total <= 0 {
		return LootDrop{}, false
	}
	roll := rng.IntN(total)
	for _, e := range table.Entries {
		roll -= e.Weight
		if roll < 0 {
			return LootDrop{ItemDefID: e.ItemDefID, ItemTemplateID: e.ItemTemplateID, UniqueItemID: e.UniqueItemID, SetItemID: e.SetItemID}, true
		}
	}
	last := table.Entries[len(table.Entries)-1]
	return LootDrop{ItemDefID: last.ItemDefID, ItemTemplateID: last.ItemTemplateID, UniqueItemID: last.UniqueItemID, SetItemID: last.SetItemID}, true
}

// LootDrops returns all guaranteed drops for a table, or one weighted roll for
// legacy single-drop tables.
func (r *Rules) LootDrops(tableID string, rng *RNG) []LootDrop {
	table, ok := r.LootTables[tableID]
	if !ok {
		return nil
	}
	if table.TreasureClassID != "" {
		return r.RollTreasureClass(table.TreasureClassID, rng)
	}
	if len(table.Drops) > 0 {
		out := make([]LootDrop, 0, len(table.Drops))
		for _, itemDefID := range table.Drops {
			out = append(out, LootDrop{ItemDefID: itemDefID})
		}
		return out
	}
	drop, ok := r.RollLoot(tableID, rng)
	if !ok {
		return nil
	}
	return []LootDrop{drop}
}

func (r *Rules) RollTreasureClass(classID string, rng *RNG) []LootDrop {
	classDef, ok := r.TreasureClasses[classID]
	if !ok {
		return nil
	}
	out := []LootDrop{}
	for _, attempt := range classDef.Attempts {
		total := attempt.SuccessWeight + attempt.NoDropWeight
		if total <= 0 {
			continue
		}
		if rng.IntN(total) >= attempt.SuccessWeight {
			continue
		}
		totalEntries := 0
		for _, entry := range attempt.Entries {
			totalEntries += entry.Weight
		}
		if totalEntries <= 0 {
			continue
		}
		roll := rng.IntN(totalEntries)
		for _, entry := range attempt.Entries {
			roll -= entry.Weight
			if roll < 0 {
				out = append(out, LootDrop{ItemDefID: entry.ItemDefID, ItemTemplateID: entry.ItemTemplateID, UniqueItemID: entry.UniqueItemID, SetItemID: entry.SetItemID})
				break
			}
		}
	}
	return out
}

func validateMonsterPlacementPool(placement MonsterPlacementRules, r *Rules) error {
	if placement.Count == 0 {
		if len(placement.MonsterPool) > 0 || len(placement.MinimumMonsters) > 0 {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement: pool/minimums require positive count")
		}
		return nil
	}
	if placement.PackCount.Min <= 0 || placement.PackCount.Max < placement.PackCount.Min {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.pack_count: invalid min/max")
	}
	if placement.PackSize.Min <= 0 || placement.PackSize.Max < placement.PackSize.Min {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.pack_size: invalid min/max")
	}
	if placement.PackMemberRadius <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.pack_member_radius: must be positive")
	}
	if placement.PackComposition.FrontlineMin <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.pack_composition.frontline_min: must be positive")
	}
	if placement.PackComposition.RangedMax < 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.pack_composition.ranged_max: must be non-negative")
	}
	if placement.PackComposition.FrontlineMin > placement.PackSize.Min {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.pack_composition.frontline_min: must fit minimum pack size")
	}
	if placement.PackComposition.RangedMax > placement.PackSize.Max {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.pack_composition.ranged_max: must fit maximum pack size")
	}
	if placement.ElitePackChance < 0 || placement.ElitePackChance > 100 {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.elite_pack_chance_percent: must be 0..100")
	}
	if placement.PackCount.Min*placement.PackSize.Min > placement.Count {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement: minimum pack population exceeds count")
	}
	if placement.PackCount.Max*placement.PackSize.Max < placement.Count {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement: maximum pack population below count")
	}

	poolWeight := 0
	poolIDs := map[string]bool{}
	poolRoles := map[string]bool{}
	minAssistRadius := math.MaxFloat64
	considerAssist := func(monsterID string) error {
		def, ok := r.Monsters[monsterID]
		if !ok {
			return fmt.Errorf("unknown monster %s", monsterID)
		}
		if def.effectiveBehavior() != monsterBehaviorChase {
			return fmt.Errorf("%s must use chase behavior", monsterID)
		}
		if def.PackRole == "" {
			return fmt.Errorf("%s must declare pack_role", monsterID)
		}
		minAssistRadius = math.Min(minAssistRadius, def.effectiveAssistRadius())
		poolRoles[def.PackRole] = true
		return nil
	}
	for idx, entry := range placement.MonsterPool {
		if entry.MonsterDefID == "" {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.monster_pool[%d].monster_def_id: required", idx)
		}
		if entry.Weight <= 0 {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.monster_pool[%d].weight: must be positive", idx)
		}
		if err := considerAssist(entry.MonsterDefID); err != nil {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.monster_pool[%d].monster_def_id: %v", idx, err)
		}
		poolIDs[entry.MonsterDefID] = true
		poolWeight += entry.Weight
	}
	if len(placement.MonsterPool) == 0 {
		if err := considerAssist(placement.MonsterDefID); err != nil {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.monster_def_id: %v", err)
		}
	}
	if len(placement.MonsterPool) > 0 && poolWeight <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.monster_pool: total weight must be positive")
	}
	if !poolRoles["frontline"] {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.monster_pool: at least one frontline pack_role required")
	}
	if minAssistRadius != math.MaxFloat64 && placement.PackMemberRadius*2 > minAssistRadius {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.pack_member_radius: diameter must fit within monster assist_radius")
	}

	minTotal := 0
	for idx, entry := range placement.MinimumMonsters {
		if entry.MonsterDefID == "" {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.minimum_monsters[%d].monster_def_id: required", idx)
		}
		if entry.Count < 0 {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.minimum_monsters[%d].count: must be non-negative", idx)
		}
		def, ok := r.Monsters[entry.MonsterDefID]
		if !ok {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.minimum_monsters[%d].monster_def_id: unknown monster %s", idx, entry.MonsterDefID)
		}
		if def.effectiveBehavior() != monsterBehaviorChase {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.minimum_monsters[%d].monster_def_id: %s must use chase behavior", idx, entry.MonsterDefID)
		}
		if len(placement.MonsterPool) > 0 && !poolIDs[entry.MonsterDefID] {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.minimum_monsters[%d].monster_def_id: %s missing from monster_pool", idx, entry.MonsterDefID)
		}
		minTotal += entry.Count
	}
	if minTotal > placement.Count {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.minimum_monsters: total %d exceeds count %d", minTotal, placement.Count)
	}

	return nil
}

func validateMonsterDepthScaling(s MonsterDepthScalingRules, combat Combat) error {
	if s.HPPerDepth < 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_depth_scaling.hp_per_depth: must be non-negative")
	}
	if s.DamagePerDepth < 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_depth_scaling.damage_per_depth: must be non-negative")
	}
	if s.ArmorPerDepth < 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_depth_scaling.armor_per_depth: must be non-negative")
	}
	if s.HitChancePerDepth < 0 || s.MaxHitChance <= 0 || s.MaxHitChance > 1 {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_depth_scaling.hit_chance: invalid per-depth or max")
	}
	if s.CritChancePerDepth < 0 || s.MaxCritChance < 0 || s.MaxCritChance > 1 {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_depth_scaling.crit_chance: invalid per-depth or max")
	}
	if s.BlockPercentPerDepth < 0 || s.MaxBlockPercent < 0 || s.MaxBlockPercent > float64(combat.BlockCap) {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_depth_scaling.block_percent: invalid per-depth or max")
	}
	if s.AttackCooldownMultiplierPerDepth <= 0 || s.AttackCooldownMultiplierPerDepth > 1 {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_depth_scaling.attack_cooldown_multiplier_per_depth: must be > 0 and <= 1")
	}
	if s.MinAttackCooldownTicks <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_depth_scaling.min_attack_cooldown_ticks: must be positive")
	}
	return nil
}

func validateMonsterRarities(rarities []MonsterRarityDef) error {
	if len(rarities) == 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_rarities: must not be empty")
	}
	seen := map[string]bool{}
	for idx, rarity := range rarities {
		if rarity.ID == "" || strings.ToLower(rarity.ID) != rarity.ID {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_rarities[%d].id: must be stable lowercase", idx)
		}
		if seen[rarity.ID] {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_rarities[%d].id: duplicate %s", idx, rarity.ID)
		}
		seen[rarity.ID] = true
		if rarity.Weight <= 0 {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_rarities[%d].weight: must be positive", idx)
		}
		if len(rarity.Color) != 7 || rarity.Color[0] != '#' {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_rarities[%d].color: must be #RRGGBB", idx)
		}
		if _, err := strconv.ParseUint(rarity.Color[1:], 16, 32); err != nil {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_rarities[%d].color: must be hex", idx)
		}
		if rarity.HPMultiplier <= 0 {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_rarities[%d].hp_multiplier: must be positive", idx)
		}
		if rarity.DamageMultiplier <= 0 {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_rarities[%d].damage_multiplier: must be positive", idx)
		}
		if rarity.XPMultiplier <= 0 {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_rarities[%d].xp_multiplier: must be positive", idx)
		}
		if rarity.ArmorMultiplier <= 0 {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_rarities[%d].armor_multiplier: must be positive", idx)
		}
		if rarity.ArmorBonus < 0 {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_rarities[%d].armor_bonus: must be non-negative", idx)
		}
		if rarity.HitChanceBonus < 0 || rarity.HitChanceBonus > 1 {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_rarities[%d].hit_chance_bonus: must be between 0 and 1", idx)
		}
		if rarity.CritChanceBonus < 0 || rarity.CritChanceBonus > 1 {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_rarities[%d].crit_chance_bonus: must be between 0 and 1", idx)
		}
		if rarity.BlockPercentBonus < 0 || rarity.BlockPercentBonus > 100 {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_rarities[%d].block_percent_bonus: must be between 0 and 100", idx)
		}
		if rarity.AttackCooldownMultiplier <= 0 || rarity.AttackCooldownMultiplier > 1 {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_rarities[%d].attack_cooldown_multiplier: must be > 0 and <= 1", idx)
		}
		if rarity.LootDepthOffset < 0 {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_rarities[%d].loot_depth_offset: must be non-negative", idx)
		}
		if rarity.VisualScale <= 0 {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_rarities[%d].visual_scale: must be positive", idx)
		}
	}
	return nil
}

func validateBossFloorRules(b BossFloorRules, r *Rules) error {
	if b.Cadence != 5 {
		return fmt.Errorf("game: invalid rules dungeon_generation.boss_floor.cadence: expected 5")
	}
	if b.FirstLevel != -5 {
		return fmt.Errorf("game: invalid rules dungeon_generation.boss_floor.first_level: expected -5")
	}
	if b.FloorSize.Width != 30 || b.FloorSize.Height != 30 {
		return fmt.Errorf("game: invalid rules dungeon_generation.boss_floor.floor_size: expected 30x30")
	}
	if b.MonsterCount < 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.boss_floor.monster_count: must be non-negative")
	}
	if b.ChestInteractableDefID != treasureChestDefID {
		return fmt.Errorf("game: invalid rules dungeon_generation.boss_floor.chest_interactable_def_id: expected %s", treasureChestDefID)
	}
	if _, ok := r.Interactables[b.ChestInteractableDefID]; !ok {
		return fmt.Errorf("game: invalid rules dungeon_generation.boss_floor.chest_interactable_def_id: unknown interactable %s", b.ChestInteractableDefID)
	}
	if table, ok := r.LootTables[b.ChestLootTable]; !ok {
		return fmt.Errorf("game: invalid rules dungeon_generation.boss_floor.chest_loot_table: unknown table %s", b.ChestLootTable)
	} else if table.TreasureClassID == "" {
		return fmt.Errorf("game: invalid rules dungeon_generation.boss_floor.chest_loot_table: must resolve to a treasure class")
	}
	if len(b.BossTemplatePool) == 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.boss_floor.boss_template_pool: required")
	}
	if b.LockedExitReason != "boss_alive" {
		return fmt.Errorf("game: invalid rules dungeon_generation.boss_floor.locked_exit_reason: expected boss_alive")
	}
	for label, point := range map[string]Vec2{
		"boss_spawn":           b.BossSpawn,
		"chest_position":       b.ChestPosition,
		"stairs_up_position":   b.StairsUpPosition,
		"stairs_down_position": b.StairsDownPosition,
		"teleporter_position":  b.TeleporterPosition,
	} {
		if point.X < 0 || point.Y < 0 || point.X > b.FloorSize.Width || point.Y > b.FloorSize.Height {
			return fmt.Errorf("game: invalid rules dungeon_generation.boss_floor.%s: outside floor", label)
		}
	}
	return nil
}

func validateBossTemplates(templates map[string]BossTemplateDef, r *Rules) error {
	if len(templates) == 0 {
		return fmt.Errorf("game: invalid rules boss_templates.bosses: required")
	}
	for templateID, template := range templates {
		if _, ok := r.Monsters[template.BaseMonsterDefID]; !ok {
			return fmt.Errorf("game: invalid rules boss_templates.%s.base_monster_def_id: unknown monster %s", templateID, template.BaseMonsterDefID)
		}
		if len(template.PatternDeck) == 0 {
			return fmt.Errorf("game: invalid rules boss_templates.%s.pattern_deck: required", templateID)
		}
		for _, patternID := range template.PatternDeck {
			if _, ok := r.BossPatterns[patternID]; !ok {
				return fmt.Errorf("game: invalid rules boss_templates.%s.pattern_deck: unknown pattern %s", templateID, patternID)
			}
		}
		if template.HPMultiplier <= 0 || template.DamageMultiplier <= 0 {
			return fmt.Errorf("game: invalid rules boss_templates.%s: multipliers must be positive", templateID)
		}
		if err := validateBossTemplateEnrage(templateID, template.Enrage); err != nil {
			return err
		}
		if _, ok := r.LootTables[template.LootTable]; !ok {
			return fmt.Errorf("game: invalid rules boss_templates.%s.loot_table: unknown table %s", templateID, template.LootTable)
		}
		if template.Visual.Model == "" {
			return fmt.Errorf("game: invalid rules boss_templates.%s.visual.model: required", templateID)
		}
		if template.Visual.Scale <= 0 {
			return fmt.Errorf("game: invalid rules boss_templates.%s.visual.scale: must be positive", templateID)
		}
		for idx, model := range template.Visual.ModelPool {
			switch model {
			case "monster_dummy", "monster_dark_purple", "monster_crocodile_archer", "monster_quadruped", "monster_wolf", "monster_tiny_flyer", "monster_skeleton":
			default:
				return fmt.Errorf("game: invalid rules boss_templates.%s.visual.model_pool[%d]: unknown model %s", templateID, idx, model)
			}
		}
	}
	for _, templateID := range r.DungeonGeneration.BossFloor.BossTemplatePool {
		if _, ok := templates[templateID]; !ok {
			return fmt.Errorf("game: invalid rules dungeon_generation.boss_floor.boss_template_pool: unknown template %s", templateID)
		}
	}
	return nil
}

func validateShopRules(r *Rules) error {
	for shopID, shop := range r.Shops {
		if shop.Name == "" {
			return fmt.Errorf("game: invalid rules shops.%s.name: required", shopID)
		}
		seenOffers := map[string]bool{}
		for _, offer := range shop.FixedOffers {
			if offer.OfferID == "" {
				return fmt.Errorf("game: invalid rules shops.%s.fixed_offers: offer_id required", shopID)
			}
			if seenOffers[offer.OfferID] {
				return fmt.Errorf("game: invalid rules shops.%s.fixed_offers: duplicate offer_id %s", shopID, offer.OfferID)
			}
			seenOffers[offer.OfferID] = true
			item, ok := r.Items[offer.ItemDefID]
			if !ok {
				return fmt.Errorf("game: invalid rules shops.%s.fixed_offers.%s: unknown item %s", shopID, offer.OfferID, offer.ItemDefID)
			}
			if item.Category == "currency" || item.Category == "quest" {
				return fmt.Errorf("game: invalid rules shops.%s.fixed_offers.%s: currency/quest items cannot be sold", shopID, offer.OfferID)
			}
			if offer.BuyPrice <= 0 {
				return fmt.Errorf("game: invalid rules shops.%s.fixed_offers.%s.buy_price: must be positive", shopID, offer.OfferID)
			}
		}
		gen := shop.GeneratedOffers
		if gen.OfferCount <= 0 {
			return fmt.Errorf("game: invalid rules shops.%s.generated_offers.offer_count: must be positive", shopID)
		}
		if gen.Source != "common_dungeon_mob" {
			return fmt.Errorf("game: invalid rules shops.%s.generated_offers.source: unsupported source %s", shopID, gen.Source)
		}
		if gen.MinDepth <= 0 {
			return fmt.Errorf("game: invalid rules shops.%s.generated_offers.min_depth: must be positive", shopID)
		}
		if gen.SourceDepthPolicy != "character_level_plus_one_to_deepest_else_any_achieved" {
			return fmt.Errorf("game: invalid rules shops.%s.generated_offers.source_depth_policy: unsupported policy %s", shopID, gen.SourceDepthPolicy)
		}
		if gen.MaxRarity == "" {
			return fmt.Errorf("game: invalid rules shops.%s.generated_offers.max_rarity: required", shopID)
		}
		if _, ok := r.Rarities[gen.MaxRarity]; !ok {
			return fmt.Errorf("game: invalid rules shops.%s.generated_offers.max_rarity: unknown rarity %s", shopID, gen.MaxRarity)
		}
		if !shopRarityAllowedByCap(gen.MaxRarity, "rare") {
			return fmt.Errorf("game: invalid rules shops.%s.generated_offers.max_rarity: must not exceed rare", shopID)
		}
		if gen.RefreshOn != "new_non_town_waypoint" {
			return fmt.Errorf("game: invalid rules shops.%s.generated_offers.refresh_on: unsupported trigger %s", shopID, gen.RefreshOn)
		}
		if gen.MaxRollAttempts < gen.OfferCount {
			return fmt.Errorf("game: invalid rules shops.%s.generated_offers.max_roll_attempts: must be >= offer_count", shopID)
		}
		if err := validateMysteryShopRules(r, shopID, shop); err != nil {
			return err
		}
		if !shop.MysteryOffers.Enabled && !shop.Buyback.Enabled {
			return fmt.Errorf("game: invalid rules shops.%s.buyback.enabled: must be true", shopID)
		}
		if shop.Buyback.Scope != "session_town_visit" {
			return fmt.Errorf("game: invalid rules shops.%s.buyback.scope: unsupported scope %s", shopID, shop.Buyback.Scope)
		}
		if shop.Buyback.BuyPriceMultiplier <= 0 {
			return fmt.Errorf("game: invalid rules shops.%s.buyback.buy_price_multiplier: must be positive", shopID)
		}
		if !shop.Buyback.ClearOnLeaveTown {
			return fmt.Errorf("game: invalid rules shops.%s.buyback.clear_on_leave_town: must be true", shopID)
		}
		pricing := shop.Pricing
		if pricing.SellMultiplier <= 0 || pricing.SellMultiplier > 1 {
			return fmt.Errorf("game: invalid rules shops.%s.pricing.sell_multiplier: must be within (0,1]", shopID)
		}
		if pricing.RoundBuyTo <= 0 {
			return fmt.Errorf("game: invalid rules shops.%s.pricing.round_buy_to: must be positive", shopID)
		}
		for rarityID := range r.Rarities {
			if !r.rarityRandomRollable(rarityID) {
				continue
			}
			if pricing.RarityMultipliers[rarityID] <= 0 {
				return fmt.Errorf("game: invalid rules shops.%s.pricing.rarity_multipliers.%s: must be positive", shopID, rarityID)
			}
		}
		if pricing.RarityMultipliers["unique"] <= 0 {
			return fmt.Errorf("game: invalid rules shops.%s.pricing.rarity_multipliers.unique: must be positive", shopID)
		}
		if pricing.RarityMultipliers["set"] <= 0 {
			return fmt.Errorf("game: invalid rules shops.%s.pricing.rarity_multipliers.set: must be positive", shopID)
		}
		for templateID, template := range r.ItemTemplates {
			if pricing.SlotBase[template.Slot] <= 0 {
				return fmt.Errorf("game: invalid rules shops.%s.pricing.slot_base.%s: required by template %s", shopID, template.Slot, templateID)
			}
			for stat := range template.BaseStats {
				if _, ok := pricing.StatWeights[stat]; !ok {
					return fmt.Errorf("game: invalid rules shops.%s.pricing.stat_weights.%s: required by template %s", shopID, stat, templateID)
				}
			}
			for _, roll := range template.RollableStats {
				if _, ok := pricing.StatWeights[roll.Stat]; !ok {
					return fmt.Errorf("game: invalid rules shops.%s.pricing.stat_weights.%s: required by template %s", shopID, roll.Stat, templateID)
				}
			}
		}
	}
	return nil
}

func validateMysteryShopRules(r *Rules, shopID string, shop ShopDef) error {
	mystery := shop.MysteryOffers
	if !mystery.Enabled {
		return nil
	}
	if mystery.Source != "common_dungeon_mob" {
		return fmt.Errorf("game: invalid rules shops.%s.mystery_offers.source: unsupported source %s", shopID, mystery.Source)
	}
	if mystery.SourceDepthWindow <= 0 {
		return fmt.Errorf("game: invalid rules shops.%s.mystery_offers.source_depth_window: must be positive", shopID)
	}
	minRank, ok := shopRarityRank(mystery.MinRarity)
	if !ok || minRank < 1 {
		return fmt.Errorf("game: invalid rules shops.%s.mystery_offers.min_rarity: must be magic or rare", shopID)
	}
	maxRank, ok := shopRarityRank(mystery.MaxRarity)
	if !ok || maxRank < 1 {
		return fmt.Errorf("game: invalid rules shops.%s.mystery_offers.max_rarity: must be magic, rare, unique, or set", shopID)
	}
	if minRank > maxRank {
		return fmt.Errorf("game: invalid rules shops.%s.mystery_offers: min_rarity must not exceed max_rarity", shopID)
	}
	if mystery.RefreshOn != "new_non_town_waypoint" {
		return fmt.Errorf("game: invalid rules shops.%s.mystery_offers.refresh_on: unsupported trigger %s", shopID, mystery.RefreshOn)
	}
	if mystery.PriceMultiplier <= 1 {
		return fmt.Errorf("game: invalid rules shops.%s.mystery_offers.price_multiplier: must be > 1", shopID)
	}
	if mystery.RerollCost <= 0 {
		return fmt.Errorf("game: invalid rules shops.%s.mystery_offers.reroll_cost: must be positive", shopID)
	}
	if mystery.MaxRollAttempts < len(mystery.EligibleSlots) {
		return fmt.Errorf("game: invalid rules shops.%s.mystery_offers.max_roll_attempts: must be >= eligible slot count", shopID)
	}
	slotsWithTemplates := map[string]bool{}
	for _, template := range r.ItemTemplates {
		slotsWithTemplates[template.Slot] = true
	}
	seenSlots := map[string]bool{}
	for _, slot := range mystery.EligibleSlots {
		if !isEquipmentSlot(slot) && slot != "ring" {
			return fmt.Errorf("game: invalid rules shops.%s.mystery_offers.eligible_slots: unsupported slot %s", shopID, slot)
		}
		if seenSlots[slot] {
			return fmt.Errorf("game: invalid rules shops.%s.mystery_offers.eligible_slots: duplicate slot %s", shopID, slot)
		}
		seenSlots[slot] = true
		if !slotsWithTemplates[slot] {
			return fmt.Errorf("game: invalid rules shops.%s.mystery_offers.eligible_slots: no template for slot %s", shopID, slot)
		}
	}
	return nil
}

func validateDungeonLootBands(bands []DungeonLootBand, r *Rules) error {
	if len(bands) != 3 {
		return fmt.Errorf("game: invalid rules dungeon_generation.loot_bands: expected exactly 3 bands")
	}
	coverage := map[int]bool{}
	openEnded := 0
	for idx, band := range bands {
		if band.MinDepth <= 0 {
			return fmt.Errorf("game: invalid rules dungeon_generation.loot_bands[%d].min_depth: must be positive", idx)
		}
		if band.MaxDepth != nil && *band.MaxDepth < band.MinDepth {
			return fmt.Errorf("game: invalid rules dungeon_generation.loot_bands[%d].max_depth: must be >= min_depth", idx)
		}
		if band.MaxDepth == nil {
			openEnded++
			if coverage[band.MinDepth] {
				return fmt.Errorf("game: invalid rules dungeon_generation.loot_bands[%d]: overlapping depth %d", idx, band.MinDepth)
			}
			coverage[band.MinDepth] = true
		} else {
			for depth := band.MinDepth; depth <= *band.MaxDepth; depth++ {
				if coverage[depth] {
					return fmt.Errorf("game: invalid rules dungeon_generation.loot_bands[%d]: overlapping depth %d", idx, depth)
				}
				coverage[depth] = true
			}
		}
		if _, ok := r.treasureClassIDForLootTable(band.MonsterLootTable); !ok {
			return fmt.Errorf("game: invalid rules dungeon_generation.loot_bands[%d].monster_loot_table: unknown treasure table %s", idx, band.MonsterLootTable)
		}
		if _, ok := r.treasureClassIDForLootTable(band.ChestLootTable); !ok {
			return fmt.Errorf("game: invalid rules dungeon_generation.loot_bands[%d].chest_loot_table: unknown treasure table %s", idx, band.ChestLootTable)
		}
		if r.treasureClassSuccessWeightForTable(band.ChestLootTable) <= r.treasureClassSuccessWeightForTable(band.MonsterLootTable) {
			return fmt.Errorf("game: invalid rules dungeon_generation.loot_bands[%d]: chest success weight must exceed monster success weight", idx)
		}
	}
	if !coverage[1] || !coverage[2] || !coverage[3] || len(coverage) != 3 || openEnded != 1 {
		return fmt.Errorf("game: invalid rules dungeon_generation.loot_bands: must cover depths 1, 2, and open-ended 3+")
	}
	last := bands[len(bands)-1]
	if last.MinDepth != 3 || last.MaxDepth != nil {
		return fmt.Errorf("game: invalid rules dungeon_generation.loot_bands: final band must be open-ended 3+")
	}
	reachable := r.templatesReachableFromLootTable(last.MonsterLootTable)
	for templateID := range r.templatesReachableFromLootTable(last.ChestLootTable) {
		reachable[templateID] = true
	}
	requiredTemplates := []string{
		"long_sword",
		"great_sword",
		"bow",
		"shield",
		"helm",
		"mail",
		"gloves",
		"belt",
		"boots",
		"ring",
		"amulet",
	}
	missing := []string{}
	for _, templateID := range requiredTemplates {
		if !reachable[templateID] {
			missing = append(missing, templateID)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("game: invalid rules dungeon_generation.loot_bands 3+ reachability: missing templates %v", missing)
	}
	return nil
}

func (r *Rules) treasureClassIDForLootTable(tableID string) (string, bool) {
	table, ok := r.LootTables[tableID]
	if !ok || table.TreasureClassID == "" {
		return "", false
	}
	if _, ok := r.TreasureClasses[table.TreasureClassID]; !ok {
		return "", false
	}
	return table.TreasureClassID, true
}

func (r *Rules) treasureClassSuccessWeightForTable(tableID string) int {
	classID, ok := r.treasureClassIDForLootTable(tableID)
	if !ok {
		return 0
	}
	total := 0
	for _, attempt := range r.TreasureClasses[classID].Attempts {
		total += attempt.SuccessWeight
	}
	return total
}

func (r *Rules) templatesReachableFromLootTable(tableID string) map[string]bool {
	out := map[string]bool{}
	classID, ok := r.treasureClassIDForLootTable(tableID)
	if !ok {
		return out
	}
	for _, attempt := range r.TreasureClasses[classID].Attempts {
		for _, entry := range attempt.Entries {
			if entry.ItemTemplateID != "" {
				out[entry.ItemTemplateID] = true
			}
		}
	}
	return out
}

func sortedStringKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("game: read rules %s: %w", path, err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("game: parse rules %s: %w", path, err)
	}
	return nil
}

func validateDamageRange(label string, d DamageRange) error {
	if d.Min < 0 || d.Max < 0 {
		return fmt.Errorf("game: invalid rules %s: min/max must be non-negative", label)
	}
	if d.Max < d.Min {
		return fmt.Errorf("game: invalid rules %s: max must be >= min", label)
	}
	return nil
}

func validateWeaponDamageScaling(scaling map[string]WeaponDamageScalingDef) error {
	if len(scaling) == 0 {
		return fmt.Errorf("game: invalid rules combat.weapon_damage_scaling: required")
	}
	for itemType, def := range scaling {
		if itemType == "" {
			return fmt.Errorf("game: invalid rules combat.weapon_damage_scaling: item type must be non-empty")
		}
		if def.DamageMin == "" || def.DamageMax == "" || def.SourceLabel == "" {
			return fmt.Errorf("game: invalid rules combat.weapon_damage_scaling.%s: damage_min, damage_max, and source_label are required", itemType)
		}
	}

	return nil
}

func validateCoopCombatRules(coop CoopCombat) error {
	if coop.XPShare.Enabled {
		if coop.XPShare.Radius <= 0 {
			return fmt.Errorf("game: invalid rules combat.coop.xp_share.radius: must be positive")
		}
		if !coop.XPShare.FullXPPerEligiblePlayer {
			return fmt.Errorf("game: invalid rules combat.coop.xp_share.full_xp_per_eligible_player: must be true for v48")
		}
		if coop.XPShare.IncludeDeadPlayers {
			return fmt.Errorf("game: invalid rules combat.coop.xp_share.include_dead_players: must be false for v48")
		}
		if coop.XPShare.IncludeDisconnected {
			return fmt.Errorf("game: invalid rules combat.coop.xp_share.include_disconnected_players: must be false for v48")
		}
	}
	if coop.PartyChallenge.Enabled {
		if coop.PartyChallenge.PerDoubleBonus < 0 {
			return fmt.Errorf("game: invalid rules combat.coop.party_challenge.per_double_bonus: must be non-negative")
		}
		if coop.PartyChallenge.MaxBonus < 0 {
			return fmt.Errorf("game: invalid rules combat.coop.party_challenge.max_bonus: must be non-negative")
		}
		if coop.PartyChallenge.MaxBonus < coop.PartyChallenge.PerDoubleBonus {
			return fmt.Errorf("game: invalid rules combat.coop.party_challenge.max_bonus: must be >= per_double_bonus")
		}
		if !coop.PartyChallenge.HPScalesAtSpawn {
			return fmt.Errorf("game: invalid rules combat.coop.party_challenge.hp_scales_at_spawn: must be true for v48")
		}
		if !coop.PartyChallenge.DamageScalesAtAttack {
			return fmt.Errorf("game: invalid rules combat.coop.party_challenge.damage_scales_at_attack: must be true for v48")
		}
	}
	return nil
}

func isEquipmentSlot(slot string) bool {
	switch slot {
	case "head", "amulet", "chest", "gloves", "belt", "boots", "ring_left", "ring_right", "main_hand", "off_hand":
		return true
	default:
		return false
	}
}

func isHandSlot(slot string) bool {
	return slot == "main_hand" || slot == "off_hand"
}

func isSupportedRequirementStat(stat string) bool {
	switch stat {
	case "level", "str", "dex", "vit", "magic":
		return true
	default:
		return false
	}
}

func isSupportedItemStat(stat string) bool {
	if isSkillLevelRollStat(stat) {
		return true
	}
	switch stat {
	case "damage_min", "damage_max", "str", "dex", "vit", "magic", "all_skills", "max_hp", "max_mana", "armor", "block_percent", "attack_speed_percent", "hit_chance", "crit_chance", "evade_chance", "health_regen_per_10_seconds", "mana_regen_per_10_seconds", "skill_damage_percent", "skill_cooldown_reduction_percent", "skill_mana_cost_reduction", "magic_find_percent", "light_radius", "hotbar_slots", "inventory_rows", "movement_speed_percent", "bonus_cold_damage", "bonus_fire_damage", "bonus_lightning_damage", "bonus_poison_damage":
		return true
	default:
		return false
	}
}

func occupiesExactly(slots []string, wantA, wantB string) bool {
	if len(slots) != 2 {
		return false
	}
	seen := map[string]bool{slots[0]: true, slots[1]: true}
	return seen[wantA] && seen[wantB]
}

// FindSharedRulesDir walks up from the current working directory looking for a
// "shared/rules" directory, returning its absolute path. Deployments should set
// ARPG_RULES_DIR explicitly instead of relying on this search.
func FindSharedRulesDir() (string, error) {
	if dir := os.Getenv("ARPG_RULES_DIR"); dir != "" {
		return dir, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for i := 0; i < 8; i++ {
		candidate := filepath.Join(dir, "shared", "rules")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("game: could not locate shared/rules from %s (set ARPG_RULES_DIR)", cwd)
}
