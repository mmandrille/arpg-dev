package game

import (
	"fmt"
	"math"
)

// PropGenerationRules tunes server-placed room props: small, blocking clutter (barrels, crates, tables)
// that the client renders from the wall layout. All values are data (obstacle_generation.props).
type PropGenerationRules struct {
	Enabled       bool               `json:"enabled"`
	MaxAttempts   int                `json:"max_attempts"`
	WallClearance float64            `json:"wall_clearance"`
	DoorClearance float64            `json:"door_clearance"`
	Spacing       float64            `json:"spacing"`
	CountBands    []PropCountBand    `json:"count_bands"`
	Catalog       []PropCatalogEntry `json:"catalog"`
}

// PropCountBand sets how many props a floor gets at a depth range: one per AreaPerProp square units of
// room area, capped at MaxCount. MaxDepth nil means no upper depth.
type PropCountBand struct {
	MinDepth    int     `json:"min_depth"`
	MaxDepth    *int    `json:"max_depth"`
	AreaPerProp float64 `json:"area_per_prop"`
	MaxCount    int     `json:"max_count"`
}

// PropCatalogEntry is one placeable prop: a stable id shared with the client presentation catalog, a
// selection weight, and its blocking footprint (axis-aligned, centered on the prop).
type PropCatalogEntry struct {
	PropID    string `json:"prop_id"`
	Weight    int    `json:"weight"`
	Footprint Vec2   `json:"footprint"`
}

const maxPropFootprint = 3.0

func validatePropGenerationRules(p PropGenerationRules) error {
	if !p.Enabled {
		return nil
	}
	const base = "game: invalid rules dungeon_generation.obstacle_generation.props"
	if p.MaxAttempts <= 0 {
		return fmt.Errorf("%s.max_attempts: must be positive", base)
	}
	for _, field := range []struct {
		label string
		value float64
	}{{"wall_clearance", p.WallClearance}, {"door_clearance", p.DoorClearance}, {"spacing", p.Spacing}} {
		if field.value < 0 {
			return fmt.Errorf("%s.%s: must be non-negative", base, field.label)
		}
	}
	if len(p.CountBands) == 0 {
		return fmt.Errorf("%s.count_bands: at least one band is required", base)
	}
	for i, band := range p.CountBands {
		if band.MinDepth < 1 || (band.MaxDepth != nil && *band.MaxDepth < band.MinDepth) {
			return fmt.Errorf("%s.count_bands[%d]: invalid depth range", base, i)
		}
		if band.AreaPerProp <= 0 || band.MaxCount < 0 {
			return fmt.Errorf("%s.count_bands[%d]: area_per_prop must be positive and max_count non-negative", base, i)
		}
	}
	if len(p.Catalog) == 0 {
		return fmt.Errorf("%s.catalog: at least one prop is required", base)
	}
	seen := make(map[string]bool, len(p.Catalog))
	totalWeight := 0
	for i, entry := range p.Catalog {
		if entry.PropID == "" || seen[entry.PropID] {
			return fmt.Errorf("%s.catalog[%d].prop_id: must be unique and non-empty", base, i)
		}
		seen[entry.PropID] = true
		if entry.Weight < 0 {
			return fmt.Errorf("%s.catalog[%d].weight: must be non-negative", base, i)
		}
		if entry.Footprint.X <= 0 || entry.Footprint.Y <= 0 || math.Max(entry.Footprint.X, entry.Footprint.Y) > maxPropFootprint {
			return fmt.Errorf("%s.catalog[%d].footprint: each side must be positive and at most %v", base, i, maxPropFootprint)
		}
		totalWeight += entry.Weight
	}
	if totalWeight <= 0 {
		return fmt.Errorf("%s.catalog: at least one weight must be positive", base)
	}
	return nil
}

func (p PropGenerationRules) bandForDepth(depth int) (PropCountBand, bool) {
	for _, band := range p.CountBands {
		if depth >= band.MinDepth && (band.MaxDepth == nil || depth <= *band.MaxDepth) {
			return band, true
		}
	}
	return PropCountBand{}, false
}

// weightedEntry picks a catalog entry by weight using one RNG draw.
func (p PropGenerationRules) weightedEntry(rng *RNG) PropCatalogEntry {
	total := 0
	for _, entry := range p.Catalog {
		total += entry.Weight
	}
	roll := rng.IntN(total)
	for _, entry := range p.Catalog {
		roll -= entry.Weight
		if roll < 0 {
			return entry
		}
	}
	return p.Catalog[len(p.Catalog)-1]
}
