package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

// Rules is the in-memory form of the shared rules-as-data (shared/rules). The
// Go server and the Godot client read the same files (ADR-0001 D6); this is the
// server's loader and typed view.
type Rules struct {
	Combat               Combat
	Navigation           NavigationRules
	Items                map[string]ItemDef
	ItemTemplates        map[string]ItemTemplateDef
	Rarities             map[string]RarityDef
	RarityOrder          []string
	TreasureClasses      map[string]TreasureClassDef
	CharacterProgression CharacterProgressionRules
	Monsters             map[string]MonsterDef
	LootTables           map[string]LootTable
	Interactables        map[string]InteractableDef
	Worlds               map[string]WorldDef
	DungeonGeneration    DungeonGenerationRules
}

// DamageRange is an inclusive [Min, Max] integer range.
type DamageRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// Combat holds combat parameters.
type Combat struct {
	BaseHitChance float64     `json:"base_hit_chance"`
	PlayerDamage  DamageRange `json:"player_damage"`
	UnarmedReach  float64     `json:"unarmed_reach"`
}

// NavigationRules bounds server-owned auto-navigation.
type NavigationRules struct {
	CellSize     float64    `json:"cell_size"`
	MaxAutoSteps int        `json:"max_auto_steps"`
	GridBounds   GridBounds `json:"grid_bounds"`
	StopDistance float64    `json:"stop_distance"`
}

// DungeonGenerationRules controls deterministic generated dungeon floors.
type DungeonGenerationRules struct {
	FloorSize                DungeonFloorSize         `json:"floor_size"`
	WallThickness            float64                  `json:"wall_thickness"`
	PlayerSpawn              Vec2                     `json:"player_spawn"`
	StairPlacement           StairPlacementRules      `json:"stair_placement"`
	TeleporterPlacement      TeleporterPlacementRules `json:"teleporter_placement"`
	MonsterPlacement         MonsterPlacementRules    `json:"monster_placement"`
	ChestPlacement           ChestPlacementRules      `json:"chest_placement"`
	LevelNames               map[string]string        `json:"level_names"`
	DefaultLevelNameTemplate string                   `json:"default_level_name_template"`
}

// CharacterProgressionRules controls XP thresholds, level-up points, base
// stats, and derived-stat formulas.
type CharacterProgressionRules struct {
	BaseStats      BaseStatsView
	PointsPerLevel int
	LevelCap       int
	XPThresholds   map[int]int
	DerivedStats   map[string]LinearStatFormula
}

type LinearStatFormula struct {
	Type     string   `json:"type"`
	Base     float64  `json:"base"`
	PerStr   float64  `json:"per_str"`
	PerDex   float64  `json:"per_dex"`
	PerVit   float64  `json:"per_vit"`
	PerMagic float64  `json:"per_magic"`
	Min      *float64 `json:"min"`
	Max      *float64 `json:"max"`
}

type DungeonFloorSize struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
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
	Count            int     `json:"count"`
	MonsterDefID     string  `json:"monster_def_id"`
	MarginFromWall   float64 `json:"margin_from_wall"`
	MinSpawnDistance float64 `json:"min_spawn_distance"`
	MaxAttempts      int     `json:"max_attempts"`
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
	Slot            string       `json:"slot"`
	Equippable      bool         `json:"equippable"`
	AttackMode      string       `json:"attack_mode,omitempty"`
	Damage          *DamageRange `json:"damage,omitempty"`
	Reach           *float64     `json:"reach,omitempty"`
	ProjectileSpeed *float64     `json:"projectile_speed,omitempty"`
	Heal            *DamageRange `json:"heal,omitempty"`
}

// RarityDef controls how many bounded stat rolls a rolled item gets.
type RarityDef struct {
	Weight     int    `json:"weight"`
	StatRolls  int    `json:"stat_rolls"`
	NamePrefix string `json:"name_prefix"`
}

