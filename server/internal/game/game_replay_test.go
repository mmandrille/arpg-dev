package game_test

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/replay"
)

func TestDungeonTeleportersReplayGolden(t *testing.T) {
	golden := loadDungeonTeleportersGolden(t)
	rules := loadRules(t)
	// This replay fixture owns teleporter determinism, not dungeon combat pressure.
	dungeonMob := rules.Monsters["dungeon_mob"]
	dungeonMob.Behavior = ""
	dungeonMob.AttackDamage = nil
	rules.Monsters["dungeon_mob"] = dungeonMob
	inputs, maxTick := buildDungeonTeleporterReplayInputs(t, rules, golden.Seed, golden.WorldID)

	recon, err := replay.ReconstructFromInputs("sess_dungeon_tp_replay", golden.Seed, rules, golden.WorldID, inputs, maxTick)
	if err != nil {
		t.Fatalf("reconstruct: %v", err)
	}

	want := golden.DiscoverDescendTeleport
	if recon.Snapshot.CurrentLevel != want.ExpectedLevel {
		t.Fatalf("reconstructed currentLevel = %d, want %d", recon.Snapshot.CurrentLevel, want.ExpectedLevel)
	}
	player := snapshotEntityByID(recon.Snapshot, "1001")
	if player == nil || player.Position != want.ExpectedPlayerPosition {
		t.Fatalf("reconstructed player = %+v, want at %+v", player, want.ExpectedPlayerPosition)
	}
	assertTeleporterDiscoveryView(t, recon.Snapshot.DiscoveredTeleporters, want.DiscoveredTeleporters)
}

func buildDungeonTeleporterReplayInputs(t *testing.T, rules *game.Rules, seed, worldID string) ([]replay.RecordedInput, int64) {
	t.Helper()
	scratch, err := game.NewSimWithWorld("sess_tp_replay_build", seed, rules, worldID)
	if err != nil {
		t.Fatalf("new scratch sim: %v", err)
	}

	var (
		inputs  []replay.RecordedInput
		tick    int64
		msgStep int
	)
	nextMsg := func(prefix string) string {
		msgStep++
		return fmt.Sprintf("%s_%d", prefix, msgStep)
	}

	townDown := findSnapshotInteractable(scratch.Snapshot(), "stairs_down")
	if townDown == nil {
		t.Fatal("missing town down stairs")
	}
	tick = appendMoveToAndAdvance(t, scratch, rules, &inputs, tick, nextMsg("move"), townDown.Position)
	inputs = append(inputs, replay.RecordedInput{
		Tick: tick,
		Input: game.Input{
			MessageID: nextMsg("descend"),
			Type:      "descend_intent",
			Descend:   &game.DescendIntent{},
		},
	})
	scratch.TickResults([]game.Input{inputs[len(inputs)-1].Input})
	tick++

	for depth := 2; depth <= 3; depth++ {
		down := findSnapshotInteractable(scratch.Snapshot(), "stairs_down")
		if down == nil {
			t.Fatalf("missing down stairs before level -%d", depth)
		}
		tick = appendMoveToAndAdvance(t, scratch, rules, &inputs, tick, nextMsg("move"), down.Position)
		inputs = append(inputs, replay.RecordedInput{
			Tick: tick,
			Input: game.Input{
				MessageID: nextMsg("descend"),
				Type:      "descend_intent",
				Descend:   &game.DescendIntent{},
			},
		})
		scratch.TickResults([]game.Input{inputs[len(inputs)-1].Input})
		tick++
	}

	level3Teleporter := findSnapshotTeleporter(scratch.Snapshot())
	if level3Teleporter == nil {
		t.Fatal("missing level -3 teleporter")
	}
	tick = appendMoveToAndAdvance(t, scratch, rules, &inputs, tick, nextMsg("move"), level3Teleporter.Position)
	inputs = append(inputs, replay.RecordedInput{
		Tick: tick,
		Input: game.Input{
			MessageID: nextMsg("discover"),
			Type:      "action_intent",
			Action:    &game.ActionIntent{TargetID: level3Teleporter.ID},
		},
	})
	scratch.Tick([]game.Input{inputs[len(inputs)-1].Input})
	tick++

	inputs = append(inputs, replay.RecordedInput{
		Tick: tick,
		Input: game.Input{
			MessageID: nextMsg("teleport"),
			Type:      "teleport_intent",
			Teleport:  &game.TeleportIntent{TargetLevel: 0},
		},
	})
	scratch.TickResults([]game.Input{inputs[len(inputs)-1].Input})
	tick++

	townTeleporter := findSnapshotTeleporter(scratch.Snapshot())
	if townTeleporter == nil {
		t.Fatal("missing town teleporter")
	}
	tick = appendMoveToAndAdvance(t, scratch, rules, &inputs, tick, nextMsg("move"), townTeleporter.Position)
	inputs = append(inputs, replay.RecordedInput{
		Tick: tick,
		Input: game.Input{
			MessageID: nextMsg("teleport"),
			Type:      "teleport_intent",
			Teleport:  &game.TeleportIntent{TargetLevel: -3},
		},
	})
	return inputs, tick
}

