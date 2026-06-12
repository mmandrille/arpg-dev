package game

import (
	"math"
	"testing"
)

func TestEffectiveAttackSpeedUsesWeaponAndItemPercent(t *testing.T) {
	sim := MustNewSim("sess_attack_speed", "01", loadRules(t))
	blade := addRolledInventoryItem(t, sim, 6400, "cave_blade", nil)
	gloves := addRolledInventoryItem(t, sim, 6401, "cave_gloves", map[string]int{"attack_speed_percent": 10})
	assertAck(t, sim.Tick([]Input{{MessageID: "blade", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(blade.instanceID), Slot: mainHandSlot}}}), "blade")
	assertAck(t, sim.Tick([]Input{{MessageID: "gloves", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(gloves.instanceID), Slot: "gloves"}}}), "gloves")

	view := sim.CharacterProgressionView()
	baseAttackSpeed := sim.characterDerivedStatsView().AttackSpeed
	expectedBladeSpeed := sim.clampEffectiveAttackSpeed(baseAttackSpeed * sim.rules.ItemTemplates["cave_blade"].AttackSpeed * (1.0 + float64(gloves.rollPayload.Stats["attack_speed_percent"])/100.0))
	expectedBladeInterval := sim.attackIntervalTicksFromSpeed(expectedBladeSpeed)
	if math.Abs(view.DerivedStats.AttackSpeed-expectedBladeSpeed) > 0.000001 || view.DerivedStats.AttackIntervalTicks != expectedBladeInterval {
		t.Fatalf("attack speed/interval with blade+gloves = %+v, want %v / %d", view.DerivedStats, expectedBladeSpeed, expectedBladeInterval)
	}
	speed := findStatBreakdown(view.StatBreakdowns, "attack_speed")
	interval := findStatBreakdown(view.StatBreakdowns, "attack_interval_ticks")
	if speed == nil || interval == nil || !hasBreakdownSource(speed.Sources, "equipment_base") || !hasBreakdownSource(speed.Sources, "equipment_roll") {
		t.Fatalf("attack speed breakdowns = speed:%+v interval:%+v all:%+v", speed, interval, view.StatBreakdowns)
	}

	slow := MustNewSim("sess_attack_speed_slow", "01", loadRules(t))
	greatsword := addRolledInventoryItem(t, slow, 6402, "cave_greatsword", nil)
	assertAck(t, slow.Tick([]Input{{MessageID: "greatsword", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(greatsword.instanceID), Slot: mainHandSlot}}}), "greatsword")
	slowView := slow.CharacterProgressionView()
	expectedGreatswordSpeed := slow.clampEffectiveAttackSpeed(slow.characterDerivedStatsView().AttackSpeed * slow.rules.ItemTemplates["cave_greatsword"].AttackSpeed)
	expectedGreatswordInterval := slow.attackIntervalTicksFromSpeed(expectedGreatswordSpeed)
	if math.Abs(slowView.DerivedStats.AttackSpeed-expectedGreatswordSpeed) > 0.000001 || slowView.DerivedStats.AttackIntervalTicks != expectedGreatswordInterval {
		t.Fatalf("attack speed/interval with greatsword = %+v, want %v / %d", slowView.DerivedStats, expectedGreatswordSpeed, expectedGreatswordInterval)
	}
}

func TestHealthAndManaRegenUseStatsAndItemRolls(t *testing.T) {
	rules := loadRules(t)
	base := MustNewSim("sess_regen_base", "01", rules)
	player := base.entities[base.playerID]
	player.hp = player.maxHP - 2
	player.mana = player.maxMana - 2

	for i := 0; i < 199; i++ {
		base.Tick(nil)
	}
	if player.hp != player.maxHP-2 || player.mana != player.maxMana-2 {
		t.Fatalf("base regen before threshold hp/mana = %d/%d, want %d/%d", player.hp, player.mana, player.maxHP-2, player.maxMana-2)
	}
	heal := base.Tick(nil)
	if player.hp != player.maxHP-1 || player.mana != player.maxMana-1 {
		t.Fatalf("base regen after 10s hp/mana = %d/%d, want %d/%d", player.hp, player.mana, player.maxHP-1, player.maxMana-1)
	}
	if !hasPlayerResourceUpdate(heal, player.hp, player.mana) {
		t.Fatalf("base regen missing player update: %+v", heal.Changes)
	}

	geared := MustNewSim("sess_regen_item", "01", rules)
	ring := addRolledInventoryItem(t, geared, 6410, "cave_ring", map[string]int{"health_regen_per_10_seconds": 5, "mana_regen_per_10_seconds": 5})
	assertAck(t, geared.Tick([]Input{{MessageID: "ring", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(ring.instanceID), Slot: ringLeftSlot}}}), "ring")
	gearedPlayer := geared.entities[geared.playerID]
	gearedPlayer.hp = gearedPlayer.maxHP - 2
	gearedPlayer.mana = gearedPlayer.maxMana - 2
	view := geared.CharacterProgressionView()
	if math.Abs(view.DerivedStats.HealthRegenPerSecond-0.6) > 0.000001 || math.Abs(view.DerivedStats.ManaRegenPerSecond-0.6) > 0.000001 {
		t.Fatalf("geared regen stats = %+v, want 0.6/0.6", view.DerivedStats)
	}
	hpRegen := findStatBreakdown(view.StatBreakdowns, "health_regen_per_second")
	manaRegen := findStatBreakdown(view.StatBreakdowns, "mana_regen_per_second")
	if hpRegen == nil || manaRegen == nil || !hasBreakdownSource(hpRegen.Sources, "equipment_roll") || !hasBreakdownSource(manaRegen.Sources, "equipment_roll") {
		t.Fatalf("missing regen equipment breakdowns hp=%+v mana=%+v all=%+v", hpRegen, manaRegen, view.StatBreakdowns)
	}

	for i := 0; i < 33; i++ {
		geared.Tick(nil)
	}
	if gearedPlayer.hp != gearedPlayer.maxHP-2 || gearedPlayer.mana != gearedPlayer.maxMana-2 {
		t.Fatalf("geared regen before threshold hp/mana = %d/%d, want %d/%d", gearedPlayer.hp, gearedPlayer.mana, gearedPlayer.maxHP-2, gearedPlayer.maxMana-2)
	}
	gearedHeal := geared.Tick(nil)
	if gearedPlayer.hp != gearedPlayer.maxHP-1 || gearedPlayer.mana != gearedPlayer.maxMana-1 {
		t.Fatalf("geared regen after boosted ticks hp/mana = %d/%d, want %d/%d", gearedPlayer.hp, gearedPlayer.mana, gearedPlayer.maxHP-1, gearedPlayer.maxMana-1)
	}
	if !hasPlayerResourceUpdate(gearedHeal, gearedPlayer.hp, gearedPlayer.mana) {
		t.Fatalf("geared regen missing player update: %+v", gearedHeal.Changes)
	}
}
