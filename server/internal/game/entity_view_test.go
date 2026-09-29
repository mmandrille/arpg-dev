package game

import (
	"sort"
	"testing"
)

// coopSimWithGuestClass builds a host + guest session where the guest plays a
// class other than the host's, picked from the loaded rules.
func coopSimWithGuestClass(t *testing.T) (sim *Sim, hostID, guestID uint64, hostClass, guestClass string) {
	t.Helper()
	rules := loadRules(t)
	sim, err := NewSim("sess_player_class_view", "player_class_view_seed", rules)
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	hostID = sim.playerID
	hostClass = sim.progression.CharacterClass
	classIDs := make([]string, 0, len(rules.CharacterProgression.Classes))
	for id := range rules.CharacterProgression.Classes { //nolint:determinism // keys sorted below
		classIDs = append(classIDs, id)
	}
	sort.Strings(classIDs)
	for _, id := range classIDs {
		if id != hostClass {
			guestClass = id
			break
		}
	}
	if hostClass == "" || guestClass == "" {
		t.Fatalf("need two distinct classes: host=%q guest=%q rules=%v", hostClass, guestClass, classIDs)
	}
	progression := rules.DefaultCharacterProgressionState()
	progression.CharacterClass = guestClass
	guestID, err = sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", progression)
	if err != nil {
		t.Fatalf("add guest: %v", err)
	}
	hostLevel, _ := sim.PlayerCurrentLevel(hostID)
	guestLevel, _ := sim.PlayerCurrentLevel(guestID)
	if hostLevel != guestLevel {
		t.Fatalf("host level %d != guest level %d; test needs both players visible", hostLevel, guestLevel)
	}
	return sim, hostID, guestID, hostClass, guestClass
}

func entityViewByID(views []EntityView, id uint64) (EntityView, bool) {
	for _, v := range views {
		if v.ID == idStr(id) {
			return v, true
		}
	}
	return EntityView{}, false
}

func TestPlayerEntityViewsCarryEachMembersClass(t *testing.T) {
	sim, hostID, guestID, hostClass, guestClass := coopSimWithGuestClass(t)
	want := map[uint64]string{hostID: hostClass, guestID: guestClass}
	for _, viewer := range []uint64{hostID, guestID} {
		snap := sim.SnapshotForPlayer(viewer)
		for _, member := range []uint64{hostID, guestID} {
			view, ok := entityViewByID(snap.Entities, member)
			if !ok {
				t.Fatalf("viewer %d snapshot missing player entity %d", viewer, member)
			}
			if view.CharacterClass != want[member] {
				t.Fatalf("viewer %d sees player %d class %q, want %q", viewer, member, view.CharacterClass, want[member])
			}
		}
	}
}

// Class is fixed for a session today (chosen at character creation), so this
// pins the property a future in-session class change would rely on: player
// views are derived from live progression, never cached at spawn.
func TestPlayerEntityClassTracksLiveProgressionInNextDelta(t *testing.T) {
	sim, hostID, guestID, hostClass, _ := coopSimWithGuestClass(t)
	sim.usePlayer(sim.players[hostID])
	sim.players[guestID].Progression.CharacterClass = hostClass

	var guestUpdate *EntityView
	move := []Input{{MessageID: "guest_move", ActorPlayerID: guestID, Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 3}}}
	for i := 0; i < 5 && guestUpdate == nil; i++ {
		for _, res := range sim.TickResults(move) {
			for _, change := range res.Changes {
				if change.Op == OpEntityUpdate && change.Entity != nil && change.Entity.ID == idStr(guestID) {
					guestUpdate = change.Entity
				}
			}
		}
		move = nil
	}
	if guestUpdate == nil {
		t.Fatal("guest move produced no entity_update for the guest player")
	}
	if guestUpdate.CharacterClass != hostClass {
		t.Fatalf("guest delta class = %q, want live progression class %q", guestUpdate.CharacterClass, hostClass)
	}
}

func TestNonPlayerEntityViewsOmitCharacterClass(t *testing.T) {
	sim, _, _, _, _ := coopSimWithGuestClass(t)
	for _, view := range sim.Snapshot().Entities {
		if view.Type != playerEntity && view.Type != companionEntity && view.CharacterClass != "" {
			t.Fatalf("%s entity %s carries character_class %q", view.Type, view.ID, view.CharacterClass)
		}
	}
}
