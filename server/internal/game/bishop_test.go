package game

import "testing"

func TestMarketBoardOpenEmitsServiceEvent(t *testing.T) {
	sim, err := NewSimWithWorld("sess_market_board_open", "v_market_board_open", loadRules(t), "vendor_lab")
	if err != nil {
		t.Fatal(err)
	}
	board := findInteractableByDefID(t, sim, "town_market_board")
	sim.activeLevel().entities[sim.playerID].pos = Vec2{X: board.pos.X - 0.5, Y: board.pos.Y}

	open := sim.Tick([]Input{{
		MessageID:     "open_market",
		CorrelationID: "corr_market_open",
		Type:          "action_intent",
		Action:        &ActionIntent{TargetID: idStr(board.id)},
	}})

	assertAck(t, open, "open_market")
	ev := findEvent(open.Events, "market_service_opened")
	if ev == nil || ev.EntityID != idStr(board.id) || ev.Service != "market" {
		t.Fatalf("market service event = %+v", ev)
	}
}

func TestBishopServiceRestoresResourcesOnOpen(t *testing.T) {
	sim, err := NewSimWithWorld("sess_bishop_open", "v92_bishop_open", loadRules(t), "vendor_lab")
	if err != nil {
		t.Fatal(err)
	}
	bishop := findInteractableByDefID(t, sim, "town_bishop")
	player := sim.activeLevel().entities[sim.playerID]
	player.pos = Vec2{X: bishop.pos.X - 0.5, Y: bishop.pos.Y}
	player.hp = player.maxHP - 2
	player.mana = player.maxMana - 2

	open := sim.Tick([]Input{{
		MessageID:     "open_bishop",
		CorrelationID: "corr_bishop_open",
		Type:          "action_intent",
		Action:        &ActionIntent{TargetID: idStr(bishop.id)},
	}})

	assertAck(t, open, "open_bishop")
	if player.hp != player.maxHP || player.mana != player.maxMana {
		t.Fatalf("resources after bishop open hp/mana=%d/%d max=%d/%d", player.hp, player.mana, player.maxHP, player.maxMana)
	}
	ev := findEvent(open.Events, "bishop_service_opened")
	wantCost := sim.bishopRespecResourceCost()
	wantAffordable := sim.gold >= sim.rules.MainConfig.Gameplay.RespecCostGold && sim.canPayBishopResourceCost(wantCost)
	if ev == nil || ev.Price == nil || *ev.Price != sim.rules.MainConfig.Gameplay.RespecCostGold || ev.Affordable == nil || *ev.Affordable != wantAffordable || ev.ResourceID != wantCost.ResourceID || ev.ResourceAmount == nil || *ev.ResourceAmount != wantCost.Count {
		t.Fatalf("bishop service event = %+v", ev)
	}
}

