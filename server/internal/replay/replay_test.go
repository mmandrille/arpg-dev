package replay

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

const (
	testSessionID = "sess_replay_v5"
	testSeed      = "deadbeefdeadbeef"
)

func TestReconstructFromInputsRestoresCombatStateAndMetadata(t *testing.T) {
	rules := reliableReplayHitRules(t)
	inputs, maxTick := scriptedRecordedInputs()

	recon, err := ReconstructFromInputs(testSessionID, testSeed, rules, game.DefaultWorldID, inputs, maxTick)
	if err != nil {
		t.Fatalf("reconstruct: %v", err)
	}
	if recon.Sim == nil {
		t.Fatal("reconstructed sim is nil")
	}
	assertRestoredSlice(t, recon.Snapshot)

	if recon.Metadata.NextSequence != 4 {
		t.Fatalf("next sequence = %d, want 4", recon.Metadata.NextSequence)
	}
	for _, id := range []string{"msg-move", "msg-attack", "msg-pickup", "msg-equip"} {
		if !recon.Metadata.SeenMessageIDs[id] {
			t.Fatalf("metadata missing seen message id %s", id)
		}
	}

	if !hasDerivedEvent(recon.DerivedEvents, "monster_killed") || !hasDerivedEvent(recon.DerivedEvents, "player_damaged") {
		t.Fatalf("derived events missing combat outcomes: %+v", recon.DerivedEvents)
	}
}

func TestReconstructFromInputsWithDirectionalAttackAggro(t *testing.T) {
	rules := reliableReplayHitRules(t)
	mob := rules.Monsters["dungeon_mob"]
	mob.MaxHP = 20
	rules.Monsters["dungeon_mob"] = mob
	rows := []store.SessionInput{
		storedInput(t, "inp-pick", "msg-pick", 0, 0, "action_intent", map[string]any{"target_id": "1002"}),
		storedInput(t, "inp-equip", "msg-equip", 1, 1, "equip_intent", map[string]any{"item_instance_id": "1004", "slot": "main_hand"}),
		storedInput(t, "inp-move", "msg-move", 2, 2, "move_intent", map[string]any{"direction": map[string]any{"x": 1, "y": 0}, "duration_ticks": 1}),
		storedInput(t, "inp-dir", "msg-dir", 3, 3, "directional_attack_intent", map[string]any{"direction": map[string]any{"x": 1, "y": 0}}),
	}
	inputs, _, err := StoredInputs(rows)
	if err != nil {
		t.Fatalf("stored inputs: %v", err)
	}
	recon, err := ReconstructFromInputs(testSessionID, "cafebabecafebabe", rules, "combat_control_lab", inputs, 10)
	if err != nil {
		t.Fatalf("reconstruct: %v", err)
	}
	if !hasDerivedEvent(recon.DerivedEvents, "monster_damaged") || !hasDerivedEvent(recon.DerivedEvents, "monster_aggro") {
		t.Fatalf("derived events missing directional hit/aggro: %+v", recon.DerivedEvents)
	}
	monster := findSnapshotEntity(recon.Snapshot, "monster", "dungeon_mob")
	if monster == nil {
		t.Fatal("missing dungeon_mob after replay")
	}
	if monster.HP == nil || monster.MaxHP == nil || *monster.HP >= *monster.MaxHP {
		t.Fatalf("monster hp = %v/%v, want damaged", monster.HP, monster.MaxHP)
	}
	if monster.Position.X >= 13 {
		t.Fatalf("monster position = %+v, want chased left after aggro", monster.Position)
	}
}

func TestReconstructFromInputsWithPassiveGoldAutoPickup(t *testing.T) {
	rules := loadRules(t)
	dummy := rules.Monsters["training_dummy"]
	dummy.MaxHP = 1
	dummy.LootTable = "reward_drop"
	dummy.RetaliationDamage = nil
	rules.Monsters["training_dummy"] = dummy
	rules.Combat.PlayerDamage = game.DamageRange{Min: 1, Max: 1}
	rows := []store.SessionInput{
		storedInput(t, "inp-step", "msg-step", 0, 0, "move_intent", map[string]any{"direction": map[string]any{"x": 1, "y": 0}, "duration_ticks": 1}),
		storedInput(t, "inp-kill", "msg-kill", 1, 1, "action_intent", map[string]any{"target_id": "1002"}),
		storedInput(t, "inp-step-after", "msg-step-after", 2, 2, "move_intent", map[string]any{"direction": map[string]any{"x": 1, "y": 0}, "duration_ticks": 1}),
	}
	inputs, maxTick, err := StoredInputs(rows)
	if err != nil {
		t.Fatalf("stored inputs: %v", err)
	}

	recon, err := ReconstructFromInputs(testSessionID, "v49-passive-gold-replay", rules, game.DefaultWorldID, inputs, maxTick+2)
	if err != nil {
		t.Fatalf("reconstruct: %v", err)
	}

	if !hasDerivedEvent(recon.DerivedEvents, "monster_killed") || !hasDerivedEvent(recon.DerivedEvents, "loot_dropped") || !hasDerivedEvent(recon.DerivedEvents, "gold_picked_up") {
		t.Fatalf("derived events missing kill/drop/passive pickup: %+v", recon.DerivedEvents)
	}
	if recon.Snapshot.Gold <= 0 || recon.Snapshot.CharacterProgression.Gold != recon.Snapshot.Gold {
		t.Fatalf("snapshot gold = %d progression gold = %d, want passive pickup wallet", recon.Snapshot.Gold, recon.Snapshot.CharacterProgression.Gold)
	}
	if loot := findSnapshotEntity(recon.Snapshot, "loot", "gold"); loot != nil {
		t.Fatalf("gold loot remains after passive replay pickup: %+v", loot)
	}
}

func TestReconstructFromInputsWithSkillSpendAndMagicBolt(t *testing.T) {
	rules := reliableReplayHitRules(t)
	zero := 0.0
	crit := rules.CharacterProgression.DerivedStats["crit_chance"]
	crit.Base = 0
	crit.PerDex = 0
	crit.Min = &zero
	crit.Max = &zero
	rules.CharacterProgression.DerivedStats["crit_chance"] = crit
	progress := rules.DefaultCharacterProgressionState()
	progress.Level = 3
	progress.CharacterClass = "sorcerer"
	progress.UnspentStatPoints = 6
	progress.UnspentSkillPoints = 1
	progress.BaseStats.Magic = 15
	progress.SkillRanks = map[string]int{"magic_bolt": 0}

	rows := []store.SessionInput{
		storedInput(t, "inp-skill-spend", "msg-skill-spend", 0, 0, "allocate_skill_point_intent", map[string]any{"skill_id": "magic_bolt"}),
		storedInput(t, "inp-skill-cast", "msg-skill-cast", 1, 1, "cast_skill_intent", map[string]any{"skill_id": "magic_bolt", "target_id": "1002"}),
	}
	inputs, _, err := StoredInputs(rows)
	if err != nil {
		t.Fatalf("stored inputs: %v", err)
	}
	recon, err := ReconstructFromInputsWithProgression(testSessionID, testSeed, rules, game.DefaultWorldID, inputs, 10, nil, nil, nil, progress)
	if err != nil {
		t.Fatalf("reconstruct: %v", err)
	}
	magicBolt := replaySkillRowByID(recon.Snapshot.SkillProgression.Skills, "magic_bolt")
	if recon.Snapshot.SkillProgression.UnspentSkillPoints != 0 || magicBolt == nil || magicBolt.Rank != 1 {
		t.Fatalf("skill progression snapshot = %+v, want rank 1 and no unspent points", recon.Snapshot.SkillProgression)
	}
	if len(recon.Snapshot.SkillCooldowns) == 0 || recon.Snapshot.SkillCooldowns[0].SkillID != "magic_bolt" {
		t.Fatalf("skill cooldown snapshot = %+v, want active magic_bolt cooldown", recon.Snapshot.SkillCooldowns)
	}
	if !hasDerivedEvent(recon.DerivedEvents, "skill_rank_updated") || !hasDerivedEvent(recon.DerivedEvents, "skill_cast") || !hasDerivedEvent(recon.DerivedEvents, "monster_damaged") {
		t.Fatalf("derived skill events = %+v, want rank, cast, and damage", recon.DerivedEvents)
	}
	monster := findSnapshotEntity(recon.Snapshot, "monster", "training_dummy")
	if monster == nil || monster.HP == nil || monster.MaxHP == nil || *monster.HP >= *monster.MaxHP {
		t.Fatalf("monster after magic bolt = %+v, want damaged", monster)
	}
}

