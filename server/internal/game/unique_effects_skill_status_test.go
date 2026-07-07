package game

import "testing"

func TestBranchUniqueWarbrandGoreAppliesBurnOnGoreStrike(t *testing.T) {
	rules := cloneRules(loadRules(t))
	sim := MustNewSim("sess_warbrand_gore", "warbrand_gore", rules)
	sim.progression.Level = 24
	sim.savePlayer(sim.defaultPlayer())
	player := sim.entities[sim.playerID]
	target := uniqueTestMonster(sim, Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, 60)
	equipUniqueTestEffect(t, sim, "warbrand_gore", 9910, "barbarian_axe", mainHandSlot)

	res := &TickResult{}
	sim.damageMonsterByPlayerSkillTypedWithID(target, player.id, "gore_strike", "warbrand_hit", res, DamageRange{Min: 20, Max: 20}, damageTypeForce)
	if _, ok := eventAmount(*res, "skill_effect_started", "warbrand_gore"); !ok {
		t.Fatalf("events = %+v, want warbrand burn start", res.Events)
	}
	if !containsStringValue(target.effectIDs, "burning") {
		t.Fatalf("target effects = %v, want burning", target.effectIDs)
	}
}

func TestBranchUniqueNightbloomBlossomAppliesPoisonOnDeathBlossom(t *testing.T) {
	rules := cloneRules(loadRules(t))
	sim := MustNewSim("sess_nightbloom", "nightbloom", rules)
	sim.progression.Level = 24
	sim.progression.CharacterClass = "rogue"
	sim.progression.BaseStats = BaseStatsView{Str: 4, Dex: 8, Vit: 5, Magic: 4}
	sim.savePlayer(sim.defaultPlayer())
	player := sim.entities[sim.playerID]
	target := uniqueTestMonster(sim, Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, 60)
	equipUniqueTestEffect(t, sim, "nightbloom_blossom", 9911, "dagger", mainHandSlot)

	res := &TickResult{}
	sim.damageMonsterByPlayerSkillTypedWithID(target, player.id, "death_blossom", "nightbloom_hit", res, DamageRange{Min: 18, Max: 18}, damageTypeForce)
	if _, ok := eventAmount(*res, "skill_effect_started", "nightbloom_blossom"); !ok {
		t.Fatalf("events = %+v, want nightbloom poison start", res.Events)
	}
	if !containsStringValue(target.effectIDs, "poisoned") {
		t.Fatalf("target effects = %v, want poisoned", target.effectIDs)
	}
}

func TestBranchUniqueRimecoilLanceSplashesSlow(t *testing.T) {
	rules := cloneRules(loadRules(t))
	sim := MustNewSim("sess_rimecoil", "rimecoil", rules)
	sim.progression.Level = 16
	sim.savePlayer(sim.defaultPlayer())
	player := sim.entities[sim.playerID]
	primary := uniqueTestMonster(sim, Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, 60)
	nearby := uniqueTestMonster(sim, Vec2{X: primary.pos.X + 1, Y: primary.pos.Y}, 60)
	equipUniqueTestEffect(t, sim, "rimecoil_lance", 9912, "sorcerer_staff", mainHandSlot)

	res := &TickResult{}
	sim.damageMonsterByPlayerSkillTypedWithID(primary, player.id, "glacial_lance", "rimecoil_hit", res, DamageRange{Min: 16, Max: 16}, damageTypeCold)
	if countEventsBySkill(res.Events, "skill_effect_started", "rimecoil_lance") < 2 {
		t.Fatalf("events = %+v, want slow starts on primary and nearby target", res.Events)
	}
	if !containsStringValue(primary.effectIDs, "ice_slow") || !containsStringValue(nearby.effectIDs, "ice_slow") {
		t.Fatalf("slow effects primary=%v nearby=%v, want both slowed", primary.effectIDs, nearby.effectIDs)
	}
}

func TestBranchUniqueSunwakeHammerAppliesBurnOnDivineHammer(t *testing.T) {
	rules := cloneRules(loadRules(t))
	sim := MustNewSim("sess_sunwake", "sunwake", rules)
	sim.progression.Level = 24
	sim.savePlayer(sim.defaultPlayer())
	player := sim.entities[sim.playerID]
	target := uniqueTestMonster(sim, Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, 60)
	equipUniqueTestEffect(t, sim, "sunwake_hammer", 9913, "shield", offHandSlot)

	res := &TickResult{}
	sim.damageMonsterByPlayerSkillTypedWithID(target, player.id, "divine_hammer", "sunwake_hit", res, DamageRange{Min: 18, Max: 18}, damageTypeForce)
	if _, ok := eventAmount(*res, "skill_effect_started", "sunwake_hammer"); !ok {
		t.Fatalf("events = %+v, want sunwake burn start", res.Events)
	}
	if !containsStringValue(target.effectIDs, "burning") {
		t.Fatalf("target effects = %v, want burning", target.effectIDs)
	}
}

func countEventsBySkill(events []Event, eventType string, skillID string) int {
	count := 0
	for _, ev := range events {
		if ev.EventType == eventType && ev.SkillID == skillID {
			count++
		}
	}
	return count
}
