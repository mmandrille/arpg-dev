package game

import "testing"

func TestRangedMonsterBlockedShotRepositionsCloseBeforeFiring(t *testing.T) {
	rules := cloneRules(loadRules(t))
	forceMonsterHitChance(rules, "dungeon_archer", 1.0)
	sim, err := NewSimWithWorld("sess_ranged_monster_reposition", "v273_ranged_reposition", rules, "ranged_monster_ai_lab")
	if err != nil {
		t.Fatal(err)
	}
	player := sim.entities[sim.playerID]
	archer := findMonsterByDef(sim, "dungeon_archer")
	if player == nil || archer == nil {
		t.Fatalf("setup player=%+v archer=%+v", player, archer)
	}
	if distance(player.pos, archer.pos) <= 3.75 {
		t.Fatalf("setup archer already close: player=%+v archer=%+v", player.pos, archer.pos)
	}
	if s := sim.hasClearMonsterRangedShot(archer.pos, player); s {
		t.Fatalf("setup has clear archer shot from %+v to %+v", archer.pos, player.pos)
	}

	var projectile *EntityView
	closeBeforeShot := false
	for i := 0; i < 160; i++ {
		res := sim.Tick(nil)
		if distance(player.pos, archer.pos) <= 3.75 {
			closeBeforeShot = true
		}
		if spawned := firstChangeEntityByType(res, projectileEntity); spawned != nil {
			projectile = spawned
			break
		}
	}
	if !closeBeforeShot {
		t.Fatalf("archer did not navigate close before shot: player=%+v archer=%+v", player.pos, archer.pos)
	}
	if projectile == nil {
		t.Fatalf("archer never fired after moving close: player=%+v archer=%+v", player.pos, archer.pos)
	}
	if projectile.OwnerID != idStr(archer.id) || projectile.TargetID != idStr(player.id) {
		t.Fatalf("projectile = %+v, want archer owner/player target", projectile)
	}
	if !sim.hasClearMonsterRangedShot(archer.pos, player) {
		t.Fatalf("archer fired without clear shot from %+v to %+v", archer.pos, player.pos)
	}
	if damage, ok := waitForPlayerDamage(sim, 20); !ok {
		t.Fatalf("close ranged shot did not damage player: player=%+v archer=%+v damage=%+v", player, archer, damage)
	}
}

func TestRangedMonsterRetreatsWhenPlayerIsTooClose(t *testing.T) {
	rules := cloneRules(loadRules(t))
	sim, err := NewSimWithWorld("sess_ranged_monster_retreat", "v286_ranged_retreat", rules, "ranged_monster_retreat_lab")
	if err != nil {
		t.Fatal(err)
	}
	player := sim.entities[sim.playerID]
	archer := findMonsterByDef(sim, "dungeon_archer")
	if player == nil || archer == nil {
		t.Fatalf("setup player=%+v archer=%+v", player, archer)
	}
	preferred := rules.Monsters["dungeon_archer"].PreferredMinRange
	initialDistance := distance(player.pos, archer.pos)
	if initialDistance >= preferred {
		t.Fatalf("setup archer distance %.3f already >= preferred %.3f", initialDistance, preferred)
	}

	retreated := false
	for i := 0; i < 40; i++ {
		sim.Tick(nil)
		if distance(player.pos, archer.pos) >= preferred-0.15 {
			retreated = true
			break
		}
	}
	if !retreated {
		t.Fatalf("archer did not retreat toward preferred range: initial=%.3f final=%.3f preferred=%.3f archer=%+v player=%+v", initialDistance, distance(player.pos, archer.pos), preferred, archer.pos, player.pos)
	}
	if distance(player.pos, archer.pos) <= initialDistance+0.75 {
		t.Fatalf("archer did not materially increase range: initial=%.3f final=%.3f", initialDistance, distance(player.pos, archer.pos))
	}
	if distance(player.pos, archer.pos) > sim.monsterAttackReach(rules.Monsters["dungeon_archer"])+playerRadius+meleeRangeEpsilon {
		t.Fatalf("archer retreated outside attack reach: player=%+v archer=%+v", player.pos, archer.pos)
	}
	if !sim.hasClearMonsterRangedShot(archer.pos, player) {
		t.Fatalf("archer retreated to blocked shot: player=%+v archer=%+v", player.pos, archer.pos)
	}
}
