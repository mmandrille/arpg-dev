package game

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"
)

func TestBossLaneRules(t *testing.T) {
	rules := loadRules(t)
	base := rules.BossPatterns["shifting_bulwark"]
	if base.Lanes == nil || len(base.Phases) != 4 {
		t.Fatalf("lane pattern = %+v", base)
	}
	for _, templateID := range []string{"cave_warden", "crypt_matron"} {
		deck := rules.BossTemplates[templateID].PatternDeck
		if deck[len(deck)-1] != "shifting_bulwark" {
			t.Fatalf("%s last pattern = %q", templateID, deck[len(deck)-1])
		}
	}
	for _, tc := range []struct {
		name string
		edit func(*BossPatternDef)
	}{
		{"zero width", func(p *BossPatternDef) { p.Lanes.Width = 0 }},
		{"narrow corridor", func(p *BossPatternDef) { p.Lanes.Width = playerRadius * 2 }},
		{"non-finite width", func(p *BossPatternDef) { p.Lanes.Width = math.NaN() }},
		{"non-finite intensity", func(p *BossPatternDef) { p.Phases[0].LaneIntensity = math.NaN() }},
		{"missing safe", func(p *BossPatternDef) { p.Lanes.SafeSequence = nil }},
		{"invalid safe", func(p *BossPatternDef) { p.Lanes.SafeSequence = []int{p.Lanes.Count} }},
		{"short warning", func(p *BossPatternDef) { p.Phases[0].DurationTicks = 1 }},
		{"strike mismatch", func(p *BossPatternDef) { p.Phases[2].Shape = "rectangle" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pattern := base
			lanes := *base.Lanes
			pattern.Lanes = &lanes
			pattern.Phases = append([]BossPatternPhase(nil), base.Phases...)
			tc.edit(&pattern)
			if err := validateBossPatterns(map[string]BossPatternDef{"bad": pattern}, 20); err == nil {
				t.Fatal("invalid lane rule accepted")
			}
		})
	}
}

func TestBossLaneHitBoundariesAndLockedFrame(t *testing.T) {
	config := BossLanesDef{Count: 3, Width: playerRadius*4 + 0.2, Length: 6, SafeSequence: []int{1}}
	lane := bossLaneFrame(Vec2{X: 10, Y: 10}, Vec2{X: 1}, config, 1)
	boss := &entity{pos: Vec2{X: 10, Y: 10}, bossLane: lane}
	phase := BossPatternPhase{Shape: "lanes"}
	safe := &entity{pos: Vec2{X: 13, Y: 10}}
	if bossPhaseHitsPlayer(boss, safe, phase) {
		t.Fatal("safe corridor took lane damage")
	}
	safeEdge := config.Width/2 - playerRadius
	safe.pos.Y += safeEdge
	if bossPhaseHitsPlayer(boss, safe, phase) {
		t.Fatal("player fully inside safe corridor took damage")
	}
	safe.pos.Y += 0.01
	if !bossPhaseHitsPlayer(boss, safe, phase) {
		t.Fatal("player footprint crossing danger boundary was not hit")
	}
	boss.pos = Vec2{X: 30, Y: 30}
	if !bossPhaseHitsPlayer(boss, safe, phase) {
		t.Fatal("moving boss carried the locked strike away")
	}
	safe.pos = Vec2{X: lane.Origin.X + lane.Length + playerRadius + 0.01, Y: 10 + config.Width}
	if bossPhaseHitsPlayer(boss, safe, phase) {
		t.Fatal("player beyond lane end was hit")
	}
}

