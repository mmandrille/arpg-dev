package replay

import (
	"context"
	"encoding/json"
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
	rules := loadRules(t)
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

func TestVerifyUsesReconstructedSnapshot(t *testing.T) {
	rules := loadRules(t)
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
	if last.Tick < 10 {
		t.Fatalf("last timeline tick = %d, want at least 10 movement ticks", last.Tick)
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
				Equip:     &game.EquipIntent{ItemInstanceID: "1004", Slot: "weapon"},
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
		storedInput(t, "inp-equip", "msg-equip", 3, 4, "equip_intent", map[string]any{"item_instance_id": "1004", "slot": "weapon"}),
	}
}

func storedInput(t *testing.T, id, messageID string, tick, sequence int64, typ string, payload any) store.SessionInput {
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
		ID:        id,
		SessionID: testSessionID,
		Tick:      tick,
		Sequence:  sequence,
		MessageID: messageID,
		Payload:   raw,
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
	if snap.Equipped["weapon"] == nil || *snap.Equipped["weapon"] != "1004" {
		t.Fatalf("equipped weapon = %v, want 1004", snap.Equipped["weapon"])
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

type fakeRepo struct {
	session store.Session
	inputs  []store.SessionInput
	events  []store.SessionEvent
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
func (f *fakeRepo) CreateSession(context.Context, store.Session) error { return nil }
func (f *fakeRepo) GetSession(context.Context, string) (store.Session, error) {
	return f.session, nil
}
func (f *fakeRepo) TouchSession(context.Context, string) error { return nil }
func (f *fakeRepo) SetSessionStatus(context.Context, string, string) error {
	return nil
}
func (f *fakeRepo) ListInventory(context.Context, string) ([]store.InventoryItem, error) {
	return nil, nil
}
func (f *fakeRepo) AddInventoryItem(context.Context, store.InventoryItem) error { return nil }
func (f *fakeRepo) SetEquipped(context.Context, string, string, string, bool) error {
	return nil
}
func (f *fakeRepo) RemoveInventoryItem(context.Context, string, string) error { return nil }
func (f *fakeRepo) AppendInput(context.Context, store.SessionInput) error     { return nil }
func (f *fakeRepo) ListInputs(context.Context, string) ([]store.SessionInput, error) {
	return f.inputs, nil
}
func (f *fakeRepo) AppendEvent(context.Context, store.SessionEvent) error { return nil }
func (f *fakeRepo) ListEvents(context.Context, string) ([]store.SessionEvent, error) {
	return f.events, nil
}
func (f *fakeRepo) Ping(context.Context) error { return nil }
