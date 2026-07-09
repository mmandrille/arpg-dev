package game

import (
	"strconv"
	"testing"
)

func findStewardHuntSeedForTest(t *testing.T, levelNum int) (string, generatedStewardHunt) {
	t.Helper()
	rules := loadRules(t)
	for i := 0; i < 2000; i++ {
		seed := "v448_steward_probe_" + strconv.Itoa(i)
		if !dungeonLevelHasStewardHuntQuest(seed, levelNum, rules.DungeonGeneration, rules.QuestSteward) {
			continue
		}
		gen, err := GenerateDungeonLevel(seed, levelNum, rules.DungeonGeneration)
		if err != nil {
			continue
		}
		maybeAssignStewardHuntQuest(seed, rules.DungeonGeneration, rules.QuestSteward, &gen)
		if gen.stewardHunt == nil {
			continue
		}
		return seed, *gen.stewardHunt
	}
	t.Fatalf("no steward hunt seed found for level %d", levelNum)

	return "", generatedStewardHunt{}
}

func TestPinnedStewardHuntBotSeedContract(t *testing.T) {
	seed, hunt := findStewardHuntSeedForTest(t, -1)
	if seed != "v448_steward_probe_17" {
		t.Fatalf("update bot scenario pinned seed/trophy: got seed=%s trophy=%s monster=%s", seed, hunt.TrophyItemDefID, hunt.MonsterDefID)
	}
	if hunt.TrophyItemDefID != "quest_trophy_bat_wing" || hunt.MonsterDefID != "dungeon_bat" {
		t.Fatalf("pinned hunt metadata incomplete: %+v", hunt)
	}
}

func TestDungeonLevelStewardHuntRollIsDeterministic(t *testing.T) {
	rules := loadRules(t)
	seed, hunt := findStewardHuntSeedForTest(t, -1)
	if hunt.MonsterDefID == "" || hunt.TrophyItemDefID == "" {
		t.Fatalf("hunt metadata incomplete: %+v", hunt)
	}
	if !dungeonLevelHasStewardHuntQuest(seed, -1, rules.DungeonGeneration, rules.QuestSteward) {
		t.Fatalf("pinned seed %s lost steward hunt roll", seed)
	}
	genAgain, err := GenerateDungeonLevel(seed, -1, rules.DungeonGeneration)
	if err != nil {
		t.Fatal(err)
	}
	maybeAssignStewardHuntQuest(seed, rules.DungeonGeneration, rules.QuestSteward, &genAgain)
	if genAgain.stewardHunt == nil || genAgain.stewardHunt.MonsterDefID != hunt.MonsterDefID {
		t.Fatalf("regenerated hunt = %+v, want %+v", genAgain.stewardHunt, hunt)
	}
}

func TestStewardHuntStartedOnDungeonEntry(t *testing.T) {
	seed, _ := findStewardHuntSeedForTest(t, -1)
	sim, err := NewSimWithWorld("sess_steward_hunt_entry", seed, loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatal(err)
	}
	stairs := sim.findStair(sim.activeLevel(), stairsDownDefID)
	if stairs == nil {
		t.Fatal("missing town stairs down")
	}
	moveDefaultPlayerTo(sim, stairs.pos)
	res := sim.Tick([]Input{{
		MessageID:     "descend_hunt",
		CorrelationID: "corr_descend",
		Type:          "descend_intent",
		Descend:       &DescendIntent{},
	}})
	assertAck(t, res, "descend_hunt")
	ev := findEvent(res.Events, "steward_hunt_started")
	if ev == nil || ev.MonsterDefID == "" || ev.TrophyItemDefID == "" || ev.SourceDepth == nil || *ev.SourceDepth != 1 {
		t.Fatalf("steward_hunt_started = %+v", ev)
	}
}

