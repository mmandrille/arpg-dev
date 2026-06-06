package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Rules is the in-memory form of the shared rules-as-data (shared/rules). The
// Go server and the Godot client read the same files (ADR-0001 D6); this is the
// server's loader and typed view.
type Rules struct {
	Combat        Combat
	Navigation    NavigationRules
	Items         map[string]ItemDef
	Monsters      map[string]MonsterDef
	LootTables    map[string]LootTable
	Interactables map[string]InteractableDef
	Worlds        map[string]WorldDef
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

// InteractableDef is a single activatable world object definition.
type InteractableDef struct {
	Name              string              `json:"name"`
	InitialState      string              `json:"initial_state"`
	BarrierWhenClosed InteractableBarrier `json:"barrier_when_closed"`
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
}

// LootEntry is one weighted entry in a loot table.
type LootEntry struct {
	ItemDefID string `json:"item_def_id"`
	Weight    int    `json:"weight"`
}

// LootTable is a weighted set of loot entries.
type LootTable struct {
	Drops   []string    `json:"drops,omitempty"`
	Entries []LootEntry `json:"entries"`
}

// WorldDef is a deterministic initial session layout.
type WorldDef struct {
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

	var monsters struct {
		Monsters map[string]MonsterDef `json:"monsters"`
	}
	if err := readJSON(filepath.Join(dir, "monsters.v0.json"), &monsters); err != nil {
		return nil, err
	}
	for id, def := range monsters.Monsters {
		if def.RetaliationDamage != nil {
			if err := validateDamageRange("monsters."+id+".retaliation_damage", *def.RetaliationDamage); err != nil {
				return nil, err
			}
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
		for _, entry := range table.Entries {
			if _, ok := r.Items[entry.ItemDefID]; !ok {
				return nil, fmt.Errorf("game: invalid rules loot_tables.%s: unknown item %s", tableID, entry.ItemDefID)
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
		if def.InitialState != interactableClosed {
			return nil, fmt.Errorf("game: invalid rules interactables.%s.initial_state: must be closed", id)
		}
		if def.BarrierWhenClosed.Size.X <= 0 || def.BarrierWhenClosed.Size.Y <= 0 {
			return nil, fmt.Errorf("game: invalid rules interactables.%s.barrier_when_closed.size: must be positive", id)
		}
	}
	r.Interactables = interactables.Interactables

	var worlds struct {
		Worlds map[string]WorldDef `json:"worlds"`
	}
	if err := readJSON(filepath.Join(dir, "worlds.v0.json"), &worlds); err != nil {
		return nil, err
	}
	for worldID, world := range worlds.Worlds {
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

// RollLoot selects an item_def_id from a loot table using the RNG. A
// single-entry table is deterministic regardless of the draw.
func (r *Rules) RollLoot(tableID string, rng *RNG) (string, bool) {
	table, ok := r.LootTables[tableID]
	if !ok || len(table.Entries) == 0 {
		return "", false
	}
	total := 0
	for _, e := range table.Entries {
		total += e.Weight
	}
	if total <= 0 {
		return "", false
	}
	roll := rng.IntN(total)
	for _, e := range table.Entries {
		roll -= e.Weight
		if roll < 0 {
			return e.ItemDefID, true
		}
	}
	return table.Entries[len(table.Entries)-1].ItemDefID, true
}

// LootDrops returns all guaranteed drops for a table, or one weighted roll for
// legacy single-drop tables.
func (r *Rules) LootDrops(tableID string, rng *RNG) []string {
	table, ok := r.LootTables[tableID]
	if !ok {
		return nil
	}
	if len(table.Drops) > 0 {
		out := make([]string, len(table.Drops))
		copy(out, table.Drops)
		return out
	}
	itemDefID, ok := r.RollLoot(tableID, rng)
	if !ok {
		return nil
	}
	return []string{itemDefID}
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
