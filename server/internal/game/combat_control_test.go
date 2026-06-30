package game

import "testing"

func TestRangedProjectileGolden(t *testing.T) {
	var golden struct {
		Cases []struct {
			Name                       string   `json:"name"`
			WorldID                    string   `json:"world_id"`
			Seed                       string   `json:"seed"`
			BaseHitChance              *float64 `json:"base_hit_chance"`
			PlayerPosition             Vec2     `json:"player_position"`
			ExpectedEvent              string   `json:"expected_event"`
			ExpectedMonsterHPUnchanged bool     `json:"expected_monster_hp_unchanged"`
			ExpectedPlayerHP           int      `json:"expected_player_hp"`
			ExpectedMonsterDead        bool     `json:"expected_monster_dead"`
		} `json:"cases"`
	}
	loadGolden(t, "ranged_projectile.json", &golden)
	for _, tc := range golden.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			rules := loadRules(t)
			if tc.BaseHitChance != nil {
				rules = rulesCopyWithHitChance(rules, *tc.BaseHitChance)
			}
			sim := rangedLabWithEquippedBow(t, rules, tc.Seed)
			sim.entities[sim.playerID].pos = tc.PlayerPosition
			monster := firstEntityByKind(sim, monsterEntity)
			initialMonsterHP := monster.hp
			fire := sim.Tick([]Input{{MessageID: "fire", CorrelationID: "corr_ranged", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
			assertAck(t, fire, "fire")
			if firstEntityByKind(sim, projectileEntity) == nil && sim.autoNav == nil && !hasEvent(fire, tc.ExpectedEvent) {
				t.Fatalf("no projectile spawned, auto-nav queued, or expected event on fire tick: %+v", fire)
			}
			var impact TickResult
			resolved := false
			for i := 0; i < 150; i++ {
				r := sim.Tick(nil)
				if len(r.Events) > 0 {
					impact = r
					resolved = true
					break
				}
			}
			if !resolved {
				t.Fatal("projectile scenario did not resolve within tick budget")
			}
			if tc.ExpectedEvent != "" && !hasEvent(impact, tc.ExpectedEvent) {
				t.Fatalf("impact events = %+v, want %s", impact.Events, tc.ExpectedEvent)
			}
			if tc.ExpectedMonsterHPUnchanged && monster.hp != initialMonsterHP {
				t.Fatalf("monster hp = %d, want unchanged %d", monster.hp, initialMonsterHP)
			}
			if tc.ExpectedMonsterDead && monster.hp != 0 {
				t.Fatalf("monster hp = %d, want dead", monster.hp)
			}
			if tc.ExpectedPlayerHP != 0 {
				player := sim.entities[sim.playerID]
				if player.hp != tc.ExpectedPlayerHP {
					t.Fatalf("player hp = %d, want %d", player.hp, tc.ExpectedPlayerHP)
				}
			}
		})
	}
}

func rangedLabWithEquippedBow(t *testing.T, rules *Rules, seed string) *Sim {
	t.Helper()
	sim, err := NewSimWithWorld("sess_ranged", seed, rules, "ranged_lab")
	if err != nil {
		t.Fatalf("ranged_lab world: %v", err)
	}
	pickup := sim.Tick([]Input{{MessageID: "pick_bow", CorrelationID: "corr_pick", Type: "action_intent", Action: &ActionIntent{TargetID: "1002"}}})
	assertAck(t, pickup, "pick_bow")
	equip := sim.Tick([]Input{{MessageID: "equip_bow", CorrelationID: "corr_equip", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "1004", Slot: mainHandSlot}}})
	assertAck(t, equip, "equip_bow")
	return sim
}

func combatControlLabWithEquippedBow(t *testing.T, rules *Rules, seed string) *Sim {
	t.Helper()
	sim, err := NewSimWithWorld("sess_combat_control", seed, rules, "combat_control_lab")
	if err != nil {
		t.Fatalf("combat_control_lab world: %v", err)
	}
	pickup := sim.Tick([]Input{{MessageID: "pick_bow", CorrelationID: "corr_pick", Type: "action_intent", Action: &ActionIntent{TargetID: "1002"}}})
	assertAck(t, pickup, "pick_bow")
	equip := sim.Tick([]Input{{MessageID: "equip_bow", CorrelationID: "corr_equip", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "1004", Slot: mainHandSlot}}})
	assertAck(t, equip, "equip_bow")
	return sim
}

