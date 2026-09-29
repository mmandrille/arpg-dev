package game

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// TestDungeonWallGridAudit is a report-only probe for ADR-0018 D6: do generated
// wall rectangles snap to a common tile size, so a client auto-tiler can map them
// onto modular kit pieces without a server change? It is skipped unless
// ARPG_WALL_GRID_AUDIT_OUT names the JSON report path (see `make wall-grid-audit`).
// It asserts nothing about gameplay and changes no generation code or golden.
func TestDungeonWallGridAudit(t *testing.T) {
	outPath := os.Getenv("ARPG_WALL_GRID_AUDIT_OUT")
	if outPath == "" {
		t.Skip("set ARPG_WALL_GRID_AUDIT_OUT to write the wall grid audit report")
	}
	rulesDir := filepath.Join("..", "..", "..", "shared", "rules")
	rules, err := LoadRules(rulesDir)
	if err != nil {
		t.Fatalf("load rules: %v", err)
	}
	gen := rules.DungeonGeneration
	maxDepth := wallAuditMaxDepth(t, rulesDir, gen)
	tiles := wallAuditTiles(t)

	var edges []wallAuditEdge
	var failures []wallAuditFailure
	thickness := map[float64]int{}
	levels := make([]int, 0, maxDepth)
	for depth := 1; depth <= maxDepth; depth++ {
		levels = append(levels, -depth)
	}
	for _, seed := range wallAuditSeeds {
		for _, level := range levels {
			out, err := GenerateDungeonLevel(seed, level, gen)
			if err != nil {
				// Report-only probe: record generator failures instead of aborting the audit.
				failures = append(failures, wallAuditFailure{Seed: seed, Level: level, Error: err.Error()})
				continue
			}
			for _, w := range out.walls {
				source := w.source
				if w.kind != "" {
					source += "/" + w.kind
				}
				hx, hy := w.size.X/2, w.size.Y/2
				for _, v := range []float64{w.pos.X - hx, w.pos.X + hx, w.pos.Y - hy, w.pos.Y + hy} {
					edges = append(edges, wallAuditEdge{source: source, value: v})
				}
				thickness[roundTo(math.Min(w.size.X, w.size.Y), 0.05)]++
			}
		}
	}

	report := wallAuditReport{
		Seeds:     wallAuditSeeds,
		Levels:    levels,
		EdgeCount: len(edges),
		Failures:  failures,
	}
	for _, tile := range tiles {
		report.Candidates = append(report.Candidates, wallAuditCandidateFor(tile, edges))
	}
	report.RemainderHistogramMod1 = wallAuditRemainders(edges)
	for _, size := range sortedFloatKeys(thickness) {
		report.ThicknessHistogram = append(report.ThicknessHistogram, wallAuditBucket{Value: size, Count: thickness[size]})
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	if err := os.WriteFile(outPath, append(data, '\n'), 0o644); err != nil {
		t.Fatalf("write report: %v", err)
	}
	t.Logf("wall grid audit: %d edges over %d seeds x %d levels, %d generation failure(s) -> %s",
		len(edges), len(wallAuditSeeds), len(levels), len(failures), outPath)
}

var wallAuditSeeds = []string{
	"audit-01", "audit-02", "audit-03", "audit-04", "audit-05",
	"audit-06", "audit-07", "audit-08", "audit-09", "audit-10",
	"audit-11", "audit-12", "audit-13", "audit-14", "audit-15",
	"audit-16", "audit-17", "audit-18", "audit-19", "audit-20",
}

const wallAuditEpsilon = 1e-6

type wallAuditEdge struct {
	source string
	value  float64
}

type wallAuditFailure struct {
	Seed  string `json:"seed"`
	Level int    `json:"level"`
	Error string `json:"error"`
}

type wallAuditBucket struct {
	Value float64 `json:"value"`
	Count int     `json:"count"`
}

type wallAuditSourceFraction struct {
	Source          string  `json:"source"`
	Edges           int     `json:"edges"`
	AlignedFraction float64 `json:"aligned_fraction"`
}

type wallAuditCandidate struct {
	Tile            float64                   `json:"tile"`
	AlignedFraction float64                   `json:"aligned_fraction"`
	BySource        []wallAuditSourceFraction `json:"by_source"`
}

type wallAuditReport struct {
	Seeds                  []string             `json:"seeds"`
	Levels                 []int                `json:"levels"`
	EdgeCount              int                  `json:"edge_count"`
	Failures               []wallAuditFailure   `json:"generation_failures"`
	Candidates             []wallAuditCandidate `json:"candidates"`
	RemainderHistogramMod1 []wallAuditBucket    `json:"remainder_histogram_mod_1"`
	ThicknessHistogram     []wallAuditBucket    `json:"thickness_histogram"`
}

// wallAuditMaxDepth covers every floor profile, every biome palette (client
// presentation data read straight from the shared JSON), and two boss floors.
func wallAuditMaxDepth(t *testing.T, rulesDir string, gen DungeonGenerationRules) int {
	t.Helper()
	maxDepth := 1
	for _, profile := range gen.FloorProfiles {
		maxDepth = max(maxDepth, profile.MinDepth)
		if profile.MaxDepth != nil {
			maxDepth = max(maxDepth, *profile.MaxDepth)
		}
	}
	if gen.BossFloor.Cadence > 0 {
		maxDepth = max(maxDepth, -gen.BossFloor.FirstLevel+gen.BossFloor.Cadence)
	}
	raw, err := os.ReadFile(filepath.Join(rulesDir, "dungeon_generation.v0.json"))
	if err != nil {
		t.Fatalf("read biome palettes: %v", err)
	}
	var doc struct {
		BiomePalettes []struct {
			MinDepth int  `json:"min_depth"`
			MaxDepth *int `json:"max_depth"`
		} `json:"biome_palettes"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse biome palettes: %v", err)
	}
	for _, palette := range doc.BiomePalettes {
		maxDepth = max(maxDepth, palette.MinDepth)
		if palette.MaxDepth != nil {
			maxDepth = max(maxDepth, *palette.MaxDepth)
		}
	}
	return maxDepth
}

func wallAuditTiles(t *testing.T) []float64 {
	t.Helper()
	tiles := []float64{0.5, 1.0, 2.0}
	for raw := range strings.SplitSeq(os.Getenv("ARPG_WALL_GRID_EXTRA_TILES"), ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		tile, err := strconv.ParseFloat(raw, 64)
		if err != nil || tile <= 0 {
			t.Fatalf("invalid ARPG_WALL_GRID_EXTRA_TILES entry %q", raw)
		}
		tiles = append(tiles, tile)
	}
	sort.Float64s(tiles)
	return tiles
}

func edgeAligned(value, tile float64) bool {
	steps := value / tile
	return math.Abs(steps-math.Round(steps))*tile < wallAuditEpsilon
}

func wallAuditCandidateFor(tile float64, edges []wallAuditEdge) wallAuditCandidate {
	total, aligned := 0, 0
	perSource := map[string][2]int{} // [edges, aligned]
	for _, edge := range edges {
		counts := perSource[edge.source]
		counts[0]++
		total++
		if edgeAligned(edge.value, tile) {
			counts[1]++
			aligned++
		}
		perSource[edge.source] = counts
	}
	candidate := wallAuditCandidate{Tile: tile, AlignedFraction: fraction(aligned, total)}
	sources := make([]string, 0, len(perSource))
	for source := range perSource {
		sources = append(sources, source)
	}
	sort.Strings(sources)
	for _, source := range sources {
		counts := perSource[source]
		candidate.BySource = append(candidate.BySource, wallAuditSourceFraction{
			Source: source, Edges: counts[0], AlignedFraction: fraction(counts[1], counts[0]),
		})
	}
	return candidate
}

func wallAuditRemainders(edges []wallAuditEdge) []wallAuditBucket {
	counts := map[float64]int{}
	for _, edge := range edges {
		remainder := edge.value - math.Floor(edge.value)
		bucket := roundTo(remainder, 0.05)
		if bucket >= 1.0 {
			bucket = 0
		}
		counts[bucket]++
	}
	buckets := make([]wallAuditBucket, 0, len(counts))
	for _, value := range sortedFloatKeys(counts) {
		buckets = append(buckets, wallAuditBucket{Value: value, Count: counts[value]})
	}
	return buckets
}

func sortedFloatKeys(m map[float64]int) []float64 {
	keys := make([]float64, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Float64s(keys)
	return keys
}

func roundTo(value, step float64) float64 {
	return math.Round(value/step) * step
}

func fraction(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return math.Round(float64(part)/float64(total)*10000) / 10000
}
