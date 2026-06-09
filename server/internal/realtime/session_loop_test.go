package realtime

import (
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func TestFanoutLevelTravelScopesDepartureAndArrival(t *testing.T) {
	rulesDir, err := game.FindSharedRulesDir()
	if err != nil {
		t.Fatalf("find rules: %v", err)
	}
	rules, err := game.LoadRules(rulesDir)
	if err != nil {
		t.Fatalf("load rules: %v", err)
	}
	sim, err := game.NewSimWithWorld("sess_fanout_levels", "fanout-levels", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	hostID := sim.DefaultPlayerID()
	guestID, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatalf("add guest: %v", err)
	}

	host := &loopClient{playerID: hostID, sendCh: make(chan outEnvelope, 8), done: make(chan struct{})}
	guest := &loopClient{playerID: guestID, sendCh: make(chan outEnvelope, 8), done: make(chan struct{})}
	loop := &sessionLoop{
		sess: store.Session{ID: "sess_fanout_levels"},
		sim:  sim,
	}
	clients := []*loopClient{host, guest}

	fromLevel := 0
	toLevel := -1
	stair := findEntity(t, sim.SnapshotForPlayer(guestID), "interactable", "stairs_down")
	for i := 0; i < 200; i++ {
		results := sim.TickResults([]game.Input{{
			MessageID:     "guest_move_to_stair",
			ActorPlayerID: guestID,
			Type:          "move_to_intent",
			MoveTo:        &game.MoveToIntent{Position: stair.Position},
		}})
		if hasReject(results) {
			t.Fatalf("guest move_to_stair rejected: %+v", results)
		}
		results = sim.TickResults([]game.Input{{
			MessageID:     "guest_descend_probe",
			ActorPlayerID: guestID,
			Type:          "descend_intent",
			Descend:       &game.DescendIntent{},
		}})
		if hasLevelChanged(results) {
			break
		}
		if i == 199 {
			t.Fatalf("guest did not descend after moving to stair")
		}
	}
	if level, _ := sim.PlayerCurrentLevel(guestID); level != toLevel {
		t.Fatalf("guest level = %d, want %d", level, toLevel)
	}

	loop.fanoutResult(game.TickResult{
		Tick:          1,
		Level:         fromLevel,
		ActorPlayerID: guestID,
		Changes: []game.Change{{
			Op:       game.OpEntityRemove,
			EntityID: "1002",
		}},
		Events: []game.Event{{
			EventType: "level_changed",
			FromLevel: &fromLevel,
			ToLevel:   &toLevel,
		}},
	}, clients, nil)

	hostDeparture := mustReceiveDelta(t, host)
	if got := len(hostDeparture.Changes); got != 1 {
		t.Fatalf("host departure changes = %d, want 1", got)
	}
	if hostDeparture.Changes[0].Op != game.OpEntityRemove {
		t.Fatalf("host departure op = %s, want entity_remove", hostDeparture.Changes[0].Op)
	}
	if got := len(hostDeparture.Events); got != 0 {
		t.Fatalf("host departure events = %d, want 0", got)
	}

	guestDeparture := mustReceiveDelta(t, guest)
	if got := len(guestDeparture.Changes); got != 0 {
		t.Fatalf("guest departure changes = %d, want 0", got)
	}
	if got := len(guestDeparture.Events); got != 1 {
		t.Fatalf("guest departure events = %d, want 1", got)
	}

	loop.fanoutResult(game.TickResult{
		Tick:          1,
		Level:         toLevel,
		ActorPlayerID: guestID,
		Changes: []game.Change{{
			Op: game.OpEntitySpawn,
			Entity: &game.EntityView{
				ID:       "2001",
				Type:     "monster",
				Position: game.Vec2{X: 4, Y: 4},
			},
		}},
	}, clients, nil)

	assertNoEnvelope(t, host)
	guestArrival := mustReceiveDelta(t, guest)
	if got := len(guestArrival.Changes); got != 1 {
		t.Fatalf("guest arrival changes = %d, want 1", got)
	}
	if guestArrival.Changes[0].Op != game.OpEntitySpawn {
		t.Fatalf("guest arrival op = %s, want entity_spawn", guestArrival.Changes[0].Op)
	}
}

func TestShopDeltasAreActorScoped(t *testing.T) {
	actorID := uint64(1001)
	otherID := uint64(1002)
	gold := 80
	changes := []game.Change{
		{Op: game.OpGoldUpdate, Gold: &gold},
		{Op: game.OpInventoryAdd, Item: &game.ItemView{ItemInstanceID: "2001", ItemDefID: "red_potion"}},
		{Op: game.OpEntityUpdate, Entity: &game.EntityView{ID: "3001", Type: "monster"}},
	}
	events := []game.Event{
		{EventType: "shop_opened", EntityID: "1013", ShopID: "town_vendor", Offers: []game.ShopOfferView{{OfferID: "fixed:red_potion", Kind: "fixed", ItemDefID: "red_potion", DisplayName: "Red Potion", BuyPrice: 20}}},
		{EventType: "shop_purchase", EntityID: "1013", ShopID: "town_vendor", OfferID: "fixed:red_potion", Price: &gold},
		{EventType: "monster_aggro", EntityID: "3001"},
	}

	actorChanges := filterChangesForClient(changes, actorID, actorID)
	if len(actorChanges) != 3 {
		t.Fatalf("actor changes = %d, want 3: %+v", len(actorChanges), actorChanges)
	}
	otherChanges := filterChangesForClient(changes, actorID, otherID)
	if len(otherChanges) != 1 || otherChanges[0].Op != game.OpEntityUpdate {
		t.Fatalf("other changes = %+v, want only public entity update", otherChanges)
	}

	actorEvents := filterEventsForClient(events, actorID, actorID)
	if len(actorEvents) != 3 {
		t.Fatalf("actor events = %d, want 3: %+v", len(actorEvents), actorEvents)
	}
	otherEvents := filterEventsForClient(events, actorID, otherID)
	if len(otherEvents) != 1 || otherEvents[0].EventType != "monster_aggro" {
		t.Fatalf("other events = %+v, want only public monster_aggro", otherEvents)
	}
}

func mustReceiveDelta(t *testing.T, client *loopClient) stateDeltaPayload {
	t.Helper()
	select {
	case env := <-client.sendCh:
		payload, ok := env.Payload.(stateDeltaPayload)
		if !ok {
			t.Fatalf("payload type = %T, want stateDeltaPayload", env.Payload)
		}
		return payload
	default:
		t.Fatal("missing outbound envelope")
		return stateDeltaPayload{}
	}
}

func assertNoEnvelope(t *testing.T, client *loopClient) {
	t.Helper()
	select {
	case env := <-client.sendCh:
		t.Fatalf("unexpected outbound envelope: %+v", env)
	default:
	}
}

func findEntity(t *testing.T, snap game.Snapshot, typ, defID string) game.EntityView {
	t.Helper()
	for _, entity := range snap.Entities {
		if entity.Type != typ {
			continue
		}
		if defID != "" && entity.InteractableDefID != defID && entity.ItemDefID != defID && entity.MonsterDefID != defID {
			continue
		}
		return entity
	}
	t.Fatalf("missing %s %s in snapshot: %+v", typ, defID, snap.Entities)
	return game.EntityView{}
}

func hasReject(results []game.TickResult) bool {
	for _, res := range results {
		if len(res.Rejects) > 0 {
			return true
		}
	}
	return false
}

func hasLevelChanged(results []game.TickResult) bool {
	for _, res := range results {
		for _, ev := range res.Events {
			if ev.EventType == "level_changed" {
				return true
			}
		}
	}
	return false
}