func TestBishopRespecConsumesBadgeAndRefundsBuild(t *testing.T) {
	sim, err := NewSimWithWorld("sess_bishop_respec", "v92_bishop_respec", loadRules(t), "vendor_lab")
	if err != nil {
		t.Fatal(err)
	}
	bishop := findInteractableByDefID(t, sim, "town_bishop")
	player := sim.activeLevel().entities[sim.playerID]
	player.pos = Vec2{X: bishop.pos.X - 0.5, Y: bishop.pos.Y}
	sim.progression.CharacterClass = "sorcerer"
	sim.progression.Level = 6
	sim.progression.BaseStats = BaseStatsView{Str: 4, Dex: 6, Vit: 7, Magic: 12}
	sim.progression.UnspentStatPoints = 4
	sim.progression.UnspentSkillPoints = 0
	sim.progression.SkillRanks = map[string]int{"magic_bolt": 2, "ice_shard": 1}
	sim.progression.Gold = 300
	sim.gold = 300
	sim.resourceWallet["respec_badge"] = 1
	player.maxHP = sCurrentMaxHP(t, sim)
	player.maxMana = sim.currentMaxMana()
	player.hp = 1
	player.mana = 1
	sim.skillCooldowns["magic_bolt"] = skillCooldownState{EndsTick: sim.tick + 5, TotalTicks: 10}
	sim.savePlayer(sim.defaultPlayer())

	respec := sim.Tick([]Input{{
		MessageID:     "respec",
		CorrelationID: "corr_bishop_respec",
		Type:          "bishop_respec_intent",
		BishopRespec:  &BishopRespecIntent{BishopEntityID: idStr(bishop.id)},
	}})

	assertAck(t, respec, "respec")
	if sim.gold != 300 || sim.progression.Gold != 300 {
		t.Fatalf("gold after respec sim/progression=%d/%d, want 300", sim.gold, sim.progression.Gold)
	}
	if got := sim.resourceWallet["respec_badge"]; got != 0 {
		t.Fatalf("respec badge after respec = %d, want 0", got)
	}
	wantStats := sim.rules.CharacterProgression.Classes["sorcerer"].BaseStats
	if sim.progression.BaseStats != wantStats {
		t.Fatalf("base stats after respec = %+v, want %+v", sim.progression.BaseStats, wantStats)
	}
	if sim.progression.UnspentStatPoints != 15 {
		t.Fatalf("unspent stat points = %d, want 15", sim.progression.UnspentStatPoints)
	}
	if sim.progression.UnspentSkillPoints != 2 {
		t.Fatalf("unspent skill points = %d, want 2", sim.progression.UnspentSkillPoints)
	}
	if len(sim.progression.SkillRanks) != 0 || len(sim.skillCooldowns) != 0 {
		t.Fatalf("skills/cooldowns after respec ranks=%+v cooldowns=%+v", sim.progression.SkillRanks, sim.skillCooldowns)
	}
	if player.hp != player.maxHP || player.mana != player.maxMana {
		t.Fatalf("resources after respec hp/mana=%d/%d max=%d/%d", player.hp, player.mana, player.maxHP, player.maxMana)
	}
	if ev := findEvent(respec.Events, "bishop_respec"); ev == nil || ev.Price == nil || *ev.Price != 0 || ev.TotalGold == nil || *ev.TotalGold != 300 || ev.ResourceID != "respec_badge" || ev.ResourceAmount == nil || *ev.ResourceAmount != 1 {
		t.Fatalf("bishop_respec event = %+v", ev)
	}
	assertResourceWalletUpdate(t, respec, "respec_badge", 0)
}

func TestBishopRespecPreservesEarnedQuestGold(t *testing.T) {
	sim, err := NewSimWithWorld("sess_bishop_respec_quest_gold", "v293_bishop_respec_quest_gold", loadRules(t), "quest_turn_in_lab")
	if err != nil {
		t.Fatal(err)
	}
	giver := findInteractableByDefID(t, sim, "town_quest_giver")
	bishop := findInteractableByDefID(t, sim, "town_bishop")
	player := sim.activeLevel().entities[sim.playerID]
	player.pos = Vec2{X: giver.pos.X - 0.5, Y: giver.pos.Y}
	sim.progression.Gold = 300
	sim.progression.DeepestDungeonDepth = 125
	sim.gold = 300
	addStaticInventoryItem(sim, 29301, sim.rules.MainConfig.Gameplay.QuestTurnInItemDefID)
	sim.savePlayer(sim.defaultPlayer())

	turnIn := sim.Tick([]Input{{
		MessageID: "turn_in_quest",
		Type:      "action_intent",
		Action:    &ActionIntent{TargetID: idStr(giver.id)},
	}})
	assertAck(t, turnIn, "turn_in_quest")
	wantGold := 300 + sim.rules.MainConfig.Gameplay.QuestTurnInRewardGold
	if sim.gold != wantGold || sim.resourceWallet["respec_badge"] != 1 {
		t.Fatalf("quest turn-in gold/wallet = %d/%+v, want gold=%d respec_badge=1", sim.gold, sim.resourceWallet, wantGold)
	}

	player.pos = Vec2{X: bishop.pos.X - 0.5, Y: bishop.pos.Y}
	respec := sim.Tick([]Input{{
		MessageID:    "respec_after_quest",
		Type:         "bishop_respec_intent",
		BishopRespec: &BishopRespecIntent{BishopEntityID: idStr(bishop.id)},
	}})

	assertAck(t, respec, "respec_after_quest")
	if sim.gold != wantGold || sim.progression.Gold != wantGold {
		t.Fatalf("gold after quest-funded respec sim/progression=%d/%d, want %d", sim.gold, sim.progression.Gold, wantGold)
	}
}