func equipStaticBow(t *testing.T, sim *Sim) {
	t.Helper()
	addTestInventoryItem(sim, &invItem{instanceID: 5000, itemDefID: "training_bow"})
	equip := sim.Tick([]Input{{MessageID: "equip_bow", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "5000", Slot: mainHandSlot}}})
	assertAck(t, equip, "equip_bow")
}

func rulesWithTrainingBowReach(t *testing.T, reach float64) *Rules {
	t.Helper()
	base := loadRules(t)
	copyRules := *base
	items := make(map[string]ItemDef, len(base.Items))
	for k, v := range base.Items {
		items[k] = v
	}
	bow := items["training_bow"]
	bow.Reach = &reach
	items["training_bow"] = bow
	copyRules.Items = items
	return &copyRules
}

func rulesWithHitChance(t *testing.T, chance float64) *Rules {
	t.Helper()
	return rulesCopyWithHitChance(loadRules(t), chance)
}

func rulesCopyWithHitChance(base *Rules, chance float64) *Rules {
	copyRules := *base
	copyRules.Combat = base.Combat
	copyRules.Combat.BaseHitChance = chance
	progress := base.CharacterProgression
	derived := make(map[string]LinearStatFormula, len(base.CharacterProgression.DerivedStats))
	for key, formula := range base.CharacterProgression.DerivedStats {
		derived[key] = formula
	}
	hit := derived["hit_chance"]
	hit.Type = "linear"
	hit.Base = chance
	hit.PerStr = 0
	hit.PerDex = 0
	hit.PerVit = 0
	hit.PerMagic = 0
	hit.Stat = ""
	hit.Scale = 0
	hit.Offset = 0
	hit.Denominator = 0
	hit.Min = &chance
	hit.Max = &chance
	derived["hit_chance"] = hit
	progress.DerivedStats = derived
	copyRules.CharacterProgression = progress
	return &copyRules
}

