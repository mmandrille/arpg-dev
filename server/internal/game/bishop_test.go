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
	if ev == nil || ev.Price == nil || *ev.Price != sim.rules.MainConfig.Gameplay.RespecCostGold || ev.Affordable == nil || *ev.Affordable {
		t.Fatalf("bishop service event = %+v", ev)
	}
}

func TestBishopRespecRefundsBuildForGold(t *testing.T) {
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
	if sim.gold != 50 || sim.progression.Gold != 50 {
		t.Fatalf("gold after respec sim/progression=%d/%d, want 50", sim.gold, sim.progression.Gold)
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
	if ev := findEvent(respec.Events, "bishop_respec"); ev == nil || ev.Price == nil || *ev.Price != 250 || ev.TotalGold == nil || *ev.TotalGold != 50 {
		t.Fatalf("bishop_respec event = %+v", ev)
	}
}

func TestBishopRespecRejectsWithoutGold(t *testing.T) {
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

	assertReject(t, respec, "poor_respec", "not_enough_gold")
	if sim.gold != 249 || sim.progression.BaseStats.Vit != 7 || sim.progression.SkillRanks["magic_bolt"] != 1 {
		t.Fatalf("unaffordable respec mutated state: gold=%d progression=%+v", sim.gold, sim.progression)
	}
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
