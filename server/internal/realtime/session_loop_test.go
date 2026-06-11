package realtime

import (
	"context"
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
		{Op: game.OpShopStockReplace, ShopID: "town_vendor", RefreshKey: "wp:none", ShopStock: []game.PersistedShopStockItem{{
			ShopID:         "town_vendor",
			RefreshKey:     "wp:none",
			OfferID:        "generated:wp:none:000",
			ItemTemplateID: "cave_blade",
			Available:      true,
		}}},
		{Op: game.OpShopStockAvailability, ShopID: "town_vendor", OfferID: "generated:wp:none:000", Available: false},
		{Op: game.OpEntityUpdate, Entity: &game.EntityView{ID: "3001", Type: "monster"}},
	}
	events := []game.Event{
		{EventType: "shop_opened", EntityID: "1013", ShopID: "town_vendor", Offers: []game.ShopOfferView{{OfferID: "fixed:red_potion", Kind: "fixed", ItemDefID: "red_potion", DisplayName: "Red Potion", BuyPrice: 20}}},
		{EventType: "shop_purchase", EntityID: "1013", ShopID: "town_vendor", OfferID: "fixed:red_potion", Price: &gold},
		{EventType: "monster_aggro", EntityID: "3001"},
	}

	actorChanges := filterChangesForClient(changes, actorID, actorID)
	if len(actorChanges) != 5 {
		t.Fatalf("actor changes = %d, want 5: %+v", len(actorChanges), actorChanges)
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

func TestProgressionDeltasUseExplicitOwner(t *testing.T) {
	actorID := uint64(1001)
	ownerID := uint64(1002)
	gold := 80
	progression := game.CharacterProgressionView{Level: 2, Experience: 20, BaseStats: game.BaseStatsView{Str: 5, Dex: 5, Vit: 5, Magic: 5}}
	skills := game.SkillProgressionView{UnspentSkillPoints: 1, Skills: []game.SkillProgressionSkillView{}}
	changes := []game.Change{
		{Op: game.OpCharacterProgressionUpdate, OwnerPlayerID: ownerID, Progression: &progression},
		{Op: game.OpSkillProgressionUpdate, OwnerPlayerID: ownerID, SkillProgression: &skills},
		{Op: game.OpGoldUpdate, Gold: &gold},
		{Op: game.OpEntityUpdate, Entity: &game.EntityView{ID: "3001", Type: "monster"}},
	}
	events := []game.Event{
		{EventType: "experience_gained", EntityID: idStr(ownerID)},
		{EventType: "character_leveled", EntityID: idStr(ownerID)},
		{EventType: "skill_point_gained", EntityID: idStr(ownerID)},
		{EventType: "monster_killed", EntityID: "3001"},
	}

	actorChanges := filterChangesForClient(changes, actorID, actorID)
	if len(actorChanges) != 2 || actorChanges[0].Op != game.OpGoldUpdate || actorChanges[1].Op != game.OpEntityUpdate {
		t.Fatalf("actor changes = %+v, want actor gold plus public entity update", actorChanges)
	}
	ownerChanges := filterChangesForClient(changes, actorID, ownerID)
	if len(ownerChanges) != 3 || ownerChanges[0].Op != game.OpCharacterProgressionUpdate || ownerChanges[1].Op != game.OpSkillProgressionUpdate || ownerChanges[2].Op != game.OpEntityUpdate {
		t.Fatalf("owner changes = %+v, want owner progression/skills plus public entity update", ownerChanges)
	}
	actorEvents := filterEventsForClient(events, actorID, actorID)
	if len(actorEvents) != 1 || actorEvents[0].EventType != "monster_killed" {
		t.Fatalf("actor events = %+v, want only public monster_killed", actorEvents)
	}
	ownerEvents := filterEventsForClient(events, actorID, ownerID)
	if len(ownerEvents) != 4 {
		t.Fatalf("owner events = %+v, want all owner progression events plus public monster_killed", ownerEvents)
	}

	repo := &progressionPersistRepo{}
	loop := &sessionLoop{
		hub:  &Hub{store: repo},
		sess: store.Session{ID: "sess_owner_persist", AccountID: "acct_host", CharacterID: "char_host"},
	}
	loop.persistTick(game.TickResult{
		Tick:          7,
		ActorPlayerID: actorID,
		Changes:       []game.Change{{Op: game.OpCharacterProgressionUpdate, OwnerPlayerID: ownerID, Progression: &progression}},
	}, map[uint64]store.SessionMember{
		actorID: {AccountID: "acct_host", CharacterID: "char_host"},
		ownerID: {AccountID: "acct_guest", CharacterID: "char_guest"},
	}, 0)
	if len(repo.progressions) != 1 {
		t.Fatalf("persisted progressions = %d, want 1", len(repo.progressions))
	}
	got := repo.progressions[0]
	if got.AccountID != "acct_guest" || got.CharacterID != "char_guest" || got.Experience != 20 {
		t.Fatalf("persisted progression = %+v, want guest owner at 20 xp", got)
	}
}

func TestGoldPickupDeltasUseExplicitOwner(t *testing.T) {
	rulesDir, err := game.FindSharedRulesDir()
	if err != nil {
		t.Fatalf("find rules: %v", err)
	}
	rules, err := game.LoadRules(rulesDir)
	if err != nil {
		t.Fatalf("load rules: %v", err)
	}
	sim := game.MustNewSim("sess_gold_fanout", "v49-gold-fanout", rules)
	hostID := sim.DefaultPlayerID()
	guestID, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatalf("add guest: %v", err)
	}
	totalGold := 17
	amount := 17
	progression := game.CharacterProgressionView{Level: 1, Gold: totalGold, BaseStats: game.BaseStatsView{Str: 5, Dex: 5, Vit: 5, Magic: 5}}
	result := game.TickResult{
		Tick:          4,
		Level:         0,
		ActorPlayerID: 0,
		Changes: []game.Change{
			{Op: game.OpEntityRemove, EntityID: "3001"},
			{Op: game.OpGoldUpdate, OwnerPlayerID: guestID, Gold: &totalGold},
			{Op: game.OpCharacterProgressionUpdate, OwnerPlayerID: guestID, Progression: &progression},
		},
		Events: []game.Event{{
			EventType: "gold_picked_up",
			EntityID:  idStr(guestID),
			Amount:    &amount,
			TotalGold: &totalGold,
		}},
	}

	host := &loopClient{playerID: hostID, sendCh: make(chan outEnvelope, 8), done: make(chan struct{})}
	guest := &loopClient{playerID: guestID, sendCh: make(chan outEnvelope, 8), done: make(chan struct{})}
	loop := &sessionLoop{
		sess: store.Session{ID: "sess_gold_fanout"},
		sim:  sim,
	}
	loop.fanoutResult(result, []*loopClient{host, guest}, nil)

	hostDelta := mustReceiveDelta(t, host)
	if len(hostDelta.Changes) != 1 || hostDelta.Changes[0].Op != game.OpEntityRemove || len(hostDelta.Events) != 0 {
		t.Fatalf("host delta = %+v, want public remove only", hostDelta)
	}
	guestDelta := mustReceiveDelta(t, guest)
	if len(guestDelta.Changes) != 3 || guestDelta.Changes[0].Op != game.OpEntityRemove || guestDelta.Changes[1].Op != game.OpGoldUpdate || guestDelta.Changes[2].Op != game.OpCharacterProgressionUpdate {
		t.Fatalf("guest changes = %+v, want remove plus private gold/progression", guestDelta.Changes)
	}
	if len(guestDelta.Events) != 1 || guestDelta.Events[0].EventType != "gold_picked_up" || guestDelta.Events[0].TotalGold == nil || *guestDelta.Events[0].TotalGold != totalGold {
		t.Fatalf("guest events = %+v, want private gold_picked_up with total", guestDelta.Events)
	}

	repo := &progressionPersistRepo{}
	persistLoop := &sessionLoop{
		hub:  &Hub{store: repo},
		sess: store.Session{ID: "sess_gold_persist", AccountID: "acct_host", CharacterID: "char_host"},
	}
	persistLoop.persistTick(result, map[uint64]store.SessionMember{
		hostID:  {AccountID: "acct_host", CharacterID: "char_host"},
		guestID: {AccountID: "acct_guest", CharacterID: "char_guest"},
	}, 0)
	if len(repo.goldUpdates) != 1 {
		t.Fatalf("persisted gold updates = %d, want 1", len(repo.goldUpdates))
	}
	gotGold := repo.goldUpdates[0]
	if gotGold.accountID != "acct_guest" || gotGold.characterID != "char_guest" || gotGold.gold != totalGold {
		t.Fatalf("persisted gold = %+v, want guest gold %d", gotGold, totalGold)
	}
	if len(repo.progressions) != 1 || repo.progressions[0].AccountID != "acct_guest" || repo.progressions[0].Gold != totalGold {
		t.Fatalf("persisted progression = %+v, want guest progression gold %d", repo.progressions, totalGold)
	}
}

type progressionPersistRepo struct {
	store.Repository
	progressions []store.CharacterProgression
	goldUpdates  []persistedGoldUpdate
	events       []store.SessionEvent
}

type persistedGoldUpdate struct {
	accountID   string
	characterID string
	gold        int
}

func (r *progressionPersistRepo) UpsertCharacterProgression(_ context.Context, _ string, progression store.CharacterProgression) error {
	r.progressions = append(r.progressions, progression)
	return nil
}

func (r *progressionPersistRepo) SetCharacterGold(_ context.Context, accountID, characterID string, gold int) error {
	r.goldUpdates = append(r.goldUpdates, persistedGoldUpdate{accountID: accountID, characterID: characterID, gold: gold})
	return nil
}

func (r *progressionPersistRepo) AppendEvent(_ context.Context, event store.SessionEvent) error {
	r.events = append(r.events, event)
	return nil
}

func (r *progressionPersistRepo) TouchSession(context.Context, string) error { return nil }

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
