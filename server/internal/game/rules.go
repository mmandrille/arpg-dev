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
	Combat     Combat
	Items      map[string]ItemDef
	Monsters   map[string]MonsterDef
	LootTables map[string]LootTable
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
}

// ItemDef is a single item definition.
type ItemDef struct {
	Name       string `json:"name"`
	Slot       string `json:"slot"`
	Equippable bool   `json:"equippable"`
}

// MonsterDef is a single monster definition.
type MonsterDef struct {
	Name      string `json:"name"`
	MaxHP     int    `json:"max_hp"`
	LootTable string `json:"loot_table"`
}

// LootEntry is one weighted entry in a loot table.
type LootEntry struct {
	ItemDefID string `json:"item_def_id"`
	Weight    int    `json:"weight"`
}

// LootTable is a weighted set of loot entries.
type LootTable struct {
	Entries []LootEntry `json:"entries"`
}

// LoadRules reads and parses the v0 rules files from a directory.
func LoadRules(dir string) (*Rules, error) {
	r := &Rules{}

	var combat struct {
		Version       int         `json:"version"`
		BaseHitChance float64     `json:"base_hit_chance"`
		PlayerDamage  DamageRange `json:"player_damage"`
	}
	if err := readJSON(filepath.Join(dir, "combat.v0.json"), &combat); err != nil {
		return nil, err
	}
	r.Combat = Combat{BaseHitChance: combat.BaseHitChance, PlayerDamage: combat.PlayerDamage}

	var items struct {
		Items map[string]ItemDef `json:"items"`
	}
	if err := readJSON(filepath.Join(dir, "items.v0.json"), &items); err != nil {
		return nil, err
	}
	r.Items = items.Items

	var monsters struct {
		Monsters map[string]MonsterDef `json:"monsters"`
	}
	if err := readJSON(filepath.Join(dir, "monsters.v0.json"), &monsters); err != nil {
		return nil, err
	}
	r.Monsters = monsters.Monsters

	var loot struct {
		LootTables map[string]LootTable `json:"loot_tables"`
	}
	if err := readJSON(filepath.Join(dir, "loot_tables.v0.json"), &loot); err != nil {
		return nil, err
	}
	r.LootTables = loot.LootTables

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