func TestBishopRespecRequiresBadge(t *testing.T) {
	sim, err := NewSimWithWorld("sess_bishop_poor", "v92_bishop_poor", loadRules(t), "vendor_lab")
	if err != nil {
		t.Fatal(err)
	}
	bishop := findInteractableByDefID(t, sim, "town_bishop")
	sim.activeLevel().entities[sim.playerID].pos = Vec2{X: bishop.pos.X - 0.5, Y: bishop.pos.Y}
	sim.progression.Level = 6
	sim.progression.BaseStats.Vit += 2
	sim.progression.SkillRanks = map[string]int{"magic_bolt": 1}
	sim.progression.Gold = 249
	sim.gold = 249
	sim.savePlayer(sim.defaultPlayer())

	respec := sim.Tick([]Input{{
		MessageID:    "poor_respec",
		Type:         "bishop_respec_intent",
		BishopRespec: &BishopRespecIntent{BishopEntityID: idStr(bishop.id)},
	}})

	assertReject(t, respec, "poor_respec", "missing_resource")
	if sim.gold != 249 || sim.progression.BaseStats.Vit == sim.rules.CharacterProgression.BaseStats.Vit || len(sim.progression.SkillRanks) == 0 {
		t.Fatalf("rejected respec mutated state: gold=%d progression=%+v", sim.gold, sim.progression)
	}
}

func TestBishopReviveAllRequiresBishopRange(t *testing.T) {
	sim, err := NewSimWithWorld("sess_bishop_revive_all", "v213_bishop_revive_all", loadRules(t), "vendor_lab")
	if err != nil {
		t.Fatal(err)
	}
	bishop := findInteractableByDefID(t, sim, "town_bishop")
	sim.activeLevel().entities[sim.playerID].pos = Vec2{X: bishop.pos.X - 0.5, Y: bishop.pos.Y}
	sim.resourceWallet["resurrection_badge"] = 1
	sim.savePlayer(sim.defaultPlayer())

	res := sim.Tick([]Input{{
		MessageID:       "revive_all",
		CorrelationID:   "corr_revive_all",
		Type:            "bishop_revive_all_intent",
		BishopReviveAll: &BishopReviveAllIntent{BishopEntityID: idStr(bishop.id)},
	}})

	assertAck(t, res, "revive_all")
	if got := sim.resourceWallet["resurrection_badge"]; got != 0 {
		t.Fatalf("resurrection badge after revive all = %d, want 0", got)
	}
	if ev := findEvent(res.Events, "bishop_revive_all"); ev == nil || ev.Service != "bishop" || ev.Amount == nil || *ev.Amount != 0 || ev.ResourceID != "resurrection_badge" || ev.ResourceAmount == nil || *ev.ResourceAmount != 1 {
		t.Fatalf("bishop_revive_all event = %+v", ev)
	}
	assertResourceWalletUpdate(t, res, "resurrection_badge", 0)
}

func TestBishopReviveAllRequiresBadge(t *testing.T) {
	sim, err := NewSimWithWorld("sess_bishop_revive_badge_required", "v213_bishop_revive_badge_required", loadRules(t), "vendor_lab")
	if err != nil {
		t.Fatal(err)
	}
	bishop := findInteractableByDefID(t, sim, "town_bishop")
	sim.activeLevel().entities[sim.playerID].pos = Vec2{X: bishop.pos.X - 0.5, Y: bishop.pos.Y}

	res := sim.Tick([]Input{{
		MessageID:       "revive_all_no_badge",
		Type:            "bishop_revive_all_intent",
		BishopReviveAll: &BishopReviveAllIntent{BishopEntityID: idStr(bishop.id)},
	}})

	assertReject(t, res, "revive_all_no_badge", "missing_resource")
}

func TestBishopDebugLevelRequiresGameplayDebug(t *testing.T) {
	sim, err := NewSimWithWorld("sess_bishop_debug_disabled", "v_bishop_debug_disabled", loadRules(t), "vendor_lab")
	if err != nil {
		t.Fatal(err)
	}
	bishop := findInteractableByDefID(t, sim, "town_bishop")
	sim.activeLevel().entities[sim.playerID].pos = Vec2{X: bishop.pos.X - 0.5, Y: bishop.pos.Y}

	res := sim.Tick([]Input{{
		MessageID:        "debug_level_disabled",
		Type:             "bishop_debug_level_intent",
		BishopDebugLevel: &BishopDebugLevelIntent{BishopEntityID: idStr(bishop.id)},
	}})

	assertReject(t, res, "debug_level_disabled", "debug_disabled")
	if sim.progression.Level != 1 || sim.progression.Experience != 0 {
		t.Fatalf("disabled debug level mutated progression: %+v", sim.progression)
	}
}

