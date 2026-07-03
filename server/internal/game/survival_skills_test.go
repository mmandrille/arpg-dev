package game

import "testing"

func TestSurvivalSecondWindProcsOnLethalAndFloorsHP(t *testing.T) {
	rules := cloneRules(loadRules(t))
	forceMonsterHitChance(rules, monsterDefID, 1.0)
	sim := MustNewSim("sess_survival_second_wind", "01", rules)
	sim.progression.CharacterClass = "barbarian"
	sim.progression.Level = 10
	sim.progression.SkillRanks["second_wind"] = 1
	player := sim.entities[sim.playerID]
	player.hp = 5
	monster := addTestMonster(sim, monsterDefID, Vec2{X: player.pos.X + 1, Y: player.pos.Y}, 20)

	res := &TickResult{}
	outcome := sim.damagePlayerByMonster(monster, player, DamageRange{Min: 10, Max: 10}, "lethal", res)
	if player.hp != 1 {
		t.Fatalf("player hp = %d, want 1 after survival proc", player.hp)
	}
	if outcome.Damage != 4 {
		t.Fatalf("lethal damage applied = %d, want 4 (5 -> 1)", outcome.Damage)
	}
	if firstEventOfType(res.Events, "skill_effect_started") == nil {
		t.Fatalf("missing skill_effect_started: %+v", res.Events)
	}
	if firstEventOfType(res.Events, "skill_cooldown_started") == nil {
		t.Fatalf("missing skill_cooldown_started: %+v", res.Events)
	}
	if _, onCooldown := sim.skillCooldownRemaining("second_wind"); !onCooldown {
		t.Fatal("second_wind should be on cooldown after proc")
	}
}

func TestSurvivalDoesNotProcOnPlayerAttacker(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_survival_pvp_guard", "01", rules)
	sim.progression.CharacterClass = "barbarian"
	sim.progression.Level = 10
	sim.progression.SkillRanks["second_wind"] = 1
	player := sim.entities[sim.playerID]
	player.hp = 5
	attacker := &entity{id: sim.alloc(), kind: playerEntity, hp: 100, maxHP: 100, pos: Vec2{X: player.pos.X + 1, Y: player.pos.Y}}
	outcome := combatResolution{Damage: 10, Hit: true, Outcome: "hit"}
	sim.trySurvivalAutocast(player, attacker, &outcome, "pvp", &TickResult{})
	if outcome.Damage != 10 {
		t.Fatalf("damage = %d, want unchanged 10 when attacker is player", outcome.Damage)
	}
}

func TestSurvivalArcaneBarrierDrainsManaBeforeHP(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_survival_arcane_barrier", "01", rules)
	sim.progression.CharacterClass = "sorcerer"
	sim.progression.Level = 10
	sim.progression.SkillRanks["arcane_barrier"] = 1
	player := sim.entities[sim.playerID]
	player.hp = 50
	player.mana = 100
	player.maxMana = 100
	sim.activateSurvivalSkill(player, nil, "arcane_barrier", rules.Skills["arcane_barrier"], 1, "barrier", &TickResult{})
	monster := addTestMonster(sim, monsterDefID, Vec2{X: player.pos.X + 1, Y: player.pos.Y}, 20)
	res := &TickResult{}
	outcome := sim.damagePlayerByMonster(monster, player, DamageRange{Min: 10, Max: 10}, "mana_shield", res)
	if outcome.Damage != 0 {
		t.Fatalf("damage = %d, want 0 while mana absorbs", outcome.Damage)
	}
	if player.mana >= 100 || player.mana <= 60 {
		t.Fatalf("mana = %d, want reduced from 100 by mana shield", player.mana)
	}
}

func TestSurvivalDivineProtectionGrantsImmunityAndOutgoingDamage(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_survival_divine", "01", rules)
	sim.progression.CharacterClass = "paladin"
	sim.progression.Level = 10
	sim.progression.SkillRanks["divine_protection"] = 1
	player := sim.entities[sim.playerID]
	sim.activateSurvivalSkill(player, nil, "divine_protection", rules.Skills["divine_protection"], 1, "divine", &TickResult{})
	if _, immune := sim.playerDamageImmunityOutcome(player); !immune {
		t.Fatal("expected immunity while divine protection active")
	}
	outcome := combatResolution{Damage: 4, RawDamage: 4, Hit: true}
	sim.applyPlayerOutgoingDamageMultiplier(&outcome)
	if outcome.Damage != 20 {
		t.Fatalf("outgoing damage = %d, want 20 (4 * 5)", outcome.Damage)
	}
}

func TestSurvivalSpectralPathRedirectsDamage(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_survival_spectral", "01", rules)
	sim.progression.CharacterClass = "ranger"
	sim.progression.Level = 10
	sim.progression.SkillRanks["spectral_path"] = 1
	player := sim.entities[sim.playerID]
	player.hp = 40
	monster := addTestMonster(sim, monsterDefID, Vec2{X: player.pos.X + 1, Y: player.pos.Y}, 30)
	sim.activateSurvivalSkill(player, monster, "spectral_path", rules.Skills["spectral_path"], 1, "spectral", &TickResult{})
	beforeMonsterHP := monster.hp
	res := &TickResult{}
	outcome := combatResolution{Damage: 12, Hit: true}
	if !sim.applyPlayerDamageRedirect(player, monster, &outcome, "redirect", res) {
		t.Fatal("expected redirect to apply")
	}
	if outcome.Damage != 0 {
		t.Fatalf("player damage = %d, want 0 after redirect", outcome.Damage)
	}
	if monster.hp >= beforeMonsterHP {
		t.Fatalf("monster hp = %d, want damage redirected from %d", monster.hp, beforeMonsterHP)
	}
}

func TestSurvivalAutocastNotCastableViaIntent(t *testing.T) {
	rules := loadRules(t)
	state := rules.DefaultCharacterProgressionState()
	state.CharacterClass = "barbarian"
	state.Level = 10
	state.BaseStats = rules.CharacterProgression.Classes["barbarian"].BaseStats
	sim, err := NewSimWithWorldProgression("sess_survival_not_castable", "01", rules, DefaultWorldID, state)
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	sim.progression.SkillRanks["second_wind"] = 1
	res := sim.Tick([]Input{{
		MessageID:     "cast_second_wind",
		CorrelationID: "cast_second_wind",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "second_wind"},
	}})
	assertReject(t, res, "cast_second_wind", "passive_skill_not_castable")
}
