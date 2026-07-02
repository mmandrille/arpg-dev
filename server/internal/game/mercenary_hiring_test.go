package game

import "testing"

func TestMercenaryHireCostRejectsNegativeConfig(t *testing.T) {
	err := validateMainGameplayEconomyConfig(MainGameplayConfig{MercenaryHireCostGold: -1})
	if err == nil {
		t.Fatalf("negative mercenary hire cost was accepted")
	}
}

func TestMercenaryHireCostUsesPerLevelFormula(t *testing.T) {
	rules := loadRules(t)
	rules.MainConfig.Gameplay.MercenaryHireCostGoldPerLevel = 10
	sim, err := NewSimWithWorldProgression("sess_merc_cost", "v403_merc_cost", rules, "mercenary_hiring_lab", CharacterProgressionState{
		Level: 12,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := sim.mercenaryHireCostGold(); got != 120 {
		t.Fatalf("hire cost = %d, want 120", got)
	}
}

func TestMercenaryBoardOpenWithoutCharacterRejectsHire(t *testing.T) {
	sim, board := newMercenaryHiringSim(t, "v403_open_only")
	sim.LoadMercenaryRoster([]MercenaryCharacterSnapshot{
		{
			CharacterID:    "char_alt",
			Name:           "Alt",
			CharacterClass: "barbarian",
			Level:          5,
			Progression: CharacterProgressionState{
				CharacterClass: "barbarian",
				Level:          5,
			},
		},
	})
	open := sim.Tick([]Input{mercenaryHireInput(board, "open_only")})
	assertAck(t, open, "open_only")
	if findEvent(open.Events, "mercenary_hired") != nil {
		t.Fatalf("unexpected hire on open-only action: %+v", open.Events)
	}
	if hiredMercenary(sim) != nil {
		t.Fatal("companion spawned on open-only action")
	}
}

func TestMercenaryHiringRejectsInsufficientGold(t *testing.T) {
	sim, board := newMercenaryHiringSim(t, "v403_poor_hire")
	sim.progression.Level = 10
	sim.gold = 0
	sim.progression.Gold = 0
	sim.LoadMercenaryRoster([]MercenaryCharacterSnapshot{
		{
			CharacterID:    "char_alt",
			Name:           "Alt",
			CharacterClass: "barbarian",
			Level:          5,
			Progression: CharacterProgressionState{
				CharacterClass: "barbarian",
				Level:          5,
			},
		},
	})
	hire := sim.Tick([]Input{mercenaryHireCharacterInput(board, "poor_hire", "char_alt")})
	if hire.Rejects == nil || hire.Rejects[0].Reason != "not_enough_gold" {
		t.Fatalf("reject = %+v", hire.Rejects)
	}
}

func newMercenaryHiringSim(t *testing.T, seedPrefix string) (*Sim, *entity) {
	t.Helper()
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_mercenary_hiring", seedPrefix, rules, "mercenary_hiring_lab")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	board := findInteractableByDefID(t, sim, "town_mercenary_board")
	player := sim.activeLevel().entities[sim.playerID]
	player.pos = Vec2{X: board.pos.X - 0.5, Y: board.pos.Y}
	return sim, board
}

func mercenaryHireInput(board *entity, msgID string) Input {
	return Input{
		MessageID:     msgID,
		CorrelationID: "corr_" + msgID,
		Type:          "action_intent",
		Action:        &ActionIntent{TargetID: idStr(board.id)},
	}
}

func hiredMercenary(sim *Sim) *entity {
	for _, id := range sortedEntityIDs(sim.entities) {
		e := sim.entities[id]
		if e.kind == companionEntity && e.ownerID == sim.playerID && e.sourceSkillID == mercenaryHireSourceID {
			return e
		}
	}
	return nil
}

func mercenaryLossAttacker(sim *Sim, mercenary *entity) *entity {
	attacker := &entity{
		kind:             monsterEntity,
		pos:              mercenary.pos,
		spawnPos:         mercenary.pos,
		hp:               1,
		maxHP:            1,
		monsterDefID:     "combat_lab_crit_attacker",
		monsterHitChance: 1,
	}
	attacker.id = sim.alloc()
	sim.activeLevel().entities[attacker.id] = attacker
	return attacker
}

func hasRemovedEntity(res TickResult, id uint64) bool {
	for _, change := range res.Changes {
		if change.Op == OpEntityRemove && change.EntityID == idStr(id) {
			return true
		}
	}
	return false
}