func TestVerifyUsesReconstructedSnapshot(t *testing.T) {
	rules := reliableReplayHitRules(t)
	rows := scriptedStoredInputs(t)
	recorded, maxTick, err := StoredInputs(rows)
	if err != nil {
		t.Fatalf("stored inputs: %v", err)
	}
	expected, err := ReconstructFromInputs(testSessionID, testSeed, rules, game.DefaultWorldID, recorded, maxTick)
	if err != nil {
		t.Fatalf("reconstruct expected: %v", err)
	}

	repo := &fakeRepo{
		session: store.Session{ID: testSessionID, Seed: testSeed, WorldID: game.DefaultWorldID},
		inputs:  rows,
		events:  storeEvents(expected.DerivedEvents),
	}
	rep, err := Verify(context.Background(), repo, rules, testSessionID)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !rep.Match {
		t.Fatalf("verify mismatch: %s", rep.Mismatch)
	}
	if !reflect.DeepEqual(rep.Snapshot, expected.Snapshot) {
		t.Fatalf("verify snapshot differs\n got: %+v\nwant: %+v", rep.Snapshot, expected.Snapshot)
	}
	assertRestoredSlice(t, rep.Snapshot)
}

func TestBuildTimelineThroughTickExtendsPassiveSimulation(t *testing.T) {
	rules := loadRules(t)
	repo := &fakeRepo{
		session: store.Session{
			ID:      testSessionID,
			Seed:    "cafebabecafebabe",
			WorldID: "chase_maze",
		},
	}

	short, err := BuildTimeline(context.Background(), repo, rules, testSessionID, -1)
	if err != nil {
		t.Fatalf("short timeline: %v", err)
	}
	long, err := BuildTimeline(context.Background(), repo, rules, testSessionID, 30)
	if err != nil {
		t.Fatalf("long timeline: %v", err)
	}
	if len(short.Envelopes) != 1 {
		t.Fatalf("short timeline envelopes = %d, want snapshot only", len(short.Envelopes))
	}
	if len(long.Envelopes) <= len(short.Envelopes) {
		t.Fatalf("long timeline envelopes = %d, want more than short %d", len(long.Envelopes), len(short.Envelopes))
	}
	last := long.Envelopes[len(long.Envelopes)-1]
	if last.Tick < 2 {
		t.Fatalf("last timeline tick = %d, want passive movement ticks", last.Tick)
	}
}

