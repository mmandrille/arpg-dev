package game

import "testing"

func TestRangedMonsterProjectileTargetsEngagedCompanion(t *testing.T) {
	rules := cloneRules(loadRules(t))
	forceMonsterHitChance(rules, "dungeon_archer", 1.0)
	sim, err := NewSimWithWorld("sess_ranged_companion_projectile", "v248_ranged_companion", rules, "ranged_companion_target_lab")
	if err != nil {
		t.Fatal(err)
	}
	player := sim.entities[sim.playerID]
	archer := findMonsterByDef(sim, "dungeon_archer")
	companion := firstEntityByKind(sim, companionEntity)
	if player == nil || archer == nil || companion == nil {
		t.Fatalf("setup player=%+v archer=%+v companion=%+v", player, archer, companion)
	}

	var spawn TickResult
	var projectile *EntityView
	for i := 0; i < 20; i++ {
		spawn = sim.Tick(nil)
		projectile = firstChangeEntityByType(spawn, projectileEntity)
		if projectile != nil {
			break
		}
	}
	if projectile == nil {
		t.Fatalf("ranged monster did not spawn projectile: last=%+v archer=%+v companion=%+v", spawn, archer, companion)
	}
	if projectile.OwnerID != idStr(archer.id) || projectile.TargetID != idStr(companion.id) {
		t.Fatalf("projectile owner/target = %s/%s, want archer %s companion %s", projectile.OwnerID, projectile.TargetID, idStr(archer.id), idStr(companion.id))
	}

	beforeHP := companion.hp
	var damage TickResult
	for i := 0; i < 20; i++ {
		damage = sim.Tick(nil)
		if hasEvent(damage, "companion_damaged") || hasEvent(damage, "companion_killed") {
			break
		}
	}
	if !hasEvent(damage, "companion_damaged") && !hasEvent(damage, "companion_killed") {
		t.Fatalf("projectile did not damage companion: events=%+v companion=%+v", damage.Events, companion)
	}
	if companion.hp >= beforeHP {
		t.Fatalf("companion hp = %d, want below %d after archer projectile", companion.hp, beforeHP)
	}
	if player.hp != playerStartHP {
		t.Fatalf("player hp = %d, want unchanged %d", player.hp, playerStartHP)
	}
}
