package game

import (
	"fmt"
	"math"
	"testing"
)

// Find a seed where the level -3 teleporter is within actionable range of stairs_up,
// so no navigation is needed in TestDungeonTeleportersReplayGolden.
func TestFindTeleporterAdjacent(t *testing.T) {
	t.Skip("diagnostic seed search for teleporter golden regeneration")
	rules := loadRules(t)
	// Teleporter is reachable from stairs_up if dist <= unarmedReach + interactableInteractionRadius
	reach := rules.Combat.UnarmedReach + interactableInteractionRadius

	for i := 0; i < 2000; i++ {
		seed := fmt.Sprintf("v448_tp_adj_%04d", i)

		// Check all 3 levels generate ok
		bad := false
		for d := 1; d <= 3; d++ {
			if _, err := GenerateDungeonLevel(seed, -d, rules.DungeonGeneration); err != nil {
				bad = true
				break
			}
		}
		if bad {
			continue
		}

		// Level -3: teleporter within reach of stairs_up?
		level, _ := GenerateDungeonLevel(seed, -3, rules.DungeonGeneration)
		if len(level.teleporters) == 0 {
			continue
		}
		tp := level.teleporters[0]
		var upPos Vec2
		for _, s := range level.stairs {
			if s.defID == "stairs_up" {
				upPos = s.pos
				break
			}
		}
		dist := math.Hypot(upPos.X-tp.pos.X, upPos.Y-tp.pos.Y)
		if dist > reach {
			continue
		}

		// Run full scenario to get expected position
		sim, err := NewSimWithWorld("sess_adj", seed, rules, "dungeon_levels")
		if err != nil {
			continue
		}
		for depth := 1; depth <= 3; depth++ {
			down := sim.findStair(sim.activeLevel(), stairsDownDefID)
			if down == nil {
				break
			}
			sim.entities[sim.playerID].pos = down.pos
			sim.TickResults([]Input{{MessageID: fmt.Sprintf("d%d", depth), Type: "descend_intent", Descend: &DescendIntent{}}})
		}
		if sim.currentLevel != -3 {
			continue
		}

		tp3 := sim.findTeleporter(sim.activeLevel())
		if tp3 == nil {
			continue
		}
		// discover the teleporter
		sim.Tick([]Input{{MessageID: "disc", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(tp3.id)}}})
		r := sim.TickResults([]Input{{MessageID: "tp_town", Type: "teleport_intent", Teleport: &TeleportIntent{TargetLevel: townLevel}}})
		if len(r) < 2 {
			continue
		}
		townTP := sim.findTeleporter(sim.activeLevel())
		if townTP == nil {
			continue
		}
		// go near town teleporter and teleport back
		sim.entities[sim.playerID].pos = townTP.pos
		r2 := sim.TickResults([]Input{{MessageID: "tp_3", Type: "teleport_intent", Teleport: &TeleportIntent{TargetLevel: -3}}})
		if len(r2) < 2 {
			continue
		}

		t.Logf("FOUND: seed=%s dist=%.2f upPos=%+v tp3=%+v finalPos=%+v", seed, dist, upPos, tp.pos, sim.entities[sim.playerID].pos)
		return
	}
	t.Fatal("no seed found")
}