func TestBossLanePhaseEventAndSnapshot(t *testing.T) {
	config := BossLanesDef{Count: 3, Width: 2, Length: 6, SafeColor: "#42e5af", DangerColor: "#ff6838", SafeIntensity: 0.3, SafeSequence: []int{1}}
	lane := bossLaneFrame(Vec2{X: 10, Y: 10}, Vec2{X: 1}, config, 1)
	phase := BossPatternPhase{Kind: "telegraph", DurationTicks: 20, TelegraphType: "lanes", HitShape: "lanes", LaneIntensity: 0.35}
	boss := &entity{id: 5, pos: lane.Origin, bossPatternID: "shifting_bulwark", bossPhaseIndex: 0,
		bossPhaseKind: "telegraph", bossPhaseStarted: 10, bossPhaseEnds: 30, bossLane: lane.withPhase(0, phase)}
	runtime := bossPhaseRuntime{patternID: boss.bossPatternID, index: 0, phase: phase}
	event := bossPhaseEvent("boss_phase_started", boss, runtime)
	snapshot := boss.bossPhaseView()
	if event.Lane == nil || snapshot.Lane == nil || !reflect.DeepEqual(event.Lane, snapshot.Lane) {
		t.Fatalf("event lane %+v != snapshot lane %+v", event.Lane, snapshot.Lane)
	}
	if snapshot.Telegraph == nil || snapshot.Telegraph.Type != "lanes" {
		t.Fatalf("snapshot warning = %+v", snapshot.Telegraph)
	}
	boss.pos = Vec2{X: 20, Y: 20}
	if snapshot.Lane.Origin != (Vec2{X: 10, Y: 10}) || boss.bossPhaseView().Lane.Origin != snapshot.Lane.Origin {
		t.Fatal("boss movement changed lane origin")
	}
	encoded, err := json.Marshal(event)
	if err != nil || !json.Valid(encoded) {
		t.Fatalf("lane event encoding: %v", err)
	}
}

func TestBossLaneStagesReplayAndSafeStrike(t *testing.T) {
	rules := loadRules(t)
	newSim := func() *Sim {
		sim, err := NewSimWithWorld("sess_lane", "boss_lane_seed", rules, "dungeon_levels")
		if err != nil {
			t.Fatal(err)
		}
		level, err := sim.ensureDungeonLevel(-5)
		if err != nil {
			t.Fatal(err)
		}
		sim.currentLevel = -5
		placeDefaultPlayerOnLevel(t, sim, level, Vec2{X: 18, Y: 15})
		level.entities[sim.playerID].hp = 100000
		level.entities[sim.playerID].maxHP = 100000
		level.walls = nil
		nav := sim.activeNav()
		nav.GridBounds = GridBounds{MinX: 0, MinY: 0, MaxX: 100, MaxY: 100}
		level.nav = &nav
		boss := findBossEntity(t, level)
		boss.pos = Vec2{X: 15, Y: 15}
		boss.bossPatternID = "shifting_bulwark"
		boss.bossPatternDeckIndex = len(rules.BossTemplates[boss.bossTemplateID].PatternDeck) - 1
		boss.bossPhaseIndex = -1
		boss.bossPhaseKind = ""
		boss.bossCooldownEnds = 0
		sim.syncCompatibilityFields()
		return sim
	}
	first, second := newSim(), newSim()
	warnings := 0
	active := false
	var origin Vec2
	pattern := rules.BossPatterns["shifting_bulwark"]
	ticks := pattern.Phases[0].DurationTicks + pattern.Phases[1].DurationTicks + pattern.Phases[2].DurationTicks
	for i := 0; i < ticks; i++ {
		a, b := first.Tick(nil), second.Tick(nil)
		jsonA, errA := json.Marshal(a)
		jsonB, errB := json.Marshal(b)
		if errA != nil || errB != nil || string(jsonA) != string(jsonB) {
			t.Fatalf("tick %d replay differs: %v %v", i, errA, errB)
		}
		for _, event := range a.Events {
			if event.EventType != "boss_phase_started" || event.PatternID != "shifting_bulwark" || event.PhaseKind == "recovery" {
				continue
			}
			if event.Lane == nil {
				t.Fatalf("tick %d phase event lacks lane", i)
			}
			if warnings == 0 {
				origin = event.Lane.Origin
			} else if event.Lane.Origin != origin {
				t.Fatalf("tick %d lane moved from %+v to %+v", i, origin, event.Lane.Origin)
			}
			if event.PhaseKind == "telegraph" {
				warnings++
			} else if event.PhaseKind == "active" {
				active = true
			}
		}
		if !active && findBossEntity(t, first.activeLevel()).bossActiveHit[first.playerID] {
			t.Fatalf("lane marked an active hit during warning tick %d", i)
		}
		if i == pattern.Phases[0].DurationTicks {
			boss := findBossEntity(t, first.activeLevel())
			view := boss.bossPhaseView()
			if view.Lane == nil || view.Lane.StageIndex != 1 || view.Lane.Origin != origin {
				t.Fatalf("second-stage reconnect view = %+v", view)
			}
		}
	}
	if warnings != 2 || !active {
		t.Fatalf("warning stages %d, active %v", warnings, active)
	}
	if findBossEntity(t, first.activeLevel()).bossActiveHit[first.playerID] {
		t.Fatal("safe-corridor player was hit by active lane strike")
	}
	if first.activeLevel().entities[first.playerID].hp != second.activeLevel().entities[second.playerID].hp {
		t.Fatal("replay player state differs")
	}
	dangerSim := newSim()
	for i := 0; i < pattern.Phases[0].DurationTicks+pattern.Phases[1].DurationTicks; i++ {
		result := dangerSim.Tick(nil)
		for _, event := range result.Events {
			if event.EventType != "boss_phase_started" || event.PatternID != "shifting_bulwark" || event.PhaseKind != "active" {
				continue
			}
			lane := event.Lane
			if lane == nil {
				t.Fatal("active strike missing locked frame")
			}
			along := lane.Length / 2
			cross := laneCrossCenter(lane, lane.DangerLanes[0])
			dangerSim.activeLevel().entities[dangerSim.playerID].pos = Vec2{
				X: lane.Origin.X + lane.Forward.X*along + lane.Right.X*cross,
				Y: lane.Origin.Y + lane.Forward.Y*along + lane.Right.Y*cross,
			}
		}
	}
	dangerSim.Tick(nil)
	dangerBoss := findBossEntity(t, dangerSim.activeLevel())
	if !dangerBoss.bossActiveHit[dangerSim.playerID] {
		t.Fatal("danger-lane player did not receive an active combat outcome")
	}
	dangerSim.Tick(nil)
	if !dangerBoss.bossActiveHit[dangerSim.playerID] {
		t.Fatal("one-hit guard did not persist during active phase")
	}
}

