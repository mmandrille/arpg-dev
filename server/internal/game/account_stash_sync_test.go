package game

import "testing"

func stashSyncInput(playerID uint64, gold int, items ...PersistedStashItem) Input {
	return Input{
		MessageID: "sys_stash_sync",
		Type:      SystemAccountStashSyncInputType,
		StashSync: &AccountStashSync{PlayerID: playerID, Items: items, Gold: gold},
	}
}

func stashRow(id, itemDefID string) PersistedStashItem {
	return PersistedStashItem{StashItemID: id, ItemDefID: itemDefID, RolledStats: []byte(`{}`)}
}

func stashChangesFor(results []TickResult, playerID uint64) []Change {
	var out []Change
	for _, res := range results {
		if res.ActorPlayerID != playerID {
			continue
		}
		for _, c := range res.Changes {
			if c.Op == OpStashItemAdd || c.Op == OpStashItemRemove || c.Op == OpStashGoldUpdate {
				out = append(out, c)
			}
		}
	}
	return out
}

func TestAccountStashSyncEmitsRemovalsThenUpsertsThenGold(t *testing.T) {
	sim := MustNewSim("sess_stash_sync_diff", "01", loadRules(t))
	sim.LoadAccountStash([]PersistedStashItem{stashRow("5010", "red_potion"), stashRow("5020", "rusty_sword"), stashRow("5040", "blue_potion")}, 5, 0)
	pid := sim.DefaultPlayerID()

	// 5010 removed, 5020 changes definition (in-place update), 5030 added, 5040 unchanged, gold 5 -> 9.
	results := sim.TickResults([]Input{stashSyncInput(pid, 9, stashRow("5040", "blue_potion"), stashRow("5030", "long_sword"), stashRow("5020", "long_sword"))})
	got := stashChangesFor(results, pid)

	want := []struct{ op, id string }{
		{OpStashItemRemove, "5010"},
		{OpStashItemAdd, "5020"},
		{OpStashItemAdd, "5030"},
		{OpStashGoldUpdate, ""},
	}
	if len(got) != len(want) {
		t.Fatalf("stash changes = %+v, want %d changes", got, len(want))
	}
	for i, w := range want {
		id := got[i].StashItemID
		if got[i].StashItem != nil {
			id = got[i].StashItem.StashItemID
		}
		if got[i].Op != w.op || (w.id != "" && id != w.id) {
			t.Fatalf("change %d = %s %s, want %s %s", i, got[i].Op, id, w.op, w.id)
		}
	}
	if got[3].StashGold == nil || *got[3].StashGold != 9 {
		t.Fatalf("stash gold change = %+v, want 9", got[3].StashGold)
	}
	if len(sim.stashItems) != 3 || sim.stashGold != 9 {
		t.Fatalf("sim stash after sync: %d items, gold %d", len(sim.stashItems), sim.stashGold)
	}
}

func TestAccountStashSyncUnchangedStashEmitsNothing(t *testing.T) {
	sim := MustNewSim("sess_stash_sync_noop", "01", loadRules(t))
	sim.LoadAccountStash([]PersistedStashItem{stashRow("5010", "red_potion")}, 3, 0)
	pid := sim.DefaultPlayerID()
	results := sim.TickResults([]Input{stashSyncInput(pid, 3, stashRow("5010", "red_potion"))})
	if got := stashChangesFor(results, pid); len(got) != 0 {
		t.Fatalf("unchanged stash emitted %+v", got)
	}
}

func TestAccountStashSyncAppliesToDeadPlayerAndNeverAcks(t *testing.T) {
	sim := MustNewSim("sess_stash_sync_dead", "01", loadRules(t))
	pid := sim.DefaultPlayerID()
	sim.entities[pid].hp = 0
	results := sim.TickResults([]Input{stashSyncInput(pid, 0, stashRow("5010", "red_potion"))})
	if got := stashChangesFor(results, pid); len(got) != 1 || got[0].Op != OpStashItemAdd {
		t.Fatalf("dead player sync changes = %+v, want one stash_item_add", got)
	}
	for _, res := range results {
		if len(res.Acks) != 0 || len(res.Rejects) != 0 {
			t.Fatalf("stash sync produced acks/rejects: %+v %+v", res.Acks, res.Rejects)
		}
	}
}

func TestAccountStashSyncWithActorIsIgnored(t *testing.T) {
	sim := MustNewSim("sess_stash_sync_actor", "01", loadRules(t))
	pid := sim.DefaultPlayerID()
	in := stashSyncInput(pid, 0, stashRow("5010", "red_potion"))
	in.ActorPlayerID = pid
	results := sim.TickResults([]Input{in})
	if got := stashChangesFor(results, pid); len(got) != 0 {
		t.Fatalf("actor-bearing stash sync applied: %+v", got)
	}
	if len(sim.stashItems) != 0 {
		t.Fatalf("actor-bearing stash sync changed the stash: %d items", len(sim.stashItems))
	}
}

func TestPlayerIDsForAccount(t *testing.T) {
	sim := MustNewSim("sess_stash_sync_accounts", "01", loadRules(t))
	guest, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", sim.rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatal(err)
	}
	if got := sim.PlayerIDsForAccount("acct_guest"); len(got) != 1 || got[0] != guest {
		t.Fatalf("PlayerIDsForAccount(acct_guest) = %v, want [%d]", got, guest)
	}
	if got := sim.PlayerIDsForAccount("acct_absent"); len(got) != 0 {
		t.Fatalf("absent account players = %v", got)
	}
}

// The sync is applied before player inputs, so a deposit in the same tick lands on top of the
// synced stash instead of being wiped by it.
func TestAccountStashSyncRunsBeforeSameTickDeposit(t *testing.T) {
	sim, err := NewSimWithWorld("sess_stash_sync_order", "v488_order", loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	stash := townStashEntity(t, sim)
	moveDefaultPlayerTo(sim, Vec2{X: stash.pos.X, Y: stash.pos.Y - 0.25})
	item := &invItem{instanceID: sim.alloc(), itemDefID: "red_potion"}
	sim.inventory = append(sim.inventory, item)
	sim.savePlayer(sim.defaultPlayer())
	pid := sim.DefaultPlayerID()
	sim.Tick([]Input{{MessageID: "open_stash", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(stash.id)}}})

	sim.TickResults([]Input{
		{MessageID: "deposit_item", ActorPlayerID: pid, Sequence: 1, Type: "stash_deposit_item_intent", StashDepositItem: &StashDepositItemIntent{StashEntityID: idStr(stash.id), ItemInstanceID: idStr(item.instanceID)}},
		stashSyncInput(pid, 0, stashRow("9010", "long_sword")),
	})

	defs := map[string]bool{}
	for _, it := range sim.stashItems {
		defs[it.itemDefID] = true
	}
	if !defs["long_sword"] || !defs["red_potion"] || len(sim.stashItems) != 2 {
		t.Fatalf("stash after same-tick sync + deposit = %v, want synced long_sword and deposited red_potion", defs)
	}
}