func TestStewardHuntTrophyTurnInGrantsMagicReward(t *testing.T) {
	seed, hunt := findStewardHuntSeedForTest(t, -1)
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_steward_hunt_reward", seed, rules, "dungeon_levels")
	if err != nil {
		t.Fatal(err)
	}
	stairs := sim.findStair(sim.activeLevel(), stairsDownDefID)
	if stairs == nil {
		t.Fatal("missing town stairs down")
	}
	moveDefaultPlayerTo(sim, stairs.pos)
	descend := sim.Tick([]Input{{MessageID: "go_down", Type: "descend_intent", Descend: &DescendIntent{}}})
	assertAck(t, descend, "go_down")
	level := sim.levels[-1]
	var targetID uint64
	for _, id := range sortedEntityIDs(level.entities) {
		entity := level.entities[id]
		if entity != nil && entity.stewardHuntTarget {
			targetID = entity.id
			break
		}
	}
	if targetID == 0 {
		t.Fatal("missing steward hunt target on generated floor")
	}
	target := level.entities[targetID]
	player := level.entities[sim.playerID]
	player.pos = target.pos
	killRes := TickResult{}
	sim.finishMonsterKill(target, sim.playerID, "corr_kill", &killRes)
	if level.stewardHunt == nil || !level.stewardHunt.Complete {
		t.Fatal("steward hunt not marked complete after target kill")
	}
	trophy := addStaticInventoryItem(sim, 44801, hunt.TrophyItemDefID)
	trophy.questSourceDepth = 1
	stairsUp := sim.findStair(sim.activeLevel(), stairsUpDefID)
	if stairsUp == nil {
		t.Fatal("missing stairs up")
	}
	moveDefaultPlayerTo(sim, stairsUp.pos)
	ascend := sim.Tick([]Input{{MessageID: "go_up", Type: "ascend_intent", Ascend: &AscendIntent{}}})
	assertAck(t, ascend, "go_up")
	giver := findInteractableByDefID(t, sim, "town_quest_giver")
	player.pos = Vec2{X: giver.pos.X - 0.5, Y: giver.pos.Y}
	open := sim.Tick([]Input{{
		MessageID:     "open_offers",
		CorrelationID: "corr_open",
		Type:          "action_intent",
		Action:        &ActionIntent{TargetID: idStr(giver.id)},
	}})
	assertAck(t, open, "open_offers")
	offersEv := findEvent(open.Events, "quest_steward_offers_opened")
	if offersEv == nil || len(offersEv.QuestStewardOffers) != rules.QuestSteward.HuntQuest.ChoiceCount {
		t.Fatalf("quest_steward_offers_opened = %+v", offersEv)
	}
	offerID := offersEv.QuestStewardOffers[0].OfferID
	pick := sim.Tick([]Input{{
		MessageID:     "pick_reward",
		CorrelationID: "corr_pick",
		Type:          "quest_steward_pick_intent",
		QuestStewardPick: &QuestStewardPickIntent{
			QuestGiverEntityID: idStr(giver.id),
			OfferID:            offerID,
		},
	}})
	assertAck(t, pick, "pick_reward")
	if sim.findItemByID(trophy.instanceID) != nil {
		t.Fatal("trophy remained after pick")
	}
	rewardEv := findEvent(pick.Events, "quest_steward_reward_granted")
	if rewardEv == nil || rewardEv.FamilyID == "" || rewardEv.SourceDepth == nil || *rewardEv.SourceDepth != 1 {
		t.Fatalf("quest_steward_reward_granted = %+v", rewardEv)
	}
	if rewardEv.Item == nil || itemRarityRank(rewardEv.Item.Rarity) < itemRarityRank(rules.QuestSteward.HuntQuest.MinRarity) {
		t.Fatalf("reward rarity too low: %+v", rewardEv.Item)
	}
}

func TestQuestStewardRewardFamilyCanGrantNamedUnique(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_steward_unique_reward", "steward_unique_reward", rules)
	reward, ok := sim.rollQuestStewardReward("warbrand", 5, 44901)
	if !ok {
		t.Fatal("rollQuestStewardReward returned false for warbrand")
	}
	if reward.NamedUniqueID != "warbrand_cleaver" || reward.DisplayName != "Warbrand Cleaver" {
		t.Fatalf("reward = %+v, want named Warbrand Cleaver", reward)
	}
}

