package game

import "testing"

func TestBossEnrageValidation(t *testing.T) {
	rules := loadRules(t)
	template := rules.BossTemplates["cave_warden"]
	if template.Enrage == nil {
		t.Fatal("cave_warden missing enrage rules")
	}
	if template.Enrage.HealthRatioThreshold <= 0 || template.Enrage.HealthRatioThreshold > 1 {
		t.Fatalf("threshold = %f, want > 0 and <= 1", template.Enrage.HealthRatioThreshold)
	}
	if template.Enrage.CooldownMultiplier <= 0 {
		t.Fatalf("cooldown multiplier = %f, want positive", template.Enrage.CooldownMultiplier)
	}

	badThreshold := template
	badThreshold.Enrage = &BossEnrageDef{HealthRatioThreshold: 0, CooldownMultiplier: template.Enrage.CooldownMultiplier}
	badTemplates := map[string]BossTemplateDef{"cave_warden": badThreshold}
	if err := validateBossTemplates(badTemplates, rules); err == nil {
		t.Fatal("zero enrage threshold validated")
	}

	badCooldown := template
	badCooldown.Enrage = &BossEnrageDef{HealthRatioThreshold: template.Enrage.HealthRatioThreshold, CooldownMultiplier: 0}
	badTemplates = map[string]BossTemplateDef{"cave_warden": badCooldown}
	if err := validateBossTemplates(badTemplates, rules); err == nil {
		t.Fatal("zero enrage cooldown multiplier validated")
	}
}

func TestBossEnrageEventAndView(t *testing.T) {
	rules := loadRules(t)
	sim, level := newBossEnrageSim(t, rules)
	boss := findBossEntity(t, level)
	enrage := rules.BossTemplates[boss.bossTemplateID].Enrage
	if enrage == nil {
		t.Fatal("boss template missing enrage")
	}

	boss.hp = int(float64(boss.maxHP) * enrage.HealthRatioThreshold)
	res := sim.Tick(nil)
	event := findBossEnrageEvent(res)
	if event.EventType == "" {
		t.Fatalf("missing boss_enraged event: %+v", res.Events)
	}
	if event.EntityID != idStr(boss.id) || event.TargetEntityID != idStr(boss.id) || event.BossTemplateID != boss.bossTemplateID {
		t.Fatalf("boss_enraged ids = %+v", event)
	}
	if event.HealthRatioThreshold == nil || *event.HealthRatioThreshold != enrage.HealthRatioThreshold {
		t.Fatalf("boss_enraged threshold = %+v, want %f", event.HealthRatioThreshold, enrage.HealthRatioThreshold)
	}

	view := sim.entityView(boss)
	if !view.Enraged {
		t.Fatalf("boss view not enraged: %+v", view)
	}
	if view.EnrageHealthRatioThreshold != enrage.HealthRatioThreshold {
		t.Fatalf("view threshold = %f, want %f", view.EnrageHealthRatioThreshold, enrage.HealthRatioThreshold)
	}

	next := sim.Tick(nil)
	if findBossEnrageEvent(next).EventType != "" {
		t.Fatalf("boss_enraged emitted more than once: %+v", next.Events)
	}
}

func TestBossEnrageShortensFutureCooldown(t *testing.T) {
	rules := loadRules(t)
	sim, level := newBossEnrageSim(t, rules)
	boss := findBossEntity(t, level)
	baseCooldown := rules.BossPatterns[boss.bossPatternID].CooldownTicks
	if baseCooldown <= 1 {
		t.Fatalf("base cooldown = %d, want > 1 for multiplier proof", baseCooldown)
	}
	if got := sim.bossPatternCooldownTicks(boss, baseCooldown); got != baseCooldown {
		t.Fatalf("non-enraged cooldown = %d, want %d", got, baseCooldown)
	}

	boss.bossEnraged = true
	enrage := rules.BossTemplates[boss.bossTemplateID].Enrage
	want := int(float64(baseCooldown) * enrage.CooldownMultiplier)
	if want < 1 {
		want = 1
	}
	if got := sim.bossPatternCooldownTicks(boss, baseCooldown); got != want {
		t.Fatalf("enraged cooldown = %d, want %d", got, want)
	}

	original := rules.BossTemplates[boss.bossTemplateID]
	modified := original
	modified.Enrage = &BossEnrageDef{HealthRatioThreshold: original.Enrage.HealthRatioThreshold, CooldownMultiplier: 0.01}
	rules.BossTemplates[boss.bossTemplateID] = modified
	if got := sim.bossPatternCooldownTicks(boss, baseCooldown); got != 1 {
		t.Fatalf("minimum enraged cooldown = %d, want 1", got)
	}
}

func newBossEnrageSim(t *testing.T, rules *Rules) (*Sim, *LevelState) {
	t.Helper()
	sim, err := NewSimWithWorld("sess_boss_enrage", "boss_floor_gate", rules, "dungeon_levels")
	if err != nil {
		t.Fatal(err)
	}
	level, err := sim.ensureDungeonLevel(-5)
	if err != nil {
		t.Fatal(err)
	}
	sim.currentLevel = -5
	placeDefaultPlayerOnLevel(t, sim, level, Vec2{X: 15, Y: 15})
	sim.syncCompatibilityFields()
	return sim, level
}

func findBossEnrageEvent(res TickResult) Event {
	for _, event := range res.Events {
		if event.EventType == "boss_enraged" {
			return event
		}
	}
	return Event{}
}
