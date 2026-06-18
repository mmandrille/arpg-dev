package game

import "testing"

func TestAggroOnHitDirectionalRangedMovesFromOutsidePassiveRadius(t *testing.T) {
	rules := cloneRules(loadRules(t))
	forceCharacterHitChance(rules, 1.0)
	sim := combatControlLabWithEquippedBow(t, rules, "cafebabecafebabe")
	player := sim.entities[sim.playerID]
	player.pos = Vec2{X: 3, Y: 5}
	monster := firstEntityByKind(sim, monsterEntity)
	monster.hp = 20
	monster.maxHP = 20
	if distance(player.pos, monster.pos) <= sim.rules.Monsters[monster.monsterDefID].AggroRadius {
		t.Fatalf("setup inside passive aggro radius: player=%+v monster=%+v", player.pos, monster.pos)
	}

	fire := sim.Tick([]Input{{MessageID: "fire", CorrelationID: "corr_aggro", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
	assertAck(t, fire, "fire")
	var impact TickResult
	for i := 0; i < 20; i++ {
		impact = sim.Tick(nil)
		if hasEvent(impact, "monster_aggro") {
			break
		}
	}
	if !hasEvent(impact, "monster_aggro") {
		t.Fatalf("impact events = %+v, want monster_aggro", impact.Events)
	}
	before := monster.pos
	sim.Tick(nil)
	if distance(monster.pos, player.pos) >= distance(before, player.pos)-0.01 {
		t.Fatalf("monster did not chase aggro target: before=%+v after=%+v player=%+v", before, monster.pos, player.pos)
	}
}

func TestAggroOnHitPrefersAttackingPlayerInCoop(t *testing.T) {
	rules := cloneRules(loadRules(t))
	forceCharacterHitChance(rules, 1.0)
	sim := combatControlLabWithEquippedBow(t, rules, "cafebabecafebabe")
	hostID := sim.playerID
	sim.SetPlayerMetadata(hostID, "acct_host", "char_host", "Host", "host")
	guestID, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatalf("add guest: %v", err)
	}
	monster := firstEntityByKind(sim, monsterEntity)
	monster.hp = 20
	monster.maxHP = 20
	sim.entities[hostID].pos = Vec2{X: 3, Y: 5}
	sim.entities[guestID].pos = Vec2{X: 12.4, Y: 5}
	sim.savePlayer(sim.players[hostID])
	sim.savePlayer(sim.players[guestID])
	sim.usePlayer(sim.players[hostID])

	fire := sim.TickResults([]Input{{MessageID: "fire", ActorPlayerID: hostID, CorrelationID: "corr_aggro", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
	if len(fire) == 0 {
		t.Fatal("directional fire produced no results")
	}
	assertAck(t, fire[0], "fire")
	var impact TickResult
	for i := 0; i < 20; i++ {
		for _, res := range sim.TickResults(nil) {
			if hasEvent(res, "monster_damaged") || hasEvent(res, "attack_missed") {
				impact = res
				break
			}
		}
		if hasEvent(impact, "monster_damaged") || hasEvent(impact, "attack_missed") {
			break
		}
	}
	if !hasEvent(impact, "monster_damaged") && !hasEvent(impact, "attack_missed") {
		t.Fatalf("projectile impact events = %+v, want monster_damaged or attack_missed", impact.Events)
	}
	if monster.aiTargetPlayerID != hostID {
		t.Fatalf("monster ai target = %d, want host %d", monster.aiTargetPlayerID, hostID)
	}
	targetPlayer := sim.nearestLivingPlayerForMonster(sim.activeLevel(), monster)
	if targetPlayer == nil || targetPlayer.PlayerID != hostID {
		t.Fatalf("target player = %+v, want host %d", targetPlayer, hostID)
	}
}

func TestAggroOnHitPropagatesToNearbyMonsterGroup(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_group_aggro", "group_aggro", rules, "dungeon_levels")
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
	placeDefaultPlayerOnLevel(t, sim, level, Vec2{X: 2, Y: 5})
	sim.syncCompatibilityFields()

	primary := addTestMonster(sim, "dungeon_mob", Vec2{X: 20, Y: 10}, 20)
	near := addTestMonster(sim, "dungeon_mob", Vec2{X: 25, Y: 10}, 20)
	chained := addTestMonster(sim, "dungeon_mob", Vec2{X: 30, Y: 10}, 20)
	far := addTestMonster(sim, "dungeon_mob", Vec2{X: 45, Y: 10}, 20)
	res := TickResult{Tick: sim.tick, Level: sim.currentLevel}

	sim.aggroMonsterOnHit(primary, sim.playerID, "corr_group", &res)

	for _, monster := range []*entity{primary, near, chained} {
		if monster.aiTargetPlayerID != sim.playerID || monster.aiMode != monsterAIModeChase {
			t.Fatalf("monster %d target/mode = %d/%s, want %d/%s", monster.id, monster.aiTargetPlayerID, monster.aiMode, sim.playerID, monsterAIModeChase)
		}
	}
	if far.aiTargetPlayerID != 0 || far.aiMode != monsterAIModeIdle {
		t.Fatalf("far monster target/mode = %d/%s, want idle outside group radius", far.aiTargetPlayerID, far.aiMode)
	}

	aggroEvents := map[string]bool{}
	for _, ev := range res.Events {
		if ev.EventType == "monster_aggro" {
			aggroEvents[ev.EntityID] = true
		}
	}
	for _, monster := range []*entity{primary, near, chained} {
		if !aggroEvents[idStr(monster.id)] {
			t.Fatalf("missing monster_aggro for %d in events %+v", monster.id, res.Events)
		}
	}
	if aggroEvents[idStr(far.id)] {
		t.Fatalf("unexpected far monster_aggro for %d in events %+v", far.id, res.Events)
	}
}

func TestAggroOnHitTickCacheStillAllowsSeparateGroups(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_tick_cache_group_aggro", "tick_cache_group_aggro", rules, "dungeon_levels")
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
	placeDefaultPlayerOnLevel(t, sim, level, Vec2{X: 2, Y: 5})
	sim.syncCompatibilityFields()

	first := addTestMonster(sim, "dungeon_mob", Vec2{X: 20, Y: 10}, 20)
	firstNear := addTestMonster(sim, "dungeon_mob", Vec2{X: 25, Y: 10}, 20)
	second := addTestMonster(sim, "dungeon_mob", Vec2{X: 60, Y: 10}, 20)
	secondNear := addTestMonster(sim, "dungeon_mob", Vec2{X: 65, Y: 10}, 20)
	res := TickResult{Tick: sim.tick, Level: sim.currentLevel}

	sim.aggroMonsterOnHit(first, sim.playerID, "corr_first_group", &res)
	eventsAfterFirst := countEvents(res, "monster_aggro")
	sim.aggroMonsterOnHit(firstNear, sim.playerID, "corr_first_group_repeat", &res)
	if got := countEvents(res, "monster_aggro"); got != eventsAfterFirst {
		t.Fatalf("repeat aggro in same group emitted new events: got %d want %d events=%+v", got, eventsAfterFirst, res.Events)
	}

	sim.aggroMonsterOnHit(second, sim.playerID, "corr_second_group", &res)
	for _, monster := range []*entity{first, firstNear, second, secondNear} {
		if monster.aiTargetPlayerID != sim.playerID || monster.aiMode != monsterAIModeChase {
			t.Fatalf("monster %d target/mode = %d/%s, want %d/%s", monster.id, monster.aiTargetPlayerID, monster.aiMode, sim.playerID, monsterAIModeChase)
		}
		if !eventForEntity(res, "monster_aggro", monster.id) {
			t.Fatalf("missing monster_aggro for %d in events %+v", monster.id, res.Events)
		}
	}
}

func TestAggroOnLethalHitPropagatesToNearbyMonsterGroup(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_lethal_group_aggro", "lethal_group_aggro", rules, "dungeon_levels")
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
	placeDefaultPlayerOnLevel(t, sim, level, Vec2{X: 2, Y: 5})
	sim.syncCompatibilityFields()

	primary := addTestMonster(sim, "dungeon_mob", Vec2{X: 20, Y: 10}, 1)
	near := addTestMonster(sim, "dungeon_mob", Vec2{X: 25, Y: 10}, 20)
	far := addTestMonster(sim, "dungeon_mob", Vec2{X: 45, Y: 10}, 20)
	res := TickResult{Tick: sim.tick, Level: sim.currentLevel}

	outcome := sim.damageMonsterByPlayer(primary, sim.playerID, "corr_lethal_group", &res, DamageRange{Min: 99, Max: 99})
	if !outcome.Hit || outcome.Blocked || primary.hp != 0 {
		t.Fatalf("setup expected lethal hit outcome=%+v primary_hp=%d events=%+v", outcome, primary.hp, res.Events)
	}
	if primary.aiMode != monsterAIModeIdle || primary.aiTargetPlayerID != 0 {
		t.Fatalf("dead primary target/mode = %d/%s, want no aggro on dead source", primary.aiTargetPlayerID, primary.aiMode)
	}
	if near.aiTargetPlayerID != sim.playerID || near.aiMode != monsterAIModeChase {
		t.Fatalf("near monster target/mode = %d/%s, want %d/%s", near.aiTargetPlayerID, near.aiMode, sim.playerID, monsterAIModeChase)
	}
	if far.aiTargetPlayerID != 0 || far.aiMode != monsterAIModeIdle {
		t.Fatalf("far monster target/mode = %d/%s, want idle outside group radius", far.aiTargetPlayerID, far.aiMode)
	}
	if eventForEntity(res, "monster_aggro", primary.id) {
		t.Fatalf("dead primary should not emit monster_aggro: %+v", res.Events)
	}
	if !eventForEntity(res, "monster_aggro", near.id) {
		t.Fatalf("missing nearby monster_aggro after lethal hit: %+v", res.Events)
	}
	if eventForEntity(res, "monster_aggro", far.id) {
		t.Fatalf("unexpected far monster_aggro after lethal hit: %+v", res.Events)
	}
}

func TestAggroOnHitAlsoAggrosMonstersWithAttackerInRange(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_attack_range_aggro", "range_aggro", rules, "dungeon_levels")
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
	placeDefaultPlayerOnLevel(t, sim, level, Vec2{X: 2, Y: 5})
	sim.syncCompatibilityFields()

	primary := addTestMonster(sim, "dungeon_mob", Vec2{X: 20, Y: 10}, 20)
	attackerRange := addTestMonster(sim, "dungeon_mob", Vec2{X: 7, Y: 5}, 20)
	outsideBoth := addTestMonster(sim, "dungeon_mob", Vec2{X: 45, Y: 10}, 20)
	res := TickResult{Tick: sim.tick, Level: sim.currentLevel}

	sim.aggroMonsterOnHit(primary, sim.playerID, "corr_attacker_range", &res)

	for _, monster := range []*entity{primary, attackerRange} {
		if monster.aiTargetPlayerID != sim.playerID || monster.aiMode != monsterAIModeChase {
			t.Fatalf("monster %d target/mode = %d/%s, want %d/%s", monster.id, monster.aiTargetPlayerID, monster.aiMode, sim.playerID, monsterAIModeChase)
		}
	}
	if outsideBoth.aiTargetPlayerID != 0 || outsideBoth.aiMode != monsterAIModeIdle {
		t.Fatalf("outside monster target/mode = %d/%s, want idle outside attacker and group radius", outsideBoth.aiTargetPlayerID, outsideBoth.aiMode)
	}

	aggroEvents := map[string]bool{}
	for _, ev := range res.Events {
		if ev.EventType == "monster_aggro" {
			aggroEvents[ev.EntityID] = true
		}
	}
	for _, monster := range []*entity{primary, attackerRange} {
		if !aggroEvents[idStr(monster.id)] {
			t.Fatalf("missing monster_aggro for %d in events %+v", monster.id, res.Events)
		}
	}
	if aggroEvents[idStr(outsideBoth.id)] {
		t.Fatalf("unexpected outside monster_aggro for %d in events %+v", outsideBoth.id, res.Events)
	}
}

func TestAggroOnHitUsesAssistRadiusWithoutGrowingPassiveAggro(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_assist_radius_aggro", "assist_radius_aggro", rules, "dungeon_levels")
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
	placeDefaultPlayerOnLevel(t, sim, level, Vec2{X: 2, Y: 5})
	player := level.entities[sim.playerID]
	sim.syncCompatibilityFields()

	def := sim.rules.Monsters["dungeon_mob"]
	if def.effectiveAssistRadius() <= def.AggroRadius {
		t.Fatalf("test requires assist radius > aggro radius, got aggro %.2f assist %.2f", def.AggroRadius, def.effectiveAssistRadius())
	}
	primary := addTestMonster(sim, "dungeon_mob", Vec2{X: 20, Y: 10}, 20)
	assistOnly := addTestMonster(sim, "dungeon_mob", Vec2{X: player.pos.X + def.AggroRadius + 1, Y: player.pos.Y}, 20)
	outside := addTestMonster(sim, "dungeon_mob", Vec2{X: 45, Y: 10}, 20)

	passive := TickResult{Tick: sim.tick, Level: sim.currentLevel}
	sim.updateMonsterAIMode(assistOnly, player, def, assistOnly.aiMode, &passive)
	if assistOnly.aiMode != monsterAIModeIdle || eventForEntity(passive, "monster_aggro", assistOnly.id) {
		t.Fatalf("assist-only monster passively aggroed at distance beyond aggro radius: mode=%s events=%+v", assistOnly.aiMode, passive.Events)
	}

	res := TickResult{Tick: sim.tick, Level: sim.currentLevel}
	sim.aggroMonsterOnHit(primary, sim.playerID, "corr_assist_radius", &res)
	if assistOnly.aiTargetPlayerID != sim.playerID || assistOnly.aiMode != monsterAIModeChase {
		t.Fatalf("assist-only monster target/mode = %d/%s, want %d/%s", assistOnly.aiTargetPlayerID, assistOnly.aiMode, sim.playerID, monsterAIModeChase)
	}
	if outside.aiTargetPlayerID != 0 || outside.aiMode != monsterAIModeIdle {
		t.Fatalf("outside monster target/mode = %d/%s, want idle outside assist radius", outside.aiTargetPlayerID, outside.aiMode)
	}
	if !eventForEntity(res, "monster_aggro", assistOnly.id) {
		t.Fatalf("missing assist radius monster_aggro: %+v", res.Events)
	}
	if eventForEntity(res, "monster_aggro", outside.id) {
		t.Fatalf("unexpected outside monster_aggro: %+v", res.Events)
	}
}
