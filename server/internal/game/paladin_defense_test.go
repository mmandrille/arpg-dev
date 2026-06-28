package game

import (
	"math"
	"testing"
)

func TestHolyShieldAreaBuffAppliesDefenseVisualStateAndExpires(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_holy_shield", "01", rules)
	sim.progression.CharacterClass = "paladin"
	sim.savePlayer(sim.defaultPlayer())
	player := sim.entities[sim.playerID]
	sim.progression.BaseStats.Vit = 8
	sim.progression.BaseStats.Magic = 8
	sim.progression.SkillRanks["holy_shield"] = 1
	player.mana = player.maxMana
	beforeMana := player.mana
	sim.savePlayer(sim.defaultPlayer())

	before, _ := sim.playerEffectiveCombatStats()
	cast := sim.Tick([]Input{{
		MessageID:     "cast_holy_shield",
		CorrelationID: "corr_holy_shield",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "holy_shield"},
	}})
	assertAck(t, cast, "cast_holy_shield")
	if player.mana != beforeMana-5 {
		t.Fatalf("holy shield mana after cast = %d, want %d", player.mana, beforeMana-5)
	}
	if !hasEvent(cast, "skill_cast") || !hasEvent(cast, "skill_effect_started") || !hasEvent(cast, "skill_cooldown_started") {
		t.Fatalf("holy shield missing cast/effect/cooldown events: %+v", cast.Events)
	}
	if !sameStringSlice(player.effectIDs, []string{"holy_shield"}) {
		t.Fatalf("player effect ids = %v, want holy_shield", player.effectIDs)
	}
	progression := characterProgressionUpdate(cast)
	if progression == nil {
		t.Fatalf("holy shield cast missing character progression update: %+v", cast.Changes)
	}
	if progression.DerivedStats.BlockPercent <= before.BlockPercent {
		t.Fatalf("holy shield progression block = %v, want above %v", progression.DerivedStats.BlockPercent, before.BlockPercent)
	}
	if block := findStatBreakdown(progression.StatBreakdowns, "block_percent"); block == nil || !statBreakdownHasSourceKind(*block, "skill_effect") {
		t.Fatalf("holy shield progression block breakdown missing skill source: %+v", block)
	}
	after, breakdowns := sim.playerEffectiveCombatStats()
	if after.Armor <= before.Armor || after.BlockPercent <= before.BlockPercent {
		t.Fatalf("holy shield stats before=%+v after=%+v", before, after)
	}
	if armor := findStatBreakdown(breakdowns, "armor"); armor == nil || !statBreakdownHasSourceKind(*armor, "skill_effect") {
		t.Fatalf("holy shield armor breakdown missing skill source: %+v", armor)
	}
	if block := findStatBreakdown(breakdowns, "block_percent"); block == nil || !statBreakdownHasSourceKind(*block, "skill_effect") {
		t.Fatalf("holy shield block breakdown missing skill source: %+v", block)
	}

	var expired TickResult
	for i := 0; i < 300; i++ {
		expired = sim.Tick(nil)
	}
	if !hasEvent(expired, "skill_effect_ended") {
		t.Fatalf("holy shield expiry missing event: %+v", expired.Events)
	}
	if len(player.effectIDs) != 0 {
		t.Fatalf("holy shield effect ids after expiry = %v, want empty", player.effectIDs)
	}
	finalStats, _ := sim.playerEffectiveCombatStats()
	if math.Abs(finalStats.Armor-before.Armor) > 0.000001 || math.Abs(finalStats.BlockPercent-before.BlockPercent) > 0.000001 {
		t.Fatalf("holy shield stats after expiry=%+v want before=%+v", finalStats, before)
	}
}

