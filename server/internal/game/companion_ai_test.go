package game

import "testing"

func TestCompanionAppearsInSnapshotWithOwner(t *testing.T) {
	sim, err := NewSimWithWorld("sess_companion_identity", "v182_companion_identity", loadRules(t), "companion_ai_lab")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	companion := firstEntityByKind(sim, companionEntity)
	if companion == nil {
		t.Fatalf("missing companion in entities: %+v", sim.entities)
	}
	view := sim.entityView(companion)
	if view.Type != companionEntity || view.OwnerID != idStr(sim.playerID) || view.MonsterDefID != "combat_lab_crit_attacker" {
		t.Fatalf("companion view = %+v", view)
	}
	if view.HP == nil || view.MaxHP == nil || *view.HP <= 0 || *view.MaxHP <= 0 {
		t.Fatalf("companion hp view = %+v", view)
	}
}

func TestMercenaryFoundationCompanionAppearsWithOwnedStats(t *testing.T) {
	sim, err := NewSimWithWorld("sess_mercenary_foundation", "v198_mercenary_foundation", loadRules(t), "mercenary_foundation_lab")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	mercenary := firstEntityByKind(sim, companionEntity)
	if mercenary == nil {
		t.Fatalf("missing mercenary companion in entities: %+v", sim.entities)
	}
	view := sim.entityView(mercenary)
	if view.Type != companionEntity || view.OwnerID != idStr(sim.playerID) || view.MonsterDefID != "mercenary_guard" {
		t.Fatalf("mercenary view = %+v", view)
	}
	def := sim.rules.Monsters["mercenary_guard"]
	if mercenary.hp != def.MaxHP || mercenary.monsterAttackDamage == nil || *mercenary.monsterAttackDamage != *def.AttackDamage {
		t.Fatalf("mercenary stats hp=%d damage=%+v, want hp %d damage %+v", mercenary.hp, mercenary.monsterAttackDamage, def.MaxHP, def.AttackDamage)
	}
	stats := view.CombatStats
	if stats == nil || stats.DamageMin != def.AttackDamage.Min || stats.DamageMax != def.AttackDamage.Max ||
		stats.AttackCooldownTicks != def.AttackCooldown || stats.Armor != float64(def.Armor) ||
		stats.BlockPercent != float64(def.BlockPercent) || stats.HitChance != def.effectiveHitChance(sim.rules.Combat) ||
		stats.CritChance != def.effectiveCritChance(sim.rules.Combat) {
		t.Fatalf("mercenary combat stats view = %+v, want rules-backed stats for %+v", stats, def)
	}
}

func TestCompanionAttacksWhenOwnerBelowMeleeDeadzone(t *testing.T) {
	sim, err := NewSimWithWorld("sess_companion_deadzone", "v182_companion_ai_foundation", loadRules(t), "companion_ai_lab")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	player := sim.entities[sim.playerID]
	companion := firstEntityByKind(sim, companionEntity)
	target := findMonsterByDef(sim, "combat_lab_soft_target")
	player.pos = Vec2{X: 5.5, Y: 5}
	startHP := target.hp
	var sawDamage bool
	for i := 0; i < 120; i++ {
		for _, res := range sim.TickResults(nil) {
			for _, ev := range res.Events {
				if ev.EventType == "monster_damaged" && ev.SourceEntityID == idStr(companion.id) {
					sawDamage = true
				}
			}
		}
	}
	if !sawDamage || target.hp >= startHP {
		t.Fatalf("companion did not damage when owner stayed short of melee deadzone: saw=%v hp=%d dist=%.3f",
			sawDamage, target.hp, distance(companion.pos, target.pos))
	}
}

func TestCompanionAttackAppearsInActorZeroTickResult(t *testing.T) {
	sim, err := NewSimWithWorld("sess_companion_actor_zero", "v182_companion_ai_foundation", loadRules(t), "companion_ai_lab")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	player := sim.entities[sim.playerID]
	companion := firstEntityByKind(sim, companionEntity)
	target := findMonsterByDef(sim, "combat_lab_soft_target")
	if player == nil || companion == nil || target == nil {
		t.Fatalf("setup player=%+v companion=%+v target=%+v", player, companion, target)
	}
	player.pos = Vec2{X: 6, Y: 5}
	var sawDamage bool
	for i := 0; i < 140; i++ {
		for _, res := range sim.TickResults(nil) {
			if res.ActorPlayerID != 0 {
				continue
			}
			for _, ev := range res.Events {
				if ev.EventType == "monster_damaged" && ev.SourceEntityID == idStr(companion.id) && ev.TargetEntityID == idStr(target.id) {
					sawDamage = true
				}
			}
		}
		if sawDamage {
			break
		}
	}
	if !sawDamage {
		t.Fatalf("companion damage missing from actor-zero tick results")
	}
}

