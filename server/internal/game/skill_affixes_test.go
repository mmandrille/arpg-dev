package game

import "testing"

func TestSkillAffixRollsReduceManaCostAndCooldown(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_skill_affixes", "01", rules)
	sim.progression.CharacterClass = "sorcerer"
	sim.progression.BaseStats = rules.CharacterProgression.Classes["sorcerer"].BaseStats
	sim.progression.SkillRanks["magic_bolt"] = 1
	player := sim.entities[sim.playerID]
	player.mana = 12
	staff := addRolledInventoryItem(t, sim, 6600, "starter_sorcerer_staff", map[string]int{
		"skill_mana_cost_reduction":        1,
		"skill_cooldown_reduction_percent": 50,
	})
	assertAck(t, sim.Tick([]Input{{MessageID: "staff", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(staff.instanceID), Slot: mainHandSlot}}}), "staff")

	def := rules.Skills["magic_bolt"]
	baseMana := skillManaCost(def, 1)
	if got := sim.effectiveSkillManaCost(def, 1); got != baseMana-1 {
		t.Fatalf("effective mana cost = %d, want %d", got, baseMana-1)
	}
	baseCooldown := sim.baseSkillCooldownTicks(def)
	if got := sim.skillCooldownTicks(def); got >= baseCooldown || got != (baseCooldown+1)/2 {
		t.Fatalf("effective cooldown = %d, want reduced half of %d", got, baseCooldown)
	}

	cast := sim.Tick([]Input{{MessageID: "cast", Type: "cast_skill_intent", CastSkill: &CastSkillIntent{SkillID: "magic_bolt", Direction: &Vec2{X: 1}}}})
	assertAck(t, cast, "cast")
	if player.mana != 12-(baseMana-1) {
		t.Fatalf("player mana after reduced cast = %d, want %d", player.mana, 12-(baseMana-1))
	}
	if !hasSkillCastMana(cast, "magic_bolt", baseMana-1) {
		t.Fatalf("skill_cast mana event missing reduced cost %d: %+v", baseMana-1, cast.Events)
	}
	cooldowns := skillCooldownUpdate(cast)
	if len(cooldowns) != 1 || cooldowns[0].TotalTicks != sim.skillCooldownTicks(def) {
		t.Fatalf("cooldown update = %+v, want total %d", cooldowns, sim.skillCooldownTicks(def))
	}
}

func TestSkillAffixRollCandidates(t *testing.T) {
	rules := loadRules(t)
	staff := rules.rollableStatsForRarity(rules.ItemTemplates["starter_sorcerer_staff"].RollableStats, "rare", 1)
	for _, stat := range []string{"skill_damage_percent", "skill_cooldown_reduction_percent", "skill_mana_cost_reduction"} {
		if _, ok := findRollableStat(staff, stat); !ok {
			t.Fatalf("rare starter_sorcerer_staff pool missing %s", stat)
		}
	}
	amulet := rules.rollableStatsForRarity(rules.ItemTemplates["amulet"].RollableStats, "rare", 1)
	if _, ok := findRollableStat(amulet, "skill_cooldown_reduction_percent"); !ok {
		t.Fatal("rare amulet pool missing skill_cooldown_reduction_percent")
	}
}

func hasSkillCastMana(r TickResult, skillID string, mana int) bool {
	for _, event := range r.Events {
		if event.EventType == "skill_cast" && event.SkillID == skillID && event.Mana != nil && *event.Mana == mana {
			return true
		}
	}
	return false
}