func TestHolyShieldIgnoresAimAndAlwaysBuffsCaster(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_holy_shield_caster_centered", "01", rules)
	sim.progression.CharacterClass = "paladin"
	sim.progression.BaseStats.Vit = 8
	sim.progression.BaseStats.Magic = 8
	sim.progression.SkillRanks["holy_shield"] = 1
	sim.savePlayer(sim.defaultPlayer())
	player := sim.entities[sim.playerID]
	player.mana = player.maxMana
	beforeMana := player.mana
	sim.savePlayer(sim.defaultPlayer())

	cast := sim.Tick([]Input{{
		MessageID:     "cast_holy_shield_bad_target",
		CorrelationID: "corr_holy_shield_bad_target",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "holy_shield", TargetID: "999999", Direction: &Vec2{X: 1, Y: 0}},
	}})
	assertAck(t, cast, "cast_holy_shield_bad_target")
	if !hasEvent(cast, "skill_effect_started") {
		t.Fatalf("holy shield with bad target missing caster buff event: %+v", cast.Events)
	}
	if !sameStringSlice(player.effectIDs, []string{"holy_shield"}) {
		t.Fatalf("player effect ids = %v, want holy_shield", player.effectIDs)
	}
	if player.mana != beforeMana-5 {
		t.Fatalf("holy shield mana after bad-target cast = %d, want %d", player.mana, beforeMana-5)
	}
}

func TestSanctuaryGrantsTemporaryDamageImmunity(t *testing.T) {
	rules := cloneRules(loadRules(t))
	forceMonsterHitChance(rules, monsterDefID, 1.0)
	sim := MustNewSim("sess_sanctuary_immunity", "01", rules)
	sim.progression.CharacterClass = "paladin"
	sim.progression.BaseStats.Vit = 10
	sim.progression.BaseStats.Magic = 10
	sim.progression.SkillRanks["holy_shield"] = 1
	sim.progression.SkillRanks["sanctuary"] = 1
	sim.savePlayer(sim.defaultPlayer())
	player := sim.entities[sim.playerID]
	player.mana = player.maxMana
	attacker := addTestMonster(sim, monsterDefID, Vec2{X: player.pos.X + 1, Y: player.pos.Y}, 20)

	cast := sim.Tick([]Input{{
		MessageID:     "cast_sanctuary",
		CorrelationID: "corr_sanctuary",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "sanctuary"},
	}})
	assertAck(t, cast, "cast_sanctuary")
	if !sameStringSlice(player.effectIDs, []string{"sanctuary"}) {
		t.Fatalf("player effect ids = %v, want sanctuary", player.effectIDs)
	}
	if started := skillEvent(cast.Events, "skill_effect_started", "sanctuary"); started == nil || started.RemainingTicks == nil || *started.RemainingTicks != 60 {
		t.Fatalf("sanctuary start event = %+v, want 60 ticks", started)
	}
	if cooldown := skillEvent(cast.Events, "skill_cooldown_started", "sanctuary"); cooldown == nil || cooldown.RemainingTicks == nil || *cooldown.RemainingTicks != 598 {
		t.Fatalf("sanctuary cooldown event = %+v, want 598 ticks", cooldown)
	}

	beforeHP := player.hp
	res := &TickResult{}
	outcome := sim.damagePlayerByMonster(attacker, player, DamageRange{Min: 10, Max: 10}, "sanctuary_hit", res)
	if outcome.Outcome != "immune" || outcome.Damage != 0 || player.hp != beforeHP {
		t.Fatalf("sanctuary damage outcome=%+v hp=%d want immune and hp %d", outcome, player.hp, beforeHP)
	}
	if ev := firstEventOfType(res.Events, "player_damaged"); ev == nil || ev.Outcome != "immune" || ev.Damage == nil || *ev.Damage != 0 {
		t.Fatalf("sanctuary damage events = %+v, want immune zero-damage player_damaged", res.Events)
	}

	var expired TickResult
	for i := 0; i < 60; i++ {
		expired = sim.Tick(nil)
	}
	if !hasEvent(expired, "skill_effect_ended") || len(player.effectIDs) != 0 {
		t.Fatalf("sanctuary expiry events/effects = %+v / %v, want ended and empty", expired.Events, player.effectIDs)
	}

	after := &TickResult{}
	outcome = sim.damagePlayerByMonster(attacker, player, DamageRange{Min: 10, Max: 10}, "after_sanctuary", after)
	if outcome.Damage <= 0 || player.hp >= beforeHP {
		t.Fatalf("post-sanctuary damage outcome=%+v hp=%d want normal damage below %d", outcome, player.hp, beforeHP)
	}
}

func firstEventOfType(events []Event, eventType string) *Event {
	for idx := range events {
		if events[idx].EventType == eventType {
			return &events[idx]
		}
	}
	return nil
}

func skillEvent(events []Event, eventType string, skillID string) *Event {
	for idx := range events {
		if events[idx].EventType == eventType && events[idx].SkillID == skillID {
			return &events[idx]
		}
	}
	return nil
}
