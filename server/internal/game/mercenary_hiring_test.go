package game

import "testing"

func TestMercenaryHireCostRejectsNegativeConfig(t *testing.T) {
	err := validateMainGameplayEconomyConfig(MainGameplayConfig{MercenaryHireCostGold: -1})
	if err == nil {
		t.Fatalf("negative mercenary hire cost was accepted")
	}
}

func TestMercenaryHiringSpendsGoldAndSpawnsCompanion(t *testing.T) {
	sim, board := newMercenaryHiringSim(t, "v206_merc_hire_success")
	cost := sim.rules.MainConfig.Gameplay.MercenaryHireCostGold
	sim.gold = cost
	sim.progression.Gold = cost
	sim.savePlayer(sim.defaultPlayer())

	hire := sim.Tick([]Input{mercenaryHireInput(board, "hire_mercenary")})

	assertAck(t, hire, "hire_mercenary")
	if sim.gold != 0 || sim.progression.Gold != 0 {
		t.Fatalf("gold after hire sim/progression=%d/%d, want 0/0", sim.gold, sim.progression.Gold)
	}
	if !hasChange(hire, OpGoldUpdate) || !hasChange(hire, OpCharacterProgressionUpdate) {
		t.Fatalf("hire changes missing gold/progression update: %+v", hire.Changes)
	}
	opened := findEvent(hire.Events, "mercenary_board_opened")
	if opened == nil || opened.EntityID != idStr(board.id) || opened.Service != mercenaryService ||
		opened.OfferID != mercenaryGuardOfferID || opened.MonsterDefID != mercenaryGuardMonsterDefID ||
		opened.Price == nil || *opened.Price != cost || opened.Affordable == nil || !*opened.Affordable ||
		opened.TotalGold == nil || *opened.TotalGold != cost {
		t.Fatalf("mercenary_board_opened = %+v", opened)
	}
	hired := findEvent(hire.Events, "mercenary_hired")
	mercenary := hiredMercenary(sim)
	if hired == nil || mercenary == nil || hired.EntityID != idStr(board.id) || hired.TargetEntityID != idStr(mercenary.id) ||
		hired.Service != mercenaryService || hired.OfferID != mercenaryGuardOfferID ||
		hired.MonsterDefID != mercenaryGuardMonsterDefID || hired.Price == nil || *hired.Price != cost ||
		hired.TotalGold == nil || *hired.TotalGold != 0 {
		t.Fatalf("mercenary_hired=%+v mercenary=%+v", hired, mercenary)
	}
	if mercenary.ownerID != sim.playerID || mercenary.sourceSkillID != mercenaryHireSourceID {
		t.Fatalf("mercenary owner/source=%d/%s, want %d/%s", mercenary.ownerID, mercenary.sourceSkillID, sim.playerID, mercenaryHireSourceID)
	}
	def := sim.rules.Monsters[mercenaryGuardMonsterDefID]
	if mercenary.hp != def.MaxHP || mercenary.maxHP != def.MaxHP || mercenary.monsterAttackDamage == nil || mercenary.monsterAttackDamage.Max != def.AttackDamage.Max {
		t.Fatalf("mercenary stats hp=%d/%d damage=%+v, want hp %d damage %+v", mercenary.hp, mercenary.maxHP, mercenary.monsterAttackDamage, def.MaxHP, def.AttackDamage)
	}
}

