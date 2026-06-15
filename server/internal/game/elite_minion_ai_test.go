package game

import "testing"

func eliteMinionLab(t *testing.T) (*Sim, *LevelState, *entity, *entity, *entity) {
	t.Helper()
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_elite_minion_ai", "elite_minion_ai", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("dungeon world: %v", err)
	}
	level, err := sim.ensureDungeonLevel(-1)
	if err != nil {
		t.Fatal(err)
	}
	for id, candidate := range level.entities {
		if candidate.kind == monsterEntity {
			delete(level.entities, id)
		}
	}
	placeDefaultPlayerOnLevel(t, sim, level, Vec2{X: 8, Y: 5})
	sim.syncCompatibilityFields()

	leader := addTestMonster(sim, "dungeon_mob", Vec2{X: 20, Y: 5}, 30)
	leader.monsterPackID = "pack_ai"
	leader.monsterPackLeader = true
	minion := addTestMonster(sim, "dungeon_mob", Vec2{X: 24, Y: 5}, 30)
	minion.monsterPackID = leader.monsterPackID
	player := level.entities[sim.playerID]
	return sim, level, player, leader, minion
}

func TestEliteMinionFollowsLeaderWithoutPassiveAggro(t *testing.T) {
	sim, _, player, leader, minion := eliteMinionLab(t)
	def := sim.rules.Monsters[minion.monsterDefID]
	player.pos = Vec2{X: minion.pos.X + def.AggroRadius*0.5, Y: minion.pos.Y}
	beforeLeaderDistance := distance(minion.pos, leader.pos)

	res := TickResult{Tick: sim.tick, Level: sim.currentLevel}
	sim.advanceMonsterMovement(&res)

	if minion.aiMode != monsterAIModeIdle || minion.aiTargetPlayerID != 0 {
		t.Fatalf("minion target/mode = %d/%s, want idle without passive target", minion.aiTargetPlayerID, minion.aiMode)
	}
	if eventForEntity(res, "monster_aggro", minion.id) {
		t.Fatalf("idle elite minion should not passive aggro: %+v", res.Events)
	}
	if distance(minion.pos, leader.pos) >= beforeLeaderDistance-0.01 {
		t.Fatalf("minion did not follow leader: before %.3f after %.3f", beforeLeaderDistance, distance(minion.pos, leader.pos))
	}
}

func TestEliteMinionAssistsLeaderTarget(t *testing.T) {
	sim, _, player, leader, minion := eliteMinionLab(t)
	player.pos = Vec2{X: 16, Y: 5}
	leader.aiMode = monsterAIModeChase
	leader.aiTargetPlayerID = player.id
	beforePlayerDistance := distance(minion.pos, player.pos)

	res := TickResult{Tick: sim.tick, Level: sim.currentLevel}
	sim.advanceMonsterMovement(&res)

	if minion.aiMode != monsterAIModeChase || minion.aiTargetPlayerID != player.id {
		t.Fatalf("minion target/mode = %d/%s, want leader target %d/chase", minion.aiTargetPlayerID, minion.aiMode, player.id)
	}
	if distance(minion.pos, player.pos) >= beforePlayerDistance-0.01 {
		t.Fatalf("minion did not assist toward leader target: before %.3f after %.3f", beforePlayerDistance, distance(minion.pos, player.pos))
	}
}

func TestEliteMinionDoesNotAttackWithoutLeaderEngagement(t *testing.T) {
	sim, _, player, _, minion := eliteMinionLab(t)
	player.pos = minion.pos
	startHP := player.hp

	res := TickResult{Tick: sim.tick, Level: sim.currentLevel}
	sim.advanceMonsterAttack(&res)

	if player.hp != startHP {
		t.Fatalf("player hp = %d, want %d without idle minion attack", player.hp, startHP)
	}
	if hasEvent(res, "player_damaged") || hasEvent(res, "player_killed") {
		t.Fatalf("idle minion unexpectedly attacked: %+v", res.Events)
	}
}
