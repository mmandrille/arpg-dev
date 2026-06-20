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

func TestCharacterProgressionViewEffectiveBaseStatsAndBreakdowns(t *testing.T) {
	sim := MustNewSim("sess_effective_base_stat_breakdowns", "01", loadRules(t))
	before := sim.CharacterProgressionView()
	ring := addRolledInventoryItem(t, sim, 6403, "cave_ring", map[string]int{
		"str": 10,
		"vit": 8,
	})
	assertAck(t, sim.Tick([]Input{{MessageID: "ring", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(ring.instanceID), Slot: ringLeftSlot}}}), "ring")

	view := sim.CharacterProgressionView()
	if view.EffectiveBaseStats.Str != before.BaseStats.Str+10 || view.EffectiveBaseStats.Vit != before.BaseStats.Vit+8 {
		t.Fatalf("effective base stats = %+v, base before %+v", view.EffectiveBaseStats, before.BaseStats)
	}
	strRow := findStatBreakdown(view.StatBreakdowns, "str")
	if strRow == nil || strRow.Value != float64(view.EffectiveBaseStats.Str) {
		t.Fatalf("strength breakdown = %+v, effective stats %+v", strRow, view.EffectiveBaseStats)
	}
	if !statBreakdownHasSourceKind(*strRow, "base_stat") || !statBreakdownHasSourceKind(*strRow, "equipment_roll") {
		t.Fatalf("strength breakdown missing base/equipment source: %+v", strRow)
	}
	if !statBreakdownHasItemSource(*strRow, idStr(ring.instanceID), sim.itemDisplayName(ring), 10) {
		t.Fatalf("strength breakdown missing item name/value source: %+v", strRow)
	}
}

func TestCritDamageUsesDexterityAsStandardDerivedStat(t *testing.T) {
	rules := cloneRules(loadRules(t))
	base := MustNewSim("sess_crit_damage_dex_base", "01", rules)
	highDexState := rules.DefaultCharacterProgressionState()
	highDexState.BaseStats.Dex += 10
	highDex, err := NewSimWithWorldProgression("sess_crit_damage_dex_high", "01", rules, DefaultWorldID, highDexState)
	if err != nil {
		t.Fatalf("new high dex sim: %v", err)
	}

	formula := rules.CharacterProgression.DerivedStats["crit_damage"]
	wantDelta := formula.PerDex * 10
	gotDelta := highDex.characterDerivedStatsView().CritDamage - base.characterDerivedStatsView().CritDamage
	if math.Abs(gotDelta-wantDelta) > 0.000001 {
		t.Fatalf("crit damage dex delta = %.6f, want %.6f from rule %+v", gotDelta, wantDelta, formula)
	}
}

