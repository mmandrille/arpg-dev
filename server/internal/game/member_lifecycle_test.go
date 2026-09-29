package game

import "testing"

func TestMemberLeaveAndRejoinMutations(t *testing.T) {
	sim, err := NewSimWithWorld("sess_member_lifecycle", "member_lifecycle_seed", loadRules(t), "combat_control_lab")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	guest, err := sim.AddGuestPlayer("acct_g", "char_g", "Guest", sim.rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatalf("add guest: %v", err)
	}
	inPlay := func() bool {
		_, ok := sim.levels[townLevel].entities[guest]
		return ok && sim.PlayerConnected(guest)
	}

	sim.ApplyMemberLeave(guest)
	if inPlay() {
		t.Fatal("leave must remove the guest entity and disconnect it")
	}
	if err := sim.ApplyMemberRejoin(guest, false); err != nil || inPlay() {
		t.Fatalf("rejoin without respawn reconnects only, err=%v inPlay=%v", err, inPlay())
	}
	sim.ApplyMemberLeave(guest)
	if err := sim.ApplyMemberRejoin(guest, true); err != nil || !inPlay() {
		t.Fatalf("rejoin with respawn must put the guest back in town, err=%v inPlay=%v", err, inPlay())
	}
}

func TestIsMemberLifecycleInput(t *testing.T) {
	for _, typ := range []string{SystemMemberJoinInputType, SystemMemberLeaveInputType, SystemMemberRejoinInputType} {
		if !IsMemberLifecycleInput(Input{Type: typ}) {
			t.Fatalf("%s must be a member lifecycle input", typ)
		}
	}
	if IsMemberLifecycleInput(Input{Type: SystemLoadShedInputType}) || IsMemberLifecycleInput(Input{Type: "move_intent"}) {
		t.Fatal("only membership types are member lifecycle inputs")
	}
}