func TestMercenaryHiringRejectsInsufficientGold(t *testing.T) {
	sim, board := newMercenaryHiringSim(t, "v206_merc_hire_poor")
	cost := sim.rules.MainConfig.Gameplay.MercenaryHireCostGold
	sim.gold = cost - 1
	sim.progression.Gold = cost - 1
	sim.savePlayer(sim.defaultPlayer())

	hire := sim.Tick([]Input{mercenaryHireInput(board, "poor_hire")})

	assertReject(t, hire, "poor_hire", "not_enough_gold")
	if hiredMercenary(sim) != nil {
		t.Fatalf("unaffordable hire spawned mercenary")
	}
	opened := findEvent(hire.Events, "mercenary_board_opened")
	if opened == nil || opened.Affordable == nil || *opened.Affordable || opened.TotalGold == nil || *opened.TotalGold != cost-1 {
		t.Fatalf("unaffordable board event = %+v", opened)
	}
	if sim.gold != cost-1 || sim.progression.Gold != cost-1 {
		t.Fatalf("unaffordable hire mutated gold sim/progression=%d/%d", sim.gold, sim.progression.Gold)
	}
}

func TestMercenaryHiringRejectsInvalidTarget(t *testing.T) {
	sim, _ := newMercenaryHiringSim(t, "v206_merc_hire_invalid")
	cost := sim.rules.MainConfig.Gameplay.MercenaryHireCostGold
	sim.gold = cost
	sim.progression.Gold = cost
	sim.savePlayer(sim.defaultPlayer())

	hire := sim.Tick([]Input{{
		MessageID: "invalid_hire",
		Type:      "action_intent",
		Action:    &ActionIntent{TargetID: "missing_board"},
	}})

	assertReject(t, hire, "invalid_hire", "invalid_target")
	if hiredMercenary(sim) != nil {
		t.Fatalf("invalid hire spawned mercenary")
	}
	if sim.gold != cost || sim.progression.Gold != cost {
		t.Fatalf("invalid hire mutated gold sim/progression=%d/%d", sim.gold, sim.progression.Gold)
	}
}

func TestMercenaryHiringReplacesExistingHire(t *testing.T) {
	sim, board := newMercenaryHiringSim(t, "v206_merc_hire_replace")
	cost := sim.rules.MainConfig.Gameplay.MercenaryHireCostGold
	sim.gold = cost * 2
	sim.progression.Gold = cost * 2
	sim.savePlayer(sim.defaultPlayer())

	first := sim.Tick([]Input{mercenaryHireInput(board, "first_hire")})
	assertAck(t, first, "first_hire")
	firstMercenary := hiredMercenary(sim)
	if firstMercenary == nil {
		t.Fatalf("first hire missing mercenary")
	}
	second := sim.Tick([]Input{mercenaryHireInput(board, "second_hire")})
	assertAck(t, second, "second_hire")
	secondMercenary := hiredMercenary(sim)
	if secondMercenary == nil || secondMercenary.id == firstMercenary.id {
		t.Fatalf("second hire did not replace mercenary: first=%+v second=%+v", firstMercenary, secondMercenary)
	}
	if countHiredMercenaries(sim) != 1 {
		t.Fatalf("hired mercenary count=%d, want 1", countHiredMercenaries(sim))
	}
	if !hasRemovedEntity(second, firstMercenary.id) {
		t.Fatalf("second hire did not remove first mercenary: changes=%+v", second.Changes)
	}
	if sim.gold != 0 || sim.progression.Gold != 0 {
		t.Fatalf("gold after two hires sim/progression=%d/%d, want 0/0", sim.gold, sim.progression.Gold)
	}
}

