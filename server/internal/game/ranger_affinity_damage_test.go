package game

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRangerAffinityFormulaGolden(t *testing.T) {
	var golden struct {
		Rule  RangerAffinityDamageDef `json:"rule"`
		Cases []struct {
			Name            string                   `json:"name"`
			Rank            int                      `json:"rank"`
			EffectiveStat   int                      `json:"effective_stat"`
			Affinities      int                      `json:"affinities"`
			RawMin          int                      `json:"raw_min"`
			RawMax          int                      `json:"raw_max"`
			ExpectedPercent int                      `json:"expected_percent"`
			ExpectedMin     int                      `json:"expected_min"`
			ExpectedMax     int                      `json:"expected_max"`
			RuleOverride    *RangerAffinityDamageDef `json:"rule_override"`
		} `json:"cases"`
	}
	loadGolden(t, "ranger_affinity_damage.json", &golden)
	deadeye := loadRules(t).Skills["deadeye"]
	if deadeye.PassiveStats.RangerAffinityDamage == nil || *deadeye.PassiveStats.RangerAffinityDamage != golden.Rule {
		t.Fatalf("Deadeye rule differs from shared golden: %+v", deadeye.PassiveStats.RangerAffinityDamage)
	}
	for _, tc := range golden.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			rule := golden.Rule
			if tc.RuleOverride != nil {
				rule = *tc.RuleOverride
			}
			percent := rangerAffinityPercent(rule, tc.EffectiveStat, tc.Affinities, tc.Rank)
			if percent != tc.ExpectedPercent {
				t.Fatalf("bonus = %d, want %d", percent, tc.ExpectedPercent)
			}
			got := scaleRangerAffinityDamage(DamageRange{Min: tc.RawMin, Max: tc.RawMax}, percent)
			if got.Min != tc.ExpectedMin || got.Max != tc.ExpectedMax {
				t.Fatalf("scaled range = %+v, want %d..%d", got, tc.ExpectedMin, tc.ExpectedMax)
			}
		})
	}
}

func TestRangerAffinityRulesRejectInvalid(t *testing.T) {
	deadeye := loadRules(t).Skills["deadeye"]
	cases := []struct {
		name   string
		mutate func(*RangerAffinityDamageDef)
	}{
		{"invalid stat", func(v *RangerAffinityDamageDef) { v.AffinityStat = "luck" }},
		{"zero denominator", func(v *RangerAffinityDamageDef) { v.StatPointsPerExtraPercent = 0 }},
		{"negative base", func(v *RangerAffinityDamageDef) { v.BasePercentPerAffinity = -1 }},
		{"excessive count", func(v *RangerAffinityDamageDef) { v.MaxActiveAffinities = 4 }},
		{"excessive cap", func(v *RangerAffinityDamageDef) { v.MaxBonusPercent = 26 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bad := deadeye
			rule := *deadeye.PassiveStats.RangerAffinityDamage
			tc.mutate(&rule)
			bad.PassiveStats.RangerAffinityDamage = &rule
			if err := validatePassiveStatSkillPayload("deadeye", bad); err == nil || !strings.Contains(err.Error(), "ranger_affinity_damage") {
				t.Fatalf("invalid rule accepted: %v", err)
			}
		})
	}
	bad := deadeye
	bad.Class = "rogue"
	if err := validatePassiveStatSkillPayload("deadeye", bad); err == nil {
		t.Fatal("wrong class accepted")
	}
	unsupported := loadRules(t).Skills["magic_bolt"]
	selected := true
	unsupported.RangerAffinityEligible = &selected
	if err := validateSkillKindPayload("magic_bolt", unsupported, nil); err == nil {
		t.Fatal("non-Ranger projectile tag accepted")
	}
	unsupported = loadRules(t).Skills["piercing_shot"]
	selected = false
	unsupported.RangerAffinityEligible = &selected
	if err := validateSkillKindPayload("piercing_shot", unsupported, nil); err == nil {
		t.Fatal("false Ranger projectile tag accepted")
	}
}