func TestBossLaneBlockedCastSkips(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_lane_blocked", "boss_lane_blocked_seed", rules, "dungeon_levels")
	if err != nil {
		t.Fatal(err)
	}
	level, err := sim.ensureDungeonLevel(-5)
	if err != nil {
		t.Fatal(err)
	}
	sim.currentLevel = -5
	placeDefaultPlayerOnLevel(t, sim, level, Vec2{X: 18, Y: 15})
	boss := findBossEntity(t, level)
	boss.pos = Vec2{X: 15, Y: 15}
	boss.bossPatternID = "shifting_bulwark"
	boss.bossPatternDeckIndex = len(rules.BossTemplates[boss.bossTemplateID].PatternDeck) - 1
	boss.bossPhaseIndex = -1
	boss.bossPhaseKind = ""
	level.walls = []wallObstacle{
		{pos: Vec2{X: 17, Y: 15}, size: Vec2{X: 1, Y: 6}, source: "generated"},
		{pos: Vec2{X: 15, Y: 11}, size: Vec2{X: 6, Y: 1}, source: "generated"},
		{pos: Vec2{X: 15, Y: 19}, size: Vec2{X: 6, Y: 1}, source: "generated"},
	}
	nav := sim.activeNav()
	nav.GridBounds = GridBounds{MinX: 0, MinY: 0, MaxX: 100, MaxY: 100}
	level.nav = &nav
	sim.syncCompatibilityFields()
	result := sim.Tick(nil)
	for _, event := range result.Events {
		if event.EventType == "boss_phase_started" && event.PatternID == "shifting_bulwark" {
			t.Fatal("blocked safe corridor produced impossible warning")
		}
	}
	if boss.bossLane != nil || boss.bossPhaseKind != "" || boss.bossPatternID == "shifting_bulwark" {
		t.Fatalf("blocked cast state = lane %+v phase %s pattern %s", boss.bossLane, boss.bossPhaseKind, boss.bossPatternID)
	}
	level.walls = []wallObstacle{{pos: Vec2{X: 17, Y: 15}, size: Vec2{X: 1, Y: 1}, source: "generated"}}
	sim.syncCompatibilityFields()
	pattern := rules.BossPatterns["shifting_bulwark"]
	runtime := bossPhaseRuntime{patternID: "shifting_bulwark", index: 0, phase: pattern.Phases[0]}
	if !sim.prepareBossLane(boss, runtime) || boss.bossLane == nil {
		t.Fatal("valid alternate lane frame was not selected")
	}
	if boss.bossLane.Forward == boss.bossPhaseAim {
		t.Fatal("blocked direct aim was selected instead of a valid alternate")
	}
}