func TestBishopDebugLevelGrantsSingleLevel(t *testing.T) {
	sim, err := NewSimWithWorld("sess_bishop_debug_level", "v_bishop_debug_level", loadRules(t), "vendor_lab")
	if err != nil {
		t.Fatal(err)
	}
	sim.SetGameplayDebug(true)
	bishop := findInteractableByDefID(t, sim, "town_bishop")
	player := sim.activeLevel().entities[sim.playerID]
	player.pos = Vec2{X: bishop.pos.X - 0.5, Y: bishop.pos.Y}
	player.hp = 1
	player.mana = 0
	wantXP, ok := sim.rules.nextLevelTotalXP(1)
	if !ok {
		t.Fatal("missing level 2 xp threshold")
	}

	res := sim.Tick([]Input{{
		MessageID:        "debug_level",
		CorrelationID:    "corr_debug_level",
		Type:             "bishop_debug_level_intent",
		BishopDebugLevel: &BishopDebugLevelIntent{BishopEntityID: idStr(bishop.id)},
	}})

	assertAck(t, res, "debug_level")
	if sim.progression.Level != 2 || sim.progression.Experience != wantXP || sim.progression.UnspentStatPoints != sim.rules.CharacterProgression.PointsPerLevel {
		t.Fatalf("debug level progression = %+v, want level 2 xp %d stat points %d", sim.progression, wantXP, sim.rules.CharacterProgression.PointsPerLevel)
	}
	if player.hp != player.maxHP || player.mana != player.maxMana {
		t.Fatalf("debug level resources hp/mana=%d/%d max=%d/%d", player.hp, player.mana, player.maxHP, player.maxMana)
	}
	ev := findEvent(res.Events, "bishop_debug_level_gained")
	if ev == nil || ev.Amount == nil || *ev.Amount != wantXP || ev.FromLevel == nil || *ev.FromLevel != 1 || ev.ToLevel == nil || *ev.ToLevel != 2 {
		t.Fatalf("bishop_debug_level_gained event = %+v", ev)
	}
}

func TestBishopDebugPointsGrantOnePoint(t *testing.T) {
	sim, err := NewSimWithWorld("sess_bishop_debug_points", "v_bishop_debug_points", loadRules(t), "vendor_lab")
	if err != nil {
		t.Fatal(err)
	}
	sim.SetGameplayDebug(true)
	bishop := findInteractableByDefID(t, sim, "town_bishop")
	sim.activeLevel().entities[sim.playerID].pos = Vec2{X: bishop.pos.X - 0.5, Y: bishop.pos.Y}
	startSkillPoints := sim.progression.UnspentSkillPoints
	startStatPoints := sim.progression.UnspentStatPoints

	skill := sim.Tick([]Input{{
		MessageID:        "debug_skill",
		Type:             "bishop_debug_skill_point_intent",
		BishopDebugSkill: &BishopDebugSkillPointIntent{BishopEntityID: idStr(bishop.id)},
	}})
	assertAck(t, skill, "debug_skill")
	if sim.progression.UnspentSkillPoints != startSkillPoints+1 || sim.progression.UnspentStatPoints != startStatPoints {
		t.Fatalf("debug skill progression = %+v", sim.progression)
	}
	if ev := findEvent(skill.Events, "bishop_debug_skill_point_gained"); ev == nil || ev.Amount == nil || *ev.Amount != 1 {
		t.Fatalf("bishop_debug_skill_point_gained event = %+v", ev)
	}

	stat := sim.Tick([]Input{{
		MessageID:       "debug_stat",
		Type:            "bishop_debug_stat_point_intent",
		BishopDebugStat: &BishopDebugStatPointIntent{BishopEntityID: idStr(bishop.id)},
	}})
	assertAck(t, stat, "debug_stat")
	if sim.progression.UnspentStatPoints != startStatPoints+1 || sim.progression.UnspentSkillPoints != startSkillPoints+1 {
		t.Fatalf("debug stat progression = %+v", sim.progression)
	}
	if ev := findEvent(stat.Events, "bishop_debug_stat_point_gained"); ev == nil || ev.Amount == nil || *ev.Amount != 1 {
		t.Fatalf("bishop_debug_stat_point_gained event = %+v", ev)
	}
}

