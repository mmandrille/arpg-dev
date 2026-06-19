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
