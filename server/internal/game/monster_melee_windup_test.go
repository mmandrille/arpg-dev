package game

import "testing"

func TestDungeonMobMeleeWindupDelaysDamage(t *testing.T) {
	rules := loadRules(t)
	mobDef := rules.Monsters["dungeon_mob"]
	mobDef.HitChance = floatPtr(1)
	mobDef.AttackDamage = &DamageRange{Min: 2, Max: 2}
	mobDef.AttackCooldown = 30
	mobDef.AttackWindupTicks = 6
	rules.Monsters["dungeon_mob"] = mobDef

	sim, err := NewSimWithWorld("sess_mob_melee_windup", "01", rules, "inventory_lab")
	if err != nil {
		t.Fatal(err)
	}
	for id, candidate := range sim.activeLevel().entities {
		if candidate.kind == monsterEntity {
			delete(sim.activeLevel().entities, id)
		}
	}
	player := sim.entities[sim.playerID]
	player.pos = Vec2{X: 5, Y: 5}
	player.hp = playerStartHP

	mob := addTestMonster(sim, "dungeon_mob", Vec2{X: 6, Y: 5}, rules.Monsters["dungeon_mob"].MaxHP)
	mob.aiMode = monsterAIModeChase
	start := TickResult{Tick: sim.tick, Level: sim.currentLevel}
	sim.advanceMonsterAttack(&start)
	windup := firstEventByType(start, "monster_attack_windup")
	if windup == nil {
		t.Fatalf("start events = %+v, want monster_attack_windup", start.Events)
	}
	if firstEventBySource(start, "player_damaged", mob.id) != nil {
		t.Fatalf("damage should not land on windup start: %+v", start.Events)
	}
	if windup.TotalTicks == nil || *windup.TotalTicks != 6 {
		t.Fatalf("windup total ticks = %+v, want 6", windup.TotalTicks)
	}

	for tick := 0; tick < 5; tick++ {
		sim.tick++
		mid := TickResult{Tick: sim.tick, Level: sim.currentLevel}
		sim.advanceMonsterMeleeWindups(&mid)
		if firstEventBySource(mid, "player_damaged", mob.id) != nil {
			t.Fatalf("damage landed early on tick %d: %+v", tick+1, mid.Events)
		}
	}

	sim.tick++
	finish := TickResult{Tick: sim.tick, Level: sim.currentLevel}
	sim.advanceMonsterMeleeWindups(&finish)
	damage := firstEventBySource(finish, "player_damaged", mob.id)
	if damage == nil {
		t.Fatalf("finish events = %+v, want player_damaged after windup", finish.Events)
	}
}

func TestWolfPounceWindupDelaysDamage(t *testing.T) {
	rules := loadRules(t)
	wolfDef := rules.Monsters["dungeon_wolf"]
	wolfDef.HitChance = floatPtr(1)
	wolfDef.AttackDamage = &DamageRange{Min: 2, Max: 2}
	wolfDef.AttackCooldown = 30
	wolfDef.AttackWindupTicks = 5
	rules.Monsters["dungeon_wolf"] = wolfDef

	sim, err := NewSimWithWorld("sess_wolf_pounce_windup", "01", rules, "inventory_lab")
	if err != nil {
		t.Fatal(err)
	}
	for id, candidate := range sim.activeLevel().entities {
		if candidate.kind == monsterEntity {
			delete(sim.activeLevel().entities, id)
		}
	}
	player := sim.entities[sim.playerID]
	player.pos = Vec2{X: 5, Y: 5}
	player.hp = playerStartHP
	wolfDistance := rules.Combat.UnarmedReach + playerRadius + 0.35

	wolf := addTestMonster(sim, "dungeon_wolf", Vec2{X: player.pos.X + wolfDistance, Y: player.pos.Y}, rules.Monsters["dungeon_wolf"].MaxHP)
	wolf.aiMode = monsterAIModeChase
	start := TickResult{Tick: sim.tick, Level: sim.currentLevel}
	sim.advanceMonsterAttack(&start)
	windup := firstEventByType(start, "monster_attack_windup")
	if windup == nil || windup.AttackStyle != monsterAttackStylePounce {
		t.Fatalf("start events = %+v, want pounce windup", start.Events)
	}
}

func firstEventByType(r TickResult, eventType string) *Event {
	for idx := range r.Events {
		event := &r.Events[idx]
		if event.EventType == eventType {
			return event
		}
	}
	return nil
}