func TestProjectileBusyRejectsSecondFire(t *testing.T) {
	sim := rangedLabWithEquippedBow(t, loadRules(t), "cafebabecafebabe")
	monster := firstEntityByKind(sim, monsterEntity)
	first := sim.Tick([]Input{{MessageID: "fire1", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertAck(t, first, "fire1")
	// Wait for the approach to complete and the first projectile to spawn before
	// testing the busy-rejection.  With wall-adjacent cells excluded from
	// pathfinding, the approach goal is farther from the spawn point.
	for i := 0; i < 100; i++ {
		if firstEntityByKind(sim, projectileEntity) != nil {
			break
		}
		sim.Tick(nil)
	}
	if firstEntityByKind(sim, projectileEntity) == nil {
		t.Fatal("fire1 did not spawn projectile within 100 ticks")
	}
	second := sim.Tick([]Input{{MessageID: "fire2", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertReject(t, second, "fire2", "projectile_busy")
}

func TestBasicAttackCooldownRejectsRapidMelee(t *testing.T) {
	sim := MustNewSim("sess_basic_attack_cooldown", "01", loadRules(t))
	monster := firstEntityByKind(sim, monsterEntity)
	monster.pos = sim.entities[sim.playerID].pos
	monster.hp = 50
	monster.maxHP = 50
	interval := sim.DerivedStatsView().AttackIntervalTicks

	first := sim.Tick([]Input{{MessageID: "hit1", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertAck(t, first, "hit1")
	second := sim.Tick([]Input{{MessageID: "hit2", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertReject(t, second, "hit2", "basic_attack_on_cooldown")
	for i := 0; i < interval-1; i++ {
		sim.Tick(nil)
	}
	third := sim.Tick([]Input{{MessageID: "hit3", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertAck(t, third, "hit3")
}

func TestDirectionalAttackRejectsInvalidDirection(t *testing.T) {
	sim := MustNewSim("sess_directional_invalid", "01", loadRules(t))
	r := sim.Tick([]Input{{MessageID: "dir", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{}}}})
	assertReject(t, r, "dir", "invalid_direction")
}

func TestDirectionalMeleeHitsMonsterInFront(t *testing.T) {
	sim := MustNewSim("sess_directional_melee", "01", loadRules(t))
	monster := firstEntityByKind(sim, monsterEntity)
	monster.pos = Vec2{X: 11.2, Y: 5}
	monster.hp = 10
	monster.maxHP = 10

	r := sim.Tick([]Input{{MessageID: "dir", CorrelationID: "corr_dir", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
	assertAck(t, r, "dir")
	if !hasEvent(r, "monster_damaged") {
		t.Fatalf("directional melee events = %+v, want monster_damaged", r.Events)
	}
	if monster.hp >= monster.maxHP {
		t.Fatalf("monster hp = %d, want reduced", monster.hp)
	}
}

func TestDirectionalMeleeMissesBehindAndOutsideCapsule(t *testing.T) {
	t.Run("behind", func(t *testing.T) {
		sim := MustNewSim("sess_directional_behind", "01", loadRules(t))
		monster := firstEntityByKind(sim, monsterEntity)
		monster.pos = Vec2{X: 9.2, Y: 5}
		initialHP := monster.hp
		r := sim.Tick([]Input{{MessageID: "dir", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
		assertAck(t, r, "dir")
		if len(r.Events) != 0 {
			t.Fatalf("behind swing emitted events: %+v", r.Events)
		}
		if monster.hp != initialHP {
			t.Fatalf("behind monster hp = %d, want %d", monster.hp, initialHP)
		}
	})

	t.Run("outside capsule", func(t *testing.T) {
		sim := MustNewSim("sess_directional_lateral", "01", loadRules(t))
		monster := firstEntityByKind(sim, monsterEntity)
		monster.pos = Vec2{X: 11.0, Y: 6.2}
		initialHP := monster.hp
		r := sim.Tick([]Input{{MessageID: "dir", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
		assertAck(t, r, "dir")
		if len(r.Events) != 0 {
			t.Fatalf("outside capsule swing emitted events: %+v", r.Events)
		}
		if monster.hp != initialHP {
			t.Fatalf("outside capsule monster hp = %d, want %d", monster.hp, initialHP)
		}
	})
}

func TestDirectionalMeleeTieBreaksByEntityID(t *testing.T) {
	sim := MustNewSim("sess_directional_tie", "01", loadRules(t))
	first := firstEntityByKind(sim, monsterEntity)
	first.pos = Vec2{X: 11, Y: 4.8}
	first.hp = 10
	first.maxHP = 10
	second := addTestMonster(sim, "training_dummy", Vec2{X: 11, Y: 5.2}, 10)

	r := sim.Tick([]Input{{MessageID: "dir", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
	assertAck(t, r, "dir")
	if first.hp >= first.maxHP {
		t.Fatalf("first monster hp = %d, want damaged", first.hp)
	}
	if second.hp != second.maxHP {
		t.Fatalf("second monster hp = %d, want unchanged %d", second.hp, second.maxHP)
	}
}

func TestDirectionalMeleeStopsMovementAndAcksEmptySwing(t *testing.T) {
	sim := MustNewSim("sess_directional_stop", "01", loadRules(t))
	monster := firstEntityByKind(sim, monsterEntity)
	monster.pos = Vec2{X: 9, Y: 5}
	move := sim.Tick([]Input{{MessageID: "move", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 3}}})
	assertAck(t, move, "move")
	beforeAttack := sim.entities[sim.playerID].pos

	r := sim.Tick([]Input{{MessageID: "dir", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
	assertAck(t, r, "dir")
	if len(r.Events) != 0 {
		t.Fatalf("empty directional swing emitted events: %+v", r.Events)
	}
	if sim.move != nil {
		t.Fatalf("directional attack did not clear movement: %+v", sim.move)
	}
	if sim.entities[sim.playerID].pos != beforeAttack {
		t.Fatalf("directional attack moved player from %+v to %+v", beforeAttack, sim.entities[sim.playerID].pos)
	}
}

func TestDirectionalRangedFreeShotHitsAndOmitsTargetID(t *testing.T) {
	sim := combatControlLabWithEquippedBow(t, rulesWithHitChance(t, 1.0), "cafebabecafebabe")
	player := sim.entities[sim.playerID]
	// Same Y as monster so A* approach is horizontal; 11.5 units keeps the monster
	// outside effective aggro radius (10) until the projectile hits and triggers aggro.
	player.pos = Vec2{X: 1.5, Y: 5}
	monster := firstEntityByKind(sim, monsterEntity)
	monster.hp = 20
	monster.maxHP = 20
	initialDistance := distance(monster.pos, player.pos)

	fire := sim.Tick([]Input{{MessageID: "dir_fire", CorrelationID: "corr_dir_fire", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
	assertAck(t, fire, "dir_fire")
	spawn := firstChangeEntityByType(fire, projectileEntity)
	if spawn == nil {
		t.Fatalf("directional ranged did not spawn projectile: %+v", fire.Changes)
	}
	if spawn.TargetID != "" {
		t.Fatalf("free-shot projectile target_id = %q, want omitted", spawn.TargetID)
	}

	var impact TickResult
	for i := 0; i < 150; i++ {
		impact = sim.Tick(nil)
		if hasEvent(impact, "monster_damaged") || hasEvent(impact, "monster_killed") || hasEvent(impact, "attack_missed") {
			break
		}
	}
	if !hasEvent(impact, "monster_damaged") {
		t.Fatalf("directional ranged impact events = %+v, want monster_damaged", impact.Events)
	}
	if !hasEvent(impact, "monster_aggro") {
		t.Fatalf("directional ranged impact events = %+v, want monster_aggro", impact.Events)
	}
	if monster.aiTargetPlayerID != sim.playerID || monster.aiMode != monsterAIModeChase {
		t.Fatalf("monster ai target/mode = %d/%s, want %d/%s", monster.aiTargetPlayerID, monster.aiMode, sim.playerID, monsterAIModeChase)
	}

	moved := false
	for i := 0; i < 10; i++ {
		sim.Tick(nil)
		if distance(monster.pos, player.pos) < initialDistance-0.01 {
			moved = true
			break
		}
	}
	if !moved {
		t.Fatalf("aggroed monster did not move toward player: start dist %.3f now %.3f", initialDistance, distance(monster.pos, player.pos))
	}
}

func TestDirectionalRangedProjectileBusy(t *testing.T) {
	sim := combatControlLabWithEquippedBow(t, loadRules(t), "cafebabecafebabe")
	first := sim.Tick([]Input{{MessageID: "fire1", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
	assertAck(t, first, "fire1")
	second := sim.Tick([]Input{{MessageID: "fire2", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
	assertReject(t, second, "fire2", "projectile_busy")
}

func TestDirectionalRangedProjectileBlockedAndExpires(t *testing.T) {
	t.Run("closed interactable blocks", func(t *testing.T) {
		sim, err := NewSimWithWorld("sess_directional_blocked", "01", loadRules(t), "door_lab")
		if err != nil {
			t.Fatalf("door world: %v", err)
		}
		equipStaticBow(t, sim)
		fire := sim.Tick([]Input{{MessageID: "fire", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
		assertAck(t, fire, "fire")
		var resolved TickResult
		for i := 0; i < 10; i++ {
			resolved = sim.Tick(nil)
			if hasEvent(resolved, "projectile_blocked") {
				break
			}
		}
		if !hasEvent(resolved, "projectile_blocked") {
			t.Fatalf("blocked projectile events = %+v, want projectile_blocked", resolved.Events)
		}
	})

	t.Run("expires without hit", func(t *testing.T) {
		rules := rulesWithTrainingBowReach(t, 2.0)
		sim := combatControlLabWithEquippedBow(t, rules, "cafebabecafebabe")
		sim.entities[sim.playerID].pos = Vec2{X: 3, Y: 5}
		fire := sim.Tick([]Input{{MessageID: "fire", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{Y: 1}}}})
		assertAck(t, fire, "fire")
		var resolved TickResult
		for i := 0; i < 10; i++ {
			resolved = sim.Tick(nil)
			if hasEvent(resolved, "projectile_expired") {
				break
			}
		}
		if !hasEvent(resolved, "projectile_expired") {
			t.Fatalf("expired projectile events = %+v, want projectile_expired", resolved.Events)
		}
	})
}

func TestRangedAutoApproachThenFire(t *testing.T) {
	sim := rangedLabWithEquippedBow(t, loadRules(t), "cafebabecafebabe")
	monster := firstEntityByKind(sim, monsterEntity)
	sim.entities[sim.playerID].pos = Vec2{X: 2, Y: 8}
	monster.pos = Vec2{X: 12, Y: 5}
	r := sim.Tick([]Input{{MessageID: "far_fire", CorrelationID: "corr_far", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertAck(t, r, "far_fire")
	sawProjectile := false
	sawImpact := false
	for i := 0; i < 150 && !sawImpact; i++ {
		r := sim.Tick(nil)
		for _, c := range r.Changes {
			if c.Op == OpEntitySpawn && c.Entity != nil && c.Entity.Type == projectileEntity {
				sawProjectile = true
			}
		}
		if hasEvent(r, "monster_damaged") || hasEvent(r, "attack_missed") || hasEvent(r, "projectile_blocked") {
			sawImpact = true
		}
	}
	if !sawProjectile {
		t.Fatal("auto-approach did not spawn projectile")
	}
	if !sawImpact {
		t.Fatal("auto-approach projectile did not resolve")
	}
}

func TestRangedDummyDropsSeparatedLootItems(t *testing.T) {
	sim := rangedLabWithEquippedBow(t, rulesWithHitChance(t, 1.0), "cafebabecafebabe")
	monster := firstEntityByKind(sim, monsterEntity)
	monster.hp = 1
	r := sim.Tick([]Input{{MessageID: "kill", CorrelationID: "corr_kill", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertAck(t, r, "kill")
	for i := 0; i < 150 && !hasEvent(r, "monster_killed"); i++ {
		r = sim.Tick(nil)
	}
	if !hasEvent(r, "monster_killed") {
		t.Fatalf("ranged kill did not resolve: %+v", r.Events)
	}

	want := map[string]bool{
		"gold":        false,
		"quest_leaf":  false,
		"red_potion":  false,
		"blue_potion": false,
	}
	poolExtras := map[string]bool{
		"upgrade_shard": false,
		"renew_stone":   false,
	}
	positions := map[Vec2]string{}
	for _, c := range r.Changes {
		if c.Op != OpEntitySpawn || c.Entity == nil || c.Entity.Type != lootEntity {
			continue
		}
		itemDefID := c.Entity.ItemDefID
		if _, ok := want[itemDefID]; !ok {
			if _, poolOK := poolExtras[itemDefID]; poolOK {
				poolExtras[itemDefID] = true
				if positions[c.Entity.Position] != "" {
					t.Fatalf("loot overlap at %+v: %s and %s", c.Entity.Position, positions[c.Entity.Position], itemDefID)
				}
				if sim.lootDropBlocked(c.Entity.Position) {
					t.Fatalf("loot spawned inside blocked geometry at %+v", c.Entity.Position)
				}
				positions[c.Entity.Position] = itemDefID
				continue
			}
			t.Fatalf("unexpected ranged loot %s in %+v", itemDefID, r.Changes)
		}
		if positions[c.Entity.Position] != "" {
			t.Fatalf("loot overlap at %+v: %s and %s", c.Entity.Position, positions[c.Entity.Position], itemDefID)
		}
		if sim.lootDropBlocked(c.Entity.Position) {
			t.Fatalf("loot spawned inside blocked geometry at %+v", c.Entity.Position)
		}
		positions[c.Entity.Position] = itemDefID
		want[itemDefID] = true
	}
	for itemDefID, seen := range want {
		if !seen {
			t.Fatalf("missing ranged loot %s in %+v", itemDefID, r.Changes)
		}
	}
}

func TestRangedBowLootRequiresMeleeReach(t *testing.T) {
	sim := rangedLabWithEquippedBow(t, rulesWithHitChance(t, 1.0), "cafebabecafebabe")
	if sim.playerActionReach() != 12.0 {
		t.Fatalf("playerActionReach = %v, want bow reach 12.0", sim.playerActionReach())
	}
	if sim.playerMeleeReach() != sim.rules.Combat.UnarmedReach {
		t.Fatalf("playerMeleeReach = %v, want unarmed %v", sim.playerMeleeReach(), sim.rules.Combat.UnarmedReach)
	}

	sim.entities[sim.playerID].pos = Vec2{X: 2, Y: 8}
	monster := firstEntityByKind(sim, monsterEntity)
	monster.hp = 1
	fire := sim.Tick([]Input{{MessageID: "kill", CorrelationID: "corr_kill", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertAck(t, fire, "kill")
	var loot *entity
	for i := 0; i < 60; i++ {
		r := sim.Tick(nil)
		if hasEvent(r, "monster_killed") {
			for _, c := range r.Changes {
				if c.Op == OpEntitySpawn && c.Entity != nil && c.Entity.Type == lootEntity {
					loot = sim.findEntity(c.Entity.ID)
					break
				}
			}
		}
		if loot != nil {
			break
		}
	}
	if loot == nil {
		t.Fatal("missing loot after ranged kill")
	}
	if sim.inMeleeRange(loot) {
		t.Fatalf("player at %+v should not be in melee range of loot at %+v with bow equipped", sim.entities[sim.playerID].pos, loot.pos)
	}

	pickup := sim.Tick([]Input{{MessageID: "loot_pick", CorrelationID: "corr_loot", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(loot.id)}}})
	assertAck(t, pickup, "loot_pick")
	if sim.autoNav == nil {
		t.Fatal("loot pickup from range should queue auto-nav, not dispatch immediately")
	}
	pickupEvent := "item_picked_up"
	if loot.itemDefID == goldItemDefID {
		pickupEvent = "gold_picked_up"
	}
	if hasEvent(pickup, pickupEvent) {
		t.Fatal("loot picked up instantly from ranged distance")
	}

	picked := false
	for i := 0; i < 150; i++ {
		r := sim.Tick(nil)
		if hasEvent(r, pickupEvent) {
			picked = true
			break
		}
	}
	if !picked {
		t.Fatal("auto-nav did not complete loot pickup within tick budget")
	}
}
func TestRangedBlockedLineAutoMovesUntilClearThenFires(t *testing.T) {
	sim := rangedLabWithEquippedBow(t, loadRules(t), "deadbeefdeadbeef")
	monster := firstEntityByKind(sim, monsterEntity)
	sim.entities[sim.playerID].pos = Vec2{X: 2, Y: 8}
	if sim.hasClearRangedShot(sim.entities[sim.playerID].pos, monster) {
		t.Fatal("test setup has clear shot; want wall-blocked line")
	}

	r := sim.Tick([]Input{{MessageID: "covered_fire", CorrelationID: "corr_covered", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertAck(t, r, "covered_fire")
	if sim.autoNav == nil {
		t.Fatal("blocked ranged click fired immediately; want auto-nav")
	}
	if firstEntityByKind(sim, projectileEntity) != nil {
		t.Fatal("projectile spawned before line was clear")
	}

	sawProjectile := false
	sawImpact := false
	for i := 0; i < 300 && !sawImpact; i++ {
		r := sim.Tick(nil)
		for _, c := range r.Changes {
			if c.Op == OpEntitySpawn && c.Entity != nil && c.Entity.Type == projectileEntity {
				sawProjectile = true
				player := sim.entities[sim.playerID]
				if !sim.hasClearRangedShot(player.pos, monster) {
					t.Fatalf("projectile spawned without clear shot from %+v to %+v", player.pos, monster.pos)
				}
				playerMonsterDistance := distance(player.pos, monster.pos)
				if meleeInRange(playerMonsterDistance, sim.rules.Combat.UnarmedReach, monsterRadius) {
					t.Fatalf("ranged auto-nav entered melee range at %+v", player.pos)
				}
			}
		}
		if hasEvent(r, "monster_damaged") || hasEvent(r, "monster_killed") || hasEvent(r, "attack_missed") || hasEvent(r, "projectile_blocked") {
			sawImpact = true
			if hasEvent(r, "projectile_blocked") {
				t.Fatalf("projectile was still blocked after auto-nav: %+v", r.Events)
			}
		}
	}
	if !sawProjectile {
		t.Fatal("auto-nav never spawned projectile")
	}
	if !sawImpact {
		t.Fatal("auto-nav projectile did not resolve")
	}
}
