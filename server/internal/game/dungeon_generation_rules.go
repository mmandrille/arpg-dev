package game

// Dungeon generation rule types and helpers. Extracted from rules.go to shrink the
// rules loader while keeping dungeon-generation JSON shape in one focused module.

// DungeonGenerationRules controls deterministic generated dungeon floors.
type DungeonGenerationRules struct {
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
	ObstacleGeneration       ObstacleGenerationRules  `json:"obstacle_generation"`
	BossFloor                BossFloorRules           `json:"boss_floor"`
	MonsterRarityNote        string                   `json:"monster_rarity_note"`
	MonsterDepthScaling      MonsterDepthScalingRules `json:"monster_depth_scaling"`
	MonsterRarities          []MonsterRarityDef       `json:"monster_rarities"`
	LootBandNote             string                   `json:"loot_band_note"`
	LootBands                []DungeonLootBand        `json:"loot_bands"`
	LevelNames               map[string]string        `json:"level_names"`
	DefaultLevelNameTemplate string                   `json:"default_level_name_template"`
	monsterPackRoles         map[string]string
}

type DungeonFloorSize struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type BossFloorRules struct {
	Cadence                int              `json:"cadence"`
	FirstLevel             int              `json:"first_level"`
	FloorSize              DungeonFloorSize `json:"floor_size"`
	MonsterCount           int              `json:"monster_count"`
	ChestInteractableDefID string           `json:"chest_interactable_def_id"`
	ChestLootTable         string           `json:"chest_loot_table"`
	BossTemplatePool       []string         `json:"boss_template_pool"`
	BossSpawn              Vec2             `json:"boss_spawn"`
	ChestPosition          Vec2             `json:"chest_position"`
	StairsUpPosition       Vec2             `json:"stairs_up_position"`
	StairsDownPosition     Vec2             `json:"stairs_down_position"`
	TeleporterPosition     Vec2             `json:"teleporter_position"`
	LockedExitReason       string           `json:"locked_exit_reason"`
}

type StairPlacementRules struct {
	MinSeparation  float64 `json:"min_separation"`
	MarginFromWall float64 `json:"margin_from_wall"`
	MaxAttempts    int     `json:"max_attempts"`
}

type TeleporterPlacementRules struct {
	MarginFromWall   float64 `json:"margin_from_wall"`
	MinStairDistance float64 `json:"min_stair_distance"`
	MaxAttempts      int     `json:"max_attempts"`
}

type MonsterPlacementRules struct {
	Count             int                   `json:"-"`
	MonsterDefID      string                `json:"monster_def_id"`
	PopulationFormula AreaCountFormula      `json:"population_formula"`
	PackCount         IntRange              `json:"-"`
	PackCountFormula  AreaRangeFormula      `json:"pack_count_formula"`
	PackSize          IntRange              `json:"pack_size"`
	PackMemberRadius  float64               `json:"pack_member_radius"`
	PackComposition   PackCompositionRules  `json:"pack_composition"`
	ElitePackChance   int                   `json:"elite_pack_chance_percent"`
	EliteAura         *EliteAuraRules       `json:"elite_aura,omitempty"`
	MonsterPool       []MonsterPoolEntry    `json:"monster_pool,omitempty"`
	MinimumMonsters   []MinimumMonsterEntry `json:"minimum_monsters,omitempty"`
	MarginFromWall    float64               `json:"margin_from_wall"`
	MinSpawnDistance  float64               `json:"min_spawn_distance"`
	MaxAttempts       int                   `json:"max_attempts"`
}

type EliteAuraRules struct {
	ID                 string  `json:"id"`
	Radius             float64 `json:"radius"`
	DamageBonusPercent int     `json:"damage_bonus_percent"`
}

type PackCompositionRules struct {
	FrontlineMin int `json:"frontline_min"`
	RangedMax    int `json:"ranged_max"`
}

type MonsterPoolEntry struct {
	MonsterDefID string `json:"monster_def_id"`
	Weight       int    `json:"weight"`
}

type MinimumMonsterEntry struct {
	MonsterDefID string `json:"monster_def_id"`
	Count        int    `json:"count"`
}

type ChestPlacementRules struct {
	Enabled           bool    `json:"enabled"`
	ChanceWeight      int     `json:"chance_weight"`
	NoChestWeight     int     `json:"no_chest_weight"`
	InteractableDefID string  `json:"interactable_def_id"`
	LootTable         string  `json:"loot_table"`
	MonsterCountBonus int     `json:"monster_count_bonus"`
	MinStairDistance  float64 `json:"min_stair_distance"`
	MaxAttempts       int     `json:"max_attempts"`
}

type ObstacleGenerationRules struct {
	Enabled                 bool                        `json:"enabled"`
	MaxAttempts             int                         `json:"max_attempts"`
	TargetGroupCount        IntRange                    `json:"-"`
	TargetGroupCountFormula AreaRangeFormula            `json:"target_group_count_formula"`
	WallSegment             WallSegmentRules            `json:"wall_segment"`
	SolidBlock              SolidBlockRules             `json:"solid_block"`
	ShapeWeights            ObstacleShapeWeights        `json:"shape_weights"`
	SolidKindWeights        SolidObstacleKindWeights    `json:"solid_kind_weights"`
	Doors                   DoorGenerationRules         `json:"doors"`
	Water                   FloorFeatureGenerationRules `json:"water"`
	Holes                   FloorFeatureGenerationRules `json:"holes"`
	Clearance               ObstacleClearanceRules      `json:"clearance"`
}

type IntRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type WallSegmentRules struct {
	MinLength int     `json:"min_length"`
	MaxLength int     `json:"max_length"`
	Thickness float64 `json:"thickness"`
}

type SolidBlockRules struct {
	MinSize Vec2 `json:"min_size"`
	MaxSize Vec2 `json:"max_size"`
}

type ObstacleShapeWeights struct {
	Line  int `json:"line"`
	L     int `json:"l"`
	T     int `json:"t"`
	Block int `json:"block"`
}

func (w ObstacleShapeWeights) total() int {
	return w.Line + w.L + w.T + w.Block
}

type ObstacleClearanceRules struct {
	PlayerSpawn float64 `json:"player_spawn"`
	Stairs      float64 `json:"stairs"`
	Teleporter  float64 `json:"teleporter"`
	Chest       float64 `json:"chest"`
	Monster     float64 `json:"monster"`
	Loot        float64 `json:"loot"`
}

type DungeonLootBand struct {
	MinDepth         int    `json:"min_depth"`
	MaxDepth         *int   `json:"max_depth"`
	MonsterLootTable string `json:"monster_loot_table"`
	ChestLootTable   string `json:"chest_loot_table"`
}

type MonsterRarityDef struct {
	ID                       string  `json:"id"`
	Weight                   int     `json:"weight"`
	Color                    string  `json:"color"`
	HPMultiplier             float64 `json:"hp_multiplier"`
	DamageMultiplier         float64 `json:"damage_multiplier"`
	XPMultiplier             float64 `json:"xp_multiplier"`
	ArmorMultiplier          float64 `json:"armor_multiplier"`
	ArmorBonus               float64 `json:"armor_bonus"`
	HitChanceBonus           float64 `json:"hit_chance_bonus"`
	CritChanceBonus          float64 `json:"crit_chance_bonus"`
	BlockPercentBonus        float64 `json:"block_percent_bonus"`
	AttackCooldownMultiplier float64 `json:"attack_cooldown_multiplier"`
	LootDepthOffset          int     `json:"loot_depth_offset"`
	VisualScale              float64 `json:"visual_scale"`
}

type MonsterDepthScalingRules struct {
	HPPerDepth                       float64 `json:"hp_per_depth"`
	DamagePerDepth                   float64 `json:"damage_per_depth"`
	ArmorPerDepth                    float64 `json:"armor_per_depth"`
	HitChancePerDepth                float64 `json:"hit_chance_per_depth"`
	MaxHitChance                     float64 `json:"max_hit_chance"`
	CritChancePerDepth               float64 `json:"crit_chance_per_depth"`
	MaxCritChance                    float64 `json:"max_crit_chance"`
	BlockPercentPerDepth             float64 `json:"block_percent_per_depth"`
	MaxBlockPercent                  float64 `json:"max_block_percent"`
	AttackCooldownMultiplierPerDepth float64 `json:"attack_cooldown_multiplier_per_depth"`
	MinAttackCooldownTicks           int     `json:"min_attack_cooldown_ticks"`
}

func (d DungeonGenerationRules) LootBandForLevel(levelNum int) (DungeonLootBand, bool) {
	if levelNum >= 0 {
		return DungeonLootBand{}, false
	}
	depth := absInt(levelNum)
	for _, band := range d.LootBands {
		if depth < band.MinDepth {
			continue
		}
		if band.MaxDepth != nil && depth > *band.MaxDepth {
			continue
		}
		return band, true
	}
	return DungeonLootBand{}, false
}

func (d DungeonGenerationRules) LootBandForDepth(depth int) (DungeonLootBand, bool) {
	if depth <= 0 {
		return DungeonLootBand{}, false
	}
	for _, band := range d.LootBands {
		if depth < band.MinDepth {
			continue
		}
		if band.MaxDepth != nil && depth > *band.MaxDepth {
			continue
		}
		return band, true
	}
	return DungeonLootBand{}, false
}

func (d DungeonGenerationRules) MonsterRarity(id string) (MonsterRarityDef, bool) {
	for _, rarity := range d.MonsterRarities {
		if rarity.ID == id {
			return rarity, true
		}
	}
	return MonsterRarityDef{}, false
}

func (d DungeonGenerationRules) MonsterRole(monsterID string) string {
	if d.monsterPackRoles == nil {
		return ""
	}
	return d.monsterPackRoles[monsterID]
}

func (d DungeonGenerationRules) RollMonsterRarity(rng *RNG) MonsterRarityDef {
	total := 0
	for _, rarity := range d.MonsterRarities {
		total += rarity.Weight
	}
	if total <= 0 {
		return MonsterRarityDef{ID: "common", Weight: 1, HPMultiplier: 1, DamageMultiplier: 1, XPMultiplier: 1}
	}
	roll := rng.IntN(total)
	for _, rarity := range d.MonsterRarities {
		roll -= rarity.Weight
		if roll < 0 {
			return rarity
		}
	}
	return d.MonsterRarities[len(d.MonsterRarities)-1]
}