func TestBuildTimelineIncludesGeneratedMonsterRarity(t *testing.T) {
	rules := loadRules(t)
	rules.DungeonGeneration.MonsterPlacement.ElitePackChance = 0
	sim, err := game.NewSimWithWorld(testSessionID, "v30_monster_rarity", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	actorID := sim.DefaultPlayerID()
	rows := []store.SessionInput{}
	events := []store.SessionEvent{}
	sequence := int64(0)
	tick := int64(0)
	for level := 0; level > -2; level-- {
		down := findSnapshotEntity(sim.SnapshotForPlayer(actorID), "interactable", "stairs_down")
		if down == nil {
			t.Fatalf("missing stairs_down on level %d", level)
		}
		tick = appendMoveToAndAdvanceReplay(t, sim, rules, &rows, &events, tick, &sequence, actorID, down.Position)
		tick = appendInputAndAdvanceReplay(t, sim, &rows, &events, tick, &sequence, game.Input{
			ActorPlayerID: actorID,
			Type:          "descend_intent",
			Descend:       &game.DescendIntent{},
		})
	}
	repo := &fakeRepo{
		session: store.Session{
			ID:      testSessionID,
			Seed:    "v30_monster_rarity",
			WorldID: "dungeon_levels",
		},
		inputs: rows,
	}
	timeline, err := BuildTimeline(context.Background(), repo, rules, testSessionID, tick)
	if err != nil {
		t.Fatalf("timeline: %v", err)
	}
	found := false
	for _, envelope := range timeline.Envelopes {
		if envelope.Type != "state_delta" {
			continue
		}
		delta, ok := envelope.Payload.(StateDeltaPayload)
		if !ok {
			t.Fatalf("delta payload type = %T", envelope.Payload)
		}
		for _, change := range delta.Changes {
			if change.Op == game.OpEntitySpawn && change.Entity != nil &&
				change.Entity.Type == "monster" && change.Entity.Rarity == "unique" {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("timeline did not include unique generated monster rarity: %+v", timeline.Envelopes)
	}
}

func TestVerifyBossFloorGateReplay(t *testing.T) {
	rules := loadRules(t)
	progression := game.CharacterProgressionState{
		Level: 1,
		BaseStats: game.BaseStatsView{
			Str:   200,
			Dex:   5,
			Vit:   200,
			Magic: 5,
		},
	}
	sim, err := game.NewSimWithWorldProgression(testSessionID, "boss_floor_gate", rules, "boss_floor_gate_lab", progression)
	if err != nil {
		t.Fatal(err)
	}
	actorID := sim.DefaultPlayerID()
	rows := []store.SessionInput{}
	events := []store.SessionEvent{}
	sequence := int64(0)
	tick := int64(0)

	down := findSnapshotEntity(sim.SnapshotForPlayer(actorID), "interactable", "stairs_down")
	if down == nil || down.State != "locked" {
		t.Fatalf("boss floor down = %+v, want locked", down)
	}
	tick = appendMoveToAndAdvanceReplay(t, sim, rules, &rows, &events, tick, &sequence, actorID, down.Position)
	tick = appendInputAndAdvanceReplay(t, sim, &rows, &events, tick, &sequence, game.Input{
		ActorPlayerID: actorID,
		Type:          "descend_intent",
		Descend:       &game.DescendIntent{},
	})
	if !hasStoreEvent(events, "descend_blocked") {
		t.Fatalf("missing descend_blocked in recorded events")
	}

	teleporter := findSnapshotEntity(sim.SnapshotForPlayer(actorID), "interactable", "teleporter")
	if teleporter != nil {
		t.Fatalf("boss floor teleporter = %+v, want absent", teleporter)
	}

	boss := findBossSnapshotEntity(t, sim.SnapshotForPlayer(actorID))
	for guard := 0; guard < 2000 && !hasStoreEvent(events, "monster_killed"); guard++ {
		if hasStoreEvent(events, "player_killed") {
			t.Fatalf("player died before boss kill")
		}
		if guard%3 == 0 {
			tick = appendInputAndAdvanceReplay(t, sim, &rows, &events, tick, &sequence, game.Input{
				ActorPlayerID: actorID,
				Type:          "action_intent",
				Action:        &game.ActionIntent{TargetID: boss.ID},
			})
			continue
		}
		results := sim.TickResults(nil)
		collectReplayEvents(&events, results)
		tick++
	}
	if !hasStoreEvent(events, "monster_killed") || !hasStoreEvent(events, "interactable_state_changed") {
		t.Fatalf("missing boss kill/unlock events")
	}

	down = findSnapshotEntity(sim.SnapshotForPlayer(actorID), "interactable", "stairs_down")
	if down == nil || down.State != "ready" {
		t.Fatalf("unlocked down = %+v, want ready", down)
	}
	tick = appendMoveToAndAdvanceReplay(t, sim, rules, &rows, &events, tick, &sequence, actorID, down.Position)
	_ = appendInputAndAdvanceReplay(t, sim, &rows, &events, tick, &sequence, game.Input{
		ActorPlayerID: actorID,
		Type:          "descend_intent",
		Descend:       &game.DescendIntent{},
	})

	repo := &fakeRepo{
		session: store.Session{
			ID:          testSessionID,
			AccountID:   "acct_host",
			CharacterID: "char_host",
			Seed:        "boss_floor_gate",
			WorldID:     "boss_floor_gate_lab",
		},
		inputs: rows,
		events: events,
		start: store.SessionStartSnapshot{
			SessionID:   testSessionID,
			AccountID:   "acct_host",
			CharacterID: "char_host",
			Progression: &store.CharacterProgression{
				AccountID:   "acct_host",
				CharacterID: "char_host",
				Level:       progression.Level,
				Stats: store.CharacterBaseStats{
					Str:   progression.BaseStats.Str,
					Dex:   progression.BaseStats.Dex,
					Vit:   progression.BaseStats.Vit,
					Magic: progression.BaseStats.Magic,
				},
			},
		},
	}
	rep, err := Verify(context.Background(), repo, rules, testSessionID)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !rep.Match {
		t.Fatalf("verify mismatch: %s", rep.Mismatch)
	}
	if rep.Snapshot.CurrentLevel != -6 {
		t.Fatalf("replay current level = %d, want -6", rep.Snapshot.CurrentLevel)
	}
}

func TestReconstructFromInputsUsesWorldID(t *testing.T) {
	rules := loadRules(t)
	recon, err := ReconstructFromInputs(testSessionID, testSeed, rules, "gear_before_combat", nil, -1)
	if err != nil {
		t.Fatalf("reconstruct: %v", err)
	}

	snap := recon.Snapshot
	player := entityByID(snap, "1001")
	wantPlayer := rules.Worlds["gear_before_combat"].Player.Position
	if player == nil || player.Position != wantPlayer {
		t.Fatalf("player = %+v, want 1001 at %+v", player, wantPlayer)
	}
	loot := entityByID(snap, "1002")
	if loot == nil || loot.Type != "loot" || loot.ItemDefID != "rusty_sword" {
		t.Fatalf("loot = %+v, want rusty_sword 1002", loot)
	}
	monster := entityByID(snap, "1003")
	if monster == nil || monster.Type != "monster" || monster.MonsterDefID != "training_dummy_reward" {
		t.Fatalf("monster = %+v, want training_dummy_reward 1003", monster)
	}
}

func TestReconstructLoadsSessionStartHotbarAndInputs(t *testing.T) {
	rules := loadRules(t)
	repo := &fakeRepo{
		session: store.Session{ID: testSessionID, Seed: testSeed, WorldID: game.DefaultWorldID},
		inputs: []store.SessionInput{
			storedInput(t, "inp-assign-hotbar", "msg-assign-hotbar", 0, 0, "assign_hotbar_intent", map[string]any{"slot_index": 1, "item_instance_id": "9001"}),
		},
		start: store.SessionStartSnapshot{
			Items: []store.CharacterItemInstance{{
				ID:          "9001",
				AccountID:   "acct_1",
				CharacterID: "char_1",
				ItemDefID:   "red_potion",
				Location:    store.ItemLocationInventory,
				RolledStats: json.RawMessage(`{}`),
			}},
			Hotbar: []store.CharacterHotbarSlot{
				{AccountID: "acct_1", CharacterID: "char_1", SlotIndex: 0, ItemInstanceID: stringPtr("9001")},
			},
		},
	}
	recon, err := Reconstruct(context.Background(), repo, rules, testSessionID)
	if err != nil {
		t.Fatalf("reconstruct: %v", err)
	}
	if len(recon.Snapshot.Hotbar) != 10 {
		t.Fatalf("hotbar len = %d, want 10", len(recon.Snapshot.Hotbar))
	}
	if recon.Snapshot.Hotbar[0].ItemInstanceID == nil || *recon.Snapshot.Hotbar[0].ItemInstanceID != "9001" {
		t.Fatalf("session-start hotbar[0] = %+v, want 9001", recon.Snapshot.Hotbar[0])
	}
	if recon.Snapshot.Hotbar[1].ItemInstanceID == nil || *recon.Snapshot.Hotbar[1].ItemInstanceID != "9001" {
		t.Fatalf("replayed hotbar[1] = %+v, want 9001", recon.Snapshot.Hotbar[1])
	}
}

func TestReconstructLoadsSessionStartShopStock(t *testing.T) {
	rules := loadRules(t)
	stock := []store.CharacterShopStockItem{
		shopStockFixture("generated:wp:none:000", 0, false, "cave_blade", `{"damage_min":2,"damage_max":4}`),
		shopStockFixture("generated:wp:none:001", 1, true, "cave_bow", `{"damage_min":2,"damage_max":2}`),
	}
	repo := &fakeRepo{
		session: store.Session{ID: testSessionID, AccountID: "acct_1", CharacterID: "char_1", Seed: testSeed, WorldID: "dungeon_levels"},
		start: store.SessionStartSnapshot{
			SessionID:   testSessionID,
			AccountID:   "acct_1",
			CharacterID: "char_1",
			ShopStock:   stock,
		},
	}
	scratch, _, _, err := sessionStartSim(context.Background(), repo, rules, repo.session)
	if err != nil {
		t.Fatalf("scratch sim: %v", err)
	}
	actorID := scratch.DefaultPlayerID()
	vendor := findSnapshotEntity(scratch.SnapshotForPlayer(actorID), "interactable", "town_vendor")
	if vendor == nil {
		t.Fatal("missing town vendor")
	}
	var rows []store.SessionInput
	var events []store.SessionEvent
	sequence := int64(0)
	tick := int64(0)
	tick = appendMoveToAndAdvanceReplay(t, scratch, rules, &rows, &events, tick, &sequence, actorID, vendor.Position)
	_ = appendInputAndAdvanceReplay(t, scratch, &rows, &events, tick, &sequence, game.Input{
		ActorPlayerID: actorID,
		Type:          "action_intent",
		Action:        &game.ActionIntent{TargetID: vendor.ID},
	})
	repo.inputs = rows
	repo.events = events

	recon, err := Reconstruct(context.Background(), repo, rules, testSessionID)
	if err != nil {
		t.Fatalf("reconstruct: %v", err)
	}
	opened := derivedShopEvent(t, recon.DerivedEvents, "shop_opened")
	if findReplayOffer(opened.Offers, "generated:wp:none:000") != nil {
		t.Fatalf("consumed stock offer surfaced in replay shop_opened: %+v", opened.Offers)
	}
	available := findReplayOffer(opened.Offers, "generated:wp:none:001")
	if available == nil || available.ItemTemplateID != "cave_bow" || available.SourceDepth != 1 {
		t.Fatalf("available stock offer = %+v, offers=%+v", available, opened.Offers)
	}
}

func TestReconstructCoopSessionRestoresMembersAndActorInputs(t *testing.T) {
	rules := loadRules(t)
	repo := &fakeRepo{
		session: store.Session{ID: testSessionID, Seed: testSeed, WorldID: game.DefaultWorldID, Mode: store.SessionModeCoop},
		members: []store.SessionMember{
			{
				SessionID:      testSessionID,
				AccountID:      "acct_host",
				CharacterID:    "char_host",
				PlayerEntityID: "1001",
				Role:           store.SessionMemberHost,
				Status:         store.SessionMemberActive,
				Connected:      true,
			},
			{
				SessionID:      testSessionID,
				AccountID:      "acct_guest",
				CharacterID:    "char_guest",
				PlayerEntityID: "1003",
				Role:           store.SessionMemberGuest,
				Status:         store.SessionMemberActive,
				Connected:      true,
			},
		},
		starts: map[string]store.SessionStartSnapshot{
			startKey("acct_host", "char_host"): {
				SessionID:   testSessionID,
				AccountID:   "acct_host",
				CharacterID: "char_host",
			},
			startKey("acct_guest", "char_guest"): {
				SessionID:   testSessionID,
				AccountID:   "acct_guest",
				CharacterID: "char_guest",
				Items: []store.CharacterItemInstance{{
					ID:          "9003",
					AccountID:   "acct_guest",
					CharacterID: "char_guest",
					ItemDefID:   "red_potion",
					Location:    store.ItemLocationInventory,
					RolledStats: json.RawMessage(`{}`),
				}},
			},
		},
		inputs: []store.SessionInput{
			storedInputWithActor(t, "inp-guest-hotbar", "msg-guest-hotbar", "1003", 0, 0, "assign_hotbar_intent", map[string]any{"slot_index": 1, "item_instance_id": "9003"}),
		},
	}
	recorded, _, err := StoredInputs(repo.inputs)
	if err != nil {
		t.Fatalf("stored inputs: %v", err)
	}
	if recorded[0].Input.ActorPlayerID != 1003 {
		t.Fatalf("actor player id = %d, want 1003", recorded[0].Input.ActorPlayerID)
	}

	recon, err := Reconstruct(context.Background(), repo, rules, testSessionID)
	if err != nil {
		t.Fatalf("reconstruct: %v", err)
	}
	if recon.Metadata.NextSequence != 1 || !recon.Metadata.SeenMessageIDs["msg-guest-hotbar"] {
		t.Fatalf("metadata = %+v, want replay resume after guest input", recon.Metadata)
	}
	hostSnap := recon.Sim.SnapshotForPlayer(1001)
	guestSnap := recon.Sim.SnapshotForPlayer(1003)
	if hostSnap.LocalPlayerID != "1001" || guestSnap.LocalPlayerID != "1003" {
		t.Fatalf("local players host=%q guest=%q", hostSnap.LocalPlayerID, guestSnap.LocalPlayerID)
	}
	if len(guestSnap.Party) != 2 {
		t.Fatalf("guest party = %+v, want host and guest", guestSnap.Party)
	}
	if entityByID(guestSnap, "1003") == nil || entityByID(guestSnap, "1003").CharacterID != "char_guest" {
		t.Fatalf("guest entity = %+v, want char_guest", entityByID(guestSnap, "1003"))
	}
	if guestSnap.Hotbar[1].ItemInstanceID == nil || *guestSnap.Hotbar[1].ItemInstanceID != "9003" {
		t.Fatalf("guest hotbar[1] = %+v, want 9003", guestSnap.Hotbar[1])
	}
	if hostSnap.Hotbar[1].ItemInstanceID != nil {
		t.Fatalf("host hotbar[1] = %+v, want untouched by guest input", hostSnap.Hotbar[1])
	}
}

func TestReconstructThreeMemberCoopSessionRestoresActorScopes(t *testing.T) {
	rules := loadRules(t)
	repo := &fakeRepo{
		session: store.Session{ID: testSessionID, Seed: testSeed, WorldID: game.DefaultWorldID, Mode: store.SessionModeCoop, Listed: true},
		members: []store.SessionMember{
			{SessionID: testSessionID, AccountID: "acct_host", CharacterID: "char_host", PlayerEntityID: "1001", Role: store.SessionMemberHost, Status: store.SessionMemberActive, Connected: true},
			{SessionID: testSessionID, AccountID: "acct_guest_a", CharacterID: "char_guest_a", PlayerEntityID: "1003", Role: store.SessionMemberGuest, Status: store.SessionMemberActive, Connected: true, JoinedTick: 1},
			{SessionID: testSessionID, AccountID: "acct_guest_b", CharacterID: "char_guest_b", PlayerEntityID: "1004", Role: store.SessionMemberGuest, Status: store.SessionMemberActive, Connected: true, JoinedTick: 2},
		},
		starts: map[string]store.SessionStartSnapshot{
			startKey("acct_host", "char_host"):       {SessionID: testSessionID, AccountID: "acct_host", CharacterID: "char_host"},
			startKey("acct_guest_a", "char_guest_a"): {SessionID: testSessionID, AccountID: "acct_guest_a", CharacterID: "char_guest_a"},
			startKey("acct_guest_b", "char_guest_b"): {
				SessionID:   testSessionID,
				AccountID:   "acct_guest_b",
				CharacterID: "char_guest_b",
				Items: []store.CharacterItemInstance{{
					ID:          "9100",
					AccountID:   "acct_guest_b",
					CharacterID: "char_guest_b",
					ItemDefID:   "red_potion",
					Location:    store.ItemLocationInventory,
					RolledStats: json.RawMessage(`{}`),
				}},
			},
		},
		inputs: []store.SessionInput{
			storedInputWithActor(t, "inp-guest-b-hotbar", "msg-guest-b-hotbar", "1004", 2, 0, "assign_hotbar_intent", map[string]any{"slot_index": 2, "item_instance_id": "9100"}),
			storedInputWithActor(t, "inp-guest-b-move", "msg-guest-b-move", "1004", 3, 1, "move_intent", map[string]any{"direction": map[string]any{"x": 1, "y": 0}, "duration_ticks": 1}),
		},
	}

	recon, err := Reconstruct(context.Background(), repo, rules, testSessionID)
	if err != nil {
		t.Fatalf("reconstruct: %v", err)
	}
	if recon.Metadata.NextSequence != 2 || !recon.Metadata.SeenMessageIDs["msg-guest-b-hotbar"] || !recon.Metadata.SeenMessageIDs["msg-guest-b-move"] {
		t.Fatalf("metadata = %+v, want both guest B messages", recon.Metadata)
	}
	hostSnap := recon.Sim.SnapshotForPlayer(1001)
	guestASnap := recon.Sim.SnapshotForPlayer(1003)
	guestBSnap := recon.Sim.SnapshotForPlayer(1004)
	if hostSnap.LocalPlayerID != "1001" || guestASnap.LocalPlayerID != "1003" || guestBSnap.LocalPlayerID != "1004" {
		t.Fatalf("local players host=%q guestA=%q guestB=%q", hostSnap.LocalPlayerID, guestASnap.LocalPlayerID, guestBSnap.LocalPlayerID)
	}
	if len(hostSnap.Party) != 3 || len(guestASnap.Party) != 3 || len(guestBSnap.Party) != 3 {
		t.Fatalf("party lengths host=%d guestA=%d guestB=%d", len(hostSnap.Party), len(guestASnap.Party), len(guestBSnap.Party))
	}
	if guestBSnap.Hotbar[2].ItemInstanceID == nil || *guestBSnap.Hotbar[2].ItemInstanceID != "9100" {
		t.Fatalf("guest B hotbar[2] = %+v, want 9100", guestBSnap.Hotbar[2])
	}
	if hostSnap.Hotbar[2].ItemInstanceID != nil || guestASnap.Hotbar[2].ItemInstanceID != nil {
		t.Fatalf("guest B hotbar assignment leaked host=%+v guestA=%+v", hostSnap.Hotbar[2], guestASnap.Hotbar[2])
	}
}

func TestReconstructCoopDisconnectedMemberIsRemovedForReconnect(t *testing.T) {
	rules := loadRules(t)
	repo := &fakeRepo{
		session: store.Session{ID: testSessionID, Seed: testSeed, WorldID: game.DefaultWorldID, Mode: store.SessionModeCoop},
		members: []store.SessionMember{
			{
				SessionID:      testSessionID,
				AccountID:      "acct_host",
				CharacterID:    "char_host",
				PlayerEntityID: "1001",
				Role:           store.SessionMemberHost,
				Status:         store.SessionMemberActive,
				Connected:      true,
			},
			{
				SessionID:      testSessionID,
				AccountID:      "acct_guest",
				CharacterID:    "char_guest",
				PlayerEntityID: "1003",
				Role:           store.SessionMemberGuest,
				Status:         store.SessionMemberActive,
				Connected:      false,
				CurrentLevel:   -1,
			},
		},
		starts: map[string]store.SessionStartSnapshot{
			startKey("acct_host", "char_host"):   {SessionID: testSessionID, AccountID: "acct_host", CharacterID: "char_host"},
			startKey("acct_guest", "char_guest"): {SessionID: testSessionID, AccountID: "acct_guest", CharacterID: "char_guest"},
		},
	}

	recon, err := Reconstruct(context.Background(), repo, rules, testSessionID)
	if err != nil {
		t.Fatalf("reconstruct: %v", err)
	}
	if recon.Sim.PlayerConnected(1003) {
		t.Fatal("guest should be disconnected after applying current member state")
	}
	if entityByID(recon.Sim.SnapshotForPlayer(1001), "1003") != nil {
		t.Fatalf("disconnected guest entity still visible: %+v", recon.Sim.SnapshotForPlayer(1001).Entities)
	}
	if err := recon.Sim.RespawnPlayerInTown(1003); err != nil {
		t.Fatalf("respawn guest: %v", err)
	}
	recon.Sim.SetPlayerConnected(1003, true)
	if level, ok := recon.Sim.PlayerCurrentLevel(1003); !ok || level != 0 {
		t.Fatalf("guest reconnect level = %d,%v want town", level, ok)
	}
}

func TestVerifyCoopReplayMatchesActorEventsAndLevelTransition(t *testing.T) {
	rules := loadRules(t)
	repo := &fakeRepo{
		session: store.Session{ID: testSessionID, Seed: "v33_coop_replay", WorldID: "dungeon_levels", Mode: store.SessionModeCoop},
		members: []store.SessionMember{
			{
				SessionID:      testSessionID,
				AccountID:      "acct_host",
				CharacterID:    "char_host",
				PlayerEntityID: "1001",
				Role:           store.SessionMemberHost,
				Status:         store.SessionMemberActive,
				Connected:      true,
			},
			{
				SessionID:   testSessionID,
				AccountID:   "acct_guest",
				CharacterID: "char_guest",
				Role:        store.SessionMemberGuest,
				Status:      store.SessionMemberActive,
				Connected:   true,
			},
		},
		starts: map[string]store.SessionStartSnapshot{
			startKey("acct_host", "char_host"):   {SessionID: testSessionID, AccountID: "acct_host", CharacterID: "char_host"},
			startKey("acct_guest", "char_guest"): {SessionID: testSessionID, AccountID: "acct_guest", CharacterID: "char_guest"},
		},
	}
	scratch, players, _, err := sessionStartSim(context.Background(), repo, rules, repo.session)
	if err != nil {
		t.Fatalf("scratch sim: %v", err)
	}
	guestPlayerID := playerIDForCharacter(t, players, "char_guest")
	for i := range repo.members {
		if repo.members[i].CharacterID == "char_guest" {
			repo.members[i].PlayerEntityID = fmt.Sprintf("%d", guestPlayerID)
		}
	}

	var rows []store.SessionInput
	var events []store.SessionEvent
	tick := int64(0)
	sequence := int64(0)
	tick = appendMoveToAndAdvanceReplay(t, scratch, rules, &rows, &events, tick, &sequence, guestPlayerID, game.Vec2{X: 5, Y: 10})

	stairs := findSnapshotEntity(scratch.SnapshotForPlayer(1001), "interactable", "stairs_down")
	if stairs == nil {
		t.Fatal("missing town stairs down")
	}
	tick = appendMoveToAndAdvanceReplay(t, scratch, rules, &rows, &events, tick, &sequence, 1001, stairs.Position)
	tick = appendInputAndAdvanceReplay(t, scratch, &rows, &events, tick, &sequence, game.Input{
		ActorPlayerID: 1001,
		Type:          "descend_intent",
		Descend:       &game.DescendIntent{},
	})
	if scratch.SnapshotForPlayer(1001).CurrentLevel != -1 {
		t.Fatalf("host level after descend = %d, want -1", scratch.SnapshotForPlayer(1001).CurrentLevel)
	}

	if loot := findSnapshotEntity(scratch.SnapshotForPlayer(1001), "loot", ""); loot != nil {
		tick = appendMoveToAndAdvanceReplay(t, scratch, rules, &rows, &events, tick, &sequence, 1001, loot.Position)
		tick = appendInputAndAdvanceReplay(t, scratch, &rows, &events, tick, &sequence, game.Input{
			ActorPlayerID: 1001,
			Type:          "action_intent",
			Action:        &game.ActionIntent{TargetID: loot.ID},
		})
	}

	repo.inputs = rows
	repo.events = events
	for i := range repo.members {
		if repo.members[i].CharacterID == "char_guest" {
			repo.members[i].Connected = false
			repo.members[i].CurrentLevel = 0
		}
	}

	rep, err := Verify(context.Background(), repo, rules, testSessionID)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !rep.Match {
		t.Fatalf("verify mismatch: %s", rep.Mismatch)
	}
	if entityByID(rep.Snapshot, fmt.Sprintf("%d", guestPlayerID)) != nil {
		t.Fatalf("disconnected guest should be absent from replay snapshot: %+v", rep.Snapshot.Entities)
	}
	if hasStoreEvent(events, "item_picked_up") {
		assertRecordedEventHasEntity(t, events, "item_picked_up", "1001")
	}
}

func playerIDForCharacter(t *testing.T, players []memberPlayer, characterID string) uint64 {
	t.Helper()
	for _, player := range players {
		if player.member.CharacterID == characterID {
			return player.playerID
		}
	}
	t.Fatalf("missing replay member for character %s in %+v", characterID, players)
	return 0
}

func TestVerifyCoopReplayMatchesActorCombatAndLootEvents(t *testing.T) {
	rules := reliableReplayHitRules(t)
	repo := &fakeRepo{
		session: store.Session{ID: testSessionID, Seed: "v33_coop_combat_replay", WorldID: "gear_before_combat", Mode: store.SessionModeCoop},
		members: []store.SessionMember{
			{
				SessionID:      testSessionID,
				AccountID:      "acct_host",
				CharacterID:    "char_host",
				PlayerEntityID: "1001",
				Role:           store.SessionMemberHost,
				Status:         store.SessionMemberActive,
				Connected:      true,
			},
			{
				SessionID:      testSessionID,
				AccountID:      "acct_guest",
				CharacterID:    "char_guest",
				PlayerEntityID: "1004",
				Role:           store.SessionMemberGuest,
				Status:         store.SessionMemberActive,
				Connected:      true,
			},
		},
		starts: map[string]store.SessionStartSnapshot{
			startKey("acct_host", "char_host"):   {SessionID: testSessionID, AccountID: "acct_host", CharacterID: "char_host"},
			startKey("acct_guest", "char_guest"): {SessionID: testSessionID, AccountID: "acct_guest", CharacterID: "char_guest"},
		},
	}
	scratch, _, _, err := sessionStartSim(context.Background(), repo, rules, repo.session)
	if err != nil {
		t.Fatalf("scratch sim: %v", err)
	}

	var rows []store.SessionInput
	var events []store.SessionEvent
	tick := int64(0)
	sequence := int64(0)
	tick = appendMoveToAndAdvanceReplay(t, scratch, rules, &rows, &events, tick, &sequence, 1004, game.Vec2{X: 3, Y: 5})

	loot := findSnapshotEntity(scratch.SnapshotForPlayer(1001), "loot", "")
	if loot == nil {
		t.Fatal("missing static loot")
	}
	tick = appendMoveToAndAdvanceReplay(t, scratch, rules, &rows, &events, tick, &sequence, 1001, loot.Position)
	tick = appendInputAndAdvanceReplay(t, scratch, &rows, &events, tick, &sequence, game.Input{
		ActorPlayerID: 1001,
		Type:          "action_intent",
		Action:        &game.ActionIntent{TargetID: loot.ID},
	})

	monster := findSnapshotEntity(scratch.SnapshotForPlayer(1001), "monster", "")
	if monster == nil {
		t.Fatal("missing static monster")
	}
	attackPosition := game.Vec2{X: monster.Position.X - 1, Y: monster.Position.Y}
	tick = appendMoveToAndAdvanceReplay(t, scratch, rules, &rows, &events, tick, &sequence, 1001, attackPosition)
	monsterKilled := false
	for guard := 0; guard < 20 && !monsterKilled; guard++ {
		before := len(events)
		tick = appendInputAndAdvanceReplay(t, scratch, &rows, &events, tick, &sequence, game.Input{
			ActorPlayerID: 1001,
			Type:          "action_intent",
			Action:        &game.ActionIntent{TargetID: monster.ID},
		})
		for _, ev := range events[before:] {
			if ev.EventType == "monster_killed" {
				monsterKilled = true
				break
			}
		}
	}
	if !monsterKilled {
		t.Fatalf("monster was not killed; events=%+v", events)
	}
	repo.inputs = rows
	repo.events = events

	rep, err := Verify(context.Background(), repo, rules, testSessionID)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !rep.Match {
		t.Fatalf("verify mismatch: %s", rep.Mismatch)
	}
	assertRecordedEventHasActor(t, events, "monster_killed", "1001", monster.ID)
	assertRecordedEventHasEntity(t, events, "item_picked_up", "1001")
}

func TestReconstructCoopReplaySharesXPWithNearbyGuest(t *testing.T) {
	rules := reliableReplayHitRules(t)
	dummy := rules.Monsters["training_dummy"]
	dummy.MaxHP = 1
	dummy.XPReward = 10
	dummy.RetaliationDamage = nil
	rules.Monsters["training_dummy"] = dummy

	repo := &fakeRepo{
		session: store.Session{ID: testSessionID, Seed: "v48_coop_rewards_replay", WorldID: game.DefaultWorldID, Mode: store.SessionModeCoop},
		members: []store.SessionMember{
			{SessionID: testSessionID, AccountID: "acct_host", CharacterID: "char_host", PlayerEntityID: "1001", Role: store.SessionMemberHost, Status: store.SessionMemberActive, Connected: true},
			{SessionID: testSessionID, AccountID: "acct_guest", CharacterID: "char_guest", PlayerEntityID: "1003", Role: store.SessionMemberGuest, Status: store.SessionMemberActive, Connected: true},
		},
		starts: map[string]store.SessionStartSnapshot{
			startKey("acct_host", "char_host"):   {SessionID: testSessionID, AccountID: "acct_host", CharacterID: "char_host"},
			startKey("acct_guest", "char_guest"): {SessionID: testSessionID, AccountID: "acct_guest", CharacterID: "char_guest"},
		},
	}
	scratch, _, _, err := sessionStartSim(context.Background(), repo, rules, repo.session)
	if err != nil {
		t.Fatalf("scratch sim: %v", err)
	}
	monster := findSnapshotEntity(scratch.SnapshotForPlayer(1001), "monster", "training_dummy")
	if monster == nil {
		t.Fatal("missing training dummy")
	}

	var rows []store.SessionInput
	var events []store.SessionEvent
	tick := int64(0)
	sequence := int64(0)
	tick = appendMoveToAndAdvanceReplay(t, scratch, rules, &rows, &events, tick, &sequence, 1003, game.Vec2{X: monster.Position.X - 4, Y: monster.Position.Y})
	tick = appendMoveToAndAdvanceReplay(t, scratch, rules, &rows, &events, tick, &sequence, 1001, game.Vec2{X: monster.Position.X - 1, Y: monster.Position.Y})
	_ = appendInputAndAdvanceReplay(t, scratch, &rows, &events, tick, &sequence, game.Input{
		ActorPlayerID: 1001,
		Type:          "action_intent",
		Action:        &game.ActionIntent{TargetID: monster.ID},
	})
	if countStoreEvents(events, "experience_gained") != 2 {
		t.Fatalf("recorded xp events = %+v, want host and guest xp", events)
	}

	repo.inputs = rows
	repo.events = events
	recon, err := Reconstruct(context.Background(), repo, rules, testSessionID)
	if err != nil {
		t.Fatalf("reconstruct: %v", err)
	}
	hostXP := recon.Sim.SnapshotForPlayer(1001).CharacterProgression.Experience
	guestXP := recon.Sim.SnapshotForPlayer(1003).CharacterProgression.Experience
	if hostXP != 10 || guestXP != 10 {
		t.Fatalf("replayed xp host=%d guest=%d, want both 10", hostXP, guestXP)
	}
	if !hasDerivedEvent(recon.DerivedEvents, "experience_gained") || !hasDerivedEvent(recon.DerivedEvents, "monster_killed") {
		t.Fatalf("derived events missing xp/kill: %+v", recon.DerivedEvents)
	}
}

func scriptedRecordedInputs() ([]RecordedInput, int64) {
	return []RecordedInput{
		{
			Tick: 0,
			Input: game.Input{
				MessageID: "msg-move",
				Sequence:  0,
				Type:      "move_intent",
				Move:      &game.MoveIntent{Direction: game.Vec2{X: 1}, DurationTicks: 1},
			},
		},
		{
			Tick: 1,
			Input: game.Input{
				MessageID: "msg-attack",
				Sequence:  1,
				Type:      "action_intent",
				Action:    &game.ActionIntent{TargetID: "1002"},
			},
		},
		{
			Tick: 2,
			Input: game.Input{
				MessageID: "msg-pickup",
				Sequence:  2,
				Type:      "action_intent",
				Action:    &game.ActionIntent{TargetID: "1003"},
			},
		},
		{
			Tick: 3,
			Input: game.Input{
				MessageID: "msg-equip",
				Sequence:  3,
				Type:      "equip_intent",
				Equip:     &game.EquipIntent{ItemInstanceID: "1004", Slot: "main_hand"},
			},
		},
	}, 3
}

func scriptedStoredInputs(t *testing.T) []store.SessionInput {
	t.Helper()
	return []store.SessionInput{
		storedInput(t, "inp-move", "msg-move", 0, 0, "move_intent", map[string]any{"direction": map[string]any{"x": 1, "y": 0}, "duration_ticks": 1}),
		storedInput(t, "inp-move-to", "msg-move-to", 0, 1, "move_to_intent", map[string]any{"position": map[string]any{"x": 10, "y": 5}}),
		storedInput(t, "inp-attack", "msg-attack", 1, 2, "action_intent", map[string]any{"target_id": "1002"}),
		storedInput(t, "inp-pickup", "msg-pickup", 2, 3, "action_intent", map[string]any{"target_id": "1003"}),
		storedInput(t, "inp-equip", "msg-equip", 3, 4, "equip_intent", map[string]any{"item_instance_id": "1004", "slot": "main_hand"}),
	}
}

func storedInput(t *testing.T, id, messageID string, tick, sequence int64, typ string, payload any) store.SessionInput {
	t.Helper()
	return storedInputWithActor(t, id, messageID, "", tick, sequence, typ, payload)
}

func storedInputWithActor(t *testing.T, id, messageID, actorPlayerID string, tick, sequence int64, typ string, payload any) store.SessionInput {
	t.Helper()
	raw, err := json.Marshal(map[string]any{
		"type":       typ,
		"message_id": messageID,
		"session_id": testSessionID,
		"tick":       tick,
		"payload":    payload,
	})
	if err != nil {
		t.Fatalf("marshal input: %v", err)
	}
	return store.SessionInput{
		ID:                  id,
		SessionID:           testSessionID,
		Tick:                tick,
		Sequence:            sequence,
		MessageID:           messageID,
		ActorPlayerEntityID: actorPlayerID,
		Payload:             raw,
	}
}

func storeEvents(events []derivedEvent) []store.SessionEvent {
	out := make([]store.SessionEvent, 0, len(events))
	for i, ev := range events {
		out = append(out, store.SessionEvent{
			ID:        "evt-" + string(rune('a'+i)),
			SessionID: testSessionID,
			Tick:      ev.Tick,
			Sequence:  ev.Sequence,
			EventType: ev.EventType,
			Payload:   ev.Payload,
		})
	}
	return out
}

func shopStockFixture(offerID string, index int, available bool, templateID, statsJSON string) store.CharacterShopStockItem {
	return store.CharacterShopStockItem{
		AccountID:      "acct_1",
		CharacterID:    "char_1",
		ShopID:         "town_vendor",
		RefreshKey:     "wp:none",
		OfferIndex:     index,
		OfferID:        offerID,
		SourceDepth:    1,
		ItemTemplateID: templateID,
		RolledPayload: json.RawMessage(fmt.Sprintf(
			`{"item_template_id":%q,"display_name":%q,"rarity":"common","stats":%s,"requirements":{"level":1},"effect_ids":[]}`,
			templateID,
			"Common Replay "+templateID,
			statsJSON,
		)),
		BuyPrice:  50 + index,
		Available: available,
	}
}

func derivedShopEvent(t *testing.T, events []derivedEvent, eventType string) game.Event {
	t.Helper()
	for _, ev := range events {
		if ev.EventType != eventType {
			continue
		}
		var payload game.Event
		if err := json.Unmarshal(ev.Payload, &payload); err != nil {
			t.Fatalf("unmarshal derived %s: %v", eventType, err)
		}
		return payload
	}
	t.Fatalf("missing derived event %s in %+v", eventType, events)
	return game.Event{}
}

func findReplayOffer(offers []game.ShopOfferView, offerID string) *game.ShopOfferView {
	for i := range offers {
		if offers[i].OfferID == offerID {
			return &offers[i]
		}
	}
	return nil
}

func stringPtr(v string) *string {
	return &v
}

func appendMoveToAndAdvanceReplay(
	t *testing.T,
	sim *game.Sim,
	rules *game.Rules,
	rows *[]store.SessionInput,
	events *[]store.SessionEvent,
	tick int64,
	sequence *int64,
	actorID uint64,
	pos game.Vec2,
) int64 {
	t.Helper()
	tick = appendInputAndAdvanceReplay(t, sim, rows, events, tick, sequence, game.Input{
		ActorPlayerID: actorID,
		Type:          "move_to_intent",
		MoveTo:        &game.MoveToIntent{Position: pos},
	})
	for guard := 0; guard < 2000; guard++ {
		player := entityByID(sim.SnapshotForPlayer(actorID), fmt.Sprintf("%d", actorID))
		if player != nil && replayDistance(player.Position, pos) <= replayInteractableReach(rules) {
			return tick
		}
		results := sim.TickResults(nil)
		collectReplayEvents(events, results)
		tick++
	}
	t.Fatalf("player %d did not reach %+v", actorID, pos)
	return tick
}

func replayInteractableReach(rules *game.Rules) float64 {
	return rules.Combat.UnarmedReach + 0.5 + 0.001
}

func appendInputAndAdvanceReplay(
	t *testing.T,
	sim *game.Sim,
	rows *[]store.SessionInput,
	events *[]store.SessionEvent,
	tick int64,
	sequence *int64,
	in game.Input,
) int64 {
	t.Helper()
	if in.MessageID == "" {
		in.MessageID = fmt.Sprintf("msg-%03d", *sequence)
	}
	in.Sequence = *sequence
	*sequence = *sequence + 1
	*rows = append(*rows, storedInputFromGameInput(t, fmt.Sprintf("inp-%03d", len(*rows)), tick, in))
	results := sim.TickResults([]game.Input{in})
	collectReplayEvents(events, results)
	return tick + 1
}

func collectReplayEvents(events *[]store.SessionEvent, results []game.TickResult) {
	sequence := int64(0)
	for _, res := range results {
		for _, ev := range res.Events {
			payload, _ := json.Marshal(ev)
			*events = append(*events, store.SessionEvent{
				ID:            fmt.Sprintf("evt-%03d", len(*events)),
				SessionID:     testSessionID,
				Tick:          int64(res.Tick),
				Sequence:      sequence,
				EventType:     ev.EventType,
				CorrelationID: ev.CorrelationID,
				Payload:       payload,
			})
			sequence++
		}
	}
}

func storedInputFromGameInput(t *testing.T, id string, tick int64, in game.Input) store.SessionInput {
	t.Helper()
	var payload any
	switch in.Type {
	case "move_to_intent":
		payload = map[string]any{"position": map[string]any{"x": in.MoveTo.Position.X, "y": in.MoveTo.Position.Y}}
	case "action_intent":
		payload = map[string]any{"target_id": in.Action.TargetID}
	case "descend_intent":
		payload = map[string]any{}
	default:
		t.Fatalf("unsupported replay test input type %s", in.Type)
	}
	return storedInputWithActor(t, id, in.MessageID, fmt.Sprintf("%d", in.ActorPlayerID), tick, in.Sequence, in.Type, payload)
}

func findSnapshotEntity(snap game.Snapshot, typ, defID string) *game.EntityView {
	for i := range snap.Entities {
		e := &snap.Entities[i]
		if e.Type != typ {
			continue
		}
		if defID == "" || e.InteractableDefID == defID || e.MonsterDefID == defID || e.ItemDefID == defID {
			return e
		}
	}
	return nil
}

func findBossSnapshotEntity(t *testing.T, snap game.Snapshot) *game.EntityView {
	t.Helper()
	for i := range snap.Entities {
		e := &snap.Entities[i]
		if e.Type == "monster" && e.IsBoss {
			return e
		}
	}
	t.Fatal("missing boss entity in snapshot")
	return nil
}

func assertRecordedEventHasActor(t *testing.T, events []store.SessionEvent, eventType, sourceID, targetID string) {
	t.Helper()
	for _, ev := range events {
		if ev.EventType != eventType {
			continue
		}
		var payload game.Event
		if err := json.Unmarshal(ev.Payload, &payload); err != nil {
			t.Fatalf("unmarshal event %s: %v", ev.EventType, err)
		}
		if payload.SourceEntityID == sourceID && (targetID == "" || payload.TargetEntityID == targetID) {
			return
		}
	}
	t.Fatalf("missing %s with source=%s target=%s in %+v", eventType, sourceID, targetID, events)
}

func assertRecordedEventHasEntity(t *testing.T, events []store.SessionEvent, eventType, entityID string) {
	t.Helper()
	for _, ev := range events {
		if ev.EventType != eventType {
			continue
		}
		var payload game.Event
		if err := json.Unmarshal(ev.Payload, &payload); err != nil {
			t.Fatalf("unmarshal event %s: %v", ev.EventType, err)
		}
		if payload.EntityID == entityID {
			return
		}
	}
	t.Fatalf("missing %s with entity=%s in %+v", eventType, entityID, events)
}

func hasStoreEvent(events []store.SessionEvent, eventType string) bool {
	for _, ev := range events {
		if ev.EventType == eventType {
			return true
		}
	}
	return false
}

func countStoreEvents(events []store.SessionEvent, eventType string) int {
	count := 0
	for _, ev := range events {
		if ev.EventType == eventType {
			count++
		}
	}
	return count
}

func replayDistance(a, b game.Vec2) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}

func assertRestoredSlice(t *testing.T, snap game.Snapshot) {
	t.Helper()
	if snap.ServerTick != 4 {
		t.Fatalf("server tick = %d, want 4", snap.ServerTick)
	}
	player := entityByID(snap, "1001")
	if player == nil || player.HP == nil || *player.HP >= 10 {
		t.Fatalf("player hp = %+v, want reduced below 10", player)
	}
	monster := entityByID(snap, "1002")
	if monster == nil || monster.HP == nil || *monster.HP != 0 {
		t.Fatalf("monster = %+v, want hp 0", monster)
	}
	if len(snap.Inventory) != 1 || snap.Inventory[0].ItemDefID != "rusty_sword" || !snap.Inventory[0].Equipped {
		t.Fatalf("inventory = %+v, want equipped rusty_sword", snap.Inventory)
	}
	if snap.Equipped["main_hand"] == nil || *snap.Equipped["main_hand"] != "1004" {
		t.Fatalf("equipped main_hand = %v, want 1004", snap.Equipped["main_hand"])
	}
}

func entityByID(snap game.Snapshot, id string) *game.EntityView {
	for i := range snap.Entities {
		if snap.Entities[i].ID == id {
			return &snap.Entities[i]
		}
	}
	return nil
}

func hasDerivedEvent(events []derivedEvent, typ string) bool {
	for _, ev := range events {
		if ev.EventType == typ {
			return true
		}
	}
	return false
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

func reliableReplayHitRules(t *testing.T) *game.Rules {
	t.Helper()
	rules := loadRules(t)
	forceReplayCharacterHitChance(rules, 1.0)
	forceReplayMonsterHitChance(rules, "training_dummy", 1.0)
	forceReplayMonsterHitChance(rules, "dungeon_mob", 1.0)
	return rules
}

func forceReplayCharacterHitChance(rules *game.Rules, chance float64) {
	hit := rules.CharacterProgression.DerivedStats["hit_chance"]
	hit.Base = chance
	hit.PerDex = 0
	hit.PerStr = 0
	hit.PerVit = 0
	hit.PerMagic = 0
	hit.Min = &chance
	hit.Max = &chance
	rules.CharacterProgression.DerivedStats["hit_chance"] = hit
}

func forceReplayMonsterHitChance(rules *game.Rules, monsterDefID string, chance float64) {
	def := rules.Monsters[monsterDefID]
	def.HitChance = &chance
	rules.Monsters[monsterDefID] = def
}

func replaySkillRowByID(rows []game.SkillProgressionSkillView, skillID string) *game.SkillProgressionSkillView {
	for i := range rows {
		if rows[i].SkillID == skillID {
			return &rows[i]
		}
	}
	return nil
}

type fakeRepo struct {
	session store.Session
	inputs  []store.SessionInput
	events  []store.SessionEvent
	start   store.SessionStartSnapshot
	starts  map[string]store.SessionStartSnapshot
	members []store.SessionMember
}

func (f *fakeRepo) UpsertAccountByEmail(context.Context, string, string) (store.Account, error) {
	return store.Account{}, nil
}
func (f *fakeRepo) GetAccount(context.Context, string) (store.Account, error) {
	return store.Account{}, nil
}
func (f *fakeRepo) GetOrCreateDefaultCharacter(context.Context, string, string, string) (store.Character, error) {
	return store.Character{}, nil
}
func (f *fakeRepo) GetCharacter(context.Context, string) (store.Character, error) {
	return store.Character{}, nil
}
func (f *fakeRepo) ListCharacters(context.Context, string) ([]store.CharacterSummary, error) {
	return nil, nil
}
func (f *fakeRepo) CreateCharacter(context.Context, string, string, string, string) (store.Character, error) {
	return store.Character{}, nil
}
func (f *fakeRepo) DeleteCharacter(context.Context, string, string) error { return nil }
func (f *fakeRepo) RenameCharacter(context.Context, string, string, string) (store.Character, error) {
	return store.Character{}, nil
}
func (f *fakeRepo) MarkCharacterDead(context.Context, string, string, int) error { return nil }
func (f *fakeRepo) ReviveDeadCharacters(context.Context, string) (int, error)    { return 0, nil }
func (f *fakeRepo) CreateSession(context.Context, store.Session) error           { return nil }
func (f *fakeRepo) GetSession(context.Context, string) (store.Session, error) {
	return f.session, nil
}
func (f *fakeRepo) ListActiveListedSessions(context.Context) ([]store.SessionSummary, error) {
	return nil, nil
}
func (f *fakeRepo) TouchSession(context.Context, string) error { return nil }
func (f *fakeRepo) SetSessionStatus(context.Context, string, string) error {
	return nil
}
func (f *fakeRepo) EndListedSessionIfNoConnected(context.Context, string) (bool, error) {
	return false, nil
}
func (f *fakeRepo) CreateSessionHostMember(context.Context, store.SessionMember) error {
	return nil
}
func (f *fakeRepo) CreateSessionGuestMember(context.Context, store.SessionMember) error {
	return nil
}
func (f *fakeRepo) ListSessionMembers(context.Context, string) ([]store.SessionMember, error) {
	return f.members, nil
}
func (f *fakeRepo) GetSessionMemberByAccount(context.Context, string, string) (store.SessionMember, error) {
	return store.SessionMember{}, nil
}
func (f *fakeRepo) GetSessionMember(context.Context, string, string, string) (store.SessionMember, error) {
	return store.SessionMember{}, nil
}
func (f *fakeRepo) ClaimSessionMemberConnection(context.Context, string, string, string) (bool, error) {
	return true, nil
}
func (f *fakeRepo) SetSessionMemberConnected(context.Context, string, string, string, string, int, int64) error {
	return nil
}
func (f *fakeRepo) SetSessionMemberDisconnected(context.Context, string, string, string, int, int64) error {
	return nil
}
func (f *fakeRepo) SetSessionMemberPlayer(context.Context, string, string, string, string, int) error {
	return nil
}
func (f *fakeRepo) ListCharacterItems(context.Context, string, string) ([]store.CharacterItemInstance, error) {
	return nil, nil
}
func (f *fakeRepo) ListRecoverableCharacterCorpses(context.Context, string, string) ([]store.CharacterCorpse, error) {
	return nil, nil
}
func (f *fakeRepo) TransferCorpseItemToCharacter(context.Context, string, string, string, string, string) (store.CharacterItemInstance, error) {
	return store.CharacterItemInstance{}, nil
}
func (f *fakeRepo) AddCharacterItem(context.Context, store.CharacterItemInstance) error { return nil }
func (f *fakeRepo) SetCharacterItemLocation(context.Context, string, string, string, string) error {
	return nil
}
func (f *fakeRepo) SetCharacterItemEquipped(context.Context, string, string, string, string, bool, int) error {
	return nil
}
func (f *fakeRepo) RemoveCharacterItem(context.Context, string, string, string) error { return nil }
func (f *fakeRepo) ListAccountWaypoints(context.Context, string, string) ([]store.CharacterWaypoint, error) {
	return nil, nil
}
func (f *fakeRepo) AddAccountWaypoint(context.Context, string, int) (bool, error) { return true, nil }
func (f *fakeRepo) GetOrCreateCharacterProgression(context.Context, string, string, store.CharacterProgressionDefaults) (store.CharacterProgression, error) {
	return store.CharacterProgression{}, nil
}
func (f *fakeRepo) GetCharacterProgression(context.Context, string, string) (store.CharacterProgression, error) {
	return store.CharacterProgression{}, nil
}
func (f *fakeRepo) UpsertCharacterProgression(context.Context, string, store.CharacterProgression) error {
	return nil
}
func (f *fakeRepo) SetCharacterGold(context.Context, string, string, int) error {
	return nil
}
func (f *fakeRepo) ListCharacterHotbar(context.Context, string, string) ([]store.CharacterHotbarSlot, error) {
	return nil, nil
}
func (f *fakeRepo) SetCharacterHotbarSlot(context.Context, string, string, int, *string) error {
	return nil
}
func (f *fakeRepo) GetOrCreateCharacterSkillBindings(_ context.Context, accountID, characterID string) (store.CharacterSkillBindings, error) {
	return store.CharacterSkillBindings{AccountID: accountID, CharacterID: characterID, FunctionKeys: make([]string, 16)}, nil
}
func (f *fakeRepo) SetCharacterSkillBindings(context.Context, store.CharacterSkillBindings) error {
	return nil
}
func (f *fakeRepo) ListCharacterShopStock(context.Context, string, string) ([]store.CharacterShopStockItem, error) {
	return nil, nil
}
func (f *fakeRepo) ReplaceCharacterShopStock(context.Context, string, string, string, string, []store.CharacterShopStockItem) error {
	return nil
}
func (f *fakeRepo) SetCharacterShopStockAvailable(context.Context, string, string, string, string, bool) error {
	return nil
}
func (f *fakeRepo) ListAccountStashItems(context.Context, string) ([]store.AccountStashItem, error) {
	return nil, nil
}
func (f *fakeRepo) GetOrCreateAccountStashGold(_ context.Context, accountID string) (store.AccountStashGold, error) {
	return store.AccountStashGold{AccountID: accountID}, nil
}
func (f *fakeRepo) TransferCharacterItemToAccountStash(context.Context, string, string, string, string) (store.AccountStashItem, error) {
	return store.AccountStashItem{}, nil
}
func (f *fakeRepo) TransferAccountStashItemToCharacter(context.Context, string, string, string, string) (store.CharacterItemInstance, error) {
	return store.CharacterItemInstance{}, nil
}
func (f *fakeRepo) TransferAccountStashItemToCharacterWithPlacement(context.Context, string, string, string, string, string, string, bool) (store.CharacterItemInstance, error) {
	return store.CharacterItemInstance{}, nil
}
func (f *fakeRepo) TransferCharacterGoldToAccountStash(context.Context, string, string, int) (int, int, error) {
	return 0, 0, nil
}
func (f *fakeRepo) TransferAccountStashGoldToCharacter(context.Context, string, string, int) (int, int, error) {
	return 0, 0, nil
}
func (f *fakeRepo) ListAccountResources(context.Context, string) ([]store.AccountResourceAmount, error) {
	return nil, nil
}
func (f *fakeRepo) AddAccountResource(context.Context, string, string, int) (store.AccountResourceAmount, error) {
	return store.AccountResourceAmount{}, nil
}
func (f *fakeRepo) SpendAccountResource(context.Context, string, string, int) (store.AccountResourceAmount, error) {
	return store.AccountResourceAmount{}, nil
}
func (f *fakeRepo) UpgradeAccountStashItem(context.Context, string, string, int, int, int, int, int, int, map[string]struct{}) (store.AccountStashItem, int, int, bool, error) {
	return store.AccountStashItem{}, 0, 0, true, nil
}
func (f *fakeRepo) UpgradeAccountStashItemWithWallet(context.Context, string, string, string, int, int, int, int, int, int, map[string]struct{}) (store.AccountStashItem, int, int, int, bool, error) {
	return store.AccountStashItem{}, 0, 0, 0, true, nil
}
func (f *fakeRepo) ListActiveMarketListings(context.Context) ([]store.MarketListing, error) {
	return nil, nil
}
func (f *fakeRepo) CreateMarketListingFromStash(context.Context, string, string, string, int) (store.MarketListing, error) {
	return store.MarketListing{}, nil
}
func (f *fakeRepo) CancelMarketListing(context.Context, string, string) (store.MarketListing, error) {
	return store.MarketListing{}, nil
}
func (f *fakeRepo) PurchaseMarketListing(context.Context, string, string) (store.MarketListing, error) {
	return store.MarketListing{}, nil
}
func (f *fakeRepo) CreateMarketOffer(context.Context, string, string, string, []string) (store.MarketOffer, error) {
	return store.MarketOffer{}, nil
}
func (f *fakeRepo) CancelMarketOffer(context.Context, string, string, string) (store.MarketOffer, error) {
	return store.MarketOffer{}, nil
}
func (f *fakeRepo) ListMarketOffersForSeller(context.Context, string, string) ([]store.MarketOffer, error) {
	return nil, nil
}
func (f *fakeRepo) ListMarketOffersForBidder(context.Context, string) ([]store.MarketOffer, error) {
	return nil, nil
}
func (f *fakeRepo) ListMarketAuditRecordsForAccount(context.Context, string, int) ([]store.MarketAuditRecord, error) {
	return nil, nil
}
func (f *fakeRepo) AcceptMarketOffer(context.Context, string, string, string) (store.MarketOffer, error) {
	return store.MarketOffer{}, nil
}
func (f *fakeRepo) ExpireMarketListings(context.Context) (int, error) {
	return 0, nil
}
func (f *fakeRepo) GetMarketSummary(context.Context, string) (store.MarketSummary, error) {
	return store.MarketSummary{}, nil
}
func (f *fakeRepo) CreateSessionStartSnapshot(context.Context, string, string, string, []store.CharacterItemInstance, []store.CharacterWaypoint, []store.CharacterHotbarSlot, store.CharacterSkillBindings, []store.CharacterShopStockItem, []store.AccountStashItem, store.AccountStashGold, []store.AccountResourceAmount, store.CharacterProgression) error {
	return nil
}
func (f *fakeRepo) LoadSessionStartSnapshot(context.Context, string) (store.SessionStartSnapshot, error) {
	return f.start, nil
}
func (f *fakeRepo) LoadSessionStartSnapshotForMember(_ context.Context, sessionID, accountID, characterID string) (store.SessionStartSnapshot, error) {
	if f.starts != nil {
		if snap, ok := f.starts[startKey(accountID, characterID)]; ok {
			return snap, nil
		}
	}
	if f.start.SessionID == sessionID && f.start.AccountID == accountID && f.start.CharacterID == characterID {
		return f.start, nil
	}
	return store.SessionStartSnapshot{}, store.ErrNotFound
}
func (f *fakeRepo) LoadSessionStartSnapshots(context.Context, string) ([]store.SessionStartSnapshot, error) {
	out := make([]store.SessionStartSnapshot, 0, len(f.starts))
	for _, snap := range f.starts {
		out = append(out, snap)
	}
	if len(out) == 0 && f.start.SessionID != "" {
		out = append(out, f.start)
	}
	return out, nil
}
func (f *fakeRepo) AppendInput(context.Context, store.SessionInput) error { return nil }
func (f *fakeRepo) ListInputs(context.Context, string) ([]store.SessionInput, error) {
	return f.inputs, nil
}
func (f *fakeRepo) AppendEvent(context.Context, store.SessionEvent) error { return nil }
func (f *fakeRepo) ListEvents(context.Context, string) ([]store.SessionEvent, error) {
	return f.events, nil
}
func (f *fakeRepo) Ping(context.Context) error { return nil }

func startKey(accountID, characterID string) string {
	return accountID + "/" + characterID
}
