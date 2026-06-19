package game

import "testing"

func TestQuestTurnInConsumesQuestItemAndRewardsGold(t *testing.T) {
	sim, giver := newQuestTurnInSim(t, "v291_turn_in_success")
	player := sim.activeLevel().entities[sim.playerID]
	player.pos = Vec2{X: giver.pos.X - 0.5, Y: giver.pos.Y}
	reward := sim.rules.MainConfig.Gameplay.QuestTurnInRewardGold
	sim.gold = 7
	sim.progression.Gold = 7
	item := addStaticInventoryItem(sim, 29101, sim.rules.MainConfig.Gameplay.QuestTurnInItemDefID)

	turnIn := sim.Tick([]Input{{
		MessageID:     "turn_in_quest",
		CorrelationID: "corr_turn_in",
		Type:          "action_intent",
		Action:        &ActionIntent{TargetID: idStr(giver.id)},
	}})

	assertAck(t, turnIn, "turn_in_quest")
	if sim.findItemByID(item.instanceID) != nil {
		t.Fatalf("quest item remained in inventory after turn-in")
	}
	if sim.gold != 7+reward || sim.progression.Gold != sim.gold {
		t.Fatalf("gold after turn-in sim/progression=%d/%d, want %d", sim.gold, sim.progression.Gold, 7+reward)
	}
	if !hasChange(turnIn, OpInventoryRemove) || !hasChange(turnIn, OpGoldUpdate) || !hasChange(turnIn, OpCharacterProgressionUpdate) {
		t.Fatalf("turn-in changes missing inventory/gold/progression update: %+v", turnIn.Changes)
	}
	ev := findEvent(turnIn.Events, "quest_turn_in_completed")
	if ev == nil || ev.EntityID != idStr(giver.id) || ev.Service != questTurnInService ||
		ev.ItemInstanceID != idStr(item.instanceID) || ev.Item == nil ||
		ev.Item.ItemDefID != sim.rules.MainConfig.Gameplay.QuestTurnInItemDefID ||
		ev.Amount == nil || *ev.Amount != 1 || ev.Price == nil || *ev.Price != reward ||
		ev.TotalGold == nil || *ev.TotalGold != sim.gold {
		t.Fatalf("quest_turn_in_completed = %+v", ev)
	}
}

func TestQuestTurnInRejectsMissingQuestItem(t *testing.T) {
	sim, giver := newQuestTurnInSim(t, "v291_turn_in_missing")
	player := sim.activeLevel().entities[sim.playerID]
	player.pos = Vec2{X: giver.pos.X - 0.5, Y: giver.pos.Y}
	sim.gold = 11
	sim.progression.Gold = 11
	sim.savePlayer(sim.defaultPlayer())

	turnIn := sim.Tick([]Input{{
		MessageID: "turn_in_missing",
		Type:      "action_intent",
		Action:    &ActionIntent{TargetID: idStr(giver.id)},
	}})

	assertReject(t, turnIn, "turn_in_missing", "missing_quest_item")
	if sim.gold != 11 || sim.progression.Gold != 11 {
		t.Fatalf("missing quest item mutated gold sim/progression=%d/%d", sim.gold, sim.progression.Gold)
	}
}

func TestQuestTurnInRejectsWrongServiceTarget(t *testing.T) {
	sim, _ := newQuestTurnInSim(t, "v291_turn_in_wrong_target")
	vendor := findInteractableByDefID(t, sim, "town_vendor")
	res := TickResult{}

	sim.turnInTownQuest(vendor, Input{MessageID: "wrong_target"}, &res, true)

	assertReject(t, res, "wrong_target", "invalid_target")
}

func TestQuestTurnInRewardUsesMainConfig(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.MainConfig.Gameplay.QuestTurnInRewardGold = 41
	sim, err := NewSimWithWorld("sess_v291_turn_in_config", "v291_turn_in_config", rules, "vendor_lab")
	if err != nil {
		t.Fatal(err)
	}
	giver := findInteractableByDefID(t, sim, "town_quest_giver")
	player := sim.activeLevel().entities[sim.playerID]
	player.pos = Vec2{X: giver.pos.X - 0.5, Y: giver.pos.Y}
	addStaticInventoryItem(sim, 29102, rules.MainConfig.Gameplay.QuestTurnInItemDefID)

	turnIn := sim.Tick([]Input{{
		MessageID: "turn_in_config",
		Type:      "action_intent",
		Action:    &ActionIntent{TargetID: idStr(giver.id)},
	}})

	assertAck(t, turnIn, "turn_in_config")
	if sim.gold != 41 {
		t.Fatalf("config reward gold = %d, want 41", sim.gold)
	}
}

func TestMainConfigRejectsInvalidQuestTurnInTuning(t *testing.T) {
	if err := validateMainGameplayEconomyConfig(MainGameplayConfig{QuestTurnInRewardGold: 1}); err == nil {
		t.Fatalf("empty quest turn-in item was accepted")
	}
	if err := validateMainGameplayEconomyConfig(MainGameplayConfig{QuestTurnInItemDefID: "quest_leaf", QuestTurnInRewardGold: -1}); err == nil {
		t.Fatalf("negative quest turn-in reward was accepted")
	}
}

func newQuestTurnInSim(t *testing.T, seed string) (*Sim, *entity) {
	t.Helper()
	sim, err := NewSimWithWorld("sess_"+seed, seed, loadRules(t), "vendor_lab")
	if err != nil {
		t.Fatal(err)
	}
	return sim, findInteractableByDefID(t, sim, "town_quest_giver")
}
