package game

import (
	"fmt"
	"testing"
)

func TestAffixDisplayNameElementalColdWeapon(t *testing.T) {
	rules := loadRules(t)
	template := rules.ItemTemplates["long_sword"]
	stats := cloneIntMap(template.BaseStats)
	stats["bonus_cold_damage"] = 3

	got := rules.affixDisplayName(template, "magic", stats)
	if got != "Freezing Long Sword" {
		t.Fatalf("affix display name = %q, want Freezing Long Sword", got)
	}
}

func TestDominantElementalDamageType(t *testing.T) {
	stats := map[string]int{
		"bonus_cold_damage":      2,
		"bonus_fire_damage":      1,
		"bonus_lightning_damage": 0,
		"bonus_poison_damage":    0,
	}
	if got := dominantElementalDamageType(stats); got != damageTypeCold {
		t.Fatalf("dominant elemental = %q, want %q", got, damageTypeCold)
	}
}

func TestElementalBonusDamageAddsFlatDamage(t *testing.T) {
	stats := map[string]int{
		"damage_min":        2,
		"damage_max":        4,
		"bonus_cold_damage": 2,
		"bonus_fire_damage": 1,
	}
	if got := elementalBonusDamage(stats); got != 3 {
		t.Fatalf("elemental bonus = %d, want 3", got)
	}
}

func TestRolledSetDisplayNameGrammar(t *testing.T) {
	rules := loadRules(t)
	payload, ok := rules.setItemPayload("verdant_vanguard_blade")
	if !ok {
		t.Fatal("setItemPayload returned false")
	}
	if payload.DisplayName != "Savage Long Sword of Verdant Vanguard" {
		t.Fatalf("set display name = %q, want Savage Long Sword of Verdant Vanguard", payload.DisplayName)
	}
	if payload.SetPieceID != "verdant_vanguard_blade" {
		t.Fatalf("set piece id = %q, want verdant_vanguard_blade", payload.SetPieceID)
	}
}

func TestNamedUniqueDisplayNameGrammar(t *testing.T) {
	rules := loadRules(t)
	payload, ok := rules.namedUniquePayload("embercall_blade")
	if !ok {
		t.Fatal("namedUniquePayload returned false")
	}
	if payload.DisplayName != "Embercall Blade" {
		t.Fatalf("unique display name = %q, want Embercall Blade", payload.DisplayName)
	}
}

func TestArchetypeLabBotSeeds(t *testing.T) {
	rules := loadRules(t)
	for i := uint64(0); i < 50000; i++ {
		seed := fmt.Sprintf("%016x", i)
		rng := NewRNG(SeedToUint64(seed))
		first, ok := rules.rollItemTemplateWithRNG("long_sword", rng, 5)
		if !ok || first.Rarity != "magic" || first.Stats["bonus_cold_damage"] <= 0 {
			continue
		}
		_, ok = rules.rollItemTemplateWithRNG("long_sword", rng, 5)
		if !ok {
			continue
		}
		_, ok = rules.rollItemTemplateWithRNG("long_sword", rng, 5)
		if !ok {
			continue
		}
		third, ok := rules.rollItemTemplateWithRNG("long_sword", rng, 5)
		if !ok || third.Rarity != "common" || third.DisplayName != "Long Sword" {
			continue
		}
		t.Logf("bot_seed=%s first=%q third=%q", seed, first.DisplayName, third.DisplayName)
		return
	}
	t.Fatal("no bot seed found")
}