func TestQuestTurnInConsumesQuestLeafFromResourceBag(t *testing.T) {
	TestQuestTurnInConsumesQuestItemAndRewardsGold(t)
}

func TestQuestStewardTrophyTurnInFromResourceBag(t *testing.T) {
	seed, hunt := findStewardHuntSeedForTest(t, -1)
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_steward_bag_turn_in", seed, rules, "dungeon_levels")
	if err != nil {
		t.Fatal(err)
	}
	sim.progression.DeepestDungeonDepth = 3
	bagItem := addTestResourceBagItem(sim, 44802, hunt.TrophyItemDefID)
	giver := findInteractableByDefID(t, sim, "town_quest_giver")
	player := sim.activeLevel().entities[sim.playerID]
	player.pos = Vec2{X: giver.pos.X - 0.5, Y: giver.pos.Y}

	open := sim.Tick([]Input{{
		MessageID:     "open_bag_offers",
		CorrelationID: "corr_open_bag",
		Type:          "action_intent",
		Action:        &ActionIntent{TargetID: idStr(giver.id)},
	}})
	assertAck(t, open, "open_bag_offers")
	offersEv := findEvent(open.Events, "quest_steward_offers_opened")
	if offersEv == nil || len(offersEv.QuestStewardOffers) != rules.QuestSteward.HuntQuest.ChoiceCount {
		t.Fatalf("quest_steward_offers_opened = %+v", offersEv)
	}
	if sim.findResourceBagItem(idStr(bagItem.stashItemID)) == nil {
		t.Fatal("trophy should remain in resource bag until pick")
	}
	offerID := offersEv.QuestStewardOffers[0].OfferID
	pick := sim.Tick([]Input{{
		MessageID:     "pick_bag_reward",
		CorrelationID: "corr_pick_bag",
		Type:          "quest_steward_pick_intent",
		QuestStewardPick: &QuestStewardPickIntent{
			QuestGiverEntityID: idStr(giver.id),
			OfferID:            offerID,
		},
	}})
	assertAck(t, pick, "pick_bag_reward")
	if sim.findResourceBagItem(idStr(bagItem.stashItemID)) != nil {
		t.Fatal("trophy remained in resource bag after pick")
	}
	if findEvent(pick.Events, "quest_steward_reward_granted") == nil {
		t.Fatalf("quest_steward_reward_granted = %+v", pick.Events)
	}
}

func TestQuestStewardTrophyTurnInWithoutSourceDepth(t *testing.T) {
	seed, hunt := findStewardHuntSeedForTest(t, -1)
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_steward_no_depth", seed, rules, "dungeon_levels")
	if err != nil {
		t.Fatal(err)
	}
	sim.progression.DeepestDungeonDepth = 4
	trophy := addStaticInventoryItem(sim, 44803, hunt.TrophyItemDefID)
	giver := findInteractableByDefID(t, sim, "town_quest_giver")
	player := sim.activeLevel().entities[sim.playerID]
	player.pos = Vec2{X: giver.pos.X - 0.5, Y: giver.pos.Y}

	open := sim.Tick([]Input{{
		MessageID: "open_no_depth",
		Type:      "action_intent",
		Action:    &ActionIntent{TargetID: idStr(giver.id)},
	}})
	assertAck(t, open, "open_no_depth")
	offersEv := findEvent(open.Events, "quest_steward_offers_opened")
	if offersEv == nil || offersEv.SourceDepth == nil || *offersEv.SourceDepth != 4 {
		t.Fatalf("quest_steward_offers_opened = %+v, want source depth 4", offersEv)
	}
	if trophy.questSourceDepth != 0 {
		t.Fatalf("test setup should not set quest source depth")
	}
}