func TestBossLaneBossFloorWallFallback(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_lane_floor", "boss_lane_floor_seed", rules, "boss_floor_gate_lab")
	if err != nil {
		t.Fatal(err)
	}
	boss := findBossEntity(t, sim.activeLevel())
	boss.pos = Vec2{X: 25, Y: 15}
	sim.activeLevel().entities[sim.playerID].pos = Vec2{X: 26, Y: 15}
	pattern := rules.BossPatterns["shifting_bulwark"]
	runtime := bossPhaseRuntime{patternID: "shifting_bulwark", index: 0, phase: pattern.Phases[0]}
	if !sim.prepareBossLane(boss, runtime) || boss.bossLane == nil {
		t.Fatal("boss-floor perimeter blocked every authored lane frame")
	}
	if boss.bossLane.Forward == (Vec2{X: 1}) {
		t.Fatal("direct lane frame crossed the boss-floor perimeter")
	}
}

func TestBossLaneBossFloorLabCast(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_lane_bot_lab", "boss_lane_telegraphs", rules, "boss_floor_gate_lab")
	if err != nil {
		t.Fatal(err)
	}
	player := sim.activeLevel().entities[sim.playerID]
	player.hp, player.maxHP = 100000, 100000
	for tick := 0; tick < 1000; tick++ {
		result := sim.Tick(nil)
		for _, event := range result.Events {
			if event.EventType == "boss_phase_started" && event.PatternID == "shifting_bulwark" && event.PhaseIndex != nil && *event.PhaseIndex == 0 {
				if event.Lane == nil || !sim.bossLaneCorridorWalkable(event.Lane) {
					t.Fatalf("boss-floor warning has no walkable safe corridor: %+v", event.Lane)
				}
				t.Logf("lane warning tick=%d player=%+v lane=%+v", tick, player.pos, event.Lane)
				return
			}
		}
	}
	t.Fatal("boss-floor lab did not reach the lane pattern")
}

func TestBossLaneCoopSafeAndDanger(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_lane_coop", "boss_lane_coop_seed", rules, "dungeon_levels")
	if err != nil {
		t.Fatal(err)
	}
	hostID := sim.playerID
	guestID, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatal(err)
	}
	level, err := sim.ensureDungeonLevel(-5)
	if err != nil {
		t.Fatal(err)
	}
	sim.usePlayer(sim.players[hostID])
	placeDefaultPlayerOnLevel(t, sim, level, Vec2{X: 18, Y: 15})
	guest := sim.levels[townLevel].entities[guestID]
	delete(sim.levels[townLevel].entities, guestID)
	guest.pos = Vec2{X: 18, Y: 17}
	level.entities[guestID] = guest
	sim.players[guestID].CurrentLevel = -5
	sim.syncCompatibilityFields()

	boss := findBossEntity(t, level)
	boss.pos = Vec2{X: 15, Y: 15}
	pattern := rules.BossPatterns["shifting_bulwark"]
	boss.bossLane = bossLaneFrame(boss.pos, Vec2{X: 1}, *pattern.Lanes, 1)
	boss.bossActiveHit = map[uint64]bool{}
	phase := pattern.Phases[2]
	var result TickResult
	sim.applyBossActivePhase(boss, phase, &result)
	sim.applyBossActivePhase(boss, phase, &result)
	if boss.bossActiveHit[hostID] || !boss.bossActiveHit[guestID] {
		t.Fatalf("co-op active-hit state = %+v, host=%d guest=%d", boss.bossActiveHit, hostID, guestID)
	}
	guestOutcomes := 0
	for _, event := range result.Events {
		if event.TargetEntityID == idStr(hostID) {
			t.Fatalf("safe player received combat event %+v", event)
		}
		if event.TargetEntityID == idStr(guestID) {
			guestOutcomes++
		}
	}
	if guestOutcomes != 1 {
		t.Fatalf("danger player received %d combat outcomes, want one", guestOutcomes)
	}
}

