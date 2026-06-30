package game

import "testing"

func TestExperienceGainAndLevelUpFromMonsterKill(t *testing.T) {
	rules := cloneRules(loadRules(t))
	def := rules.Monsters["dungeon_mob"]
	def.XPReward = 20
	rules.Monsters["dungeon_mob"] = def
	sim := MustNewSim("sess_xp_kill", "01", rules)
	player := sim.entities[sim.playerID]
	player.hp = max(1, player.maxHP-4)
	player.mana = max(0, player.mana-3)
	monster := &entity{
		id:           sim.alloc(),
		kind:         monsterEntity,
		pos:          Vec2{X: player.pos.X + 0.5, Y: player.pos.Y},
		hp:           1,
		maxHP:        1,
		monsterDefID: "dungeon_mob",
		lootTable:    "no_drop",
	}
	sim.entities[monster.id] = monster

	res := sim.Tick([]Input{{MessageID: "kill_xp", CorrelationID: "corr_xp", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertAck(t, res, "kill_xp")
	if !hasEvent(res, "monster_killed") || !hasEvent(res, "experience_gained") || !hasEvent(res, "character_leveled") {
		t.Fatalf("missing kill/xp/level events: %+v", res.Events)
	}
	view := sim.CharacterProgressionView()
	if view.Experience != 20 || view.Level != 2 || view.UnspentStatPoints != 3 {
		t.Fatalf("progression after kill = %+v, want exp 20 level 2 unspent 3", view)
	}
	if !hasProgressionChange(res) {
		t.Fatalf("missing progression update change: %+v", res.Changes)
	}
	if player.hp != player.maxHP || player.mana != player.maxMana {
		t.Fatalf("level up restore hp/mana = %d/%d, want full %d/%d", player.hp, player.mana, player.maxHP, player.maxMana)
	}
	if !eventWithReason(res, "player_healed", "level_up") || !eventWithReason(res, "player_mana_restored", "level_up") {
		t.Fatalf("missing level-up restore events: %+v", res.Events)
	}

	reject := sim.Tick([]Input{{MessageID: "kill_again", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertReject(t, reject, "kill_again", "invalid_target")
	if sim.CharacterProgressionView().Experience != 20 {
		t.Fatalf("dead monster granted XP twice: %+v", sim.CharacterProgressionView())
	}
}

func eventWithReason(r TickResult, eventType string, reason string) bool {
	for _, ev := range r.Events {
		if ev.EventType == eventType && ev.Reason == reason {
			return true
		}
	}
	return false
}