func TestHealthAndManaRegenUseStatsAndItemRolls(t *testing.T) {
	rules := loadRules(t)
	base := MustNewSim("sess_regen_base", "01", rules)
	player := base.entities[base.playerID]
	player.hp = player.maxHP - 2
	player.mana = player.maxMana - 2

	for i := 0; i < 99; i++ {
		base.Tick(nil)
	}
	if player.hp != player.maxHP-2 || player.mana != player.maxMana-2 {
		t.Fatalf("base regen before threshold hp/mana = %d/%d, want %d/%d", player.hp, player.mana, player.maxHP-2, player.maxMana-2)
	}
	manaHeal := base.Tick(nil)
	if player.hp != player.maxHP-2 || player.mana != player.maxMana-1 {
		t.Fatalf("base regen after mana threshold hp/mana = %d/%d, want %d/%d", player.hp, player.mana, player.maxHP-2, player.maxMana-1)
	}
	if !hasPlayerResourceUpdate(manaHeal, player.hp, player.mana) {
		t.Fatalf("base mana regen missing player update: %+v", manaHeal.Changes)
	}
	if !hasManaRegenEvent(manaHeal, player.id, 1) {
		t.Fatalf("base mana regen missing event: %+v", manaHeal.Events)
	}
	for i := 0; i < 33; i++ {
		base.Tick(nil)
	}
	hpHeal := base.Tick(nil)
	if player.hp != player.maxHP-1 || player.mana != player.maxMana-1 {
		t.Fatalf("base regen after hp threshold hp/mana = %d/%d, want %d/%d", player.hp, player.mana, player.maxHP-1, player.maxMana-1)
	}
	if !hasPlayerResourceUpdate(hpHeal, player.hp, player.mana) {
		t.Fatalf("base hp regen missing player update: %+v", hpHeal.Changes)
	}
	if hasEvent(hpHeal, "player_mana_regenerated") {
		t.Fatalf("base hp regen emitted mana event: %+v", hpHeal.Events)
	}

	geared := MustNewSim("sess_regen_item", "01", rules)
	ring := addRolledInventoryItem(t, geared, 6410, "cave_ring", map[string]int{"health_regen_per_10_seconds": 5, "mana_regen_per_10_seconds": 5})
	assertAck(t, geared.Tick([]Input{{MessageID: "ring", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(ring.instanceID), Slot: ringLeftSlot}}}), "ring")
	gearedPlayer := geared.entities[geared.playerID]
	gearedPlayer.hp = gearedPlayer.maxHP - 2
	gearedPlayer.mana = gearedPlayer.maxMana - 2
	view := geared.CharacterProgressionView()
	if math.Abs(view.DerivedStats.HealthRegenPerSecond-0.65) > 0.000001 || math.Abs(view.DerivedStats.ManaRegenPerSecond-0.7) > 0.000001 {
		t.Fatalf("geared regen stats = %+v, want 0.65/0.7", view.DerivedStats)
	}
	hpRegen := findStatBreakdown(view.StatBreakdowns, "health_regen_per_second")
	manaRegen := findStatBreakdown(view.StatBreakdowns, "mana_regen_per_second")
	if hpRegen == nil || manaRegen == nil || !hasBreakdownSource(hpRegen.Sources, "equipment_roll") || !hasBreakdownSource(manaRegen.Sources, "equipment_roll") {
		t.Fatalf("missing regen equipment breakdowns hp=%+v mana=%+v all=%+v", hpRegen, manaRegen, view.StatBreakdowns)
	}

	for i := 0; i < 28; i++ {
		geared.Tick(nil)
	}
	if gearedPlayer.hp != gearedPlayer.maxHP-2 || gearedPlayer.mana != gearedPlayer.maxMana-2 {
		t.Fatalf("geared regen before threshold hp/mana = %d/%d, want %d/%d", gearedPlayer.hp, gearedPlayer.mana, gearedPlayer.maxHP-2, gearedPlayer.maxMana-2)
	}
	gearedManaHeal := geared.Tick(nil)
	if gearedPlayer.hp != gearedPlayer.maxHP-2 || gearedPlayer.mana != gearedPlayer.maxMana-1 {
		t.Fatalf("geared regen after mana threshold hp/mana = %d/%d, want %d/%d", gearedPlayer.hp, gearedPlayer.mana, gearedPlayer.maxHP-2, gearedPlayer.maxMana-1)
	}
	if !hasPlayerResourceUpdate(gearedManaHeal, gearedPlayer.hp, gearedPlayer.mana) {
		t.Fatalf("geared mana regen missing player update: %+v", gearedManaHeal.Changes)
	}
	if !hasManaRegenEvent(gearedManaHeal, gearedPlayer.id, 1) {
		t.Fatalf("geared mana regen missing event: %+v", gearedManaHeal.Events)
	}
	geared.Tick(nil)
	gearedHPHeal := geared.Tick(nil)
	if gearedPlayer.hp != gearedPlayer.maxHP-1 || gearedPlayer.mana != gearedPlayer.maxMana-1 {
		t.Fatalf("geared regen after hp threshold hp/mana = %d/%d, want %d/%d", gearedPlayer.hp, gearedPlayer.mana, gearedPlayer.maxHP-1, gearedPlayer.maxMana-1)
	}
	if !hasPlayerResourceUpdate(gearedHPHeal, gearedPlayer.hp, gearedPlayer.mana) {
		t.Fatalf("geared hp regen missing player update: %+v", gearedHPHeal.Changes)
	}
	if hasEvent(gearedHPHeal, "player_mana_regenerated") {
		t.Fatalf("geared hp regen emitted mana event: %+v", gearedHPHeal.Events)
	}
}

func hasManaRegenEvent(res TickResult, entityID uint64, mana int) bool {
	for _, event := range res.Events {
		if event.EventType == "player_mana_regenerated" && event.EntityID == idStr(entityID) && event.Mana != nil && *event.Mana == mana {
			return true
		}
	}
	return false
}

func statBreakdownHasItemSource(row StatBreakdownView, itemID string, label string, value float64) bool {
	for _, source := range row.Sources {
		if source.ItemInstanceID == itemID && source.Label == label && source.Value == value {
			return true
		}
	}
	return false
}

func TestStarterStaffAddsMaxManaAndSkillDamage(t *testing.T) {
	sim := MustNewSim("sess_starter_staff_stats", "01", loadRules(t))
	staff := addRolledInventoryItem(t, sim, 6420, "starter_sorcerer_staff", map[string]int{
		"damage_min":           2,
		"damage_max":           4,
		"max_mana":             5,
		"skill_damage_percent": 50,
	})
	assertAck(t, sim.Tick([]Input{{MessageID: "staff", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(staff.instanceID), Slot: mainHandSlot}}}), "staff")

	view := sim.CharacterProgressionView()
	baseMaxMana := sim.characterDerivedStatsView().MaxMana
	if math.Abs(view.DerivedStats.MaxMana-(baseMaxMana+5)) > 0.000001 {
		t.Fatalf("max mana with staff = %+v, want %v", view.DerivedStats, baseMaxMana+5)
	}
	maxMana := findStatBreakdown(view.StatBreakdowns, "max_mana")
	if maxMana == nil || !hasBreakdownSource(maxMana.Sources, "equipment_base") || !hasBreakdownSource(maxMana.Sources, "equipment_roll") {
		t.Fatalf("max mana breakdown = %+v all=%+v", maxMana, view.StatBreakdowns)
	}

	scaled := sim.applySkillDamageBonus(DamageRange{Min: 4, Max: 4})
	outcome := sim.resolveSkillDamage(effectiveCombatStats{}, scaled)
	if outcome.RawDamage != 6 || outcome.Damage != 6 {
		t.Fatalf("skill damage with staff = %+v, want raw/final 6", outcome)
	}
}