func TestRangerAffinityDamageEquipAndLaunch(t *testing.T) {
	sim := rangerSkillSim(t, "sess_ranger_affinity_launch")
	sim.progression.BaseStats.Dex = 17
	sim.progression.SkillRanks["deadeye"] = 1
	sim.defaultPlayer().Progression = sim.progression
	bow := addRolledItemWithAffinities(t, sim, 8850, "affinity_compound_bow", []ClassAffinityRoll{{Class: "ranger", Stat: "reach_percent", Value: 20}})
	quiver := addRolledItemWithAffinities(t, sim, 8851, "affinity_ranger_quiver", []ClassAffinityRoll{{Class: "ranger", Stat: "reach_percent", Value: 10}})
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_bow", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(bow.instanceID), Slot: mainHandSlot}}}), "equip_bow")
	first := sim.rangerAffinityBonusPercent()
	if first <= 0 || first != sim.DerivedStatsView().RangedDamageBonusPercent {
		t.Fatalf("one affinity bonus = %d, view = %d", first, sim.DerivedStatsView().RangedDamageBonusPercent)
	}
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_quiver", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(quiver.instanceID), Slot: "belt"}}}), "equip_quiver")
	second := sim.rangerAffinityBonusPercent()
	if second <= first || sim.beneficialRangerAffinityCount() != 2 {
		t.Fatalf("two affinity bonus/count = %d/%d, first = %d", second, sim.beneficialRangerAffinityCount(), first)
	}
	base := sim.resolvePlayerAttackDamage()
	fire := sim.Tick([]Input{{MessageID: "fire", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
	assertAck(t, fire, "fire")
	projectile := firstEntityByKind(sim, projectileEntity)
	if projectile == nil {
		t.Fatal("basic bow projectile was not spawned")
	}
	want := scaleRangerAffinityDamage(base, second)
	if projectile.damageRange != want {
		t.Fatalf("launched range = %+v, want %+v", projectile.damageRange, want)
	}
	assertAck(t, sim.Tick([]Input{{MessageID: "unequip_quiver", Type: "unequip_intent", Unequip: &UnequipIntent{Slot: "belt"}}}), "unequip_quiver")
	if sim.rangerAffinityBonusPercent() != first {
		t.Fatalf("unequip bonus = %d, want %d", sim.rangerAffinityBonusPercent(), first)
	}
	if projectile.damageRange != want {
		t.Fatalf("in-flight range changed: %+v, want %+v", projectile.damageRange, want)
	}
}

func TestRangerAffinityDamageFiltersAndCap(t *testing.T) {
	sim := rangerSkillSim(t, "sess_ranger_affinity_filters")
	sim.progression.BaseStats.Dex = 18
	sim.progression.SkillRanks["deadeye"] = 1
	sim.defaultPlayer().Progression = sim.progression
	quiver := addRolledItemWithAffinities(t, sim, 8860, "affinity_ranger_quiver", []ClassAffinityRoll{
		{Class: "ranger", Stat: "reach_percent", Value: 10},
		{Class: "rogue", Stat: "crit_chance", Value: 10},
		{Class: "ranger", Stat: "reach_percent", Value: -2},
		{Class: "rogue", Stat: "crit_chance", Value: -2, Mode: "penalty_if_not_class"},
	})
	if got := sim.rangerAffinityBonusPercent(); got != 0 {
		t.Fatalf("unequipped affinity bonus = %d, want 0", got)
	}
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_quiver", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(quiver.instanceID), Slot: "belt"}}}), "equip_quiver")
	if got := sim.beneficialRangerAffinityCount(); got != 1 {
		t.Fatalf("beneficial count = %d, want 1", got)
	}
	before := sim.rangerAffinityBonusPercent()
	sim.progression.BaseStats.Dex++
	sim.defaultPlayer().Progression = sim.progression
	if got := sim.rangerAffinityBonusPercent(); got != before+1 {
		t.Fatalf("Dex threshold bonus = %d, want %d", got, before+1)
	}
	quiver.rollPayload.ClassAffinities = append(quiver.rollPayload.ClassAffinities,
		ClassAffinityRoll{Class: "ranger", Stat: "reach_percent", Value: 9},
		ClassAffinityRoll{Class: "ranger", Stat: "reach_percent", Value: 8},
	)
	rule := *sim.rules.Skills["deadeye"].PassiveStats.RangerAffinityDamage
	if got, want := sim.rangerAffinityBonusPercent(), rangerAffinityPercent(rule, 20, 2, 1); got != want {
		t.Fatalf("three affinity cap = %d, want %d", got, want)
	}
	sim.progression.CharacterClass = "rogue"
	sim.defaultPlayer().Progression = sim.progression
	if got := sim.rangerAffinityBonusPercent(); got != 0 {
		t.Fatalf("wrong-class bonus = %d, want 0", got)
	}
}

func TestRangerAffinityProjectileEligibilityAndSnapshot(t *testing.T) {
	sim := rangerSkillSim(t, "sess_ranger_affinity_skill_launch")
	sim.progression.BaseStats.Dex = 40
	sim.progression.SkillRanks["deadeye"] = 1
	sim.defaultPlayer().Progression = sim.progression
	quiver := addRolledItemWithAffinities(t, sim, 8870, "affinity_ranger_quiver", []ClassAffinityRoll{
		{Class: "ranger", Stat: "reach_percent", Value: 10},
		{Class: "ranger", Stat: "reach_percent", Value: 9},
	})
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_quiver", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(quiver.instanceID), Slot: "belt"}}}), "equip_quiver")
	def := sim.rules.Skills["snipe"]
	base := sim.scaleSkillDamageForMagic(def, 1, sim.skillDamageRangeForSkill("snipe", def, 1))
	projectile := sim.spawnSkillProjectile(sim.activeLevel().entities[sim.playerID], "snipe", def, 1, Vec2{X: 1}, 0, Input{MessageID: "snipe"})
	want := scaleRangerAffinityDamage(base, sim.rangerAffinityBonusPercent())
	if projectile.damageRange != want {
		t.Fatalf("tagged skill range = %+v, want %+v", projectile.damageRange, want)
	}
	untagged := def
	untagged.RangerAffinityEligible = nil
	if got := sim.rangerAffinitySkillDamage(untagged, DamageRange{Min: 100, Max: 200}); got != (DamageRange{Min: 100, Max: 200}) {
		t.Fatalf("untagged skill scaled: %+v", got)
	}
	if got := sim.rangerAffinityBasicDamage(DamageRange{Min: 100, Max: 200}); got != (DamageRange{Min: 100, Max: 200}) {
		t.Fatalf("unarmed basic damage scaled: %+v", got)
	}
	assertAck(t, sim.Tick([]Input{{MessageID: "unequip_quiver", Type: "unequip_intent", Unequip: &UnequipIntent{Slot: "belt"}}}), "unequip_quiver")
	if projectile.damageRange != want {
		t.Fatalf("in-flight skill range changed: %+v, want %+v", projectile.damageRange, want)
	}
}

func TestRangerAffinityReplay(t *testing.T) {
	newSim := func() *Sim {
		sim := rangerSkillSim(t, "sess_ranger_affinity_replay")
		sim.progression.BaseStats.Dex = 17
		sim.progression.SkillRanks["deadeye"] = 1
		sim.defaultPlayer().Progression = sim.progression
		player := sim.activeLevel().entities[sim.playerID]
		player.pos = Vec2{X: 2, Y: 2}
		addRangerSkillMonster(sim, Vec2{X: 6, Y: 2}, 100)
		bow := addRolledItemWithAffinities(t, sim, 8880, "affinity_compound_bow", []ClassAffinityRoll{{Class: "ranger", Stat: "reach_percent", Value: 20}})
		quiver := addRolledItemWithAffinities(t, sim, 8881, "affinity_ranger_quiver", []ClassAffinityRoll{{Class: "ranger", Stat: "reach_percent", Value: 10}})
		assertAck(t, sim.Tick([]Input{{MessageID: "equip_bow", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(bow.instanceID), Slot: mainHandSlot}}}), "equip_bow")
		assertAck(t, sim.Tick([]Input{{MessageID: "equip_quiver", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(quiver.instanceID), Slot: "belt"}}}), "equip_quiver")
		return sim
	}
	first, second := newSim(), newSim()
	inputs := [][]Input{
		{{MessageID: "pierce", Type: "cast_skill_intent", CastSkill: &CastSkillIntent{SkillID: "piercing_shot", Direction: &Vec2{X: 1}}}},
		nil,
		{{MessageID: "unequip_quiver", Type: "unequip_intent", Unequip: &UnequipIntent{Slot: "belt"}}},
		nil,
	}
	for tick, in := range inputs {
		a, b := first.Tick(in), second.Tick(in)
		jsonA, errA := json.Marshal(a)
		jsonB, errB := json.Marshal(b)
		if errA != nil || errB != nil || string(jsonA) != string(jsonB) {
			t.Fatalf("tick %d replay differs: %v %v", tick, errA, errB)
		}
		if tick == 0 && !hasSkillDamageEvent(a, "piercing_shot") {
			t.Fatalf("replay shot has no server damage event: %+v", a.Events)
		}
	}
}

func TestRangerAffinityPiercingShotScalesEachHitOnce(t *testing.T) {
	newSim := func() (*Sim, []*entity) {
		sim := rangerSkillSim(t, "sess_ranger_affinity_pierce")
		sim.progression.BaseStats.Dex = 40
		sim.progression.SkillRanks["deadeye"] = 1
		sim.defaultPlayer().Progression = sim.progression
		item := addRolledItemWithAffinities(t, sim, 8890, "affinity_ranger_quiver", []ClassAffinityRoll{
			{Class: "ranger", Stat: "reach_percent", Value: 10},
			{Class: "ranger", Stat: "reach_percent", Value: 9},
		})
		assertAck(t, sim.Tick([]Input{{MessageID: "equip", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(item.instanceID), Slot: "belt"}}}), "equip")
		return sim, []*entity{
			addRangerSkillMonster(sim, Vec2{X: 6, Y: 2}, 1000),
			addRangerSkillMonster(sim, Vec2{X: 9, Y: 2}, 1000),
		}
	}
	withBonus, bonusTargets := newSim()
	baseline, baselineTargets := newSim()
	def := withBonus.rules.Skills["piercing_shot"]
	def.Damage = SkillDamageDef{Type: "rank_linear_range", MinBase: 100, MaxBase: 100}
	def.Pierce.MaxHits = 2
	def.Pierce.DamagePercentPerExtraHit = 100
	plainDef := def
	plainDef.RangerAffinityEligible = nil
	apply := func(sim *Sim, skill SkillDef, targets []*entity) {
		rows := []rangerLineTarget{{Target: targets[0]}, {Target: targets[1]}}
		res := TickResult{}
		sim.applyRangerShot(sim.activeLevel().entities[sim.playerID], "piercing_shot", skill, 1, rows, "corr", &res)
		if countSkillDamageEvents(res, "piercing_shot") != 2 {
			t.Fatalf("piercing hit events = %+v", res.Events)
		}
	}
	apply(withBonus, def, bonusTargets)
	apply(baseline, plainDef, baselineTargets)
	for i := range bonusTargets {
		baseDamage := baselineTargets[i].maxHP - baselineTargets[i].hp
		bonusDamage := bonusTargets[i].maxHP - bonusTargets[i].hp
		if baseDamage <= 0 || bonusDamage <= baseDamage {
			t.Fatalf("hit %d baseline/bonus = %d/%d", i, baseDamage, bonusDamage)
		}
	}
}
