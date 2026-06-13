package game

import "testing"

func TestEliteAuraAppliesOnlyToNearbyPackFollowers(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.DungeonGeneration.MonsterPlacement.EliteAura = &EliteAuraRules{
		ID:                 "elite_command",
		Radius:             4,
		DamageBonusPercent: 50,
	}
	sim := MustNewSim("sess_elite_aura", "elite_aura_seed", rules)
	level := sim.activeLevel()

	leader := &entity{
		id:                sim.alloc(),
		kind:              monsterEntity,
		pos:               Vec2{X: 10, Y: 10},
		hp:                20,
		maxHP:             20,
		monsterDefID:      "dungeon_mob",
		monsterPackID:     "pack_01",
		monsterPackLeader: true,
	}
	follower := &entity{
		id:            sim.alloc(),
		kind:          monsterEntity,
		pos:           Vec2{X: 12, Y: 10},
		hp:            10,
		maxHP:         10,
		monsterDefID:  "dungeon_mob",
		monsterPackID: "pack_01",
	}
	level.entities[leader.id] = leader
	level.entities[follower.id] = follower

	got := sim.applyEliteAuraToMonsterDamage(follower, DamageRange{Min: 10, Max: 12})
	if got != (DamageRange{Min: 15, Max: 18}) {
		t.Fatalf("aura damage = %+v, want 15..18", got)
	}

	leaderDamage := sim.applyEliteAuraToMonsterDamage(leader, DamageRange{Min: 10, Max: 12})
	if leaderDamage != (DamageRange{Min: 10, Max: 12}) {
		t.Fatalf("leader damage = %+v, want unchanged", leaderDamage)
	}

	follower.pos = Vec2{X: 20, Y: 10}
	farDamage := sim.applyEliteAuraToMonsterDamage(follower, DamageRange{Min: 10, Max: 12})
	if farDamage != (DamageRange{Min: 10, Max: 12}) {
		t.Fatalf("far follower damage = %+v, want unchanged", farDamage)
	}

	follower.pos = Vec2{X: 12, Y: 10}
	leader.hp = 0
	deadLeaderDamage := sim.applyEliteAuraToMonsterDamage(follower, DamageRange{Min: 10, Max: 12})
	if deadLeaderDamage != (DamageRange{Min: 10, Max: 12}) {
		t.Fatalf("dead leader damage = %+v, want unchanged", deadLeaderDamage)
	}

	leader.hp = 20
	follower.monsterPackID = "pack_02"
	otherPackDamage := sim.applyEliteAuraToMonsterDamage(follower, DamageRange{Min: 10, Max: 12})
	if otherPackDamage != (DamageRange{Min: 10, Max: 12}) {
		t.Fatalf("other pack damage = %+v, want unchanged", otherPackDamage)
	}

	follower.monsterPackID = ""
	noPackDamage := sim.applyEliteAuraToMonsterDamage(follower, DamageRange{Min: 10, Max: 12})
	if noPackDamage != (DamageRange{Min: 10, Max: 12}) {
		t.Fatalf("no pack damage = %+v, want unchanged", noPackDamage)
	}
}

func TestGeneratedDungeonMonstersPreservePackMetadata(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.DungeonGeneration.MonsterPlacement.ElitePackChance = 100
	sim, err := NewSimWithWorld("sess_elite_aura_metadata", "v112_pack_metadata", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	descendFromCurrentLevel(t, sim, "descend_to_dungeon")

	packMembers := 0
	packLeaders := 0
	for _, monster := range liveDungeonMonsters(sim.activeLevel()) {
		if monster.monsterPackID == "" {
			continue
		}
		packMembers++
		if monster.monsterPackLeader {
			packLeaders++
			if monster.monsterRarityID != "champion" {
				t.Fatalf("leader rarity = %s, want champion", monster.monsterRarityID)
			}
		}
	}
	if packMembers < rules.DungeonGeneration.MonsterPlacement.PackSize.Min {
		t.Fatalf("pack members = %d, want at least %d", packMembers, rules.DungeonGeneration.MonsterPlacement.PackSize.Min)
	}
	if packLeaders < rules.DungeonGeneration.MonsterPlacement.PackCount.Min {
		t.Fatalf("pack leaders = %d, want at least %d", packLeaders, rules.DungeonGeneration.MonsterPlacement.PackCount.Min)
	}
}