func appendMoveToAndAdvance(
	t *testing.T,
	sim *game.Sim,
	rules *game.Rules,
	inputs *[]replay.RecordedInput,
	tick int64,
	messageID string,
	pos game.Vec2,
) int64 {
	t.Helper()
	*inputs = append(*inputs, replay.RecordedInput{
		Tick: tick,
		Input: game.Input{
			MessageID: messageID,
			Type:      "move_to_intent",
			MoveTo:    &game.MoveToIntent{Position: pos},
		},
	})
	sim.TickResults([]game.Input{(*inputs)[len(*inputs)-1].Input})
	// Large tick budget: generated dungeon levels can have very winding paths,
	// and navigation may re-plan multiple times to escape wall-pocket positions.
	const navBudget = 30000
	for guard := 0; guard < navBudget; guard++ {
		player := snapshotEntityByID(sim.Snapshot(), "1001")
		if player != nil && distance(player.Position, pos) <= replayInteractableReach(rules) {
			break
		}
		tick++
		sim.Tick(nil)
		if guard == navBudget-1 {
			t.Fatalf("player did not reach %+v; last pos %+v", pos, player)
		}
	}
	return tick + 1
}

func replayInteractableReach(rules *game.Rules) float64 {
	return rules.Combat.UnarmedReach + 0.5 + 0.001
}

func findSnapshotTeleporter(snap game.Snapshot) *game.EntityView {
	for i := range snap.Entities {
		e := &snap.Entities[i]
		if e.Type == "interactable" && e.InteractableDefID == "teleporter" {
			return e
		}
	}
	return nil
}

func findSnapshotInteractable(snap game.Snapshot, defID string) *game.EntityView {
	for i := range snap.Entities {
		e := &snap.Entities[i]
		if e.Type == "interactable" && e.InteractableDefID == defID {
			return e
		}
	}
	return nil
}

func snapshotEntityByID(snap game.Snapshot, id string) *game.EntityView {
	for i := range snap.Entities {
		if snap.Entities[i].ID == id {
			return &snap.Entities[i]
		}
	}
	return nil
}

func distance(a, b game.Vec2) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func assertTeleporterDiscoveryView(t *testing.T, got []game.TeleporterDiscoveryView, want []teleporterDiscoveryGolden) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("discovery view len = %d, want %d: got=%+v", len(got), len(want), got)
	}
	for i, row := range want {
		if got[i].Level != row.Level || got[i].Discovered != row.Discovered {
			t.Fatalf("discovery[%d] = %+v, want level=%d discovered=%v", i, got[i], row.Level, row.Discovered)
		}
	}
}

type dungeonTeleportersGolden struct {
	Seed                    string `json:"seed"`
	WorldID                 string `json:"world_id"`
	DiscoverDescendTeleport struct {
		ExpectedLevel          int                         `json:"expected_level"`
		ExpectedPlayerPosition game.Vec2                   `json:"expected_player_position"`
		DiscoveredTeleporters  []teleporterDiscoveryGolden `json:"discovered_teleporters"`
	} `json:"discover_descend_teleport"`
}

type teleporterDiscoveryGolden struct {
	Level      int  `json:"level"`
	Discovered bool `json:"discovered"`
}

func loadDungeonTeleportersGolden(t *testing.T) dungeonTeleportersGolden {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(sharedDir(t), "golden", "dungeon_teleporters.json"))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	var golden dungeonTeleportersGolden
	if err := json.Unmarshal(b, &golden); err != nil {
		t.Fatalf("parse golden: %v", err)
	}
	return golden
}

func loadRules(t *testing.T) *game.Rules {
	t.Helper()
	dir, err := game.FindSharedRulesDir()
	if err != nil {
		t.Fatalf("find shared rules: %v", err)
	}
	rules, err := game.LoadRules(dir)
	if err != nil {
		t.Fatalf("load rules: %v", err)
	}
	return rules
}

func sharedDir(t *testing.T) string {
	t.Helper()
	rulesDir, err := game.FindSharedRulesDir()
	if err != nil {
		t.Fatalf("locate shared/rules: %v", err)
	}
	return filepath.Dir(rulesDir)
}