func TestBishopDebugDropUpgradeShardRequiresGameplayDebug(t *testing.T) {
	sim, err := NewSimWithWorld("sess_bishop_debug_shard_disabled", "v_bishop_debug_shard_disabled", loadRules(t), "vendor_lab")
	if err != nil {
		t.Fatal(err)
	}
	bishop := findInteractableByDefID(t, sim, "town_bishop")
	sim.activeLevel().entities[sim.playerID].pos = Vec2{X: bishop.pos.X - 0.5, Y: bishop.pos.Y}

	res := sim.Tick([]Input{{
		MessageID:                   "debug_shard_disabled",
		Type:                        "bishop_debug_drop_upgrade_shard_intent",
		BishopDebugDropUpgradeShard: &BishopDebugDropUpgradeShardIntent{BishopEntityID: idStr(bishop.id)},
	}})

	assertReject(t, res, "debug_shard_disabled", "debug_disabled")
}

func TestBishopDebugDropUpgradeShardSpawnsLoot(t *testing.T) {
	sim, err := NewSimWithWorld("sess_bishop_debug_shard", "v_bishop_debug_shard", loadRules(t), "vendor_lab")
	if err != nil {
		t.Fatal(err)
	}
	sim.SetGameplayDebug(true)
	sim.progression.DeepestDungeonDepth = 20
	bishop := findInteractableByDefID(t, sim, "town_bishop")
	sim.activeLevel().entities[sim.playerID].pos = Vec2{X: bishop.pos.X - 0.5, Y: bishop.pos.Y}
	startLoot := countEntitiesByType(sim, lootEntity)

	res := sim.Tick([]Input{{
		MessageID:                   "debug_shard",
		CorrelationID:               "corr_debug_shard",
		Type:                        "bishop_debug_drop_upgrade_shard_intent",
		BishopDebugDropUpgradeShard: &BishopDebugDropUpgradeShardIntent{BishopEntityID: idStr(bishop.id)},
	}})

	assertAck(t, res, "debug_shard")
	if countEntitiesByType(sim, lootEntity) != startLoot+1 {
		t.Fatalf("loot entities = %d, want %d", countEntitiesByType(sim, lootEntity), startLoot+1)
	}
	ev := findEvent(res.Events, "bishop_debug_upgrade_shard_dropped")
	if ev == nil || ev.TargetEntityID == "" || ev.Amount == nil || *ev.Amount < 1 {
		t.Fatalf("bishop_debug_upgrade_shard_dropped event = %+v", ev)
	}
	if findEvent(res.Events, "loot_dropped") == nil {
		t.Fatalf("missing loot_dropped: %+v", res.Events)
	}
}

func TestBishopDebugDropRenewStoneSpawnsLoot(t *testing.T) {
	sim, err := NewSimWithWorld("sess_bishop_debug_renew", "v_bishop_debug_renew", loadRules(t), "vendor_lab")
	if err != nil {
		t.Fatal(err)
	}
	sim.SetGameplayDebug(true)
	sim.progression.DeepestDungeonDepth = 20
	bishop := findInteractableByDefID(t, sim, "town_bishop")
	sim.activeLevel().entities[sim.playerID].pos = Vec2{X: bishop.pos.X - 0.5, Y: bishop.pos.Y}
	startLoot := countEntitiesByType(sim, lootEntity)

	res := sim.Tick([]Input{{
		MessageID:                 "debug_renew",
		CorrelationID:             "corr_debug_renew",
		Type:                      "bishop_debug_drop_renew_stone_intent",
		BishopDebugDropRenewStone: &BishopDebugDropRenewStoneIntent{BishopEntityID: idStr(bishop.id)},
	}})

	assertAck(t, res, "debug_renew")
	if countEntitiesByType(sim, lootEntity) != startLoot+1 {
		t.Fatalf("loot entities = %d, want %d", countEntitiesByType(sim, lootEntity), startLoot+1)
	}
	ev := findEvent(res.Events, "bishop_debug_renew_stone_dropped")
	if ev == nil || ev.TargetEntityID == "" || ev.Amount == nil || *ev.Amount < 1 {
		t.Fatalf("bishop_debug_renew_stone_dropped event = %+v", ev)
	}
}