// ItemTemplateDef is a server-authoritative rolled item template.
type ItemTemplateDef struct {
	Name          string            `json:"name"`
	Category      string            `json:"category"`
	ItemType      string            `json:"item_type"`
	Slot          string            `json:"slot"`
	Equippable    bool              `json:"equippable"`
	AttackMode    string            `json:"attack_mode,omitempty"`
	Reach         float64           `json:"reach"`
	Requirements  map[string]int    `json:"requirements"`
	BaseStats     map[string]int    `json:"base_stats"`
	RollableStats []RollableStatDef `json:"rollable_stats"`
	EffectPool    []string          `json:"effect_pool"`
}

// RollableStatDef is one weighted bounded stat increment.
type RollableStatDef struct {
	Stat   string `json:"stat"`
	Min    int    `json:"min"`
	Max    int    `json:"max"`
	Weight int    `json:"weight"`
}

// InteractableDef is a single activatable world object definition.
type InteractableDef struct {
	Name              string               `json:"name"`
	InitialState      string               `json:"initial_state"`
	Transition        string               `json:"transition,omitempty"`
	BarrierWhenClosed *InteractableBarrier `json:"barrier_when_closed,omitempty"`
}

// InteractableBarrier is the closed-state movement blocker for an interactable.
type InteractableBarrier struct {
	Size Vec2 `json:"size"`
}

// MonsterDef is a single monster definition.
type MonsterDef struct {
	Name              string       `json:"name"`
	MaxHP             int          `json:"max_hp"`
	LootTable         string       `json:"loot_table"`
	RetaliationDamage *DamageRange `json:"retaliation_damage,omitempty"`
	AttackDamage      *DamageRange `json:"attack_damage,omitempty"`
	AttackCooldown    int          `json:"attack_cooldown_ticks,omitempty"`
	Behavior          string       `json:"behavior,omitempty"`
	AggroRadius       float64      `json:"aggro_radius,omitempty"`
	LeashRadius       float64      `json:"leash_radius,omitempty"`
	MoveSpeed         float64      `json:"move_speed,omitempty"`
	XPReward          int          `json:"xp_reward,omitempty"`
}

func (d MonsterDef) effectiveBehavior() string {
	if d.Behavior == "" {
		return monsterBehaviorStatic
	}

	return d.Behavior
}

func (d MonsterDef) effectiveMoveSpeed(nav NavigationRules) float64 {
	if d.MoveSpeed > 0 {
		return d.MoveSpeed
	}

	return nav.CellSize
}

// LootEntry is one weighted entry in a loot table.
type LootEntry struct {
	ItemDefID      string `json:"item_def_id"`
	ItemTemplateID string `json:"item_template_id"`
	Weight         int    `json:"weight"`
}