func TestCompanionDamagesMonsterViaTickResultsWithMoveInputs(t *testing.T) {
	sim, err := NewSimWithWorld("sess_companion_ai_protocol", "v182_companion_ai_foundation", loadRules(t), "companion_ai_lab")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	companion := firstEntityByKind(sim, companionEntity)
	target := findMonsterByDef(sim, "combat_lab_soft_target")
	if companion == nil || target == nil {
		t.Fatalf("setup companion=%+v target=%+v", companion, target)
	}
	startCompanionPos := companion.pos
	startTargetHP := target.hp
	moveRight := Input{
		MessageID: "move_right",
		Type:      "move_intent",
		Move:      &MoveIntent{Direction: Vec2{X: 1, Y: 0}, DurationTicks: 1},
	}
	for i := 0; i < 8; i++ {
		moveRight.MessageID = idStr(uint64(i + 1))
		sim.TickResults([]Input{moveRight})
	}
	player := sim.entities[sim.playerID]
	player.pos = Vec2{X: 5.5, Y: 5}
	var sawDamage bool
	for i := 0; i < 95; i++ {
		pulse := Input{
			MessageID: idStr(uint64(100 + i)),
			Type:      "move_intent",
			Move:      &MoveIntent{Direction: Vec2{X: 1, Y: 0}, DurationTicks: 1},
		}
		if i%2 == 1 {
			pulse.Move.Direction = Vec2{X: -1, Y: 0}
		}
		for _, res := range sim.TickResults([]Input{pulse}) {
			for _, ev := range res.Events {
				if ev.EventType == "monster_damaged" && ev.SourceEntityID == idStr(companion.id) && ev.TargetEntityID == idStr(target.id) {
					sawDamage = true
				}
			}
		}
	}
	if distance(companion.pos, startCompanionPos) < 0.5 {
		t.Fatalf("companion did not follow owner: start=%+v now=%+v", startCompanionPos, companion.pos)
	}
	if !sawDamage || target.hp >= startTargetHP {
		t.Fatalf("companion did not damage target via TickResults: sawDamage=%v hp=%d start=%d companion=%+v target=%+v dist=%.3f target_id=%d",
			sawDamage, target.hp, startTargetHP, companion.pos, target.pos, distance(companion.pos, target.pos), companion.targetID)
	}
}

func TestCompanionFollowsOwnerAndDamagesMonster(t *testing.T) {
	sim, err := NewSimWithWorld("sess_companion_ai", "v182_companion_ai", loadRules(t), "companion_ai_lab")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	player := sim.entities[sim.playerID]
	companion := firstEntityByKind(sim, companionEntity)
	target := findMonsterByDef(sim, "combat_lab_soft_target")
	if player == nil || companion == nil || target == nil {
		t.Fatalf("setup player=%+v companion=%+v target=%+v", player, companion, target)
	}
	startCompanionPos := companion.pos
	startTargetHP := target.hp
	player.pos = Vec2{X: 6, Y: 5}

	var sawDamage bool
	for i := 0; i < 140; i++ {
		res := sim.Tick(nil)
		for _, ev := range res.Events {
			if ev.EventType == "monster_damaged" && ev.SourceEntityID == idStr(companion.id) && ev.TargetEntityID == idStr(target.id) {
				sawDamage = true
			}
		}
		if sawDamage {
			break
		}
	}
	if distance(companion.pos, startCompanionPos) < 0.5 {
		t.Fatalf("companion did not follow owner: start=%+v now=%+v", startCompanionPos, companion.pos)
	}
	if !sawDamage || target.hp >= startTargetHP {
		t.Fatalf("companion did not damage target: sawDamage=%v hp=%d start=%d companion=%+v target=%+v dist=%.3f target_id=%d", sawDamage, target.hp, startTargetHP, companion.pos, target.pos, distance(companion.pos, target.pos), companion.targetID)
	}
}

func TestMainConfigCompanionFollowDistanceDrivesSpawnAndTravel(t *testing.T) {
	rules := loadRulesWithMainGameplay(t, map[string]any{
		"companion_follow_distance": 2.25,
	})
	sim, err := NewSimWithWorld("sess_companion_follow_config", "v232_companion_follow_config", rules, "companion_ai_lab")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	player := sim.entities[sim.playerID]
	companion := firstEntityByKind(sim, companionEntity)
	if player == nil || companion == nil {
		t.Fatalf("setup player=%+v companion=%+v", player, companion)
	}
	spawn := sim.companionSpawnPosition(player)
	if spawn.X != player.pos.X+2.25 || spawn.Y != player.pos.Y {
		t.Fatalf("spawn position = %+v, player=%+v", spawn, player.pos)
	}
	travel := sim.companionTravelPosition(player.pos, 0)
	if travel.X != player.pos.X+2.25 || travel.Y != player.pos.Y {
		t.Fatalf("travel position = %+v, player=%+v", travel, player.pos)
	}
}
