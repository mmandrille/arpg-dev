package game

import "testing"

func TestBadgeRewardChanceScalesFromUnlockDepth(t *testing.T) {
	rule := BadgeRewardRule{ResourceItemDefID: "upgrade_shard", UnlockDepth: 10, BaseChancePercent: 25, ChancePerDepthPercent: 1}
	cases := []struct {
		depth int
		want  int
	}{
		{depth: 9, want: 0},
		{depth: 10, want: 25},
		{depth: 20, want: 35},
		{depth: 100, want: 100},
	}
	for _, tc := range cases {
		if got := badgeRewardChancePercent(rule, tc.depth); got != tc.want {
			t.Fatalf("chance at depth %d = %d, want %d", tc.depth, got, tc.want)
		}
	}
}

func TestBadgeRewardValidationRejectsInvalidRows(t *testing.T) {
	gameplay := loadRules(t).MainConfig.Gameplay
	gameplay.BadgeRewardRules = []BadgeRewardRule{{ResourceItemDefID: "upgrade_shard", UnlockDepth: 0, BaseChancePercent: 25, ChancePerDepthPercent: 1}}
	if err := validateMainGameplayEconomyConfig(gameplay); err == nil {
		t.Fatalf("invalid unlock depth was accepted")
	}
	gameplay.BadgeRewardRules = []BadgeRewardRule{{ResourceItemDefID: "upgrade_shard", UnlockDepth: 10, BaseChancePercent: 101, ChancePerDepthPercent: 1}}
	if err := validateMainGameplayEconomyConfig(gameplay); err == nil {
		t.Fatalf("invalid base chance was accepted")
	}
}

func TestQuestTurnInGrantsDepthBadgeReward(t *testing.T) {
	rules := cloneRules(loadRules(t))
	forceBadgeRewardRules(rules, "respec_badge", 20)
	progression := rules.DefaultCharacterProgressionState()
	progression.DeepestDungeonDepth = 20
	sim, err := NewSimWithWorldProgression("sess_badge_quest_turn_in", "badge_quest_turn_in", rules, "vendor_lab", progression)
	if err != nil {
		t.Fatal(err)
	}
	giver := findInteractableByDefID(t, sim, "town_quest_giver")
	player := sim.activeLevel().entities[sim.playerID]
	player.pos = Vec2{X: giver.pos.X - 0.5, Y: giver.pos.Y}
	addStaticInventoryItem(sim, 29201, rules.MainConfig.Gameplay.QuestTurnInItemDefID)

	turnIn := sim.Tick([]Input{{
		MessageID:     "turn_in_badge",
		CorrelationID: "corr_turn_in_badge",
		Type:          "action_intent",
		Action:        &ActionIntent{TargetID: idStr(giver.id)},
	}})

	assertAck(t, turnIn, "turn_in_badge")
	if got := sim.resourceWallet["respec_badge"]; got != 1 {
		t.Fatalf("respec badge wallet = %d, want 1", got)
	}
	if !hasChange(turnIn, OpResourceWalletUpdate) {
		t.Fatalf("turn-in missing wallet update: %+v", turnIn.Changes)
	}
	ev := findBadgeRewardEvent(turnIn.Events, "respec_badge")
	if ev == nil || ev.Service != questTurnInService || ev.Level == nil || *ev.Level != 20 || ev.Amount == nil || *ev.Amount != 1 {
		t.Fatalf("quest badge event = %+v", ev)
	}
}

func TestBossKillGrantsDepthBadgeReward(t *testing.T) {
	rules := cloneRules(loadRules(t))
	forceBadgeRewardRules(rules, "skill_badge", 40)
	sim := MustNewSim("sess_boss_badge_reward", "boss_badge_reward", rules)
	level := sim.activeLevel()
	level.levelNum = -40
	player := level.entities[sim.playerID]
	boss := &entity{
		id:             sim.alloc(),
		kind:           monsterEntity,
		pos:            Vec2{X: player.pos.X + 1, Y: player.pos.Y},
		hp:             0,
		maxHP:          1,
		monsterDefID:   "dungeon_mob",
		lootTable:      "no_drop",
		isBoss:         true,
		bossTemplateID: "cave_warden",
	}
	level.entities[boss.id] = boss
	res := TickResult{}

	sim.finishMonsterKill(boss, sim.playerID, "corr_boss_badge", &res)

	if got := sim.resourceWallet["skill_badge"]; got != 1 {
		t.Fatalf("skill badge wallet = %d, want 1", got)
	}
	ev := findBadgeRewardEvent(res.Events, "skill_badge")
	if ev == nil || ev.BossTemplateID != "cave_warden" || ev.Level == nil || *ev.Level != 40 {
		t.Fatalf("boss badge event = %+v", ev)
	}
}

func TestConfiguredBadgesAreWalletResourceItems(t *testing.T) {
	sim := MustNewSim("sess_badge_wallet_item", "badge_wallet_item", loadRules(t))
	if !sim.isWalletResourceItem("respec_badge") || !sim.isWalletResourceItem("upgrade_shard") {
		t.Fatalf("configured badges should be wallet resources")
	}
	if sim.isWalletResourceItem("quest_leaf") {
		t.Fatalf("quest item should not be wallet resource")
	}
}

func forceBadgeRewardRules(rules *Rules, resourceID string, unlockDepth int) {
	rules.MainConfig.Gameplay.BadgeRewardRules = []BadgeRewardRule{{
		ResourceItemDefID:     resourceID,
		UnlockDepth:           unlockDepth,
		BaseChancePercent:     100,
		ChancePerDepthPercent: 0,
	}}
}

func findBadgeRewardEvent(events []Event, resourceID string) *Event {
	for i := range events {
		if events[i].EventType == badgeRewardEventType && events[i].ResourceID == resourceID {
			return &events[i]
		}
	}
	return nil
}