func TestMercenaryLossRemovesHireAndEmitsEvent(t *testing.T) {
	sim, board := newMercenaryHiringSim(t, "v220_mercenary_loss")
	cost := sim.rules.MainConfig.Gameplay.MercenaryHireCostGold
	sim.gold = cost
	sim.progression.Gold = cost
	sim.savePlayer(sim.defaultPlayer())
	hire := sim.Tick([]Input{mercenaryHireInput(board, "hire_for_loss")})
	assertAck(t, hire, "hire_for_loss")
	mercenary := hiredMercenary(sim)
	if mercenary == nil {
		t.Fatalf("hire missing mercenary")
	}

	res := &TickResult{}
	attacker := mercenaryLossAttacker(sim, mercenary)
	damage := mercenary.maxHP + int(mercenary.monsterArmor) + 1
	sim.damageCompanionByMonster(attacker, mercenary, DamageRange{Min: damage, Max: damage}, "merc_loss", res)

	if hiredMercenary(sim) != nil || countHiredMercenaries(sim) != 0 {
		t.Fatalf("lost mercenary still active: count=%d mercenary=%+v", countHiredMercenaries(sim), hiredMercenary(sim))
	}
	if !hasRemovedEntity(*res, mercenary.id) {
		t.Fatalf("loss did not remove mercenary entity: changes=%+v", res.Changes)
	}
	killed := findEvent(res.Events, "companion_killed")
	if killed == nil || killed.SourceEntityID != idStr(attacker.id) || killed.TargetEntityID != idStr(mercenary.id) {
		t.Fatalf("companion_killed event = %+v", killed)
	}
	lost := findEvent(res.Events, "mercenary_lost")
	if lost == nil || lost.EntityID != idStr(mercenary.id) || lost.SourceEntityID != idStr(attacker.id) ||
		lost.TargetEntityID != idStr(mercenary.id) || lost.Service != mercenaryService ||
		lost.OfferID != mercenaryGuardOfferID || lost.MonsterDefID != mercenaryGuardMonsterDefID {
		t.Fatalf("mercenary_lost event = %+v", lost)
	}
}

func TestMercenaryCanRehireAfterLoss(t *testing.T) {
	sim, board := newMercenaryHiringSim(t, "v220_mercenary_rehire_after_loss")
	cost := sim.rules.MainConfig.Gameplay.MercenaryHireCostGold
	sim.gold = cost * 2
	sim.progression.Gold = cost * 2
	sim.savePlayer(sim.defaultPlayer())

	first := sim.Tick([]Input{mercenaryHireInput(board, "first_hire_for_loss")})
	assertAck(t, first, "first_hire_for_loss")
	firstMercenary := hiredMercenary(sim)
	if firstMercenary == nil {
		t.Fatalf("first hire missing mercenary")
	}
	loss := &TickResult{}
	attacker := mercenaryLossAttacker(sim, firstMercenary)
	damage := firstMercenary.maxHP + int(firstMercenary.monsterArmor) + 1
	sim.damageCompanionByMonster(attacker, firstMercenary, DamageRange{Min: damage, Max: damage}, "merc_loss_rehire", loss)
	if findEvent(loss.Events, "mercenary_lost") == nil {
		t.Fatalf("loss events missing mercenary_lost: %+v", loss.Events)
	}

	second := sim.Tick([]Input{mercenaryHireInput(board, "rehire_after_loss")})
	assertAck(t, second, "rehire_after_loss")
	secondMercenary := hiredMercenary(sim)
	if secondMercenary == nil || secondMercenary.id == firstMercenary.id {
		t.Fatalf("rehire did not spawn replacement: first=%+v second=%+v", firstMercenary, secondMercenary)
	}
	if countHiredMercenaries(sim) != 1 {
		t.Fatalf("hired mercenary count=%d, want 1", countHiredMercenaries(sim))
	}
	if sim.gold != 0 || sim.progression.Gold != 0 {
		t.Fatalf("gold after hire/loss/rehire sim/progression=%d/%d, want 0/0", sim.gold, sim.progression.Gold)
	}
}

func newMercenaryHiringSim(t *testing.T, seed string) (*Sim, *entity) {
	t.Helper()
	sim, err := NewSimWithWorld("sess_mercenary_hiring", seed, loadRules(t), "mercenary_hiring_lab")
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

func countHiredMercenaries(sim *Sim) int {
	count := 0
	for _, id := range sortedEntityIDs(sim.entities) {
		e := sim.entities[id]
		if e.kind == companionEntity && e.ownerID == sim.playerID && e.sourceSkillID == mercenaryHireSourceID {
			count++
		}
	}
	return count
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
