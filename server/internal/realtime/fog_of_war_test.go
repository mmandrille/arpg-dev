package realtime

import (
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func TestFogOfWarFanoutSuppressesFarMonsterDeltas(t *testing.T) {
	sim := mustRealtimeSim(t, "combat_control_lab")
	hostID := sim.DefaultPlayerID()
	monster := game.EntityView{ID: "1003", Type: "monster", MonsterDefID: "dungeon_mob", Position: game.Vec2{X: 13, Y: 5}}
	host := &loopClient{playerID: hostID, sendCh: make(chan outEnvelope, 8), done: make(chan struct{})}
	loop := &sessionLoop{sess: store.Session{ID: "sess_fog_far"}, sim: sim}

	loop.fanoutResult(game.TickResult{
		Tick:  1,
		Level: 0,
		Changes: []game.Change{{
			Op:     game.OpEntityUpdate,
			Entity: &monster,
		}},
		Events: []game.Event{{EventType: "monster_aggro", EntityID: monster.ID}},
	}, []*loopClient{host}, nil, map[uint64]int{hostID: 0})

	assertNoEnvelope(t, host)
}

func TestFogOfWarFanoutKeepsNearMonsterDeltas(t *testing.T) {
	sim := mustRealtimeSim(t, "vertical_slice")
	hostID := sim.DefaultPlayerID()
	monster := findEntity(t, sim.Snapshot(), "monster", "training_dummy")
	host := &loopClient{playerID: hostID, sendCh: make(chan outEnvelope, 8), done: make(chan struct{})}
	loop := &sessionLoop{sess: store.Session{ID: "sess_fog_near"}, sim: sim}

	loop.fanoutResult(game.TickResult{
		Tick:  1,
		Level: 0,
		Changes: []game.Change{{
			Op:     game.OpEntityUpdate,
			Entity: &monster,
		}},
		Events: []game.Event{{EventType: "monster_aggro", EntityID: monster.ID}},
	}, []*loopClient{host}, nil, map[uint64]int{hostID: 0})

	delta := mustReceiveDelta(t, host)
	if len(delta.Changes) != 1 || delta.Changes[0].Entity == nil || delta.Changes[0].Entity.ID != monster.ID {
		t.Fatalf("near monster changes = %+v, want visible monster delta", delta.Changes)
	}
	if len(delta.Events) != 1 || delta.Events[0].EventType != "monster_aggro" || delta.Events[0].EntityID != monster.ID {
		t.Fatalf("near monster events = %+v, want visible monster event", delta.Events)
	}
}

func TestFogOfWarSessionSnapshotUsesRecipientScope(t *testing.T) {
	sim := mustRealtimeSim(t, "combat_control_lab")
	monsterID := "1003"
	hostID := sim.DefaultPlayerID()
	loop := &sessionLoop{sess: store.Session{ID: "sess_fog_snapshot"}, sim: sim}

	env := loop.snapshotEnvelope(hostID)
	snap, ok := env.Payload.(game.Snapshot)
	if !ok {
		t.Fatalf("snapshot payload type = %T, want game.Snapshot", env.Payload)
	}

	if snapshotContainsEntity(snap, monsterID) {
		t.Fatalf("session snapshot leaked hidden monster %s: %+v", monsterID, snap.Entities)
	}
}

func mustRealtimeSim(t *testing.T, worldID string) *game.Sim {
	t.Helper()
	rulesDir, err := game.FindSharedRulesDir()
	if err != nil {
		t.Fatalf("find rules: %v", err)
	}
	rules, err := game.LoadRules(rulesDir)
	if err != nil {
		t.Fatalf("load rules: %v", err)
	}
	sim, err := game.NewSimWithWorld("sess_"+worldID, worldID+"_seed", rules, worldID)
	if err != nil {
		t.Fatalf("new sim %s: %v", worldID, err)
	}
	sim.SetFogOfWarEnabled(true)
	return sim
}

func snapshotContainsEntity(snap game.Snapshot, entityID string) bool {
	for _, entity := range snap.Entities {
		if entity.ID == entityID {
			return true
		}
	}
	return false
}
