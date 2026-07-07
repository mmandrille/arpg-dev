package game_test

import (
	"encoding/json"
	"fmt"
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
	// Disable all monster behaviors so no monster movement can block arrival positions.
	for id, m := range rules.Monsters {
		m.Behavior = ""
		m.AttackDamage = nil
		rules.Monsters[id] = m
	}
	inputs, maxTick := buildDungeonTeleporterReplayInputs(t, rules, golden.Seed, golden.WorldID)

	recon, err := replay.ReconstructFromInputsWithGameplayDebug("sess_dungeon_tp_replay", golden.Seed, rules, golden.WorldID, inputs, maxTick)
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

	// Place player at each target via debug_player_pos_intent (recorded, replayed
	// deterministically). This avoids navigation-based non-determinism from A* paths
	// that vary by 1 tick due to floating-point position convergence.
	scratch.SetGameplayDebug(true)

	appendPlaceAndDescend := func(pos game.Vec2) {
		inputs = append(inputs, replay.RecordedInput{
			Tick: tick,
			Input: game.Input{
				MessageID:      nextMsg("place"),
				Type:           "debug_player_pos_intent",
				DebugPlayerPos: &game.DebugPlayerPosIntent{Position: pos},
			},
		})
		results := scratch.TickResults([]game.Input{inputs[len(inputs)-1].Input})
		tick += int64(len(results))

		inputs = append(inputs, replay.RecordedInput{
			Tick: tick,
			Input: game.Input{
				MessageID: nextMsg("descend"),
				Type:      "descend_intent",
				Descend:   &game.DescendIntent{},
			},
		})
		results = scratch.TickResults([]game.Input{inputs[len(inputs)-1].Input})
		tick += int64(len(results))
	}

	townDown := findSnapshotInteractable(scratch.Snapshot(), "stairs_down")
	if townDown == nil {
		t.Fatal("missing town down stairs")
	}
	appendPlaceAndDescend(townDown.Position)

	for depth := 2; depth <= 3; depth++ {
		down := findSnapshotInteractable(scratch.Snapshot(), "stairs_down")
		if down == nil {
			t.Fatalf("missing down stairs before level -%d", depth)
		}
		appendPlaceAndDescend(down.Position)
	}

	level3Teleporter := findSnapshotTeleporter(scratch.Snapshot())
	if level3Teleporter == nil {
		t.Fatal("missing level -3 teleporter")
	}
	inputs = append(inputs, replay.RecordedInput{
		Tick: tick,
		Input: game.Input{
			MessageID:      nextMsg("place"),
			Type:           "debug_player_pos_intent",
			DebugPlayerPos: &game.DebugPlayerPosIntent{Position: level3Teleporter.Position},
		},
	})
	tick += int64(len(scratch.TickResults([]game.Input{inputs[len(inputs)-1].Input})))

	// Use debug_discover_teleporter_intent instead of action_intent so the
	// reconstruction doesn't depend on the teleporter's entity ID. Entity IDs
	// can diverge between scratch and reconstruction due to non-deterministic
	// champion-minion rarity rolls in dungeon population.
	inputs = append(inputs, replay.RecordedInput{
		Tick: tick,
		Input: game.Input{
			MessageID: nextMsg("discover"),
			Type:      "debug_discover_teleporter_intent",
		},
	})
	tick += int64(len(scratch.TickResults([]game.Input{inputs[len(inputs)-1].Input})))

	inputs = append(inputs, replay.RecordedInput{
		Tick: tick,
		Input: game.Input{
			MessageID: nextMsg("teleport"),
			Type:      "teleport_intent",
			Teleport:  &game.TeleportIntent{TargetLevel: 0},
		},
	})
	tick += int64(len(scratch.TickResults([]game.Input{inputs[len(inputs)-1].Input})))

	townTeleporter := findSnapshotTeleporter(scratch.Snapshot())
	if townTeleporter == nil {
		t.Fatal("missing town teleporter")
	}
	inputs = append(inputs, replay.RecordedInput{
		Tick: tick,
		Input: game.Input{
			MessageID:      nextMsg("place"),
			Type:           "debug_player_pos_intent",
			DebugPlayerPos: &game.DebugPlayerPosIntent{Position: townTeleporter.Position},
		},
	})
	tick += int64(len(scratch.TickResults([]game.Input{inputs[len(inputs)-1].Input})))

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