func TestBossLaneRuleVariantControlsWidthAndStageDuration(t *testing.T) {
	probe := func(extraTicks int, extraWidth float64) (int, float64) {
		rules := loadRules(t)
		pattern := rules.BossPatterns["shifting_bulwark"]
		lanes := *pattern.Lanes
		lanes.Width += extraWidth
		pattern.Lanes = &lanes
		pattern.Phases = append([]BossPatternPhase(nil), pattern.Phases...)
		pattern.Phases[0].DurationTicks += extraTicks
		rules.BossPatterns["shifting_bulwark"] = pattern
		if err := validateBossPatterns(map[string]BossPatternDef{"shifting_bulwark": pattern}, 20); err != nil {
			t.Fatal(err)
		}
		sim, err := NewSimWithWorld("sess_lane_variant", "boss_lane_variant_seed", rules, "dungeon_levels")
		if err != nil {
			t.Fatal(err)
		}
		level, err := sim.ensureDungeonLevel(-5)
		if err != nil {
			t.Fatal(err)
		}
		sim.currentLevel = -5
		placeDefaultPlayerOnLevel(t, sim, level, Vec2{X: 18, Y: 15})
		level.entities[sim.playerID].hp = 100000
		level.entities[sim.playerID].maxHP = 100000
		level.walls = nil
		nav := sim.activeNav()
		nav.GridBounds = GridBounds{MinX: 0, MinY: 0, MaxX: 100, MaxY: 100}
		level.nav = &nav
		boss := findBossEntity(t, level)
		boss.pos = Vec2{X: 15, Y: 15}
		boss.bossPatternID = "shifting_bulwark"
		boss.bossPatternDeckIndex = len(rules.BossTemplates[boss.bossTemplateID].PatternDeck) - 1
		boss.bossPhaseIndex = -1
		boss.bossPhaseKind = ""
		sim.syncCompatibilityFields()
		start := sim.Tick(nil)
		var width float64
		for _, event := range start.Events {
			if event.EventType == "boss_phase_started" && event.PatternID == "shifting_bulwark" && event.Lane != nil {
				width = event.Lane.Width
			}
		}
		if width == 0 {
			t.Fatal("lane cast did not start")
		}
		for ticks := 1; ticks < 100; ticks++ {
			result := sim.Tick(nil)
			for _, event := range result.Events {
				if event.EventType == "boss_phase_started" && event.PatternID == "shifting_bulwark" && event.PhaseIndex != nil && *event.PhaseIndex == 1 {
					return ticks, width
				}
			}
		}
		t.Fatal("second lane warning stage did not start")
		return 0, 0
	}
	baselineTicks, baselineWidth := probe(0, 0)
	variantTicks, variantWidth := probe(5, 0.4)
	if variantTicks-baselineTicks != 5 || variantWidth-baselineWidth < 0.399 || variantWidth-baselineWidth > 0.401 {
		t.Fatalf("variant timing/width = (%d,%v), baseline (%d,%v)", variantTicks, variantWidth, baselineTicks, baselineWidth)
	}
}