func TestBishopDebugDropRespecBadgeSpawnsLoot(t *testing.T) {
	sim, err := NewSimWithWorld("sess_bishop_debug_respec_badge", "v_bishop_debug_respec_badge", loadRules(t), "vendor_lab")
	if err != nil {
		t.Fatal(err)
	}
	sim.SetGameplayDebug(true)
	bishop := findInteractableByDefID(t, sim, "town_bishop")
	sim.activeLevel().entities[sim.playerID].pos = Vec2{X: bishop.pos.X - 0.5, Y: bishop.pos.Y}
	startLoot := countEntitiesByType(sim, lootEntity)

	res := sim.Tick([]Input{{
		MessageID:                  "debug_respec_badge",
		CorrelationID:              "corr_debug_respec_badge",
		Type:                       "bishop_debug_drop_respec_badge_intent",
		BishopDebugDropRespecBadge: &BishopDebugDropWalletBadgeIntent{BishopEntityID: idStr(bishop.id)},
	}})

	assertAck(t, res, "debug_respec_badge")
	if countEntitiesByType(sim, lootEntity) != startLoot+1 {
		t.Fatalf("loot entities = %d, want %d", countEntitiesByType(sim, lootEntity), startLoot+1)
	}
	ev := findEvent(res.Events, "bishop_debug_respec_badge_dropped")
	if ev == nil || ev.TargetEntityID == "" || ev.Amount == nil || *ev.Amount != 1 {
		t.Fatalf("bishop_debug_respec_badge_dropped event = %+v", ev)
	}
}

func TestBishopDebugDropResurrectionBadgeSpawnsLoot(t *testing.T) {
	sim, err := NewSimWithWorld("sess_bishop_debug_resurrection_badge", "v_bishop_debug_resurrection_badge", loadRules(t), "vendor_lab")
	if err != nil {
		t.Fatal(err)
	}
	sim.SetGameplayDebug(true)
	bishop := findInteractableByDefID(t, sim, "town_bishop")
	sim.activeLevel().entities[sim.playerID].pos = Vec2{X: bishop.pos.X - 0.5, Y: bishop.pos.Y}
	startLoot := countEntitiesByType(sim, lootEntity)

	res := sim.Tick([]Input{{
		MessageID:                        "debug_resurrection_badge",
		CorrelationID:                    "corr_debug_resurrection_badge",
		Type:                             "bishop_debug_drop_resurrection_badge_intent",
		BishopDebugDropResurrectionBadge: &BishopDebugDropWalletBadgeIntent{BishopEntityID: idStr(bishop.id)},
	}})

	assertAck(t, res, "debug_resurrection_badge")
	if countEntitiesByType(sim, lootEntity) != startLoot+1 {
		t.Fatalf("loot entities = %d, want %d", countEntitiesByType(sim, lootEntity), startLoot+1)
	}
	ev := findEvent(res.Events, "bishop_debug_resurrection_badge_dropped")
	if ev == nil || ev.TargetEntityID == "" || ev.Amount == nil || *ev.Amount != 1 {
		t.Fatalf("bishop_debug_resurrection_badge_dropped event = %+v", ev)
	}
}

func countEntitiesByType(sim *Sim, kind string) int {
	count := 0
	for _, e := range sim.activeLevel().entities {
		if e.kind == kind {
			count++
		}
	}

	return count
}

func findInteractableByDefID(t *testing.T, sim *Sim, defID string) *entity {
	t.Helper()
	for _, e := range sim.activeLevel().entities {
		if e.kind == interactableEntity && e.interactableDefID == defID {
			return e
		}
	}
	t.Fatalf("missing interactable %s", defID)
	return nil
}

func sCurrentMaxHP(t *testing.T, sim *Sim) int {
	t.Helper()
	return sim.currentMaxHP()
}

func assertResourceWalletUpdate(t *testing.T, res TickResult, resourceID string, amount int) {
	t.Helper()
	for _, change := range res.Changes {
		if change.Op == OpResourceWalletUpdate && change.ResourceID == resourceID && change.ResourceAmount != nil && *change.ResourceAmount == amount {
			return
		}
	}
	t.Fatalf("missing resource wallet update %s=%d changes=%+v", resourceID, amount, res.Changes)
}
