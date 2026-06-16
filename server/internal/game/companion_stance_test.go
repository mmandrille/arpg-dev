package game

import "testing"

func TestCompanionStanceCommandUpdatesOwnedCompanions(t *testing.T) {
	sim, board := newMercenaryHiringSim(t, "v208_companion_stance_command")
	cost := sim.rules.MainConfig.Gameplay.MercenaryHireCostGold
	sim.gold = cost
	sim.progression.Gold = cost
	sim.savePlayer(sim.defaultPlayer())
	hire := sim.Tick([]Input{mercenaryHireInput(board, "hire_for_stance")})
	assertAck(t, hire, "hire_for_stance")
	mercenary := hiredMercenary(sim)
	if mercenary == nil {
		t.Fatal("missing hired mercenary")
	}
	if got := sim.entityView(mercenary).CompanionStance; got != CompanionStanceAssist {
		t.Fatalf("default companion stance = %q, want %q", got, CompanionStanceAssist)
	}

	res := sim.Tick([]Input{{
		MessageID:        "stance_passive",
		CorrelationID:    "corr_stance_passive",
		Type:             "companion_command_intent",
		CompanionCommand: &CompanionCommandIntent{Stance: CompanionStancePassive},
	}})

	assertAck(t, res, "stance_passive")
	if mercenary.companionStanceOrDefault() != CompanionStancePassive {
		t.Fatalf("mercenary stance = %q", mercenary.companionStanceOrDefault())
	}
	if !hasChange(res, OpEntityUpdate) {
		t.Fatalf("stance command missing entity update: %+v", res.Changes)
	}
	event := findEvent(res.Events, "companion_stance_changed")
	if event == nil || event.EntityID != idStr(sim.playerID) || event.Stance != CompanionStancePassive || event.Amount == nil || *event.Amount != 1 {
		t.Fatalf("companion_stance_changed = %+v", event)
	}
}

func TestCompanionStancePassivePreventsTargetingAndAttack(t *testing.T) {
	sim, companion, target := newCompanionStanceAISim(t)
	companion.companionStance = CompanionStancePassive
	companion.targetID = target.id
	startHP := target.hp

	for i := 0; i < 80; i++ {
		sim.Tick(nil)
	}

	if companion.targetID != 0 {
		t.Fatalf("passive companion target_id=%d, want 0", companion.targetID)
	}
	if target.hp != startHP {
		t.Fatalf("passive companion damaged target hp=%d, want %d", target.hp, startHP)
	}
}

func TestCompanionStanceDefendUsesOwnerProximity(t *testing.T) {
	sim, companion, target := newCompanionStanceAISim(t)
	player := sim.entities[sim.playerID]
	player.pos = Vec2{X: 0, Y: 0}
	companion.pos = Vec2{X: 10, Y: 0}
	target.pos = Vec2{X: 10.5, Y: 0}
	companion.companionStance = CompanionStanceDefend

	if got := sim.companionTarget(companion); got != nil {
		t.Fatalf("defend targeted monster near companion but far from owner: %+v", got)
	}

	target.pos = Vec2{X: 1, Y: 0}
	if got := sim.companionTarget(companion); got == nil || got.id != target.id {
		t.Fatalf("defend target = %+v, want owner-near monster %d", got, target.id)
	}
}

func TestCompanionStanceCommandRejectsInvalidOrMissingCompanion(t *testing.T) {
	sim := MustNewSim("sess_v208_no_companion", "v208_no_companion", loadRules(t))
	bad := sim.Tick([]Input{{
		MessageID:        "bad_stance",
		Type:             "companion_command_intent",
		CompanionCommand: &CompanionCommandIntent{Stance: "hold"},
	}})
	assertReject(t, bad, "bad_stance", "invalid_stance")

	missing := sim.Tick([]Input{{
		MessageID:        "no_companion",
		Type:             "companion_command_intent",
		CompanionCommand: &CompanionCommandIntent{Stance: CompanionStanceAssist},
	}})
	assertReject(t, missing, "no_companion", "no_companion")
}

func newCompanionStanceAISim(t *testing.T) (*Sim, *entity, *entity) {
	t.Helper()
	sim := MustNewSim("sess_v208_companion_stance_ai", "v208_companion_stance_ai", loadRules(t))
	player := sim.entities[sim.playerID]
	def := sim.rules.Monsters["dungeon_mob"]
	companion := &entity{
		id:                    sim.alloc(),
		kind:                  companionEntity,
		pos:                   Vec2{X: player.pos.X + 1, Y: player.pos.Y},
		hp:                    20,
		maxHP:                 20,
		ownerID:               player.id,
		monsterDefID:          mercenaryGuardMonsterDefID,
		monsterAttackDamage:   &DamageRange{Min: 1, Max: 1},
		monsterAttackCooldown: 1,
		aiMode:                monsterAIModeIdle,
		speed:                 1,
	}
	target := &entity{
		id:           sim.alloc(),
		kind:         monsterEntity,
		pos:          Vec2{X: companion.pos.X + 0.5, Y: companion.pos.Y},
		hp:           20,
		maxHP:        20,
		monsterDefID: "dungeon_mob",
		lootTable:    "no_drop",
	}
	if def.AttackDamage != nil {
		target.monsterAttackDamage = def.AttackDamage
	}
	sim.entities[companion.id] = companion
	sim.entities[target.id] = target
	return sim, companion, target
}