type TreasureClassEntry struct {
	ItemDefID      string `json:"item_def_id"`
	ItemTemplateID string `json:"item_template_id"`
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

// WorldDef is a deterministic initial session layout.
type WorldDef struct {
	Mode     string        `json:"mode,omitempty"`
	Player   WorldPlayer   `json:"player"`
	Entities []WorldEntity `json:"entities"`
}

// WorldPlayer is the initial player placement for a world.
type WorldPlayer struct {
	Position Vec2 `json:"position"`
}

// WorldEntity is an initial non-player entity in a world.
type WorldEntity struct {
	Type              string `json:"type"`
	MonsterDefID      string `json:"monster_def_id,omitempty"`
	ItemDefID         string `json:"item_def_id,omitempty"`
	InteractableDefID string `json:"interactable_def_id,omitempty"`
	Position          Vec2   `json:"position"`
	Size              Vec2   `json:"size,omitempty"`
}

// LoadRules reads and parses the v0 rules files from a directory.
func LoadRules(dir string) (*Rules, error) {
	r := &Rules{}

	var combat struct {
		Version       int         `json:"version"`
		BaseHitChance float64     `json:"base_hit_chance"`
		PlayerDamage  DamageRange `json:"player_damage"`
		UnarmedReach  float64     `json:"unarmed_reach"`
	}
	if err := readJSON(filepath.Join(dir, "combat.v0.json"), &combat); err != nil {
		return nil, err
	}
	if err := validateDamageRange("combat.player_damage", combat.PlayerDamage); err != nil {
		return nil, err
	}
	if combat.UnarmedReach <= 0 {
		return nil, fmt.Errorf("game: invalid rules combat.unarmed_reach: must be positive")
	}
	r.Combat = Combat{BaseHitChance: combat.BaseHitChance, PlayerDamage: combat.PlayerDamage, UnarmedReach: combat.UnarmedReach}

	var navigation struct {
		Version      int        `json:"version"`
		CellSize     float64    `json:"cell_size"`
		MaxAutoSteps int        `json:"max_auto_steps"`
		GridBounds   GridBounds `json:"grid_bounds"`
		StopDistance float64    `json:"stop_distance"`
	}
	if err := readJSON(filepath.Join(dir, "navigation.v0.json"), &navigation); err != nil {
		return nil, err
	}
	if navigation.CellSize <= 0 {
		return nil, fmt.Errorf("game: invalid rules navigation.cell_size: must be positive")
	}
	if navigation.CellSize != moveSpeed {
		return nil, fmt.Errorf("game: invalid rules navigation.cell_size: must equal moveSpeed %.1f for v11", moveSpeed)
	}
	if navigation.MaxAutoSteps <= 0 {
		return nil, fmt.Errorf("game: invalid rules navigation.max_auto_steps: must be positive")
	}
	if navigation.GridBounds.MaxX < navigation.GridBounds.MinX || navigation.GridBounds.MaxY < navigation.GridBounds.MinY {
		return nil, fmt.Errorf("game: invalid rules navigation.grid_bounds: max must be >= min")
	}
	if navigation.StopDistance < 0 {
		return nil, fmt.Errorf("game: invalid rules navigation.stop_distance: must be non-negative")
	}
	r.Navigation = NavigationRules{
		CellSize:     navigation.CellSize,
		MaxAutoSteps: navigation.MaxAutoSteps,
		GridBounds:   navigation.GridBounds,
		StopDistance: navigation.StopDistance,
	}

	var progression struct {
		Version         int           `json:"version"`
		BaseStats       BaseStatsView `json:"base_stats"`
		PointsPerLevel  int           `json:"points_per_level"`
		LevelCap        int           `json:"level_cap"`
		ExperienceCurve struct {
			Type   string `json:"type"`
			Levels []struct {
				Level            int `json:"level"`
				NextLevelTotalXP int `json:"next_level_total_xp"`
			} `json:"levels"`
		} `json:"experience_curve"`
		DerivedStats map[string]LinearStatFormula `json:"derived_stats"`
	}
	if err := readJSON(filepath.Join(dir, "character_progression.v0.json"), &progression); err != nil {
		return nil, err
	}
	if progression.BaseStats.Str <= 0 || progression.BaseStats.Dex <= 0 || progression.BaseStats.Vit <= 0 || progression.BaseStats.Magic <= 0 {
		return nil, fmt.Errorf("game: invalid rules character_progression.base_stats: all stats must be positive")
	}
	if progression.PointsPerLevel <= 0 {
		return nil, fmt.Errorf("game: invalid rules character_progression.points_per_level: must be positive")
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
	requiredDerived := []string{"damage_min", "damage_max", "armor", "attack_speed", "hit_chance", "crit_chance", "crit_damage", "movement_speed", "max_hp", "max_mana"}
	for _, key := range requiredDerived {
		formula, ok := progression.DerivedStats[key]
		if !ok {
			return nil, fmt.Errorf("game: invalid rules character_progression.derived_stats: missing %s", key)
		}
		if formula.Type != "linear" {
			return nil, fmt.Errorf("game: invalid rules character_progression.derived_stats.%s.type: %s", key, formula.Type)
		}
		if formula.Min != nil && formula.Max != nil && *formula.Max < *formula.Min {
			return nil, fmt.Errorf("game: invalid rules character_progression.derived_stats.%s: max must be >= min", key)
		}
	}
	r.CharacterProgression = CharacterProgressionRules{
		BaseStats:      progression.BaseStats,
		PointsPerLevel: progression.PointsPerLevel,
		LevelCap:       progression.LevelCap,
		XPThresholds:   thresholds,
		DerivedStats:   progression.DerivedStats,
	}

	var items struct {
		Items map[string]ItemDef `json:"items"`
	}
	if err := readJSON(filepath.Join(dir, "items.v0.json"), &items); err != nil {
		return nil, err
	}
	for id, def := range items.Items {
		if def.Equippable && def.Slot == "" {
			return nil, fmt.Errorf("game: invalid rules items.%s: equippable item must declare slot", id)
		}
		if !def.Equippable && def.Slot != "" {
			return nil, fmt.Errorf("game: invalid rules items.%s: non-equippable item must not declare slot", id)
		}
		if def.Damage != nil {
			if !def.Equippable || def.Slot != weaponSlot {
				return nil, fmt.Errorf("game: invalid rules items.%s.damage: damage is only valid on equippable weapons", id)
			}
			if err := validateDamageRange("items."+id+".damage", *def.Damage); err != nil {
				return nil, err
			}
		}
		if def.Reach != nil {
			if !def.Equippable || def.Slot != weaponSlot {
				return nil, fmt.Errorf("game: invalid rules items.%s.reach: reach is only valid on equippable weapons", id)
			}
			if *def.Reach <= 0 {
				return nil, fmt.Errorf("game: invalid rules items.%s.reach: must be positive", id)
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
			if !def.Equippable || def.Slot != weaponSlot || def.Damage == nil || def.Reach == nil || def.ProjectileSpeed == nil {
				return nil, fmt.Errorf("game: invalid rules items.%s: ranged weapons require slot, damage, reach, and projectile_speed", id)
			}
			if *def.ProjectileSpeed <= 0 {
				return nil, fmt.Errorf("game: invalid rules items.%s.projectile_speed: must be positive", id)
			}
		default:
			return nil, fmt.Errorf("game: invalid rules items.%s.attack_mode: %s", id, def.AttackMode)
		}
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
		if rarity.StatRolls <= 0 {
			return nil, fmt.Errorf("game: invalid rules item_templates.rarities.%s.stat_rolls: must be positive", id)
		}
		if rarity.NamePrefix == "" {
			return nil, fmt.Errorf("game: invalid rules item_templates.rarities.%s.name_prefix: required", id)
		}
	}
	for id, def := range itemTemplates.Templates {
		if !def.Equippable || def.Slot != weaponSlot || def.Category != "equipment" {
			return nil, fmt.Errorf("game: invalid rules item_templates.%s: v23 templates must be equippable equipment weapons", id)
		}
		if def.AttackMode == "" {
			def.AttackMode = attackModeMelee
		}
		if def.AttackMode != attackModeMelee {
			return nil, fmt.Errorf("game: invalid rules item_templates.%s.attack_mode: v23 supports melee only", id)
		}
		if def.Reach <= 0 {
			return nil, fmt.Errorf("game: invalid rules item_templates.%s.reach: must be positive", id)
		}
		if def.Requirements["level"] > 1 {
			return nil, fmt.Errorf("game: invalid rules item_templates.%s.requirements.level: v23 supports level <= 1", id)
		}
		min, max := def.BaseStats["damage_min"], def.BaseStats["damage_max"]
		if max < min {
			return nil, fmt.Errorf("game: invalid rules item_templates.%s.base_stats: damage_max must be >= damage_min", id)
		}
		seen := map[string]bool{}
		for _, roll := range def.RollableStats {
			switch roll.Stat {
			case "damage_min", "damage_max", "max_hp":
			default:
				return nil, fmt.Errorf("game: invalid rules item_templates.%s.rollable_stats: unsupported stat %s", id, roll.Stat)
			}
			if roll.Max < roll.Min {
				return nil, fmt.Errorf("game: invalid rules item_templates.%s.rollable_stats.%s: max must be >= min", id, roll.Stat)
			}
			if roll.Weight <= 0 {
				return nil, fmt.Errorf("game: invalid rules item_templates.%s.rollable_stats.%s: weight must be positive", id, roll.Stat)
			}
			seen[roll.Stat] = true
		}
		if !seen["damage_min"] || !seen["damage_max"] {
			return nil, fmt.Errorf("game: invalid rules item_templates.%s.rollable_stats: damage_min and damage_max are required", id)
		}
		itemTemplates.Templates[id] = def
	}
	r.ItemTemplates = itemTemplates.Templates
	r.Rarities = itemTemplates.Rarities
	r.RarityOrder = rarityOrder

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
				if (entry.ItemDefID == "") == (entry.ItemTemplateID == "") {
					return nil, fmt.Errorf("game: invalid rules treasure_classes.%s.attempts.%s: entry must declare exactly one of item_def_id or item_template_id", classID, attempt.AttemptID)
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
		behavior := def.effectiveBehavior()
		switch behavior {
		case monsterBehaviorStatic:
			if def.AggroRadius > 0 || def.LeashRadius > 0 || def.MoveSpeed > 0 {
				return nil, fmt.Errorf("game: invalid rules monsters.%s: aggro/leash/move_speed only valid for chase behavior", id)
			}
			if def.AttackDamage != nil {
				return nil, fmt.Errorf("game: invalid rules monsters.%s: attack_damage only valid for chase behavior", id)
			}
		case monsterBehaviorChase:
			if def.AggroRadius <= 0 {
				return nil, fmt.Errorf("game: invalid rules monsters.%s: chase requires positive aggro_radius", id)
			}
			if def.LeashRadius > 0 && def.LeashRadius < def.AggroRadius {
				return nil, fmt.Errorf("game: invalid rules monsters.%s: leash_radius must be >= aggro_radius", id)
			}
			if def.MoveSpeed > 0 && def.MoveSpeed != r.Navigation.CellSize {
				return nil, fmt.Errorf("game: invalid rules monsters.%s: move_speed must equal navigation.cell_size %.1f in v17", id, r.Navigation.CellSize)
			}
		default:
			return nil, fmt.Errorf("game: invalid rules monsters.%s.behavior: %s", id, def.Behavior)
		}
	}
	r.Monsters = monsters.Monsters

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
			if (entry.ItemDefID == "") == (entry.ItemTemplateID == "") {
				return nil, fmt.Errorf("game: invalid rules loot_tables.%s: entry must declare exactly one of item_def_id or item_template_id", tableID)
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
			if def.BarrierWhenClosed != nil && (def.BarrierWhenClosed.Size.X <= 0 || def.BarrierWhenClosed.Size.Y <= 0) {
				return nil, fmt.Errorf("game: invalid rules interactables.%s.barrier_when_closed.size: must be positive", id)
			}
		case interactableReady:
			if def.BarrierWhenClosed != nil {
				return nil, fmt.Errorf("game: invalid rules interactables.%s.barrier_when_closed: ready interactable must not declare barrier", id)
			}
			switch def.Transition {
			case interactableTransitionAscend, interactableTransitionDescend, interactableTransitionWaypoint:
			default:
				return nil, fmt.Errorf("game: invalid rules interactables.%s.transition: must be ascend, descend, or waypoint", id)
			}
		default:
			return nil, fmt.Errorf("game: invalid rules interactables.%s.initial_state: unsupported state %s", id, def.InitialState)
		}
	}
	r.Interactables = interactables.Interactables

	var dungeonGeneration struct {
		Version                  int                      `json:"version"`
		FloorSize                DungeonFloorSize         `json:"floor_size"`
		WallThickness            float64                  `json:"wall_thickness"`
		PlayerSpawn              Vec2                     `json:"player_spawn"`
		StairPlacement           StairPlacementRules      `json:"stair_placement"`
		TeleporterPlacement      TeleporterPlacementRules `json:"teleporter_placement"`
		MonsterPlacement         MonsterPlacementRules    `json:"monster_placement"`
		ChestPlacement           ChestPlacementRules      `json:"chest_placement"`
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
	if dungeonGeneration.MonsterPlacement.Count < 0 {
		return nil, fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.count: must be non-negative")
	}
	if dungeonGeneration.MonsterPlacement.Count > 0 {
		monsterID := dungeonGeneration.MonsterPlacement.MonsterDefID
		monsterDef, ok := r.Monsters[monsterID]
		if !ok {
			return nil, fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.monster_def_id: unknown monster %s", monsterID)
		}
		if monsterDef.effectiveBehavior() != monsterBehaviorChase {
			return nil, fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.monster_def_id: %s must use chase behavior", monsterID)
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
	for key := range dungeonGeneration.LevelNames {
		level, err := strconv.Atoi(key)
		if err != nil || level >= 0 {
			return nil, fmt.Errorf("game: invalid rules dungeon_generation.level_names.%s: key must be a negative integer string", key)
		}
	}
	r.DungeonGeneration = DungeonGenerationRules{
		FloorSize:                dungeonGeneration.FloorSize,
		WallThickness:            dungeonGeneration.WallThickness,
		PlayerSpawn:              dungeonGeneration.PlayerSpawn,
		StairPlacement:           dungeonGeneration.StairPlacement,
		TeleporterPlacement:      dungeonGeneration.TeleporterPlacement,
		MonsterPlacement:         dungeonGeneration.MonsterPlacement,
		ChestPlacement:           dungeonGeneration.ChestPlacement,
		LevelNames:               dungeonGeneration.LevelNames,
		DefaultLevelNameTemplate: dungeonGeneration.DefaultLevelNameTemplate,
	}

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
		for i, entity := range world.Entities {
			label := fmt.Sprintf("worlds.%s.entities[%d]", worldID, i)
			switch entity.Type {
			case monsterEntity:
				if entity.MonsterDefID == "" {
					return nil, fmt.Errorf("game: invalid rules %s: missing monster_def_id", label)
				}
				if _, ok := r.Monsters[entity.MonsterDefID]; !ok {
					return nil, fmt.Errorf("game: invalid rules %s: unknown monster %s", label, entity.MonsterDefID)
				}
			case lootEntity:
				if entity.ItemDefID == "" {
					return nil, fmt.Errorf("game: invalid rules %s: missing item_def_id", label)
				}
				if _, ok := r.Items[entity.ItemDefID]; !ok {
					return nil, fmt.Errorf("game: invalid rules %s: unknown item %s", label, entity.ItemDefID)
				}
			case wallEntity:
				if entity.Size.X <= 0 || entity.Size.Y <= 0 {
					return nil, fmt.Errorf("game: invalid rules %s: wall size must be positive", label)
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

	return r, nil
}

// LootDrop is one resolved loot table result.
type LootDrop struct {
	ItemDefID      string `json:"item_def_id,omitempty"`
	ItemTemplateID string `json:"item_template_id,omitempty"`
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
			return LootDrop{ItemDefID: e.ItemDefID, ItemTemplateID: e.ItemTemplateID}, true
		}
	}
	last := table.Entries[len(table.Entries)-1]
	return LootDrop{ItemDefID: last.ItemDefID, ItemTemplateID: last.ItemTemplateID}, true
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
				out = append(out, LootDrop{ItemDefID: entry.ItemDefID, ItemTemplateID: entry.ItemTemplateID})
				break
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
